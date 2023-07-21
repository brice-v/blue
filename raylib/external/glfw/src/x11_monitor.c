//========================================================================
// GLFW 3.4 X11 - www.glfw.org
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
// It is fine to use C99 in this file because it will not be built with VS
//========================================================================

#include "internal.h"

#include <limits.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>


// Check whether the display mode should be included in enumeration
//
static GLFWbool modeIsGood(const XRRModeInfo* mi)
{
    return (mi->modeFlags & RR_Interlace) == 0;
}

// Calculates the refresh rate, in Hz, from the specified RandR mode info
//
static int calculateRefreshRate(const XRRModeInfo* mi)
{
    if (mi->hTotal && mi->vTotal)
        return (int) round((double) mi->dotClock / ((double) mi->hTotal * (double) mi->vTotal));
    else
        return 0;
}

// Returns the mode info for a RandR mode XID
//
static const XRRModeInfo* getModeInfo(const XRRScreenResources* sr, RRMode id)
{
    for (int i = 0;  i < sr->nmode;  i++)
    {
        if (sr->modes[i].id == id)
            return sr->modes + i;
    }

    return NULL;
}

// Convert RandR mode info to GLFW video mode
//
static GLFWvidmode vidmodeFromModeInfo(const XRRModeInfo* mi,
                                       const XRRCrtcInfo* ci)
{
    GLFWvidmode mode;

    if (ci->rotation == RR_Rotate_90 || ci->rotation == RR_Rotate_270)
    {
        mode.width  = mi->height;
        mode.height = mi->width;
    }
    else
    {
        mode.width  = mi->width;
        mode.height = mi->height;
    }

    mode.refreshRate = calculateRefreshRate(mi);

    ___glfwSplitBPP(DefaultDepth(__glfw.x11.display, __glfw.x11.screen),
                  &mode.redBits, &mode.greenBits, &mode.blueBits);

    return mode;
}


//////////////////////////////////////////////////////////////////////////
//////                       GLFW internal API                      //////
//////////////////////////////////////////////////////////////////////////

