//========================================================================
// GLFW 3.4 EGL - www.glfw.org
//------------------------------------------------------------------------
// Copyright (c) 2002-2006 Marcus Geelnard
// Copyright (c) 2006-2019 Camilla Löwy <elmindreda@glfw.org>
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

#include "internal.h"

#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <assert.h>


// Return a description of the specified EGL error
//
static const char* getEGLErrorString(EGLint error)
{
    switch (error)
    {
        case EGL_SUCCESS:
            return "Success";
        case EGL_NOT_INITIALIZED:
            return "EGL is not or could not be initialized";
        case EGL_BAD_ACCESS:
            return "EGL cannot access a requested resource";
        case EGL_BAD_ALLOC:
            return "EGL failed to allocate resources for the requested operation";
        case EGL_BAD_ATTRIBUTE:
            return "An unrecognized attribute or attribute value was passed in the attribute list";
        case EGL_BAD_CONTEXT:
            return "An EGLContext argument does not name a valid EGL rendering context";
        case EGL_BAD_CONFIG:
            return "An EGLConfig argument does not name a valid EGL frame buffer configuration";
        case EGL_BAD_CURRENT_SURFACE:
            return "The current surface of the calling thread is a window, pixel buffer or pixmap that is no longer valid";
        case EGL_BAD_DISPLAY:
            return "An EGLDisplay argument does not name a valid EGL display connection";
        case EGL_BAD_SURFACE:
            return "An EGLSurface argument does not name a valid surface configured for GL rendering";
        case EGL_BAD_MATCH:
            return "Arguments are inconsistent";
        case EGL_BAD_PARAMETER:
            return "One or more argument values are invalid";
        case EGL_BAD_NATIVE_PIXMAP:
            return "A NativePixmapType argument does not refer to a valid native pixmap";
        case EGL_BAD_NATIVE_WINDOW:
            return "A NativeWindowType argument does not refer to a valid native window";
        case EGL_CONTEXT_LOST:
            return "The application must destroy all contexts and reinitialise";
        default:
            return "ERROR: UNKNOWN EGL ERROR";
    }
}

// Returns the specified attribute of the specified EGLConfig
//
static int getEGLConfigAttrib(EGLConfig config, int attrib)
{
    int value;
    eglGetConfigAttrib(__glfw.egl.display, config, attrib, &value);
    return value;
}

