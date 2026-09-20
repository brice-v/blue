package ml

import (
	"fmt"

	"blue/borncgo/tensor"
)

// CrossEntropy is the mean cross-entropy over logits [batch, classes] and
// integer targets [batch]. It is the default-reduction case of CrossEntropyWith,
// so there is one implementation.
func CrossEntropy(logits, target *Tensor) (*Tensor, error) {
	return CrossEntropyWith(logits, target, CrossEntropyOpts{Reduction: "mean"})
}

// CrossEntropyGrad returns the gradient of the mean cross-entropy loss with
// respect to the logits: (softmax(logits) - onehot(target)) / batch.
//
// It is the seed a compiled backward needs for a classification model: the
// compiled graph is the model's forward (logits), so the loss lives outside it
// and its gradient is supplied here. The value is computed on the real backend
// and is not recorded on the tape, so it can be passed to Compiled.Backward
// without disturbing eager bookkeeping.
func CrossEntropyGrad(logits, target *Tensor) (*Tensor, error) {
	shape := logits.Shape()
	if len(shape) != 2 {
		return nil, fmt.Errorf("cross_entropy_grad: logits must be 2D, got %v", shape)
	}
	batch, classes := shape[0], shape[1]

	// Compute under no-grad so the seed's own ops never land on the tape.
	was := SetGradEnabled(false)
	defer SetGradEnabled(was)

	probs, err := Softmax(logits, 1)
	if err != nil {
		return nil, err
	}
	// onehot(targets) as a constant tensor on logits' backend.
	labels := target.ContiguousData()
	oh := make([]float32, batch*classes)
	for i := 0; i < batch && i < len(labels); i++ {
		c := int(labels[i])
		if c < 0 || c >= classes {
			return nil, fmt.Errorf("cross_entropy_grad: target %d out of range for %d classes", c, classes)
		}
		oh[i*classes+c] = 1
	}
	ohT, err := tensor.FromSlice[float32](oh, tensor.Shape{batch, classes}, logits.be)
	if err != nil {
		return nil, err
	}
	diff, err := Sub(probs, wrap(ohT))
	if err != nil {
		return nil, err
	}
	scale := 1.0 / float32(batch)
	return wrapRaw(logits.be, logits.be.MulScalar(diff.t.Raw(), scale)), nil
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
