package lsp

import (
	"strings"
	"testing"
)

// buildTestIndex is a shortcut for indexing source text the way the server does.
func buildTestIndex(t *testing.T, text string) *fileIndex {
	t.Helper()
	src := newDocSource("test.b", text)
	index := buildIndex(src)
	index.resolveExtents()
	return index
}

func declNamed(index *fileIndex, name string) *declaration {
	for _, d := range index.decls {
		if d.name == name {
			return d
		}
	}
	return nil
}

// Extents must stay well formed because a reversed range makes editors drop the
// whole symbol response. The samples deliberately include half typed and invalid
// code, which is the normal state of a buffer while someone is typing.
func TestIndexExtentsAreNeverInverted(t *testing.T) {
	samples := []string{
		"fun main() {\n\tval x = 1\n\tfor i in 0..10 { val y = x + i }\n}\n",
		"val a = 1\nvar b = 2\n",
		"fun f(a, b = 2, c) {\n## inner doc\nval inner = 0\nreturn a + b + c + inner\n}\n",
		"try {\n\throw 1\n} catch (e) {\n\tprint(e)\n}\n",
		"while (true) { break }",
		"if x { 1 } else if y { 2 } else { 3 }",
		"fun unterminated(x {\nval after = 0\n",
		"val trailing =",
		"const z = 3",
		"fun g(): int { return 1 }",
		"for k, v in m { print(k) }",
		"try { f() } catch e { }",
		"",
		"\n\n   \n",
	}

	for _, sample := range samples {
		index := buildTestIndex(t, sample)
		for _, d := range index.decls {
			if d.nameStart > d.nameEnd {
				t.Errorf("source %q: decl %q name start %d after end %d", sample, d.name, d.nameStart, d.nameEnd)
			}
			if d.end >= 0 && d.end < d.nameStart {
				t.Errorf("source %q: decl %q ends (%d) before its name starts (%d)", sample, d.name, d.nameStart, d.end)
			}
			for _, other := range index.decls {
				if other == d || d.end < 0 || other.end < 0 {
					continue
				}
				if d.kind == declFunction && other.kind == declFunction &&
					d.start <= other.start && other.start < d.end && other.end > d.end {
					t.Errorf("source %q: nested function %q escapes %q", sample, other.name, d.name)
				}
			}
		}
	}
}

// Statements after an import must still be indexed, which is easy to break when
// the scanner jumps past the token that follows the statement.
func TestIndexKeepsStatementsAfterImports(t *testing.T) {
	index := buildTestIndex(t, "import color\nfoo.bar.baz()\nimport foo.bar as b\nval x = 1\nfrom abc import {hello, doSomething}\nother.call()\nfrom other import *\n")

	want := []string{"color", "foo.bar", "abc", "other"}
	if len(index.imports) != len(want) {
		t.Fatalf("got %d imports, want %d: %+v", len(index.imports), len(want), index.imports)
	}
	for i, path := range want {
		if index.imports[i].path != path {
			t.Errorf("import %d = %q, want %q", i, index.imports[i].path, path)
		}
	}
	if got := index.imports[1].alias; got != "b" {
		t.Errorf("`import foo.bar as b` alias = %q, want \"b\"", got)
	}
	if got := strings.Join(index.imports[2].names, ","); got != "hello,doSomething" {
		t.Errorf("`from abc import {...}` names = %q, want \"hello,doSomething\"", got)
	}
	if got := strings.Join(index.imports[3].names, ","); got != "*" {
		t.Errorf("`from other import *` names = %q, want \"*\"", got)
	}

	// `val` binds immutably in blue (the compiler sets immutable from the
	// statement kind), so it is reported as a constant while `var` is not.
	if decl := declNamed(index, "x"); decl == nil {
		t.Fatal("no declaration for x")
	} else if decl.kind != declConstant {
		t.Errorf("x kind = %v, want a constant", decl.kind)
	}
}

