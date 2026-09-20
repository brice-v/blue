package ml

import (
	"fmt"
	"math"

	"blue/borncgo/nn"
	"blue/borncgo/optim"
	"blue/borncgo/tensor"
)

// OptimBinding pairs a borncgo parameter with the blue tensor it updates.
//
// Two cases exist:
//   - A raw blue tensor (a leaf the user owns). It is wrapped in a borncgo
//     parameter, and after Step the tensor's storage is rebound to the
//     parameter's new tensor, so every holder of the blue tensor sees the
//     update.
//   - A module-owned parameter. The module rebinds it, so Tensor is nil.
type OptimBinding struct {
	Param  *Param
	Tensor *Tensor
}

// BindTensor wraps a blue tensor so it can be optimized directly, without an
// nn module. This is what makes custom models trainable.
func BindTensor(name string, t *Tensor) OptimBinding {
	return OptimBinding{Param: nn.NewParameter(name, t.t), Tensor: t}
}

// BindParam binds a module-owned borncgo parameter.
func BindParam(p *Param) OptimBinding {
	return OptimBinding{Param: p}
}

// Optimizer wraps a borncgo optimizer over bound parameters. The backend is
// kept so Step can read the gradient map that the matching backward produced.
type Optimizer struct {
	bindings []OptimBinding
	inner    optim.Optimizer
	backend  tensor.Backend
}

func paramsOf(bindings []OptimBinding) []*Param {
	params := make([]*Param, len(bindings))
	for i, b := range bindings {
		params[i] = b.Param
	}
	return params
}

// SGD builds SGD over the bound parameters. weightDecay is L2.
func SGD(bindings []OptimBinding, lr, momentum, weightDecay float32) (*Optimizer, error) {
	if len(bindings) == 0 {
		return nil, fmt.Errorf("optimizer: no parameters")
	}
	be := backendForBorn(bindings[0].Param.Tensor().Device())
	return &Optimizer{
		bindings: bindings,
		inner:    optim.NewSGD(paramsOf(bindings), optim.SGDConfig{LR: lr, Momentum: momentum, WeightDecay: weightDecay}, be),
		backend:  be,
	}, nil
}

// Adam builds Adam over the bound parameters. weightDecay is decoupled (AdamW
// style) in the engine, so a non-zero value gives AdamW behavior.
func Adam(bindings []OptimBinding, lr, beta1, beta2, eps, weightDecay float32) (*Optimizer, error) {
	if len(bindings) == 0 {
		return nil, fmt.Errorf("optimizer: no parameters")
	}
	be := backendForBorn(bindings[0].Param.Tensor().Device())
	return &Optimizer{
		bindings: bindings,
		inner: optim.NewAdam(paramsOf(bindings), optim.AdamConfig{
			LR:          lr,
			Betas:       [2]float32{beta1, beta2},
			Eps:         eps,
			WeightDecay: weightDecay,
		}, be),
		backend: be,
	}, nil
}

// Step applies the update using the gradient map from the last Backward, then
// rebinds every raw tensor to its parameter's new storage.
func (o *Optimizer) Step() error {
	grads := lastGradsFor(o.backend)
	if grads == nil {
		return fmt.Errorf("optimizer: no gradients (call backward first)")
	}
	o.inner.Step(grads)
	for _, b := range o.bindings {
		if b.Tensor != nil {
			b.Tensor.t = b.Param.Tensor()
		}
	}
	return nil
}

// ZeroGrad clears the parameters' gradients.
func (o *Optimizer) ZeroGrad() {
	o.inner.ZeroGrad()
}

// ClipGradNorm scales the gradients from the last backward in place so their
// global L2 norm over the bound parameters is at most maxNorm, matching
// torch.nn.utils.clip_grad_norm_. It returns the norm before clipping. Call it
// after Backward and before Step.
func ClipGradNorm(bindings []OptimBinding, maxNorm float32) (float32, error) {
	if len(bindings) == 0 {
		return 0, fmt.Errorf("clip_grad_norm: no parameters")
	}
	be := bindings[0].Param.Tensor().Backend()
	grads := lastGradsFor(be)
	if grads == nil {
		return 0, fmt.Errorf("clip_grad_norm: no gradients (call backward first)")
	}
	gradOf := func(b OptimBinding) *tensor.RawTensor {
		return grads[b.Param.Tensor().Raw()]
	}
	var total float64
	for _, b := range bindings {
		if g := gradOf(b); g != nil {
			for _, v := range g.AsFloat32() {
				total += float64(v) * float64(v)
			}
		}
	}
	norm := float32(math.Sqrt(total))
	if norm == 0 || norm <= maxNorm {
		return norm, nil
	}
	scale := maxNorm / norm
	for _, b := range bindings {
		if g := gradOf(b); g != nil {
			s := g.AsFloat32()
			for i := range s {
				s[i] *= scale
			}
		}
	}
	return norm, nil
}

// LRSchedule computes a learning rate from a step, keeping scheduling as data
// rather than state on the optimizer. kind is one of:
//
//	constant: base
//	step:     base * gamma^(step/stepSize), with a = stepSize, b = gamma
//	cosine:   minLr + 0.5*(base-minLr)*(1+cos(pi*step/total)), a = total, b = minLr
//	warmup:   base * min(1, (step+1)/warmupSteps), with a = warmupSteps
func LRSchedule(kind string, base float32, step int, a, b float32) (float32, error) {
	switch kind {
	case "constant":
		return base, nil
	case "step":
		stepSize := int(a)
		gamma := b
		if stepSize <= 0 {
			return 0, fmt.Errorf("lr_schedule: step_size must be positive")
		}
		// Integer division, matching torch.optim.lr_scheduler.StepLR.
		exp := step / stepSize
		return base * float32(math.Pow(float64(gamma), float64(exp))), nil
	case "cosine":
		total := a
		minLr := b
		if total <= 0 {
			return 0, fmt.Errorf("lr_schedule: total_steps must be positive")
		}
		t := float64(step) / float64(total)
		if t > 1 {
			t = 1
		}
		return minLr + 0.5*(base-minLr)*(1+float32(math.Cos(math.Pi*t))), nil
	case "warmup":
		warmup := a
		if warmup <= 0 {
			return base, nil
		}
		if float32(step) >= warmup {
			return base, nil
		}
		return base * (float32(step) + 1) / warmup, nil
	default:
		return 0, fmt.Errorf("lr_schedule: unknown schedule %q", kind)
	}
}

// lrSetter is implemented by the engine optimizers (SGD, Adam) but is not part
// of the Optimizer interface, so it is reached by type assertion.
type lrSetter interface {
	SetLR(float32)
}

// SetLR updates the learning rate, for a schedule driven from blue.
func (o *Optimizer) SetLR(lr float32) {
	if s, ok := o.inner.(lrSetter); ok {
		s.SetLR(lr)
	}
}

// LR returns the current learning rate.
func (o *Optimizer) LR() float32 {
	return o.inner.GetLR()
}
