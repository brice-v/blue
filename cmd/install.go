package cmd

import (
	"blue/cmd/srcbundle"
	"blue/consts"
	"crypto/sha256"
	"encoding/hex"
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
			return fmt.Errorf("no embedded source archive present (rebuild with ./make_src_bundle.sh first)")
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
			if err := srcbundle.WriteArchive(installedArchivePath(root)); err != nil {
				return fmt.Errorf("record source archive: %w", err)
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

// installedArchivePath returns where install keeps a copy of the
// embedded source archive to detect updates on later runs.
func installedArchivePath(root string) string {
	return filepath.Join(root, "blue-src.tar.gz")
}

// storedArchiveHash returns the hex SHA-256 of the previously installed
// archive, or empty when no record exists.
func storedArchiveHash(root string) string {
	data, err := os.ReadFile(installedArchivePath(root))
	if err != nil || len(data) == 0 {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// srcUpToDate reports whether srcDir was installed from the currently
// embedded archive. Blue and blues builds from the same commit embed
// the same archive, so they share one source tree.
func srcUpToDate(root, srcDir string) bool {
	stored := storedArchiveHash(root)
	if stored == "" || stored != srcbundle.ArchiveSHA256() {
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

// handleInstallCommand parses `blue install` flags and runs the install.
func handleInstallCommand(argc int, arguments []string) {
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
				consts.ErrorPrinter("`install` flag --prefix requires a directory\n")
				os.Exit(1)
			}
			opts.Prefix = arguments[i+2]
		default:
			if arg == opts.Prefix && opts.Prefix != "" {
				continue
			}
			consts.ErrorPrinter("unexpected `install` argument. got=%s\n", arg)
			os.Exit(1)
		}
	}
	if err := RunInstall(opts); err != nil {
		consts.ErrorPrinter("install failed: %s\n", err.Error())
		os.Exit(1)
	}
}
