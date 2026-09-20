package ml

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"slices"

	"blue/borncgo/tensor"
)

// NamedTensor is one entry of a model state dictionary: a name and the tensor
// it refers to. The tensor may be a view that shares storage with a model
// parameter, so copying into it updates the model in place.
type NamedTensor struct {
	Name   string
	Tensor *Tensor
}

// stateMagic identifies a blue state archive: a named collection of tensors.
const stateMagic = "BLUESTATE1"

// SaveState writes named tensors to one file. Unlike Save, which stores a
// single unnamed tensor, this keeps the names so a model can be rebuilt and
// loaded by name. All values are stored as float32 plus their dtype, so int and
// bool tensors round-trip too.
func SaveState(entries []NamedTensor, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return writeState(f, entries)
}

// LoadState reads a state archive written by SaveState. Tensors are loaded on
// the CPU; use CopyInto to place them into a model, which handles the device.
func LoadState(path string) ([]NamedTensor, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return readState(f)
}

func writeState(w io.Writer, entries []NamedTensor) error {
	if _, err := io.WriteString(w, stateMagic); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(len(entries))); err != nil {
		return err
	}
	for _, e := range entries {
		if err := binary.Write(w, binary.LittleEndian, uint32(len(e.Name))); err != nil {
			return err
		}
		if _, err := io.WriteString(w, e.Name); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, uint8(e.Tensor.DType())); err != nil {
			return err
		}
		shape := e.Tensor.Shape()
		if err := binary.Write(w, binary.LittleEndian, uint32(len(shape))); err != nil {
			return err
		}
		for _, d := range shape {
			if err := binary.Write(w, binary.LittleEndian, int64(d)); err != nil {
				return err
			}
		}
		if err := binary.Write(w, binary.LittleEndian, e.Tensor.ContiguousData()); err != nil {
			return err
		}
	}
	return nil
}

func readState(r io.Reader) ([]NamedTensor, error) {
	magic := make([]byte, len(stateMagic))
	if _, err := io.ReadFull(r, magic); err != nil {
		return nil, err
	}
	if string(magic) != stateMagic {
		return nil, fmt.Errorf("load_state: not a blue state file")
	}
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, err
	}
	out := make([]NamedTensor, count)
	for i := range out {
		var nameLen uint32
		if err := binary.Read(r, binary.LittleEndian, &nameLen); err != nil {
			return nil, err
		}
		name := make([]byte, nameLen)
		if _, err := io.ReadFull(r, name); err != nil {
			return nil, err
		}
		var dt uint8
		if err := binary.Read(r, binary.LittleEndian, &dt); err != nil {
			return nil, err
		}
		var rank uint32
		if err := binary.Read(r, binary.LittleEndian, &rank); err != nil {
			return nil, err
		}
		shape := make([]int, rank)
		for j := range shape {
			var d int64
			if err := binary.Read(r, binary.LittleEndian, &d); err != nil {
				return nil, err
			}
			shape[j] = int(d)
		}
		total := 1
		for _, d := range shape {
			total *= d
		}
		data := make([]float32, total)
		if err := binary.Read(r, binary.LittleEndian, data); err != nil {
			return nil, err
		}
		t, err := NewTensor(data, shape, DType(dt), CPU)
		if err != nil {
			return nil, err
		}
		out[i] = NamedTensor{Name: string(name), Tensor: t}
	}
	return out, nil
}

// CopyInto copies src's elements into dst in place. Shapes must match. It is
// dtype-aware, so a float32 source can fill an int32 or bool destination.
func CopyInto(dst, src *Tensor) error {
	if !slices.Equal(dst.Shape(), src.Shape()) {
		return fmt.Errorf("copy: shape mismatch %v vs %v", dst.Shape(), src.Shape())
	}
	data := src.ContiguousData()
	raw := dst.t.Raw()
	switch raw.DType() {
	case tensor.Float32:
		copy(raw.AsFloat32(), data)
	case tensor.Float64:
		s := raw.AsFloat64()
		for i := range s {
			s[i] = float64(data[i])
		}
	case tensor.Int32:
		s := raw.AsInt32()
		for i := range s {
			s[i] = int32(data[i])
		}
	case tensor.Int64:
		s := raw.AsInt64()
		for i := range s {
			s[i] = int64(data[i])
		}
	case tensor.Uint8:
		s := raw.AsUint8()
		for i := range s {
			s[i] = uint8(data[i])
		}
	case tensor.Bool:
		s := raw.AsBool()
		for i := range s {
			s[i] = data[i] != 0
		}
	default:
		return fmt.Errorf("copy: unsupported dtype %s", raw.DType())
	}
	return nil
}
