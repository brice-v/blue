package ml

import (
	"math"
	"testing"

	"blue/borncgo/tensor"
)

func mustTensor(t *testing.T, data []float32, shape []int) *Tensor {
	t.Helper()
	tt, err := NewTensor(data, shape, Float32, CPU)
	if err != nil {
		t.Fatalf("NewTensor: %v", err)
	}
	return tt
}

func closeData(a, b []float32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if math.Abs(float64(a[i]-b[i])) > 1e-5 {
			return false
		}
	}
	return true
}

// TestCompiledForwardMatchesEager captures a small module stack, optimises it,
// and checks the replayed result equals the eager one. The Linear layers use the
// transpose + matmul + add pattern, so the fusion pass should fire.
func TestCompiledForwardMatchesEager(t *testing.T) {
	l1, err := NNLinear(4, 5, CPU)
	if err != nil {
		t.Fatalf("NNLinear: %v", err)
	}
	l2, err := NNLinear(5, 2, CPU)
	if err != nil {
		t.Fatalf("NNLinear: %v", err)
	}
	fwd := func(x *Tensor) *Tensor {
		h := NNForward(l1, x)
		h, _ = Relu(h)
		return NNForward(l2, h)
	}

	ex := mustTensor(t, []float32{
		0.1, 0.2, 0.3, 0.4,
		-0.5, 0.6, -0.7, 0.8,
		0.9, -1.0, 1.1, -1.2,
	}, []int{3, 4})
	want := fwd(ex).ContiguousData()

	tr, err := NewTracer(ex)
	if err != nil {
		t.Fatalf("NewTracer: %v", err)
	}
	out := fwd(tr.InputTensor())
	g, err := tr.Finish(out)
	if err != nil {
		t.Fatalf("Finish: %v", err)
	}
	g.Optimize()
	if g.fused == 0 {
		t.Fatalf("expected the fusion pass to fire, stats=%s", g.Stats())
	}
	raw, err := g.Run(ex.be, ex.t.Raw())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	got := wrapRaw(ex.be, raw).ContiguousData()
	if !closeData(want, got) {
		t.Fatalf("compiled forward differs from eager:\n want %v\n got  %v", want, got)
	}

	// A graph captured for one shape must not silently accept another.
	bad := mustTensor(t, make([]float32, 8), []int{2, 4})
	if _, err := g.Run(ex.be, bad.t.Raw()); err == nil {
		t.Fatal("expected replay to reject a different input shape")
	}
}

// TestCompiledBackwardMatchesEager checks that a replayed graph still
// differentiates to the same gradients as the eager graph.
func TestCompiledBackwardMatchesEager(t *testing.T) {
	w1 := mustTensor(t, []float32{0.1, -0.2, 0.3, 0.4, 0.5, -0.6, 0.7, 0.8, -0.9, 1.0, 0.2, -0.3, 0.4, 0.5, -0.6}, []int{3, 5})
	b1 := mustTensor(t, []float32{0.1, -0.1, 0.2, -0.2, 0.3}, []int{5})
	w2 := mustTensor(t, []float32{0.5, -0.5, 0.25, 0.75, -0.25, 0.6, -0.6, 0.1, 0.2, 0.3}, []int{5, 2})
	b2 := mustTensor(t, []float32{0.01, -0.02}, []int{2})
	for _, p := range []*Tensor{w1, b1, w2, b2} {
		p.SetRequiresGrad(true)
	}
	fwd := func(x *Tensor) *Tensor {
		h, _ := MatMul(x, w1)
		h, _ = Add(h, b1)
		h, _ = Relu(h)
		h, _ = MatMul(h, w2)
		return mustAdd(t, h, b2)
	}
	ex := mustTensor(t, []float32{0.2, -0.4, 0.6, 0.8, -1.0, 1.2, 0.3, -0.5, 0.7, -0.9, 1.1, -1.3}, []int{4, 3})

	// Eager gradients.
	sum, err := Sum(fwd(ex), nil, false)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if err := sum.Backward(); err != nil {
		t.Fatalf("eager Backward: %v", err)
	}
	eager := map[*Tensor][]float32{
		w1: gradData(w1), b1: gradData(b1), w2: gradData(w2), b2: gradData(b2),
	}
	for _, p := range []*Tensor{w1, b1, w2, b2} {
		p.ZeroGrad()
	}

	// Compiled gradients.
	tr, err := NewTracer(ex)
	if err != nil {
		t.Fatalf("NewTracer: %v", err)
	}
	g, err := tr.Finish(fwd(tr.InputTensor()))
	if err != nil {
		t.Fatalf("Finish: %v", err)
	}
	g.Optimize()
	raw, err := g.Run(ex.be, ex.t.Raw())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	sum2, err := Sum(wrapRaw(ex.be, raw), nil, false)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if err := sum2.Backward(); err != nil {
		t.Fatalf("compiled Backward: %v", err)
	}
	for _, p := range []*Tensor{w1, b1, w2, b2} {
		if !closeData(eager[p], gradData(p)) {
			t.Errorf("gradient mismatch for a leaf:\n eager %v\n got   %v", eager[p], gradData(p))
		}
	}
}

