//========================================================================
// GLFW 3.4 OSMesa - www.glfw.org
//------------------------------------------------------------------------
// Copyright (c) 2016 Google Inc.
// Copyright (c) 2016-2017 Camilla Löwy <elmindreda@glfw.org>
//
// This software is provided 'as-is', without any express or implied
// warranty. In no event will the authors be held liable for any damages
// arising from the use of this software.
//
// Permission is granted to anyone to use this software for any purpose,
// including commercial applications, and to alter it and redistribute it
// freely, subject to the following restrictions:
//
// 1. The origin of this software must not be misrepresented; you must not
//    claim that you wrote the original software. If you use this software
//    in a product, an acknowledgment in the product documentation would
//    be appreciated but is not required.
//
// 2. Altered source versions must be plainly marked as such, and must not
//    be misrepresented as being the original software.
//
// 3. This notice may not be removed or altered from any source
//    distribution.
//
//========================================================================
// Please use C89 style variable declarations in this file because VS 2010
//========================================================================

#include <stdlib.h>
#include <string.h>
#include <assert.h>

#include "internal.h"


static void makeContextCurrentOSMesa(_GLFWwindow* window)
{
    if (window)
    {
        int width, height;
        __glfw.platform.getFramebufferSize(window, &width, &height);

        // Check to see if we need to allocate a new buffer
        if ((window->context.osmesa.buffer == NULL) ||
            (width != window->context.osmesa.width) ||
            (height != window->context.osmesa.height))
        {
            __glfw_free(window->context.osmesa.buffer);

            // Allocate the new buffer (width * height * 8-bit RGBA)
            window->context.osmesa.buffer = __glfw_calloc(4, (size_t) width * height);
            window->context.osmesa.width  = width;
            window->context.osmesa.height = height;
        }

        if (!OSMesaMakeCurrent(window->context.osmesa.handle,
                               window->context.osmesa.buffer,
                               GL_UNSIGNED_BYTE,
                               width, height))
        {
            ___glfwInputError(GLFW_PLATFORM_ERROR,
                            "OSMesa: Failed to make context current");
            return;
        }
    }

    ___glfwPlatformSetTls(&__glfw.contextSlot, window);
}

static GLFWglproc getProcAddressOSMesa(const char* procname)
{
    return (GLFWglproc) OSMesaGetProcAddress(procname);
}

static void destroyContextOSMesa(_GLFWwindow* window)
{
    if (window->context.osmesa.handle)
    {
        OSMesaDestroyContext(window->context.osmesa.handle);
        window->context.osmesa.handle = NULL;
    }

    if (window->context.osmesa.buffer)
    {
        __glfw_free(window->context.osmesa.buffer);
        window->context.osmesa.width = 0;
        window->context.osmesa.height = 0;
    }
}

static void swapBuffersOSMesa(_GLFWwindow* window)
{
    // No double buffering on OSMesa
}

static void swapIntervalOSMesa(int interval)
{
    // No swap interval on OSMesa
}

static int extensionSupportedOSMesa(const char* extension)
{
    // OSMesa does not have extensions
    return GLFW_FALSE;
}


//////////////////////////////////////////////////////////////////////////
//////                       GLFW internal API                      //////
//////////////////////////////////////////////////////////////////////////

