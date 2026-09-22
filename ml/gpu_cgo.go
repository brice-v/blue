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

// GPUInfo returns a description of the GPU adapter (name, vendor, device,
// backend) for `ml.gpu.info()`. ok is false when no GPU backend is available,
// which is the case for static and non-cgo builds.
func GPUInfo() (map[string]string, bool) {
	be := ensureGPU()
	if be == nil {
		return nil, false
	}
	// newGPUEngine wraps the webgpu backend in the autodiff decorator, so
	// unwrap it to reach the adapter details.
	if ad, ok := be.(*autodiff.Backend[*webgpu.Backend]); ok {
		if wb := ad.Inner(); wb != nil {
			return wb.Info(), true
		}
	}
	return map[string]string{"name": be.Name(), "device": be.Device().String()}, true
}
