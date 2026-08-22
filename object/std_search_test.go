package object

import (
	"regexp"
	"testing"
)

func searchBuiltinFn(t *testing.T, name string) BuiltinFunction {
	t.Helper()
	for _, b := range SearchBuiltins {
		if b.Name == name {
			if b.Fun == nil {
				t.Fatalf("search builtin %q has nil Fun", name)
			}
			return b.Fun
		}
	}
	t.Fatalf("search builtin %q not found", name)
	return nil
}

func TestSearchRegistry(t *testing.T) {
	seen := make(map[string]bool)
	for _, b := range SearchBuiltins {
		if b.Name == "" || b.HelpStr == "" {
			t.Fatalf("search builtin %q missing Name or HelpStr", b.Name)
		}
		if seen[b.Name] {
			t.Errorf("duplicate search builtin name %q", b.Name)
		}
		seen[b.Name] = true
		if b.Fun == nil {
			t.Errorf("search builtin %q has nil Fun", b.Name)
		}
	}
}

const testHtmlDoc = `<html><body><div id="abc">one</div><p>mid</p><div id="def">two</div></body></html>`

func TestByXpathBuiltin(t *testing.T) {
	fn := searchBuiltinFn(t, "_by_xpath")

	single := fn(&Stringo{Value: testHtmlDoc}, &Stringo{Value: `//*[@id="abc"]`}, TRUE)
	s, ok := single.(*Stringo)
	if !ok {
		t.Fatalf("by_xpath(find_one=true) returned %T", single)
	}
	if s.Value != `<div id="abc">one</div>` {
		t.Errorf("xpath single match = %q", s.Value)
	}

	all := fn(&Stringo{Value: testHtmlDoc}, &Stringo{Value: "//div"}, FALSE)
	l, ok := all.(*List)
	if !ok {
		t.Fatalf("by_xpath(find_one=false) returned %T", all)
	}
	if len(l.Elements) != 2 {
		t.Fatalf("xpath multi match found %d elements, want 2", len(l.Elements))
	}
	if l.Elements[0].(*Stringo).Value != `<div id="abc">one</div>` ||
		l.Elements[1].(*Stringo).Value != `<div id="def">two</div>` {
		t.Errorf("xpath multi match contents wrong: %q / %q",
			l.Elements[0].Inspect(), l.Elements[1].Inspect())
	}

	noMatch := fn(&Stringo{Value: testHtmlDoc}, &Stringo{Value: "//span"}, TRUE)
	if ns := noMatch.(*Stringo); ns.Value != "" {
		t.Errorf("xpath with no match should be empty STRING, got %q", ns.Value)
	}
	noMatchList := fn(&Stringo{Value: testHtmlDoc}, &Stringo{Value: "//span"}, FALSE)
	if nl := noMatchList.(*List); len(nl.Elements) != 0 {
		t.Errorf("xpath list with no match should be empty LIST, got %s", nl.Inspect())
	}

	tests := []builtinTestCase{
		{name: "empty doc", args: []Object{&Stringo{Value: ""}, &Stringo{Value: "//div"}, TRUE}, err: "str_to_search argument is empty"},
		{name: "empty query", args: []Object{&Stringo{Value: "<div/>"}, &Stringo{Value: ""}, TRUE}, err: "query argument is empty"},
		{name: "invalid xpath", args: []Object{&Stringo{Value: testHtmlDoc}, &Stringo{Value: "///[[["}, TRUE}, err: "`by_xpath` error"},
		{name: "doc not string", args: []Object{in(1), &Stringo{Value: "//div"}, TRUE}, err: "PositionalTypeError"},
		{name: "query not string", args: []Object{&Stringo{Value: "<div/>"}, in(1), TRUE}, err: "PositionalTypeError"},
		{name: "flag not bool", args: []Object{&Stringo{Value: "<div/>"}, &Stringo{Value: "//div"}, in(1)}, err: "PositionalTypeError"},
		{name: "two args", args: []Object{&Stringo{Value: "<div/>"}, &Stringo{Value: "//div"}}, err: "InvalidArgCountError"},
	}
	runBuiltinTestsFor(t, SearchBuiltins, "_by_xpath", tests)
}

func TestByRegexBuiltin(t *testing.T) {
	fn := searchBuiltinFn(t, "_by_regex")

	all := fn(&Stringo{Value: "abc 123 def 456"}, &Stringo{Value: "[0-9]+"}, FALSE)
	l, ok := all.(*List)
	if !ok {
		t.Fatalf("by_regex(find_one=false) returned %T", all)
	}
	if len(l.Elements) != 2 || l.Elements[0].(*Stringo).Value != "123" || l.Elements[1].(*Stringo).Value != "456" {
		t.Errorf("regex multi match wrong: %s", l.Inspect())
	}

	one := fn(&Stringo{Value: "abc 123 def 456"}, &Stringo{Value: "[0-9]+"}, TRUE)
	if o := one.(*Stringo); o.Value != "123" {
		t.Errorf("regex single match = %q, want 123", o.Value)
	}

	reObj := &Regex{Value: regexp.MustCompile(`[a-z]+`)}
	viaObj := fn(&Stringo{Value: "ABC xyz"}, reObj, TRUE)
	if v := viaObj.(*Stringo); v.Value != "xyz" {
		t.Errorf("regex object match = %q, want xyz", v.Value)
	}

	noMatch := fn(&Stringo{Value: "abc"}, &Stringo{Value: "[0-9]+"}, FALSE)
	if nm := noMatch.(*List); len(nm.Elements) != 0 {
		t.Errorf("regex no match should be empty LIST, got %s", nm.Inspect())
	}
	noMatchOne := fn(&Stringo{Value: "abc"}, &Stringo{Value: "[0-9]+"}, TRUE)
	if nmo := noMatchOne.(*Stringo); nmo.Value != "" {
		t.Errorf("regex no match should be empty STRING, got %q", nmo.Value)
	}

	tests := []builtinTestCase{
		{name: "invalid pattern", args: []Object{&Stringo{Value: "abc"}, &Stringo{Value: "("}, TRUE}, err: "failed to compile regexp"},
		{name: "target not string", args: []Object{in(1), &Stringo{Value: "a"}, TRUE}, err: "PositionalTypeError"},
		{name: "pattern wrong type", args: []Object{&Stringo{Value: "abc"}, fl(1), TRUE}, err: "PositionalTypeError"},
		{name: "flag not bool", args: []Object{&Stringo{Value: "abc"}, &Stringo{Value: "a"}, in(1)}, err: "PositionalTypeError"},
		{name: "two args", args: []Object{&Stringo{Value: "abc"}, &Stringo{Value: "a"}}, err: "InvalidArgCountError"},
	}
	runBuiltinTestsFor(t, SearchBuiltins, "_by_regex", tests)
}