// Poll for changes in the set of connected monitors
//
void ___glfwPollMonitorsX11(void)
{
    if (__glfw.x11.randr.available && !__glfw.x11.randr.monitorBroken)
    {
        int disconnectedCount, screenCount = 0;
        _GLFWmonitor** disconnected = NULL;
        XineramaScreenInfo* screens = NULL;
        XRRScreenResources* sr = XRRGetScreenResourcesCurrent(__glfw.x11.display,
                                                              __glfw.x11.root);
        RROutput primary = XRRGetOutputPrimary(__glfw.x11.display,
                                               __glfw.x11.root);

        if (__glfw.x11.xinerama.available)
            screens = XineramaQueryScreens(__glfw.x11.display, &screenCount);

        disconnectedCount = __glfw.monitorCount;
        if (disconnectedCount)
        {
            disconnected = __glfw_calloc(__glfw.monitorCount, sizeof(_GLFWmonitor*));
            memcpy(disconnected,
                   __glfw.monitors,
                   __glfw.monitorCount * sizeof(_GLFWmonitor*));
        }

        for (int i = 0;  i < sr->noutput;  i++)
        {
            int j, type, widthMM, heightMM;

            XRROutputInfo* oi = XRRGetOutputInfo(__glfw.x11.display, sr, sr->outputs[i]);
            if (oi->connection != RR_Connected || oi->crtc == None)
            {
                XRRFreeOutputInfo(oi);
                continue;
            }

            for (j = 0;  j < disconnectedCount;  j++)
            {
                if (disconnected[j] &&
                    disconnected[j]->x11.output == sr->outputs[i])
                {
                    disconnected[j] = NULL;
                    break;
                }
            }

            if (j < disconnectedCount)
            {
                XRRFreeOutputInfo(oi);
                continue;
            }

            XRRCrtcInfo* ci = XRRGetCrtcInfo(__glfw.x11.display, sr, oi->crtc);
            if (ci->rotation == RR_Rotate_90 || ci->rotation == RR_Rotate_270)
            {
                widthMM  = oi->mm_height;
                heightMM = oi->mm_width;
            }
            else
            {
                widthMM  = oi->mm_width;
                heightMM = oi->mm_height;
            }

            if (widthMM <= 0 || heightMM <= 0)
            {
                // HACK: If RandR does not provide a physical size, assume the
                //       X11 default 96 DPI and calculate from the CRTC viewport
                // NOTE: These members are affected by rotation, unlike the mode
                //       info and output info members
                widthMM  = (int) (ci->width * 25.4f / 96.f);
                heightMM = (int) (ci->height * 25.4f / 96.f);
            }

            _GLFWmonitor* monitor = ___glfwAllocMonitor(oi->name, widthMM, heightMM);
            monitor->x11.output = sr->outputs[i];
            monitor->x11.crtc   = oi->crtc;

            for (j = 0;  j < screenCount;  j++)
            {
                if (screens[j].x_org == ci->x &&
                    screens[j].y_org == ci->y &&
                    screens[j].width == ci->width &&
                    screens[j].height == ci->height)
                {
                    monitor->x11.index = j;
                    break;
                }
            }

            if (monitor->x11.output == primary)
                type = _GLFW_INSERT_FIRST;
            else
                type = _GLFW_INSERT_LAST;

            ___glfwInputMonitor(monitor, GLFW_CONNECTED, type);

            XRRFreeOutputInfo(oi);
            XRRFreeCrtcInfo(ci);
        }

        XRRFreeScreenResources(sr);

        if (screens)
            XFree(screens);

        for (int i = 0;  i < disconnectedCount;  i++)
        {
            if (disconnected[i])
                ___glfwInputMonitor(disconnected[i], GLFW_DISCONNECTED, 0);
        }

        __glfw_free(disconnected);
    }
    else
    {
        const int widthMM = DisplayWidthMM(__glfw.x11.display, __glfw.x11.screen);
        const int heightMM = DisplayHeightMM(__glfw.x11.display, __glfw.x11.screen);

        ___glfwInputMonitor(___glfwAllocMonitor("Display", widthMM, heightMM),
                          GLFW_CONNECTED,
                          _GLFW_INSERT_FIRST);
    }
}

// Set the current video mode for the specified monitor
//
void ___glfwSetVideoModeX11(_GLFWmonitor* monitor, const GLFWvidmode* desired)
{
    if (__glfw.x11.randr.available && !__glfw.x11.randr.monitorBroken)
    {
        GLFWvidmode current;
        RRMode native = None;

        const GLFWvidmode* best = ___glfwChooseVideoMode(monitor, desired);
        ___glfwGetVideoModeX11(monitor, &current);
        if (___glfwCompareVideoModes(&current, best) == 0)
            return;

        XRRScreenResources* sr =
            XRRGetScreenResourcesCurrent(__glfw.x11.display, __glfw.x11.root);
        XRRCrtcInfo* ci = XRRGetCrtcInfo(__glfw.x11.display, sr, monitor->x11.crtc);
        XRROutputInfo* oi = XRRGetOutputInfo(__glfw.x11.display, sr, monitor->x11.output);

        for (int i = 0;  i < oi->nmode;  i++)
        {
            const XRRModeInfo* mi = getModeInfo(sr, oi->modes[i]);
            if (!modeIsGood(mi))
                continue;

            const GLFWvidmode mode = vidmodeFromModeInfo(mi, ci);
            if (___glfwCompareVideoModes(best, &mode) == 0)
            {
                native = mi->id;
                break;
            }
        }

        if (native)
        {
            if (monitor->x11.oldMode == None)
                monitor->x11.oldMode = ci->mode;

            XRRSetCrtcConfig(__glfw.x11.display,
                             sr, monitor->x11.crtc,
                             CurrentTime,
                             ci->x, ci->y,
                             native,
                             ci->rotation,
                             ci->outputs,
                             ci->noutput);
        }

        XRRFreeOutputInfo(oi);
        XRRFreeCrtcInfo(ci);
        XRRFreeScreenResources(sr);
    }
}

