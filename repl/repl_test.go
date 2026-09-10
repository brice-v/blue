package repl

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func scriptedReplInput(t *testing.T, lines string) io.ReadCloser {
	t.Helper()
	return io.NopCloser(strings.NewReader(lines))
}

func TestStartLexerReplEmitsTokensThenExitsOnEOF(t *testing.T) {
	var out bytes.Buffer
	if err := startLexerRepl(scriptedReplInput(t, "1 + 2\n"), &out, "tester"); err != nil {
		t.Fatalf("expected clean exit on EOF, got %v", err)
	}
	s := out.String()
	if !strings.Contains(s, "MODE: LEX") {
		t.Fatalf("missing banner in output: %s", s)
	}
	for _, want := range []string{"Type:INT", "Literal:+"} {
		if !strings.Contains(s, want) {
			t.Fatalf("expected %s token in output, got: %s", want, s)
		}
	}
}

func TestStartParserReplPrintsProgramThenExitsOnEOF(t *testing.T) {
	var out bytes.Buffer
	if err := startParserRepl(scriptedReplInput(t, "let x = 1\n"), &out, "tester"); err != nil {
		t.Fatalf("expected clean exit on EOF, got %v", err)
	}
	s := out.String()
	if !strings.Contains(s, "MODE: PARSE") || !strings.Contains(s, "x") {
		t.Fatalf("unexpected output: %s", s)
	}
}

func TestStartParserReplReportsParseErrorsWithoutExiting(t *testing.T) {
	var out bytes.Buffer
	if err := startParserRepl(scriptedReplInput(t, "let x =\nlet y = 2\n"), &out, "tester"); err != nil {
		t.Fatalf("expected clean exit on EOF, got %v", err)
	}
	s := out.String()
	if !strings.Contains(s, "y") {
		t.Fatalf("valid line after a bad one should still be parsed: %s", s)
	}
}

func TestStartVmReplEvaluatesAndStoresResultVars(t *testing.T) {
	var out bytes.Buffer
	if err := startVmRepl(scriptedReplInput(t, "1 + 2\n3 * 4\n"), &out, "tester", "", ""); err != nil {
		t.Fatalf("expected clean exit on EOF, got %v", err)
	}
	s := out.String()
	if !strings.Contains(s, "MODE: VM") {
		t.Fatalf("missing banner in output: %s", s)
	}
	for _, want := range []string{"_1 => 3", "_2 => 12"} {
		if !strings.Contains(s, want) {
			t.Fatalf("expected %q in output, got: %s", want, s)
		}
	}
}

func TestStartVmReplExitsOnDotExit(t *testing.T) {
	var out bytes.Buffer
	if err := startVmRepl(scriptedReplInput(t, "1 + 2\n.exit\nnever-evaluated"), &out, "tester", "", ""); err != nil {
		t.Fatalf("expected clean exit on .exit, got %v", err)
	}
	s := out.String()
	if strings.Contains(s, "never-evaluated") {
		t.Fatalf("input after .exit should not be processed: %s", s)
	}
}

func TestStartVmReplSurfacesRuntimeErrors(t *testing.T) {
	var out bytes.Buffer
	if err := startVmRepl(scriptedReplInput(t, "1 + true\n"), &out, "tester", "", ""); err != nil {
		t.Fatalf("runtime errors should not end the session, got %v", err)
	}
}

func TestStartVmReplKeepsSessionAfterCompilerErrors(t *testing.T) {
	var out bytes.Buffer
	if err := startVmRepl(scriptedReplInput(t, "let x =\n2 + 3\n"), &out, "tester", "", ""); err != nil {
		t.Fatalf("expected clean exit on EOF, got %v", err)
	}
	if !strings.Contains(out.String(), "_1 => 5") {
		t.Fatalf("valid line after a compiler error should still be evaluated: %s", out.String())
	}
}
