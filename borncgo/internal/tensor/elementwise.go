package tensor

// ElementwiseStep is one op in a fused elementwise chain, shared by the graph
// compiler and by backends that can run such a chain in one kernel.
//
// Slots are positional: a chain's leaves occupy indices 0..NumInputs-1, then
// each step gets the next index. A and B are slot indices and B is -1 for unary
// ops. DType is the op's result dtype, which matters for a cast step.
type ElementwiseStep struct {
	Op     string
	A      int
	B      int
	Scalar float32
	Min    float32
	Max    float32
	DType  DataType
}

// ElementwiseChainBackend is implemented by backends that can evaluate a fused
// elementwise chain in a single kernel. It returns nil when the chain is
// outside the backend's supported subset; callers then run the per-op sequence.
type ElementwiseChainBackend interface {
	ElementwiseChain(inputs []*RawTensor, steps []ElementwiseStep, outShape Shape, outDType DataType) *RawTensor
}
