package ml

import (
	"blue/borncgo/nn"
	"blue/borncgo/tensor"
)

// Module is borncgo's module interface specialised to the engine backend type.
type Module = nn.Module[tensor.Backend]

// Param is a borncgo parameter handle used by the optimizer bridge.
type Param = nn.Parameter[tensor.Backend]

// NNLinear builds a borncgo Linear layer (weight [out, in], Xavier init) on the
// requested device.
func NNLinear(inFeatures, outFeatures int, device Device) (Module, error) {
	be, err := backendFor(device)
	if err != nil {
		return nil, err
	}
	return nn.NewLinear(inFeatures, outFeatures, be), nil
}

func NNReLU() Module    { return nn.NewReLU[tensor.Backend]() }
func NNSigmoid() Module { return nn.NewSigmoid[tensor.Backend]() }

// NNForward runs a module on a blue tensor.
func NNForward(m Module, x *Tensor) *Tensor {
	return wrap(m.Forward(x.t))
}

// NNParameters returns the module's borncgo parameters.
func NNParameters(m Module) []*Param {
	return m.Parameters()
}

// NNTo moves a module's parameters to the requested device, like PyTorch's
// Module.to. Parameters are rebound, so the module runs on the new device.
func NNTo(m Module, device Device) error {
	be, err := backendFor(device)
	if err != nil {
		return err
	}
	for _, p := range m.Parameters() {
		cur := p.Tensor()
		if cur.Device() == be.Device() {
			continue
		}
		moved, err := tensor.FromSlice[float32](wrap(cur).ContiguousData(), cur.Shape(), be)
		if err != nil {
			return err
		}
		p.SetTensor(moved)
	}
	return nil
}
