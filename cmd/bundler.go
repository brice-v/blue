package cmd

import (
	"blue/consts"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/uuid"
	"github.com/gookit/color"

	cp "github.com/otiai10/copy"
)

const header = `package main

import (
	"blue/evaluator"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"blue/repl"
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

var out = os.Stderr

//go:embed embed_files
var files embed.FS
`

const mainFunc = `func main() {
	entryPoint, err := files.ReadFile("embed_files/" + entryPointPath)
	if err != nil {
		out.WriteString("Failed to read EntryPoint File '" + entryPointPath + "'\n")
		os.Exit(1)
	}
	input := string(entryPoint)
	evaluator.IsEmbed = true
	evaluator.Files = files
	l := lexer.New(input, "<embed: "+entryPointPath+">")
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		repl.PrintParserErrors(out, p.Errors())
		os.Exit(1)
	}
	evaluator := evaluator.New()
	evaluator.CurrentFile = "<embed>"
	evaluator.EvalBasePath = filepath.Dir(".")
	val := evaluator.Eval(program)
	if val.Type() == object.ERROR_OBJ {
		errorObj := val.(*object.Error)
		var buf bytes.Buffer
		buf.WriteString(errorObj.Message)
		buf.WriteByte('\n')
		for evaluator.ErrorTokens.Len() > 0 {
			buf.WriteString(lexer.GetErrorLineMessage(evaluator.ErrorTokens.PopBack()))
			buf.WriteByte('\n')
		}
		out.WriteString(fmt.Sprintf("EvaluatorError: %s", buf.String()))
		os.Exit(1)
	}
}
`

// bundleFile takes the given file as an entry point
// and bundles the interpreter with the code into a go executable
func bundleFile(fpath string) error {
	entryPointPath := fmt.Sprintf("const entryPointPath = `%s`\n", fpath)
	gomain := fmt.Sprintf("%s\n%s\n%s", header, entryPointPath, mainFunc)
	defer color.Reset()

	// save current directory
	savedCurrentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("`savedCurrentDir` error: %s", err.Error())
	}
	consts.InfoPrinter("Saved Current Directory is `%s`\n", savedCurrentDir)

	// These steps need to executed in this order

	// change dir to tmp
	tmpDir, err := createTmpWorkspace()
	defer changeBackToSavedDir(savedCurrentDir)
	if err != nil {
		return fmt.Errorf("`createTmpWorkspace` error: %s", err.Error())
	}
	consts.InfoPrinter("Temporary Directory for Building Created at `%s`\n", tmpDir)
	// check BLUE_INSTALL_PATH is set with files, if not git clone them else error out
	err = checkAndGetBlueSouce()
	if err != nil {
		return fmt.Errorf("`checkAndGetBlueSource` error: %s", err.Error())
	}
	// Copy files from BLUE_INSTALL_PATH into current tmp dir
	err = copyFilesFromBlueSourceToTmpDir()
	if err != nil {
		return fmt.Errorf("`copyFilesFromBlueSourceToTmpDir` error: %s", err.Error())
	}
	consts.InfoPrinter("Copied Files from BLUE_INSTALL_PATH `%s` to Temporary Directory\n", os.Getenv(consts.BLUE_INSTALL_PATH))

	// make the directory for embedding the files
	err = makeEmbedFilesDir()
	if err != nil {
		return fmt.Errorf("`makeEmbedFilesDir` error: %s", err.Error())
	}
	consts.InfoPrinter("Made Embedded Files Directory with prefix `%s`\n", consts.EMBED_FILES_PREFIX)

	// Copy currentSavedDir files into embed_files dir, (we are in the tmpWorkspace (so /tmp/blue-build-xxx/embed_files))
	err = copyFilesFromDirToTmpDirEmbedFiles(savedCurrentDir, tmpDir)
	if err != nil {
		return fmt.Errorf("`copyFilesFromDirToTmpDirEmbedFiles` error: %s", err.Error())
	}
	consts.InfoPrinter("Copied Files from Directory `%s` to `%s`\n", savedCurrentDir, tmpDir)
	err = renameOriginalMainGoFile()
	if err != nil {
		return fmt.Errorf("`renameOriginalMainGoFile` error: %s", err.Error())
	}
	err = writeMainGoFile(gomain)
	if err != nil {
		return fmt.Errorf("`writeMainGoFile` error: %s", err.Error())
	}
	consts.InfoPrinter("Renamed Original Main Go File to bundler Main\n")
	consts.InfoPrinter("Building Exe with go toolchain...\n")
	err = buildExeAndWriteToSavedDir(fpath, tmpDir, savedCurrentDir)
	if err != nil {
		return fmt.Errorf("`buildExeAndWriteToSavedDir` error: %s", err.Error())
	}
	err = removeMainGoFile()
	if err != nil {
		return fmt.Errorf("`removeMainGoFile` error: %s", err.Error())
	}
	err = revertRenameOfOriginalGoFile()
	if err != nil {
		return fmt.Errorf("`revertRenameOfOriginalGoFile` error: %s", err.Error())
	}
	return nil
}

func changeBackToSavedDir(savedCurrentDir string) {
	if err := os.Chdir(savedCurrentDir); err != nil {
		consts.ErrorPrinter("`changeBackToSavedDir` error: %s\n", err.Error())
		os.Exit(1)
	}
}

func createTmpWorkspace() (string, error) {
	tmpDir := os.TempDir() + string(os.PathSeparator) + "blue-build-" + uuid.NewString()
	err := os.Mkdir(tmpDir, 0700)
	if err != nil {
		return "", err
	}
	return tmpDir, os.Chdir(tmpDir)
}

