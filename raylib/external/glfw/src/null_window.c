//========================================================================
// GLFW 3.4 - www.glfw.org
//------------------------------------------------------------------------
// Copyright (c) 2016 Google Inc.
// Copyright (c) 2016-2019 Camilla Löwy <elmindreda@glfw.org>
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
// It is fine to use C99 in this file because it will not be built with VS
//========================================================================

#include "internal.h"

#include <stdlib.h>

static void applySizeLimits(_GLFWwindow* window, int* width, int* height)
{
    if (window->numer != GLFW_DONT_CARE && window->denom != GLFW_DONT_CARE)
    {
        const float ratio = (float) window->numer / (float) window->denom;
        *height = (int) (*width / ratio);
    }

    if (window->minwidth != GLFW_DONT_CARE)
        *width = ___glfw_max(*width, window->minwidth);
    else if (window->maxwidth != GLFW_DONT_CARE)
        *width = ___glfw_min(*width, window->maxwidth);

    if (window->minheight != GLFW_DONT_CARE)
        *height = ___glfw_min(*height, window->minheight);
    else if (window->maxheight != GLFW_DONT_CARE)
        *height = ___glfw_max(*height, window->maxheight);
}

static void fitToMonitor(_GLFWwindow* window)
{
    GLFWvidmode mode;
    ___glfwGetVideoModeNull(window->monitor, &mode);
    ___glfwGetMonitorPosNull(window->monitor,
                           &window->null.xpos,
                           &window->null.ypos);
    window->null.width = mode.width;
    window->null.height = mode.height;
}

static void acquireMonitor(_GLFWwindow* window)
{
    ____glfwInputMonitorWindow(window->monitor, window);
}

static void releaseMonitor(_GLFWwindow* window)
{
    if (window->monitor->window != window)
        return;

    ____glfwInputMonitorWindow(window->monitor, NULL);
}

static int createNativeWindow(_GLFWwindow* window,
                              const _GLFWwndconfig* wndconfig,
                              const _GLFWfbconfig* fbconfig)
{
    if (window->monitor)
        fitToMonitor(window);
    else
    {
        if (wndconfig->xpos == GLFW_ANY_POSITION && wndconfig->ypos == GLFW_ANY_POSITION)
        {
            window->null.xpos = 17;
            window->null.ypos = 17;
        }
        else
        {
            window->null.xpos = wndconfig->xpos;
            window->null.ypos = wndconfig->ypos;
        }

        window->null.width = wndconfig->width;
        window->null.height = wndconfig->height;
    }

    window->null.visible = wndconfig->visible;
    window->null.decorated = wndconfig->decorated;
    window->null.maximized = wndconfig->maximized;
    window->null.floating = wndconfig->floating;
    window->null.transparent = fbconfig->transparent;
    window->null.opacity = 1.f;

    return GLFW_TRUE;
}


//////////////////////////////////////////////////////////////////////////
//////                       GLFW platform API                      //////
//////////////////////////////////////////////////////////////////////////

GLFWbool ___glfwCreateWindowNull(_GLFWwindow* window,
                               const _GLFWwndconfig* wndconfig,
                               const _GLFWctxconfig* ctxconfig,
                               const _GLFWfbconfig* fbconfig)
{
    if (!createNativeWindow(window, wndconfig, fbconfig))
        return GLFW_FALSE;

    if (ctxconfig->client != GLFW_NO_API)
    {
        if (ctxconfig->source == GLFW_NATIVE_CONTEXT_API ||
            ctxconfig->source == GLFW_OSMESA_CONTEXT_API)
        {
            if (!____glfwInitOSMesa())
                return GLFW_FALSE;
            if (!___glfwCreateContextOSMesa(window, ctxconfig, fbconfig))
                return GLFW_FALSE;
        }
        else if (ctxconfig->source == GLFW_EGL_CONTEXT_API)
        {
            if (!____glfwInitEGL())
                return GLFW_FALSE;
            if (!___glfwCreateContextEGL(window, ctxconfig, fbconfig))
                return GLFW_FALSE;
        }

        if (!___glfwRefreshContextAttribs(window, ctxconfig))
            return GLFW_FALSE;
    }

    if (wndconfig->mousePassthrough)
        __glfwSetWindowMousePassthroughNull(window, GLFW_TRUE);

    if (window->monitor)
    {
        ___glfwShowWindowNull(window);
        ___glfwFocusWindowNull(window);
        acquireMonitor(window);

        if (wndconfig->centerCursor)
            ___glfwCenterCursorInContentArea(window);
    }
    else
    {
        if (wndconfig->visible)
        {
            ___glfwShowWindowNull(window);
            if (wndconfig->focused)
                ___glfwFocusWindowNull(window);
        }
    }

    return GLFW_TRUE;
}