// Return the EGLConfig most closely matching the specified hints
//
static GLFWbool chooseEGLConfig(const _GLFWctxconfig* ctxconfig,
                                const _GLFWfbconfig* desired,
                                EGLConfig* result)
{
    EGLConfig* nativeConfigs;
    _GLFWfbconfig* usableConfigs;
    const _GLFWfbconfig* closest;
    int i, nativeCount, usableCount;

    eglGetConfigs(__glfw.egl.display, NULL, 0, &nativeCount);
    if (!nativeCount)
    {
        ___glfwInputError(GLFW_API_UNAVAILABLE, "EGL: No EGLConfigs returned");
        return GLFW_FALSE;
    }

    nativeConfigs = __glfw_calloc(nativeCount, sizeof(EGLConfig));
    eglGetConfigs(__glfw.egl.display, nativeConfigs, nativeCount, &nativeCount);

    usableConfigs = __glfw_calloc(nativeCount, sizeof(_GLFWfbconfig));
    usableCount = 0;

    for (i = 0;  i < nativeCount;  i++)
    {
        const EGLConfig n = nativeConfigs[i];
        _GLFWfbconfig* u = usableConfigs + usableCount;

        // Only consider RGB(A) EGLConfigs
        if (getEGLConfigAttrib(n, EGL_COLOR_BUFFER_TYPE) != EGL_RGB_BUFFER)
            continue;

        // Only consider window EGLConfigs
        if (!(getEGLConfigAttrib(n, EGL_SURFACE_TYPE) & EGL_WINDOW_BIT))
            continue;

#if defined(_GLFW_X11)
        if (__glfw.platform.platformID == GLFW_PLATFORM_X11)
        {
            XVisualInfo vi = {0};

            // Only consider EGLConfigs with associated Visuals
            vi.visualid = getEGLConfigAttrib(n, EGL_NATIVE_VISUAL_ID);
            if (!vi.visualid)
                continue;

            if (desired->transparent)
            {
                int count;
                XVisualInfo* vis =
                    XGetVisualInfo(__glfw.x11.display, VisualIDMask, &vi, &count);
                if (vis)
                {
                    u->transparent = ___glfwIsVisualTransparentX11(vis[0].visual);
                    XFree(vis);
                }
            }
        }
#endif // _GLFW_X11

        if (ctxconfig->client == GLFW_OPENGL_ES_API)
        {
            if (ctxconfig->major == 1)
            {
                if (!(getEGLConfigAttrib(n, EGL_RENDERABLE_TYPE) & EGL_OPENGL_ES_BIT))
                    continue;
            }
            else
            {
                if (!(getEGLConfigAttrib(n, EGL_RENDERABLE_TYPE) & EGL_OPENGL_ES2_BIT))
                    continue;
            }
        }
        else if (ctxconfig->client == GLFW_OPENGL_API)
        {
            if (!(getEGLConfigAttrib(n, EGL_RENDERABLE_TYPE) & EGL_OPENGL_BIT))
                continue;
        }

        u->redBits = getEGLConfigAttrib(n, EGL_RED_SIZE);
        u->greenBits = getEGLConfigAttrib(n, EGL_GREEN_SIZE);
        u->blueBits = getEGLConfigAttrib(n, EGL_BLUE_SIZE);

        u->alphaBits = getEGLConfigAttrib(n, EGL_ALPHA_SIZE);
        u->depthBits = getEGLConfigAttrib(n, EGL_DEPTH_SIZE);
        u->stencilBits = getEGLConfigAttrib(n, EGL_STENCIL_SIZE);

#if defined(_GLFW_WAYLAND)
        if (__glfw.platform.platformID == GLFW_PLATFORM_WAYLAND)
        {
            // NOTE: The wl_surface opaque region is no guarantee that its buffer
            //       is presented as opaque, if it also has an alpha channel
            // HACK: If EGL_EXT_present_opaque is unavailable, ignore any config
            //       with an alpha channel to ensure the buffer is opaque
            if (!__glfw.egl.EXT_present_opaque)
            {
                if (!desired->transparent && u->alphaBits > 0)
                    continue;
            }
        }
#endif // _GLFW_WAYLAND

        u->samples = getEGLConfigAttrib(n, EGL_SAMPLES);
        u->doublebuffer = desired->doublebuffer;

        u->handle = (uintptr_t) n;
        usableCount++;
    }

    closest = ___glfwChooseFBConfig(desired, usableConfigs, usableCount);
    if (closest)
        *result = (EGLConfig) closest->handle;

    __glfw_free(nativeConfigs);
    __glfw_free(usableConfigs);

    return closest != NULL;
}

static void makeContextCurrentEGL(_GLFWwindow* window)
{
    if (window)
    {
        if (!eglMakeCurrent(__glfw.egl.display,
                            window->context.egl.surface,
                            window->context.egl.surface,
                            window->context.egl.handle))
        {
            ___glfwInputError(GLFW_PLATFORM_ERROR,
                            "EGL: Failed to make context current: %s",
                            getEGLErrorString(eglGetError()));
            return;
        }
    }
    else
    {
        if (!eglMakeCurrent(__glfw.egl.display,
                            EGL_NO_SURFACE,
                            EGL_NO_SURFACE,
                            EGL_NO_CONTEXT))
        {
            ___glfwInputError(GLFW_PLATFORM_ERROR,
                            "EGL: Failed to clear current context: %s",
                            getEGLErrorString(eglGetError()));
            return;
        }
    }

    ___glfwPlatformSetTls(&__glfw.contextSlot, window);
}

static void swapBuffersEGL(_GLFWwindow* window)
{
    if (window != ___glfwPlatformGetTls(&__glfw.contextSlot))
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "EGL: The context must be current on the calling thread when swapping buffers");
        return;
    }

