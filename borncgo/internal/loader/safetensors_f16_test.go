package loader

import (
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"blue/borncgo/internal/backend/cpu"
	"blue/borncgo/internal/half"
	"blue/borncgo/internal/tensor"
)

// u16leBytes encodes 16-bit values in little-endian order.
func u16leBytes(vals []uint16) []byte {
	b := make([]byte, len(vals)*2)
	for i, v := range vals {
		binary.LittleEndian.PutUint16(b[i*2:], v)
	}
	return b
}

// writeSafeTensorsHalfFixture writes a SafeTensors file with one F16 tensor
// ("half") and one BF16 tensor ("brain"), each holding the given 16-bit
// little-endian element bits, laid out contiguously in that order.
func writeSafeTensorsHalfFixture(t *testing.T, path string, f16, bf16 []uint16) {
	t.Helper()

	f16Bytes := u16leBytes(f16)
	bf16Bytes := u16leBytes(bf16)

	tensors := map[string]SafeTensorInfo{
		"half": {
			DType:       SafeTensorsF16,
			Shape:       []int{len(f16)},
			DataOffsets: [2]int64{0, int64(len(f16Bytes))},
		},
		"brain": {
			DType:       SafeTensorsBF16,
			Shape:       []int{len(bf16)},
			DataOffsets: [2]int64{int64(len(f16Bytes)), int64(len(f16Bytes) + len(bf16Bytes))},
		},
	}

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
		t.Fatalf("create fixture: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("close fixture: %v", err)
		}
	}()

	if err := binary.Write(file, binary.LittleEndian, uint64(len(headerJSON))); err != nil {
		t.Fatalf("write header size: %v", err)
	}
	if _, err := file.Write(headerJSON); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := file.Write(f16Bytes); err != nil {
		t.Fatalf("write f16 data: %v", err)
	}
	if _, err := file.Write(bf16Bytes); err != nil {
		t.Fatalf("write bf16 data: %v", err)
	}
}

func assertFloat32s(t *testing.T, got, want []float32) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len(data) = %d, want %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("data[%d] = %v, want %v", i, got[i], w)
		}
	}
}

func TestSafeTensorsReader_LoadTensorHalf(t *testing.T) {
	path := filepath.Join(t.TempDir(), "half.safetensors")
	writeSafeTensorsHalfFixture(t, path,
		[]uint16{0x3C00, 0x4000, 0x3800, 0xC000}, // F16:  1, 2, 0.5, -2
		[]uint16{0x3F80, 0xBF80, 0x4040},         // BF16: 1, -1, 3
	)

	reader, err := NewSafeTensorsReader(path)
	if err != nil {
		t.Fatalf("NewSafeTensorsReader: %v", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()

	backend := cpu.New()

	t.Run("F16", func(t *testing.T) {
		raw, err := reader.LoadTensor("half", backend)
		if err != nil {
			t.Fatalf("LoadTensor(half): %v", err)
		}
		if raw.DType() != tensor.Float32 {
			t.Errorf("dtype = %v, want %v", raw.DType(), tensor.Float32)
		}
		assertFloat32s(t, raw.AsFloat32(), []float32{1, 2, 0.5, -2})
	})

	t.Run("BF16", func(t *testing.T) {
		raw, err := reader.LoadTensor("brain", backend)
		if err != nil {
			t.Fatalf("LoadTensor(brain): %v", err)
		}
		if raw.DType() != tensor.Float32 {
			t.Errorf("dtype = %v, want %v", raw.DType(), tensor.Float32)
		}
		assertFloat32s(t, raw.AsFloat32(), []float32{1, -1, 3})
	})
}

// TestSafeTensorsReader_LoadTensorHalf_ReadFails checks that the half path
// reports a failed read rather than carrying on into the widener. The dtype
// sends LoadTensor down the half branch and the offsets make the read behind
// it fail, so the error can only come from there.
func TestSafeTensorsReader_LoadTensorHalf_ReadFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "half-corrupt.safetensors")
	writeSafeTensors(t, path, map[string]SafeTensorInfo{
		"half": {DType: SafeTensorsF16, Shape: []int{3}, DataOffsets: [2]int64{0, 1 << 40}},
	}, 6)

	reader, err := NewSafeTensorsReader(path)
	if err != nil {
		t.Fatalf("NewSafeTensorsReader: %v", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()

	_, err = reader.LoadTensor("half", cpu.New())
	if err == nil {
		t.Fatalf("LoadTensor = nil error, want an out-of-range error")
	}
	if !strings.Contains(err.Error(), "out of range") {
		t.Errorf("error = %q, want it to contain %q", err, "out of range")
	}
}

func TestHalfWidener(t *testing.T) {
	tests := []struct {
		dtype    SafeTensorsDType
		wantHalf bool
	}{
		{SafeTensorsF16, true},
		{SafeTensorsBF16, true},
		{SafeTensorsF32, false},
		{SafeTensorsI64, false},
		{SafeTensorsBool, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.dtype), func(t *testing.T) {
			widen, ok := halfWidener(tt.dtype)
			if ok != tt.wantHalf {
				t.Errorf("halfWidener(%s) ok = %v, want %v", tt.dtype, ok, tt.wantHalf)
			}
			// A widener is returned exactly when the dtype is half-precision.
			if (widen != nil) != ok {
				t.Errorf("halfWidener(%s) widen != nil = %v, ok = %v", tt.dtype, widen != nil, ok)
			}
		})
	}
}

func TestLoadHalfAsFloat32Errors(t *testing.T) {
	backend := cpu.New()

	tests := []struct {
		name    string
		shape   tensor.Shape
		data    []byte
		wantErr string
	}{
		{"odd-length", tensor.Shape{1}, []byte{0x00}, "not a multiple of 2"},
		// Two F16 elements but a shape claiming three.
		{"count-mismatch", tensor.Shape{3}, u16leBytes([]uint16{0x3C00, 0x4000}), "does not match shape"},
		// The element count wraps to zero, which the count check above matches
		// against empty data, so the shape only gets rejected when the tensor is
		// allocated. LoadTensor validates the shape before calling in, but the
		// helper does not get to assume that.
		{"overflowing-shape", tensor.Shape{4, 1 << 62}, nil, "overflows"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadHalfAsFloat32("t", tt.shape, tt.data, half.Float16ToFloat32, backend)
			if err == nil {
				t.Fatalf("loadHalfAsFloat32 = nil error, want one containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}
