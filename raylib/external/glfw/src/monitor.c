//========================================================================
// GLFW 3.4 - www.glfw.org
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

#include <assert.h>
#include <math.h>
#include <float.h>
#include <string.h>
#include <stdlib.h>
#include <limits.h>


// Lexically compare video modes, used by qsort
//
static int compareVideoModes(const void* fp, const void* sp)
{
    const GLFWvidmode* fm = fp;
    const GLFWvidmode* sm = sp;
    const int fbpp = fm->redBits + fm->greenBits + fm->blueBits;
    const int sbpp = sm->redBits + sm->greenBits + sm->blueBits;
    const int farea = fm->width * fm->height;
    const int sarea = sm->width * sm->height;

    // First sort on color bits per pixel
    if (fbpp != sbpp)
        return fbpp - sbpp;

    // Then sort on screen area
    if (farea != sarea)
        return farea - sarea;

    // Then sort on width
    if (fm->width != sm->width)
        return fm->width - sm->width;

    // Lastly sort on refresh rate
    return fm->refreshRate - sm->refreshRate;
}

// Retrieves the available modes for the specified monitor
//
static GLFWbool refreshVideoModes(_GLFWmonitor* monitor)
{
    int modeCount;
    GLFWvidmode* modes;

    if (monitor->modes)
        return GLFW_TRUE;

    modes = __glfw.platform.getVideoModes(monitor, &modeCount);
    if (!modes)
        return GLFW_FALSE;

    qsort(modes, modeCount, sizeof(GLFWvidmode), compareVideoModes);

    __glfw_free(monitor->modes);
    monitor->modes = modes;
    monitor->modeCount = modeCount;

    return GLFW_TRUE;
}


//////////////////////////////////////////////////////////////////////////
//////                         GLFW event API                       //////
//////////////////////////////////////////////////////////////////////////

// Notifies shared code of a monitor connection or disconnection
//
void ___glfwInputMonitor(_GLFWmonitor* monitor, int action, int placement)
{
    assert(monitor != NULL);
    assert(action == GLFW_CONNECTED || action == GLFW_DISCONNECTED);
    assert(placement == _GLFW_INSERT_FIRST || placement == _GLFW_INSERT_LAST);

    if (action == GLFW_CONNECTED)
    {
        __glfw.monitorCount++;
        __glfw.monitors =
            __glfw_realloc(__glfw.monitors,
                          sizeof(_GLFWmonitor*) * __glfw.monitorCount);

        if (placement == _GLFW_INSERT_FIRST)
        {
            memmove(__glfw.monitors + 1,
                    __glfw.monitors,
                    ((size_t) __glfw.monitorCount - 1) * sizeof(_GLFWmonitor*));
            __glfw.monitors[0] = monitor;
        }
        else
            __glfw.monitors[__glfw.monitorCount - 1] = monitor;
    }
    else if (action == GLFW_DISCONNECTED)
    {
        int i;
        _GLFWwindow* window;

        for (window = __glfw.windowListHead;  window;  window = window->next)
        {
            if (window->monitor == monitor)
            {
                int width, height, xoff, yoff;
                __glfw.platform.getWindowSize(window, &width, &height);
                __glfw.platform.setWindowMonitor(window, NULL, 0, 0, width, height, 0);
                __glfw.platform.getWindowFrameSize(window, &xoff, &yoff, NULL, NULL);
                __glfw.platform.setWindowPos(window, xoff, yoff);
            }
        }

        for (i = 0;  i < __glfw.monitorCount;  i++)
        {
            if (__glfw.monitors[i] == monitor)
            {
                __glfw.monitorCount--;
                memmove(__glfw.monitors + i,
                        __glfw.monitors + i + 1,
                        ((size_t) __glfw.monitorCount - i) * sizeof(_GLFWmonitor*));
                break;
            }
        }
    }

    if (__glfw.callbacks.monitor)
        __glfw.callbacks.monitor((GLFWmonitor*) monitor, action);

    if (action == GLFW_DISCONNECTED)
        ___glfwFreeMonitor(monitor);
}