#if defined(_GLFW_WAYLAND)
    if (__glfw.platform.platformID == GLFW_PLATFORM_WAYLAND)
    {
        // NOTE: Swapping buffers on a hidden window on Wayland makes it visible
        if (!window->wl.visible)
            return;
    }
#endif

    eglSwapBuffers(__glfw.egl.display, window->context.egl.surface);
}

static void swapIntervalEGL(int interval)
{
    eglSwapInterval(__glfw.egl.display, interval);
}

static int extensionSupportedEGL(const char* extension)
{
    const char* extensions = eglQueryString(__glfw.egl.display, EGL_EXTENSIONS);
    if (extensions)
    {
        if (___glfwStringInExtensionString(extension, extensions))
            return GLFW_TRUE;
    }

    return GLFW_FALSE;
}

static GLFWglproc getProcAddressEGL(const char* procname)
{
    _GLFWwindow* window = ___glfwPlatformGetTls(&__glfw.contextSlot);

    if (window->context.egl.client)
    {
        GLFWglproc proc = (GLFWglproc)
            __glfwPlatformGetModuleSymbol(window->context.egl.client, procname);
        if (proc)
            return proc;
    }

    return eglGetProcAddress(procname);
}

static void destroyContextEGL(_GLFWwindow* window)
{
    // NOTE: Do not unload libGL.so.1 while the X11 display is still open,
    //       as it will make XCloseDisplay segfault
    if (__glfw.platform.platformID != GLFW_PLATFORM_X11 ||
        window->context.client != GLFW_OPENGL_API)
    {
        if (window->context.egl.client)
        {
            __glfwPlatformFreeModule(window->context.egl.client);
            window->context.egl.client = NULL;
        }
    }

    if (window->context.egl.surface)
    {
        eglDestroySurface(__glfw.egl.display, window->context.egl.surface);
        window->context.egl.surface = EGL_NO_SURFACE;
    }

    if (window->context.egl.handle)
    {
        eglDestroyContext(__glfw.egl.display, window->context.egl.handle);
        window->context.egl.handle = EGL_NO_CONTEXT;
    }
}


//////////////////////////////////////////////////////////////////////////
//////                       GLFW internal API                      //////
//////////////////////////////////////////////////////////////////////////

