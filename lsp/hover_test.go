package lsp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// hoverAt runs hover on the nth occurrence of word in the given line.
func hoverAt(t *testing.T, s *session, uri, text string, line int, word string) any {
	t.Helper()
	col := strings.Index(strings.Split(text, "\n")[line], word)
	if col < 0 {
		t.Fatalf("word %q not found on line %d", word, line)
	}
	return s.hover(textDocumentHoverParams{
		TextDocument: textDocumentIdentifier{URI: uri},
		Position:     position{Line: line, Character: col},
	})
}

func hoverValue(res any) (string, bool) {
	hr, ok := res.(*hoverResult)
	if !ok || hr == nil {
		return "", false
	}
	return hr.Contents.Value, true
}

// Hovering a plain name shows the statement that bound it.
func TestHoverDeclaration(t *testing.T) {
	text := "val pi = 3.14\nvar count = 0\nprint(pi)\nprint(count)\n"
	s, uri := openDoc(t, text)

	got, ok := hoverValue(hoverAt(t, s, uri, text, 2, "pi"))
	if !ok {
		t.Fatalf("hover over a used constant returned nothing")
	}
	if !strings.Contains(got, "val pi = 3.14") {
		t.Errorf("hover = %q, want the binding statement", got)
	}

	got2, ok := hoverValue(hoverAt(t, s, uri, text, 3, "count"))
	if !ok {
		t.Fatalf("hover over a used variable returned nothing")
	}
	if !strings.Contains(got2, "var count = 0") {
		t.Errorf("hover = %q, want the binding statement", got2)
	}
}

// A function header must be shown exactly as written, because blue has no
// return types and hand written parameter types would be made up.
func TestHoverFunctionHeaderVerbatim(t *testing.T) {
	text := "fun connect(host, port = 5432, dbname = \":memory:\") {\n\tprint(host)\n}\n\nprint(connect)\n"
	s, uri := openDoc(t, text)

	got, ok := hoverValue(hoverAt(t, s, uri, text, 4, "connect"))
	if !ok {
		t.Fatalf("hover over a declared function returned nothing")
	}
	want := "fun connect(host, port = 5432, dbname = \":memory:\")"
	if !strings.Contains(got, want) {
		t.Errorf("hover = %q, want the verbatim header %q", got, want)
	}
	if strings.Contains(got, "->") {
		t.Errorf("hover invented a return type: %q", got)
	}

	// Hovering on the declaration itself must work as well.
	if _, ok := hoverValue(hoverAt(t, s, uri, text, 0, "connect")); !ok {
		t.Error("hover over the function's own name returned nothing")
	}
}

// Docstrings are `##` lines standing alone above the declaration or right after
// its opening brace. A trailing comment on another statement's line belongs to
// that statement and must not leak into whatever comes next.
func TestHoverDocstringRules(t *testing.T) {
	text := strings.Join([]string{
		"val e = 2.718 # note about e",
		"## documents pi",
		"## second doc line",
		"val pi = 3.14",
		"print(pi)",
		"",
		"fun damped(x) {",
		"\t## decays with noise",
		"\tx",
		"}",
		"print(damped)",
	}, "\n")
	s, uri := openDoc(t, text)

	got, ok := hoverValue(hoverAt(t, s, uri, text, 4, "pi"))
	if !ok {
		t.Fatalf("hover over pi returned nothing")
	}
	if !strings.Contains(got, "documents pi") || !strings.Contains(got, "second doc line") {
		t.Errorf("own-line docstrings above a val were not shown: %q", got)
	}

	got2, ok := hoverValue(hoverAt(t, s, uri, text, 10, "damped"))
	if !ok {
		t.Fatalf("hover over damped returned nothing")
	}
	if !strings.Contains(got2, "decays with noise") {
		t.Errorf("in-body docstring was not shown: %q", got2)
	}
	if strings.Contains(got2, "note about e") {
		t.Errorf("trailing comment of an earlier statement leaked into %q", got2)
	}
}

// A dot call on a plain value compiles as the bare function or builtin of that
// name with the value passed as first argument, so hovering it must show that
// name's docs, not the receiver's.
func TestHoverDotCallOnValue(t *testing.T) {
	text := "val s = \"abc\"\nfun twice(x) {\n\tx * 2\n}\nprint(s.len())\nprint(21.twice())\n"
	s, uri := openDoc(t, text)

	got, ok := hoverValue(hoverAt(t, s, uri, text, 4, "len"))
	if !ok {
		t.Fatalf("hover over s.len() returned nothing")
	}
	if !strings.Contains(got, "returns the INTEGER length") {
		t.Errorf("hover = %q, want the builtin len's help text", got)
	}

	got2, ok := hoverValue(hoverAt(t, s, uri, text, 5, "twice"))
	if !ok {
		t.Fatalf("hover over 21.twice() returned nothing")
	}
	if !strings.Contains(got2, "fun twice(x)") {
		t.Errorf("hover = %q, want the local function's declaration", got2)
	}

	text2 := "val m = {a: 1}\nprint(m.a)\n"
	s2, uri2 := openDoc(t, text2)
	if res := hoverAt(t, s2, uri2, text2, 1, "a"); res != nil {
		t.Errorf("hover over a bare index access returned %v, want nil", res)
	}
}

