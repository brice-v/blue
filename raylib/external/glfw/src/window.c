//========================================================================
// GLFW 3.4 - www.glfw.org
//------------------------------------------------------------------------
// Copyright (c) 2002-2006 Marcus Geelnard
// Copyright (c) 2006-2019 Camilla Löwy <elmindreda@glfw.org>
// Copyright (c) 2012 Torsten Walluhn <tw@mad-cad.net>
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

#include <assert.h>
#include <string.h>
#include <stdlib.h>
#include <float.h>


//////////////////////////////////////////////////////////////////////////
//////                         GLFW event API                       //////
//////////////////////////////////////////////////////////////////////////

// Notifies shared code that a window has lost or received input focus
//
void ___glfwInputWindowFocus(_GLFWwindow* window, GLFWbool focused)
{
    assert(window != NULL);
    assert(focused == GLFW_TRUE || focused == GLFW_FALSE);

    if (window->callbacks.focus)
        window->callbacks.focus((GLFWwindow*) window, focused);

    if (!focused)
    {
        int key, button;

        for (key = 0;  key <= GLFW_KEY_LAST;  key++)
        {
            if (window->keys[key] == GLFW_PRESS)
            {
                const int scancode = __glfw.platform.getKeyScancode(key);
                ___glfwInputKey(window, key, scancode, GLFW_RELEASE, 0);
            }
        }

        for (button = 0;  button <= GLFW_MOUSE_BUTTON_LAST;  button++)
        {
            if (window->mouseButtons[button] == GLFW_PRESS)
                ___glfwInputMouseClick(window, button, GLFW_RELEASE, 0);
        }
    }
}

// Notifies shared code that a window has moved
// The position is specified in content area relative screen coordinates
//
void ___glfwInputWindowPos(_GLFWwindow* window, int x, int y)
{
    assert(window != NULL);

    if (window->callbacks.pos)
        window->callbacks.pos((GLFWwindow*) window, x, y);
}

// Notifies shared code that a window has been resized
// The size is specified in screen coordinates
//
void ___glfwInputWindowSize(_GLFWwindow* window, int width, int height)
{
    assert(window != NULL);
    assert(width >= 0);
    assert(height >= 0);

    if (window->callbacks.size)
        window->callbacks.size((GLFWwindow*) window, width, height);
}

// Notifies shared code that a window has been iconified or restored
//
void ___glfwInputWindowIconify(_GLFWwindow* window, GLFWbool iconified)
{
    assert(window != NULL);
    assert(iconified == GLFW_TRUE || iconified == GLFW_FALSE);

    if (window->callbacks.iconify)
        window->callbacks.iconify((GLFWwindow*) window, iconified);
}

// Notifies shared code that a window has been maximized or restored
//
void ___glfwInputWindowMaximize(_GLFWwindow* window, GLFWbool maximized)
{
    assert(window != NULL);
    assert(maximized == GLFW_TRUE || maximized == GLFW_FALSE);

    if (window->callbacks.maximize)
        window->callbacks.maximize((GLFWwindow*) window, maximized);
}

// Notifies shared code that a window framebuffer has been resized
// The size is specified in pixels
//
void ___glfwInputFramebufferSize(_GLFWwindow* window, int width, int height)
{
    assert(window != NULL);
    assert(width >= 0);
    assert(height >= 0);

    if (window->callbacks.fbsize)
        window->callbacks.fbsize((GLFWwindow*) window, width, height);
}

// Notifies shared code that a window content scale has changed
// The scale is specified as the ratio between the current and default DPI
//
void ___glfwInputWindowContentScale(_GLFWwindow* window, float xscale, float yscale)
{
    assert(window != NULL);
    assert(xscale > 0.f);
    assert(xscale < FLT_MAX);
    assert(yscale > 0.f);
    assert(yscale < FLT_MAX);

    if (window->callbacks.scale)
        window->callbacks.scale((GLFWwindow*) window, xscale, yscale);
}

// Notifies shared code that the window contents needs updating
//
void ___glfwInputWindowDamage(_GLFWwindow* window)
{
    assert(window != NULL);

    if (window->callbacks.refresh)
        window->callbacks.refresh((GLFWwindow*) window);
}

