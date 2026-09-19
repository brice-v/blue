package loader

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"blue/borncgo/internal/half"
	"blue/borncgo/internal/tensor"
)

// SafeTensors format:
// [8 bytes: header_size (uint64 LE)]
// [header_size bytes: JSON header]
// [tensor data: raw bytes]

// SafeTensorsDType represents supported SafeTensors data types.
type SafeTensorsDType string

// Supported SafeTensors dtypes.
const (
	SafeTensorsF16  SafeTensorsDType = "F16"
	SafeTensorsF32  SafeTensorsDType = "F32"
	SafeTensorsF64  SafeTensorsDType = "F64"
	SafeTensorsBF16 SafeTensorsDType = "BF16"
	SafeTensorsI32  SafeTensorsDType = "I32"
	SafeTensorsI64  SafeTensorsDType = "I64"
	SafeTensorsU8   SafeTensorsDType = "U8"
	SafeTensorsBool SafeTensorsDType = "BOOL"
)

// SafeTensorInfo describes a tensor in SafeTensors format.
type SafeTensorInfo struct {
	DType       SafeTensorsDType `json:"dtype"`
	Shape       []int            `json:"shape"`
	DataOffsets [2]int64         `json:"data_offsets"` // [start, end]
}

// SafeTensorsHeader is the JSON header in SafeTensors format.
type SafeTensorsHeader struct {
	Metadata map[string]string          `json:"__metadata__"`
	Tensors  map[string]SafeTensorInfo  `json:"-"`
	RawMap   map[string]json.RawMessage `json:"-"`
}

