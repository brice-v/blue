// Package binc defines and implements the blue binary container format
// ("BLUEBC"): a versioned, checksummed envelope for compiled blue programs
// so compiled output can be reused instead of re-running the lexer, parser
// and compiler.
//
// A container holds one merged program image (the compiler already merges
// lib/core/core.b, std modules and the user program into a single
// instruction stream and constant pool) plus the token table used for
// error traces.
//
// Layout (all integers little-endian; instruction operands keep their own
// big-endian encoding as defined by package code):
//
//	header:
//	  magic          8 bytes   "BLUEBC\x00"
//	  formatVersion  u16       container format version (FormatVersion)
//	  flags          u16       bit 0: tokens section stripped (FlagNoTokens)
//	  blueVersion    lp-string u32 length + bytes (consts.VERSION at compile time)
//	  fingerprint    lp-string build fingerprint (see Fingerprint())
//	  crc32          u32       IEEE CRC-32 of every byte after the CRC field
//	sections, each u32 length + bytes:
//	  instructions             raw code.Instructions blob
//	  constants                object.EncodeConstantPool output (cbor)
//	  tokens                   compact token table (see tokens.go), empty when stripped
//	  meta                     reserved for future metadata (empty in v1)
//	trailer (only meaningful when the payload is APPENDED to an executable,
//	but always written so both layouts are byte-identical):
//	  payloadSize    u64       size of header + sections + trailer
//	  reverseMagic   8 bytes   "BLUEBC\x00" reversed
//
// The trailer lets a packed executable locate its payload by seeking to the
// end of the file and reading backwards, without scanning the whole binary.
//
// Version bump policy: ANY change to the opcode set semantics, the constant
// pool layout, or this envelope must bump FormatVersion. Loaders reject
// containers whose FormatVersion they do not implement, and reject
// containers whose fingerprint does not match the running build.
package binc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"

	"blue/code"
	"blue/object"
	"blue/token"
)

// Magic is the leading byte sequence of every container.
var Magic = [8]byte{'B', 'L', 'U', 'E', 'B', 'C', 0x00}

// ReverseMagic closes an appended payload so it can be discovered from EOF.
var ReverseMagic = func() [8]byte {
	var r [8]byte
	for i := range Magic {
		r[i] = Magic[len(Magic)-1-i]
	}
	return r
}()

// FormatVersion is the current container format version.
const FormatVersion uint16 = 1

// FlagNoTokens marks images compiled without their token table.
const FlagNoTokens uint16 = 1 << 0

const (
	crc32Size   = 4
	trailerSize = 8 + 8 // payloadSize u64 + reverse magic
	lpStrLen    = 4     // u32 length prefix of a length-prefixed string
)

// ErrBadMagic is returned when the container magic (or trailing reverse
// magic) does not match.
var ErrBadMagic = errors.New("binc: not a blue binary container (bad magic)")

// ErrBadVersion is returned when the container was written by a different
// (usually newer or older) version of the format.
var ErrBadVersion = errors.New("binc: unsupported container format version")

// ErrFingerprintMismatch is returned when the image was compiled for a
// different build flavor (build tags, opcode set, blue version).
var ErrFingerprintMismatch = errors.New("binc: build fingerprint mismatch")

// ErrBadCRC is returned when the payload does not match its checksum.
var ErrBadCRC = errors.New("binc: checksum mismatch (corrupted container)")

// ErrTruncated is returned when the container ends unexpectedly.
var ErrTruncated = errors.New("binc: truncated container")

// Bytecode is the neutral home of what the VM needs to run a program.
// Package compiler aliases this type (compiler.Bytecode = binc.Bytecode) so
// existing call sites keep working, while packages that must not import the
// compiler (vm, minimal runners) use it directly.
type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
	Tokens       []*token.Token
}

// EncodeOptions tweaks how an image is encoded.
type EncodeOptions struct {
	// NoTokens strips the token table to shrink the file. Runtime error
	// traces from such an image cannot show source lines.
	NoTokens bool
}

