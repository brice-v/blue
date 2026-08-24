package object

import (
	"bytes"
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

// This file owns the constant-pool slice of the blue binary container
// (see package bluec). It encodes the pool as a cbor array of
// ObjectWrapper values and preserves the reserved constant slots
// (OBJECT_CONSTANTS, indices 0..len-1) by never writing them: the decoder
// reconstructs them with NewObjectConstants() first so constant indices in
// compiled instructions round-trip identically.

// ValidateReservedPrefix checks that constants[0:len(OBJECT_CONSTANTS)] are
// exactly the reserved constant objects (compared by identity, matching how
// IsConstantObject works). A compiler always starts its pool from
// NewObjectConstants() so this should hold for any compilable program.
func ValidateReservedPrefix(constants []Object) error {
	reserved := OBJECT_CONSTANTS
	if len(constants) < len(reserved) {
		return fmt.Errorf("constant pool too small: have %d constants, need at least %d reserved slots", len(constants), len(reserved))
	}
	for i, r := range reserved {
		if constants[i] != r {
			return fmt.Errorf("reserved constant slot %d does not hold the expected reserved object (got %T)", i, constants[i])
		}
	}
	return nil
}

// EncodeConstantPool encodes the non-reserved tail of the given pool. The
// reserved prefix is validated and skipped, see ValidateReservedPrefix.
func EncodeConstantPool(constants []Object) ([]byte, error) {
	if err := ValidateReservedPrefix(constants); err != nil {
		return nil, err
	}
	tail := constants[len(OBJECT_CONSTANTS):]
	wrappers := make([]ObjectWrapper, len(tail))
	for i, o := range tail {
		w, err := marshalObjectDepth(o, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to encode constant %d (%T): %w", i+len(OBJECT_CONSTANTS), o, err)
		}
		wrappers[i] = w
	}
	return cbor.Marshal(wrappers)
}

// DecodeConstantPool decodes a pool encoded by EncodeConstantPool,
// reconstructing the reserved constants first so that decoded indices match
// the ones the compiler emitted.
func DecodeConstantPool(data []byte) ([]Object, error) {
	var wrappers []ObjectWrapper
	if err := cbor.Unmarshal(data, &wrappers); err != nil {
		return nil, fmt.Errorf("failed to decode constant pool: %w", err)
	}
	constants := NewObjectConstants()
	for i, w := range wrappers {
		obj, err := decodeFromType(w.Type, w.Data, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to decode constant %d: %w", i+len(OBJECT_CONSTANTS), err)
		}
		constants = append(constants, obj)
	}
	return constants, nil
}

// FindUnserializableConstant returns the index of the first constant that
// cannot be serialized into a binary image, along with the error describing
// why. It returns -1 and nil when everything is serializable.
func FindUnserializableConstant(constants []Object) (int, error) {
	for i := len(OBJECT_CONSTANTS); i < len(constants); i++ {
		if _, err := marshalObjectDepth(constants[i], 0); err != nil {
			return i, err
		}
	}
	return -1, nil
}

// DebugDumpConstants renders every pool entry via Inspect for error messages.
func DebugDumpConstants(constants []Object) string {
	var buf bytes.Buffer
	for i, c := range constants {
		insp := "<nil>"
		if c != nil {
			insp = c.Inspect()
		}
		fmt.Fprintf(&buf, "%d: %s\n", i, insp)
	}
	return buf.String()
}
