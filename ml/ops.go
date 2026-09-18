package ml

// apply is the low-level "autograd key": resolve the backend, run the forward
// kernel, wrap with track. The backward also receives the forward output `out`,
// which several gradients need (exp, sigmoid, softmax, max).
func apply(name string, in []*Tensor,
	fwd func(be Backend, in []*Tensor) (*Tensor, error),
	bwd func(be Backend, g, out *Tensor, in []*Tensor) ([]*Tensor, error),
) (*Tensor, error) {
	be, err := backendForAll(in...)
	if err != nil {
		return nil, err
	}
	out, err := fwd(be, in)
	if err != nil {
		return nil, err
	}
	track(out, name, in, func(g *Tensor) ([]*Tensor, error) {
		return bwd(be, g, out, in)
	})
	return out, nil
}

func applyUnary(name string, a *Tensor,
	fwd func(be Backend, a *Tensor) (*Tensor, error),
	bwd func(be Backend, g, out, a *Tensor) (*Tensor, error),
) (*Tensor, error) {
	return apply(name, []*Tensor{a},
		func(be Backend, in []*Tensor) (*Tensor, error) { return fwd(be, in[0]) },
		func(be Backend, g, out *Tensor, in []*Tensor) ([]*Tensor, error) {
			d, err := bwd(be, g, out, in[0])
			if err != nil {
				return nil, err
			}
			return []*Tensor{d}, nil
		})
}

func applyBinary(name string, a, b *Tensor,
	fwd func(be Backend, a, b *Tensor) (*Tensor, error),
	bwd func(be Backend, g, out, a, b *Tensor) (*Tensor, *Tensor, error),
) (*Tensor, error) {
	return apply(name, []*Tensor{a, b},
		func(be Backend, in []*Tensor) (*Tensor, error) { return fwd(be, in[0], in[1]) },
		func(be Backend, g, out *Tensor, in []*Tensor) ([]*Tensor, error) {
			da, db, err := bwd(be, g, out, in[0], in[1])
			if err != nil {
				return nil, err
			}
			return []*Tensor{da, db}, nil
		})
}

func MatMul(a, b *Tensor) (*Tensor, error) {
	return applyBinary("matmul", a, b,
		func(be Backend, a, b *Tensor) (*Tensor, error) { return be.MatMul(a, b) },
		func(be Backend, g, out, a, b *Tensor) (*Tensor, *Tensor, error) {
			// dA = g @ b^T, dB = a^T @ g
			bT, err := be.Transpose(b, 0, 1)
			if err != nil {
				return nil, nil, err
			}
			aT, err := be.Transpose(a, 0, 1)
			if err != nil {
				return nil, nil, err
			}
			dA, err := be.MatMul(g, bT)
			if err != nil {
				return nil, nil, err
			}
			dB, err := be.MatMul(aT, g)
			if err != nil {
				return nil, nil, err
			}
			return dA, dB, nil
		})
}

func Add(a, b *Tensor) (*Tensor, error) {
	return applyBinary("add", a, b,
		func(be Backend, a, b *Tensor) (*Tensor, error) { return be.Add(a, b) },
		func(be Backend, g, out, a, b *Tensor) (*Tensor, *Tensor, error) {
			da, err := unbroadcast(be, g, a)
			if err != nil {
				return nil, nil, err
			}
			db, err := unbroadcast(be, g, b)
			if err != nil {
				return nil, nil, err
			}
			return da, db, nil
		})
}

func Sub(a, b *Tensor) (*Tensor, error) {
	return applyBinary("sub", a, b,
		func(be Backend, a, b *Tensor) (*Tensor, error) { return be.Sub(a, b) },
		func(be Backend, g, out, a, b *Tensor) (*Tensor, *Tensor, error) {
			da, err := unbroadcast(be, g, a)
			if err != nil {
				return nil, nil, err
			}
			db, err := unbroadcast(be, g, b)
			if err != nil {
				return nil, nil, err
			}
			neg, err := be.Neg(db)
			if err != nil {
				return nil, nil, err
			}
			return da, neg, nil
		})
}

func Mul(a, b *Tensor) (*Tensor, error) {
	return applyBinary("mul", a, b,
		func(be Backend, a, b *Tensor) (*Tensor, error) { return be.Mul(a, b) },
		func(be Backend, g, out, a, b *Tensor) (*Tensor, *Tensor, error) {
			da, err := be.Mul(g, b) // dz/da = b
			if err != nil {
				return nil, nil, err
			}
			db, err := be.Mul(g, a) // dz/db = a
			if err != nil {
				return nil, nil, err
			}
			return da, db, nil
		})
}

func Div(a, b *Tensor) (*Tensor, error) {
	return applyBinary("div", a, b,
		func(be Backend, a, b *Tensor) (*Tensor, error) { return be.Div(a, b) },
		func(be Backend, g, out, a, b *Tensor) (*Tensor, *Tensor, error) {
			// da = g / b, db = -g * a / b^2
			da, err := be.Div(g, b)
			if err != nil {
				return nil, nil, err
			}
			b2, err := be.Mul(b, b)
			if err != nil {
				return nil, nil, err
			}
			num, err := be.Mul(g, a)
			if err != nil {
				return nil, nil, err
			}
			q, err := be.Div(num, b2)
			if err != nil {
				return nil, nil, err
			}
			db, err := be.Neg(q)
			if err != nil {
				return nil, nil, err
			}
			return da, db, nil
		})
}

