# Plan: Blue Binary Format and Standalone VM

Goal: define a binary format for compiled blue programs (bytecode, constants, tokens) so compiled output can be reused when running, then use it to build a minimal VM executable that embeds no lexer/parser/compiler and can run precompiled bytecode, including as a single packed executable.

STATUS: IMPLEMENTED (all phases). See "Implementation notes" at the bottom for decisions taken, deviations, measured numbers and caveats discovered along the way.

Current state notes (from exploring the code):

- `compiler.Bytecode` = `{Instructions code.Instructions, Constants []object.Object, Tokens []*token.Token}`.
  A normal compile already merges `lib/core/core.b` + std modules + user program into ONE instruction stream and one shared constant pool (see `cmd/util.go newCompiler`, `compiler/compiler_core.go compileCore`, marker opcode `OpCoreCompiled`).
  So v1 of the format can store this final merged image directly, no module splitting needed yet.
- Reserved constants `object.OBJECT_CONSTANTS` occupy fixed indices 0..6 (`TRUE, FALSE, NULL, VM_IGNORE, BREAK, CONTINUE, USE_PARAM_STR_OBJ`) and must round-trip identically.
- `object/encoding.go` has partial CBOR encode/decode but explicitly does not support `CompiledFunction`, `Module`, `BlueStruct`, `ExecString`, `DefaultArgs`, `Ignore`, and closure decode is TODO.
- Runtime dependencies on lexer/parser/compiler that live OUTSIDE the compiler today:
  - `eval()` builtin / `OpEval`: `vm/vm.go vmStr()` runs lexer+parser+compiler+VM on a string at runtime.
  - `from_json()` builtin: `object/builtins.go` lexes/parses JSON using the blue lexer/parser, converts via `ParseJson(ast)`.
  - REPL (`repl/repl.go`), `lex`, `parse`, `doc`, `compile` CLI commands.
  - Error stack traces read `Bytecode.Tokens` via `OpNode` operands (`vm.prepareStackTraceAndReturnError`).
- Build tags select different std sources/builtins (`static`, `rgfw`; ui/gg static vs non-static variants), so a compiled image is only valid for the matching build flavor.

---

## Phase 1: Design the container format

- [x] Create new package `bluec` (name TBD) that depends only on `code`, `token`, `object` (no compiler import) and owns:
  - format constants (magic bytes, e.g. `BLUEBC\x00`, format version u16)
  - `Image` struct mirroring what the VM needs: instructions blob, constants blob, tokens blob, metadata
  - `Encode(img *bluec.Image) ([]byte, error)` and `Decode(data []byte) (*bluec.Image, error)`
- [x] Define layout (all little-endian unless noted, opcodes already big-endian operands, keep as-is):
  - header: magic, format version, blue version string (`consts.VERSION`), build-flavor fingerprint (build tags + opcode-set hash + builtin count), CRC32 (or SHA256) of payload, flags (e.g. tokens-stripped bit)
  - payload sections, each length-prefixed u32: instructions, constants, tokens, trailer
  - trailing trailer: total-payload size u64 + reverse magic, so a payload can also be found when APPENDED to an executable (scan from EOF)
- [x] Decide constant encoding style (decision point):
  - Option A: reuse existing CBOR `ObjectWrapper` infra in `object/encoding.go`
  - Option B: custom compact binary writer (smaller, faster, more work)
  - Recommendation: start with A (reuse + free fuzzing via cbor), keep interface narrow enough to swap later
  - CHOSEN: Option A. Pool encoding lives in `object/constpool.go` (`EncodeConstantPool` / `DecodeConstantPool` / `FindUnserializableConstant`) on top of the existing CBOR wrappers; `bluec` just length-prefixes blobs so the codec can be swapped later without touching the envelope.
- [x] Add `Fingerprint()` helper covering: opcode count/names hash, reserved-constant count, build tags, blue version. Loader refuses mismatched images.

Note: the struct is named `bluec.Bytecode` (not `Image`) since it doubles as the neutral home for the compiler's old Bytecode type; see Phase 3.

## Phase 2: Complete object serialization for the constant pool

