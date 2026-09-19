package cpu

import (
	"testing"

	"blue/borncgo/internal/tensor"
)

// ---- Reduce benchmarks ----

// BenchmarkReduce_SumDim_LastDim benchmarks SumDim along the last (inner) dimension.
// This is the hot path in RMSNorm and layer norm patterns.
func BenchmarkReduce_SumDim_LastDim(b *testing.B) {
	backend := New()
	x, _ := tensor.NewRaw(tensor.Shape{64, 128, 512}, tensor.Float32, backend.Device())
	data := x.AsFloat32()
	for i := range data {
		data[i] = float32(i) * 0.001
	}
	b.ResetTimer()
	for range b.N {
		_ = backend.SumDim(x, 2, true)
	}
}

// BenchmarkReduce_SumDim_MiddleDim benchmarks SumDim along a middle dimension.
func BenchmarkReduce_SumDim_MiddleDim(b *testing.B) {
	backend := New()
	x, _ := tensor.NewRaw(tensor.Shape{32, 256, 128}, tensor.Float32, backend.Device())
	data := x.AsFloat32()
	for i := range data {
		data[i] = float32(i) * 0.001
	}
	b.ResetTimer()
	for range b.N {
		_ = backend.SumDim(x, 1, false)
	}
}

// BenchmarkReduce_SumDim_FirstDim benchmarks SumDim along the first (outer) dimension.
func BenchmarkReduce_SumDim_FirstDim(b *testing.B) {
	backend := New()
	x, _ := tensor.NewRaw(tensor.Shape{128, 64, 64}, tensor.Float32, backend.Device())
	data := x.AsFloat32()
	for i := range data {
		data[i] = float32(i) * 0.001
	}
	b.ResetTimer()
	for range b.N {
		_ = backend.SumDim(x, 0, false)
	}
}

// BenchmarkReduce_MeanDim_BirdNET benchmarks MeanDim along the last dim.
// Represents the BirdNET spectrogram normalisation pattern.
func BenchmarkReduce_MeanDim_BirdNET(b *testing.B) {
	backend := New()
	// Typical BirdNET tensor: [batch, time, mel_bands]
	x, _ := tensor.NewRaw(tensor.Shape{8, 256, 128}, tensor.Float32, backend.Device())
	data := x.AsFloat32()
	for i := range data {
		data[i] = float32(i) * 0.001
	}
	b.ResetTimer()
	for range b.N {
		_ = backend.MeanDim(x, 2, true)
	}
}

// ---- ScatterAdd benchmarks ----

// BenchmarkScatterAdd_3D benchmarks ScatterAdd on a typical 3D gather-backward pattern.
func BenchmarkScatterAdd_3D(b *testing.B) {
	backend := New()

	// dest: [32, 128, 64], src/indices: [32, 128, 16] gather along dim=2
	destShape := tensor.Shape{32, 128, 64}
	srcShape := tensor.Shape{32, 128, 16}

	dest, _ := tensor.NewRaw(destShape, tensor.Float32, backend.Device())
	src, _ := tensor.NewRaw(srcShape, tensor.Float32, backend.Device())
	indices, _ := tensor.NewRaw(srcShape, tensor.Int32, backend.Device())

	srcData := src.AsFloat32()
	for i := range srcData {
		srcData[i] = float32(i) * 0.001
	}
	idxData := indices.AsInt32()
	for i := range idxData {
		idxData[i] = int32(i % 64)
	}

	b.ResetTimer()
	for range b.N {
		_ = backend.ScatterAdd(dest, 2, indices, src)
	}
}

// BenchmarkScatterAdd_2D benchmarks ScatterAdd on a 2D embedding gradient scatter.
func BenchmarkScatterAdd_2D(b *testing.B) {
	backend := New()

	// Typical embedding backward: dest [vocab=8192, embed=128], src [seq=512, embed=128]
	destShape := tensor.Shape{8192, 128}
	srcShape := tensor.Shape{512, 128}

	dest, _ := tensor.NewRaw(destShape, tensor.Float32, backend.Device())
	src, _ := tensor.NewRaw(srcShape, tensor.Float32, backend.Device())
	indices, _ := tensor.NewRaw(srcShape, tensor.Int32, backend.Device())

	srcData := src.AsFloat32()
	for i := range srcData {
		srcData[i] = float32(i) * 0.001
	}
	idxData := indices.AsInt32()
	for i := range idxData {
		idxData[i] = int32(i % 8192)
	}

	b.ResetTimer()
	for range b.N {
		_ = backend.ScatterAdd(dest, 0, indices, src)
	}
}

// ---- BatchMatMul broadcast benchmarks ----

// BenchmarkBatchMatMul_Broadcast_SingletonA benchmarks broadcast matmul
// where A has batch=1 broadcast to B's batch dimension.
// This is the attention Q/K/V computation pattern with shared weights.
func BenchmarkBatchMatMul_Broadcast_SingletonA(b *testing.B) {
	backend := New()

	// A: [1, 64, 64] broadcast to B's batch=16 → output [16, 64, 64]
	aData := make([]float32, 64*64)
	for i := range aData {
		aData[i] = 0.01
	}
	a, _ := tensor.FromSlice(aData, tensor.Shape{1, 64, 64}, backend)

	bData := make([]float32, 16*64*64)
	for i := range bData {
		bData[i] = 0.01
	}
	bRaw, _ := tensor.FromSlice(bData, tensor.Shape{16, 64, 64}, backend)

	b.ResetTimer()
	for range b.N {
		_ = backend.BatchMatMul(a.Raw(), bRaw.Raw())
	}
}

// BenchmarkBatchMatMul_Broadcast_BothSides benchmarks broadcast matmul
// where both A and B have singleton batch dims that broadcast against each other.
func BenchmarkBatchMatMul_Broadcast_BothSides(b *testing.B) {
	backend := New()

	// A: [8, 1, 32, 32], B: [1, 4, 32, 32] → output [8, 4, 32, 32]
	aData := make([]float32, 8*32*32)
	for i := range aData {
		aData[i] = 0.01
	}
	a, _ := tensor.FromSlice(aData, tensor.Shape{8, 1, 32, 32}, backend)

	bData := make([]float32, 4*32*32)
	for i := range bData {
		bData[i] = 0.01
	}
	bRaw, _ := tensor.FromSlice(bData, tensor.Shape{1, 4, 32, 32}, backend)

	b.ResetTimer()
	for range b.N {
		_ = backend.BatchMatMul(a.Raw(), bRaw.Raw())
	}
}

// BenchmarkBatchMatMul_Broadcast_MultiHead benchmarks the multi-head attention
// broadcast pattern: [batch, 1, seq, dim] @ [batch, heads, dim, seq].
func BenchmarkBatchMatMul_Broadcast_MultiHead(b *testing.B) {
	backend := New()

	// Q: [2, 1, 64, 64], K: [2, 8, 64, 64] → scores [2, 8, 64, 64]
	qData := make([]float32, 2*64*64)
	for i := range qData {
		qData[i] = 0.01
	}
	q, _ := tensor.FromSlice(qData, tensor.Shape{2, 1, 64, 64}, backend)

	kData := make([]float32, 2*8*64*64)
	for i := range kData {
		kData[i] = 0.01
	}
	k, _ := tensor.FromSlice(kData, tensor.Shape{2, 8, 64, 64}, backend)

	b.ResetTimer()
	for range b.N {
		_ = backend.BatchMatMul(q.Raw(), k.Raw())
	}
}
