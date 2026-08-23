package binc

import (
	"encoding/binary"
	"fmt"

	"blue/token"
)

// Token table codec.
//
// The token table backs runtime error traces (the VM's OpNode operands are
// indices into it). Tokens repeat a lot (same filepaths, same keyword
// literals), so all strings are interned into one string table and each
// token is stored as five varints:
//
//	typeIdx      uvarint  index into the string table of token.Type
//	literalIdx   uvarint  index into the string table of Literal
//	filepathIdx  uvarint  index into the string table of Filepath
//	lineDelta    svarint  zigzag delta vs the previous token's line number
//	posInLine    uvarint  PositionInLine
//
// Layout: [uvarint numStrings][strings...: uvarint len + bytes]
//         [uvarint numTokens][tokens...]

func encodeTokens(tokens []*token.Token) ([]byte, error) {
	strIdx := map[string]uint64{}
	stringTable := make([]string, 0, 64)
	intern := func(s string) uint64 {
		if i, ok := strIdx[s]; ok {
			return i
		}
		i := uint64(len(stringTable))
		stringTable = append(stringTable, s)
		strIdx[s] = i
		return i
	}

	var body []byte
	prevLine := 0
	for _, t := range tokens {
		if t == nil {
			return nil, fmt.Errorf("binc: nil token in token table")
		}
		body = binary.AppendUvarint(body, intern(string(t.Type)))
		body = binary.AppendUvarint(body, intern(t.Literal))
		body = binary.AppendUvarint(body, intern(t.Filepath))
		body = binary.AppendVarint(body, int64(t.LineNumber-prevLine))
		prevLine = t.LineNumber
		body = binary.AppendUvarint(body, uint64(t.PositionInLine))
	}

	out := binary.AppendUvarint(nil, uint64(len(stringTable)))
	for _, s := range stringTable {
		out = binary.AppendUvarint(out, uint64(len(s)))
		out = append(out, s...)
	}
	out = binary.AppendUvarint(out, uint64(len(tokens)))
	out = append(out, body...)
	return out, nil
}

func decodeTokens(data []byte) ([]*token.Token, error) {
	readUvarint := func() (uint64, error) {
		v, n := binary.Uvarint(data)
		if n <= 0 {
			return 0, fmt.Errorf("binc: malformed token table")
		}
		data = data[n:]
		return v, nil
	}
	readVarint := func() (int64, error) {
		v, n := binary.Varint(data)
		if n <= 0 {
			return 0, fmt.Errorf("binc: malformed token table")
		}
		data = data[n:]
		return v, nil
	}
	readString := func() (string, error) {
		n, err := readUvarint()
		if err != nil {
			return "", err
		}
		if n > uint64(len(data)) {
			return "", fmt.Errorf("binc: malformed token table")
		}
		s := string(data[:n])
		data = data[n:]
		return s, nil
	}

	numStrings, err := readUvarint()
	if err != nil {
		return nil, err
	}
	stringTable := make([]string, 0, min(numStrings, 1<<20))
	for i := uint64(0); i < numStrings; i++ {
		s, err := readString()
		if err != nil {
			return nil, err
		}
		stringTable = append(stringTable, s)
	}
	get := func(i uint64) (string, error) {
		if i >= uint64(len(stringTable)) {
			return "", fmt.Errorf("binc: token table string index %d out of range (%d strings)", i, len(stringTable))
		}
		return stringTable[i], nil
	}

	numTokens, err := readUvarint()
	if err != nil {
		return nil, err
	}
	if numTokens > uint64(len(data)) { // every token needs at least 5 bytes
		return nil, fmt.Errorf("binc: malformed token table")
	}
	tokens := make([]*token.Token, 0, numTokens)
	prevLine := int64(0)
	for i := uint64(0); i < numTokens; i++ {
		typeIdx, err := readUvarint()
		if err != nil {
			return nil, err
		}
		literalIdx, err := readUvarint()
		if err != nil {
			return nil, err
		}
		filepathIdx, err := readUvarint()
		if err != nil {
			return nil, err
		}
		lineDelta, err := readVarint()
		if err != nil {
			return nil, err
		}
		posInLine, err := readUvarint()
		if err != nil {
			return nil, err
		}
		tokTypeStr, err := get(typeIdx)
		if err != nil {
			return nil, err
		}
		literal, err := get(literalIdx)
		if err != nil {
			return nil, err
		}
		filepath, err := get(filepathIdx)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, &token.Token{
			Type:           token.Type(tokTypeStr),
			Literal:        literal,
			Filepath:       filepath,
			LineNumber:     int(prevLine + lineDelta),
			PositionInLine: int(posInLine),
		})
		prevLine += lineDelta
	}
	return tokens, nil
}
