//go:build !js

package webgpu

import (
	"fmt"

	wgpu "github.com/oliverbestmann/webgpu/wgpu"
)

// This file contains small helpers that bridge the born GPU backend to the
// CGO-based oliverbestmann/webgpu bindings (wgpu-native).
//
// The two APIs differ in a few systematic ways:
//
//   - Creation methods are split into panicking (Create*) and fallible
//     (TryCreate*) variants. Born uses the Try* variants to preserve the
//     previous error-handling behavior.
//   - Queue.Submit has no error return; validation failures panic.
//   - Device.Poll takes (wait bool, submissionIndex *uint64):
//     Poll(true, nil) waits, Poll(false, nil) is a non-blocking pump.
//   - Compute passes must be released after End and before Submit
//     (see https://github.com/gfx-rs/wgpu/issues/6145).
//   - Buffer mapping is callback-based: TryMapAsync + Device.Poll to drive
//     the callback, then GetMappedRange to access bytes.

// endComputePass ends a compute pass and releases it. The Release is required
// before the enclosing encoder is submitted.
func endComputePass(pass *wgpu.ComputePassEncoder) error {
	if err := pass.TryEnd(); err != nil {
		return err
	}
	pass.Release()
	return nil
}

// mapReadSync maps a MapRead staging buffer for reading, blocking until the
// GPU completes all submitted work. Returns a CPU copy of the mapped bytes;
// the buffer is unmapped before returning.
func mapReadSync(device *wgpu.Device, buf *wgpu.Buffer, size uint64) ([]byte, error) {
	mapStatus := wgpu.MapAsyncStatusError
	if err := buf.TryMapAsync(wgpu.MapModeRead, 0, size, func(s wgpu.MapAsyncStatus) {
		mapStatus = s
	}); err != nil {
		return nil, fmt.Errorf("webgpu: map async: %w", err)
	}

	// Drive the mapping callback to completion.
	device.Poll(true, nil)

	if mapStatus != wgpu.MapAsyncStatusSuccess {
		return nil, fmt.Errorf("webgpu: map failed with status %v (size=%d)", mapStatus, size)
	}
	defer func() { _ = buf.TryUnmap() }()

	mapped := buf.GetMappedRange(0, uint(size))
	if mapped == nil {
		return nil, fmt.Errorf("webgpu: mapped range nil (buffer released?)")
	}
	if uint64(len(mapped)) < size {
		return nil, fmt.Errorf("webgpu: got %d bytes, need %d", len(mapped), size)
	}
	out := make([]byte, size)
	copy(out, mapped)
	return out, nil
}

// writeBufferInit creates a buffer with MappedAtCreation and copies data into
// it, mirroring the previous gogpu/wgpu createBuffer behavior.
func writeBufferInit(device *wgpu.Device, data []byte, usage wgpu.BufferUsage, label string) (*wgpu.Buffer, error) {
	size := uint64(len(data))
	buffer, err := device.TryCreateBuffer(&wgpu.BufferDescriptor{
		Label:            label,
		Usage:            usage,
		Size:             size,
		MappedAtCreation: true,
	})
	if err != nil {
		return nil, err
	}

	mapped := buffer.GetMappedRange(0, uint(size))
	copy(mapped, data)

	if err := buffer.TryUnmap(); err != nil {
		buffer.Release()
		return nil, err
	}

	return buffer, nil
}
