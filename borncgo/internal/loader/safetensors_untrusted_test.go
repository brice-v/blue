package loader

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"blue/borncgo/internal/backend/cpu"
)

// writeSafeTensors writes a SafeTensors file with the given tensor header and a
// data section of dataLen zero bytes. dataLen is decoupled from what the header
// claims, so tests can point offsets past the real data or under-fill a tensor.
func writeSafeTensors(t *testing.T, path string, tensors map[string]SafeTensorInfo, dataLen int) {
	t.Helper()

	headerMap := make(map[string]any)
	headerMap["__metadata__"] = map[string]string{"format": "pt"}
	for name, info := range tensors {
		headerMap[name] = info
	}

	headerJSON, err := json.Marshal(headerMap)
	if err != nil {
		t.Fatalf("marshal header: %v", err)
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("close file: %v", err)
		}
	}()

	if err := binary.Write(file, binary.LittleEndian, uint64(len(headerJSON))); err != nil {
		t.Fatalf("write header size: %v", err)
	}
	if _, err := file.Write(headerJSON); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := file.Write(make([]byte, dataLen)); err != nil {
		t.Fatalf("write data: %v", err)
	}
}

// TestSafeTensorsReader_ReadTensorData_UntrustedOffsets checks that offsets the
// header should not be trusted about are rejected with a clean error rather than
// an out-of-range allocation panic or a backwards seek into the header.
func TestSafeTensorsReader_ReadTensorData_UntrustedOffsets(t *testing.T) {
	// The data section is 12 bytes; each case overrides the tensor's offsets.
	cases := []struct {
		name    string
		offsets [2]int64
	}{
		{"negative-start", [2]int64{-4, 8}},
		{"end-before-start", [2]int64{8, 4}},
		{"end-past-data", [2]int64{0, 1 << 40}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "corrupt.safetensors")
			writeSafeTensors(t, path, map[string]SafeTensorInfo{
				"t": {DType: SafeTensorsF32, Shape: []int{3}, DataOffsets: c.offsets},
			}, 12)

			reader, err := NewSafeTensorsReader(path)
			if err != nil {
				t.Fatalf("NewSafeTensorsReader: %v", err)
			}
			defer func() {
				if err := reader.Close(); err != nil {
					t.Errorf("Close: %v", err)
				}
			}()

			_, err = reader.ReadTensorData("t")
			if err == nil {
				t.Fatalf("ReadTensorData = nil error, want an out-of-range error")
			}
			if !strings.Contains(err.Error(), "out of range") {
				t.Errorf("error = %q, want it to contain %q", err, "out of range")
			}
		})
	}
}

