package ops

import "blue/borncgo/internal/tensor"

// MatMulOp represents a matrix multiplication operation: output = a @ b.
//
// Backward pass:
//   - d(A@B)/dA = outputGrad @ B^T
//   - d(A@B)/dB = A^T @ outputGrad
//
// Where @ denotes matrix multiplication and ^T denotes transpose.
type MatMulOp struct {
	inputs []*tensor.RawTensor // [a, b]
	output *tensor.RawTensor   // a @ b
}

// NewMatMulOp creates a new MatMulOp.
func NewMatMulOp(a, b, output *tensor.RawTensor) *MatMulOp {
	return &MatMulOp{
		inputs: []*tensor.RawTensor{a, b},
		output: output,
	}
}

// matMulTransposed is implemented by backends that can multiply with transposed
// operands directly, so the matmul backward does not materialize B^T and A^T.
type matMulTransposed interface {
	MatMulTransposed(a, b *tensor.RawTensor, transA, transB bool) *tensor.RawTensor
}

// Backward computes input gradients for matrix multiplication.
func (op *MatMulOp) Backward(outputGrad *tensor.RawTensor, backend tensor.Backend) []*tensor.RawTensor {
	a, b := op.inputs[0], op.inputs[1]

	// grad_a = outputGrad @ b^T, grad_b = a^T @ outputGrad. Prefer the transposed
	// matmul so the backward does not create two extra full-size copies.
	if tb, ok := backend.(matMulTransposed); ok {
		gradA := tb.MatMulTransposed(outputGrad, b, false, true)
		gradB := tb.MatMulTransposed(a, outputGrad, true, false)
		return []*tensor.RawTensor{gradA, gradB}
	}

	// grad_a = outputGrad @ b^T
	bT := backend.Transpose(b, 1, 0)
	gradA := backend.MatMul(outputGrad, bT)

	// grad_b = a^T @ outputGrad
	aT := backend.Transpose(a, 1, 0)
	gradB := backend.MatMul(aT, outputGrad)

	return []*tensor.RawTensor{gradA, gradB}
}

// Inputs returns the input tensors [a, b].
func (op *MatMulOp) Inputs() []*tensor.RawTensor {
	return op.inputs
}

// Output returns the output tensor a @ b.
func (op *MatMulOp) Output() *tensor.RawTensor {
	return op.output
}