// Notifies shared code that the user wishes to close a window
//
void ___glfwInputWindowCloseRequest(_GLFWwindow* window)
{
    assert(window != NULL);

    window->shouldClose = GLFW_TRUE;

    if (window->callbacks.close)
        window->callbacks.close((GLFWwindow*) window);
}

// Notifies shared code that a window has changed its desired monitor
//
void ___glfwInputWindowMonitor(_GLFWwindow* window, _GLFWmonitor* monitor)
{
    assert(window != NULL);
    window->monitor = monitor;
}

//////////////////////////////////////////////////////////////////////////
//////                        GLFW public API                       //////
//////////////////////////////////////////////////////////////////////////

GLFWAPI GLFWwindow* __glfwCreateWindow(int width, int height,
                                     const char* title,
                                     GLFWmonitor* monitor,
                                     GLFWwindow* share)
{
    _GLFWfbconfig fbconfig;
    _GLFWctxconfig ctxconfig;
    _GLFWwndconfig wndconfig;
    _GLFWwindow* window;

    assert(title != NULL);
    assert(width >= 0);
    assert(height >= 0);

    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);

    if (width <= 0 || height <= 0)
    {
        ___glfwInputError(GLFW_INVALID_VALUE,
                        "Invalid window size %ix%i",
                        width, height);

        return NULL;
    }

    fbconfig  = __glfw.hints.framebuffer;
    ctxconfig = __glfw.hints.context;
    wndconfig = __glfw.hints.window;

    wndconfig.width   = width;
    wndconfig.height  = height;
    wndconfig.title   = title;
    ctxconfig.share   = (_GLFWwindow*) share;

    if (!___glfwIsValidContextConfig(&ctxconfig))
        return NULL;

    window = __glfw_calloc(1, sizeof(_GLFWwindow));
    window->next = __glfw.windowListHead;
    __glfw.windowListHead = window;

    window->videoMode.width       = width;
    window->videoMode.height      = height;
    window->videoMode.redBits     = fbconfig.redBits;
    window->videoMode.greenBits   = fbconfig.greenBits;
    window->videoMode.blueBits    = fbconfig.blueBits;
    window->videoMode.refreshRate = __glfw.hints.refreshRate;

    window->monitor          = (_GLFWmonitor*) monitor;
    window->resizable        = wndconfig.resizable;
    window->decorated        = wndconfig.decorated;
    window->autoIconify      = wndconfig.autoIconify;
    window->floating         = wndconfig.floating;
    window->focusOnShow      = wndconfig.focusOnShow;
    window->mousePassthrough = wndconfig.mousePassthrough;
    window->cursorMode       = GLFW_CURSOR_NORMAL;

    window->doublebuffer = fbconfig.doublebuffer;

    window->minwidth    = GLFW_DONT_CARE;
    window->minheight   = GLFW_DONT_CARE;
    window->maxwidth    = GLFW_DONT_CARE;
    window->maxheight   = GLFW_DONT_CARE;
    window->numer       = GLFW_DONT_CARE;
    window->denom       = GLFW_DONT_CARE;

    if (!__glfw.platform.createWindow(window, &wndconfig, &ctxconfig, &fbconfig))
    {
        __glfwDestroyWindow((GLFWwindow*) window);
        return NULL;
    }

    return (GLFWwindow*) window;
}