// Initialize EGL
//
GLFWbool ____glfwInitEGL(void)
{
    int i;
    EGLint* attribs = NULL;
    const char* extensions;
    const char* sonames[] =
    {
#if defined(_GLFW_EGL_LIBRARY)
        _GLFW_EGL_LIBRARY,
#elif defined(_GLFW_WIN32)
        "libEGL.dll",
        "EGL.dll",
#elif defined(_GLFW_COCOA)
        "libEGL.dylib",
#elif defined(__CYGWIN__)
        "libEGL-1.so",
#elif defined(__OpenBSD__) || defined(__NetBSD__)
        "libEGL.so",
#else
        "libEGL.so.1",
#endif
        NULL
    };

    if (__glfw.egl.handle)
        return GLFW_TRUE;

    for (i = 0;  sonames[i];  i++)
    {
        __glfw.egl.handle = __glfwPlatformLoadModule(sonames[i]);
        if (__glfw.egl.handle)
            break;
    }

    if (!__glfw.egl.handle)
    {
        ___glfwInputError(GLFW_API_UNAVAILABLE, "EGL: Library not found");
        return GLFW_FALSE;
    }

    __glfw.egl.prefix = (strncmp(sonames[i], "lib", 3) == 0);

    __glfw.egl.GetConfigAttrib = (PFN_eglGetConfigAttrib)
        __glfwPlatformGetModuleSymbol(__glfw.egl.handle, "eglGetConfigAttrib");
    __glfw.egl.GetConfigs = (PFN_eglGetConfigs)
        __glfwPlatformGetModuleSymbol(__glfw.egl.handle, "eglGetConfigs");
    __glfw.egl.GetDisplay = (PFN_eglGetDisplay)
        __glfwPlatformGetModuleSymbol(__glfw.egl.handle, "eglGetDisplay");
    __glfw.egl.GetError = (PFN_eglGetError)
        __glfwPlatformGetModuleSymbol(__glfw.egl.handle, "eglGetError");
    __glfw.egl.Initialize = (PFN_eglInitialize)
        __glfwPlatformGetModuleSymbol(__glfw.egl.handle, "eglInitialize");
    __glfw.egl.Terminate = (PFN_eglTerminate)
        __glfwPlatformGetModuleSymbol(__glfw.egl.handle, "eglTerminate");
    __glfw.egl.BindAPI = (PFN_eglBindAPI)
        __glfwPlatformGetModuleSymbol(__glfw.egl.handle, "eglBindAPI");
    __glfw.egl.CreateContext = (PFN_eglCreateContext)
        __glfwPlatformGetModuleSymbol(__glfw.egl.handle, "eglCreateContext");
    __glfw.egl.DestroySurface = (PFN_eglDestroySurface)
        __glfwPlatformGetModuleSymbol(__glfw.egl.handle, "eglDestroySurface");
    __glfw.egl.DestroyContext = (PFN_eglDestroyContext)
        __glfwPlatformGetModuleSymbol(__glfw.egl.handle, "eglDestroyContext");
    __glfw.egl.CreateWindowSurface = (PFN_eglCreateWindowSurface)
        __glfwPlatformGetModuleSymbol(__glfw.egl.handle, "eglCreateWindowSurface");
    __glfw.egl.MakeCurrent = (PFN_eglMakeCurrent)
        __glfwPlatformGetModuleSymbol(__glfw.egl.handle, "eglMakeCurrent");
    __glfw.egl.SwapBuffers = (PFN_eglSwapBuffers)
        __glfwPlatformGetModuleSymbol(__glfw.egl.handle, "eglSwapBuffers");
    __glfw.egl.SwapInterval = (PFN_eglSwapInterval)
        __glfwPlatformGetModuleSymbol(__glfw.egl.handle, "eglSwapInterval");
    __glfw.egl.QueryString = (PFN_eglQueryString)
        __glfwPlatformGetModuleSymbol(__glfw.egl.handle, "eglQueryString");
    __glfw.egl.GetProcAddress = (PFN_eglGetProcAddress)
        __glfwPlatformGetModuleSymbol(__glfw.egl.handle, "eglGetProcAddress");

    if (!__glfw.egl.GetConfigAttrib ||
        !__glfw.egl.GetConfigs ||
        !__glfw.egl.GetDisplay ||
        !__glfw.egl.GetError ||
        !__glfw.egl.Initialize ||
        !__glfw.egl.Terminate ||
        !__glfw.egl.BindAPI ||
        !__glfw.egl.CreateContext ||
        !__glfw.egl.DestroySurface ||
        !__glfw.egl.DestroyContext ||
        !__glfw.egl.CreateWindowSurface ||
        !__glfw.egl.MakeCurrent ||
        !__glfw.egl.SwapBuffers ||
        !__glfw.egl.SwapInterval ||
        !__glfw.egl.QueryString ||
        !__glfw.egl.GetProcAddress)
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "EGL: Failed to load required entry points");

        ____glfwTerminateEGL();
        return GLFW_FALSE;
    }

    extensions = eglQueryString(EGL_NO_DISPLAY, EGL_EXTENSIONS);
    if (extensions && eglGetError() == EGL_SUCCESS)
        __glfw.egl.EXT_client_extensions = GLFW_TRUE;

    if (__glfw.egl.EXT_client_extensions)
    {
        __glfw.egl.EXT_platform_base =
            ___glfwStringInExtensionString("EGL_EXT_platform_base", extensions);
        __glfw.egl.EXT_platform_x11 =
            ___glfwStringInExtensionString("EGL_EXT_platform_x11", extensions);
        __glfw.egl.EXT_platform_wayland =
            ___glfwStringInExtensionString("EGL_EXT_platform_wayland", extensions);
        __glfw.egl.ANGLE_platform_angle =
            ___glfwStringInExtensionString("EGL_ANGLE_platform_angle", extensions);
        __glfw.egl.ANGLE_platform_angle_opengl =
            ___glfwStringInExtensionString("EGL_ANGLE_platform_angle_opengl", extensions);
        __glfw.egl.ANGLE_platform_angle_d3d =
            ___glfwStringInExtensionString("EGL_ANGLE_platform_angle_d3d", extensions);
        __glfw.egl.ANGLE_platform_angle_vulkan =
            ___glfwStringInExtensionString("EGL_ANGLE_platform_angle_vulkan", extensions);
        __glfw.egl.ANGLE_platform_angle_metal =
            ___glfwStringInExtensionString("EGL_ANGLE_platform_angle_metal", extensions);
    }

    if (__glfw.egl.EXT_platform_base)
    {
        __glfw.egl.GetPlatformDisplayEXT = (PFNEGLGETPLATFORMDISPLAYEXTPROC)
            eglGetProcAddress("eglGetPlatformDisplayEXT");
        __glfw.egl.CreatePlatformWindowSurfaceEXT = (PFNEGLCREATEPLATFORMWINDOWSURFACEEXTPROC)
            eglGetProcAddress("eglCreatePlatformWindowSurfaceEXT");
    }

    __glfw.egl.platform = __glfw.platform.getEGLPlatform(&attribs);
    if (__glfw.egl.platform)
    {
        __glfw.egl.display =
            eglGetPlatformDisplayEXT(__glfw.egl.platform,
                                     __glfw.platform.getEGLNativeDisplay(),
                                     attribs);
    }
    else
        __glfw.egl.display = eglGetDisplay(__glfw.platform.getEGLNativeDisplay());

    __glfw_free(attribs);

    if (__glfw.egl.display == EGL_NO_DISPLAY)
    {
        ___glfwInputError(GLFW_API_UNAVAILABLE,
                        "EGL: Failed to get EGL display: %s",
                        getEGLErrorString(eglGetError()));

        ____glfwTerminateEGL();
        return GLFW_FALSE;
    }

    if (!eglInitialize(__glfw.egl.display, &__glfw.egl.major, &__glfw.egl.minor))
    {
        ___glfwInputError(GLFW_API_UNAVAILABLE,
                        "EGL: Failed to initialize EGL: %s",
                        getEGLErrorString(eglGetError()));

        ____glfwTerminateEGL();
        return GLFW_FALSE;
    }

    __glfw.egl.KHR_create_context =
        extensionSupportedEGL("EGL_KHR_create_context");
    __glfw.egl.KHR_create_context_no_error =
        extensionSupportedEGL("EGL_KHR_create_context_no_error");
    __glfw.egl.KHR_gl_colorspace =
        extensionSupportedEGL("EGL_KHR_gl_colorspace");
    __glfw.egl.KHR_get_all_proc_addresses =
        extensionSupportedEGL("EGL_KHR_get_all_proc_addresses");
    __glfw.egl.KHR_context_flush_control =
        extensionSupportedEGL("EGL_KHR_context_flush_control");
    __glfw.egl.EXT_present_opaque =
        extensionSupportedEGL("EGL_EXT_present_opaque");

    return GLFW_TRUE;
}

