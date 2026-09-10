package cmd

import (
	"blue/cmd/srcbundle"
	"blue/consts"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// DefaultInstallRoot returns the default blue installation root.
// It is ~/.local/blue on Unix and %USERPROFILE%\.blue on Windows.
func DefaultInstallRoot() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(home, ".blue")
	}
	return filepath.Join(home, ".local", "blue")
}

// DefaultSrcDir returns the default directory the embedded source tree
// is extracted to.
func DefaultSrcDir() string {
	if root := DefaultInstallRoot(); root != "" {
		return filepath.Join(root, "src")
	}
	return ""
}

// DefaultBinDir returns the default directory the running executable
// is copied to.
func DefaultBinDir() string {
	if root := DefaultInstallRoot(); root != "" {
		return filepath.Join(root, "bin")
	}
	return ""
}

// defaultSourceCandidates returns source directories checked when
// BLUE_INSTALL_PATH is unset, newest layout first.
func defaultSourceCandidates() []string {
	var out []string
	if src := DefaultSrcDir(); src != "" {
		out = append(out, src)
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		legacy := filepath.Join(home, ".blue", "src")
		if len(out) == 0 || legacy != out[0] {
			out = append(out, legacy)
		}
	}
	return out
}

// goAvailable reports whether the go toolchain is on PATH.
func goAvailable() bool {
	_, err := exec.LookPath("go")
	return err == nil
}

// InstallOptions controls which install steps run.
type InstallOptions struct {
	Prefix string
	Force  bool
	NoSrc  bool
	NoBin  bool
}

// RunInstall extracts the embedded source, pre-fetches modules, and
// copies the running binary into the bin directory.
func RunInstall(opts InstallOptions) error {
	root := opts.Prefix
	if root == "" {
		root = DefaultInstallRoot()
	}
	if root == "" {
		return fmt.Errorf("cannot determine install root: set --prefix <dir>")
	}
	srcDir := filepath.Join(root, "src")
	binDir := filepath.Join(root, "bin")

	if !opts.NoSrc {
		if !goAvailable() {
			consts.ErrorPrinter("warning: go toolchain not found on PATH, source not installed and `blue bundle` will be unavailable\n")
		} else if !srcbundle.Available() {
			return fmt.Errorf("no embedded source tree in this build")
		} else if srcUpToDate(root, srcDir) && !opts.Force {
			fmt.Printf("source up to date at %s\n", srcDir)
		} else {
			if _, err := os.Stat(srcDir); err == nil {
				fmt.Printf("refreshing source at %s\n", srcDir)
				if err := os.RemoveAll(srcDir); err != nil {
					return err
				}
			}
			if err := os.MkdirAll(srcDir, 0o755); err != nil {
				return err
			}
			if err := srcbundle.Extract(srcDir); err != nil {
				return fmt.Errorf("extract source: %w", err)
			}
			if err := writeTreeHash(root); err != nil {
				return fmt.Errorf("record source tree: %w", err)
			}
			fmt.Printf("extracted source to %s\n", srcDir)
			if err := runGoModDownload(srcDir); err != nil {
				return err
			}
		}
	}

	if !opts.NoBin {
		if err := os.MkdirAll(binDir, 0o755); err != nil {
			return err
		}
		dest, err := installExecutable(binDir)
		if err != nil {
			return err
		}
		fmt.Printf("installed binary to %s\n", dest)
	}

	printInstallEnv(root, srcDir, binDir)
	return nil
}

// treeHashPath returns where install records the SHA-256 of the embedded source
// tree it extracted, so later runs can tell whether the running binary still
// carries the same source.
func treeHashPath(root string) string {
	return filepath.Join(root, ".blue-src-sha256")
}

// writeTreeHash records the hash of the embedded source tree that was just
// extracted to root/src.
func writeTreeHash(root string) error {
	return os.WriteFile(treeHashPath(root), []byte(srcbundle.TreeSHA256()+"\n"), 0o644)
}