void __glfwDefaultWindowHints(void)
{
    _GLFW_REQUIRE_INIT();

    // The default is OpenGL with minimum version 1.0
    memset(&__glfw.hints.context, 0, sizeof(__glfw.hints.context));
    __glfw.hints.context.client = GLFW_OPENGL_API;
    __glfw.hints.context.source = GLFW_NATIVE_CONTEXT_API;
    __glfw.hints.context.major  = 1;
    __glfw.hints.context.minor  = 0;

    // The default is a focused, visible, resizable window with decorations
    memset(&__glfw.hints.window, 0, sizeof(__glfw.hints.window));
    __glfw.hints.window.resizable    = GLFW_TRUE;
    __glfw.hints.window.visible      = GLFW_TRUE;
    __glfw.hints.window.decorated    = GLFW_TRUE;
    __glfw.hints.window.focused      = GLFW_TRUE;
    __glfw.hints.window.autoIconify  = GLFW_TRUE;
    __glfw.hints.window.centerCursor = GLFW_TRUE;
    __glfw.hints.window.focusOnShow  = GLFW_TRUE;
    __glfw.hints.window.xpos         = GLFW_ANY_POSITION;
    __glfw.hints.window.ypos         = GLFW_ANY_POSITION;

    // The default is 24 bits of color, 24 bits of depth and 8 bits of stencil,
    // double buffered
    memset(&__glfw.hints.framebuffer, 0, sizeof(__glfw.hints.framebuffer));
    __glfw.hints.framebuffer.redBits      = 8;
    __glfw.hints.framebuffer.greenBits    = 8;
    __glfw.hints.framebuffer.blueBits     = 8;
    __glfw.hints.framebuffer.alphaBits    = 8;
    __glfw.hints.framebuffer.depthBits    = 24;
    __glfw.hints.framebuffer.stencilBits  = 8;
    __glfw.hints.framebuffer.doublebuffer = GLFW_TRUE;

    // The default is to select the highest available refresh rate
    __glfw.hints.refreshRate = GLFW_DONT_CARE;

    // The default is to use full Retina resolution framebuffers
    __glfw.hints.window.ns.retina = GLFW_TRUE;
}

GLFWAPI void __glfwWindowHint(int hint, int value)
{
    _GLFW_REQUIRE_INIT();

    switch (hint)
    {
        case GLFW_RED_BITS:
            __glfw.hints.framebuffer.redBits = value;
            return;
        case GLFW_GREEN_BITS:
            __glfw.hints.framebuffer.greenBits = value;
            return;
        case GLFW_BLUE_BITS:
            __glfw.hints.framebuffer.blueBits = value;
            return;
        case GLFW_ALPHA_BITS:
            __glfw.hints.framebuffer.alphaBits = value;
            return;
        case GLFW_DEPTH_BITS:
            __glfw.hints.framebuffer.depthBits = value;
            return;
        case GLFW_STENCIL_BITS:
            __glfw.hints.framebuffer.stencilBits = value;
            return;
        case GLFW_ACCUM_RED_BITS:
            __glfw.hints.framebuffer.accumRedBits = value;
            return;
        case GLFW_ACCUM_GREEN_BITS:
            __glfw.hints.framebuffer.accumGreenBits = value;
            return;
        case GLFW_ACCUM_BLUE_BITS:
            __glfw.hints.framebuffer.accumBlueBits = value;
            return;
        case GLFW_ACCUM_ALPHA_BITS:
            __glfw.hints.framebuffer.accumAlphaBits = value;
            return;
        case GLFW_AUX_BUFFERS:
            __glfw.hints.framebuffer.auxBuffers = value;
            return;
        case GLFW_STEREO:
            __glfw.hints.framebuffer.stereo = value ? GLFW_TRUE : GLFW_FALSE;
            return;
        case GLFW_DOUBLEBUFFER:
            __glfw.hints.framebuffer.doublebuffer = value ? GLFW_TRUE : GLFW_FALSE;
            return;
        case GLFW_TRANSPARENT_FRAMEBUFFER:
            __glfw.hints.framebuffer.transparent = value ? GLFW_TRUE : GLFW_FALSE;
            return;
        case GLFW_SAMPLES:
            __glfw.hints.framebuffer.samples = value;
            return;
        case GLFW_SRGB_CAPABLE:
            __glfw.hints.framebuffer.sRGB = value ? GLFW_TRUE : GLFW_FALSE;
            return;
        case GLFW_RESIZABLE:
            __glfw.hints.window.resizable = value ? GLFW_TRUE : GLFW_FALSE;
            return;
        case GLFW_DECORATED:
            __glfw.hints.window.decorated = value ? GLFW_TRUE : GLFW_FALSE;
            return;
        case GLFW_FOCUSED:
            __glfw.hints.window.focused = value ? GLFW_TRUE : GLFW_FALSE;
            return;
        case GLFW_AUTO_ICONIFY:
            __glfw.hints.window.autoIconify = value ? GLFW_TRUE : GLFW_FALSE;
            return;
        case GLFW_FLOATING:
            __glfw.hints.window.floating = value ? GLFW_TRUE : GLFW_FALSE;
            return;
        case GLFW_MAXIMIZED:
            __glfw.hints.window.maximized = value ? GLFW_TRUE : GLFW_FALSE;
            return;
        case GLFW_VISIBLE:
            __glfw.hints.window.visible = value ? GLFW_TRUE : GLFW_FALSE;
            return;
        case GLFW_POSITION_X:
            __glfw.hints.window.xpos = value;
            return;
        case GLFW_POSITION_Y:
            __glfw.hints.window.ypos = value;
            return;
        case GLFW_COCOA_RETINA_FRAMEBUFFER:
            __glfw.hints.window.ns.retina = value ? GLFW_TRUE : GLFW_FALSE;
            return;
        case GLFW_WIN32_KEYBOARD_MENU:
            __glfw.hints.window.win32.keymenu = value ? GLFW_TRUE : GLFW_FALSE;
            return;
        case GLFW_COCOA_GRAPHICS_SWITCHING:
            __glfw.hints.context.nsgl.offline = value ? GLFW_TRUE : GLFW_FALSE;
            return;
        case GLFW_SCALE_TO_MONITOR:
            __glfw.hints.window.scaleToMonitor = value ? GLFW_TRUE : GLFW_FALSE;
            return;
        case GLFW_CENTER_CURSOR:
            __glfw.hints.window.centerCursor = value ? GLFW_TRUE : GLFW_FALSE;
            return;
        case GLFW_FOCUS_ON_SHOW:
            __glfw.hints.window.focusOnShow = value ? GLFW_TRUE : GLFW_FALSE;
            return;
        case GLFW_MOUSE_PASSTHROUGH:
            __glfw.hints.window.mousePassthrough = value ? GLFW_TRUE : GLFW_FALSE;
            return;
        case GLFW_CLIENT_API:
            __glfw.hints.context.client = value;
            return;
        case GLFW_CONTEXT_CREATION_API:
            __glfw.hints.context.source = value;
            return;
        case GLFW_CONTEXT_VERSION_MAJOR:
            __glfw.hints.context.major = value;
            return;
        case GLFW_CONTEXT_VERSION_MINOR:
            __glfw.hints.context.minor = value;
            return;
        case GLFW_CONTEXT_ROBUSTNESS:
            __glfw.hints.context.robustness = value;
            return;
        case GLFW_OPENGL_FORWARD_COMPAT:
            __glfw.hints.context.forward = value ? GLFW_TRUE : GLFW_FALSE;
            return;
        case GLFW_CONTEXT_DEBUG:
            __glfw.hints.context.debug = value ? GLFW_TRUE : GLFW_FALSE;
            return;
        case GLFW_CONTEXT_NO_ERROR:
            __glfw.hints.context.noerror = value ? GLFW_TRUE : GLFW_FALSE;
            return;
        case GLFW_OPENGL_PROFILE:
            __glfw.hints.context.profile = value;
            return;
        case GLFW_CONTEXT_RELEASE_BEHAVIOR:
            __glfw.hints.context.release = value;
            return;
        case GLFW_REFRESH_RATE:
            __glfw.hints.refreshRate = value;
            return;
    }

    ___glfwInputError(GLFW_INVALID_ENUM, "Invalid window hint 0x%08X", hint);
}

