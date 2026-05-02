package lib

import (
	"strings"
	"testing"
)

func TestCoreFileNotEmpty(t *testing.T) {
	if CoreFile == "" {
		t.Fatal("CoreFile should not be empty")
	}
}

func TestCoreFileContainsFunctions(t *testing.T) {
	if !strings.Contains(CoreFile, "fun") {
		t.Error("CoreFile should contain function definitions")
	}
}

func TestCoreFileContainsDocComments(t *testing.T) {
	if !strings.Contains(CoreFile, "##") {
		t.Error("CoreFile should contain doc comments")
	}
}

func TestReadStdFileToString(t *testing.T) {
	tests := []string{
		"math.b",
		"http.b",
		"crypto.b",
		"db.b",
		"time.b",
		"net.b",
		"csv.b",
		"config.b",
		"color.b",
		"psutil.b",
		"search.b",
		"wasm.b",
		"gg.b",
		"gg-static.b",
		"ui.b",
		"ui-static.b",
	}

	for _, fname := range tests {
		t.Run(fname, func(t *testing.T) {
			content := ReadStdFileToString(fname)
			if content == "" {
				t.Errorf("ReadStdFileToString(%q) returned empty string", fname)
			}
		})
	}
}

func TestReadStdFileToStringInvalidFile(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("ReadStdFileToString should panic on invalid file")
		}
	}()
	ReadStdFileToString("nonexistent.b")
}

func TestReadStdFileToStringMath(t *testing.T) {
	content := ReadStdFileToString("math.b")
	if content == "" {
		t.Fatal("math.b should not be empty")
	}
	if !strings.Contains(content, "fun") {
		t.Error("math.b should contain function definitions")
	}
}

func TestReadStdFileToStringHttp(t *testing.T) {
	content := ReadStdFileToString("http.b")
	if content == "" {
		t.Fatal("http.b should not be empty")
	}
}

func TestReadStdFileToStringCrypto(t *testing.T) {
	content := ReadStdFileToString("crypto.b")
	if content == "" {
		t.Fatal("crypto.b should not be empty")
	}
}

func TestReadStdFileToStringDb(t *testing.T) {
	content := ReadStdFileToString("db.b")
	if content == "" {
		t.Fatal("db.b should not be empty")
	}
}

func TestReadStdFileToStringTime(t *testing.T) {
	content := ReadStdFileToString("time.b")
	if content == "" {
		t.Fatal("time.b should not be empty")
	}
}

func TestReadStdFileToStringNet(t *testing.T) {
	content := ReadStdFileToString("net.b")
	if content == "" {
		t.Fatal("net.b should not be empty")
	}
}

func TestReadStdFileToStringCsv(t *testing.T) {
	content := ReadStdFileToString("csv.b")
	if content == "" {
		t.Fatal("csv.b should not be empty")
	}
}

func TestReadStdFileToStringConfig(t *testing.T) {
	content := ReadStdFileToString("config.b")
	if content == "" {
		t.Fatal("config.b should not be empty")
	}
}

func TestReadStdFileToStringColor(t *testing.T) {
	content := ReadStdFileToString("color.b")
	if content == "" {
		t.Fatal("color.b should not be empty")
	}
}

func TestReadStdFileToStringPsutil(t *testing.T) {
	content := ReadStdFileToString("psutil.b")
	if content == "" {
		t.Fatal("psutil.b should not be empty")
	}
}

func TestReadStdFileToStringSearch(t *testing.T) {
	content := ReadStdFileToString("search.b")
	if content == "" {
		t.Fatal("search.b should not be empty")
	}
}

func TestReadStdFileToStringWasm(t *testing.T) {
	content := ReadStdFileToString("wasm.b")
	if content == "" {
		t.Fatal("wasm.b should not be empty")
	}
}

func TestReadStdFileToStringGg(t *testing.T) {
	content := ReadStdFileToString("gg.b")
	if content == "" {
		t.Fatal("gg.b should not be empty")
	}
}

func TestReadStdFileToStringGgStatic(t *testing.T) {
	content := ReadStdFileToString("gg-static.b")
	if content == "" {
		t.Fatal("gg-static.b should not be empty")
	}
}

func TestReadStdFileToStringUi(t *testing.T) {
	content := ReadStdFileToString("ui.b")
	if content == "" {
		t.Fatal("ui.b should not be empty")
	}
}

func TestReadStdFileToStringUiStatic(t *testing.T) {
	content := ReadStdFileToString("ui-static.b")
	if content == "" {
		t.Fatal("ui-static.b should not be empty")
	}
}

func TestReadStdFileToStringWithSubdir(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("ReadStdFileToString should panic on file in subdirectory")
		}
	}()
	ReadStdFileToString("core/core.b")
}