GLFWbool ____glfwInitOSMesa(void)
{
    int i;
    const char* sonames[] =
    {
#if defined(_GLFW_OSMESA_LIBRARY)
        _GLFW_OSMESA_LIBRARY,
#elif defined(_WIN32)
        "libOSMesa.dll",
        "OSMesa.dll",
#elif defined(__APPLE__)
        "libOSMesa.8.dylib",
#elif defined(__CYGWIN__)
        "libOSMesa-8.so",
#elif defined(__OpenBSD__) || defined(__NetBSD__)
        "libOSMesa.so",
#else
        "libOSMesa.so.8",
        "libOSMesa.so.6",
#endif
        NULL
    };

    if (__glfw.osmesa.handle)
        return GLFW_TRUE;

    for (i = 0;  sonames[i];  i++)
    {
        __glfw.osmesa.handle = __glfwPlatformLoadModule(sonames[i]);
        if (__glfw.osmesa.handle)
            break;
    }

    if (!__glfw.osmesa.handle)
    {
        ___glfwInputError(GLFW_API_UNAVAILABLE, "OSMesa: Library not found");
        return GLFW_FALSE;
    }

    __glfw.osmesa.CreateContextExt = (PFN_OSMesaCreateContextExt)
        __glfwPlatformGetModuleSymbol(__glfw.osmesa.handle, "OSMesaCreateContextExt");
    __glfw.osmesa.CreateContextAttribs = (PFN_OSMesaCreateContextAttribs)
        __glfwPlatformGetModuleSymbol(__glfw.osmesa.handle, "OSMesaCreateContextAttribs");
    __glfw.osmesa.DestroyContext = (PFN_OSMesaDestroyContext)
        __glfwPlatformGetModuleSymbol(__glfw.osmesa.handle, "OSMesaDestroyContext");
    __glfw.osmesa.MakeCurrent = (PFN_OSMesaMakeCurrent)
        __glfwPlatformGetModuleSymbol(__glfw.osmesa.handle, "OSMesaMakeCurrent");
    __glfw.osmesa.GetColorBuffer = (PFN_OSMesaGetColorBuffer)
        __glfwPlatformGetModuleSymbol(__glfw.osmesa.handle, "OSMesaGetColorBuffer");
    __glfw.osmesa.GetDepthBuffer = (PFN_OSMesaGetDepthBuffer)
        __glfwPlatformGetModuleSymbol(__glfw.osmesa.handle, "OSMesaGetDepthBuffer");
    __glfw.osmesa.GetProcAddress = (PFN_OSMesaGetProcAddress)
        __glfwPlatformGetModuleSymbol(__glfw.osmesa.handle, "OSMesaGetProcAddress");

    if (!__glfw.osmesa.CreateContextExt ||
        !__glfw.osmesa.DestroyContext ||
        !__glfw.osmesa.MakeCurrent ||
        !__glfw.osmesa.GetColorBuffer ||
        !__glfw.osmesa.GetDepthBuffer ||
        !__glfw.osmesa.GetProcAddress)
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "OSMesa: Failed to load required entry points");

        ____glfwTerminateOSMesa();
        return GLFW_FALSE;
    }

    return GLFW_TRUE;
}

void ____glfwTerminateOSMesa(void)
{
    if (__glfw.osmesa.handle)
    {
        __glfwPlatformFreeModule(__glfw.osmesa.handle);
        __glfw.osmesa.handle = NULL;
    }
}

#define SET_ATTRIB(a, v) \
{ \
    assert(((size_t) index + 1) < sizeof(attribs) / sizeof(attribs[0])); \
    attribs[index++] = a; \
    attribs[index++] = v; \
}

GLFWbool ___glfwCreateContextOSMesa(_GLFWwindow* window,
                                  const _GLFWctxconfig* ctxconfig,
                                  const _GLFWfbconfig* fbconfig)
{
    OSMesaContext share = NULL;
    const int accumBits = fbconfig->accumRedBits +
                          fbconfig->accumGreenBits +
                          fbconfig->accumBlueBits +
                          fbconfig->accumAlphaBits;

    if (ctxconfig->client == GLFW_OPENGL_ES_API)
    {
        ___glfwInputError(GLFW_API_UNAVAILABLE,
                        "OSMesa: OpenGL ES is not available on OSMesa");
        return GLFW_FALSE;
    }

    if (ctxconfig->share)
        share = ctxconfig->share->context.osmesa.handle;

    if (OSMesaCreateContextAttribs)
    {
        int index = 0, attribs[40];

        SET_ATTRIB(OSMESA_FORMAT, OSMESA_RGBA);
        SET_ATTRIB(OSMESA_DEPTH_BITS, fbconfig->depthBits);
        SET_ATTRIB(OSMESA_STENCIL_BITS, fbconfig->stencilBits);
        SET_ATTRIB(OSMESA_ACCUM_BITS, accumBits);

        if (ctxconfig->profile == GLFW_OPENGL_CORE_PROFILE)
        {
            SET_ATTRIB(OSMESA_PROFILE, OSMESA_CORE_PROFILE);
        }
        else if (ctxconfig->profile == GLFW_OPENGL_COMPAT_PROFILE)
        {
            SET_ATTRIB(OSMESA_PROFILE, OSMESA_COMPAT_PROFILE);
        }

        if (ctxconfig->major != 1 || ctxconfig->minor != 0)
        {
            SET_ATTRIB(OSMESA_CONTEXT_MAJOR_VERSION, ctxconfig->major);
            SET_ATTRIB(OSMESA_CONTEXT_MINOR_VERSION, ctxconfig->minor);
        }

        if (ctxconfig->forward)
        {
            ___glfwInputError(GLFW_VERSION_UNAVAILABLE,
                            "OSMesa: Forward-compatible contexts not supported");
            return GLFW_FALSE;
        }

        SET_ATTRIB(0, 0);

        window->context.osmesa.handle =
            OSMesaCreateContextAttribs(attribs, share);
    }
    else
    {
        if (ctxconfig->profile)
        {
            ___glfwInputError(GLFW_VERSION_UNAVAILABLE,
                            "OSMesa: OpenGL profiles unavailable");
            return GLFW_FALSE;
        }

        window->context.osmesa.handle =
            OSMesaCreateContextExt(OSMESA_RGBA,
                                   fbconfig->depthBits,
                                   fbconfig->stencilBits,
                                   accumBits,
                                   share);
    }

    if (window->context.osmesa.handle == NULL)
    {
        ___glfwInputError(GLFW_VERSION_UNAVAILABLE,
                        "OSMesa: Failed to create context");
        return GLFW_FALSE;
    }

    window->context.makeCurrent = makeContextCurrentOSMesa;
    window->context.swapBuffers = swapBuffersOSMesa;
    window->context.swapInterval = swapIntervalOSMesa;
    window->context.extensionSupported = extensionSupportedOSMesa;
    window->context.getProcAddress = getProcAddressOSMesa;
    window->context.destroy = destroyContextOSMesa;

    return GLFW_TRUE;
}

