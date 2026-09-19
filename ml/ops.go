package ml

import (
	"fmt"
	"slices"
)

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

func Pow(a, b *Tensor) (*Tensor, error) {
	return applyBinary("pow", a, b,
		func(be Backend, a, b *Tensor) (*Tensor, error) { return be.Pow(a, b) },
		func(be Backend, g, out, a, b *Tensor) (*Tensor, *Tensor, error) {
			var da, db *Tensor
			if a.RequiresGrad() {
				// da = g * b * a^(b-1)
				bMinus1, err := be.Sub(b, onesLike(b))
				if err != nil {
					return nil, nil, err
				}
				p, err := be.Pow(a, bMinus1)
				if err != nil {
					return nil, nil, err
				}
				bp, err := be.Mul(b, p)
				if err != nil {
					return nil, nil, err
				}
				da, err = be.Mul(g, bp)
				if err != nil {
					return nil, nil, err
				}
			}
			if b.RequiresGrad() {
				// db = g * out * ln(a); NaN for a <= 0, matching PyTorch
				la, err := be.Log(a)
				if err != nil {
					return nil, nil, err
				}
				o, err := be.Mul(out, la)
				if err != nil {
					return nil, nil, err
				}
				db, err = be.Mul(g, o)
				if err != nil {
					return nil, nil, err
				}
			}
			return da, db, nil
		})
}

