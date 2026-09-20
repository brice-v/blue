//go:build !js

package webgpu

import (
	"fmt"
	"strings"

	"blue/borncgo/internal/tensor"
	wgpu "github.com/oliverbestmann/webgpu/wgpu"
)

// ElementwiseChain evaluates a fused elementwise chain in one dispatch, with a
// WGSL kernel generated from the chain's op sequence. This is the fast path
// behind ml.compile's elementwise fusion: the alternatives are one dispatch and
// one full buffer round trip per op, so a chain of N ops becomes 1 dispatch and
// reads/writes each intermediate N-1 fewer times.
//
// Supported subset: float32 where every leaf has exactly the result's shape
// (no broadcasting and no scalar-shaped leaves) and there is no binding-count
// overflow. Anything else returns nil and the caller runs the per-op sequence,
// so correctness never depends on this kernel.
func (b *Backend) ElementwiseChain(inputs []*tensor.RawTensor, steps []tensor.ElementwiseStep, outShape tensor.Shape, outDType tensor.DataType) *tensor.RawTensor {
	if !b.LazyMode {
		return nil
	}
	if outDType != tensor.Float32 || len(inputs) == 0 || len(steps) == 0 {
		return nil
	}
	// One binding per leaf, plus result and params. WebGPU guarantees at least
	// 4 storage bindings per stage; keep a generous ceiling and fall back above.
	if len(inputs) > 8 {
		return nil
	}
	for _, in := range inputs {
		if in.DType() != tensor.Float32 || !in.Shape().Equal(outShape) {
			return nil
		}
	}
	// The generated kernel carries every intermediate as f32, so any cast whose
	// target is not float32 (or a bool kept as 0/1) would be miscompiled. Decline
	// those and let the eager path run them.
	for _, s := range steps {
		if s.Op == "cast" && s.DType != tensor.Float32 && s.DType != tensor.Bool {
			return nil
		}
	}

	code, err := generateElementwiseWGSL(inputs, steps, outShape)
	if err != nil {
		return nil
	}

	result, err := b.runElementwiseChainLazy(inputs, steps, code, outShape)
	if err != nil {
		return nil
	}
	return result
}

// runElementwiseChainLazy dispatches the generated kernel and returns a lazy
// result, following the same ownership rules as the other lazy ops.
func (b *Backend) runElementwiseChainLazy(inputs []*tensor.RawTensor, steps []tensor.ElementwiseStep, code string, outShape tensor.Shape) (*tensor.RawTensor, error) {
	numElements := outShape.NumElements()
	resultSize := uint64(numElements * 4) //nolint:gosec // G115: dims are small positive ints

	shaderName := "elementwise_chain"
	if sig := elementwiseSignature(steps); len(sig) > 0 {
		shaderName += "_" + sig
	}
	shader := b.compileShader(shaderName, code)

	// Layout: one read-only storage binding per leaf, then the result, then params.
	entries := make([]wgpu.BindGroupLayoutEntry, 0, len(inputs)+2)
	for i := range inputs {
		entries = append(entries, bglStorage(uint32(i), true)) //nolint:gosec // G115: small positive index
	}
	entries = append(entries, bglStorage(uint32(len(inputs)), false)) //nolint:gosec // G115
	entries = append(entries, bglUniform(uint32(len(inputs)+1)))      //nolint:gosec // G115
	entry := b.getOrCreatePipeline(shaderName, shader, entries)

	var transientBufs []*wgpu.Buffer
	var inputLazyDatas []*LazyGPUData
	bufs := make([]bindGroupBuffer, 0, len(inputs)+2)
	for _, in := range inputs {
		ib := b.getOrCreateInputBuffer(in)
		if !ib.cached {
			transientBufs = append(transientBufs, ib.buffer)
		} else if ib.gpuData != nil {
			inputLazyDatas = append(inputLazyDatas, ib.gpuData)
		}
		bufs = append(bufs, bufBinding(ib.buffer, uint64(in.ByteSize()))) //nolint:gosec // G115
	}

	bufferResult, err := b.gpuPool.Acquire(resultSize)
	if err != nil {
		return nil, fmt.Errorf("elementwiseChain: create result buffer: %w", err)
	}
	bufs = append(bufs, bufBinding(bufferResult, resultSize))

	params := make([]byte, 16)
	params[0] = byte(numElements)
	params[1] = byte(numElements >> 8)
	params[2] = byte(numElements >> 16)
	params[3] = byte(numElements >> 24)
	paramsBuffer := b.createUniformBuffer(params)
	bufs = append(bufs, bufBinding(paramsBuffer, 16))

	bg := b.createBindGroupFromBuffers(entry.layout, bufs)

	workgroups := uint32((numElements + workgroupSize - 1) / workgroupSize) //nolint:gosec // G115
	return b.addComputePassToEncoder(entry.pipeline, bg, workgroups, 1, 1, bufferResult, resultSize, outShape, tensor.Float32,
		lazyResources{
			buffers:    append(transientBufs, paramsBuffer),
			bindGroups: []*wgpu.BindGroup{},
			lazyDatas:  inputLazyDatas,
		})
}

// elementwiseSignature names the op sequence, for shader cache reuse.
func elementwiseSignature(steps []tensor.ElementwiseStep) string {
	parts := make([]string, len(steps))
	for i, s := range steps {
		parts[i] = s.Op
	}
	return strings.Join(parts, "_")
}