// UnmarshalJSON implements custom JSON unmarshaling for SafeTensorsHeader.
func (h *SafeTensorsHeader) UnmarshalJSON(data []byte) error {
	// First parse as generic map
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return err
	}
	h.RawMap = rawMap

	// Extract metadata
	if metadataRaw, ok := rawMap["__metadata__"]; ok {
		if err := json.Unmarshal(metadataRaw, &h.Metadata); err != nil {
			return fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	// Extract tensors (everything except __metadata__)
	h.Tensors = make(map[string]SafeTensorInfo)
	for key, value := range rawMap {
		if key == "__metadata__" {
			continue
		}
		var info SafeTensorInfo
		if err := json.Unmarshal(value, &info); err != nil {
			return fmt.Errorf("failed to unmarshal tensor %s: %w", key, err)
		}
		h.Tensors[key] = info
	}

	return nil
}

// SafeTensorsReader reads SafeTensors format files.
type SafeTensorsReader struct {
	file       *os.File
	header     SafeTensorsHeader
	headerSize uint64
	dataOffset int64 // Offset where tensor data starts
	fileSize   int64 // Total file size, for bounds-checking header offsets
}

// NewSafeTensorsReader creates a new SafeTensors reader.
func NewSafeTensorsReader(path string) (*SafeTensorsReader, error) {
	//nolint:gosec // G304: File path comes from user input, which is expected for model loading
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Record the real file size, so tensor byte ranges from the (untrusted)
	// header can be bounds-checked against it before any allocation or read.
	// Stat on a descriptor opened a line earlier fails only if that descriptor
	// is revoked underneath us, which no portable test can arrange, so this
	// error arm stays uncovered on purpose.
	fileInfo, err := file.Stat()
	if err != nil {
		_ = file.Close() // Best effort close on error
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Read header size (8 bytes, little-endian uint64)
	var headerSize uint64
	if err := binary.Read(file, binary.LittleEndian, &headerSize); err != nil {
		_ = file.Close() // Best effort close on error
		return nil, fmt.Errorf("failed to read header size: %w", err)
	}

	// Validate header size (should be reasonable, < 100MB)
	if headerSize > 100*1024*1024 {
		_ = file.Close() // Best effort close on error
		return nil, fmt.Errorf("invalid header size: %d (too large)", headerSize)
	}

	// Read header JSON
	headerBytes := make([]byte, headerSize)
	if _, err := io.ReadFull(file, headerBytes); err != nil {
		_ = file.Close() // Best effort close on error
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Parse header
	var header SafeTensorsHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		_ = file.Close() // Best effort close on error
		return nil, fmt.Errorf("failed to parse header JSON: %w", err)
	}

	// Calculate data offset
	dataOffset := int64(8 + headerSize)

	return &SafeTensorsReader{
		file:       file,
		header:     header,
		headerSize: headerSize,
		dataOffset: dataOffset,
		fileSize:   fileInfo.Size(),
	}, nil
}

// Close closes the SafeTensors file.
func (r *SafeTensorsReader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// Metadata returns the metadata map from the header.
func (r *SafeTensorsReader) Metadata() map[string]string {
	return r.header.Metadata
}

// TensorNames returns a list of all tensor names in the file.
func (r *SafeTensorsReader) TensorNames() []string {
	names := make([]string, 0, len(r.header.Tensors))
	for name := range r.header.Tensors {
		names = append(names, name)
	}
	return names
}

// TensorInfo returns information about a specific tensor.
func (r *SafeTensorsReader) TensorInfo(name string) (*SafeTensorInfo, error) {
	info, ok := r.header.Tensors[name]
	if !ok {
		return nil, fmt.Errorf("tensor %s not found", name)
	}
	return &info, nil
}

// ReadTensorData reads raw tensor data for a given tensor name.
func (r *SafeTensorsReader) ReadTensorData(name string) ([]byte, error) {
	info, err := r.TensorInfo(name)
	if err != nil {
		return nil, err
	}

	// The header is untrusted, so validate the byte range it claims before
	// allocating or reading. Offsets are relative to the data section; compare
	// against its length rather than absolute file positions to avoid int64
	// overflow on a malicious end offset.
	start, end := info.DataOffsets[0], info.DataOffsets[1]
	dataLen := r.fileSize - r.dataOffset
	if start < 0 || end < start || end > dataLen {
		return nil, fmt.Errorf("tensor %s: data offsets [%d, %d] out of range for data section of %d bytes",
			name, start, end, dataLen)
	}

	// ReadAt reads from an explicit offset without moving a shared file cursor,
	// so concurrent reads on the same reader cannot interleave. The bounds check
	// above compares against the size sampled at open, so a file that shrinks
	// after that still has to be caught here.
	data := make([]byte, end-start)
	if _, err := r.file.ReadAt(data, r.dataOffset+start); err != nil {
		return nil, fmt.Errorf("failed to read tensor data for %s: %w", name, err)
	}

	return data, nil
}

// safeTensorsDTypeToDataType converts SafeTensors dtype to Born DataType.
func safeTensorsDTypeToDataType(dtype SafeTensorsDType) (tensor.DataType, error) {
	switch dtype {
	case SafeTensorsF32:
		return tensor.Float32, nil
	case SafeTensorsF64:
		return tensor.Float64, nil
	case SafeTensorsI32:
		return tensor.Int32, nil
	case SafeTensorsI64:
		return tensor.Int64, nil
	case SafeTensorsU8:
		return tensor.Uint8, nil
	case SafeTensorsBool:
		return tensor.Bool, nil
	default:
		return 0, fmt.Errorf("unsupported dtype: %s", dtype)
	}
}

// halfWidener returns the function widening a half-precision SafeTensors dtype
// to float32, and reports whether dtype is half-precision at all. Born has no
// native half-precision dtype, so these have no safeTensorsDTypeToDataType
// mapping; this is the only place that decides which dtypes are half.
func halfWidener(dtype SafeTensorsDType) (func(uint16) float32, bool) {
	switch dtype {
	case SafeTensorsF16:
		return half.Float16ToFloat32, true
	case SafeTensorsBF16:
		return half.BFloat16ToFloat32, true
	default:
		return nil, false
	}
}

// LoadTensor loads a tensor from a SafeTensors file into a Born tensor.
// F16 and BF16 tensors are widened to float32, since Born has no native
// half-precision dtype; all other supported dtypes are copied verbatim.
func (r *SafeTensorsReader) LoadTensor(name string, backend tensor.Backend) (*tensor.RawTensor, error) {
	info, err := r.TensorInfo(name)
	if err != nil {
		return nil, err
	}

	// Convert shape.
	shape := tensor.Shape(info.Shape)
	if err := shape.Validate(); err != nil {
		return nil, fmt.Errorf("invalid shape for tensor %s: %w", name, err)
	}

	// F16 and BF16 have no native Born dtype, so they are widened to float32
	// rather than mapped; LoadTensor always yields a directly usable tensor.
	if widen, isHalf := halfWidener(info.DType); isHalf {
		data, err := r.ReadTensorData(name)
		if err != nil {
			return nil, err
		}
		return loadHalfAsFloat32(name, shape, data, widen, backend)
	}

	// Resolve the dtype before reading, so an unsupported one is rejected
	// without pulling the tensor's data off disk.
	dtype, err := safeTensorsDTypeToDataType(info.DType)
	if err != nil {
		return nil, fmt.Errorf("failed to convert dtype for tensor %s: %w", name, err)
	}

	// Read raw data.
	data, err := r.ReadTensorData(name)
	if err != nil {
		return nil, err
	}

	// The header is untrusted: a data section shorter than the shape and dtype
	// demand would otherwise be copied verbatim, silently zero-filling the rest.
	if want := shape.NumElements() * dtype.Size(); len(data) != want {
		return nil, fmt.Errorf("tensor %s: data length %d does not match shape %v (dtype %s, %d bytes)",
			name, len(data), shape, info.DType, want)
	}

	// Create RawTensor.
	raw, err := tensor.NewRaw(shape, dtype, backend.Device())
	if err != nil {
		return nil, fmt.Errorf("failed to create tensor: %w", err)
	}

	// Copy data.
	copy(raw.Data(), data)

	return raw, nil
}

// loadHalfAsFloat32 widens raw half-precision tensor bytes into a float32
// RawTensor, element by element, through widen. Born has no native
// half-precision dtype, so half-precision SafeTensors tensors are converted on
// load. widen comes from halfWidener, which is what decides that a dtype is
// half-precision in the first place.
func loadHalfAsFloat32(name string, shape tensor.Shape, data []byte, widen func(uint16) float32, backend tensor.Backend) (*tensor.RawTensor, error) {
	if len(data)%2 != 0 {
		return nil, fmt.Errorf("tensor %s: half-precision data length %d is not a multiple of 2", name, len(data))
	}

	n := len(data) / 2
	if got := shape.NumElements(); got != n {
		return nil, fmt.Errorf("tensor %s: element count %d does not match shape %v (%d elements)", name, n, shape, got)
	}

	raw, err := tensor.NewRaw(shape, tensor.Float32, backend.Device())
	if err != nil {
		return nil, fmt.Errorf("failed to create tensor: %w", err)
	}

	out := raw.AsFloat32()
	for i := range n {
		out[i] = widen(binary.LittleEndian.Uint16(data[i*2:]))
	}

	return raw, nil
}
