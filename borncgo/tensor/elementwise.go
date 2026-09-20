package tensor

import "blue/borncgo/internal/tensor"

// ElementwiseStep is one op in a fused elementwise chain, shared by the graph
// compiler and by backends that can run such a chain in one kernel.
//
// Slots are positional: a chain's leaves occupy indices 0..NumInputs-1, then
// each step gets the next index. A and B are slot indices and B is -1 for unary
// ops. It lives here so ml and every backend agree on one declaration.
type ElementwiseStep = tensor.ElementwiseStep

// ElementwiseChainBackend is implemented by backends that can evaluate a fused
// elementwise chain in a single kernel. It returns nil when the chain is
// outside the backend's supported subset; callers then run the per-op sequence.
type ElementwiseChainBackend = tensor.ElementwiseChainBackend