func Relu(a *Tensor) (*Tensor, error) {
	return applyUnary("relu", a,
		func(be Backend, a *Tensor) (*Tensor, error) { return be.Relu(a) },
		func(be Backend, g, out, a *Tensor) (*Tensor, error) {
			mask, err := be.Greater(a, zerosLike(a)) // 1 where a > 0, else 0
			if err != nil {
				return nil, err
			}
			return be.Mul(g, mask)
		})
}

func Exp(a *Tensor) (*Tensor, error) {
	return applyUnary("exp", a,
		func(be Backend, a *Tensor) (*Tensor, error) { return be.Exp(a) },
		func(be Backend, g, out, a *Tensor) (*Tensor, error) {
			return be.Mul(g, out) // exp' = exp, already computed
		})
}

func Log(a *Tensor) (*Tensor, error) {
	return applyUnary("log", a,
		func(be Backend, a *Tensor) (*Tensor, error) { return be.Log(a) },
		func(be Backend, g, out, a *Tensor) (*Tensor, error) { return be.Div(g, a) })
}

func Sqrt(a *Tensor) (*Tensor, error) {
	return applyUnary("sqrt", a,
		func(be Backend, a *Tensor) (*Tensor, error) { return be.Sqrt(a) },
		func(be Backend, g, out, a *Tensor) (*Tensor, error) {
			// 1/(2 sqrt x), reusing the output
			two := scalarLike(a, 2)
			den, err := be.Mul(two, out)
			if err != nil {
				return nil, err
			}
			return be.Div(g, den)
		})
}

func Transpose(a *Tensor, dim0, dim1 int) (*Tensor, error) {
	return applyUnary("transpose", a,
		func(be Backend, a *Tensor) (*Tensor, error) { return be.Transpose(a, dim0, dim1) },
		func(be Backend, g, out, a *Tensor) (*Tensor, error) {
			return be.Transpose(g, dim0, dim1) // a swap is its own inverse
		})
}

func Reshape(a *Tensor, shape ...int) (*Tensor, error) {
	return applyUnary("reshape", a,
		func(be Backend, a *Tensor) (*Tensor, error) { return be.Reshape(a, shape...) },
		func(be Backend, g, out, a *Tensor) (*Tensor, error) { return be.Reshape(g, a.Shape()...) })
}

func Sum(a *Tensor, dim int, keepdim bool) (*Tensor, error) {
	return apply("sum", []*Tensor{a},
		func(be Backend, in []*Tensor) (*Tensor, error) { return be.Sum(in[0], dim, keepdim) },
		func(be Backend, g, out *Tensor, in []*Tensor) ([]*Tensor, error) {
			// every element that was summed gets the same incoming gradient
			d, err := expandToDim(be, g, in[0], dim, keepdim)
			if err != nil {
				return nil, err
			}
			return []*Tensor{d}, nil
		})
}

func Max(a *Tensor, dim int, keepdim bool) (*Tensor, error) {
	return apply("max", []*Tensor{a},
		func(be Backend, in []*Tensor) (*Tensor, error) { return be.Max(in[0], dim, keepdim) },
		func(be Backend, g, out *Tensor, in []*Tensor) ([]*Tensor, error) {
			a := in[0]
			// mask: 1 where a equals the max along dim, else 0
			red := out
			if !keepdim {
				sh := append([]int(nil), a.Shape()...)
				sh[resolveDim(a.Shape(), dim)] = 1
				var err error
				red, err = be.Reshape(out, sh...)
				if err != nil {
					return nil, err
				}
			}
			broad, err := broadcastTo(red, a.Shape())
			if err != nil {
				return nil, err
			}
			mask, err := be.Eq(a, broad.materialize())
			if err != nil {
				return nil, err
			}
			exp, err := expandToDim(be, g, a, dim, keepdim)
			if err != nil {
				return nil, err
			}
			d, err := be.Mul(exp, mask)
			if err != nil {
				return nil, err
			}
			return []*Tensor{d}, nil
		})
}

func Softmax(a *Tensor, dim int) (*Tensor, error) {
	return apply("softmax", []*Tensor{a},
		func(be Backend, in []*Tensor) (*Tensor, error) { return be.Softmax(in[0], dim) },
		func(be Backend, g, out *Tensor, in []*Tensor) ([]*Tensor, error) {
			a := in[0]
			// dA = out * (g - sum(g * out, dim, keepdim=true))
			gy, err := be.Mul(g, out)
			if err != nil {
				return nil, err
			}
			s, err := be.Sum(gy, dim, true)
			if err != nil {
				return nil, err
			}
			broad, err := broadcastTo(s, a.Shape())
			if err != nil {
				return nil, err
			}
			diff, err := be.Sub(g, broad.materialize())
			if err != nil {
				return nil, err
			}
			d, err := be.Mul(out, diff)
			if err != nil {
				return nil, err
			}
			return []*Tensor{d}, nil
		})
}

// Eq is a comparison, so it is not differentiable and records no graph node.
// It exists for masks (for example Max's backward) and for `_eq` later.
func Eq(a, b *Tensor) (*Tensor, error) {
	be, err := backendForAll(a, b)
	if err != nil {
		return nil, err
	}
	return be.Eq(a, b)
}
