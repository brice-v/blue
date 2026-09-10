package cmd

import (
	"blue/consts"
	"errors"
	"path/filepath"
	"testing"
)

func TestRunVersion(t *testing.T) {
	if err := Run("blue", "version"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := Run("blue", "version", "--full"); err != nil {
		t.Fatalf("expected no error with --full, got %v", err)
	}
	if err := Run("blue", "version", "bogus"); err == nil {
		t.Fatal("expected error for unexpected version argument")
	}
}

func TestRunHelp(t *testing.T) {
	if err := Run("blue", "help"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRunUnknownCommandFallsBackToFileNotFound(t *testing.T) {
	err := Run("blue", "definitely-not-a-file.b")
	if err == nil {
		t.Fatal("expected error for missing file argument")
	}
}

func TestRunVmEvaluatesString(t *testing.T) {
	if err := Run("blue", "vm", "println(1 + 2)"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRunVmReportsCompileError(t *testing.T) {
	if err := Run("blue", "vm", "let x = "); err == nil {
		t.Fatal("expected error for source with parser errors")
	}
}

func TestRunVmReportsRuntimeError(t *testing.T) {
	err := Run("blue", "vm", "1 + true")
	if !errors.Is(err, ErrProgramFailed) {
		t.Fatalf("expected ErrProgramFailed, got %v", err)
	}
}

func TestRunVmUnexpectedArguments(t *testing.T) {
	if err := Run("blue", "vm", "a", "b", "c"); err == nil {
		t.Fatal("expected error for too many vm arguments")
	}
}

func TestRunLexCommand(t *testing.T) {
	fpath := writeTempProgram(t, "println(1)\n")
	if err := Run("blue", "lex", fpath); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := Run("blue", "lex", filepath.Join(t.TempDir(), "missing.b")); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestRunParseCommand(t *testing.T) {
	fpath := writeTempProgram(t, "let x = 1\n")
	if err := Run("blue", "parse", fpath); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := Run("blue", "parse", filepath.Join(t.TempDir(), "missing.b")); err == nil {
		t.Fatal("expected error for missing file")
	}
	if err := Run("blue", "parse", writeTempProgram(t, "let x = \n")); err == nil {
		t.Fatal("expected error for unparseable source")
	}
}

func TestRunCompileCommand(t *testing.T) {
	if err := Run("blue", "compile", "println(1)"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	outPath := filepath.Join(t.TempDir(), "out.bluec")
	if err := Run("blue", "compile", "-o", outPath, writeTempProgram(t, "println(1)\n")); err != nil {
		t.Fatalf("expected no error with -o, got %v", err)
	}
	if !isFile(outPath) {
		t.Fatal("compiled image was not written")
	}
	if err := Run("blue", "compile"); err == nil {
		t.Fatal("expected error for missing compile source")
	}
	if err := Run("blue", "compile", "-o"); err == nil {
		t.Fatal("expected error when -o is missing its path")
	}
}

func TestRunDocCommand(t *testing.T) {
	if err := Run("blue", "doc"); err == nil {
		t.Fatal("expected error for missing doc argument")
	}
}

func TestRunDefaultFallsBackToProgramFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(consts.BLUE_NO_CACHE, "1")
	t.Chdir(dir)
	if err := Run("blue", writeTempProgram(t, "println(42)\n")); err != nil {
		t.Fatalf("expected no error running a program file, got %v", err)
	}
}

func TestParseVersionArgs(t *testing.T) {
	full, err := parseVersionArgs([]string{"version"})
	if err != nil || full {
		t.Fatalf("unexpected result: full=%v err=%v", full, err)
	}
	full, err = parseVersionArgs([]string{"version", "--full"})
	if err != nil || !full {
		t.Fatalf("unexpected result: full=%v err=%v", full, err)
	}
	if _, err := parseVersionArgs([]string{"version", "bogus"}); err == nil {
		t.Fatal("expected error for unexpected argument")
	}
}

func TestParseInstallArgs(t *testing.T) {
	opts, err := parseInstallArgs(5, []string{"install", "--prefix", "/tmp/x", "-f", "--no-src"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Prefix != "/tmp/x" || !opts.Force || !opts.NoSrc {
		t.Fatalf("unexpected options: %+v", opts)
	}
	if _, err := parseInstallArgs(2, []string{"install", "--prefix"}); err == nil {
		t.Fatal("expected error when --prefix is missing its value")
	}
	if _, err := parseInstallArgs(3, []string{"install", "bogus"}); err == nil {
		t.Fatal("expected error for unexpected argument")
	}
}

func TestParseLspArgs(t *testing.T) {
	opts, err := parseLspArgs([]string{"--addr", "127.0.0.1:8080", "--trace"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Addr != "127.0.0.1:8080" || !opts.Trace {
		t.Fatalf("unexpected options: %+v", opts)
	}
	if _, err := parseLspArgs([]string{"--addr"}); err == nil {
		t.Fatal("expected error when --addr is missing its value")
	}
	if _, err := parseLspArgs([]string{"bogus"}); err == nil {
		t.Fatal("expected error for unexpected argument")
	}
}