#undef SET_ATTRIB


//////////////////////////////////////////////////////////////////////////
//////                        GLFW native API                       //////
//////////////////////////////////////////////////////////////////////////

GLFWAPI int __glfwGetOSMesaColorBuffer(GLFWwindow* handle, int* width,
                                     int* height, int* format, void** buffer)
{
    void* mesaBuffer;
    GLint mesaWidth, mesaHeight, mesaFormat;
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT_OR_RETURN(GLFW_FALSE);

    if (window->context.source != GLFW_OSMESA_CONTEXT_API)
    {
        ___glfwInputError(GLFW_NO_WINDOW_CONTEXT, NULL);
        return GLFW_FALSE;
    }

    if (!OSMesaGetColorBuffer(window->context.osmesa.handle,
                              &mesaWidth, &mesaHeight,
                              &mesaFormat, &mesaBuffer))
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "OSMesa: Failed to retrieve color buffer");
        return GLFW_FALSE;
    }

    if (width)
        *width = mesaWidth;
    if (height)
        *height = mesaHeight;
    if (format)
        *format = mesaFormat;
    if (buffer)
        *buffer = mesaBuffer;

    return GLFW_TRUE;
}

GLFWAPI int __glfwGetOSMesaDepthBuffer(GLFWwindow* handle,
                                     int* width, int* height,
                                     int* bytesPerValue,
                                     void** buffer)
{
    void* mesaBuffer;
    GLint mesaWidth, mesaHeight, mesaBytes;
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT_OR_RETURN(GLFW_FALSE);

    if (window->context.source != GLFW_OSMESA_CONTEXT_API)
    {
        ___glfwInputError(GLFW_NO_WINDOW_CONTEXT, NULL);
        return GLFW_FALSE;
    }

    if (!OSMesaGetDepthBuffer(window->context.osmesa.handle,
                              &mesaWidth, &mesaHeight,
                              &mesaBytes, &mesaBuffer))
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "OSMesa: Failed to retrieve depth buffer");
        return GLFW_FALSE;
    }

    if (width)
        *width = mesaWidth;
    if (height)
        *height = mesaHeight;
    if (bytesPerValue)
        *bytesPerValue = mesaBytes;
    if (buffer)
        *buffer = mesaBuffer;

    return GLFW_TRUE;
}

GLFWAPI OSMesaContext __glfwGetOSMesaContext(GLFWwindow* handle)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);

    if (window->context.source != GLFW_OSMESA_CONTEXT_API)
    {
        ___glfwInputError(GLFW_NO_WINDOW_CONTEXT, NULL);
        return NULL;
    }

    return window->context.osmesa.handle;
}

