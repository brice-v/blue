# Born ML Framework - Architecture

**Status**: Living Document
**Last Updated**: 2026-06-20

---

## Overview

Born is a layered ML framework following Burn (Rust) architecture with Go adaptations.

```
┌─────────────────────────────────────────────────────┐
│  Application: models/llama, examples/mnist, HRM     │
├─────────────────────────────────────────────────────┤
│  generate: TextGenerator, Sampling, KV-Cache        │
├─────────────────────────────────────────────────────┤
│  nn: Linear, Attention, GQA, SwiGLU, RMSNorm, RoPE │
├─────────────────────────────────────────────────────┤
│  autodiff: Gradient tape, backward ops              │
├─────────────────────────────────────────────────────┤
│  tensor: Tensor[T, B], RawTensor, Shape, Backend    │
├──────────────────┬──────────────────────────────────┤
│  backend/cpu     │  backend/webgpu (Vulkan)         │
└──────────────────┴──────────────────────────────────┘
```

---

## Backend Interface

All tensor operations route through the `Backend` interface (`internal/tensor/backend.go`). CPU and WebGPU backends implement the same interface — switching backends changes *where* computation runs without changing *what* is computed.

```go
type Backend interface {
    // Element-wise: Add, Sub, Mul, Div, MulScalar, ...
    // Matrix: MatMul, BatchMatMul
    // Shape: Reshape, Transpose, Cat, Chunk, Expand
    // Reduction: Sum, SumDim, MeanDim, Argmax
    // Activation: Softmax, SiLU, Sigmoid, ReLU, Tanh
    // Indexing: Gather, Embedding, SelectAdd, ScatterAdd
    // Comparison: Greater, Equal, Where
    Name() string
    Device() Device
}
```

---

## Autodiff (Automatic Differentiation)

### Decorator Pattern

```go
base := cpu.New()                   // or webgpu.New()
backend := autodiff.New(base)       // Wraps any backend with gradient tracking
```

`AutodiffBackend` wraps any `Backend` and records operations on a gradient tape during forward pass. Backward traverses the tape in reverse, computing gradients via chain rule.

### Backward = Forward Ops Composition

All backward operations are implemented by composing forward backend ops. Tensors never leave the compute device during backward:

```
Forward:  x → backend.MatMul → backend.SiLU → backend.Add → loss
Backward: ← backend.Mul ← backend.Sigmoid ← backend.MatMul ← grad
```

This follows the Burn (Rust) architecture where `B::float_matmul()`, `B::float_mul()`, etc. are used in both forward and backward passes.

**No CPU fallback**: Prior to v0.8.2, some backward ops (SiLU, CrossEntropy, Embedding, etc.) read tensor data to CPU via `AsFloat32()` and computed gradients in Go loops. This forced GPU→CPU synchronization and broke the pipeline. v0.8.2 migrated all 7 affected ops to forward composition per [ADR-009](dev/ADR-009-backward-ops-composition.md).

### Gradient Tape

```go
backend.Tape().StartRecording()     // Begin tracking
output := model.Forward(input)      // Operations recorded
grads := autodiff.Backward(output)  // Reverse-mode differentiation
backend.Tape().StopRecording()
```

The tape stops recording during backward to prevent double-recording of gradient computation ops.

---

## Tensor System

### Typed Tensors

```go
type Tensor[T DType, B Backend] struct {
    raw     *RawTensor    // Low-level data
    backend B             // Compute backend
}
```

Generic type parameters provide compile-time safety:
- `T` constrains element type (`float32`, `float64`, `int32`, etc.)
- `B` constrains backend (enables backend-specific optimizations)

### Lazy GPU Evaluation

WebGPU backend uses lazy evaluation — operations queue GPU dispatches without reading results back to CPU. Data transfer only happens when `Data()` or `AsFloat32()` is explicitly called.

```
GPU Op → GPU Op → GPU Op → AsFloat32()
  ↓         ↓         ↓         ↓
[lazy]   [lazy]   [lazy]   [readback]
```

---

## CPU Backend Performance

### Parallel BatchMatMul

Batches are independent — parallelized via `sync.WaitGroup` + goroutines. Threshold: B ≤ 4 sequential, B > 4 parallel. Workers: `runtime.NumCPU()`.

### Cache-Tiled Blocked MatMul

i-block→k-block→j-block loop order keeps both A and B blocks in L1 cache. Block size: 64 (float32, 16KB), 32 (float64, 8KB). 3-5x speedup for matrices > 64×64.

### SIMD Acceleration

Three-layer SIMD architecture with always-on dispatch via `golang.org/x/sys/cpu`:

| Layer | Build | Platform |
|-------|-------|----------|
| **Vendored asm** (avo-generated Plan 9 `.s`) | Default `go build` | amd64 |
| **archsimd** (Go 1.26+ intrinsics) | `GOEXPERIMENT=simd` | amd64 |
| **Scalar fallback** | Always | All platforms |

**AVX2+FMA GEMM micro-kernel** — 6×16 register-blocked (12 YMM accumulators), avo-generated. BirdNET v2.4: 1450ms → 408ms per inference (3.55× end-to-end). Allocation-free via pooled packing scratch (`sync.Pool`).

**Vectorized sigmoid/SiLU** — Cephes expf approximation (~1 ULP). BirdNET sigmoid/exp: 18% → 3% of inference time.

**AVX2 depthwise 3×3** — 9 taps in persistent YMM registers, 11.6× geomean speedup.