// Restore the saved (original) video mode for the specified monitor
//
void ___glfwRestoreVideoModeX11(_GLFWmonitor* monitor)
{
    if (__glfw.x11.randr.available && !__glfw.x11.randr.monitorBroken)
    {
        if (monitor->x11.oldMode == None)
            return;

        XRRScreenResources* sr =
            XRRGetScreenResourcesCurrent(__glfw.x11.display, __glfw.x11.root);
        XRRCrtcInfo* ci = XRRGetCrtcInfo(__glfw.x11.display, sr, monitor->x11.crtc);

        XRRSetCrtcConfig(__glfw.x11.display,
                         sr, monitor->x11.crtc,
                         CurrentTime,
                         ci->x, ci->y,
                         monitor->x11.oldMode,
                         ci->rotation,
                         ci->outputs,
                         ci->noutput);

        XRRFreeCrtcInfo(ci);
        XRRFreeScreenResources(sr);

        monitor->x11.oldMode = None;
    }
}


//////////////////////////////////////////////////////////////////////////
//////                       GLFW platform API                      //////
//////////////////////////////////////////////////////////////////////////

void ___glfwFreeMonitorX11(_GLFWmonitor* monitor)
{
}

void ___glfwGetMonitorPosX11(_GLFWmonitor* monitor, int* xpos, int* ypos)
{
    if (__glfw.x11.randr.available && !__glfw.x11.randr.monitorBroken)
    {
        XRRScreenResources* sr =
            XRRGetScreenResourcesCurrent(__glfw.x11.display, __glfw.x11.root);
        XRRCrtcInfo* ci = XRRGetCrtcInfo(__glfw.x11.display, sr, monitor->x11.crtc);

        if (ci)
        {
            if (xpos)
                *xpos = ci->x;
            if (ypos)
                *ypos = ci->y;

            XRRFreeCrtcInfo(ci);
        }

        XRRFreeScreenResources(sr);
    }
}

void ___glfwGetMonitorContentScaleX11(_GLFWmonitor* monitor,
                                    float* xscale, float* yscale)
{
    if (xscale)
        *xscale = __glfw.x11.contentScaleX;
    if (yscale)
        *yscale = __glfw.x11.contentScaleY;
}

void ___glfwGetMonitorWorkareaX11(_GLFWmonitor* monitor,
                                int* xpos, int* ypos,
                                int* width, int* height)
{
    int areaX = 0, areaY = 0, areaWidth = 0, areaHeight = 0;

    if (__glfw.x11.randr.available && !__glfw.x11.randr.monitorBroken)
    {
        XRRScreenResources* sr =
            XRRGetScreenResourcesCurrent(__glfw.x11.display, __glfw.x11.root);
        XRRCrtcInfo* ci = XRRGetCrtcInfo(__glfw.x11.display, sr, monitor->x11.crtc);

        areaX = ci->x;
        areaY = ci->y;

        const XRRModeInfo* mi = getModeInfo(sr, ci->mode);

        if (ci->rotation == RR_Rotate_90 || ci->rotation == RR_Rotate_270)
        {
            areaWidth  = mi->height;
            areaHeight = mi->width;
        }
        else
        {
            areaWidth  = mi->width;
            areaHeight = mi->height;
        }

        XRRFreeCrtcInfo(ci);
        XRRFreeScreenResources(sr);
    }
    else
    {
        areaWidth  = DisplayWidth(__glfw.x11.display, __glfw.x11.screen);
        areaHeight = DisplayHeight(__glfw.x11.display, __glfw.x11.screen);
    }

    if (__glfw.x11.NET_WORKAREA && __glfw.x11.NET_CURRENT_DESKTOP)
    {
        Atom* extents = NULL;
        Atom* desktop = NULL;
        const unsigned long extentCount =
            ___glfwGetWindowPropertyX11(__glfw.x11.root,
                                      __glfw.x11.NET_WORKAREA,
                                      XA_CARDINAL,
                                      (unsigned char**) &extents);

        if (___glfwGetWindowPropertyX11(__glfw.x11.root,
                                      __glfw.x11.NET_CURRENT_DESKTOP,
                                      XA_CARDINAL,
                                      (unsigned char**) &desktop) > 0)
        {
            if (extentCount >= 4 && *desktop < extentCount / 4)
            {
                const int globalX = extents[*desktop * 4 + 0];
                const int globalY = extents[*desktop * 4 + 1];
                const int globalWidth  = extents[*desktop * 4 + 2];
                const int globalHeight = extents[*desktop * 4 + 3];

                if (areaX < globalX)
                {
                    areaWidth -= globalX - areaX;
                    areaX = globalX;
                }

                if (areaY < globalY)
                {
                    areaHeight -= globalY - areaY;
                    areaY = globalY;
                }

                if (areaX + areaWidth > globalX + globalWidth)
                    areaWidth = globalX - areaX + globalWidth;
                if (areaY + areaHeight > globalY + globalHeight)
                    areaHeight = globalY - areaY + globalHeight;
            }
        }

        if (extents)
            XFree(extents);
        if (desktop)
            XFree(desktop);
    }

    if (xpos)
        *xpos = areaX;
    if (ypos)
        *ypos = areaY;
    if (width)
        *width = areaWidth;
    if (height)
        *height = areaHeight;
}

