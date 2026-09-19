package ml

import (
	"fmt"

	"blue/borncgo/optim"
)

// Optimizer wraps a borncgo optimizer over borncgo parameters. Step consumes
// the gradient map from the last Backward and rebinds the parameters in place,
// so the model that owns them sees the update.
type Optimizer struct {
	params []*Param
	inner  optim.Optimizer
}

// SGD builds borncgo's SGD optimizer over the given parameters.
func SGD(params []*Param, lr, momentum float32) (*Optimizer, error) {
	if len(params) == 0 {
		return nil, fmt.Errorf("optimizer: no parameters")
	}
	be := backendForBorn(params[0].Tensor().Device())
	return &Optimizer{params: params, inner: optim.NewSGD(params, optim.SGDConfig{LR: lr, Momentum: momentum}, be)}, nil
}

// Adam builds borncgo's Adam optimizer over the given parameters.
func Adam(params []*Param, lr, beta1, beta2, eps float32) (*Optimizer, error) {
	if len(params) == 0 {
		return nil, fmt.Errorf("optimizer: no parameters")
	}
	be := backendForBorn(params[0].Tensor().Device())
	return &Optimizer{params: params, inner: optim.NewAdam(params, optim.AdamConfig{LR: lr, Betas: [2]float32{beta1, beta2}, Eps: eps}, be)}, nil
}

// Step applies borncgo's update using the gradient map from the last Backward.
func (o *Optimizer) Step() error {
	if lastGrads == nil {
		return fmt.Errorf("optimizer: no gradients (call backward first)")
	}
	o.inner.Step(lastGrads)
	return nil
}

// ZeroGrad clears the parameters' gradients.
func (o *Optimizer) ZeroGrad() {
	o.inner.ZeroGrad()
}
