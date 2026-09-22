//go:build !cgo || static

package ml

import (
	"fmt"

	"blue/borncgo/tensor"
)

// newGPUEngine is the pure-Go fallback: static builds and builds without cgo
// have no WebGPU backend, so device="gpu" reports a clear error.
func newGPUEngine() (tensor.Backend, error) {
	return nil, fmt.Errorf("gpu backend not built (needs cgo and a non-static build)")
}

// gpuIsAvailable is false in static and non-cgo builds.
func gpuIsAvailable() bool {
	return false
}

// GPUInfo has no adapter to describe in static and non-cgo builds.
func GPUInfo() (map[string]string, bool) {
	return nil, false
}