func mustAdd(t *testing.T, a, b *Tensor) *Tensor {
	t.Helper()
	out, err := Add(a, b)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	return out
}

func gradData(t *Tensor) []float32 {
	if t.Grad() == nil {
		return nil
	}
	return t.Grad().ContiguousData()
}

// TestFuseMatMulBiasRewrite builds the transpose -> matmul -> add -> relu graph
// directly and checks the rewrite fuses it and still replays the right values.
func TestFuseMatMulBiasRewrite(t *testing.T) {
	x := mustTensor(t, []float32{1, 2, 3, 4, 5, 6}, []int{2, 3})
	w := mustTensor(t, []float32{1, -1, 0.5, 2, -2, 1, 3, 0.5, -0.5, -1, 1, 2}, []int{4, 3})
	b := mustTensor(t, []float32{0.5, -0.5, 0.25, -0.25}, []int{4})

	g := &Graph{device: tensor.CPU}
	g.input = len(g.values)
	g.values = append(g.values, graphValue{shape: tensor.Shape(x.Shape()), dtype: tensor.Float32, device: tensor.CPU, node: -1, input: true})
	wID := len(g.values)
	g.values = append(g.values, graphValue{shape: tensor.Shape(w.Shape()), dtype: tensor.Float32, device: tensor.CPU, node: -1, raw: w.Raw()})
	bID := len(g.values)
	g.values = append(g.values, graphValue{shape: tensor.Shape(b.Shape()), dtype: tensor.Float32, device: tensor.CPU, node: -1, raw: b.Raw()})

	addNode := func(op string, ins []int, outShape tensor.Shape, cfg func(*graphNode)) int {
		out := len(g.values)
		g.values = append(g.values, graphValue{shape: outShape, dtype: tensor.Float32, device: tensor.CPU, node: len(g.nodes)})
		n := graphNode{op: op, ins: ins, out: out}
		if cfg != nil {
			cfg(&n)
		}
		g.nodes = append(g.nodes, n)
		return out
	}
	tID := addNode("transpose", []int{wID}, tensor.Shape{3, 4}, func(n *graphNode) { n.axes = []int{1, 0} })
	mmID := addNode("matmul", []int{g.input, tID}, tensor.Shape{2, 4}, nil)
	rID := addNode("reshape", []int{bID}, tensor.Shape{1, 4}, func(n *graphNode) { n.shape = tensor.Shape{1, 4} })
	aID := addNode("add", []int{mmID, rID}, tensor.Shape{2, 4}, nil)
	g.output = addNode("relu", []int{aID}, tensor.Shape{2, 4}, nil)

	g.Optimize()
	if g.fused < 1 {
		t.Fatalf("expected matmul_bias fusion, stats=%s", g.Stats())
	}
	// The whole transpose/matmul/add/relu chain collapses to one fused node.
	live := 0
	var only *graphNode
	for i := range g.nodes {
		if !g.nodes[i].dead {
			live++
			only = &g.nodes[i]
		}
	}
	if live != 1 || only.op != "matmul_bias" || !only.relu {
		t.Fatalf("expected a single fused matmul_bias(relu) node, stats=%s", g.Stats())
	}

	wT, err := Transpose(w, 0, 1)
	if err != nil {
		t.Fatalf("Transpose: %v", err)
	}
	h, err := MatMul(x, wT)
	if err != nil {
		t.Fatalf("MatMul: %v", err)
	}
	h, err = Add(h, b)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	want, err := Relu(h)
	if err != nil {
		t.Fatalf("Relu: %v", err)
	}

	raw, err := g.Run(x.be, x.t.Raw())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	got := wrapRaw(x.be, raw).ContiguousData()
	if !closeData(want.ContiguousData(), got) {
		t.Fatalf("fused rewrite differs from eager:\n want %v\n got  %v", want.ContiguousData(), got)
	}
}