GLFWvidmode* ____glfwGetVideoModesX11(_GLFWmonitor* monitor, int* count)
{
    GLFWvidmode* result;

    *count = 0;

    if (__glfw.x11.randr.available && !__glfw.x11.randr.monitorBroken)
    {
        XRRScreenResources* sr =
            XRRGetScreenResourcesCurrent(__glfw.x11.display, __glfw.x11.root);
        XRRCrtcInfo* ci = XRRGetCrtcInfo(__glfw.x11.display, sr, monitor->x11.crtc);
        XRROutputInfo* oi = XRRGetOutputInfo(__glfw.x11.display, sr, monitor->x11.output);

        result = __glfw_calloc(oi->nmode, sizeof(GLFWvidmode));

        for (int i = 0;  i < oi->nmode;  i++)
        {
            const XRRModeInfo* mi = getModeInfo(sr, oi->modes[i]);
            if (!modeIsGood(mi))
                continue;

            const GLFWvidmode mode = vidmodeFromModeInfo(mi, ci);
            int j;

            for (j = 0;  j < *count;  j++)
            {
                if (___glfwCompareVideoModes(result + j, &mode) == 0)
                    break;
            }

            // Skip duplicate modes
            if (j < *count)
                continue;

            (*count)++;
            result[*count - 1] = mode;
        }

        XRRFreeOutputInfo(oi);
        XRRFreeCrtcInfo(ci);
        XRRFreeScreenResources(sr);
    }
    else
    {
        *count = 1;
        result = __glfw_calloc(1, sizeof(GLFWvidmode));
        ___glfwGetVideoModeX11(monitor, result);
    }

    return result;
}

void ___glfwGetVideoModeX11(_GLFWmonitor* monitor, GLFWvidmode* mode)
{
    if (__glfw.x11.randr.available && !__glfw.x11.randr.monitorBroken)
    {
        XRRScreenResources* sr =
            XRRGetScreenResourcesCurrent(__glfw.x11.display, __glfw.x11.root);
        XRRCrtcInfo* ci = XRRGetCrtcInfo(__glfw.x11.display, sr, monitor->x11.crtc);

        if (ci)
        {
            const XRRModeInfo* mi = getModeInfo(sr, ci->mode);
            if (mi)  // mi can be NULL if the monitor has been disconnected
                *mode = vidmodeFromModeInfo(mi, ci);

            XRRFreeCrtcInfo(ci);
        }

        XRRFreeScreenResources(sr);
    }
    else
    {
        mode->width = DisplayWidth(__glfw.x11.display, __glfw.x11.screen);
        mode->height = DisplayHeight(__glfw.x11.display, __glfw.x11.screen);
        mode->refreshRate = 0;

        ___glfwSplitBPP(DefaultDepth(__glfw.x11.display, __glfw.x11.screen),
                      &mode->redBits, &mode->greenBits, &mode->blueBits);
    }
}

