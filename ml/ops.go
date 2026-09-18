package ml

// apply is the low-level "autograd key": resolve the backend, run the forward
// kernel, wrap with track. Every op goes through here.
func apply(name string, in []*Tensor,
	fwd func(be Backend, in []*Tensor) (*Tensor, error),
	bwd func(be Backend, g *Tensor, in []*Tensor) []*Tensor,
) (*Tensor, error) {
	be, err := backendForAll(in...)
	if err != nil {
		return nil, err
	}
	out, err := fwd(be, in)
	if err != nil {
		return nil, err
	}
	track(out, name, in, func(g *Tensor) []*Tensor { return bwd(be, g, in) })
	return out, nil
}

func applyUnary(name string, a *Tensor,
	fwd func(be Backend, a *Tensor) (*Tensor, error),
	bwd func(be Backend, g, a *Tensor) *Tensor,
) (*Tensor, error) {
	return apply(name, []*Tensor{a},
		func(be Backend, in []*Tensor) (*Tensor, error) { return fwd(be, in[0]) },
		func(be Backend, g *Tensor, in []*Tensor) []*Tensor { return []*Tensor{bwd(be, g, in[0])} })
}

func applyBinary(name string, a, b *Tensor,
	fwd func(be Backend, a, b *Tensor) (*Tensor, error),
	bwd func(be Backend, g, a, b *Tensor) (*Tensor, *Tensor),
) (*Tensor, error) {
	return apply(name, []*Tensor{a, b},
		func(be Backend, in []*Tensor) (*Tensor, error) { return fwd(be, in[0], in[1]) },
		func(be Backend, g *Tensor, in []*Tensor) []*Tensor {
			da, db := bwd(be, g, in[0], in[1])
			return []*Tensor{da, db}
		})
}

func MatMul(a, b *Tensor) (*Tensor, error) {
	return applyBinary("matmul", a, b,
		func(be Backend, a, b *Tensor) (*Tensor, error) { return be.MatMul(a, b) },
		func(be Backend, g, a, b *Tensor) (*Tensor, *Tensor) {
			bT := must(be.Transpose(b, 0, 1))
			aT := must(be.Transpose(a, 0, 1))
			return must(be.MatMul(g, bT)), must(be.MatMul(aT, g))
		})
}

func Add(a, b *Tensor) (*Tensor, error) {
	return applyBinary("add", a, b,
		func(be Backend, a, b *Tensor) (*Tensor, error) { return be.Add(a, b) },
		func(be Backend, g, a, b *Tensor) (*Tensor, *Tensor) {
			return unbroadcast(be, g, a), unbroadcast(be, g, b)
		})
}

func Sub(a, b *Tensor) (*Tensor, error) {
	return applyBinary("sub", a, b,
		func(be Backend, a, b *Tensor) (*Tensor, error) { return be.Sub(a, b) },
		func(be Backend, g, a, b *Tensor) (*Tensor, *Tensor) {
			return unbroadcast(be, g, a), must(be.Neg(unbroadcast(be, g, b)))
		})
}

func Mul(a, b *Tensor) (*Tensor, error) {
	return applyBinary("mul", a, b,
		func(be Backend, a, b *Tensor) (*Tensor, error) { return be.Mul(a, b) },
		func(be Backend, g, a, b *Tensor) (*Tensor, *Tensor) {
			return must(be.Mul(g, b)), must(be.Mul(g, a))
		})
}

func Div(a, b *Tensor) (*Tensor, error) {
	return applyBinary("div", a, b,
		func(be Backend, a, b *Tensor) (*Tensor, error) { return be.Div(a, b) },
		func(be Backend, g, a, b *Tensor) (*Tensor, *Tensor) {
			da := must(be.Div(g, b))
			db := must(be.Neg(must(be.Div(must(be.Mul(g, a)), must(be.Mul(b, b))))))
			return da, db
		})
}

func Relu(a *Tensor) (*Tensor, error) {
	return applyUnary("relu", a,
		func(be Backend, a *Tensor) (*Tensor, error) { return be.Relu(a) },
		func(be Backend, g, a *Tensor) *Tensor {
			mask := must(be.Greater(a, zerosLike(a)))
			return must(be.Mul(g, mask))
		})
}

func Exp(a *Tensor) (*Tensor, error) {
	return applyUnary("exp", a,
		func(be Backend, a *Tensor) (*Tensor, error) { return be.Exp(a) },
		func(be Backend, g, a *Tensor) *Tensor {
			return must(be.Mul(g, must(be.Exp(a)))) // exp' = exp
		})
}

func Log(a *Tensor) (*Tensor, error) {
	return applyUnary("log", a,
		func(be Backend, a *Tensor) (*Tensor, error) { return be.Log(a) },
		func(be Backend, g, a *Tensor) *Tensor { return must(be.Div(g, a)) })
}

func Sqrt(a *Tensor) (*Tensor, error) {
	return applyUnary("sqrt", a,
		func(be Backend, a *Tensor) (*Tensor, error) { return be.Sqrt(a) },
		func(be Backend, g, a *Tensor) *Tensor {
			two := scalarLike(a, 2)
			return must(be.Div(g, must(be.Mul(two, must(be.Sqrt(a)))))) // 1/(2 sqrt x)
		})
}

func Reshape(a *Tensor, shape ...int) (*Tensor, error) {
	return applyUnary("reshape", a,
		func(be Backend, a *Tensor) (*Tensor, error) { return be.Reshape(a, shape...) },
		func(be Backend, g, a *Tensor) *Tensor { return must(be.Reshape(g, a.Shape()...)) })
}

func Transpose(a *Tensor, dim0, dim1 int) (*Tensor, error) {
	return applyUnary("transpose", a,
		func(be Backend, a *Tensor) (*Tensor, error) { return be.Transpose(a, dim0, dim1) },
		func(be Backend, g, a *Tensor) *Tensor { return must(be.Transpose(g, dim0, dim1)) })
}