// Encode encodes the image into the container format. The reserved
// constants prefix is validated before writing.
func Encode(bc *Bytecode, opts EncodeOptions) ([]byte, error) {
	if bc == nil {
		return nil, errors.New("binc: cannot encode nil bytecode")
	}
	constantsBlob, err := object.EncodeConstantPool(bc.Constants)
	if err != nil {
		return nil, fmt.Errorf("binc: %w", err)
	}
	flags := uint16(0)
	tokensBlob := []byte{}
	if opts.NoTokens || bc.Tokens == nil {
		flags |= FlagNoTokens
	} else {
		tokensBlob, err = encodeTokens(bc.Tokens)
		if err != nil {
			return nil, fmt.Errorf("binc: %w", err)
		}
	}
	meta := []byte{} // reserved for future metadata

	var buf bytes.Buffer
	buf.Write(Magic[:])
	writeUint16(&buf, FormatVersion)
	writeUint16(&buf, flags)
	writeLPString(&buf, BlueVersion())
	writeLPString(&buf, Fingerprint())

	sections := make([]byte, 0, len(bc.Instructions)+len(constantsBlob)+len(tokensBlob)+len(meta)+4*lpStrLen)
	sections = appendSection(sections, []byte(bc.Instructions))
	sections = appendSection(sections, constantsBlob)
	sections = appendSection(sections, tokensBlob)
	sections = appendSection(sections, meta)

	writeUint32(&buf, crc32.ChecksumIEEE(sections))
	buf.Write(sections)

	// Everything written so far (header + crc + sections), plus the trailer.
	writeUint64(&buf, uint64(buf.Len()+trailerSize))
	buf.Write(ReverseMagic[:])
	return buf.Bytes(), nil
}

// Decode decodes a container previously produced by Encode (or found
// appended to an executable, see FindAppendedPayload). It validates magic,
// format version, payload size, CRC-32 and the trailing reverse magic. Set
// checkEnv to also verify blue version and build fingerprint against the
// running process.
func Decode(data []byte, checkEnv bool) (*Bytecode, error) {
	minSize := 8 + 2 + 2 + lpStrLen + lpStrLen + crc32Size + trailerSize
	if len(data) < minSize {
		return nil, ErrTruncated
	}
	if !bytes.Equal(data[:8], Magic[:]) {
		return nil, ErrBadMagic
	}
	version := binary.LittleEndian.Uint16(data[8:10])
	if version != FormatVersion {
		return nil, fmt.Errorf("%w: got %d, supported %d", ErrBadVersion, version, FormatVersion)
	}
	flags := binary.LittleEndian.Uint16(data[10:12])

	off := 12
	blueVersion, n, err := readLPStringAt(data, off)
	if err != nil {
		return nil, err
	}
	off += n
	fingerprint, n, err := readLPStringAt(data, off)
	if err != nil {
		return nil, err
	}
	off += n

	trailer := data[len(data)-trailerSize:]
	if !bytes.Equal(trailer[8:], ReverseMagic[:]) {
		return nil, ErrBadMagic
	}
	payloadSize := binary.LittleEndian.Uint64(trailer[:8])
	if payloadSize != uint64(len(data)) {
		return nil, fmt.Errorf("%w: trailer says payload is %d bytes, got %d", ErrTruncated, payloadSize, len(data))
	}

	sectionsEnd := len(data) - trailerSize
	if sectionsEnd < off+crc32Size {
		return nil, ErrTruncated
	}
	storedCRC := binary.LittleEndian.Uint32(data[off : off+crc32Size])
	sections := data[off+crc32Size : sectionsEnd]
	if got := crc32.ChecksumIEEE(sections); got != storedCRC {
		return nil, fmt.Errorf("%w: stored %#08x, computed %#08x", ErrBadCRC, storedCRC, got)
	}
	if checkEnv {
		if err := CheckEnvironment(fingerprint, blueVersion); err != nil {
			return nil, err
		}
	}

	reader := bytes.NewReader(sections)
	instructionsBlob, err := readSection(reader)
	if err != nil {
		return nil, err
	}
	constantsBlob, err := readSection(reader)
	if err != nil {
		return nil, err
	}
	tokensBlob, err := readSection(reader)
	if err != nil {
		return nil, err
	}
	if _, err := readSection(reader); err != nil { // meta, unused in v1
		return nil, err
	}

	constants, err := object.DecodeConstantPool(constantsBlob)
	if err != nil {
		return nil, fmt.Errorf("binc: %w", err)
	}
	var tokens []*token.Token
	if flags&FlagNoTokens == 0 && len(tokensBlob) > 0 {
		tokens, err = decodeTokens(tokensBlob)
		if err != nil {
			return nil, fmt.Errorf("binc: %w", err)
		}
	}
	return &Bytecode{
		Instructions: code.Instructions(instructionsBlob),
		Constants:    constants,
		Tokens:       tokens,
	}, nil
}

