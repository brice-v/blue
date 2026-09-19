//go:build !js

package wgpu

/*

#include <stdlib.h>
#include "gen_wgpu_wrappers.h"

#cgo noescape wgpuTextureViewRelease
#cgo nocallback wgpuTextureViewRelease

#cgo noescape go_wgpuTextureCreateView
#cgo nocallback go_wgpuTextureCreateView

*/
import "C"

func (g *Texture) TryCreateView(descriptor *TextureViewDescriptor) (*TextureView, error) {
	var desc *C.WGPUTextureViewDescriptor

	if descriptor != nil {
		desc = &C.WGPUTextureViewDescriptor{
			format:          C.WGPUTextureFormat(descriptor.Format),
			dimension:       C.WGPUTextureViewDimension(descriptor.Dimension),
			baseMipLevel:    C.uint32_t(descriptor.BaseMipLevel),
			mipLevelCount:   C.uint32_t(descriptor.MipLevelCount),
			baseArrayLayer:  C.uint32_t(descriptor.BaseArrayLayer),
			arrayLayerCount: C.uint32_t(descriptor.ArrayLayerCount),
			aspect:          C.WGPUTextureAspect(descriptor.Aspect),
		}

		label := stringViewOf(descriptor.Label)
		defer label.Release()
		desc.label = label.ToC()
	}

	errh := acquireErrorCallback()
	defer errh.Done()

	ref := C.go_wgpuTextureCreateView(
		g.device,
		errh.ToPointer(),
		g.ref,
		desc,
	)
	if err := errh.ToError(); err != nil {
		C.wgpuTextureViewRelease(ref)
		return nil, err
	}

	return releaseOnGC(&TextureView{ref: ref}), nil
}

func (g *Texture) Destroy() {
	C.wgpuTextureDestroy(g.ref)
}

func (g *Texture) GetDepthOrArrayLayers() uint32 {
	return uint32(C.wgpuTextureGetDepthOrArrayLayers(g.ref))
}

func (g *Texture) GetDimension() TextureDimension {
	return TextureDimension(C.wgpuTextureGetDimension(g.ref))
}

func (g *Texture) GetFormat() TextureFormat {
	return TextureFormat(C.wgpuTextureGetFormat(g.ref))
}

func (g *Texture) GetHeight() uint32 {
	return uint32(C.wgpuTextureGetHeight(g.ref))
}

func (g *Texture) GetMipLevelCount() uint32 {
	return uint32(C.wgpuTextureGetMipLevelCount(g.ref))
}

func (g *Texture) GetSampleCount() uint32 {
	return uint32(C.wgpuTextureGetSampleCount(g.ref))
}

func (g *Texture) GetUsage() TextureUsage {
	return TextureUsage(C.wgpuTextureGetUsage(g.ref))
}

func (g *Texture) GetWidth() uint32 {
	return uint32(C.wgpuTextureGetWidth(g.ref))
}
