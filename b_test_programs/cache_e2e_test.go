package b_program_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// End-to-end tests for the run cache (__blue_cache): running a program file
// caches its compiled image next to the program and later runs reuse it
// until the main source, any imported module, or the binary itself changes.
func setupCacheProject(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	main := filepath.Join(dir, "main.b")
	mod := filepath.Join(dir, "util.b")
	if err := os.WriteFile(mod, []byte("fun double(x) {\n    return x * 2\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(main, []byte("import util\nprintln(util.double(21))\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return main, mod
}

func cacheEntries(t *testing.T, main string) []os.DirEntry {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(filepath.Dir(main), "__blue_cache"))
	if err != nil {
		t.Fatalf("cache folder missing: %v", err)
	}
	return entries
}

func TestRunCacheReusesAndInvalidates(t *testing.T) {
	bin := buildBlueBinary(t)
	main, mod := setupCacheProject(t)

	out1, code1 := runCmd(bin, main)
	if code1 != 0 || !strings.Contains(out1, "42") {
		t.Fatalf("first run failed (exit %d):\n%s", code1, out1)
	}
	entries := cacheEntries(t, main)
	var images int
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".bluec") {
			images++
		}
	}
	if images == 0 {
		t.Fatalf("no cached images after first run: %d entries", len(entries))
	}

	out2, code2 := runCmd(bin, main)
	if code2 != 0 || out2 != out1 {
		t.Fatalf("cached run output differs:\n--- first ---\n%s\n--- cached ---\n%s", out1, out2)
	}

	// Editing an imported module must invalidate the entry.
	if err := os.WriteFile(mod, []byte("fun double(x) {\n    return x * 3\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out3, code3 := runCmd(bin, main)
	if code3 != 0 || !strings.Contains(out3, "63") {
		t.Fatalf("dependency change was not picked up (exit %d):\n%s", code3, out3)
	}

	// Editing the main source must produce a fresh entry.
	if err := os.WriteFile(main, []byte("import util\nval y = util.double(5)\nprintln(y + 1)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out4, code4 := runCmd(bin, main)
	if code4 != 0 || !strings.Contains(out4, "16") {
		t.Fatalf("main source change was not picked up (exit %d):\n%s", code4, out4)
	}
}

func TestRunCacheDisabledByEnv(t *testing.T) {
	bin := buildBlueBinary(t)
	main, _ := setupCacheProject(t)
	t.Setenv("BLUE_NO_CACHE", "1")

	out, code := runCmd(bin, main)
	if code != 0 || !strings.Contains(out, "42") {
		t.Fatalf("run with cache disabled failed (exit %d):\n%s", code, out)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(main), "__blue_cache")); !os.IsNotExist(err) {
		t.Fatalf("__blue_cache created despite BLUE_NO_CACHE")
	}
}

func TestRunCacheRecoversFromCorruption(t *testing.T) {
	bin := buildBlueBinary(t)
	main, _ := setupCacheProject(t)

	if out, code := runCmd(bin, main); code != 0 || !strings.Contains(out, "42") {
		t.Fatalf("priming run failed (exit %d):\n%s", code, out)
	}
	entries := cacheEntries(t, main)
	corrupted := false
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".bluec") {
			p := filepath.Join(filepath.Dir(main), "__blue_cache", e.Name())
			if err := os.WriteFile(p, []byte("garbage that is not an image"), 0o644); err != nil {
				t.Fatal(err)
			}
			corrupted = true
		}
	}
	if !corrupted {
		t.Fatal("found no image to corrupt")
	}

	out, code := runCmd(bin, main)
	if code != 0 || !strings.Contains(out, "42") {
		t.Fatalf("corrupted cache broke the run (exit %d):\n%s", code, out)
	}
}

func TestRunCacheWorksForDefaultCommandAndVm(t *testing.T) {
	bin := buildBlueBinary(t)
	main, _ := setupCacheProject(t)

	if out, code := runCmd(bin, main); code != 0 || !strings.Contains(out, "42") {
		t.Fatalf("default-command run failed (exit %d):\n%s", code, out)
	}
	if out, code := runCmd(bin, "vm", main); code != 0 || !strings.Contains(out, "42") {
		t.Fatalf("`vm` cached run failed (exit %d):\n%s", code, out)
	}
}