// A module member's hover shows its docstring above the signature, the way a
// builtin's help does, and never leaks compiler directives such as
// `std:this,__style` into the docs.
func TestHoverModuleMemberDocs(t *testing.T) {
	text := "import color\nprint(color.style)\n"
	s, uri := openDoc(t, text)

	got, ok := hoverValue(hoverAt(t, s, uri, text, 1, "style"))
	if !ok {
		t.Fatalf("hover over color.style returned nothing")
	}
	if strings.Contains(got, "std:this") {
		t.Errorf("hover leaked the help directive: %q", got)
	}
	if !strings.Contains(got, "takes a text style") {
		t.Errorf("hover = %q, want the member's own docs", got)
	}
	docAt := strings.Index(got, "takes a text style")
	codeAt := strings.Index(got, "```blue")
	if docAt < 0 || codeAt < 0 || docAt > codeAt {
		t.Errorf("hover = %q, want the docstring above the signature code sample", got)
	}
	if !strings.Contains(got, "fun style(text=normal, fg_color=normal, bg_color=normal)") {
		t.Errorf("hover = %q, want the function header as the signature", got)
	}
	if strings.Contains(got, "__style(text, fg_color, bg_color)") {
		t.Errorf("hover = %q, want the body left out of the signature", got)
	}

	if !strings.Contains(got, "Signature:  style(") || !strings.Contains(got, "Example(s):") {
		t.Errorf("hover = %q, want the wrapped builtin's help text", got)
	}

	text2 := "import math\nprint(math.rand)\n"
	s2, uri2 := openDoc(t, text2)
	got2, ok := hoverValue(hoverAt(t, s2, uri2, text2, 1, "rand"))
	if !ok {
		t.Fatalf("hover over math.rand returned nothing")
	}
	if !strings.Contains(got2, "`rand` returns a random float") || !strings.Contains(got2, "rand() -> float") {
		t.Errorf("hover = %q, want the member's docs with their signature", got2)
	}
	if strings.Contains(got2, "__rand()") {
		t.Errorf("hover = %q, want the body left out", got2)
	}
	if strings.Index(got2, "`rand` returns") > strings.Index(got2, "```blue") {
		t.Errorf("hover = %q, want the docstring above the signature code sample", got2)
	}
}

// The module name in an import statement is explained by the docs the module
// writes about itself at the top of its source, and by nothing else.
func TestHoverImportShowsModuleDocs(t *testing.T) {
	text := "import math\nimport color\nimport time as clock\nprint(math.pi)\nprint(clock.now())\n"
	s, uri := openDoc(t, text)

	got, ok := hoverValue(hoverAt(t, s, uri, text, 0, "math"))
	if !ok {
		t.Fatalf("hover over `import math` returned nothing")
	}
	if !strings.Contains(got, "**module** `math`") {
		t.Errorf("hover = %q, want it labelled as the math module", got)
	}
	if !strings.Contains(got, "deals with most math related") {
		t.Errorf("hover = %q, want the module's own docs", got)
	}
	// Only module level docs, no member signatures.
	if strings.Contains(got, "```blue") {
		t.Errorf("hover = %q, want docs alone without any member signature", got)
	}

	// A local module documents itself the same way a std one does.
	got2, ok := hoverValue(hoverAt(t, s, uri, text, 1, "color"))
	if !ok {
		t.Fatalf("hover over `import color` returned nothing")
	}
	if !strings.Contains(got2, "print to the console with colors") {
		t.Errorf("hover = %q, want the module's own docs", got2)
	}

	// An aliased import answers for the path and for the alias alike.
	for _, at := range []string{"time", "clock"} {
		got3, ok := hoverValue(hoverAt(t, s, uri, text, 2, at))
		if !ok {
			t.Fatalf("hover over %q in `import time as clock` returned nothing", at)
		}
		if !strings.Contains(got3, "**module** `time`") || !strings.Contains(got3, "time related functions") {
			t.Errorf("hover over %q = %q, want the time module's docs", at, got3)
		}
	}

	// Reaching for the module in code explains it the same way.
	got4, ok := hoverValue(hoverAt(t, s, uri, text, 4, "clock"))
	if !ok {
		t.Fatalf("hover over a module used in code returned nothing")
	}
	if !strings.Contains(got4, "time related functions") {
		t.Errorf("hover = %q, want the module's own docs", got4)
	}
}