// Blue has no return types and parameters are plain names with optional default
// values, so the signature shown for a function is exactly its parameter list.
// Defaults containing colons and braces (used all over lib/std) must survive.
func TestIndexFunctionSignatures(t *testing.T) {
	index := buildTestIndex(t, `fun plain() {}
fun add(a, b = 2, c) { return a + b + c }
fun serve(addr_port="localhost:3001", use_tls=true, mounts={'.':'/'}) {}
`)

	cases := []struct {
		name   string
		detail string
		params []string
	}{
		{name: "plain", detail: "()", params: nil},
		{name: "add", detail: "(a, b = 2, c)", params: []string{"a", "b", "c"}},
		{name: "serve", detail: `(addr_port="localhost:3001", use_tls=true, mounts={'.':'/'})`, params: []string{"addr_port", "use_tls", "mounts"}},
	}

	for _, tc := range cases {
		decl := declNamed(index, tc.name)
		if decl == nil {
			t.Fatalf("no declaration for %s", tc.name)
		}
		if decl.detail != tc.detail {
			t.Errorf("%s detail = %q, want %q", tc.name, decl.detail, tc.detail)
		}
		if got := strings.Join(decl.params, ","); got != strings.Join(tc.params, ",") {
			t.Errorf("%s params = %q, want %q", tc.name, got, strings.Join(tc.params, ","))
		}
	}

	// A header whose parameter list never closes is skipped entirely (nothing
	// sensible can be said about its signature), but the rest still indexes.
	noClose := buildTestIndex(t, "fun nope(x {\nval after = 0\n")
	if decl := declNamed(noClose, "nope"); decl != nil {
		t.Errorf("declared %q even though its parameter list never closes", decl.name)
	}

	// A function whose body never closes reaches to the end of the buffer.
	openBody := buildTestIndex(t, "fun openBody(a, b) {\nval inner = a + b\n")
	fn := declNamed(openBody, "openBody")
	if fn == nil {
		t.Fatal("no declaration for openBody")
	}
	if fn.detail != "(a, b)" {
		t.Errorf("openBody detail = %q, want \"(a, b)\"", fn.detail)
	}
	if fn.end != len(openBody.src.runes) {
		t.Errorf("openBody end = %d, want the end of the buffer (%d)", fn.end, len(openBody.src.runes))
	}
	if inner := declNamed(openBody, "inner"); inner == nil {
		t.Error("no declaration for the val inside an unclosed body")
	}
}

// Destructuring binds every identifier inside the brackets or braces, and the
// value after `=` must not be mistaken for another binding.
func TestIndexDestructuringAndDefaults(t *testing.T) {
	index := buildTestIndex(t, "val [a, b] = pair\nvar {x, y} = point\nval scaled = a * 2\n")

	for _, want := range []string{"a", "b", "x", "y"} {
		if declNamed(index, want) == nil {
			t.Errorf("no declaration for destructured name %q", want)
		}
	}

	scaled := declNamed(index, "scaled")
	if scaled == nil {
		t.Fatal("no declaration for scaled")
	}
	if got := strings.Join(scaled.params, ","); got != "" {
		t.Errorf("scaled should have no params, got %q", got)
	}
	if strings.ContainsAny(scaled.doc, "*2") && scaled.doc != "" {
		t.Errorf("value expression leaked somewhere unexpected: %q", scaled.doc)
	}
}

// Docstrings are the `##` comment lines sitting above a declaration.
func TestIndexDocstrings(t *testing.T) {
	index := buildTestIndex(t, `## explains nothing much
var pi = 3.14

## first line
## second line
fun documented(x) {
	return x
}

val undoc = 1
`)

	if got := declNamed(index, "pi").doc; !strings.Contains(got, "explains nothing much") {
		t.Errorf("pi doc = %q, want it to contain \"explains nothing much\"", got)
	}

	fn := declNamed(index, "documented")
	if fn == nil {
		t.Fatal("no declaration for documented")
	}
	if !strings.Contains(fn.doc, "first line") || !strings.Contains(fn.doc, "second line") {
		t.Errorf("documented doc = %q, want both comment lines", fn.doc)
	}

	if got := declNamed(index, "undoc").doc; got != "" {
		t.Errorf("undoc doc = %q, want empty", got)
	}
}

// Definition lookup only sees bindings declared before the cursor, which is what
// makes hover and completion agree with blue's own scoping.
func TestIndexDefinitionFor(t *testing.T) {
	src := newDocSource("test.b", "fun outer(p) {\n\tval local = p * 2\n\treturn local\n}\n")
	index := buildTestIndex(t, src.text)

	if got := index.definitionFor("outer", 0); got == nil || got.kind != declFunction {
		t.Fatalf("definitionFor(outer) = %+v, want the function", got)
	}

	// The offset of `local` inside the body.
	localOff := strings.Index(src.text, "local")
	local := index.definitionFor("local", localOff)
	if local == nil || local.name != "local" {
		t.Fatalf("definitionFor(local) = %+v, want the val inside outer", local)
	}

	param := index.definitionFor("p", src.offsetOf(position{Line: 1, Character: 14}))
	if param == nil || param.kind != declParam {
		t.Fatalf("definitionFor(p) inside the body = %+v, want the parameter", param)
	}

	params := 0
	for _, d := range index.decls {
		if d.kind == declParam && d.name == "p" {
			params++
		}
	}
	if params != 1 {
		t.Errorf("got %d param declarations for p, want 1", params)
	}

	if got := index.definitionFor("neverDeclaredAnywhere", 40); got != nil {
		t.Errorf("definitionFor(unknown) = %+v, want nil", got)
	}
}