// SniffMagic reports whether data starts with the container magic.
func SniffMagic(data []byte) bool {
	return len(data) >= len(Magic) && bytes.Equal(data[:len(Magic)], Magic[:])
}

func writeUint16(buf *bytes.Buffer, v uint16) {
	buf.WriteByte(byte(v))
	buf.WriteByte(byte(v >> 8))
}

func writeUint32(buf *bytes.Buffer, v uint32) {
	writeUint16(buf, uint16(v))
	writeUint16(buf, uint16(v>>16))
}

func writeUint64(buf *bytes.Buffer, v uint64) {
	writeUint32(buf, uint32(v))
	writeUint32(buf, uint32(v>>32))
}

func appendSection(dst, section []byte) []byte {
	var lenBuf [4]byte
	binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(section)))
	dst = append(dst, lenBuf[:]...)
	return append(dst, section...)
}

func readSection(r *bytes.Reader) ([]byte, error) {
	var lenBuf [4]byte
	if _, err := r.Read(lenBuf[:]); err != nil {
		return nil, ErrTruncated
	}
	n := binary.LittleEndian.Uint32(lenBuf[:])
	section := make([]byte, n)
	read, _ := r.Read(section)
	if uint32(read) != n {
		return nil, ErrTruncated
	}
	return section, nil
}

func writeLPString(buf *bytes.Buffer, s string) {
	writeUint32(buf, uint32(len(s)))
	buf.WriteString(s)
}

// readLPStringAt reads a u32-length-prefixed string starting at off. It
// returns the string and the total number of bytes consumed (including the
// length prefix).
func readLPStringAt(data []byte, off int) (string, int, error) {
	if off+lpStrLen > len(data) {
		return "", 0, ErrTruncated
	}
	n := int(binary.LittleEndian.Uint32(data[off : off+lpStrLen]))
	start := off + lpStrLen
	end := start + n
	if n < 0 || end > len(data) || end < start {
		return "", 0, ErrTruncated
	}
	return string(data[start:end]), lpStrLen + n, nil
}

// FindAppendedPayload looks for a container appended to raw executable
// bytes: it validates the reverse-magic trailer at the end and returns the
// full payload slice (header through trailer).
func FindAppendedPayload(exeBytes []byte) ([]byte, bool) {
	minSize := 8 + 2 + 2 + lpStrLen + lpStrLen + crc32Size + trailerSize
	if len(exeBytes) < minSize {
		return nil, false
	}
	trailer := exeBytes[len(exeBytes)-trailerSize:]
	if !bytes.Equal(trailer[8:], ReverseMagic[:]) {
		return nil, false
	}
	payloadSize := binary.LittleEndian.Uint64(trailer[:8])
	if payloadSize < uint64(minSize) || payloadSize > uint64(len(exeBytes)) {
		return nil, false
	}
	payload := exeBytes[uint64(len(exeBytes))-payloadSize:]
	if !SniffMagic(payload) {
		return nil, false
	}
	return payload, true
}