// A member must keep its own docs: hovering `math.pi` is about pi, and the names
// inside `from x import {a, b}` are members of the module, not the module itself.
func TestHoverImportKeepsMemberDocs(t *testing.T) {
	text := "import math\nprint(math.pi)\nfrom color import {style}\nprint(style)\n"
	s, uri := openDoc(t, text)

	got, ok := hoverValue(hoverAt(t, s, uri, text, 1, "pi"))
	if !ok {
		t.Fatalf("hover over math.pi returned nothing")
	}
	if strings.Contains(got, "deals with most math related") {
		t.Errorf("hover over math.pi = %q, want pi's docs rather than the module's", got)
	}
	if !strings.Contains(got, "val pi =") {
		t.Errorf("hover over math.pi = %q, want pi's declaration", got)
	}

	// `style` is not bound in this buffer, so it gets no hover, but the module
	// docs of its import statement must not stand in for it either.
	if res := hoverAt(t, s, uri, text, 3, "style"); res != nil {
		t.Errorf("hover over an imported name = %v, want nil", res)
	}
	if got, ok := hoverValue(hoverAt(t, s, uri, text, 2, "color")); !ok ||
		!strings.Contains(got, "print to the console with colors") {
		t.Errorf("hover over `from color import` = %q, want the color module's docs", got)
	}
}

// A local module documents itself the same way a std one does, and every segment
// of a dotted import names it, so hovering any of them answers the same way.
func TestHoverImportLocalModuleDocs(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "pkg"), 0o755); err != nil {
		t.Fatal(err)
	}
	module := "## `pkg.tool` does tool things\n\n## and more about it\n\nval x = 1\n"
	if err := os.WriteFile(filepath.Join(dir, "pkg", "tool.b"), []byte(module), 0o600); err != nil {
		t.Fatal(err)
	}

	text := strings.Join([]string{
		"import pkg.tool",
		"import pkg.tool as t2",
		"from pkg.tool import {x}",
		"print(t2.x)",
	}, "\n")
	s, uri := openDocIn(t, dir, "main.b", text)

	// Any segment of the path names the module, including the last one, which
	// looks like a `module.member` reference because of the dot before it.
	for _, at := range []string{"pkg", "tool"} {
		got, ok := hoverValue(hoverAt(t, s, uri, text, 0, at))
		if !ok {
			t.Fatalf("hover over %q in `import pkg.tool` returned nothing", at)
		}
		if !strings.Contains(got, "**module** `pkg.tool`") || !strings.Contains(got, "does tool things") {
			t.Errorf("hover over %q = %q, want the module's own docs", at, got)
		}
	}

	// The alias and a `from` import answer the same way.
	got, ok := hoverValue(hoverAt(t, s, uri, text, 1, "t2"))
	if !ok || !strings.Contains(got, "does tool things") {
		t.Errorf("hover over the alias = %q, want the module's own docs", got)
	}
	got2, ok := hoverValue(hoverAt(t, s, uri, text, 2, "tool"))
	if !ok || !strings.Contains(got2, "does tool things") {
		t.Errorf("hover over `from pkg.tool import` = %q, want the module's own docs", got2)
	}
	got3, ok := hoverValue(hoverAt(t, s, uri, text, 3, "t2"))
	if !ok || !strings.Contains(got3, "and more about it") {
		t.Errorf("hover over a module used in code = %q, want the module's own docs", got3)
	}
}

// A module that documents itself nowhere has nothing to show, and a name the
// buffer binds itself keeps explaining itself through its own declaration.
func TestHoverModuleWithoutDocs(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bare.b"), []byte("val x = 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	text := "import bare\nprint(bare.x)\n"
	s, uri := openDocIn(t, dir, "main.b", text)
	if res := hoverAt(t, s, uri, text, 0, "bare"); res != nil {
		t.Errorf("hover over a module with no docs = %v, want nil", res)
	}

	// A name this buffer binds itself is not a module, even when an import
	// happens to share the name, so its own declaration keeps explaining it.
	text2 := "import math\nval math = 1\nprint(math)\n"
	s2, uri2 := openDoc(t, text2)
	got, ok := hoverValue(hoverAt(t, s2, uri2, text2, 1, "math"))
	if !ok || !strings.Contains(got, "val math = 1") {
		t.Errorf("hover over a locally bound math = %q, want its own declaration", got)
	}
}

// Nothing may be invented: unknown names and members blue does not have get no
// hover at all.
func TestHoverUnknown(t *testing.T) {
	text := "import math\nprint(nope)\nprint(math.PI)\n"
	s, uri := openDoc(t, text)

	if res := hoverAt(t, s, uri, text, 1, "nope"); res != nil {
		t.Errorf("hover over an undeclared name returned %v, want nil", res)
	}

	// math.b defines pi in lower case only, so hovering PI must find nothing.
	if res := hoverAt(t, s, uri, text, 2, "PI"); res != nil {
		t.Errorf("hover over math.PI returned %v, want nil because blue has no PI", res)
	}
}