// Terminate EGL
//
void ____glfwTerminateEGL(void)
{
    if (__glfw.egl.display)
    {
        eglTerminate(__glfw.egl.display);
        __glfw.egl.display = EGL_NO_DISPLAY;
    }

    if (__glfw.egl.handle)
    {
        __glfwPlatformFreeModule(__glfw.egl.handle);
        __glfw.egl.handle = NULL;
    }
}

#define SET_ATTRIB(a, v) \
{ \
    assert(((size_t) index + 1) < sizeof(attribs) / sizeof(attribs[0])); \
    attribs[index++] = a; \
    attribs[index++] = v; \
}

// Create the OpenGL or OpenGL ES context
//
GLFWbool ___glfwCreateContextEGL(_GLFWwindow* window,
                               const _GLFWctxconfig* ctxconfig,
                               const _GLFWfbconfig* fbconfig)
{
    EGLint attribs[40];
    EGLConfig config;
    EGLContext share = NULL;
    EGLNativeWindowType native;
    int index = 0;

    if (!__glfw.egl.display)
    {
        ___glfwInputError(GLFW_API_UNAVAILABLE, "EGL: API not available");
        return GLFW_FALSE;
    }

    if (ctxconfig->share)
        share = ctxconfig->share->context.egl.handle;

    if (!chooseEGLConfig(ctxconfig, fbconfig, &config))
    {
        ___glfwInputError(GLFW_FORMAT_UNAVAILABLE,
                        "EGL: Failed to find a suitable EGLConfig");
        return GLFW_FALSE;
    }

    if (ctxconfig->client == GLFW_OPENGL_ES_API)
    {
        if (!eglBindAPI(EGL_OPENGL_ES_API))
        {
            ___glfwInputError(GLFW_API_UNAVAILABLE,
                            "EGL: Failed to bind OpenGL ES: %s",
                            getEGLErrorString(eglGetError()));
            return GLFW_FALSE;
        }
    }
    else
    {
        if (!eglBindAPI(EGL_OPENGL_API))
        {
            ___glfwInputError(GLFW_API_UNAVAILABLE,
                            "EGL: Failed to bind OpenGL: %s",
                            getEGLErrorString(eglGetError()));
            return GLFW_FALSE;
        }
    }

    if (__glfw.egl.KHR_create_context)
    {
        int mask = 0, flags = 0;

        if (ctxconfig->client == GLFW_OPENGL_API)
        {
            if (ctxconfig->forward)
                flags |= EGL_CONTEXT_OPENGL_FORWARD_COMPATIBLE_BIT_KHR;

            if (ctxconfig->profile == GLFW_OPENGL_CORE_PROFILE)
                mask |= EGL_CONTEXT_OPENGL_CORE_PROFILE_BIT_KHR;
            else if (ctxconfig->profile == GLFW_OPENGL_COMPAT_PROFILE)
                mask |= EGL_CONTEXT_OPENGL_COMPATIBILITY_PROFILE_BIT_KHR;
        }

        if (ctxconfig->debug)
            flags |= EGL_CONTEXT_OPENGL_DEBUG_BIT_KHR;

        if (ctxconfig->robustness)
        {
            if (ctxconfig->robustness == GLFW_NO_RESET_NOTIFICATION)
            {
                SET_ATTRIB(EGL_CONTEXT_OPENGL_RESET_NOTIFICATION_STRATEGY_KHR,
                           EGL_NO_RESET_NOTIFICATION_KHR);
            }
            else if (ctxconfig->robustness == GLFW_LOSE_CONTEXT_ON_RESET)
            {
                SET_ATTRIB(EGL_CONTEXT_OPENGL_RESET_NOTIFICATION_STRATEGY_KHR,
                           EGL_LOSE_CONTEXT_ON_RESET_KHR);
            }

            flags |= EGL_CONTEXT_OPENGL_ROBUST_ACCESS_BIT_KHR;
        }

        if (ctxconfig->noerror)
        {
            if (__glfw.egl.KHR_create_context_no_error)
                SET_ATTRIB(EGL_CONTEXT_OPENGL_NO_ERROR_KHR, GLFW_TRUE);
        }

        if (ctxconfig->major != 1 || ctxconfig->minor != 0)
        {
            SET_ATTRIB(EGL_CONTEXT_MAJOR_VERSION_KHR, ctxconfig->major);
            SET_ATTRIB(EGL_CONTEXT_MINOR_VERSION_KHR, ctxconfig->minor);
        }

        if (mask)
            SET_ATTRIB(EGL_CONTEXT_OPENGL_PROFILE_MASK_KHR, mask);

        if (flags)
            SET_ATTRIB(EGL_CONTEXT_FLAGS_KHR, flags);
    }
    else
    {
        if (ctxconfig->client == GLFW_OPENGL_ES_API)
            SET_ATTRIB(EGL_CONTEXT_CLIENT_VERSION, ctxconfig->major);
    }

    if (__glfw.egl.KHR_context_flush_control)
    {
        if (ctxconfig->release == GLFW_RELEASE_BEHAVIOR_NONE)
        {
            SET_ATTRIB(EGL_CONTEXT_RELEASE_BEHAVIOR_KHR,
                       EGL_CONTEXT_RELEASE_BEHAVIOR_NONE_KHR);
        }
        else if (ctxconfig->release == GLFW_RELEASE_BEHAVIOR_FLUSH)
        {
            SET_ATTRIB(EGL_CONTEXT_RELEASE_BEHAVIOR_KHR,
                       EGL_CONTEXT_RELEASE_BEHAVIOR_FLUSH_KHR);
        }
    }

    SET_ATTRIB(EGL_NONE, EGL_NONE);

    window->context.egl.handle = eglCreateContext(__glfw.egl.display,
                                                  config, share, attribs);

    if (window->context.egl.handle == EGL_NO_CONTEXT)
    {
        ___glfwInputError(GLFW_VERSION_UNAVAILABLE,
                        "EGL: Failed to create context: %s",
                        getEGLErrorString(eglGetError()));
        return GLFW_FALSE;
    }

    // Set up attributes for surface creation
    index = 0;

    if (fbconfig->sRGB)
    {
        if (__glfw.egl.KHR_gl_colorspace)
            SET_ATTRIB(EGL_GL_COLORSPACE_KHR, EGL_GL_COLORSPACE_SRGB_KHR);
    }

    if (!fbconfig->doublebuffer)
        SET_ATTRIB(EGL_RENDER_BUFFER, EGL_SINGLE_BUFFER);

    if (__glfw.egl.EXT_present_opaque)
        SET_ATTRIB(EGL_PRESENT_OPAQUE_EXT, !fbconfig->transparent);

    SET_ATTRIB(EGL_NONE, EGL_NONE);

    native = __glfw.platform.getEGLNativeWindow(window);
    // HACK: ANGLE does not implement eglCreatePlatformWindowSurfaceEXT
    //       despite reporting EGL_EXT_platform_base
    if (__glfw.egl.platform && __glfw.egl.platform != EGL_PLATFORM_ANGLE_ANGLE)
    {
        window->context.egl.surface =
            eglCreatePlatformWindowSurfaceEXT(__glfw.egl.display, config, native, attribs);
    }
    else
    {
        window->context.egl.surface =
            eglCreateWindowSurface(__glfw.egl.display, config, native, attribs);
    }

    if (window->context.egl.surface == EGL_NO_SURFACE)
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "EGL: Failed to create window surface: %s",
                        getEGLErrorString(eglGetError()));
        return GLFW_FALSE;
    }

    window->context.egl.config = config;

    // Load the appropriate client library
    if (!__glfw.egl.KHR_get_all_proc_addresses)
    {
        int i;
        const char** sonames;
        const char* es1sonames[] =
        {
#if defined(_GLFW_GLESV1_LIBRARY)
            _GLFW_GLESV1_LIBRARY,
#elif defined(_GLFW_WIN32)
            "GLESv1_CM.dll",
            "libGLES_CM.dll",
#elif defined(_GLFW_COCOA)
            "libGLESv1_CM.dylib",
#elif defined(__OpenBSD__) || defined(__NetBSD__)
            "libGLESv1_CM.so",
#else
            "libGLESv1_CM.so.1",
            "libGLES_CM.so.1",
#endif
            NULL
        };
        const char* es2sonames[] =
        {
#if defined(_GLFW_GLESV2_LIBRARY)
            _GLFW_GLESV2_LIBRARY,
#elif defined(_GLFW_WIN32)
            "GLESv2.dll",
            "libGLESv2.dll",
#elif defined(_GLFW_COCOA)
            "libGLESv2.dylib",
#elif defined(__CYGWIN__)
            "libGLESv2-2.so",
#elif defined(__OpenBSD__) || defined(__NetBSD__)
            "libGLESv2.so",
#else
            "libGLESv2.so.2",
#endif
            NULL
        };
        const char* glsonames[] =
        {
#if defined(_GLFW_OPENGL_LIBRARY)
            _GLFW_OPENGL_LIBRARY,
#elif defined(_GLFW_WIN32)
#elif defined(_GLFW_COCOA)
#elif defined(__OpenBSD__) || defined(__NetBSD__)
            "libGL.so",
#else
            "libOpenGL.so.0",
            "libGL.so.1",
#endif
            NULL
        };

        if (ctxconfig->client == GLFW_OPENGL_ES_API)
        {
            if (ctxconfig->major == 1)
                sonames = es1sonames;
            else
                sonames = es2sonames;
        }
        else
            sonames = glsonames;

        for (i = 0;  sonames[i];  i++)
        {
            // HACK: Match presence of lib prefix to increase chance of finding
            //       a matching pair in the jungle that is Win32 EGL/GLES
            if (__glfw.egl.prefix != (strncmp(sonames[i], "lib", 3) == 0))
                continue;

            window->context.egl.client = __glfwPlatformLoadModule(sonames[i]);
            if (window->context.egl.client)
                break;
        }

        if (!window->context.egl.client)
        {
            ___glfwInputError(GLFW_API_UNAVAILABLE,
                            "EGL: Failed to load client library");
            return GLFW_FALSE;
        }
    }

    window->context.makeCurrent = makeContextCurrentEGL;
    window->context.swapBuffers = swapBuffersEGL;
    window->context.swapInterval = swapIntervalEGL;
    window->context.extensionSupported = extensionSupportedEGL;
    window->context.getProcAddress = getProcAddressEGL;
    window->context.destroy = destroyContextEGL;

    return GLFW_TRUE;
}

