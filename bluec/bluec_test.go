package bluec_test

import (
	"bytes"
	"strings"
	"testing"

	"blue/bluec"
	"blue/compiler"
	"blue/consts"
	"blue/lexer"
	"blue/object"
	"blue/parser"
)

func compileSource(t *testing.T, src string) *bluec.Bytecode {
	t.Helper()
	l := lexer.New(src, "<test>")
	p := parser.New(l)
	program := p.ParseProgram()
	if p.HasErrors() {
		t.Fatalf("parser errors: %v", p.ErrorMessages())
	}
	symbolTable := compiler.NewSymbolTable()
	constants := object.NewObjectConstants()
	for i, v := range object.AllBuiltins[0].Builtins {
		symbolTable.DefineBuiltin(i, v.Name, 0, v.Help())
	}
	c := compiler.NewWithStateAndCore(symbolTable, constants)
	if err := c.Compile(program); err != nil {
		t.Fatalf("compiler error: %s", err.Error())
	}
	return c.Bytecode()
}

const sampleProgram = `
fun addDefault(a, b=10) {
    return a + b
}
var m = {"key": [1, 2.5, 0x1f, true, null]}
m["set-like"] = {1, 2, 3}
val big = 123456789012345678901234567890
val bf = 1.79769313486231570000000001
val re = r/ab+c/
val bs = to_bytes("hello")
println(addDefault(5))
for i in [1, 2, 3] {
    println(i)
}
try {
    println("Try")
} catch (e) {
    println("Catch #{e}")
} finally {
    println("Finally")
}
`

func TestEncodeDecodeRoundTrip(t *testing.T) {
	bc := compileSource(t, sampleProgram)

	data, err := bluec.Encode(bc, bluec.EncodeOptions{})
	if err != nil {
		t.Fatalf("Encode failed: %s", err.Error())
	}

	decoded, err := bluec.Decode(data, false)
	if err != nil {
		t.Fatalf("Decode failed: %s", err.Error())
	}

	if !bytes.Equal([]byte(bc.Instructions), []byte(decoded.Instructions)) {
		t.Fatalf("instructions did not round-trip byte-for-byte")
	}

	if len(bc.Constants) != len(decoded.Constants) {
		t.Fatalf("constant count mismatch: %d vs %d", len(bc.Constants), len(decoded.Constants))
	}
	for i := range bc.Constants {
		want, got := bc.Constants[i], decoded.Constants[i]
		if want.Type() != got.Type() {
			t.Errorf("constant %d type mismatch: %s vs %s", i, want.Type(), got.Type())
			continue
		}
		if want.Inspect() != got.Inspect() {
			t.Errorf("constant %d (%s) inspect mismatch:\nwant %q\ngot  %q", i, want.Type(), want.Inspect(), got.Inspect())
		}
	}

	if len(bc.Tokens) != len(decoded.Tokens) {
		t.Fatalf("token count mismatch: %d vs %d", len(bc.Tokens), len(decoded.Tokens))
	}
	for i := range bc.Tokens {
		w, g := *bc.Tokens[i], *decoded.Tokens[i]
		if w != g {
			t.Errorf("token %d mismatch: %+v vs %+v", i, w, g)
		}
	}
}

func TestReservedConstantsPreserved(t *testing.T) {
	bc := compileSource(t, sampleProgram)
	data, err := bluec.Encode(bc, bluec.EncodeOptions{})
	if err != nil {
		t.Fatalf("Encode failed: %s", err.Error())
	}
	decoded, err := bluec.Decode(data, false)
	if err != nil {
		t.Fatalf("Decode failed: %s", err.Error())
	}
	reserved := object.NewObjectConstants()
	for i, r := range reserved {
		if decoded.Constants[i] != r {
			t.Errorf("reserved constant slot %d not identical after decode", i)
		}
	}
}

