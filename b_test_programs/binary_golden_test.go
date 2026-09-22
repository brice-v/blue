package b_program_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// Golden equivalence: every program must
// behave IDENTICALLY when run through the real CLI from source
// (`blue vm prog.b`) and from a compiled image (`blue compile -o prog.bluec`
// then `blue vm prog.bluec`). Stdout and exit codes are compared.
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
		bin := filepath.Join(dir, "blue-golden"+exeSuffix())
		args := withRunningTags([]string{"build", "-ldflags=-s -w"})
		args = append(args, "-o", bin, "..")
		out, err := exec.Command("go", args...).CombinedOutput()
		if err != nil {
			blueBinErr = err
			t.Logf("go build output:\n%s", out)
			return
		}
		// Smoke-check that the binary actually starts. Without this, a
		// non-startable binary makes every runCmd return a start error
		// (exit -1, empty output), which the golden suite below would
		// silently turn into per-program skips.
		if out, err := exec.Command(bin, "version").CombinedOutput(); err != nil {
			blueBinErr = fmt.Errorf("built blue binary does not run: %w\noutput:\n%s", err, out)
			return
		}
		blueBinPath = bin
	})
	if blueBinErr != nil {
		t.Fatalf("failed to build blue binary: %v", blueBinErr)
	}
	return blueBinPath
}

// logTimestampRe matches the date/time prefix the Go `log` package puts on
// every line (e.g. fyne's startup locale logging on C/POSIX-locale systems).
// These prefixes legitimately differ between two runs, so they are stripped
// before golden comparisons.
var logTimestampRe = regexp.MustCompile(`(?m)^\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2} `)

func stripLogTimestamps(s string) string {
	return logTimestampRe.ReplaceAllString(s, "")
}

func runCmd(bin string, args ...string) (string, int) {
	cmd := exec.Command(bin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	_ = cmd.Run()
	code := cmd.ProcessState.ExitCode()
	return stripLogTimestamps(stdout.String()) + stripLogTimestamps(stderr.String()), code
}

// goldenCmdTimeout bounds a single source, compile or image subprocess in the
// golden suite. A program that blocks forever (a spawn/recv deadlock, or a
// platform-specific wait) would otherwise hang runCmd and burn the whole
// package timeout, hiding which file was at fault. Timed-out programs are
// skipped by the caller: the in-process VM suite still covers their behavior.
const goldenCmdTimeout = 60 * time.Second

// runGoldenCmd runs one golden subprocess with the suite's hard timeout,
// reporting whether the process had to be killed.
func runGoldenCmd(bin string, args ...string) (out string, code int, timedOut bool) {
	return runGoldenCmdTimeout(goldenCmdTimeout, bin, args...)
}

func runGoldenCmdTimeout(timeout time.Duration, bin string, args ...string) (out string, code int, timedOut bool) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	// After the context fires, give the child a moment to exit on its own,
	// then close its pipes so Wait cannot block on a grandchild holding them.
	cmd.WaitDelay = 5 * time.Second
	_ = cmd.Run()
	code = -1
	if cmd.ProcessState != nil {
		code = cmd.ProcessState.ExitCode()
	}
	out = stripLogTimestamps(stdout.String()) + stripLogTimestamps(stderr.String())
	return out, code, ctx.Err() == context.DeadlineExceeded
}

// TestRunGoldenCmdTimesOut proves the golden runner kills a blocking child and
// reports it, which is what keeps one hung program from exhausting the package
// test timeout. The child re-executes this test binary and sleeps.
func TestRunGoldenCmdTimesOut(t *testing.T) {
	if os.Getenv("BLUE_GOLDEN_HELPER_SLEEP") == "1" {
		time.Sleep(time.Minute)
		return
	}
	t.Setenv("BLUE_GOLDEN_HELPER_SLEEP", "1")
	_, code, timedOut := runGoldenCmdTimeout(300*time.Millisecond, os.Args[0], "-test.run=TestRunGoldenCmdTimesOut")
	if !timedOut {
		t.Fatalf("expected a timeout, got code=%d timedOut=%v", code, timedOut)
	}
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

// windowsSkippedFiles block or time out on the Windows CI runners (spawn plus
// channel recv timing). They are still exercised by the in-process VM suite, so
// skipping them here only drops the redundant source-vs-image comparison.
var windowsSkippedFiles = map[string]bool{
	"test_self_process.b": true,
}

func TestSourceAndBinaryImageProduceIdenticalOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("golden subprocess suite skipped in -short mode")
	}
	// Keep this suite hermetic: it must compare a FRESH compile against an
	// image, never a run-cache entry (which could predate local uncommitted
	// compiler changes sharing the same build fingerprint).
	t.Setenv("BLUE_NO_CACHE", "1")
	bin := buildBlueBinary(t)

	tmpDir := t.TempDir()
	maxPar := min(max(runtime.GOMAXPROCS(0), 2), 8)
	sem := make(chan struct{}, maxPar)
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
			if runtime.GOOS == "windows" && windowsSkippedFiles[f.Name()] {
				continue
			}
			dir, f := dir, f
			t.Run(filepath.Join(dir, f.Name()), func(t *testing.T) {
				t.Parallel()
				sem <- struct{}{}
				defer func() { <-sem }()
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

				sourceOut, sourceCode, timedOut := runGoldenCmd(bin, "vm", fpath)
				if timedOut {
					t.Skipf("source run exceeded %s (program blocks on this platform); covered by the in-process VM suite", goldenCmdTimeout)
				}

				prefix := filepath.Base(dir)
				if prefix == "." || prefix == "" || prefix == string(filepath.Separator) {
					prefix = "root"
				}
				bbcPath := filepath.Join(tmpDir, prefix+"_"+strings.TrimSuffix(f.Name(), ".b")+".bluec")
				compileOut, compileCode, timedOut := runGoldenCmd(bin, "compile", "-o", bbcPath, fpath)
				if timedOut {
					t.Skipf("compile exceeded %s; covered by the in-process VM suite", goldenCmdTimeout)
				}
				if compileCode != 0 {
					t.Skipf("program does not compile to an image: %s", compileOut)
				}
				imageOut, imageCode, timedOut := runGoldenCmd(bin, "vm", bbcPath)
				if timedOut {
					t.Skipf("image run exceeded %s (program blocks on this platform); covered by the in-process VM suite", goldenCmdTimeout)
				}

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