- [x] Enumerate every object type the compiler can place in constants and mark serializable/not:
  - yes: Integer, UInteger, BigInteger, BigFloat, Float, Boolean, Null, Stringo, Bytes, Regex, List, Map, Set, ExecString, DefaultArgs, Ignore, Module (Name + HelpStr only, Env is nil at compile time), CompiledFunction
  - reject at compile-to-file time: GoObj, Process, Closure, Builtin, anything holding live Go state (error out with a clear message listing the constant index)
  - IMPLEMENTATION NOTE: the compiler emits ONE legitimate GoObj constant, `GoObj[[]string]` struct field-name lists (`compiler.go addConstant(object.NewGoObj(node.Fields))`). These get a dedicated faithful encoding (`i_STRUCT_FIELDS_OBJ`) instead of being rejected or lossily gob'd; every other GoObj now errors from the pool path with a clear message. `Module.Encode()` stays rejected for the runtime `save()` builtin (live Env) while compile-time module constants serialize Name+HelpStr.
  - BUG THIS FIXED: the old decoder turned any GoObj into `GoObjectGob`, which silently broke struct matching after an image round trip ("struct did not have fields in index: 0" in test_nested_match.b). Caught by the golden suite.
- [x] Implement `CompiledFunction` encode/decode covering ALL fields:
  - `Instructions`, `NumLocals`, `NumParameters`, `Parameters`, `ParameterHasDefault`, `NumDefaultParams`, `DisplayString`, `HelpStr`, `SpecialFunctionParameters map[NameIndexKey]map[NameIndexKey]Object`
- [x] Preserve reserved constant slots: decoder reconstructs `object.NewObjectConstants()` first, then decodes remaining entries in order so indices match exactly. Encoder validates the prefix via `ValidateReservedPrefix` (identity check) and skips writing it.
- [x] Handle nested objects inside List/Map/Set/DefaultArgs/SpecialFunctionParameters recursively with cycle safety (constants should be acyclic, assert anyway): depth limit of 512 returns an error instead of overflowing the stack.
- [x] Token table codec: serialize `token.Token{Type, Literal, Filepath, LineNumber, PositionInLine}` compactly (intern filepath strings, delta-encoded line numbers optional): all three strings interned into one table; per token five varints with zigzag line deltas.
- [x] Round-trip unit tests: compile sample programs, assert before/after equality of Instructions byte-for-byte, constants Inspect()-equal type-by-type, tokens equal (`bluec/binc_test.go`)
- [x] Fuzz or property-test Decode against malformed input (must return error, never panic): `FuzzDecode` plus a table of hand-made corruptions (bad magic, truncated header/body, flipped bit, corrupt CRC, bad version). One real bounds bug found and fixed by fuzzing (sections slice when trailer offset underflows).

## Phase 3: Write and run compiled files from the CLI

- [x] Extend `handleCompileCommand` in `cmd/cmd.go` / `cmd/util.go`:
  - `blue compile -o out.bluec file.b` writes the container file (keep current debug-print behavior when no `-o`)
  - optional `--no-tokens` flag to strip token section (smaller file, degraded error traces)
  - optional `--split-core` later: emit core image separately for caching (stretch goal) - NOT DONE, deferred as planned
- [x] New loader used by `vm` command: if input ends in `.bluec` (or magic sniff), skip lexer/parser/compiler entirely, decode, construct `compiler.Bytecode` equivalent, `vm.NewWithGlobalsStore(...)` and run (`cmd/util.go looksLikeImage/loadImageFile`)
- [x] Keep a neutral home for the Bytecode struct decision: either move `Bytecode` out of package `compiler` into `code` or `bluec` so the minimal runner does not import `compiler` (decision point, recommend moving to `bluec` and aliasing in compiler for compatibility). DONE: `type Bytecode = bluec.Bytecode` alias in compiler; vm takes `*bluec.Bytecode`.
- [x] Golden tests: for each program in `b_test_programs`, running source vs running `.bluec` must produce identical stdout and exit codes (`b_test_programs/binary_golden_test.go`). Runs each program in its own subprocess via the real CLI (in-process capture was polluted by leftover spawned goroutines between files). Deterministic files compare stdout + exit code exactly; network/pid/metrics files compare exit codes only.
- [x] Benchmark startup: time `blue vm file.b` vs `blue vm file.bluec` to prove reuse value (document numbers in this file when done):
  - Measured on darwin/arm64, rgfw flavor, warm cache, best-of-5 style:
    - trivial program: source ~20ms vs image ~20ms per run (both dominated by process startup)
    - program importing math+crypto+search: 10 runs source 0.264s (~26ms) vs image 0.199s (~20ms), ~25% faster startup
    - savings grow with std-module import count since compilation (incl. core.b merge) dominates startup