// Notifies shared code that a full screen window has acquired or released
// a monitor
//
void ____glfwInputMonitorWindow(_GLFWmonitor* monitor, _GLFWwindow* window)
{
    assert(monitor != NULL);
    monitor->windowww = window;
}


//////////////////////////////////////////////////////////////////////////
//////                       GLFW internal API                      //////
//////////////////////////////////////////////////////////////////////////

// Allocates and returns a monitor object with the specified name and dimensions
//
_GLFWmonitor* ___glfwAllocMonitor(const char* name, int widthMM, int heightMM)
{
    _GLFWmonitor* monitor = __glfw_calloc(1, sizeof(_GLFWmonitor));
    monitor->widthMM = widthMM;
    monitor->heightMM = heightMM;

    strncpy(monitor->name, name, sizeof(monitor->name) - 1);

    return monitor;
}

// Frees a monitor object and any data associated with it
//
void ___glfwFreeMonitor(_GLFWmonitor* monitor)
{
    if (monitor == NULL)
        return;

    __glfw.platform.freeMonitor(monitor);

    ___glfwFreeGammaArrays(&monitor->originalRamp);
    ___glfwFreeGammaArrays(&monitor->currentRamp);

    __glfw_free(monitor->modes);
    __glfw_free(monitor);
}

// Allocates red, green and blue value arrays of the specified size
//
void ___glfwAllocGammaArrays(GLFWgammaramp* ramp, unsigned int size)
{
    ramp->red = __glfw_calloc(size, sizeof(unsigned short));
    ramp->green = __glfw_calloc(size, sizeof(unsigned short));
    ramp->blue = __glfw_calloc(size, sizeof(unsigned short));
    ramp->size = size;
}

// Frees the red, green and blue value arrays and clears the struct
//
void ___glfwFreeGammaArrays(GLFWgammaramp* ramp)
{
    __glfw_free(ramp->red);
    __glfw_free(ramp->green);
    __glfw_free(ramp->blue);

    memset(ramp, 0, sizeof(GLFWgammaramp));
}

// Chooses the video mode most closely matching the desired one
//
const GLFWvidmode* ___glfwChooseVideoMode(_GLFWmonitor* monitor,
                                        const GLFWvidmode* desired)
{
    int i;
    unsigned int sizeDiff, leastSizeDiff = UINT_MAX;
    unsigned int rateDiff, leastRateDiff = UINT_MAX;
    unsigned int colorDiff, leastColorDiff = UINT_MAX;
    const GLFWvidmode* current;
    const GLFWvidmode* closest = NULL;

    if (!refreshVideoModes(monitor))
        return NULL;

    for (i = 0;  i < monitor->modeCount;  i++)
    {
        current = monitor->modes + i;

        colorDiff = 0;

        if (desired->redBits != GLFW_DONT_CARE)
            colorDiff += abs(current->redBits - desired->redBits);
        if (desired->greenBits != GLFW_DONT_CARE)
            colorDiff += abs(current->greenBits - desired->greenBits);
        if (desired->blueBits != GLFW_DONT_CARE)
            colorDiff += abs(current->blueBits - desired->blueBits);

        sizeDiff = abs((current->width - desired->width) *
                       (current->width - desired->width) +
                       (current->height - desired->height) *
                       (current->height - desired->height));

        if (desired->refreshRate != GLFW_DONT_CARE)
            rateDiff = abs(current->refreshRate - desired->refreshRate);
        else
            rateDiff = UINT_MAX - current->refreshRate;

        if ((colorDiff < leastColorDiff) ||
            (colorDiff == leastColorDiff && sizeDiff < leastSizeDiff) ||
            (colorDiff == leastColorDiff && sizeDiff == leastSizeDiff && rateDiff < leastRateDiff))
        {
            closest = current;
            leastSizeDiff = sizeDiff;
            leastRateDiff = rateDiff;
            leastColorDiff = colorDiff;
        }
    }

    return closest;
}

