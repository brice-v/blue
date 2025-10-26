package object

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/scrypt"
)

var CryptoBuiltins = NewBuiltinSliceType{
	{Name: "_sha", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("sha", len(args), 2, "")
			}
			if args[0].Type() != STRING_OBJ && args[0].Type() != BYTES_OBJ {
				return newPositionalTypeError("sha", 1, "STRING or BYTES", args[0].Type())
			}
			if args[1].Type() != INTEGER_OBJ {
				return newPositionalTypeError("sha", 2, INTEGER_OBJ, args[1].Type())
			}
			var bs []byte
			if args[0].Type() == STRING_OBJ {
				bs = []byte(args[0].(*Stringo).Value)
			} else {
				bs = args[0].(*Bytes).Value
			}
			i := args[1].(*Integer).Value
			var hasher hash.Hash
			switch i {
			case 1:
				hasher = sha1.New()
			case 256:
				hasher = sha256.New()
			case 512:
				hasher = sha512.New()
			default:
				return newError("argument 2 to `sha` should be 1, 256, or 512. got=%d", i)
			}
			hasher.Write(bs)
			return &Stringo{Value: fmt.Sprintf("%x", hasher.Sum(nil))}
		},
		HelpStr: helpStrArgs{
			explanation: "`sha` returns the sha 1, 256, or 512 sum of the given content as a STRING",
			signature:   "sha(content: str|bytes, type: int(1|256|512)) -> str",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "sha('a',1) => '86f7e437faa5a7fce15d1ddcb9eaeaea377667b8'",
		}.String(),
	}},
	{Name: "_md5", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("md5", len(args), 1, "")
			}
			if args[0].Type() != STRING_OBJ && args[0].Type() != BYTES_OBJ {
				return newPositionalTypeError("md5", 1, "STRING or BYTES", args[0].Type())
			}
			var bs []byte
			if args[0].Type() == STRING_OBJ {
				bs = []byte(args[0].(*Stringo).Value)
			} else {
				bs = args[0].(*Bytes).Value
			}
			hasher := md5.New()
			hasher.Write(bs)
			return &Stringo{Value: fmt.Sprintf("%x", hasher.Sum(nil))}
		},
		HelpStr: helpStrArgs{
			explanation: "`md5` returns the md5 sum of the given content as a STRING",
			signature:   "md5(content: str|bytes) -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "md5('a') => '0cc175b9c0f1b6a831c399e269772661'",
		}.String(),
	}},
	{Name: "_generate_from_password", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("generate_from_password", len(args), 1, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("generate_from_password", 1, STRING_OBJ, args[0].Type())
			}
			pw := []byte(args[0].(*Stringo).Value)
			hashedPw, err := bcrypt.GenerateFromPassword(pw, bcrypt.DefaultCost)
			if err != nil {
				return newError("bcrypt error: %s", err.Error())
			}
			return &Stringo{Value: string(hashedPw)}
		},
		HelpStr: helpStrArgs{
			explanation: "`generate_from_password` returns a bcyrpt STRING for the given password STRING",
			signature:   "generate_from_password(pw: str) -> str",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "generate_from_password('a') => '$2a$10$4GjpUS8/60qPsxFtPbo.3e5ueULg4Llk0iCwVsGAV9LBDuw2FkSa2'",
		}.String(),
	}},
	{Name: "_compare_hash_and_password", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("compare_hash_and_password", len(args), 2, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("compare_hash_and_password", 1, STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("compare_hash_and_password", 2, STRING_OBJ, args[1].Type())
			}
			hashedPw := []byte(args[0].(*Stringo).Value)
			pw := []byte(args[1].(*Stringo).Value)
			err := bcrypt.CompareHashAndPassword(hashedPw, pw)
			if err != nil {
				return newError("bcrypt error: %s", err.Error())
			}
			return TRUE
		},
		HelpStr: helpStrArgs{
			explanation: "`compare_hash_and_password` returns a true if the given hashed password matches the given password",
			signature:   "compare_hash_and_password(hashed_pw: str, pw: str) -> bool",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "compare_hash_and_password('$2a$10$4GjpUS8/60qPsxFtPbo.3e5ueULg4Llk0iCwVsGAV9LBDuw2FkSa2', 'a') => true",
		}.String(),
	}},
	{Name: "_encrypt", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("encrypt", len(args), 2, "")
			}
			if args[0].Type() != STRING_OBJ && args[0].Type() != BYTES_OBJ {
				return newPositionalTypeError("encrypt", 1, "STRING or BYTES", args[0].Type())
			}
			if args[1].Type() != STRING_OBJ && args[1].Type() != BYTES_OBJ {
				return newPositionalTypeError("encrypt", 2, "STRING or BYTES", args[1].Type())
			}
			var pw []byte
			if args[0].Type() == STRING_OBJ {
				pw = []byte(args[0].(*Stringo).Value)
			} else {
				pw = args[0].(*Bytes).Value
			}
			var data []byte
			if args[1].Type() == STRING_OBJ {
				data = []byte(args[1].(*Stringo).Value)
			} else {
				data = args[1].(*Bytes).Value
			}

			// Deriving key from pw as it needs to be 32 bytes
			salt := make([]byte, 32)
			if _, err := rand.Read(salt); err != nil {
				return newError("encrypt error: %s", err.Error())
			}
			key, err := scrypt.Key(pw, salt, 1048576, 8, 1, 32)
			if err != nil {
				return newError("encrypt error: %s", err.Error())
			}
			// Done Deriving key

			blockCipher, err := aes.NewCipher(key)
			if err != nil {
				return newError("encrypt error: %s", err.Error())
			}
			gcm, err := cipher.NewGCM(blockCipher)
			if err != nil {
				return newError("encrypt error: %s", err.Error())
			}
			nonce := make([]byte, gcm.NonceSize())
			if _, err = rand.Read(nonce); err != nil {
				return newError("encrypt error: %s", err.Error())
			}
			ciphertext := gcm.Seal(nonce, nonce, data, nil)
			ciphertext = append(ciphertext, salt...)
			return &Bytes{Value: ciphertext}
		},
		HelpStr: helpStrArgs{
			explanation: "`encrypt` encrypts the data given with the password given",
			signature:   "encrypt(pw: str|bytes, data: str|bytes) -> bytes",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "encrypt('a','test') => bytes",
		}.String(),
	}},
	{Name: "_decrypt", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("decrypt", len(args), 3, "")
			}
			if args[0].Type() != STRING_OBJ && args[0].Type() != BYTES_OBJ {
				return newPositionalTypeError("decrypt", 1, "STRING or BYTES", args[0].Type())
			}
			if args[1].Type() != BYTES_OBJ {
				return newPositionalTypeError("decrypt", 2, BYTES_OBJ, args[1].Type())
			}
			if args[2].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("decrypt", 3, BOOLEAN_OBJ, args[2].Type())
			}
			var pw []byte
			if args[0].Type() == STRING_OBJ {
				pw = []byte(args[0].(*Stringo).Value)
			} else {
				pw = args[0].(*Bytes).Value
			}
			data := args[1].(*Bytes).Value
			getDataAsBytes := args[2].(*Boolean).Value

			// Deriving key from pw as it needs to be 32 bytes
			salt, data := data[len(data)-32:], data[:len(data)-32]
			key, err := scrypt.Key(pw, salt, 1048576, 8, 1, 32)
			if err != nil {
				return newError("decrypt error: %s", err.Error())
			}
			// Done Deriving key

			blockCipher, err := aes.NewCipher(key)
			if err != nil {
				return newError("decrypt error: %s", err.Error())
			}
			gcm, err := cipher.NewGCM(blockCipher)
			if err != nil {
				return newError("decrypt error: %s", err.Error())
			}
			nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
			plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
			if err != nil {
				return newError("decrypt error: %s", err.Error())
			}
			if getDataAsBytes {
				return &Bytes{Value: plaintext}
			} else {
				return &Stringo{Value: string(plaintext)}
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`decrypt` decrypts the data given with the password given, bytes are returned if as_bytes is set to true",
			signature:   "decrypt(pw: str|bytes, data: bytes, as_bytes: bool=false) -> str|bytes",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "decrypt('a',bs) => 'test'",
		}.String(),
	}},
	{Name: "_encode_base_64_32", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("encode_base_64_32", len(args), 3, "")
			}
			if args[0].Type() != STRING_OBJ && args[0].Type() != BYTES_OBJ {
				return newPositionalTypeError("encode_base_64_32", 1, "STRING or BYTES", args[0].Type())
			}
			if args[1].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("encode_base_64_32", 2, BOOLEAN_OBJ, args[1].Type())
			}
			if args[2].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("encode_base_64_32", 3, BOOLEAN_OBJ, args[2].Type())
			}
			useBase64 := args[2].(*Boolean).Value
			var bs []byte
			if args[0].Type() == STRING_OBJ {
				bs = []byte(args[0].(*Stringo).Value)
			} else {
				bs = args[0].(*Bytes).Value
			}
			asBytes := args[1].(*Boolean).Value
			var encoded string
			if useBase64 {
				encoded = base64.StdEncoding.EncodeToString(bs)
			} else {
				encoded = base32.StdEncoding.EncodeToString(bs)
			}
			if asBytes {
				return &Bytes{Value: []byte(encoded)}
			}
			return &Stringo{Value: encoded}
		},
		HelpStr: helpStrArgs{
			explanation: "`encode_base_64_32` encodes the data given in base64 if true, else base32, bytes are returned if as_bytes is set to true. Note: this function should only be called from encode",
			signature:   "encode_base_64_32(data: str|bytes, is_base_64: bool=false, as_bytes: bool=false) -> str|bytes",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "encode_base_64_32('a', true, false) => 'YQ=='",
		}.String(),
	}},
	{Name: "_decode_base_64_32", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("decode_base_64_32", len(args), 3, "")
			}
			if args[0].Type() != STRING_OBJ && args[0].Type() != BYTES_OBJ {
				return newPositionalTypeError("decode_base_64_32", 1, "STRING or BYTES", args[0].Type())
			}
			if args[1].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("decode_base_64_32", 2, BOOLEAN_OBJ, args[1].Type())
			}
			if args[2].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("decode_base_64_32", 3, BOOLEAN_OBJ, args[2].Type())
			}
			useBase64 := args[2].(*Boolean).Value
			var s string
			if args[0].Type() == STRING_OBJ {
				s = args[0].(*Stringo).Value
			} else {
				s = string(args[0].(*Bytes).Value)
			}
			asBytes := args[1].(*Boolean).Value
			var decoded []byte
			var err error
			if useBase64 {
				decoded, err = base64.StdEncoding.DecodeString(s)
			} else {
				decoded, err = base32.StdEncoding.DecodeString(s)
			}
			if err != nil {
				return newError("`decode_base_64_32` error: %s", err.Error())
			}
			if !asBytes {
				return &Stringo{Value: string(decoded)}
			}
			return &Bytes{Value: decoded}
		},
		HelpStr: helpStrArgs{
			explanation: "`decode_base_64_32` decodes the data given in base64 if true, else base32, bytes are returned if as_bytes is set to true. Note: this function should only be called from decode",
			signature:   "decode_base_64_32(data: str|bytes, is_base_64: bool=false, as_bytes: bool=false) -> str|bytes",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "decode_base_64_32('YQ==', true, false) => 'a'",
		}.String(),
	}},
	{Name: "_decode_hex", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("decode_hex", len(args), 2, "")
			}
			if args[0].Type() != STRING_OBJ && args[0].Type() != BYTES_OBJ {
				return newPositionalTypeError("decode_hex", 1, "STRING or BYTES", args[0].Type())
			}
			if args[1].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("encode_hex", 2, BOOLEAN_OBJ, args[1].Type())
			}
			asBytes := args[1].(*Boolean).Value
			var bs []byte
			if args[0].Type() == STRING_OBJ {
				s := args[0].(*Stringo).Value
				data, err := hex.DecodeString(s)
				if err != nil {
					return newError("`decode_hex` error: %s", err.Error())
				}
				bs = data
			} else if args[0].Type() == BYTES_OBJ {
				b := args[0].(*Bytes).Value
				bs = make([]byte, hex.DecodedLen(len(b)))
				l, err := hex.Decode(bs, b)
				if err != nil {
					return newError("`decode_hex` error: %s", err.Error())
				}
				if l != len(b) {
					return newError("`decode_hex` error: length of bytes does not match bytes written. got=%d, want=%d", l, len(b))
				}
			}
			if !asBytes {
				return &Stringo{Value: string(bs)}
			}
			return &Bytes{Value: bs}
		},
		HelpStr: helpStrArgs{
			explanation: "`decode_hex` decodes the data given in hex, bytes are returned if as_bytes is set to true. Note: this function should only be called from decode",
			signature:   "decode_hex(data: str|bytes, as_bytes: bool=false) -> str|bytes",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "decode_hex('61') => 'a'",
		}.String(),
	}},
	{Name: "_encode_hex", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("encode_hex", len(args), 2, "")
			}
			if args[0].Type() != STRING_OBJ && args[0].Type() != BYTES_OBJ {
				return newPositionalTypeError("encode_hex", 1, "STRING or BYTES", args[0].Type())
			}
			if args[1].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("encode_hex", 2, BOOLEAN_OBJ, args[1].Type())
			}
			asBytes := args[1].(*Boolean).Value
			var s string
			if args[0].Type() == BYTES_OBJ {
				b := args[0].(*Bytes).Value
				s = hex.EncodeToString(b)
			} else if args[0].Type() == STRING_OBJ {
				b := args[0].(*Stringo).Value
				bs := make([]byte, hex.EncodedLen(len(b)))
				hex.Encode(bs, []byte(b))
				// if l != len(b) {
				// 	return newError("`encode_hex` error: length of bytes does not match bytes written. got=%d, want=%d", l, len(b))
				// }
				s = string(bs)
			}
			if asBytes {
				return &Bytes{Value: []byte(s)}
			}
			return &Stringo{Value: s}
		},
		HelpStr: helpStrArgs{
			explanation: "`encode_hex` encodes the data given as hex, bytes are returned if as_bytes is set to true. Note: this function should only be called from encode",
			signature:   "encode_hex(data: str|bytes, as_bytes: bool=false) -> str|bytes",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "encode_hex('a') => '61'",
		}.String(),
	}},
}