#undef SET_ATTRIB

// Returns the Visual and depth of the chosen EGLConfig
//
#if defined(_GLFW_X11)
GLFWbool ___glfwChooseVisualEGL(const _GLFWwndconfig* wndconfig,
                              const _GLFWctxconfig* ctxconfig,
                              const _GLFWfbconfig* fbconfig,
                              Visual** visual, int* depth)
{
    XVisualInfo* result;
    XVisualInfo desired;
    EGLConfig native;
    EGLint visualID = 0, count = 0;
    const long vimask = VisualScreenMask | VisualIDMask;

    if (!chooseEGLConfig(ctxconfig, fbconfig, &native))
    {
        ___glfwInputError(GLFW_FORMAT_UNAVAILABLE,
                        "EGL: Failed to find a suitable EGLConfig");
        return GLFW_FALSE;
    }

    eglGetConfigAttrib(__glfw.egl.display, native,
                       EGL_NATIVE_VISUAL_ID, &visualID);

    desired.screen = __glfw.x11.screen;
    desired.visualid = visualID;

    result = XGetVisualInfo(__glfw.x11.display, vimask, &desired, &count);
    if (!result)
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "EGL: Failed to retrieve Visual for EGLConfig");
        return GLFW_FALSE;
    }

    *visual = result->visual;
    *depth = result->depth;

    XFree(result);
    return GLFW_TRUE;
}
#endif // _GLFW_X11


//////////////////////////////////////////////////////////////////////////
//////                        GLFW native API                       //////
//////////////////////////////////////////////////////////////////////////

GLFWAPI EGLDisplay __glfwGetEGLDisplay(void)
{
    _GLFW_REQUIRE_INIT_OR_RETURN(EGL_NO_DISPLAY);
    return __glfw.egl.display;
}

GLFWAPI EGLContext __glfwGetEGLContext(GLFWwindow* handle)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    _GLFW_REQUIRE_INIT_OR_RETURN(EGL_NO_CONTEXT);

    if (window->context.source != GLFW_EGL_CONTEXT_API)
    {
        ___glfwInputError(GLFW_NO_WINDOW_CONTEXT, NULL);
        return EGL_NO_CONTEXT;
    }

    return window->context.egl.handle;
}

GLFWAPI EGLSurface __glfwGetEGLSurface(GLFWwindow* handle)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    _GLFW_REQUIRE_INIT_OR_RETURN(EGL_NO_SURFACE);

    if (window->context.source != GLFW_EGL_CONTEXT_API)
    {
        ___glfwInputError(GLFW_NO_WINDOW_CONTEXT, NULL);
        return EGL_NO_SURFACE;
    }

    return window->context.egl.surface;
}

