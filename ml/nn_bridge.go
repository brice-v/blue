package ml

import (
	"fmt"

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

// NNConv2D builds a 2D convolution on the requested device. The kernel is
// square (kernelH == kernelW), matching the common PyTorch default.
func NNConv2D(inChannels, outChannels, kernelSize, stride, padding int, useBias bool, device Device) (Module, error) {
	be, err := backendFor(device)
	if err != nil {
		return nil, err
	}
	return nn.NewConv2D(inChannels, outChannels, kernelSize, kernelSize, stride, padding, useBias, be), nil
}

// NNMaxPool2D builds a 2D max pooling layer on the requested device.
func NNMaxPool2D(kernelSize, stride int, device Device) (Module, error) {
	be, err := backendFor(device)
	if err != nil {
		return nil, err
	}
	return nn.NewMaxPool2D(kernelSize, stride, be), nil
}

// NNForward runs a module on a blue tensor. Linear layers go through the fused
// matmul+bias path so their weight stays a rebindable parameter (an optimizer
// replaces a parameter's tensor each step) and so a compiled graph can fuse the
// epilogue.
func NNForward(m Module, x *Tensor) *Tensor {
	if out, err := NNLinearForwardAct(m, x, false); err == nil {
		return out
	}
	return wrap(m.Forward(x.t))
}

// NNParameters returns the module's borncgo parameters.
func NNParameters(m Module) []*Param {
	return m.Parameters()
}

// matMulBiasBackend is implemented by the autodiff backend with the fused
// matmul+bias(+relu) kernel.
type matMulBiasBackend interface {
	MatMulBias(a, b, bias *tensor.RawTensor, relu bool) *tensor.RawTensor
}

// NNLinearForwardAct runs a Linear layer fused with the following activation
// when the engine supports the fused kernel, otherwise it falls back to the
// layer's own forward. This is the "compile" step for the common
// Linear -> ReLU pattern: one dispatch instead of matmul + add + relu, with no
// materialized activation in between.
func NNLinearForwardAct(m Module, x *Tensor, relu bool) (*Tensor, error) {
	l, ok := m.(*nn.Linear[tensor.Backend])
	if !ok || l.Bias() == nil {
		out := m.Forward(x.t)
		if relu {
			return Relu(wrap(out))
		}
		return wrap(out), nil
	}
	mb, ok := x.be.(matMulBiasBackend)
	if !ok {
		out := m.Forward(x.t)
		if relu {
			return Relu(wrap(out))
		}
		return wrap(out), nil
	}
	w := l.Weight().Tensor()
	bias := l.Bias().Tensor()
	if w.DType() != tensor.Float32 || bias.DType() != tensor.Float32 {
		return nil, fmt.Errorf("fused linear requires float32 weight and bias")
	}
	// Under a tracing backend, bind the weight and bias as rebindable
	// parameters: optimizers replace a parameter's tensor each step, so the
	// compiled graph must read them fresh rather than freeze the raw.
	if pi, ok := x.be.(paramInterner); ok {
		pi.InternParam(l.Weight())
		pi.InternParam(l.Bias())
	}
	out := mb.MatMulBias(x.t.Raw(), w.Raw(), bias.Raw(), relu)
	return wrapRaw(x.be, out), nil
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