void ___glfwDestroyWindowNull(_GLFWwindow* window)
{
    if (window->monitor)
        releaseMonitor(window);

    if (__glfw.null.focusedWindow == window)
        __glfw.null.focusedWindow = NULL;

    if (window->context.destroy)
        window->context.destroy(window);
}

void ___glfwSetWindowTitleNull(_GLFWwindow* window, const char* title)
{
}

void ___glfwSetWindowIconNull(_GLFWwindow* window, int count, const GLFWimage* images)
{
}

void ___glfwSetWindowMonitorNull(_GLFWwindow* window,
                               _GLFWmonitor* monitor,
                               int xpos, int ypos,
                               int width, int height,
                               int refreshRate)
{
    if (window->monitor == monitor)
    {
        if (!monitor)
        {
            ___glfwSetWindowPosNull(window, xpos, ypos);
            ___glfwSetWindowSizeNull(window, width, height);
        }

        return;
    }

    if (window->monitor)
        releaseMonitor(window);

    ___glfwInputWindowMonitor(window, monitor);

    if (window->monitor)
    {
        window->null.visible = GLFW_TRUE;
        acquireMonitor(window);
        fitToMonitor(window);
    }
    else
    {
        ___glfwSetWindowPosNull(window, xpos, ypos);
        ___glfwSetWindowSizeNull(window, width, height);
    }
}

void ___glfwGetWindowPosNull(_GLFWwindow* window, int* xpos, int* ypos)
{
    if (xpos)
        *xpos = window->null.xpos;
    if (ypos)
        *ypos = window->null.ypos;
}

void ___glfwSetWindowPosNull(_GLFWwindow* window, int xpos, int ypos)
{
    if (window->monitor)
        return;

    if (window->null.xpos != xpos || window->null.ypos != ypos)
    {
        window->null.xpos = xpos;
        window->null.ypos = ypos;
        ___glfwInputWindowPos(window, xpos, ypos);
    }
}

void ___glfwGetWindowSizeNull(_GLFWwindow* window, int* width, int* height)
{
    if (width)
        *width = window->null.width;
    if (height)
        *height = window->null.height;
}

void ___glfwSetWindowSizeNull(_GLFWwindow* window, int width, int height)
{
    if (window->monitor)
        return;

    if (window->null.width != width || window->null.height != height)
    {
        window->null.width = width;
        window->null.height = height;
        ___glfwInputWindowSize(window, width, height);
        ___glfwInputFramebufferSize(window, width, height);
    }
}

void ____glfwSetWindowSizeLimitsNull(_GLFWwindow* window,
                                  int minwidth, int minheight,
                                  int maxwidth, int maxheight)
{
    int width = window->null.width;
    int height = window->null.height;
    applySizeLimits(window, &width, &height);
    ___glfwSetWindowSizeNull(window, width, height);
}

void ___glfwSetWindowAspectRatioNull(_GLFWwindow* window, int n, int d)
{
    int width = window->null.width;
    int height = window->null.height;
    applySizeLimits(window, &width, &height);
    ___glfwSetWindowSizeNull(window, width, height);
}

