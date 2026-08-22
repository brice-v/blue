package object

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func configBuiltinFn(t *testing.T, name string) BuiltinFunction {
	t.Helper()
	for _, b := range ConfigBuiltins {
		if b.Name == name {
			if b.Fun == nil {
				t.Fatalf("config builtin %q has nil Fun", name)
			}
			return b.Fun
		}
	}
	t.Fatalf("config builtin %q not found", name)
	return nil
}

func TestConfigRegistry(t *testing.T) {
	seen := make(map[string]bool)
	for _, b := range ConfigBuiltins {
		if b.Name == "" || b.HelpStr == "" {
			t.Fatalf("config builtin %q missing Name or HelpStr", b.Name)
		}
		if seen[b.Name] {
			t.Errorf("duplicate config builtin name %q", b.Name)
		}
		seen[b.Name] = true
		if b.Fun == nil {
			t.Errorf("config builtin %q has nil Fun", b.Name)
		}
	}
}

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write %s: %v", p, err)
	}
	return p
}

func loadConfigAsMap(t *testing.T, path string) map[string]any {
	t.Helper()
	res := configBuiltinFn(t, "_load_file")(&Stringo{Value: path})
	s, ok := res.(*Stringo)
	if !ok {
		t.Fatalf("load_file(%s) returned %T: %v", path, res, res.Inspect())
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(s.Value), &m); err != nil {
		t.Fatalf("load_file(%s) did not return valid JSON (%q): %v", path, s.Value, err)
	}
	return m
}

func TestLoadFileBuiltin(t *testing.T) {
	tmp := t.TempDir()

	jsonPath := writeTempFile(t, tmp, "conf.json", `{"a": 1, "name": "blue"}`)
	m := loadConfigAsMap(t, jsonPath)
	if m["a"] != float64(1) || m["name"] != "blue" {
		t.Errorf("json config fields wrong: %v", m)
	}

	yamlPath := writeTempFile(t, tmp, "conf.yaml", "a: 1\nname: blue\n")
	m = loadConfigAsMap(t, yamlPath)
	if m["a"] != float64(1) || m["name"] != "blue" {
		t.Errorf("yaml config fields wrong: %v", m)
	}

	tomlPath := writeTempFile(t, tmp, "conf.toml", "a = 1\nname = 'blue'\n")
	m = loadConfigAsMap(t, tomlPath)
	if m["a"] != float64(1) || m["name"] != "blue" {
		t.Errorf("toml config fields wrong: %v", m)
	}

	iniPath := writeTempFile(t, tmp, "conf.ini", "a = 1\nname = blue\n")
	m = loadConfigAsMap(t, iniPath)
	if m["a"] == nil || m["name"] == nil {
		t.Errorf("ini config fields missing: %v", m)
	}

	propsPath := writeTempFile(t, tmp, "conf.properties", "a=1\nname=blue\n")
	m = loadConfigAsMap(t, propsPath)
	if m["a"] == nil || m["name"] == nil {
		t.Errorf("properties config fields missing: %v", m)
	}

	envPath := writeTempFile(t, tmp, "app.env", "BLUE_TEST_ENV_VAR=helloworld\n")
	res := configBuiltinFn(t, "_load_file")(&Stringo{Value: envPath})
	es, ok := res.(*Stringo)
	if !ok || es.Value != "{}" {
		t.Errorf(".env file should hit dotenv branch and return {}, got %#v", res)
	}
	if os.Getenv("BLUE_TEST_ENV_VAR") != "helloworld" {
		t.Error(".env file values were not loaded into the environment")
	}

	tests := []builtinTestCase{
		{name: "missing file", args: []Object{&Stringo{Value: filepath.Join(tmp, "nope.json")}}, err: "`load_file` error"},
		{name: "unsupported extension", args: []Object{&Stringo{Value: writeTempFile(t, tmp, "conf.xyz", "junk")}}, err: "`load_file` error"},
		{name: "malformed content", args: []Object{&Stringo{Value: writeTempFile(t, tmp, "bad.json", "{not json")}}, err: "`load_file` error"},
		{name: "path not string", args: []Object{in(1)}, err: "PositionalTypeError"},
		{name: "no args", args: []Object{}, err: "InvalidArgCountError"},
	}
	runBuiltinTestsFor(t, ConfigBuiltins, "_load_file", tests)
}

func TestDumpConfigBuiltin(t *testing.T) {
	tmp := t.TempDir()
	dumpFn := configBuiltinFn(t, "_dump_config")
	cfgJson := &Stringo{Value: `{"a": 1, "name": "blue"}`}

	formats := map[string]string{
		"JSON":       ".json",
		"YAML":       ".yaml",
		"TOML":       ".toml",
		"INI":        ".ini",
		"PROPERTIES": ".properties",
	}
	for format, ext := range formats {
		out := filepath.Join(tmp, "out"+ext)
		res := dumpFn(cfgJson, &Stringo{Value: out}, &Stringo{Value: format})
		if res != NULL {
			t.Fatalf("dump_config(%s) = %s, want NULL", format, res.Inspect())
		}
		data, err := os.ReadFile(out)
		if err != nil || len(data) == 0 {
			t.Errorf("dump_config(%s) produced no output file: %v", format, err)
			continue
		}
		backPath := filepath.Join(tmp, "round"+ext)
		if err := os.WriteFile(backPath, data, 0o600); err != nil {
			t.Fatal(err)
		}
		m := loadConfigAsMap(t, backPath)
		if fmt.Sprint(m["a"]) != "1" || fmt.Sprint(m["name"]) != "blue" {
			t.Errorf("dump_config(%s) round trip lost data: %v (file: %s)", format, m, data)
		}
	}

	badInput := dumpFn(&Stringo{Value: `{not json`}, &Stringo{Value: filepath.Join(tmp, "x.json")}, &Stringo{Value: "JSON"})
	if _, ok := badInput.(*Error); !ok {
		t.Errorf("dump_config with invalid JSON should error, got %T", badInput)
	}

	unknownFormat := dumpFn(cfgJson, &Stringo{Value: filepath.Join(tmp, "u.conf")}, &Stringo{Value: "XML"})
	errObj, ok := unknownFormat.(*Error)
	if !ok {
		t.Fatalf("dump_config with unknown format should error, got %T", unknownFormat)
	}
	if !strings.Contains(errObj.Message, "unknown format") {
		t.Errorf("unexpected error for unknown format: %s", errObj.Message)
	}

	unwritable := dumpFn(cfgJson, &Stringo{Value: filepath.Join(tmp, "missing-dir", "x.json")}, &Stringo{Value: "JSON"})
	if _, ok := unwritable.(*Error); !ok {
		t.Errorf("dump_config to unwritable path should error, got %T", unwritable)
	}

	tests := []builtinTestCase{
		{name: "cfg not string", args: []Object{in(1), &Stringo{Value: "p"}, &Stringo{Value: "JSON"}}, err: "PositionalTypeError"},
		{name: "fpath not string", args: []Object{&Stringo{Value: "{}"}, in(1), &Stringo{Value: "JSON"}}, err: "PositionalTypeError"},
		{name: "format not string", args: []Object{&Stringo{Value: "{}"}, &Stringo{Value: "p"}, in(1)}, err: "PositionalTypeError"},
		{name: "two args", args: []Object{&Stringo{Value: "{}"}, &Stringo{Value: "p"}}, err: "InvalidArgCountError"},
	}
	runBuiltinTestsFor(t, ConfigBuiltins, "_dump_config", tests)
}