GLFWbool ___glfwGetGammaRampX11(_GLFWmonitor* monitor, GLFWgammaramp* ramp)
{
    if (__glfw.x11.randr.available && !__glfw.x11.randr.gammaBroken)
    {
        const size_t size = XRRGetCrtcGammaSize(__glfw.x11.display,
                                                monitor->x11.crtc);
        XRRCrtcGamma* gamma = XRRGetCrtcGamma(__glfw.x11.display,
                                              monitor->x11.crtc);

        ___glfwAllocGammaArrays(ramp, size);

        memcpy(ramp->red,   gamma->red,   size * sizeof(unsigned short));
        memcpy(ramp->green, gamma->green, size * sizeof(unsigned short));
        memcpy(ramp->blue,  gamma->blue,  size * sizeof(unsigned short));

        XRRFreeGamma(gamma);
        return GLFW_TRUE;
    }
    else if (__glfw.x11.vidmode.available)
    {
        int size;
        XF86VidModeGetGammaRampSize(__glfw.x11.display, __glfw.x11.screen, &size);

        ___glfwAllocGammaArrays(ramp, size);

        XF86VidModeGetGammaRamp(__glfw.x11.display,
                                __glfw.x11.screen,
                                ramp->size, ramp->red, ramp->green, ramp->blue);
        return GLFW_TRUE;
    }
    else
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "X11: Gamma ramp access not supported by server");
        return GLFW_FALSE;
    }
}

void ____glfwSetGammaRampX11(_GLFWmonitor* monitor, const GLFWgammaramp* ramp)
{
    if (__glfw.x11.randr.available && !__glfw.x11.randr.gammaBroken)
    {
        if (XRRGetCrtcGammaSize(__glfw.x11.display, monitor->x11.crtc) != ramp->size)
        {
            ___glfwInputError(GLFW_PLATFORM_ERROR,
                            "X11: Gamma ramp size must match current ramp size");
            return;
        }

        XRRCrtcGamma* gamma = XRRAllocGamma(ramp->size);

        memcpy(gamma->red,   ramp->red,   ramp->size * sizeof(unsigned short));
        memcpy(gamma->green, ramp->green, ramp->size * sizeof(unsigned short));
        memcpy(gamma->blue,  ramp->blue,  ramp->size * sizeof(unsigned short));

        XRRSetCrtcGamma(__glfw.x11.display, monitor->x11.crtc, gamma);
        XRRFreeGamma(gamma);
    }
    else if (__glfw.x11.vidmode.available)
    {
        XF86VidModeSetGammaRamp(__glfw.x11.display,
                                __glfw.x11.screen,
                                ramp->size,
                                (unsigned short*) ramp->red,
                                (unsigned short*) ramp->green,
                                (unsigned short*) ramp->blue);
    }
    else
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "X11: Gamma ramp access not supported by server");
    }
}


//////////////////////////////////////////////////////////////////////////
//////                        GLFW native API                       //////
//////////////////////////////////////////////////////////////////////////

GLFWAPI RRCrtc __glfwGetX11Adapter(GLFWmonitor* handle)
{
    _GLFWmonitor* monitor = (_GLFWmonitor*) handle;
    _GLFW_REQUIRE_INIT_OR_RETURN(None);
    return monitor->x11.crtc;
}

GLFWAPI RROutput __glfwGetX11Monitor(GLFWmonitor* handle)
{
    _GLFWmonitor* monitor = (_GLFWmonitor*) handle;
    _GLFW_REQUIRE_INIT_OR_RETURN(None);
    return monitor->x11.output;
}

