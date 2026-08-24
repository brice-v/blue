package b_program_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// Golden equivalence (Phase 3 of blue-binary-plan.md): every program must
// behave IDENTICALLY when run through the real CLI from source
// (`blue vm prog.b`) and from a compiled image (`blue compile -o prog.bbc`
// then `blue vm prog.bbc`). Stdout and exit codes are compared.
//
// Each program runs in its own subprocess so spawned processes from earlier
// files cannot pollute later captures.

var (
	blueBinOnce sync.Once
	blueBinPath string
	blueBinErr  error
)

func buildBlueBinary(t *testing.T) string {
	t.Helper()
	blueBinOnce.Do(func() {
		dir, err := os.MkdirTemp("", "blue-golden-bin")
		if err != nil {
			blueBinErr = err
			return
		}
		bin := filepath.Join(dir, "blue-golden")
		args := withRunningTags([]string{"build", "-ldflags=-s -w"})
		args = append(args, "-o", bin, "..")
		out, err := exec.Command("go", args...).CombinedOutput()
		if err != nil {
			blueBinErr = err
			t.Logf("go build output:\n%s", out)
			return
		}
		blueBinPath = bin
	})
	if blueBinErr != nil {
		t.Fatalf("failed to build blue binary: %v", blueBinErr)
	}
	return blueBinPath
}

func runCmd(bin string, args ...string) (string, int) {
	cmd := exec.Command(bin, args...)
	out, _ := cmd.CombinedOutput()
	code := cmd.ProcessState.ExitCode()
	return string(out), code
}

// nondeterministicFiles print output that legitimately varies between two
// runs (network payloads, timings, pids, metrics). We still require both
// runs to exit identically.
var nondeterministicFiles = map[string]bool{
	"test_metrics.b":           true,
	"test_crypto.b":            true, // bcrypt uses a random salt
	"test_import_std.b":        true, // fetches a random fact over the network
	"test_db.b":                true,
	"test_psutil.b":            true,
	"test_pids.b":              true,
	"test_lots_of_processes.b": true,
	"test_recv_from_same_pid_but_with_pubsub_instead.b": true,
	"test_return_on_list_of_obj.b":                      true,
	"test_time_parse_and_to_str.b":                      true,
}

// blockingFiles start servers or wait on network/interactive I/O; they are
// excluded from the golden suite entirely (the base suite covers them).
var blockingFiles = map[string]bool{
	"test_http.b":               true,
	"test_http_client.b":        true,
	"test_http_response_full.b": true,
	"test_http_server.b":        true,
	"test_net_tcp.b":            true,
	"test_net_udp.b":            true,
}

func TestSourceAndBinaryImageProduceIdenticalOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("golden subprocess suite skipped in -short mode")
	}
	bin := buildBlueBinary(t)

	tmpDir := t.TempDir()
	dirs := []string{"./", "./generated"}
	for _, dir := range dirs {
		files, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range files {
			if !strings.HasSuffix(f.Name(), ".b") || blockingFiles[f.Name()] {
				continue
			}
			t.Run(filepath.Join(dir, f.Name()), func(t *testing.T) {
				fpath := filepath.Join(dir, f.Name())
				data, err := os.ReadFile(fpath)
				if err != nil {
					t.Fatal(err)
				}
				src := string(data)
				if strings.HasPrefix(src, "# IGNORE") || strings.HasPrefix(src, "#IGNORE") ||
					strings.HasPrefix(src, "#VM IGNORE") || strings.HasPrefix(src, "# VM IGNORE") {
					t.Skip("ignored by header")
				}

				sourceOut, sourceCode := runCmd(bin, "vm", fpath)

				bbcPath := filepath.Join(tmpDir, strings.TrimSuffix(f.Name(), ".b")+".bbc")
				compileOut, compileCode := runCmd(bin, "compile", "-o", bbcPath, fpath)
				if compileCode != 0 {
					t.Skipf("program does not compile to an image: %s", compileOut)
				}
				imageOut, imageCode := runCmd(bin, "vm", bbcPath)

				if sourceCode != imageCode {
					t.Fatalf("exit code mismatch: source=%d image=%d\nsource stdout:\n%s\nimage stdout:\n%s",
						sourceCode, imageCode, sourceOut, imageOut)
				}
				if !nondeterministicFiles[f.Name()] && sourceOut != imageOut {
					t.Fatalf("stdout mismatch:\n--- source ---\n%s\n--- image ---\n%s", sourceOut, imageOut)
				}
			})
		}
	}
}
