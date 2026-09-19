//go:build cgo && !static

package ml

import (
	"blue/borncgo/autodiff"
	"blue/borncgo/backend/webgpu"
	"blue/borncgo/tensor"
)

// newGPUEngine builds the WebGPU autodiff backend. Compiled in for cgo,
// non-static builds, so static and wasm builds stay pure Go.
func newGPUEngine() (tensor.Backend, error) {
	be, err := webgpu.New()
	if err != nil {
		return nil, err
	}
	return autodiff.New(be), nil
}

// gpuIsAvailable reports whether a WebGPU adapter can be created.
func gpuIsAvailable() bool {
	return webgpu.IsAvailable()
}