// Performs lexical comparison between two @ref GLFWvidmode structures
//
int ___glfwCompareVideoModes(const GLFWvidmode* fm, const GLFWvidmode* sm)
{
    return compareVideoModes(fm, sm);
}

// Splits a color depth into red, green and blue bit depths
//
void ___glfwSplitBPP(int bpp, int* red, int* green, int* blue)
{
    int delta;

    // We assume that by 32 the user really meant 24
    if (bpp == 32)
        bpp = 24;

    // Convert "bits per pixel" to red, green & blue sizes

    *red = *green = *blue = bpp / 3;
    delta = bpp - (*red * 3);
    if (delta >= 1)
        *green = *green + 1;

    if (delta == 2)
        *red = *red + 1;
}


//////////////////////////////////////////////////////////////////////////
//////                        GLFW public API                       //////
//////////////////////////////////////////////////////////////////////////

GLFWAPI GLFWmonitor** __glfwGetMonitors(int* count)
{
    assert(count != NULL);

    *count = 0;

    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);

    *count = __glfw.monitorCount;
    return (GLFWmonitor**) __glfw.monitors;
}

GLFWAPI GLFWmonitor* __glfwGetPrimaryMonitor(void)
{
    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);

    if (!__glfw.monitorCount)
        return NULL;

    return (GLFWmonitor*) __glfw.monitors[0];
}

GLFWAPI void __glfwGetMonitorPos(GLFWmonitor* handle, int* xpos, int* ypos)
{
    _GLFWmonitor* monitor = (_GLFWmonitor*) handle;
    assert(monitor != NULL);

    if (xpos)
        *xpos = 0;
    if (ypos)
        *ypos = 0;

    _GLFW_REQUIRE_INIT();

    __glfw.platform.getMonitorPos(monitor, xpos, ypos);
}

GLFWAPI void __glfwGetMonitorWorkarea(GLFWmonitor* handle,
                                    int* xpos, int* ypos,
                                    int* width, int* height)
{
    _GLFWmonitor* monitor = (_GLFWmonitor*) handle;
    assert(monitor != NULL);

    if (xpos)
        *xpos = 0;
    if (ypos)
        *ypos = 0;
    if (width)
        *width = 0;
    if (height)
        *height = 0;

    _GLFW_REQUIRE_INIT();

    __glfw.platform.getMonitorWorkarea(monitor, xpos, ypos, width, height);
}

GLFWAPI void __glfwGetMonitorPhysicalSize(GLFWmonitor* handle, int* widthMM, int* heightMM)
{
    _GLFWmonitor* monitor = (_GLFWmonitor*) handle;
    assert(monitor != NULL);

    if (widthMM)
        *widthMM = 0;
    if (heightMM)
        *heightMM = 0;

    _GLFW_REQUIRE_INIT();

    if (widthMM)
        *widthMM = monitor->widthMM;
    if (heightMM)
        *heightMM = monitor->heightMM;
}

GLFWAPI void __glfwGetMonitorContentScale(GLFWmonitor* handle,
                                        float* xscale, float* yscale)
{
    _GLFWmonitor* monitor = (_GLFWmonitor*) handle;
    assert(monitor != NULL);

    if (xscale)
        *xscale = 0.f;
    if (yscale)
        *yscale = 0.f;

    _GLFW_REQUIRE_INIT();
    __glfw.platform.getMonitorContentScale(monitor, xscale, yscale);
}

GLFWAPI const char* __glfwGetMonitorName(GLFWmonitor* handle)
{
    _GLFWmonitor* monitor = (_GLFWmonitor*) handle;
    assert(monitor != NULL);

    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);
    return monitor->name;
}

GLFWAPI void __glfwSetMonitorUserPointer(GLFWmonitor* handle, void* pointer)
{
    _GLFWmonitor* monitor = (_GLFWmonitor*) handle;
    assert(monitor != NULL);

    _GLFW_REQUIRE_INIT();
    monitor->userPointer = pointer;
}