void ___glfwGetFramebufferSizeNull(_GLFWwindow* window, int* width, int* height)
{
    if (width)
        *width = window->null.width;
    if (height)
        *height = window->null.height;
}

void ___glfwGetWindowFrameSizeNull(_GLFWwindow* window,
                                 int* left, int* top,
                                 int* right, int* bottom)
{
    if (window->null.decorated && !window->monitor)
    {
        if (left)
            *left = 1;
        if (top)
            *top = 10;
        if (right)
            *right = 1;
        if (bottom)
            *bottom = 1;
    }
    else
    {
        if (left)
            *left = 0;
        if (top)
            *top = 0;
        if (right)
            *right = 0;
        if (bottom)
            *bottom = 0;
    }
}

void ___glfwGetWindowContentScaleNull(_GLFWwindow* window, float* xscale, float* yscale)
{
    if (xscale)
        *xscale = 1.f;
    if (yscale)
        *yscale = 1.f;
}

void ___glfwIconifyWindowNull(_GLFWwindow* window)
{
    if (__glfw.null.focusedWindow == window)
    {
        __glfw.null.focusedWindow = NULL;
        ___glfwInputWindowFocus(window, GLFW_FALSE);
    }

    if (!window->null.iconified)
    {
        window->null.iconified = GLFW_TRUE;
        ___glfwInputWindowIconify(window, GLFW_TRUE);

        if (window->monitor)
            releaseMonitor(window);
    }
}

void ___glfwRestoreWindowNull(_GLFWwindow* window)
{
    if (window->null.iconified)
    {
        window->null.iconified = GLFW_FALSE;
        ___glfwInputWindowIconify(window, GLFW_FALSE);

        if (window->monitor)
            acquireMonitor(window);
    }
    else if (window->null.maximized)
    {
        window->null.maximized = GLFW_FALSE;
        ___glfwInputWindowMaximize(window, GLFW_FALSE);
    }
}

void ___glfwMaximizeWindowNull(_GLFWwindow* window)
{
    if (!window->null.maximized)
    {
        window->null.maximized = GLFW_TRUE;
        ___glfwInputWindowMaximize(window, GLFW_TRUE);
    }
}

GLFWbool __glfwWindowMaximizedNull(_GLFWwindow* window)
{
    return window->null.maximized;
}

GLFWbool __glfwWindowHoveredNull(_GLFWwindow* window)
{
    return __glfw.null.xcursor >= window->null.xpos &&
           __glfw.null.ycursor >= window->null.ypos &&
           __glfw.null.xcursor <= window->null.xpos + window->null.width - 1 &&
           __glfw.null.ycursor <= window->null.ypos + window->null.height - 1;
}

GLFWbool __glfwFramebufferTransparentNull(_GLFWwindow* window)
{
    return window->null.transparent;
}

void __glfwSetWindowResizableNull(_GLFWwindow* window, GLFWbool enabled)
{
    window->null.resizable = enabled;
}

void __glfwSetWindowDecoratedNull(_GLFWwindow* window, GLFWbool enabled)
{
    window->null.decorated = enabled;
}

void __glfwSetWindowFloatingNull(_GLFWwindow* window, GLFWbool enabled)
{
    window->null.floating = enabled;
}

void __glfwSetWindowMousePassthroughNull(_GLFWwindow* window, GLFWbool enabled)
{
}

float ___glfwGetWindowOpacityNull(_GLFWwindow* window)
{
    return window->null.opacity;
}

void ___glfwSetWindowOpacityNull(_GLFWwindow* window, float opacity)
{
    window->null.opacity = opacity;
}

void __glfwSetRawMouseMotionNull(_GLFWwindow *window, GLFWbool enabled)
{
}

GLFWbool ___glfwRawMouseMotionSupportedNull(void)
{
    return GLFW_TRUE;
}

void ___glfwShowWindowNull(_GLFWwindow* window)
{
    window->null.visible = GLFW_TRUE;
}