// storedTreeHash returns the recorded hex SHA-256 of the installed source tree,
// or empty when no record exists.
func storedTreeHash(root string) string {
	data, err := os.ReadFile(treeHashPath(root))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// srcUpToDate reports whether srcDir was installed from the currently embedded
// source tree. Blue and blues builds from the same commit carry the same tree,
// so they share one source install.
func srcUpToDate(root, srcDir string) bool {
	stored := storedTreeHash(root)
	if stored == "" || stored != srcbundle.TreeSHA256() {
		return false
	}
	ok, _ := isBlueSourceDir(srcDir)
	return ok
}

// isBlueSourceDir reports whether dir looks like an installed source tree.
func isBlueSourceDir(dir string) (bool, error) {
	if !isFile(filepath.Join(dir, "go.mod")) {
		return false, nil
	}
	if !isDir(filepath.Join(dir, runnerPackageRelPath)) {
		return false, nil
	}
	return true, nil
}

// runGoModDownload pre-fetches modules so later bundles work offline.
func runGoModDownload(srcDir string) error {
	fmt.Printf("fetching go modules in %s (first run may take a while)...\n", srcDir)
	cmd := exec.Command("go", "mod", "download")
	cmd.Dir = srcDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod download: %w", err)
	}
	fmt.Println("go modules ready")
	return nil
}

// installExecutable copies the running binary into binDir, keeping its
// file name so `blues` installs as `blues` and `blue` as `blue`.
func installExecutable(binDir string) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	name := filepath.Base(exe)
	if runtime.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(name), ".exe") {
		name += ".exe"
	}
	dest := filepath.Join(binDir, name)
	if sameFile(exe, dest) {
		return dest, nil
	}
	if err := copyFile(exe, dest); err != nil {
		return "", err
	}
	return dest, nil
}

// sameFile reports whether two paths resolve to the same file.
func sameFile(a, b string) bool {
	ra, errA := filepath.EvalSymlinks(a)
	rb, errB := filepath.EvalSymlinks(b)
	if errA != nil || errB != nil {
		return a == b
	}
	return ra == rb
}

// copyFile copies src to dst preserving the executable bit.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = in.Close()
	}()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Chmod(dst, 0o755)
}

// printInstallEnv prints the shell setup for the new install.
func printInstallEnv(root, srcDir, binDir string) {
	fmt.Printf("\ninstall root: %s\n", root)
	fmt.Printf("add to PATH and point bundles at the source with:\n")
	if runtime.GOOS == "windows" {
		fmt.Printf("  setx BLUE_INSTALL_PATH \"%s\"\n", srcDir)
		fmt.Printf("  setx PATH \"%%PATH%%;%s\"\n", binDir)
		return
	}
	fmt.Printf("  export BLUE_INSTALL_PATH=%s\n", srcDir)
	fmt.Printf("  export PATH=%s:$PATH\n", binDir)
}

// parseInstallArgs extracts InstallOptions from `blue install` arguments.
func parseInstallArgs(argc int, arguments []string) (InstallOptions, error) {
	opts := InstallOptions{}
	for i, arg := range arguments[1:] {
		switch arg {
		case "--force", "-f":
			opts.Force = true
		case "--no-src":
			opts.NoSrc = true
		case "--no-bin":
			opts.NoBin = true
		case "--prefix":
			if i+2 >= argc {
				return InstallOptions{}, fmt.Errorf("`install` flag --prefix requires a directory")
			}
			opts.Prefix = arguments[i+2]
		default:
			if arg == opts.Prefix && opts.Prefix != "" {
				continue
			}
			return InstallOptions{}, fmt.Errorf("unexpected `install` argument. got=%s", arg)
		}
	}
	return opts, nil
}

// handleInstallCommand parses `blue install` flags and runs the install.
func handleInstallCommand(argc int, arguments []string) error {
	opts, err := parseInstallArgs(argc, arguments)
	if err != nil {
		return failf("%s", err.Error())
	}
	if err := RunInstall(opts); err != nil {
		return failf("install failed: %w", err)
	}
	return nil
}
