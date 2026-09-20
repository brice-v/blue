package ops

import "blue/borncgo/internal/tensor"

// MatMulBiasOp is a fused Linear epilogue: output = relu(x @ W^T + bias), with
// x [M, K], W [N, K], bias [N], output [M, N]. Keeping it as a single recorded
// op means the fused forward still autodiffs correctly: the backward below
// reconstructs the relu mask from the output and computes the three gradients.
type MatMulBiasOp struct {
	x, w, bias *tensor.RawTensor
	output     *tensor.RawTensor
	relu       bool
}

// NewMatMulBiasOp creates a fused matmul+bias(+relu) op.
func NewMatMulBiasOp(x, w, bias, output *tensor.RawTensor, relu bool) *MatMulBiasOp {
	return &MatMulBiasOp{x: x, w: w, bias: bias, output: output, relu: relu}
}

// Inputs returns [x, w, bias].
func (op *MatMulBiasOp) Inputs() []*tensor.RawTensor {
	return []*tensor.RawTensor{op.x, op.w, op.bias}
}

// Output returns the fused result.
func (op *MatMulBiasOp) Output() *tensor.RawTensor { return op.output }

// matMulTransposedBackend is implemented by backends that multiply with
// transposed operands without materializing the transpose.
type matMulTransposedBackend interface {
	MatMulTransposed(a, b *tensor.RawTensor, transA, transB bool) *tensor.RawTensor
}

// matMul runs op(a) @ op(b), using the transposed path when available.
func matMul(backend tensor.Backend, a, b *tensor.RawTensor, transA, transB bool) *tensor.RawTensor {
	if tb, ok := backend.(matMulTransposedBackend); ok {
		return tb.MatMulTransposed(a, b, transA, transB)
	}
	x, y := a, b
	if transA {
		x = backend.Transpose(a, 1, 0)
	}
	if transB {
		y = backend.Transpose(b, 1, 0)
	}
	return backend.MatMul(x, y)
}

// Backward computes gradients for x, w, and bias.
func (op *MatMulBiasOp) Backward(outputGrad *tensor.RawTensor, backend tensor.Backend) []*tensor.RawTensor {
	// Through the activation: relu'(z) = 1 where output > 0, else 0.
	dPre := outputGrad
	if op.relu {
		zero, err := tensor.NewRaw(tensor.Shape{1}, tensor.Float32, op.output.Device())
		if err != nil {
			panic(err)
		}
		zero.AsFloat32()[0] = 0
		mask := backend.Cast(backend.Greater(op.output, zero), tensor.Float32)
		dPre = backend.Mul(outputGrad, mask)
	}

	// gradX = dPre @ W; gradW = dPre^T @ x; gradBias = sum over rows of dPre.
	gradX := matMul(backend, dPre, op.w, false, false)
	gradW := matMul(backend, dPre, op.x, true, false)
	gradBias := backend.SumDim(dPre, 0, false)

	return []*tensor.RawTensor{gradX, gradW, gradBias}
}