func checkAndGetBlueSouce() error {
	dirPath := os.Getenv(consts.BLUE_INSTALL_PATH)
	if !isDir(dirPath) {
		return errors.New("`BLUE_INSTALL_PATH` not set")
	}
	mainFpath := path.Clean(dirPath + string(filepath.Separator) + "main.go")
	if !isFile(mainFpath) {
		fmt.Printf("bundler::checkAndGetBlueSouce: filepath does not exist at path (%s), trying to clone from github...", mainFpath)
		return gitCloneFilesToBlueInstallPath(dirPath)
	}
	return nil
}

func gitCloneFilesToBlueInstallPath(dirPath string) error {
	cmd := []string{"git", "clone", "https://github.com/brice-v/blue.git", dirPath}
	if runtime.GOOS == "windows" {
		winArgs := []string{"/c"}
		winArgs = append(winArgs, cmd...)
		output, err := exec.Command("cmd", winArgs...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to exec `%s`. error: %s", strings.Join(winArgs, " "), err.Error())
		}
		if len(output) > 0 {
			return nil
		}
	} else {
		output, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to exec `%s`. error: %s", strings.Join(cmd, " "), err.Error())
		}
		if len(output) > 0 {
			return nil
		}
	}
	return nil
}

func makeEmbedFilesDir() error {
	dirName := strings.TrimRight(consts.EMBED_FILES_PREFIX, "/")
	return os.Mkdir(dirName, 0755)
}

func copyFilesFromBlueSourceToTmpDir() error {
	dirPath := os.Getenv(consts.BLUE_INSTALL_PATH)
	// copy these files into .
	return copyFilesFromDirToTmpDir(dirPath)
}

func copyFilesFromDirToTmpDir(dirPath string) error {
	return cp.Copy(dirPath, ".")
}

func copyFilesFromDirToTmpDirEmbedFiles(dirPath, tmpDir string) error {
	// Note: this should only be called from tmp workspace so the dir should already be created
	dstPath := path.Join(tmpDir, "embed_files")
	return cp.Copy(dirPath, dstPath)
}

func renameOriginalMainGoFile() error {
	err := os.Rename("main.go", "main.go.tmp")
	if err != nil {
		return fmt.Errorf("`main.go` rename failed to `main.go.tmp`. error: %s", err.Error())
	}
	return nil
}

func writeMainGoFile(fdata string) error {
	f, err := os.Create("main.go")
	if err != nil {
		return fmt.Errorf("failed to created `main.go` file. error: %s", err.Error())
	}
	_, err = f.WriteString(fdata)
	if err != nil {
		return fmt.Errorf("failed to write file data to `main.go` file. error: %s", err.Error())
	}
	err = f.Close()
	if err != nil {
		return fmt.Errorf("failed to close `main.go` file. error: %s", err.Error())
	}
	return nil
}

func buildExeAndWriteToSavedDir(fpath, tmpDir, savedCurrentDir string) error {
	exeName := strings.ReplaceAll(filepath.Base(fpath), ".b", "")
	extension := ""
	if runtime.GOOS == "windows" {
		extension = ".exe"
	}
	cmd := []string{"go", "build", "-o", exeName + extension}
	if runtime.GOOS == "windows" {
		winArgs := []string{"/c"}
		winArgs = append(winArgs, cmd...)
		command := exec.Command("cmd", winArgs...)
		command.Dir = tmpDir
		output, err := command.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to exec `%s`. error: %s", strings.Join(winArgs, " "), err.Error())
		}
		if len(output) == 0 {
			color.New(color.FgGreen, color.Bold).Printf("Successfully built `%s` as Executable!\n", cmd[len(cmd)-1])
			return copyFileToSavedDirAndPackIfPossible(exeName+extension, savedCurrentDir)
		}
	} else {
		command := exec.Command(cmd[0], cmd[1:]...)
		command.Dir = tmpDir
		output, err := command.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to exec `%s`. error: %s", strings.Join(cmd, " "), err.Error())
		}
		if len(output) == 0 {
			color.New(color.FgGreen, color.Bold).Printf("Successfully built `%s` as Executable!\n", cmd[len(cmd)-1])
			return copyFileToSavedDirAndPackIfPossible(exeName+extension, savedCurrentDir)
		}
	}
	panic("should not reach this line")
}

func copyFileToSavedDirAndPackIfPossible(exeName, savedCurrentDir string) error {
	src, err := os.Open(exeName)
	if err != nil {
		return err
	}
	defer src.Close()

	dstFile := savedCurrentDir + string(os.PathSeparator) + exeName
	dst, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	err = dst.Sync()
	if err != nil {
		return err
	}

	si, err := os.Stat(exeName)
	if err != nil {
		return err
	}

	err = os.Chmod(dstFile, si.Mode())
	if err != nil {
		return err
	}

	tryToPack(dstFile)

	return nil
}

func tryToPack(exeName string) {
	cmd := []string{"upx", exeName}
	var command *exec.Cmd
	if runtime.GOOS == "windows" {
		winArgs := []string{"/c"}
		winArgs = append(winArgs, cmd...)
		command = exec.Command("cmd", winArgs...)
	} else {
		command = exec.Command(cmd[0], cmd[1:]...)
	}
	consts.InfoPrinter("Trying to pack exe '%s' with UPX\n", exeName)
	command.CombinedOutput()
}

func removeMainGoFile() error {
	err := os.Remove("main.go")
	if err != nil {
		return fmt.Errorf("failed to remove `main.go` file. error: %s", err.Error())
	}
	return nil
}

func revertRenameOfOriginalGoFile() error {
	err := os.Rename("main.go.tmp", "main.go")
	if err != nil {
		return fmt.Errorf("`main.go.tmp` rename failed to `main.go`. error: %s", err.Error())
	}
	return nil
}