func Relu(a *Tensor) (*Tensor, error) {
	return applyUnary("relu", a,
		func(be Backend, a *Tensor) (*Tensor, error) { return be.Relu(a) },
		func(be Backend, g, out, a *Tensor) (*Tensor, error) {
			mask, err := be.Gt(a, zerosLike(a)) // 1 where a > 0, else 0
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

func Neg(a *Tensor) (*Tensor, error) {
	return applyUnary("neg", a,
		func(be Backend, a *Tensor) (*Tensor, error) { return be.Neg(a) },
		func(be Backend, g, out, a *Tensor) (*Tensor, error) { return be.Neg(g) })
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

func Sum(a *Tensor, dims []int, keepdim bool) (*Tensor, error) {
	return reduceDims(a, dims, keepdim, reduceSum)
}

func reduceSum(a *Tensor, dim int, keepdim bool) (*Tensor, error) {
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

func maxDim(a *Tensor, dim int, keepdim bool) (*Tensor, error) {
	return apply("max", []*Tensor{a},
		func(be Backend, in []*Tensor) (*Tensor, error) { return be.Max(in[0], dim, keepdim) },
		func(be Backend, g, out *Tensor, in []*Tensor) ([]*Tensor, error) {
			a := in[0]
			// mask: 1 where a equals the max along dim, else 0
			red := out
			if !keepdim {
				sh := append([]int(nil), a.Shape()...)
				rd, err := resolveDim("max", a.Shape(), dim)
				if err != nil {
					return nil, err
				}
				sh[rd] = 1
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

func Max(a *Tensor, dims []int, keepdim bool) (*Tensor, error) {
	return reduceDims(a, dims, keepdim, maxDim)
}

func Min(a *Tensor, dims []int, keepdim bool) (*Tensor, error) {
	n, err := Neg(a)
	if err != nil {
		return nil, err
	}
	m, err := Max(n, dims, keepdim)
	if err != nil {
		return nil, err
	}
	return Neg(m)
}

func Mean(a *Tensor, dims []int, keepdim bool) (*Tensor, error) {
	s, err := Sum(a, dims, keepdim)
	if err != nil {
		return nil, err
	}
	ds, err := normalizeDims(a.Shape(), dims)
	if err != nil {
		return nil, err
	}
	count := 1
	for _, d := range ds {
		count *= a.Shape()[d]
	}
	if count == 0 {
		return nil, fmt.Errorf("mean: cannot average over an empty dimension")
	}
	return Div(s, scalarLike(s, float32(count)))
}

func ArgMax(a *Tensor, dim int) (*Tensor, error) {
	be, err := backendForAll(a)
	if err != nil {
		return nil, err
	}
	return be.ArgMax(a, dim)
}

func ArgMin(a *Tensor, dim int) (*Tensor, error) {
	be, err := backendForAll(a)
	if err != nil {
		return nil, err
	}
	return be.ArgMin(a, dim)
}

// Abs is relu(x) + relu(-x), so the gradient comes from Relu.
func Abs(a *Tensor) (*Tensor, error) {
	p, err := Relu(a)
	if err != nil {
		return nil, err
	}
	n, err := Neg(a)
	if err != nil {
		return nil, err
	}
	r, err := Relu(n)
	if err != nil {
		return nil, err
	}
	return Add(p, r)
}

// Sigmoid is 1 / (1 + exp(-x)); the chain rule handles the gradient.
func Sigmoid(a *Tensor) (*Tensor, error) {
	n, err := Neg(a)
	if err != nil {
		return nil, err
	}
	e, err := Exp(n)
	if err != nil {
		return nil, err
	}
	denom, err := Add(e, onesLike(a))
	if err != nil {
		return nil, err
	}
	return Div(onesLike(a), denom)
}

// Tanh is 2*sigmoid(2x) - 1.
func Tanh(a *Tensor) (*Tensor, error) {
	two := scalarLike(a, 2)
	x2, err := Mul(a, two)
	if err != nil {
		return nil, err
	}
	s, err := Sigmoid(x2)
	if err != nil {
		return nil, err
	}
	num, err := Mul(s, two)
	if err != nil {
		return nil, err
	}
	return Sub(num, onesLike(a))
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

// compare runs a non-differentiable comparison. It records no graph node, so
// its output is detached even when an input requires grad.
func compare(a, b *Tensor, f func(be Backend, a, b *Tensor) (*Tensor, error)) (*Tensor, error) {
	be, err := backendForAll(a, b)
	if err != nil {
		return nil, err
	}
	return f(be, a, b)
}

// The comparison ops below are non-differentiable and record no graph node.
// Each is a distinct predicate rather than a negation of another, so NaN
// semantics match (NaN compares false with everything except Ne, which is
// true). They exist for masks (Relu uses Gt, Max uses Eq) and for `_eq`/`_ne`/
// `_gt`/`_ge`/`_lt`/`_le`.
func Eq(a, b *Tensor) (*Tensor, error) {
	return compare(a, b, func(be Backend, a, b *Tensor) (*Tensor, error) { return be.Eq(a, b) })
}

func Ne(a, b *Tensor) (*Tensor, error) {
	return compare(a, b, func(be Backend, a, b *Tensor) (*Tensor, error) { return be.Ne(a, b) })
}

func Gt(a, b *Tensor) (*Tensor, error) {
	return compare(a, b, func(be Backend, a, b *Tensor) (*Tensor, error) { return be.Gt(a, b) })
}

func Ge(a, b *Tensor) (*Tensor, error) {
	return compare(a, b, func(be Backend, a, b *Tensor) (*Tensor, error) { return be.Ge(a, b) })
}

func Lt(a, b *Tensor) (*Tensor, error) {
	return compare(a, b, func(be Backend, a, b *Tensor) (*Tensor, error) { return be.Lt(a, b) })
}

func Le(a, b *Tensor) (*Tensor, error) {
	return compare(a, b, func(be Backend, a, b *Tensor) (*Tensor, error) { return be.Le(a, b) })
}

func Permute(a *Tensor, perm ...int) (*Tensor, error) {
	return applyUnary("permute", a,
		func(be Backend, a *Tensor) (*Tensor, error) { return be.Permute(a, perm...) },
		func(be Backend, g, out, a *Tensor) (*Tensor, error) {
			inv := make([]int, len(perm))
			for i, p := range perm {
				rd, err := resolveDim("permute", a.Shape(), p)
				if err != nil {
					return nil, err
				}
				inv[rd] = i
			}
			return be.Permute(a, inv...)
		})
}

func Flatten(a *Tensor) (*Tensor, error) {
	return Reshape(a, a.Numel())
}

func Unsqueeze(a *Tensor, dim int) (*Tensor, error) {
	shape := a.Shape()
	d := dim
	if d < 0 {
		d += len(shape) - 1
	}
	if d < 0 || d >= len(shape) {
		return nil, fmt.Errorf("unsqueeze: dim %d out of range for shape %v", dim, shape)
	}
	out := slices.Concat(shape[:d], []int{1}, shape[d:])
	return Reshape(a, out...)
}

// Squeeze drops size-1 dims. dims nil drops every size-1 dim.
func Squeeze(a *Tensor, dims []int) (*Tensor, error) {
	shape := a.Shape()
	drop := map[int]bool{}
	if dims == nil {
		for i, d := range shape {
			if d == 1 {
				drop[i] = true
			}
		}
	} else {
		for _, d := range dims {
			rd, err := resolveDim("squeeze", shape, d)
			if err != nil {
				return nil, err
			}
			if shape[rd] != 1 {
				return nil, fmt.Errorf("squeeze: dim %d has size %d, not 1", d)
			}
			drop[rd] = true
		}
	}
	out := make([]int, 0, len(shape))
	for i, d := range shape {
		if !drop[i] {
			out = append(out, d)
		}
	}
	return Reshape(a, out...)
}

func BroadcastTo(a *Tensor, shape []int) (*Tensor, error) {
	return applyUnary("broadcast_to", a,
		func(be Backend, a *Tensor) (*Tensor, error) { return broadcastTo(a.materialize(), shape) },
		func(be Backend, g, out, a *Tensor) (*Tensor, error) {
			// sum g back over the axes that were expanded
			d := g
			var err error
			for i := range shape {
				if a.Shape()[i] == 1 && shape[i] != 1 {
					d, err = be.Sum(a, i, true)
					if err != nil {
						return nil, err
					}
				}
			}
			return d, nil
		})
}

// Slice returns a strided view, so it is not tracked (batching data needs no grad).
func Slice(a *Tensor, dim, start, end int) (*Tensor, error) {
	be, err := backendForAll(a)
	if err != nil {
		return nil, err
	}
	return be.Slice(a, dim, start, end)
}

// Clamp is lo + relu(x - lo) - relu(x - hi).
func Clamp(a *Tensor, lo, hi float32) (*Tensor, error) {
	loT, hiT := scalarLike(a, lo), scalarLike(a, hi)
	below, err := Sub(a, loT)
	if err != nil {
		return nil, err
	}
	above, err := Sub(a, hiT)
	if err != nil {
		return nil, err
	}
	rl, err := Relu(below)
	if err != nil {
		return nil, err
	}
	rh, err := Relu(above)
	if err != nil {
		return nil, err
	}
	x, err := Add(loT, rl)
	if err != nil {
		return nil, err
	}
	return Sub(x, rh)
}

// Where is a*cond + b*(1-cond). cond is a detached bool tensor.
func Where(cond, a, b *Tensor) (*Tensor, error) {
	notCond, err := Sub(onesLike(cond), cond)
	if err != nil {
		return nil, err
	}
	left, err := Mul(a, cond)
	if err != nil {
		return nil, err
	}
	right, err := Mul(b, notCond)
	if err != nil {
		return nil, err
	}
	return Add(left, right)
}

// OneHot maps a 1d label tensor to [n, classes]; labels are not differentiable.
func OneHot(labels *Tensor, classes int) (*Tensor, error) {
	be, err := backendForAll(labels)
	if err != nil {
		return nil, err
	}
	return be.OneHot(labels, classes)
}