func TestNoTokensOption(t *testing.T) {
	bc := compileSource(t, sampleProgram)
	data, err := bluec.Encode(bc, bluec.EncodeOptions{NoTokens: true})
	if err != nil {
		t.Fatalf("Encode failed: %s", err.Error())
	}
	fullData, _ := bluec.Encode(bc, bluec.EncodeOptions{})
	if len(data) >= len(fullData) {
		t.Errorf("--no-tokens image should be smaller")
	}
	decoded, err := bluec.Decode(data, false)
	if err != nil {
		t.Fatalf("Decode failed: %s", err.Error())
	}
	if len(decoded.Tokens) != 0 {
		t.Errorf("expected no tokens in stripped image, got %d", len(decoded.Tokens))
	}
	if !bytes.Equal([]byte(bc.Instructions), []byte(decoded.Instructions)) {
		t.Errorf("instructions did not round-trip with stripped tokens")
	}
}

func TestDecodeRejectsMalformedInput(t *testing.T) {
	bc := compileSource(t, "println(1)")
	valid, err := bluec.Encode(bc, bluec.EncodeOptions{})
	if err != nil {
		t.Fatal(err)
	}

	cases := map[string][]byte{
		"empty":            {},
		"garbage":          []byte("this is not a container at all"),
		"bad magic":        append([]byte("BLAHBC\x00"), valid[8:]...),
		"truncated header": valid[:10],
		"truncated body":   valid[:len(valid)/2],
		"flipped bit":      flipBit(valid, len(valid)/2),
		"corrupt crc":      flipBit(valid, crcFieldOffset()),
		"bad version":      bumpVersion(valid),
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := bluec.Decode(data, false)
			if err == nil {
				t.Fatalf("expected error decoding malformed container, got image with %d constants", len(got.Constants))
			}
		})
	}
}

// crcFieldOffset computes where the CRC lives inside an encoded container:
// after magic(8) + version(2) + flags(2) + two length-prefixed strings.
func crcFieldOffset() int {
	return 8 + 2 + 2 + 4 + len(consts.VERSION) + 4 + len(bluec.Fingerprint())
}

func TestFindAppendedPayload(t *testing.T) {
	bc := compileSource(t, sampleProgram)
	payload, err := bluec.Encode(bc, bluec.EncodeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	exe := append([]byte("fake mach-o / elf executable bytes \x00\x01\x02"), payload...)

	found, ok := bluec.FindAppendedPayload(exe)
	if !ok {
		t.Fatalf("payload not found in fake executable")
	}
	if !bytes.Equal(found, payload) {
		t.Fatalf("found payload does not match original")
	}

	if _, ok := bluec.FindAppendedPayload([]byte("short")); ok {
		t.Errorf("should not find payload in short data")
	}
	if _, ok := bluec.FindAppendedPayload(nil); ok {
		t.Errorf("should not find payload in empty data")
	}
}

func TestFingerprintIsStableAndDescriptive(t *testing.T) {
	fp1 := bluec.Fingerprint()
	fp2 := bluec.Fingerprint()
	if fp1 != fp2 {
		t.Fatalf("fingerprint not stable within one process:\n%s\n%s", fp1, fp2)
	}
	desc := bluec.DescribeFingerprintMismatch(fp1, fp2)
	if desc != "" {
		t.Fatalf("identical fingerprints reported a difference: %s", desc)
	}
	mutated := strings.Replace(fp1, "tags:", "tags:x", 1)
	desc = bluec.DescribeFingerprintMismatch(fp1, mutated)
	if desc == "" {
		t.Fatalf("no difference reported for mutated fingerprint")
	}
}

func flipBit(b []byte, pos int) []byte {
	out := make([]byte, len(b))
	copy(out, b)
	out[pos] ^= 0xff
	return out
}

// FuzzDecode asserts that arbitrary input never panics the loader: every
// failure must surface as an error return. Seed with a small valid
// container; run via `go test -fuzz FuzzDecode ./bluec/`.
func FuzzDecode(f *testing.F) {
	bc := &bluec.Bytecode{
		Instructions: []byte{byte(0)},
		Constants:    object.NewObjectConstants(),
		Tokens:       nil,
	}
	valid, err := bluec.Encode(bc, bluec.EncodeOptions{})
	if err != nil {
		f.Fatal(err)
	}
	f.Add(valid)
	f.Add(valid[:len(valid)/2])
	f.Add([]byte("BLUEBC\x00garbage"))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = bluec.Decode(data, false)
	})
}

func bumpVersion(b []byte) []byte {
	out := make([]byte, len(b))
	copy(out, b)
	out[8] = 99
	return out
}