GLFWAPI void* __glfwGetMonitorUserPointer(GLFWmonitor* handle)
{
    _GLFWmonitor* monitor = (_GLFWmonitor*) handle;
    assert(monitor != NULL);

    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);
    return monitor->userPointer;
}

GLFWAPI GLFWmonitorfun __glfwSetMonitorCallback(GLFWmonitorfun cbfun)
{
    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);
    _GLFW_SWAP(GLFWmonitorfun, __glfw.callbacks.monitor, cbfun);
    return cbfun;
}

GLFWAPI const GLFWvidmode* ___glfwGetVideoModes(GLFWmonitor* handle, int* count)
{
    _GLFWmonitor* monitor = (_GLFWmonitor*) handle;
    assert(monitor != NULL);
    assert(count != NULL);

    *count = 0;

    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);

    if (!refreshVideoModes(monitor))
        return NULL;

    *count = monitor->modeCount;
    return monitor->modes;
}

GLFWAPI const GLFWvidmode* __glfwGetVideoMode(GLFWmonitor* handle)
{
    _GLFWmonitor* monitor = (_GLFWmonitor*) handle;
    assert(monitor != NULL);

    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);

    __glfw.platform.getVideoMode(monitor, &monitor->currentMode);
    return &monitor->currentMode;
}

GLFWAPI void __glfwSetGamma(GLFWmonitor* handle, float gamma)
{
    unsigned int i;
    unsigned short* values;
    GLFWgammaramp ramp;
    const GLFWgammaramp* original;
    assert(handle != NULL);
    assert(gamma > 0.f);
    assert(gamma <= FLT_MAX);

    _GLFW_REQUIRE_INIT();

    if (gamma != gamma || gamma <= 0.f || gamma > FLT_MAX)
    {
        ___glfwInputError(GLFW_INVALID_VALUE, "Invalid gamma value %f", gamma);
        return;
    }

    original = __glfwGetGammaRamp(handle);
    if (!original)
        return;

    values = __glfw_calloc(original->size, sizeof(unsigned short));

    for (i = 0;  i < original->size;  i++)
    {
        float value;

        // Calculate intensity
        value = i / (float) (original->size - 1);
        // Apply gamma curve
        value = powf(value, 1.f / gamma) * 65535.f + 0.5f;
        // Clamp to value range
        value = ___glfw_fminf(value, 65535.f);

        values[i] = (unsigned short) value;
    }

    ramp.red = values;
    ramp.green = values;
    ramp.blue = values;
    ramp.size = original->size;

    ___glfwSetGammaRamp(handle, &ramp);
    __glfw_free(values);
}

GLFWAPI const GLFWgammaramp* __glfwGetGammaRamp(GLFWmonitor* handle)
{
    _GLFWmonitor* monitor = (_GLFWmonitor*) handle;
    assert(monitor != NULL);

    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);

    ___glfwFreeGammaArrays(&monitor->currentRamp);
    if (!__glfw.platform.getGammaRamp(monitor, &monitor->currentRamp))
        return NULL;

    return &monitor->currentRamp;
}

GLFWAPI void ___glfwSetGammaRamp(GLFWmonitor* handle, const GLFWgammaramp* ramp)
{
    _GLFWmonitor* monitor = (_GLFWmonitor*) handle;
    assert(monitor != NULL);
    assert(ramp != NULL);
    assert(ramp->size > 0);
    assert(ramp->red != NULL);
    assert(ramp->green != NULL);
    assert(ramp->blue != NULL);

    _GLFW_REQUIRE_INIT();

    if (ramp->size <= 0)
    {
        ___glfwInputError(GLFW_INVALID_VALUE,
                        "Invalid gamma ramp size %i",
                        ramp->size);
        return;
    }

    if (!monitor->originalRamp.size)
    {
        if (!__glfw.platform.getGammaRamp(monitor, &monitor->originalRamp))
            return;
    }

    __glfw.platform.setGammaRamp(monitor, ramp);
}