GLFWAPI void ___glfwWindowHintString(int hint, const char* value)
{
    assert(value != NULL);

    _GLFW_REQUIRE_INIT();

    switch (hint)
    {
        case GLFW_COCOA_FRAME_NAME:
            strncpy(__glfw.hints.window.ns.frameName, value,
                    sizeof(__glfw.hints.window.ns.frameName) - 1);
            return;
        case GLFW_X11_CLASS_NAME:
            strncpy(__glfw.hints.window.x11.className, value,
                    sizeof(__glfw.hints.window.x11.className) - 1);
            return;
        case GLFW_X11_INSTANCE_NAME:
            strncpy(__glfw.hints.window.x11.instanceName, value,
                    sizeof(__glfw.hints.window.x11.instanceName) - 1);
            return;
        case GLFW_WAYLAND_APP_ID:
            strncpy(__glfw.hints.window.wl.appId, value,
                    sizeof(__glfw.hints.window.wl.appId) - 1);
            return;
    }

    ___glfwInputError(GLFW_INVALID_ENUM, "Invalid window hint string 0x%08X", hint);
}

GLFWAPI void __glfwDestroyWindow(GLFWwindow* handle)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;

    _GLFW_REQUIRE_INIT();

    // Allow closing of NULL (to match the behavior of free)
    if (window == NULL)
        return;

    // Clear all callbacks to avoid exposing a half torn-down window object
    memset(&window->callbacks, 0, sizeof(window->callbacks));

    // The window's context must not be current on another thread when the
    // window is destroyed
    if (window == ___glfwPlatformGetTls(&__glfw.contextSlot))
        __glfwMakeContextCurrent(NULL);

    __glfw.platform.destroyWindow(window);

    // Unlink window from global linked list
    {
        _GLFWwindow** prev = &__glfw.windowListHead;

        while (*prev != window)
            prev = &((*prev)->next);

        *prev = window->next;
    }

    __glfw_free(window);
}