void ___glfwRequestWindowAttentionNull(_GLFWwindow* window)
{
}

void ___glfwHideWindowNull(_GLFWwindow* window)
{
    if (__glfw.null.focusedWindow == window)
    {
        __glfw.null.focusedWindow = NULL;
        ___glfwInputWindowFocus(window, GLFW_FALSE);
    }

    window->null.visible = GLFW_FALSE;
}

void ___glfwFocusWindowNull(_GLFWwindow* window)
{
    _GLFWwindow* previous;

    if (__glfw.null.focusedWindow == window)
        return;

    if (!window->null.visible)
        return;

    previous = __glfw.null.focusedWindow;
    __glfw.null.focusedWindow = window;

    if (previous)
    {
        ___glfwInputWindowFocus(previous, GLFW_FALSE);
        if (previous->monitor && previous->autoIconify)
            ___glfwIconifyWindowNull(previous);
    }

    ___glfwInputWindowFocus(window, GLFW_TRUE);
}

GLFWbool __glfwWindowFocusedNull(_GLFWwindow* window)
{
    return __glfw.null.focusedWindow == window;
}

GLFWbool __glfwWindowIconifiedNull(_GLFWwindow* window)
{
    return window->null.iconified;
}

GLFWbool __glfwWindowVisibleNull(_GLFWwindow* window)
{
    return window->null.visible;
}

void ___glfwPollEventsNull(void)
{
}

void ___glfwWaitEventsNull(void)
{
}

void ____glfwWaitEventsTimeoutNull(double timeout)
{
}

void ___glfwPostEmptyEventNull(void)
{
}

void ___glfwGetCursorPosNull(_GLFWwindow* window, double* xpos, double* ypos)
{
    if (xpos)
        *xpos = __glfw.null.xcursor - window->null.xpos;
    if (ypos)
        *ypos = __glfw.null.ycursor - window->null.ypos;
}

void ____glfwSetCursorPosNull(_GLFWwindow* window, double x, double y)
{
    __glfw.null.xcursor = window->null.xpos + (int) x;
    __glfw.null.ycursor = window->null.ypos + (int) y;
}

void ___glfwSetCursorModeNull(_GLFWwindow* window, int mode)
{
}

GLFWbool ___glfwCreateCursorNull(_GLFWcursor* cursor,
                               const GLFWimage* image,
                               int xhot, int yhot)
{
    return GLFW_TRUE;
}

GLFWbool ___glfwCreateStandardCursorNull(_GLFWcursor* cursor, int shape)
{
    return GLFW_TRUE;
}

void ___glfwDestroyCursorNull(_GLFWcursor* cursor)
{
}

void ___glfwSetCursorNull(_GLFWwindow* window, _GLFWcursor* cursor)
{
}

void ___glfwSetClipboardStringNull(const char* string)
{
    char* copy = ___glfw_strdup(string);
    __glfw_free(__glfw.null.clipboardString);
    __glfw.null.clipboardString = copy;
}

const char* ___glfwGetClipboardStringNull(void)
{
    return __glfw.null.clipboardString;
}

EGLenum __glfwGetEGLPlatformNull(EGLint** attribs)
{
    return 0;
}

EGLNativeDisplayType __glfwGetEGLNativeDisplayNull(void)
{
    return 0;
}

EGLNativeWindowType __glfwGetEGLNativeWindowNull(_GLFWwindow* window)
{
    return 0;
}