## Phase 4: Decouple VM/runtime packages from lexer, parser, compiler, ast

This is the enabling refactor for the minimal build. Strategy: dependency inversion via small hook variables, plus build-tag stubs, rather than keeping heavy imports.

- [x] `vm/vm.go`: remove direct lexer/parser/compiler usage from `vmStr()`. Introduce `var EvalHook func(src string) object.Object` (nil in minimal builds). Full builds (the main `blue` binary, repl, wasmmain) install the real implementation from a tiny glue file that IS allowed to import lexer/parser/compiler
  - `OpEval` handler calls the hook; nil hook returns a runtime error object: "eval is not available in this build"
  - NOTE: `vmStr` backs THREE call sites (OpEval, `to_num`, `load`); routing the hook through `vmStr` covers all of them. Hook installers: `cmd/util.go installFullBuildHooks` (desktop), wasmmain's own `evalSourceString`, and test-only glue in `vm` / `b_test_programs`.
- [x] `object/builtins.go from_json`: extract the AST-free path. Two options (decision point):
  - Option A (recommended): write a small dedicated JSON-to-Object converter (JSON is much simpler than blue); keep `ParseJson(ast)` for parity tests comparing outputs on a corpus
  - Option B: move `from_json` behind the same hook pattern as eval (unavailable in minimal builds)
  - CHOSEN: Option A. `object/json.go FromJsonString` uses encoding/json tokens with UseNumber. Number semantics mirror the parser's ExactFloat64 rule: floats that survive the float64 round-trip exactly stay Float, otherwise they promote to BigFloat (this parity was caught by test_from_json_and_is_valid.b). The historical AST path moved to `object/astjson` for corpus comparison.
- [x] Audit remaining imports so the following packages form a closed set with NO imports of lexer/parser/compiler/ast: `code`, `token`(data only), `object`, `vm`, `bluec`, `consts`, `util`
  - `object/object_util.go ParseJson` and related ast-dependent helpers move to a separate file/package excluded from the minimal build (moved to package `object/astjson`)
  - check `object/object.go ast usage and remove/relocate`: `Function.Parameters/Body` were ast nodes used ONLY by Inspect(); replaced with plain strings. `CreateHelpStringFromBodyTokensAstFun` moved into the compiler (its only consumer).
  - DEVIATION: new package `runner` (shared run/error-print path used by cmd AND bluerun) imports `lexer.GetErrorLineMessage`, which reads embedded core/std data files but pulls in no parser/compiler/ast. Documented in its package doc; the guard below encodes this two-tier rule.
- [x] Guard with CI: add a test that parses the import graph of those packages (or use a build tag `minivm` + `go build -tags minivm ./...` gate) so heavy imports cannot creep back: `tools/importguard/importguard_test.go` runs `go list -deps` over the closed set and fails on forbidden imports (strict tier forbids lexer too; runner tier allows lexer for error formatting only).
- [x] Make sure `vm.Clone`, spawn/process, http server, ws template paths do not secretly require compiled artifacts beyond constants (they operate on closures/globals, expected fine)

## Phase 5: Minimal runner executable template

- [x] New main package, e.g. `cmd/bluerun/main.go` built with `-tags minivm`:
  - imports ONLY the closed set above (vm, object, code, bluec, consts)
  - behavior:
    - `bluerun app.bluec` : load sidecar image and run
    - `bluerun` with no args: look for appended payload in own executable (`os.Executable()`, seek to trailer, validate fingerprint), run it
    - forward os.Args[1:] to the program (match current `args`/`os.args()` builtin behavior), honor BLUE_NO_COLOR etc.
    - REFINEMENT: if an appended payload exists (packed exe) ALL args are forwarded to the program and the payload always wins; sidecar mode (first arg = image path) applies only when there is no embedded payload. ARGV forwarding is implemented via `object.SetProgramArgs` because ARGV is populated from os.Args during package init, before main can adjust it.
  - exit code semantics copied from `cmd/util.go vmFileOrString` error handling