GLFWAPI int __glfwWindowShouldClose(GLFWwindow* handle)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT_OR_RETURN(0);
    return window->shouldClose;
}

GLFWAPI void __glfwSetWindowShouldClose(GLFWwindow* handle, int value)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT();
    window->shouldClose = value;
}

GLFWAPI void __glfwSetWindowTitle(GLFWwindow* handle, const char* title)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);
    assert(title != NULL);

    _GLFW_REQUIRE_INIT();
    __glfw.platform.setWindowTitle(window, title);
}

GLFWAPI void __glfwSetWindowIcon(GLFWwindow* handle,
                               int count, const GLFWimage* images)
{
    int i;
    _GLFWwindow* window = (_GLFWwindow*) handle;

    assert(window != NULL);
    assert(count >= 0);
    assert(count == 0 || images != NULL);

    _GLFW_REQUIRE_INIT();

    if (count < 0)
    {
        ___glfwInputError(GLFW_INVALID_VALUE, "Invalid image count for window icon");
        return;
    }

    for (i = 0; i < count; i++)
    {
        assert(images[i].pixels != NULL);

        if (images[i].width <= 0 || images[i].height <= 0)
        {
            ___glfwInputError(GLFW_INVALID_VALUE,
                            "Invalid image dimensions for window icon");
            return;
        }
    }

    __glfw.platform.setWindowIcon(window, count, images);
}

GLFWAPI void __glfwGetWindowPos(GLFWwindow* handle, int* xpos, int* ypos)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    if (xpos)
        *xpos = 0;
    if (ypos)
        *ypos = 0;

    _GLFW_REQUIRE_INIT();
    __glfw.platform.getWindowPos(window, xpos, ypos);
}

GLFWAPI void __glfwSetWindowPos(GLFWwindow* handle, int xpos, int ypos)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT();

    if (window->monitor)
        return;

    __glfw.platform.setWindowPos(window, xpos, ypos);
}

GLFWAPI void __glfwGetWindowSize(GLFWwindow* handle, int* width, int* height)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    if (width)
        *width = 0;
    if (height)
        *height = 0;

    _GLFW_REQUIRE_INIT();
    __glfw.platform.getWindowSize(window, width, height);
}

GLFWAPI void __glfwSetWindowSize(GLFWwindow* handle, int width, int height)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);
    assert(width >= 0);
    assert(height >= 0);

    _GLFW_REQUIRE_INIT();

    window->videoMode.width  = width;
    window->videoMode.height = height;

    __glfw.platform.setWindowSize(window, width, height);
}

GLFWAPI void ___glfwSetWindowSizeLimits(GLFWwindow* handle,
                                     int minwidth, int minheight,
                                     int maxwidth, int maxheight)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT();

    if (minwidth != GLFW_DONT_CARE && minheight != GLFW_DONT_CARE)
    {
        if (minwidth < 0 || minheight < 0)
        {
            ___glfwInputError(GLFW_INVALID_VALUE,
                            "Invalid window minimum size %ix%i",
                            minwidth, minheight);
            return;
        }
    }

    if (maxwidth != GLFW_DONT_CARE && maxheight != GLFW_DONT_CARE)
    {
        if (maxwidth < 0 || maxheight < 0 ||
            maxwidth < minwidth || maxheight < minheight)
        {
            ___glfwInputError(GLFW_INVALID_VALUE,
                            "Invalid window maximum size %ix%i",
                            maxwidth, maxheight);
            return;
        }
    }

    window->minwidth  = minwidth;
    window->minheight = minheight;
    window->maxwidth  = maxwidth;
    window->maxheight = maxheight;

    if (window->monitor || !window->resizable)
        return;

    __glfw.platform.setWindowSizeLimits(window,
                                       minwidth, minheight,
                                       maxwidth, maxheight);
}

