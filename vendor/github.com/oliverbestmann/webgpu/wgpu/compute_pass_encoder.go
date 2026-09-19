//go:build !js

package wgpu

// #include "gen_wgpu_wrappers.h"
import "C"
import (
	"unsafe"
)

func (p *ComputePassEncoder) BeginPipelineStatisticsQuery(querySet *QuerySet, queryIndex uint32) {
	C.wgpuComputePassEncoderBeginPipelineStatisticsQuery(p.ref, querySet.ref, C.uint32_t(queryIndex))
}

func (p *ComputePassEncoder) DispatchWorkgroups(workgroupCountX, workgroupCountY, workgroupCountZ uint32) {
	C.wgpuComputePassEncoderDispatchWorkgroups(p.ref, C.uint32_t(workgroupCountX), C.uint32_t(workgroupCountY), C.uint32_t(workgroupCountZ))
}

func (p *ComputePassEncoder) DispatchWorkgroupsIndirect(indirectBuffer *Buffer, indirectOffset uint64) {
	C.wgpuComputePassEncoderDispatchWorkgroupsIndirect(p.ref, indirectBuffer.ref, C.uint64_t(indirectOffset))
}

func (p *ComputePassEncoder) TryEnd() error {
	errh := acquireErrorCallback()
	defer errh.Done()

	C.go_wgpuComputePassEncoderEnd(p.device, errh.ToPointer(), p.ref)
	return errh.ToError()
}

func (p *ComputePassEncoder) EndPipelineStatisticsQuery() {
	C.wgpuComputePassEncoderEndPipelineStatisticsQuery(p.ref)
}

func (p *ComputePassEncoder) InsertDebugMarker(markerLabel string) {
	label := stringViewOf(markerLabel)
	defer label.Release()

	C.wgpuComputePassEncoderInsertDebugMarker(p.ref, label.ToC())
}

func (p *ComputePassEncoder) PopDebugGroup() {
	C.wgpuComputePassEncoderPopDebugGroup(p.ref)
}

func (p *ComputePassEncoder) PushDebugGroup(groupLabel string) {
	label := stringViewOf(groupLabel)
	defer label.Release()

	C.wgpuComputePassEncoderPushDebugGroup(p.ref, label.ToC())
}

func (p *ComputePassEncoder) SetBindGroup(groupIndex uint32, group *BindGroup, dynamicOffsets []uint32) {
	dynamicOffsetCount := len(dynamicOffsets)
	if dynamicOffsetCount == 0 {
		C.wgpuComputePassEncoderSetBindGroup(p.ref, C.uint32_t(groupIndex), group.ref, 0, nil)
	} else {
		C.wgpuComputePassEncoderSetBindGroup(
			p.ref, C.uint32_t(groupIndex), group.ref,
			C.size_t(dynamicOffsetCount), (*C.uint32_t)(unsafe.Pointer(&dynamicOffsets[0])),
		)
	}
}

func (p *ComputePassEncoder) SetPipeline(pipeline *ComputePipeline) {
	C.wgpuComputePassEncoderSetPipeline(p.ref, pipeline.ref)
}