- [x] Reuse the error printing path (VMError prefix, token trace) without importing cmd; factor shared bits into a small package both cmd and bluerun can use if needed: new package `runner` (`runner.RunBytecode` returns an exit code; cmd's `runBytecode` is now a thin os.Exit wrapper around it).
- [x] Verify size: compare stripped `blue` vs `bluerun` binary sizes, record here:
  - darwin/arm64, rgfw flavor: blue stripped ~55.8MB vs bluerun stripped ~54.9MB. The toolchain removal saves little because object's builtin surface (fyne/raylib/sqlite/...) dominates binary size; the win of bluerun is the ABSENT toolchain (no parse/compile attack surface, no source loading), not bytes. A static-flavor bluerun would shrink much more, but static builds are currently broken by a pre-existing vendored fyne issue (`isScroollerPageOnTap` undefined in vendor/fyne internal widget), unrelated to this work.
- [x] Tests: build bluerun, run N sample programs (incl. closures, try/catch, defer, spawn, maps, regex, big ints), diff output against `blue vm`: covered corpus-wide by the golden suite (subprocess isolation) plus `TestPackProducesWorkingSingleExecutable`.

## Phase 6: Packing into a single executable

- [x] `blue pack -o myapp main.b` command:
  1. compile main.b through the normal pipeline (core+std+user merged image)
  2. encode via bluec, append to a copy of the bluerun template for HOST os/arch
  3. append trailer (size + reverse magic), chmod +x
- [x] Template acquisition strategy (decision point):
  - Option A: release process builds per-platform `bluerun` templates next to `blue` (make_release.sh gains a step); pack uses matching template, errors if missing/cross-target
  - Option B: pack shells out to `go build` requiring a toolchain (portable but needs Go installed)
  - Recommendation: A for releases, B as fallback flag `--go-build`. BOTH IMPLEMENTED: make_release/make_release_static gained template steps (`bluerun-<GOOS>-<GOARCH>` next to blue, same flavor tags); `--go-build` locates the blue source tree (walks up from cwd, then $BLUE_INSTALL_PATH) and rebuilds the template with the RUNNING binary's own flavor tags so fingerprints match. Template lookup order next to the blue executable: bluerun-<GOOS>-<GOARCH>, then bluerun.
- [x] Document cross-compilation caveat: template must match target GOOS/GOARCH; CGO-enabled flavors complicate cross targets, prefer static template for packing (README + fingerprint errors make mismatches loud rather than silent)
- [x] Self-check on startup: bluerun validates magic + CRC + fingerprint, prints actionable error on mismatch (distinct errors for bad magic, bad version, truncation, CRC, fingerprint; decode failures suggest recompiling with `blue compile -o out.bluec main.b`)
- [x] End-to-end test in CI: pack a demo program on mac/linux, run, compare output; ensure upx step still works on packed output (or document skipping upx): `TestPackProducesWorkingSingleExecutable` builds both binaries, packs, runs and compares output/exit codes against `blue demo.b`. upx note: upx recompression of packed executables has NOT been re-verified here (no upx on this machine); release scripts should keep packing AFTER any upx step or verify separately.

## Phase 7: Versioning, compatibility, docs

- [x] Format version bump policy documented in bluec package doc (any opcode/constant-layout change bumps)
- [x] Loader error messages distinguish: bad magic, unknown version, fingerprint mismatch, CRC failure, truncated payload
- [x] Update `USAGE` text in `cmd/cmd.go` for new commands/flags (`compile -o`, `pack`)
- [x] README/manual section: compiling to .bluec, packing single executables, supported flags (README "Compiled programs (.bluec) and single executables"; man page regenerates from --help via gen-man.sh at release time)
- [x] Record final limitation list below in manual once implementation landed (kept here + README summary)

---

## Features unavailable or changed in the minimal/single-executable build (informative)

1. `eval(...)`: requires the full lexer/parser/compiler pipeline at runtime. In the minimal build it returns a clear runtime error instead of evaluating. Programs relying on eval cannot be packed unless a future "fat packer" variant ships the whole toolchain (possible follow-up since hooks make it easy).
2. `from_json(...)`: currently implemented via the blue parser. Either reimplemented with a dedicated JSON converter (recommended, keeps feature) or returns the same unsupported-in-this-build error (if option B chosen). Behavior identical either way for valid JSON; error messages may differ slightly during migration. RESOLVED via option A: works everywhere now.
3. REPL: absent from the runner by definition (it exists only in the full `blue` binary, unaffected there).
4. Nice error traces depend on the Tokens section. If an image is compiled with `--no-tokens`, runtime errors print messages without file/line pointers.
5. Compile-time-only concerns are safe: user module imports (`import foo.bar`), std module imports, and doc/help strings are baked into the image at pack time; `BLUE_INSTALL_PATH` style dynamic source resolution has no meaning in a packed program (there is no compiler to consume it).
6. Build-flavor lock-in: an image is tied to the build tags/flavor (static vs rgfw, ui/gg static vs non-static) used when compiling it; the packer/loader fingerprint check rejects mismatches. The structural `minivm` tag itself is filtered out of the fingerprint (it selects which main package builds, not what the runtime can do).
7. Anything returning live Go objects (GoObj values held by builtins, Process handles, open servers) obviously cannot appear in serialized constants; they only exist at runtime, which is fine, but default-arg constants referencing such values would fail compilation to .bluec with a clear error (constant index included).

## Open decisions recap (resolve before/during implementation)

ALL RESOLVED:

- Constant encoding: CBOR reuse (Option A), pool codec isolated in `object/constpool.go`
- Bytecode struct relocation: moved to `bluec.Bytecode`, aliased as `compiler.Bytecode`
- `from_json` strategy: dedicated converter (`object/json.go`), AST path kept in `object/astjson` for parity tests
- Template distribution for `pack`: shipped per-platform templates (release scripts) PLUS `--go-build` fallback that reproduces the running binary's flavor
- Package name for the format: `bluec`

## Implementation notes, numbers and caveats (added while implementing)

- Container layout v1 (see bluec.go doc for the authoritative description): header (magic `BLUEBC\x00`, u16 version=1, u16 flags, lp-string blue version, lp-string fingerprint, u32 CRC32) then four length-prefixed sections (instructions, constants, tokens, meta-reserved) then a 16-byte trailer (u64 total size + reversed magic). Sidecar files and appended payloads are byte-identical layouts.
- Fingerprint string: `v<version>|ops:<hash64>|rc:<reserved-count>|tags:<sorted-tags>|<goos>/<goarch>`, where tags come from build info `-tags` minus `minivm`.
- Machine caveat discovered: building ANY desktop binary without the `rgfw` tag links raylib's bundled glfw AND fyne's glfw -> duplicate-symbol link failure (pre-existing, documented in README). Consequences encoded here: command-line `-tags` REPLACES GOFLAGS tags, so scripts/tests that build bluerun must merge ambient GOFLAGS tags with minivm. `pack --go-build` handles this by copying the running binary's own recorded tags.
- Startup benchmark numbers recorded under Phase 3 above; packed-executable startup measured ~89ms/run for a 59MB packed binary (page-cache dependent), vs ~20ms sidecar.
- Sizes: blue(rgfw, stripped) 55.8MB; bluerun(minivm+rgfw, stripped) 54.9MB; packed demo app 59.2MB unstripped template.
- Pre-existing issues encountered (NOT caused by this work, left alone):
  - vendored fyne breakage breaks `-tags static` CGO_ENABLED=0 builds (`isScrollerPageOnTap` undefined), so static bluerun sizes could not be measured
  - parsing `"#{... m["k"] ...}"` (nested quotes inside string interpolation) sends the parser into minutes-long pathological backtracking; avoided in tests' demo programs, worth a separate investigation