GLFWAPI void __glfwSetWindowAspectRatio(GLFWwindow* handle, int numer, int denom)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);
    assert(numer != 0);
    assert(denom != 0);

    _GLFW_REQUIRE_INIT();

    if (numer != GLFW_DONT_CARE && denom != GLFW_DONT_CARE)
    {
        if (numer <= 0 || denom <= 0)
        {
            ___glfwInputError(GLFW_INVALID_VALUE,
                            "Invalid window aspect ratio %i:%i",
                            numer, denom);
            return;
        }
    }

    window->numer = numer;
    window->denom = denom;

    if (window->monitor || !window->resizable)
        return;

    __glfw.platform.setWindowAspectRatio(window, numer, denom);
}

GLFWAPI void __glfwGetFramebufferSize(GLFWwindow* handle, int* width, int* height)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    if (width)
        *width = 0;
    if (height)
        *height = 0;

    _GLFW_REQUIRE_INIT();
    __glfw.platform.getFramebufferSize(window, width, height);
}

GLFWAPI void __glfwGetWindowFrameSize(GLFWwindow* handle,
                                    int* left, int* top,
                                    int* right, int* bottom)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    if (left)
        *left = 0;
    if (top)
        *top = 0;
    if (right)
        *right = 0;
    if (bottom)
        *bottom = 0;

    _GLFW_REQUIRE_INIT();
    __glfw.platform.getWindowFrameSize(window, left, top, right, bottom);
}

GLFWAPI void __glfwGetWindowContentScale(GLFWwindow* handle,
                                       float* xscale, float* yscale)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    if (xscale)
        *xscale = 0.f;
    if (yscale)
        *yscale = 0.f;

    _GLFW_REQUIRE_INIT();
    __glfw.platform.getWindowContentScale(window, xscale, yscale);
}

GLFWAPI float __glfwGetWindowOpacity(GLFWwindow* handle)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT_OR_RETURN(1.f);
    return __glfw.platform.getWindowOpacity(window);
}

GLFWAPI void __glfwSetWindowOpacity(GLFWwindow* handle, float opacity)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);
    assert(opacity == opacity);
    assert(opacity >= 0.f);
    assert(opacity <= 1.f);

    _GLFW_REQUIRE_INIT();

    if (opacity != opacity || opacity < 0.f || opacity > 1.f)
    {
        ___glfwInputError(GLFW_INVALID_VALUE, "Invalid window opacity %f", opacity);
        return;
    }

    __glfw.platform.setWindowOpacity(window, opacity);
}

GLFWAPI void __glfwIconifyWindow(GLFWwindow* handle)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT();
    __glfw.platform.iconifyWindow(window);
}

GLFWAPI void __glfwRestoreWindow(GLFWwindow* handle)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT();
    __glfw.platform.restoreWindow(window);
}

GLFWAPI void __glfwMaximizeWindow(GLFWwindow* handle)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT();

    if (window->monitor)
        return;

    __glfw.platform.maximizeWindow(window);
}

GLFWAPI void __glfwShowWindow(GLFWwindow* handle)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT();

    if (window->monitor)
        return;

    __glfw.platform.showWindow(window);

    if (window->focusOnShow)
        __glfw.platform.focusWindow(window);
}

GLFWAPI void __glfwRequestWindowAttention(GLFWwindow* handle)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT();

    __glfw.platform.requestWindowAttention(window);
}

GLFWAPI void __glfwHideWindow(GLFWwindow* handle)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT();

    if (window->monitor)
        return;

    __glfw.platform.hideWindow(window);
}

GLFWAPI void __glfwFocusWindow(GLFWwindow* handle)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT();

    __glfw.platform.focusWindow(window);
}

