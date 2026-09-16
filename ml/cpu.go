package ml

type CPUBackend struct{}

var DefaultBackend Backend = CPUBackend{}

func (cpu CPUBackend) MatMul(a, b *Tensor) (*Tensor, error) {
	return nil, nil
}

func (cpu CPUBackend) Add(a, b *Tensor) (*Tensor, error) {
	return nil, nil
}

func (cpu CPUBackend) Sub(a, b *Tensor) (*Tensor, error) {
	return nil, nil
}

func (cpu CPUBackend) Mul(a, b *Tensor) (*Tensor, error) {
	return nil, nil
}

func (cpu CPUBackend) Div(a, b *Tensor) (*Tensor, error) {
	return nil, nil
}

func (cpu CPUBackend) Exp(a *Tensor) (*Tensor, error) {
	return nil, nil
}

func (cpu CPUBackend) Log(a *Tensor) (*Tensor, error) {
	return nil, nil
}

func (cpu CPUBackend) Sqrt(a *Tensor) (*Tensor, error) {
	return nil, nil
}

func (cpu CPUBackend) Sum(a *Tensor, dim int, keepdim bool) (*Tensor, error) {
	return nil, nil
}

func (cpu CPUBackend) Max(a *Tensor, dim int, keepdim bool) (*Tensor, error) {
	return nil, nil
}

func (cpu CPUBackend) Softmax(a *Tensor, dim int) (*Tensor, error) {
	return nil, nil
}

func (cpu CPUBackend) Reshape(a *Tensor, shape ...int) (*Tensor, error) {
	return nil, nil
}

func (cpu CPUBackend) Transpose(a *Tensor, dim0, dim1 int) (*Tensor, error) {
	return nil, nil
}