// TestSafeTensorsReader_ReadTensorData_ShrunkFile checks that the read itself
// reports a file that shrank after the reader sampled its size. The bounds
// check compares against the size recorded at open, so a range that was valid
// then passes it and can no longer be read; the reader must say so rather than
// hand back a short or zero-filled buffer.
func TestSafeTensorsReader_ReadTensorData_ShrunkFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "shrunk.safetensors")
	writeSafeTensors(t, path, map[string]SafeTensorInfo{
		"t": {DType: SafeTensorsF32, Shape: []int{3}, DataOffsets: [2]int64{0, 12}},
	}, 12)

	reader, err := NewSafeTensorsReader(path)
	if err != nil {
		t.Fatalf("NewSafeTensorsReader: %v", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()

	// Drop the whole data section, leaving the header and the size the reader
	// already recorded intact.
	if err := os.Truncate(path, reader.dataOffset); err != nil {
		t.Fatalf("Truncate: %v", err)
	}

	_, err = reader.ReadTensorData("t")
	if err == nil {
		t.Fatalf("ReadTensorData = nil error, want a read error")
	}
	if !strings.Contains(err.Error(), "failed to read tensor data") {
		t.Errorf("error = %q, want it to contain %q", err, "failed to read tensor data")
	}
}

// TestSafeTensorsReader_LoadTensor_UntrustedHeader checks that LoadTensor
// rejects tensors whose header fields cannot be trusted — an invalid shape, an
// unsupported dtype, or a data section shorter than the shape and dtype demand
// (which must be reported, not silently zero-filled) — rather than allocating a
// wrong-sized or partially-filled tensor. Every offset stays within the file,
// so each mismatch must be caught by LoadTensor itself.
func TestSafeTensorsReader_LoadTensor_UntrustedHeader(t *testing.T) {
	cases := []struct {
		name    string
		info    SafeTensorInfo
		dataLen int
		wantErr string
	}{
		{
			"invalid-shape",
			SafeTensorInfo{DType: SafeTensorsF32, Shape: []int{2, -3}, DataOffsets: [2]int64{0, 12}},
			12,
			"invalid shape",
		},
		{
			"unsupported-dtype",
			SafeTensorInfo{DType: "F8", Shape: []int{3}, DataOffsets: [2]int64{0, 12}},
			12,
			"unsupported dtype",
		},
		{
			// Every dimension is positive, but the element count wraps
			// (4 * 2^62 == 2^64 == 0), so the length check below would accept
			// the empty data section this pairs it with and hand back a tensor
			// whose declared shape has nothing to do with its buffer.
			"overflowing-shape",
			SafeTensorInfo{DType: SafeTensorsF32, Shape: []int{4, 1 << 62}, DataOffsets: [2]int64{0, 0}},
			12,
			"overflows",
		},
		{
			// Header claims a [2, 3] F32 tensor (24 bytes) but the data holds 12.
			"short-data",
			SafeTensorInfo{DType: SafeTensorsF32, Shape: []int{2, 3}, DataOffsets: [2]int64{0, 12}},
			12,
			"does not match shape",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "untrusted.safetensors")
			writeSafeTensors(t, path, map[string]SafeTensorInfo{"t": c.info}, c.dataLen)

			reader, err := NewSafeTensorsReader(path)
			if err != nil {
				t.Fatalf("NewSafeTensorsReader: %v", err)
			}
			defer func() {
				if err := reader.Close(); err != nil {
					t.Errorf("Close: %v", err)
				}
			}()

			_, err = reader.LoadTensor("t", cpu.New())
			if err == nil {
				t.Fatalf("LoadTensor = nil error, want one containing %q", c.wantErr)
			}
			if !strings.Contains(err.Error(), c.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err, c.wantErr)
			}
		})
	}
}

// TestSafeTensorsReader_ReadTensorData_Concurrent checks that concurrent reads
// on one reader each return their own tensor's bytes. A shared seek cursor would
// let them interleave; the file is internally locked, so the corruption is
// wrong-but-consistent bytes rather than a data race -race would flag.
func TestSafeTensorsReader_ReadTensorData_Concurrent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.safetensors")
	createTestSafeTensorsFile(t, path)

	reader, err := NewSafeTensorsReader(path)
	if err != nil {
		t.Fatalf("NewSafeTensorsReader: %v", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()

	// Golden bytes read one at a time, before any concurrency.
	golden := make(map[string][]byte)
	for _, name := range []string{"weight", "bias"} {
		data, err := reader.ReadTensorData(name)
		if err != nil {
			t.Fatalf("ReadTensorData(%s): %v", name, err)
		}
		golden[name] = data
	}

	var wg sync.WaitGroup
	for _, name := range []string{"weight", "bias"} {
		for range 50 {
			wg.Add(1)
			go func(name string) {
				defer wg.Done()
				data, err := reader.ReadTensorData(name)
				if err != nil {
					t.Errorf("ReadTensorData(%s): %v", name, err)
					return
				}
				if !bytes.Equal(data, golden[name]) {
					t.Errorf("ReadTensorData(%s): interleaved read returned wrong bytes", name)
				}
			}(name)
		}
	}
	wg.Wait()
}
