package object

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"strings"
	"testing"
)

func cryptoBuiltinFn(t *testing.T, name string) BuiltinFunction {
	t.Helper()
	for _, b := range CryptoBuiltins {
		if b.Name == name {
			if b.Fun == nil {
				t.Fatalf("crypto builtin %q has nil Fun", name)
			}
			return b.Fun
		}
	}
	t.Fatalf("crypto builtin %q not found", name)
	return nil
}

func TestCryptoRegistry(t *testing.T) {
	seen := make(map[string]bool)
	for _, b := range CryptoBuiltins {
		if b.Name == "" || b.HelpStr == "" {
			t.Fatalf("crypto builtin %q missing Name or HelpStr", b.Name)
		}
		if seen[b.Name] {
			t.Errorf("duplicate crypto builtin name %q", b.Name)
		}
		seen[b.Name] = true
		if b.Fun == nil {
			t.Errorf("crypto builtin %q has nil Fun", b.Name)
		}
	}
}

func TestShaBuiltin(t *testing.T) {
	tests := []builtinTestCase{
		{name: "sha1 of a", args: []Object{&Stringo{Value: "a"}, in(1)}, want: fmt.Sprintf("%x", sha1.Sum([]byte("a")))},
		{name: "sha256 of a", args: []Object{&Stringo{Value: "a"}, in(256)}, want: fmt.Sprintf("%x", sha256.Sum256([]byte("a")))},
		{name: "sha512 of a", args: []Object{&Stringo{Value: "a"}, in(512)}, want: fmt.Sprintf("%x", sha512.Sum512([]byte("a")))},
		{name: "sha1 from bytes", args: []Object{&Bytes{Value: []byte("a")}, in(1)}, want: "86f7e437faa5a7fce15d1ddcb9eaeaea377667b8"},
		{name: "empty input", args: []Object{&Stringo{Value: ""}, in(256)}, want: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{name: "invalid variant", args: []Object{&Stringo{Value: "a"}, in(128)}, err: "should be 1, 256, or 512"},
		{name: "content wrong type", args: []Object{in(5), in(256)}, err: "PositionalTypeError"},
		{name: "variant not int", args: []Object{&Stringo{Value: "a"}, fl(256)}, err: "PositionalTypeError"},
		{name: "no args", args: []Object{}, err: "InvalidArgCountError"},
		{name: "too many args", args: []Object{&Stringo{Value: "a"}, in(1), in(1)}, err: "InvalidArgCountError"},
	}
	runBuiltinTestsFor(t, CryptoBuiltins, "_sha", tests)
}

func TestMd5Builtin(t *testing.T) {
	tests := []builtinTestCase{
		{name: "md5 of a", args: []Object{&Stringo{Value: "a"}}, want: "0cc175b9c0f1b6a831c399e269772661"},
		{name: "md5 from bytes", args: []Object{&Bytes{Value: []byte("hello")}}, want: "5d41402abc4b2a76b9719d911017c592"},
		{name: "md5 empty", args: []Object{&Stringo{Value: ""}}, want: "d41d8cd98f00b204e9800998ecf8427e"},
		{name: "wrong type", args: []Object{in(5)}, err: "PositionalTypeError"},
		{name: "no args", args: []Object{}, err: "InvalidArgCountError"},
	}
	runBuiltinTestsFor(t, CryptoBuiltins, "_md5", tests)
}

func TestBcryptRoundTrip(t *testing.T) {
	genFn := cryptoBuiltinFn(t, "_generate_from_password")
	cmpFn := cryptoBuiltinFn(t, "_compare_hash_and_password")

	hashed := genFn(&Stringo{Value: "s3cret"})
	hs, ok := hashed.(*Stringo)
	if !ok {
		t.Fatalf("generate_from_password returned %T", hashed)
	}
	if !strings.HasPrefix(hs.Value, "$2") {
		t.Errorf("bcrypt hash should start with $2, got %q", hs.Value)
	}
	if res := cmpFn(&Stringo{Value: hs.Value}, &Stringo{Value: "s3cret"}); res != TRUE {
		t.Errorf("correct password did not verify, got %v", res)
	}
	wrong := cmpFn(&Stringo{Value: hs.Value}, &Stringo{Value: "wrong"})
	errObj, ok := wrong.(*Error)
	if !ok {
		t.Fatalf("wrong password should return Error, got %T", wrong)
	}
	if !strings.Contains(errObj.Message, "bcrypt error") {
		t.Errorf("unexpected bcrypt error message: %s", errObj.Message)
	}
	runBuiltinTestsFor(t, CryptoBuiltins, "_generate_from_password", []builtinTestCase{
		{name: "not a string", args: []Object{in(1)}, err: "PositionalTypeError"},
		{name: "no args", args: []Object{}, err: "InvalidArgCountError"},
	})
	runBuiltinTestsFor(t, CryptoBuiltins, "_compare_hash_and_password", []builtinTestCase{
		{name: "hash not string", args: []Object{in(1), &Stringo{Value: "pw"}}, err: "PositionalTypeError"},
		{name: "pw not string", args: []Object{&Stringo{Value: "$2a$10$x"}, in(1)}, err: "PositionalTypeError"},
		{name: "one arg", args: []Object{&Stringo{Value: "$2a$10$x"}}, err: "InvalidArgCountError"},
	})
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	encFn := cryptoBuiltinFn(t, "_encrypt")
	decFn := cryptoBuiltinFn(t, "_decrypt")

	cipherObj := encFn(&Stringo{Value: "hunter2"}, &Stringo{Value: "attack at dawn"})
	ct, ok := cipherObj.(*Bytes)
	if !ok {
		t.Fatalf("encrypt returned %T, want *Bytes", cipherObj)
	}
	if len(ct.Value) <= 32 {
		t.Fatalf("ciphertext suspiciously short: %d bytes", len(ct.Value))
	}
	ct2 := encFn(&Stringo{Value: "hunter2"}, &Stringo{Value: "attack at dawn"}).(*Bytes)
	if string(ct.Value) == string(ct2.Value) {
		t.Error("encrypt with random salt/nonce produced identical ciphertext twice")
	}

	asStr := decFn(&Stringo{Value: "hunter2"}, ct, FALSE)
	s, ok := asStr.(*Stringo)
	if !ok || s.Value != "attack at dawn" {
		t.Errorf("decrypt(as_bytes=false) = %#v, want 'attack at dawn'", asStr.Inspect())
	}
	asBs := decFn(&Bytes{Value: []byte("hunter2")}, ct, TRUE)
	b, ok := asBs.(*Bytes)
	if !ok || string(b.Value) != "attack at dawn" {
		t.Errorf("decrypt(as_bytes=true) = %#v, want bytes 'attack at dawn'", asBs.Inspect())
	}

	badPw := decFn(&Stringo{Value: "wrong"}, ct, FALSE)
	if _, ok := badPw.(*Error); !ok {
		t.Errorf("decrypt with wrong password should error, got %T", badPw)
	}

	shortData := decFn(&Stringo{Value: "pw"}, &Bytes{Value: []byte("abc")}, FALSE)
	if _, ok := shortData.(*Error); !ok {
		t.Fatalf("decrypt with short data should error, got %T %v", shortData, shortData)
	}

	tampered := make([]byte, len(ct.Value))
	copy(tampered, ct.Value)
	tampered[0] ^= 0xff
	tampRes := decFn(&Stringo{Value: "hunter2"}, &Bytes{Value: tampered}, FALSE)
	if _, ok := tampRes.(*Error); !ok {
		t.Errorf("decrypt of tampered ciphertext should error, got %T", tampRes)
	}

	runBuiltinTestsFor(t, CryptoBuiltins, "_encrypt", []builtinTestCase{
		{name: "data wrong type", args: []Object{&Stringo{Value: "pw"}, in(1)}, err: "PositionalTypeError"},
		{name: "pw wrong type", args: []Object{in(1), &Stringo{Value: "d"}}, err: "PositionalTypeError"},
		{name: "one arg", args: []Object{&Stringo{Value: "pw"}}, err: "InvalidArgCountError"},
	})
	runBuiltinTestsFor(t, CryptoBuiltins, "_decrypt", []builtinTestCase{
		{name: "pw wrong type", args: []Object{in(1), &Bytes{Value: make([]byte, 64)}, FALSE}, err: "PositionalTypeError"},
		{name: "data not bytes", args: []Object{&Stringo{Value: "pw"}, &Stringo{Value: "xx"}, FALSE}, err: "PositionalTypeError"},
		{name: "as_bytes not bool", args: []Object{&Stringo{Value: "pw"}, &Bytes{Value: make([]byte, 64)}, in(1)}, err: "PositionalTypeError"},
		{name: "two args", args: []Object{&Stringo{Value: "pw"}, &Bytes{Value: make([]byte, 64)}}, err: "InvalidArgCountError"},
	})
}

func TestBase64Base32Builtins(t *testing.T) {
	tests := []builtinTestCase{
		{name: "base32 str", args: []Object{&Stringo{Value: "a"}, FALSE, FALSE}, want: "ME======"},
		{name: "base32 str as bytes", args: []Object{&Stringo{Value: "a"}, TRUE, FALSE}, want: "[]byte{0x4d, 0x45, 0x3d, 0x3d, 0x3d, 0x3d, 0x3d, 0x3d}"},
		{name: "base32 from bytes", args: []Object{&Bytes{Value: []byte("a")}, FALSE, FALSE}, want: "ME======"},
		{name: "base64 str", args: []Object{&Stringo{Value: "a"}, FALSE, TRUE}, want: "YQ=="},
		{name: "data wrong type", args: []Object{in(1), FALSE, FALSE}, err: "PositionalTypeError"},
		{name: "as_bytes not bool", args: []Object{&Stringo{Value: "a"}, in(1), FALSE}, err: "PositionalTypeError"},
		{name: "is_base64 not bool", args: []Object{&Stringo{Value: "a"}, FALSE, in(1)}, err: "PositionalTypeError"},
		{name: "two args", args: []Object{&Stringo{Value: "a"}, FALSE}, err: "InvalidArgCountError"},
	}
	runBuiltinTestsFor(t, CryptoBuiltins, "_encode_base_64_32", tests)

	decodeTests := []builtinTestCase{
		{name: "base64 to str", args: []Object{&Stringo{Value: "YQ=="}, FALSE, TRUE}, want: "a"},
		{name: "base64 to bytes", args: []Object{&Stringo{Value: "YQ=="}, TRUE, TRUE}, want: "[]byte{0x61}"},
		{name: "base64 from bytes", args: []Object{&Bytes{Value: []byte("YQ==")}, FALSE, TRUE}, want: "a"},
		{name: "base32", args: []Object{&Stringo{Value: "ME======"}, FALSE, FALSE}, want: "a"},
		{name: "invalid base64", args: []Object{&Stringo{Value: "!!!"}, FALSE, TRUE}, err: "`decode_base_64_32` error"},
		{name: "data wrong type", args: []Object{in(1), FALSE, TRUE}, err: "PositionalTypeError"},
		{name: "flags wrong type", args: []Object{&Stringo{Value: "YQ=="}, fl(1), TRUE}, err: "PositionalTypeError"},
		{name: "four args", args: []Object{&Stringo{Value: "YQ=="}, FALSE, TRUE, TRUE}, err: "InvalidArgCountError"},
	}
	runBuiltinTestsFor(t, CryptoBuiltins, "_decode_base_64_32", decodeTests)

	for _, isB64 := range []bool{false, true} {
		enc := cryptoBuiltinFn(t, "_encode_base_64_32")(&Stringo{Value: "round trip me"}, FALSE, BOOL(isB64))
		dec := cryptoBuiltinFn(t, "_decode_base_64_32")(enc, FALSE, BOOL(isB64))
		if s := dec.(*Stringo); s.Value != "round trip me" {
			t.Errorf("base%d round trip failed: got %q", map[bool]int{true: 64, false: 32}[isB64], s.Value)
		}
	}
}

func BOOL(b bool) Object { return nativeToBooleanObject(b) }

func TestHexBuiltins(t *testing.T) {
	tests := []builtinTestCase{
		{name: "encode str", args: []Object{&Stringo{Value: "a"}, FALSE}, want: "61"},
		{name: "encode str as bytes", args: []Object{&Stringo{Value: "a"}, TRUE}, want: "[]byte{0x36, 0x31}"},
		{name: "encode bytes", args: []Object{&Bytes{Value: []byte("\x01\x02")}, FALSE}, want: "0102"},
		{name: "encode empty", args: []Object{&Stringo{Value: ""}, FALSE}, want: ""},
		{name: "data wrong type", args: []Object{in(1), FALSE}, err: "PositionalTypeError"},
		{name: "as_bytes not bool", args: []Object{&Stringo{Value: "a"}, in(1)}, err: "PositionalTypeError"},
		{name: "three args", args: []Object{&Stringo{Value: "a"}, FALSE, FALSE}, err: "InvalidArgCountError"},
	}
	runBuiltinTestsFor(t, CryptoBuiltins, "_encode_hex", tests)

	decodeTests := []builtinTestCase{
		{name: "decode to str", args: []Object{&Stringo{Value: "6869"}, FALSE}, want: "hi"},
		{name: "decode to bytes", args: []Object{&Stringo{Value: "6869"}, TRUE}, want: "[]byte{0x68, 0x69}"},
		{name: "decode from bytes", args: []Object{&Bytes{Value: []byte("6869")}, FALSE}, want: "hi"},
		{name: "odd length", args: []Object{&Stringo{Value: "abc"}, FALSE}, err: "`decode_hex` error"},
		{name: "non-hex chars", args: []Object{&Stringo{Value: "zz"}, FALSE}, err: "`decode_hex` error"},
		{name: "data wrong type", args: []Object{in(1), FALSE}, err: "PositionalTypeError"},
		{name: "as_bytes not bool", args: []Object{&Stringo{Value: "aa"}, fl(1)}, err: "PositionalTypeError"},
		{name: "one arg", args: []Object{&Stringo{Value: "aa"}}, err: "InvalidArgCountError"},
	}
	runBuiltinTestsFor(t, CryptoBuiltins, "_decode_hex", decodeTests)

	orig := &Bytes{Value: []byte{0xde, 0xad, 0xbe, 0xef}}
	hexed := cryptoBuiltinFn(t, "_encode_hex")(orig, FALSE).(*Stringo)
	back := cryptoBuiltinFn(t, "_decode_hex")(&Stringo{Value: hexed.Value}, TRUE).(*Bytes)
	if string(back.Value) != string(orig.Value) {
		t.Errorf("hex round trip failed: got %x, want %x", back.Value, orig.Value)
	}
}
