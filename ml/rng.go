package ml

import "math/rand"

var rng = rand.New(rand.NewSource(0))

func ManualSeed(seed int64) {
	rng = rand.New(rand.NewSource(seed))
}

func Randn(shape []int, dtype DType, device Device) (*Tensor, error) {
	data := make([]float32, numelOf(shape))
	for i := range data {
		data[i] = float32(rng.NormFloat64())
	}
	return NewTensorOwned(data, shape, dtype, device)
}