GLFWAPI int __glfwGetWindowAttrib(GLFWwindow* handle, int attrib)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT_OR_RETURN(0);

    switch (attrib)
    {
        case GLFW_FOCUSED:
            return __glfw.platform.windowFocused(window);
        case GLFW_ICONIFIED:
            return __glfw.platform.windowIconified(window);
        case GLFW_VISIBLE:
            return __glfw.platform.windowVisible(window);
        case GLFW_MAXIMIZED:
            return __glfw.platform.windowMaximized(window);
        case GLFW_HOVERED:
            return __glfw.platform.windowHovered(window);
        case GLFW_FOCUS_ON_SHOW:
            return window->focusOnShow;
        case GLFW_MOUSE_PASSTHROUGH:
            return window->mousePassthrough;
        case GLFW_TRANSPARENT_FRAMEBUFFER:
            return __glfw.platform.framebufferTransparent(window);
        case GLFW_RESIZABLE:
            return window->resizable;
        case GLFW_DECORATED:
            return window->decorated;
        case GLFW_FLOATING:
            return window->floating;
        case GLFW_AUTO_ICONIFY:
            return window->autoIconify;
        case GLFW_DOUBLEBUFFER:
            return window->doublebuffer;
        case GLFW_CLIENT_API:
            return window->context.client;
        case GLFW_CONTEXT_CREATION_API:
            return window->context.source;
        case GLFW_CONTEXT_VERSION_MAJOR:
            return window->context.major;
        case GLFW_CONTEXT_VERSION_MINOR:
            return window->context.minor;
        case GLFW_CONTEXT_REVISION:
            return window->context.revision;
        case GLFW_CONTEXT_ROBUSTNESS:
            return window->context.robustness;
        case GLFW_OPENGL_FORWARD_COMPAT:
            return window->context.forward;
        case GLFW_CONTEXT_DEBUG:
            return window->context.debug;
        case GLFW_OPENGL_PROFILE:
            return window->context.profile;
        case GLFW_CONTEXT_RELEASE_BEHAVIOR:
            return window->context.release;
        case GLFW_CONTEXT_NO_ERROR:
            return window->context.noerror;
    }

    ___glfwInputError(GLFW_INVALID_ENUM, "Invalid window attribute 0x%08X", attrib);
    return 0;
}

GLFWAPI void __glfwSetWindowAttrib(GLFWwindow* handle, int attrib, int value)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT();

    value = value ? GLFW_TRUE : GLFW_FALSE;

    switch (attrib)
    {
        case GLFW_AUTO_ICONIFY:
            window->autoIconify = value;
            return;

        case GLFW_RESIZABLE:
            window->resizable = value;
            if (!window->monitor)
                __glfw.platform.setWindowResizable(window, value);
            return;

        case GLFW_DECORATED:
            window->decorated = value;
            if (!window->monitor)
                __glfw.platform.setWindowDecorated(window, value);
            return;

        case GLFW_FLOATING:
            window->floating = value;
            if (!window->monitor)
                __glfw.platform.setWindowFloating(window, value);
            return;

        case GLFW_FOCUS_ON_SHOW:
            window->focusOnShow = value;
            return;

        case GLFW_MOUSE_PASSTHROUGH:
            window->mousePassthrough = value;
            __glfw.platform.setWindowMousePassthrough(window, value);
            return;
    }

    ___glfwInputError(GLFW_INVALID_ENUM, "Invalid window attribute 0x%08X", attrib);
}

GLFWAPI GLFWmonitor* __glfwGetWindowMonitor(GLFWwindow* handle)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);
    return (GLFWmonitor*) window->monitor;
}

GLFWAPI void __glfwSetWindowMonitor(GLFWwindow* wh,
                                  GLFWmonitor* mh,
                                  int xpos, int ypos,
                                  int width, int height,
                                  int refreshRate)
{
    _GLFWwindow* window = (_GLFWwindow*) wh;
    _GLFWmonitor* monitor = (_GLFWmonitor*) mh;
    assert(window != NULL);
    assert(width >= 0);
    assert(height >= 0);

    _GLFW_REQUIRE_INIT();

    if (width <= 0 || height <= 0)
    {
        ___glfwInputError(GLFW_INVALID_VALUE,
                        "Invalid window size %ix%i",
                        width, height);
        return;
    }

    if (refreshRate < 0 && refreshRate != GLFW_DONT_CARE)
    {
        ___glfwInputError(GLFW_INVALID_VALUE,
                        "Invalid refresh rate %i",
                        refreshRate);
        return;
    }

    window->videoMode.width       = width;
    window->videoMode.height      = height;
    window->videoMode.refreshRate = refreshRate;

    __glfw.platform.setWindowMonitor(window, monitor,
                                    xpos, ypos, width, height,
                                    refreshRate);
}

