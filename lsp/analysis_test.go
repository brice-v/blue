package lsp

import (
	"fmt"
	"strings"
	"testing"
)

// posKey renders a range as a comparable identity so duplicates can be counted.
func posKey(r rangeStruct) string {
	return fmt.Sprintf("%d:%d-%d:%d", r.Start.Line, r.Start.Character, r.End.Line, r.End.Character)
}

// firstDiag finds the first diagnostic whose message contains want.
func firstDiag(diags []diagnostic, want string) (diagnostic, bool) {
	for _, d := range diags {
		if strings.Contains(d.Message, want) {
			return d, true
		}
	}
	return diagnostic{}, false
}

// highlight returns the source text a diagnostic covers, which is what an editor
// will underline.
func highlight(src *docSource, d diagnostic) string {
	line := src.lineRunes(d.Range.Start.Line)
	end := int(d.Range.End.Character)
	if end > len(line) {
		end = len(line)
	}
	if int(d.Range.Start.Character) > len(line) {
		return ""
	}
	return string(line[d.Range.Start.Character:end])
}

// Correct code has to stay silent. Every case here was checked against blue's own
// lexer first, so none of them rely on guesswork about the syntax.
func TestAnalyzeCleanBuffersProduceNothing(t *testing.T) {
	clean := []string{
		"",
		"\n\n   \n",
		"# just a comment\n",
		"### block comment ###\nval y = 2\n",
		"fun main() {\n\treturn 0\n}\n",
		"import math\nval pi = math.pi\nprint(pi)\n",
		"val s = \"a string with # and } inside\"\n",
		"val ex = `echo hi`\nprint(ex)\n",
		"val raw = \"\"\"raw \"text\" here\"\"\"\n",
		"fun f(a, b = 2, c) {\n\treturn a + b + c\n}\n",
		"for i in 0..10 { print(i) }\n",
	}

	for _, text := range clean {
		if diags := analyze(newDocSource("clean.b", text)); len(diags) != 0 {
			t.Errorf("source %q: got %d diagnostics, want none: %+v", text, len(diags), diags)
		}
	}
}

// A diagnostic has to cover the offending token and nothing more. The columns
// below were measured against blue's lexer rather than assumed, because its
// recorded position sometimes lands one place before or inside a token.
func TestAnalyzeRangesCoverTheOffendingToken(t *testing.T) {
	cases := []struct {
		text  string
		want  string // substring of the message to look for
		line  int
		from  int
		to    int
		token string // source text the range must cover exactly
	}{
		{
			text: "val x = 1\nval b = = 2\n",
			want: "unexpected =",
			line: 1, from: 8, to: 9,
			token: "=",
		},
		{
			text: "fun broken(x {\n  return x\n}\n",
			want: "expected ) got { instead",
			line: 0, from: 13, to: 14,
			token: "{",
		},
		{
			text: "val y = for\n",
			want: "unexpected for",
			line: 0, from: 8, to: 11,
			token: "for",
		},
		{
			text: "}\n",
			want: "unexpected }",
			line: 0, from: 0, to: 1,
			token: "}",
		},
	}

	for _, tc := range cases {
		src := newDocSource("t.b", tc.text)
		diags := analyze(src)
		d, ok := firstDiag(diags, tc.want)
		if !ok {
			t.Errorf("source %q: no diagnostic containing %q, got %+v", tc.text, tc.want, diags)
			continue
		}
		if d.Range.Start.Line != tc.line {
			t.Errorf("source %q: reported line %d, want %d", tc.text, d.Range.Start.Line, tc.line)
		}
		if int(d.Range.Start.Character) != tc.from || int(d.Range.End.Character) != tc.to {
			t.Errorf("source %q: range columns %d-%d, want %d-%d", tc.text,
				d.Range.Start.Character, d.Range.End.Character, tc.from, tc.to)
		}
		if got := highlight(src, d); got != tc.token {
			t.Errorf("source %q: highlighted %q, want exactly %q", tc.text, got, tc.token)
		}
	}
}

// Hints blue's parser attaches to certain mistakes ride along so the editor can
// show them without knowing anything about blue.
func TestAnalyzeCarriesParserHints(t *testing.T) {
	src := newDocSource("h.b", "}\n")
	d, ok := firstDiag(analyze(src), "unexpected }")
	if !ok {
		t.Fatal("no diagnostic for a stray '}'")
	}
	if !strings.Contains(d.Message, "Unmatched closing brace") {
		t.Errorf("message = %q, want the parser hint about an unmatched brace", d.Message)
	}
}

// Half typed literals are the normal state of a buffer being edited. They must
// produce sane ranges and never take the server down.
func TestAnalyzeSurvivesHalfTypedInput(t *testing.T) {
	partial := []string{
		"val s = \"unterminated\nprintln(s)\n",
		"val s = 'single not closed\nprint(1)\n",
		"val t = `exec never closed\n",
		"### unterminated block\n",
		"fun f(\n",
		"if (x) { } else {\n",
		"val m = {'a': 1\n",
		"val x = =",
		"}\n}\n}\n",
		"val 日本 = 5\nval z = 日\n",
	}

	for _, text := range partial {
		src := newDocSource("half.b", text)
		diags := analyze(src)
		for _, d := range diags {
			if strings.Contains(d.Message, "internal error") {
				t.Errorf("source %q produced an internal error: %s", text, d.Message)
			}
			if d.Severity != sevError {
				t.Errorf("source %q: severity = %d, want an error", text, d.Severity)
			}
			if d.Source != "blue" {
				t.Errorf("source %q: source = %q, want \"blue\"", text, d.Source)
			}
			if d.Code != "parse" {
				t.Errorf("source %q: code = %v, want \"parse\"", text, d.Code)
			}
			if d.Range.End.Line < d.Range.Start.Line {
				t.Errorf("source %q: range ends before it starts: %+v", text, d.Range)
			}
			if d.Range.End.Character <= d.Range.Start.Character {
				t.Errorf("source %q: empty range %+v", text, d.Range)
			}
			if hl := highlight(src, d); strings.Contains(hl, "\n") {
				t.Errorf("source %q: highlight crosses a line boundary: %q", text, hl)
			}
		}
	}
}

// Three identical mistakes on three lines each need their own marker, and the
// exact same message at the exact same spot must collapse into one.
func TestAnalyzeRepeatedAndDeduplicated(t *testing.T) {
	src := newDocSource("dupe.b", "val a = = 1\nval b = = 2\nval c = = 3\n")
	diags := analyze(src)

	seen := map[string]int{}
	marked := map[int]bool{}
	for _, d := range diags {
		key := fmt.Sprintf("%s@%s", strings.SplitN(d.Message, "\n", 2)[0], posKey(d.Range))
		seen[key]++
		if seen[key] > 1 {
			t.Errorf("identical diagnostic published twice: %s", key)
		}
		if strings.Contains(d.Message, "unexpected =") {
			marked[d.Range.Start.Line] = true
		}
	}

	for _, want := range []int{0, 1, 2} {
		if !marked[want] {
			t.Errorf("nothing reported on line %d of three identical mistakes", want)
		}
	}
}
