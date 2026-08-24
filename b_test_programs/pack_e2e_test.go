package b_program_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// End-to-end pack test (Phase 6): compile a demo program, append it to a
// freshly built bluerun template, run the packed executable and compare its
// output against running the same program through `blue vm`.
//
// Skipped in -short mode because it shells out to `go build` twice.
func TestPackProducesWorkingSingleExecutable(t *testing.T) {
	if testing.Short() {
		t.Skip("pack end-to-end skipped in -short mode")
	}
	bin := buildBlueBinary(t)
	bluerun := buildBluerunBinary(t)

	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "demo.b")
	demo := `fun makeAdder(a) {
    return fun(b) {
        return a + b
    }
}
val add5 = makeAdder(5)
var results = [add5(3), add5(10)]
try {
    var x = 1 / 0
    println(x)
} catch (e) {
    results.push("caught")
} finally {
    results.push("done")
}
val m = {"nums": [1, 2, 3], "big": 123456789012345678901234567890n}
println(results)
println(m.nums)
println(m.big)
assert(results.len() == 4)
assert(results[1] == 15)
`
	if err := os.WriteFile(src, []byte(demo), 0o644); err != nil {
		t.Fatal(err)
	}

	// Must carry .exe on Windows: the packed file is executed directly by
	// this test, and os/exec refuses extensionless absolute paths there.
	packed := filepath.Join(tmpDir, "myapp"+exeSuffix())
	if out, code := runCmd(bin, "pack", "--go-build", "-o", packed, src); code != 0 {
		t.Fatalf("pack failed (exit %d):\n%s", code, out)
	}
	info, err := os.Stat(packed)
	if err != nil {
		t.Fatalf("packed executable missing: %v", err)
	}
	// Windows stat never reports execute bits (regular files are 0666 or
	// 0444); executability there comes from the PE format and is proven
	// below by actually running packed.
	if runtime.GOOS != "windows" && info.Mode()&0o111 == 0 {
		t.Fatalf("packed executable is not executable: %v", info.Mode())
	}

	packedOut, packedCode := runCmd(packed)
	// The default `blue prog.b` path does not print the last expression's
	// value, and neither does the packed runner; compare like-for-like.
	vmOut, vmCode := runCmd(bin, src)

	if packedCode != vmCode {
		t.Fatalf("exit code mismatch: packed=%d vm=%d\nsource stdout:\n%s\npacked stdout:\n%s",
			packedCode, vmCode, vmOut, packedOut)
	}
	if packedOut != vmOut {
		t.Fatalf("output mismatch:\n--- packed ---\n%s\n--- vm ---\n%s", packedOut, vmOut)
	}
	if !strings.Contains(packedOut, "[8, 15, caught, done]") {
		t.Fatalf("unexpected program output: %q", packedOut)
	}

	// Running with no embedded payload must print an actionable error.
	noPayloadOut, noPayloadCode := runCmd(bluerun)
	if noPayloadCode == 0 || !strings.Contains(noPayloadOut, "no compiled blue program found") {
		t.Fatalf("bare bluerun should fail with usage help, got exit %d:\n%s", noPayloadCode, noPayloadOut)
	}
}

var (
	bluerunOnce sync.Once
	bluerunPath string
	bluerunErr  error
)

func buildBluerunBinary(t *testing.T) string {
	t.Helper()
	bluerunOnce.Do(func() {
		dir, err := os.MkdirTemp("", "blue-golden-bluerun")
		if err != nil {
			bluerunErr = err
			return
		}
		bin := filepath.Join(dir, "bluerun-golden"+exeSuffix())
		// Merge the ambient GOFLAGS tags (flavors like rgfw are often set
		// machine-wide, see README) with the required minivm tag. A
		// command-line -tags REPLACES GOFLAGS tags, so both must be
		// passed together.
		envTags := ""
		if out, err := exec.Command("go", "env", "GOFLAGS").Output(); err == nil {
			for _, flag := range strings.Fields(string(out)) {
				if v, ok := strings.CutPrefix(flag, "-tags="); ok {
					envTags = v
				}
			}
		}
		tags := "minivm"
		if envTags != "" {
			tags = tags + "," + envTags
		}
		if runTags := runningBuildTags(); runTags != "" {
			tags = tags + "," + runTags
		}
		out, err := exec.Command("go", "build", "-ldflags=-s -w", "-tags", tags, "-o", bin, "../cmd/bluerun").CombinedOutput()
		if err != nil {
			bluerunErr = err
			t.Logf("go build output:\n%s", out)
			return
		}
		bluerunPath = bin
	})
	if bluerunErr != nil {
		t.Fatalf("failed to build bluerun: %v", bluerunErr)
	}
	return bluerunPath
}