GLFWAPI void __glfwSetWindowUserPointer(GLFWwindow* handle, void* pointer)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT();
    window->userPointer = pointer;
}

GLFWAPI void* __glfwGetWindowUserPointer(GLFWwindow* handle)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);
    return window->userPointer;
}

GLFWAPI GLFWwindowposfun ___glfwSetWindowPosCallback(GLFWwindow* handle,
                                                  GLFWwindowposfun cbfun)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);
    _GLFW_SWAP(GLFWwindowposfun, window->callbacks.pos, cbfun);
    return cbfun;
}

GLFWAPI GLFWwindowsizefun ___glfwSetWindowSizeCallback(GLFWwindow* handle,
                                                    GLFWwindowsizefun cbfun)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);
    _GLFW_SWAP(GLFWwindowsizefun, window->callbacks.size, cbfun);
    return cbfun;
}

GLFWAPI GLFWwindowclosefun __glfwSetWindowCloseCallback(GLFWwindow* handle,
                                                      GLFWwindowclosefun cbfun)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);
    _GLFW_SWAP(GLFWwindowclosefun, window->callbacks.close, cbfun);
    return cbfun;
}

GLFWAPI GLFWwindowrefreshfun __glfwSetWindowRefreshCallback(GLFWwindow* handle,
                                                          GLFWwindowrefreshfun cbfun)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);
    _GLFW_SWAP(GLFWwindowrefreshfun, window->callbacks.refresh, cbfun);
    return cbfun;
}

GLFWAPI GLFWwindowfocusfun __glfwSetWindowFocusCallback(GLFWwindow* handle,
                                                      GLFWwindowfocusfun cbfun)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);
    _GLFW_SWAP(GLFWwindowfocusfun, window->callbacks.focus, cbfun);
    return cbfun;
}

GLFWAPI GLFWwindowiconifyfun ___glfwSetWindowIconifyCallback(GLFWwindow* handle,
                                                          GLFWwindowiconifyfun cbfun)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);
    _GLFW_SWAP(GLFWwindowiconifyfun, window->callbacks.iconify, cbfun);
    return cbfun;
}

GLFWAPI GLFWwindowmaximizefun __glfwSetWindowMaximizeCallback(GLFWwindow* handle,
                                                            GLFWwindowmaximizefun cbfun)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);
    _GLFW_SWAP(GLFWwindowmaximizefun, window->callbacks.maximize, cbfun);
    return cbfun;
}

GLFWAPI GLFWframebuffersizefun __glfwSetFramebufferSizeCallback(GLFWwindow* handle,
                                                              GLFWframebuffersizefun cbfun)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);
    _GLFW_SWAP(GLFWframebuffersizefun, window->callbacks.fbsize, cbfun);
    return cbfun;
}

GLFWAPI GLFWwindowcontentscalefun __glfwSetWindowContentScaleCallback(GLFWwindow* handle,
                                                                    GLFWwindowcontentscalefun cbfun)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    assert(window != NULL);

    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);
    _GLFW_SWAP(GLFWwindowcontentscalefun, window->callbacks.scale, cbfun);
    return cbfun;
}

GLFWAPI void __glfwPollEvents(void)
{
    _GLFW_REQUIRE_INIT();
    __glfw.platform.pollEvents();
}

GLFWAPI void __glfwWaitEvents(void)
{
    _GLFW_REQUIRE_INIT();
    __glfw.platform.waitEvents();
}

GLFWAPI void ___glfwWaitEventsTimeout(double timeout)
{
    _GLFW_REQUIRE_INIT();
    assert(timeout == timeout);
    assert(timeout >= 0.0);
    assert(timeout <= DBL_MAX);

    if (timeout != timeout || timeout < 0.0 || timeout > DBL_MAX)
    {
        ___glfwInputError(GLFW_INVALID_VALUE, "Invalid time %f", timeout);
        return;
    }

    __glfw.platform.waitEventsTimeout(timeout);
}

GLFWAPI void __glfwPostEmptyEvent(void)
{
    _GLFW_REQUIRE_INIT();
    __glfw.platform.postEmptyEvent();
}