// generateElementwiseWGSL builds the kernel. Leaves become storage bindings
// read at the element index, then each step is emitted as straight-line WGSL in
// a local, and the final local is written out.
func generateElementwiseWGSL(inputs []*tensor.RawTensor, steps []tensor.ElementwiseStep, outShape tensor.Shape) (string, error) {
	var b strings.Builder
	b.WriteString("@group(0) @binding(0) var<storage, read> in0: array<f32>;\n")
	for i := 1; i < len(inputs); i++ {
		fmt.Fprintf(&b, "@group(0) @binding(%d) var<storage, read> in%d: array<f32>;\n", i, i)
	}
	fmt.Fprintf(&b, "@group(0) @binding(%d) var<storage, read_write> result: array<f32>;\n", len(inputs))
	b.WriteString("struct Params { n: u32 }\n")
	fmt.Fprintf(&b, "@group(0) @binding(%d) var<uniform> params: Params;\n\n", len(inputs)+1)

	// The kernel is 1D over the flattened element count.
	b.WriteString("@compute @workgroup_size(64)\n")
	b.WriteString("fn main(@builtin(global_invocation_id) gid: vec3<u32>) {\n")
	b.WriteString("    let i = gid.x;\n")
	b.WriteString("    if (i >= params.n) { return; }\n")

	// Values are addressed by chain slot: leaves 0..len(inputs)-1, then steps.
	expr := make([]string, len(inputs)+len(steps))
	for i := range inputs {
		expr[i] = fmt.Sprintf("in%d[i]", i)
	}
	for i, s := range steps {
		a := slotExpr(expr, s.A)
		var rhs string
		if s.B >= 0 {
			rhs = slotExpr(expr, s.B)
		} else {
			rhs = ""
		}
		code, err := stepWGSL(s, a, rhs)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "    let v%d = %s;\n", i, code)
		expr[len(inputs)+i] = fmt.Sprintf("v%d", i)
	}
	fmt.Fprintf(&b, "    result[i] = %s;\n", expr[len(expr)-1])
	b.WriteString("}\n")
	return b.String(), nil
}

func slotExpr(expr []string, slot int) string {
	if slot < 0 || slot >= len(expr) {
		return "0.0"
	}
	return expr[slot]
}

func stepWGSL(s tensor.ElementwiseStep, a, b string) (string, error) {
	switch s.Op {
	case "add":
		return fmt.Sprintf("(%s + %s)", a, b), nil
	case "sub":
		return fmt.Sprintf("(%s - %s)", a, b), nil
	case "mul":
		return fmt.Sprintf("(%s * %s)", a, b), nil
	case "div":
		return fmt.Sprintf("(%s / %s)", a, b), nil
	case "exp":
		return fmt.Sprintf("exp(%s)", a), nil
	case "log":
		return fmt.Sprintf("log(%s)", a), nil
	case "sqrt":
		return fmt.Sprintf("sqrt(%s)", a), nil
	case "rsqrt":
		return fmt.Sprintf("inverseSqrt(%s)", a), nil
	case "cos":
		return fmt.Sprintf("cos(%s)", a), nil
	case "sin":
		return fmt.Sprintf("sin(%s)", a), nil
	case "tanh":
		return fmt.Sprintf("tanh(%s)", a), nil
	case "abs":
		return fmt.Sprintf("abs(%s)", a), nil
	case "sign":
		return fmt.Sprintf("sign(%s)", a), nil
	case "relu":
		return fmt.Sprintf("max(%s, 0.0)", a), nil
	case "sigmoid":
		return fmt.Sprintf("(1.0 / (1.0 + exp(-%s)))", a), nil
	case "silu":
		return fmt.Sprintf("(%s / (1.0 + exp(-%s)))", a, a), nil
	case "neg":
		return fmt.Sprintf("(-%s)", a), nil
	case "greater":
		return fmt.Sprintf("select(0.0, 1.0, %s > %s)", a, b), nil
	case "lower":
		return fmt.Sprintf("select(0.0, 1.0, %s < %s)", a, b), nil
	case "greater_equal":
		return fmt.Sprintf("select(0.0, 1.0, %s >= %s)", a, b), nil
	case "lower_equal":
		return fmt.Sprintf("select(0.0, 1.0, %s <= %s)", a, b), nil
	case "equal":
		return fmt.Sprintf("select(0.0, 1.0, %s == %s)", a, b), nil
	case "not_equal":
		return fmt.Sprintf("select(0.0, 1.0, %s != %s)", a, b), nil
	case "cast":
		// The kernel carries every intermediate as f32, so a float32 or a
		// bool-as-0/1 cast is an identity here.
		return a, nil
	case "mul_scalar":
		return fmt.Sprintf("(%s * %s)", a, f32(s.Scalar)), nil
	case "add_scalar":
		return fmt.Sprintf("(%s + %s)", a, f32(s.Scalar)), nil
	case "sub_scalar":
		return fmt.Sprintf("(%s - %s)", a, f32(s.Scalar)), nil
	case "div_scalar":
		return fmt.Sprintf("(%s / %s)", a, f32(s.Scalar)), nil
	case "clamp":
		return fmt.Sprintf("clamp(%s, %s, %s)", a, f32(s.Min), f32(s.Max)), nil
	case "erf":
		// WGSL has no erf; approximate through the identity used by the CPU path
		// is not worth it. Decline so the caller falls back to eager.
		return "", fmt.Errorf("elementwise: erf has no WGSL builtin")
	default:
		return "", fmt.Errorf("elementwise: unsupported op %q", s.Op)
	}
}

// f32 formats a float so WGSL reads it as a float literal.
func f32(v float32) string {
	s := fmt.Sprintf("%v", v)
	if !strings.ContainsAny(s, ".eE") {
		s += ".0"
	}
	return s
}
