package ml

import (
	"blue/borncgo/nn"
	"blue/borncgo/tensor"
)

// CrossEntropy wraps borncgo's cross-entropy loss. blue labels are float32, so
// they are converted to the int32 class indices borncgo expects. The loss
// records on the backend's tape, so it is differentiable on CPU and GPU.
func CrossEntropy(logits, target *Tensor) (*Tensor, error) {
	tgt, err := toInt32(target)
	if err != nil {
		return nil, err
	}
	loss := nn.NewCrossEntropyLoss(logits.be)
	return wrap(loss.Forward(logits.t, tgt)), nil
}

// MSE is composed from ops rather than borncgo's nn.MSELoss, because that loss
// reads its input data and rebinds the result, which breaks the tape. Mean of
// (pred - target)^2 stays differentiable.
func MSE(pred, target *Tensor) (*Tensor, error) {
	diff, err := Sub(pred, target)
	if err != nil {
		return nil, err
	}
	sq, err := Mul(diff, diff)
	if err != nil {
		return nil, err
	}
	return Mean(sq, nil, false)
}

func toInt32(t *Tensor) (*tensor.Tensor[int32, tensor.Backend], error) {
	data := t.ContiguousData()
	ids := make([]int32, len(data))
	for i, v := range data {
		ids[i] = int32(v)
	}
	return tensor.FromSlice[int32](ids, tensor.Shape(t.Shape()), t.be)
}