const char* __glfwGetScancodeNameNull(int scancode)
{
    if (scancode < GLFW_KEY_SPACE || scancode > GLFW_KEY_LAST)
    {
        ___glfwInputError(GLFW_INVALID_VALUE, "Invalid scancode %i", scancode);
        return NULL;
    }

    switch (scancode)
    {
        case GLFW_KEY_APOSTROPHE:
            return "'";
        case GLFW_KEY_COMMA:
            return ",";
        case GLFW_KEY_MINUS:
        case GLFW_KEY_KP_SUBTRACT:
            return "-";
        case GLFW_KEY_PERIOD:
        case GLFW_KEY_KP_DECIMAL:
            return ".";
        case GLFW_KEY_SLASH:
        case GLFW_KEY_KP_DIVIDE:
            return "/";
        case GLFW_KEY_SEMICOLON:
            return ";";
        case GLFW_KEY_EQUAL:
        case GLFW_KEY_KP_EQUAL:
            return "=";
        case GLFW_KEY_LEFT_BRACKET:
            return "[";
        case GLFW_KEY_RIGHT_BRACKET:
            return "]";
        case GLFW_KEY_KP_MULTIPLY:
            return "*";
        case GLFW_KEY_KP_ADD:
            return "+";
        case GLFW_KEY_BACKSLASH:
        case GLFW_KEY_WORLD_1:
        case GLFW_KEY_WORLD_2:
            return "\\";
        case GLFW_KEY_0:
        case GLFW_KEY_KP_0:
            return "0";
        case GLFW_KEY_1:
        case GLFW_KEY_KP_1:
            return "1";
        case GLFW_KEY_2:
        case GLFW_KEY_KP_2:
            return "2";
        case GLFW_KEY_3:
        case GLFW_KEY_KP_3:
            return "3";
        case GLFW_KEY_4:
        case GLFW_KEY_KP_4:
            return "4";
        case GLFW_KEY_5:
        case GLFW_KEY_KP_5:
            return "5";
        case GLFW_KEY_6:
        case GLFW_KEY_KP_6:
            return "6";
        case GLFW_KEY_7:
        case GLFW_KEY_KP_7:
            return "7";
        case GLFW_KEY_8:
        case GLFW_KEY_KP_8:
            return "8";
        case GLFW_KEY_9:
        case GLFW_KEY_KP_9:
            return "9";
        case GLFW_KEY_A:
            return "a";
        case GLFW_KEY_B:
            return "b";
        case GLFW_KEY_C:
            return "c";
        case GLFW_KEY_D:
            return "d";
        case GLFW_KEY_E:
            return "e";
        case GLFW_KEY_F:
            return "f";
        case GLFW_KEY_G:
            return "g";
        case GLFW_KEY_H:
            return "h";
        case GLFW_KEY_I:
            return "i";
        case GLFW_KEY_J:
            return "j";
        case GLFW_KEY_K:
            return "k";
        case GLFW_KEY_L:
            return "l";
        case GLFW_KEY_M:
            return "m";
        case GLFW_KEY_N:
            return "n";
        case GLFW_KEY_O:
            return "o";
        case GLFW_KEY_P:
            return "p";
        case GLFW_KEY_Q:
            return "q";
        case GLFW_KEY_R:
            return "r";
        case GLFW_KEY_S:
            return "s";
        case GLFW_KEY_T:
            return "t";
        case GLFW_KEY_U:
            return "u";
        case GLFW_KEY_V:
            return "v";
        case GLFW_KEY_W:
            return "w";
        case GLFW_KEY_X:
            return "x";
        case GLFW_KEY_Y:
            return "y";
        case GLFW_KEY_Z:
            return "z";
    }

    return NULL;
}

int ____glfwGetKeyScancodeNull(int key)
{
    return key;
}

void ___glfwGetRequiredInstanceExtensionsNull(char** extensions)
{
}

GLFWbool ___glfwGetPhysicalDevicePresentationSupportNull(VkInstance instance,
                                                       VkPhysicalDevice device,
                                                       uint32_t queuefamily)
{
    return GLFW_FALSE;
}

VkResult ____glfwCreateWindowSurfaceNull(VkInstance instance,
                                      _GLFWwindow* window,
                                      const VkAllocationCallbacks* allocator,
                                      VkSurfaceKHR* surface)
{
    // This seems like the most appropriate error to return here
    return VK_ERROR_EXTENSION_NOT_PRESENT;
}