**Element-wise arithmetic** — Add, Sub, Mul, Div for float32/float64/int32/int64.

**Tolerance package** (`internal/tolerance`) — Burn-aligned approximate equality (RelAbs, Rel, Abs) with input validation and NaN/Inf handling. Used by all SIMD correctness tests.

**Fuzz tests** — 14 fuzz targets covering all inplace ops with IEEE 754 edge-case seed corpus (NaN, ±Inf, ±0, subnormals, ULP neighbors, min/max).

Detection via `golang.org/x/sys/cpu` (runtime). Function pointer dispatch — zero overhead when SIMD unavailable. ARM64 NEON planned ([#106](https://github.com/born-ml/born/issues/106)).

---

## WebGPU Backend

GPU backend via [gogpu/wgpu](https://github.com/gogpu/wgpu) v0.30.0 — triple-backend architecture:

- **Pure Go** (default): `go build ./...` — zero CGO, zero external dependencies
- **Rust wgpu-native** (WIP): `go build -tags=rust ./...` — Rust wgpu v29 FFI for max performance (runtime FFI incomplete)
- **Browser WASM**: `GOOS=js GOARCH=wasm go build ./...` — native WebGPU in browsers

Born works with all three backends without code changes.

- **Compute shaders**: Hand-written WGSL (not compiler DSL like Burn's CubeCL)
- **Shared encoder** (ADR-012): One `CommandEncoder` for N compute passes. GPU utilization 55→80%.
- **Input buffer cache**: `getOrCreateInputBuffer` — weight matrices uploaded once, reused across forward+backward.
- **Batched dispatch**: Auto-flush every 128 dispatches. Prevents TDR on iGPUs.
- **Vulkan primary**: `BackendsVulkan` for compute workloads.

### GPU Memory Architecture (ADR-015/016/017)

```
┌──────────────────────────────────────────────────────┐
│  TieredPool — 12 log-spaced buckets from Limits()    │
│  ┌──────────┬──────────┬──────────┬───────────────┐  │
│  │ 32KB     │ 74KB     │ 168KB    │ ... │ 256MB   │  │
│  │ Exclusive│ Exclusive│ Exclusive│     │Exclusive │  │
│  │ Pool     │ Pool     │ Pool     │     │Pool      │  │
│  └──────────┴──────────┴──────────┴───────────────┘  │
│  Acquire(size) → first-fit pool → reuse or allocate  │
│  Release(buf) → mark free (pool) or destroy (other)  │
│  Cleanup() → free pages unused for 2+ cycles         │
└──────────────────────────────────────────────────────┘
```

**Buffer lifecycle** (4 release points):
1. **Autodiff tape** — `ClearTape()` releases intermediate outputs
2. **Backward ops** — gradient buffers released after optimizer step
3. **Forward intermediates** — `ReclaimMemory()` releases non-persistent tensors
4. **Optimizer** — moment updates release old buffers

**Tensor.Persist() / Unpersist()** — marks GPU tensors to survive `ReclaimMemory`. Required for any tensor that lives across training steps: carry state, rotary embeddings, model buffers. Without `Persist()`, `ReclaimMemory` destroys all non-parameter tensors with refcount ≤ 1.

### GPU Training Loop

```go
backend := autodiff.New(webgpu.New())

for step := range steps {
    output := model.Forward(input)          // GPU lazy ops (no readback)
    grads := autodiff.Backward(output)      // GPU backward composition
    optimizer.Step(grads)                   // GPU-native Adam
    autodiff.ReleaseGradients(grads)        // Free gradient buffers
    backend.ClearTape()                     // Free tape intermediates
    reclaimer.ReclaimMemory()               // Free non-persistent GPU tensors
}
```

---

## Model Loading

```
GGUF file → gguf.ParseFile → TensorConverter → models/llama.LoadGGUF
                                    ↓
                              Dequantize (Q4_K, Q8_0, F16, ...)
                                    ↓
                              RawTensor → Model weights
```

`models/llama` provides end-to-end LLaMA inference with injectable attention for research experiments.

---

## Key Design Decisions

| Decision | Rationale | ADR |
|---|---|---|
| Backend interface in `internal/tensor` | Avoids import cycle | — |
| Decorator pattern for autodiff | Composable, testable | ADR-004 |
| Hand-written WGSL shaders | Control over GPU kernels | ADR-004 |
| Core API over HAL-direct for wgpu | Stability, portability | ADR-005 |
| Backward via forward composition | GPU-native gradients, Burn alignment | ADR-009 |
| CPU parallel + GPU batching + SIMD | Performance parity with references | ADR-010 |
| Shared encoder + buffer cache | Reduce Submit overhead, GPU utilization | ADR-012 |
| Explicit buffer Release, zero GC | Deterministic GPU memory lifecycle | ADR-015 |
| ExclusivePool (Burn/CubeCL pattern) | Buffer reuse, zero alloc after warmup | ADR-016 |
| TieredPool from device.Limits() | Size-class routing, budget enforcement | ADR-017 |

Full ADR list: `docs/dev/ADR-*.md`

---

## References

- [Burn Framework](https://github.com/tracel-ai/burn) — Rust ML, primary architecture reference
- [gogpu/wgpu](https://github.com/gogpu/wgpu) — Pure Go WebGPU backend
- [ROADMAP.md](../ROADMAP.md) — Version strategy and milestones
- [PHILOSOPHY.md](PHILOSOPHY.md) — Design principles
