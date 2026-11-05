package ml

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const (
	// Tensor file format: magic, dtype, rank, shape, element count, then float32
	// elements in row-major order.
	tensorMagic = "BLUETNSR1"
	// maxTensorRank bounds the rank read from an untrusted file before the
	// shape slice is allocated.
	maxTensorRank = 32
	// maxTensorElements bounds the total number of elements accepted from an
	// untrusted file so a small header cannot request a huge allocation.
	maxTensorElements = 1 << 26
)

// tensorElementCount validates a shape read from an untrusted file and returns
// its total number of elements. It rejects negative dimensions and shapes whose
// product would overflow or exceed maxTensorElements.
func tensorElementCount(shape []int) (int, error) {
	total := int64(1)
	for _, d := range shape {
		if d < 0 {
			return 0, fmt.Errorf("load: tensor dimension %d must not be negative", d)
		}
		if d == 0 {
			total = 0
			continue
		}
		if total > int64(maxTensorElements)/int64(d) {
			return 0, fmt.Errorf("load: tensor shape %v is too large", shape)
		}
		total *= int64(d)
	}
	return int(total), nil
}

func Save(t *Tensor, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return writeTensor(f, t)
}

// Load reads a tensor written by Save onto the given device.
func Load(path string, device Device) (*Tensor, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return readTensor(f, device)
}

func writeTensor(w io.Writer, t *Tensor) error {
	if _, err := io.WriteString(w, tensorMagic); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint8(t.DType())); err != nil {
		return err
	}
	shape := t.Shape()
	if err := binary.Write(w, binary.LittleEndian, uint32(len(shape))); err != nil {
		return err
	}
	for _, d := range shape {
		if err := binary.Write(w, binary.LittleEndian, int64(d)); err != nil {
			return err
		}
	}
	data := t.ContiguousData()
	if err := binary.Write(w, binary.LittleEndian, data); err != nil {
		return err
	}
	return nil
}

func readTensor(r io.Reader, device Device) (*Tensor, error) {
	magic := make([]byte, len(tensorMagic))
	if _, err := io.ReadFull(r, magic); err != nil {
		return nil, err
	}
	if string(magic) != tensorMagic {
		return nil, fmt.Errorf("load: not a blue tensor file")
	}
	var dt uint8
	if err := binary.Read(r, binary.LittleEndian, &dt); err != nil {
		return nil, err
	}
	var rank uint32
	if err := binary.Read(r, binary.LittleEndian, &rank); err != nil {
		return nil, err
	}
	if rank > maxTensorRank {
		return nil, fmt.Errorf("load: tensor rank %d exceeds maximum %d", rank, maxTensorRank)
	}
	shape := make([]int, rank)
	for i := range shape {
		var d int64
		if err := binary.Read(r, binary.LittleEndian, &d); err != nil {
			return nil, err
		}
		shape[i] = int(d)
	}
	total, err := tensorElementCount(shape)
	if err != nil {
		return nil, err
	}
	data := make([]float32, total)
	if err := binary.Read(r, binary.LittleEndian, data); err != nil {
		return nil, err
	}
	return NewTensor(data, shape, DType(dt), device)
}
