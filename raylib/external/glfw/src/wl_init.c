//========================================================================
// GLFW 3.4 Wayland - www.glfw.org
//------------------------------------------------------------------------
// Copyright (c) 2014 Jonas Ådahl <jadahl@gmail.com>
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

#include <errno.h>
#include <limits.h>
#include <linux/input.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mman.h>
#include <sys/timerfd.h>
#include <unistd.h>
#include <time.h>

#include "wayland-client-protocol.h"
#include "wayland-xdg-shell-client-protocol.h"
#include "wayland-xdg-decoration-client-protocol.h"
#include "wayland-viewporter-client-protocol.h"
#include "wayland-relative-pointer-unstable-v1-client-protocol.h"
#include "wayland-pointer-constraints-unstable-v1-client-protocol.h"
#include "wayland-idle-inhibit-unstable-v1-client-protocol.h"

// NOTE: Versions of wayland-scanner prior to 1.17.91 named every global array of
//       wl_interface pointers 'types', making it impossible to combine several unmodified
//       private-code files into a single compilation unit
// HACK: We override this name with a macro for each file, allowing them to coexist

#define types __glfw_wayland_types
#include "wayland-client-protocol-code.h"
#undef types

#define types __glfw_xdg_shell_types
#include "wayland-xdg-shell-client-protocol-code.h"
#undef types

#define types __glfw_xdg_decoration_types
#include "wayland-xdg-decoration-client-protocol-code.h"
#undef types

#define types __glfw_viewporter_types
#include "wayland-viewporter-client-protocol-code.h"
#undef types

#define types __glfw_relative_pointer_types
#include "wayland-relative-pointer-unstable-v1-client-protocol-code.h"
#undef types

#define types __glfw_pointer_constraints_types
#include "wayland-pointer-constraints-unstable-v1-client-protocol-code.h"
#undef types

#define types __glfw_idle_inhibit_types
#include "wayland-idle-inhibit-unstable-v1-client-protocol-code.h"
#undef types

static void wmBaseHandlePing(void* userData,
                             struct xdg_wm_base* wmBase,
                             uint32_t serial)
{
    xdg_wm_base_pong(wmBase, serial);
}

static const struct xdg_wm_base_listener wmBaseListener =
{
    wmBaseHandlePing
};

static void registryHandleGlobal(void* userData,
                                 struct wl_registry* registry,
                                 uint32_t name,
                                 const char* interface,
                                 uint32_t version)
{
    if (strcmp(interface, "wl_compositor") == 0)
    {
        __glfw.wl.compositorVersion = ___glfw_min(3, version);
        __glfw.wl.compositor =
            wl_registry_bind(registry, name, &wl_compositor_interface,
                             __glfw.wl.compositorVersion);
    }
    else if (strcmp(interface, "wl_subcompositor") == 0)
    {
        __glfw.wl.subcompositor =
            wl_registry_bind(registry, name, &wl_subcompositor_interface, 1);
    }
    else if (strcmp(interface, "wl_shm") == 0)
    {
        __glfw.wl.shm =
            wl_registry_bind(registry, name, &wl_shm_interface, 1);
    }
    else if (strcmp(interface, "wl_output") == 0)
    {
        __glfwAddOutputWayland(name, version);
    }
    else if (strcmp(interface, "wl_seat") == 0)
    {
        if (!__glfw.wl.seat)
        {
            __glfw.wl.seatVersion = ___glfw_min(4, version);
            __glfw.wl.seat =
                wl_registry_bind(registry, name, &wl_seat_interface,
                                 __glfw.wl.seatVersion);
            __glfwAddSeatListenerWayland(__glfw.wl.seat);
        }
    }
    else if (strcmp(interface, "wl_data_device_manager") == 0)
    {
        if (!__glfw.wl.dataDeviceManager)
        {
            __glfw.wl.dataDeviceManager =
                wl_registry_bind(registry, name,
                                 &wl_data_device_manager_interface, 1);
        }
    }
    else if (strcmp(interface, "xdg_wm_base") == 0)
    {
        __glfw.wl.wmBase =
            wl_registry_bind(registry, name, &xdg_wm_base_interface, 1);
        xdg_wm_base_add_listener(__glfw.wl.wmBase, &wmBaseListener, NULL);
    }
    else if (strcmp(interface, "zxdg_decoration_manager_v1") == 0)
    {
        __glfw.wl.decorationManager =
            wl_registry_bind(registry, name,
                             &zxdg_decoration_manager_v1_interface,
                             1);
    }
    else if (strcmp(interface, "wp_viewporter") == 0)
    {
        __glfw.wl.viewporter =
            wl_registry_bind(registry, name, &wp_viewporter_interface, 1);
    }
    else if (strcmp(interface, "zwp_relative_pointer_manager_v1") == 0)
    {
        __glfw.wl.relativePointerManager =
            wl_registry_bind(registry, name,
                             &zwp_relative_pointer_manager_v1_interface,
                             1);
    }
    else if (strcmp(interface, "zwp_pointer_constraints_v1") == 0)
    {
        __glfw.wl.pointerConstraints =
            wl_registry_bind(registry, name,
                             &zwp_pointer_constraints_v1_interface,
                             1);
    }
    else if (strcmp(interface, "zwp_idle_inhibit_manager_v1") == 0)
    {
        __glfw.wl.idleInhibitManager =
            wl_registry_bind(registry, name,
                             &zwp_idle_inhibit_manager_v1_interface,
                             1);
    }
}

static void registryHandleGlobalRemove(void* userData,
                                       struct wl_registry* registry,
                                       uint32_t name)
{
    for (int i = 0; i < __glfw.monitorCount; ++i)
    {
        _GLFWmonitor* monitor = __glfw.monitors[i];
        if (monitor->wl.name == name)
        {
            ___glfwInputMonitor(monitor, GLFW_DISCONNECTED, 0);
            return;
        }
    }
}


static const struct wl_registry_listener registryListener =
{
    registryHandleGlobal,
    registryHandleGlobalRemove
};

// Create key code translation tables
//
static void createKeyTables(void)
{
    memset(__glfw.wl.keycodes, -1, sizeof(__glfw.wl.keycodes));
    memset(__glfw.wl.scancodes, -1, sizeof(__glfw.wl.scancodes));

    __glfw.wl.keycodes[KEY_GRAVE]      = GLFW_KEY_GRAVE_ACCENT;
    __glfw.wl.keycodes[KEY_1]          = GLFW_KEY_1;
    __glfw.wl.keycodes[KEY_2]          = GLFW_KEY_2;
    __glfw.wl.keycodes[KEY_3]          = GLFW_KEY_3;
    __glfw.wl.keycodes[KEY_4]          = GLFW_KEY_4;
    __glfw.wl.keycodes[KEY_5]          = GLFW_KEY_5;
    __glfw.wl.keycodes[KEY_6]          = GLFW_KEY_6;
    __glfw.wl.keycodes[KEY_7]          = GLFW_KEY_7;
    __glfw.wl.keycodes[KEY_8]          = GLFW_KEY_8;
    __glfw.wl.keycodes[KEY_9]          = GLFW_KEY_9;
    __glfw.wl.keycodes[KEY_0]          = GLFW_KEY_0;
    __glfw.wl.keycodes[KEY_SPACE]      = GLFW_KEY_SPACE;
    __glfw.wl.keycodes[KEY_MINUS]      = GLFW_KEY_MINUS;
    __glfw.wl.keycodes[KEY_EQUAL]      = GLFW_KEY_EQUAL;
    __glfw.wl.keycodes[KEY_Q]          = GLFW_KEY_Q;
    __glfw.wl.keycodes[KEY_W]          = GLFW_KEY_W;
    __glfw.wl.keycodes[KEY_E]          = GLFW_KEY_E;
    __glfw.wl.keycodes[KEY_R]          = GLFW_KEY_R;
    __glfw.wl.keycodes[KEY_T]          = GLFW_KEY_T;
    __glfw.wl.keycodes[KEY_Y]          = GLFW_KEY_Y;
    __glfw.wl.keycodes[KEY_U]          = GLFW_KEY_U;
    __glfw.wl.keycodes[KEY_I]          = GLFW_KEY_I;
    __glfw.wl.keycodes[KEY_O]          = GLFW_KEY_O;
    __glfw.wl.keycodes[KEY_P]          = GLFW_KEY_P;
    __glfw.wl.keycodes[KEY_LEFTBRACE]  = GLFW_KEY_LEFT_BRACKET;
    __glfw.wl.keycodes[KEY_RIGHTBRACE] = GLFW_KEY_RIGHT_BRACKET;
    __glfw.wl.keycodes[KEY_A]          = GLFW_KEY_A;
    __glfw.wl.keycodes[KEY_S]          = GLFW_KEY_S;
    __glfw.wl.keycodes[KEY_D]          = GLFW_KEY_D;
    __glfw.wl.keycodes[KEY_F]          = GLFW_KEY_F;
    __glfw.wl.keycodes[KEY_G]          = GLFW_KEY_G;
    __glfw.wl.keycodes[KEY_H]          = GLFW_KEY_H;
    __glfw.wl.keycodes[KEY_J]          = GLFW_KEY_J;
    __glfw.wl.keycodes[KEY_K]          = GLFW_KEY_K;
    __glfw.wl.keycodes[KEY_L]          = GLFW_KEY_L;
    __glfw.wl.keycodes[KEY_SEMICOLON]  = GLFW_KEY_SEMICOLON;
    __glfw.wl.keycodes[KEY_APOSTROPHE] = GLFW_KEY_APOSTROPHE;
    __glfw.wl.keycodes[KEY_Z]          = GLFW_KEY_Z;
    __glfw.wl.keycodes[KEY_X]          = GLFW_KEY_X;
    __glfw.wl.keycodes[KEY_C]          = GLFW_KEY_C;
    __glfw.wl.keycodes[KEY_V]          = GLFW_KEY_V;
    __glfw.wl.keycodes[KEY_B]          = GLFW_KEY_B;
    __glfw.wl.keycodes[KEY_N]          = GLFW_KEY_N;
    __glfw.wl.keycodes[KEY_M]          = GLFW_KEY_M;
    __glfw.wl.keycodes[KEY_COMMA]      = GLFW_KEY_COMMA;
    __glfw.wl.keycodes[KEY_DOT]        = GLFW_KEY_PERIOD;
    __glfw.wl.keycodes[KEY_SLASH]      = GLFW_KEY_SLASH;
    __glfw.wl.keycodes[KEY_BACKSLASH]  = GLFW_KEY_BACKSLASH;
    __glfw.wl.keycodes[KEY_ESC]        = GLFW_KEY_ESCAPE;
    __glfw.wl.keycodes[KEY_TAB]        = GLFW_KEY_TAB;
    __glfw.wl.keycodes[KEY_LEFTSHIFT]  = GLFW_KEY_LEFT_SHIFT;
    __glfw.wl.keycodes[KEY_RIGHTSHIFT] = GLFW_KEY_RIGHT_SHIFT;
    __glfw.wl.keycodes[KEY_LEFTCTRL]   = GLFW_KEY_LEFT_CONTROL;
    __glfw.wl.keycodes[KEY_RIGHTCTRL]  = GLFW_KEY_RIGHT_CONTROL;
    __glfw.wl.keycodes[KEY_LEFTALT]    = GLFW_KEY_LEFT_ALT;
    __glfw.wl.keycodes[KEY_RIGHTALT]   = GLFW_KEY_RIGHT_ALT;
    __glfw.wl.keycodes[KEY_LEFTMETA]   = GLFW_KEY_LEFT_SUPER;
    __glfw.wl.keycodes[KEY_RIGHTMETA]  = GLFW_KEY_RIGHT_SUPER;
    __glfw.wl.keycodes[KEY_COMPOSE]    = GLFW_KEY_MENU;
    __glfw.wl.keycodes[KEY_NUMLOCK]    = GLFW_KEY_NUM_LOCK;
    __glfw.wl.keycodes[KEY_CAPSLOCK]   = GLFW_KEY_CAPS_LOCK;
    __glfw.wl.keycodes[KEY_PRINT]      = GLFW_KEY_PRINT_SCREEN;
    __glfw.wl.keycodes[KEY_SCROLLLOCK] = GLFW_KEY_SCROLL_LOCK;
    __glfw.wl.keycodes[KEY_PAUSE]      = GLFW_KEY_PAUSE;
    __glfw.wl.keycodes[KEY_DELETE]     = GLFW_KEY_DELETE;
    __glfw.wl.keycodes[KEY_BACKSPACE]  = GLFW_KEY_BACKSPACE;
    __glfw.wl.keycodes[KEY_ENTER]      = GLFW_KEY_ENTER;
    __glfw.wl.keycodes[KEY_HOME]       = GLFW_KEY_HOME;
    __glfw.wl.keycodes[KEY_END]        = GLFW_KEY_END;
    __glfw.wl.keycodes[KEY_PAGEUP]     = GLFW_KEY_PAGE_UP;
    __glfw.wl.keycodes[KEY_PAGEDOWN]   = GLFW_KEY_PAGE_DOWN;
    __glfw.wl.keycodes[KEY_INSERT]     = GLFW_KEY_INSERT;
    __glfw.wl.keycodes[KEY_LEFT]       = GLFW_KEY_LEFT;
    __glfw.wl.keycodes[KEY_RIGHT]      = GLFW_KEY_RIGHT;
    __glfw.wl.keycodes[KEY_DOWN]       = GLFW_KEY_DOWN;
    __glfw.wl.keycodes[KEY_UP]         = GLFW_KEY_UP;
    __glfw.wl.keycodes[KEY_F1]         = GLFW_KEY_F1;
    __glfw.wl.keycodes[KEY_F2]         = GLFW_KEY_F2;
    __glfw.wl.keycodes[KEY_F3]         = GLFW_KEY_F3;
    __glfw.wl.keycodes[KEY_F4]         = GLFW_KEY_F4;
    __glfw.wl.keycodes[KEY_F5]         = GLFW_KEY_F5;
    __glfw.wl.keycodes[KEY_F6]         = GLFW_KEY_F6;
    __glfw.wl.keycodes[KEY_F7]         = GLFW_KEY_F7;
    __glfw.wl.keycodes[KEY_F8]         = GLFW_KEY_F8;
    __glfw.wl.keycodes[KEY_F9]         = GLFW_KEY_F9;
    __glfw.wl.keycodes[KEY_F10]        = GLFW_KEY_F10;
    __glfw.wl.keycodes[KEY_F11]        = GLFW_KEY_F11;
    __glfw.wl.keycodes[KEY_F12]        = GLFW_KEY_F12;
    __glfw.wl.keycodes[KEY_F13]        = GLFW_KEY_F13;
    __glfw.wl.keycodes[KEY_F14]        = GLFW_KEY_F14;
    __glfw.wl.keycodes[KEY_F15]        = GLFW_KEY_F15;
    __glfw.wl.keycodes[KEY_F16]        = GLFW_KEY_F16;
    __glfw.wl.keycodes[KEY_F17]        = GLFW_KEY_F17;
    __glfw.wl.keycodes[KEY_F18]        = GLFW_KEY_F18;
    __glfw.wl.keycodes[KEY_F19]        = GLFW_KEY_F19;
    __glfw.wl.keycodes[KEY_F20]        = GLFW_KEY_F20;
    __glfw.wl.keycodes[KEY_F21]        = GLFW_KEY_F21;
    __glfw.wl.keycodes[KEY_F22]        = GLFW_KEY_F22;
    __glfw.wl.keycodes[KEY_F23]        = GLFW_KEY_F23;
    __glfw.wl.keycodes[KEY_F24]        = GLFW_KEY_F24;
    __glfw.wl.keycodes[KEY_KPSLASH]    = GLFW_KEY_KP_DIVIDE;
    __glfw.wl.keycodes[KEY_KPASTERISK] = GLFW_KEY_KP_MULTIPLY;
    __glfw.wl.keycodes[KEY_KPMINUS]    = GLFW_KEY_KP_SUBTRACT;
    __glfw.wl.keycodes[KEY_KPPLUS]     = GLFW_KEY_KP_ADD;
    __glfw.wl.keycodes[KEY_KP0]        = GLFW_KEY_KP_0;
    __glfw.wl.keycodes[KEY_KP1]        = GLFW_KEY_KP_1;
    __glfw.wl.keycodes[KEY_KP2]        = GLFW_KEY_KP_2;
    __glfw.wl.keycodes[KEY_KP3]        = GLFW_KEY_KP_3;
    __glfw.wl.keycodes[KEY_KP4]        = GLFW_KEY_KP_4;
    __glfw.wl.keycodes[KEY_KP5]        = GLFW_KEY_KP_5;
    __glfw.wl.keycodes[KEY_KP6]        = GLFW_KEY_KP_6;
    __glfw.wl.keycodes[KEY_KP7]        = GLFW_KEY_KP_7;
    __glfw.wl.keycodes[KEY_KP8]        = GLFW_KEY_KP_8;
    __glfw.wl.keycodes[KEY_KP9]        = GLFW_KEY_KP_9;
    __glfw.wl.keycodes[KEY_KPDOT]      = GLFW_KEY_KP_DECIMAL;
    __glfw.wl.keycodes[KEY_KPEQUAL]    = GLFW_KEY_KP_EQUAL;
    __glfw.wl.keycodes[KEY_KPENTER]    = GLFW_KEY_KP_ENTER;
    __glfw.wl.keycodes[KEY_102ND]      = GLFW_KEY_WORLD_2;

    for (int scancode = 0;  scancode < 256;  scancode++)
    {
        if (__glfw.wl.keycodes[scancode] > 0)
            __glfw.wl.scancodes[__glfw.wl.keycodes[scancode]] = scancode;
    }
}

static GLFWbool loadCursorTheme(void)
{
    int cursorSize = 32;

    const char* sizeString = getenv("XCURSOR_SIZE");
    if (sizeString)
    {
        errno = 0;
        const long cursorSizeLong = strtol(sizeString, NULL, 10);
        if (errno == 0 && cursorSizeLong > 0 && cursorSizeLong < INT_MAX)
            cursorSize = (int) cursorSizeLong;
    }

    const char* themeName = getenv("XCURSOR_THEME");

    __glfw.wl.cursorTheme = wl_cursor_theme_load(themeName, cursorSize, __glfw.wl.shm);
    if (!__glfw.wl.cursorTheme)
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "Wayland: Failed to load default cursor theme");
        return GLFW_FALSE;
    }

    // If this happens to be NULL, we just fallback to the scale=1 version.
    __glfw.wl.cursorThemeHiDPI =
        wl_cursor_theme_load(themeName, cursorSize * 2, __glfw.wl.shm);

    __glfw.wl.cursorSurface = wl_compositor_create_surface(__glfw.wl.compositor);
    __glfw.wl.cursorTimerfd = timerfd_create(CLOCK_MONOTONIC, TFD_CLOEXEC | TFD_NONBLOCK);
    return GLFW_TRUE;
}


//////////////////////////////////////////////////////////////////////////
//////                       GLFW platform API                      //////
//////////////////////////////////////////////////////////////////////////

GLFWbool __glfwConnectWayland(int platformID, _GLFWplatform* platform)
{
    const _GLFWplatform wayland =
    {
        GLFW_PLATFORM_WAYLAND,
        ___glfwInitWayland,
        ___glfwTerminateWayland,
        ___glfwGetCursorPosWayland,
        ____glfwSetCursorPosWayland,
        ___glfwSetCursorModeWayland,
        __glfwSetRawMouseMotionWayland,
        ___glfwRawMouseMotionSupportedWayland,
        ___glfwCreateCursorWayland,
        ___glfwCreateStandardCursorWayland,
        ___glfwDestroyCursorWayland,
        ___glfwSetCursorWayland,
        __glfwGetScancodeNameWayland,
        ____glfwGetKeyScancodeWayland,
        ___glfwSetClipboardStringWayland,
        ___glfwGetClipboardStringWayland,
#if defined(__linux__)
        ____glfwInitJoysticksLinux,
        ____glfwTerminateJoysticksLinux,
        __glfwPollJoystickLinux,
        __glfwGetMappingNameLinux,
        __glfwUpdateGamepadGUIDLinux,
#else
        ___glfwInitJoysticksNull,
        ___glfwTerminateJoysticksNull,
        __glfwPollJoystickNull,
        __glfwGetMappingNameNull,
        __glfwUpdateGamepadGUIDNull,
#endif
        ___glfwFreeMonitorWayland,
        ___glfwGetMonitorPosWayland,
        ___glfwGetMonitorContentScaleWayland,
        ___glfwGetMonitorWorkareaWayland,
        ____glfwGetVideoModesWayland,
        ___glfwGetVideoModeWayland,
        ___glfwGetGammaRampWayland,
        ____glfwSetGammaRampWayland,
        ___glfwCreateWindowWayland,
        ___glfwDestroyWindowWayland,
        ___glfwSetWindowTitleWayland,
        ___glfwSetWindowIconWayland,
        ___glfwGetWindowPosWayland,
        ___glfwSetWindowPosWayland,
        ___glfwGetWindowSizeWayland,
        ___glfwSetWindowSizeWayland,
        ____glfwSetWindowSizeLimitsWayland,
        ___glfwSetWindowAspectRatioWayland,
        ___glfwGetFramebufferSizeWayland,
        ___glfwGetWindowFrameSizeWayland,
        ___glfwGetWindowContentScaleWayland,
        ___glfwIconifyWindowWayland,
        ___glfwRestoreWindowWayland,
        ___glfwMaximizeWindowWayland,
        ___glfwShowWindowWayland,
        ___glfwHideWindowWayland,
        ___glfwRequestWindowAttentionWayland,
        ___glfwFocusWindowWayland,
        ___glfwSetWindowMonitorWayland,
        __glfwWindowFocusedWayland,
        __glfwWindowIconifiedWayland,
        __glfwWindowVisibleWayland,
        __glfwWindowMaximizedWayland,
        __glfwWindowHoveredWayland,
        __glfwFramebufferTransparentWayland,
        ___glfwGetWindowOpacityWayland,
        __glfwSetWindowResizableWayland,
        __glfwSetWindowDecoratedWayland,
        __glfwSetWindowFloatingWayland,
        ___glfwSetWindowOpacityWayland,
        __glfwSetWindowMousePassthroughWayland,
        ___glfwPollEventsWayland,
        ___glfwWaitEventsWayland,
        ____glfwWaitEventsTimeoutWayland,
        ___glfwPostEmptyEventWayland,
        __glfwGetEGLPlatformWayland,
        __glfwGetEGLNativeDisplayWayland,
        __glfwGetEGLNativeWindowWayland,
        ___glfwGetRequiredInstanceExtensionsWayland,
        ___glfwGetPhysicalDevicePresentationSupportWayland,
        ____glfwCreateWindowSurfaceWayland,
    };

    void* module = __glfwPlatformLoadModule("libwayland-client.so.0");
    if (!module)
    {
        if (platformID == GLFW_PLATFORM_WAYLAND)
        {
            ___glfwInputError(GLFW_PLATFORM_ERROR,
                            "Wayland: Failed to load libwayland-client");
        }

        return GLFW_FALSE;
    }

    PFN_wl_display_connect wl_display_connect = (PFN_wl_display_connect)
        __glfwPlatformGetModuleSymbol(module, "wl_display_connect");
    if (!wl_display_connect)
    {
        if (platformID == GLFW_PLATFORM_WAYLAND)
        {
            ___glfwInputError(GLFW_PLATFORM_ERROR,
                            "Wayland: Failed to load libwayland-client entry point");
        }

        __glfwPlatformFreeModule(module);
        return GLFW_FALSE;
    }

    struct wl_display* display = wl_display_connect(NULL);
    if (!display)
    {
        if (platformID == GLFW_PLATFORM_WAYLAND)
            ___glfwInputError(GLFW_PLATFORM_ERROR, "Wayland: Failed to connect to display");

        __glfwPlatformFreeModule(module);
        return GLFW_FALSE;
    }

    __glfw.wl.display = display;
    __glfw.wl.client.handle = module;

    *platform = wayland;
    return GLFW_TRUE;
}

int ___glfwInitWayland(void)
{
    // These must be set before any failure checks
    __glfw.wl.keyRepeatTimerfd = -1;
    __glfw.wl.cursorTimerfd = -1;

    __glfw.wl.client.display_flush = (PFN_wl_display_flush)
        __glfwPlatformGetModuleSymbol(__glfw.wl.client.handle, "wl_display_flush");
    __glfw.wl.client.display_cancel_read = (PFN_wl_display_cancel_read)
        __glfwPlatformGetModuleSymbol(__glfw.wl.client.handle, "wl_display_cancel_read");
    __glfw.wl.client.display_dispatch_pending = (PFN_wl_display_dispatch_pending)
        __glfwPlatformGetModuleSymbol(__glfw.wl.client.handle, "wl_display_dispatch_pending");
    __glfw.wl.client.display_read_events = (PFN_wl_display_read_events)
        __glfwPlatformGetModuleSymbol(__glfw.wl.client.handle, "wl_display_read_events");
    __glfw.wl.client.display_disconnect = (PFN_wl_display_disconnect)
        __glfwPlatformGetModuleSymbol(__glfw.wl.client.handle, "wl_display_disconnect");
    __glfw.wl.client.display_roundtrip = (PFN_wl_display_roundtrip)
        __glfwPlatformGetModuleSymbol(__glfw.wl.client.handle, "wl_display_roundtrip");
    __glfw.wl.client.display_get_fd = (PFN_wl_display_get_fd)
        __glfwPlatformGetModuleSymbol(__glfw.wl.client.handle, "wl_display_get_fd");
    __glfw.wl.client.display_prepare_read = (PFN_wl_display_prepare_read)
        __glfwPlatformGetModuleSymbol(__glfw.wl.client.handle, "wl_display_prepare_read");
    __glfw.wl.client.proxy_marshal = (PFN_wl_proxy_marshal)
        __glfwPlatformGetModuleSymbol(__glfw.wl.client.handle, "wl_proxy_marshal");
    __glfw.wl.client.proxy_add_listener = (PFN_wl_proxy_add_listener)
        __glfwPlatformGetModuleSymbol(__glfw.wl.client.handle, "wl_proxy_add_listener");
    __glfw.wl.client.proxy_destroy = (PFN_wl_proxy_destroy)
        __glfwPlatformGetModuleSymbol(__glfw.wl.client.handle, "wl_proxy_destroy");
    __glfw.wl.client.proxy_marshal_constructor = (PFN_wl_proxy_marshal_constructor)
        __glfwPlatformGetModuleSymbol(__glfw.wl.client.handle, "wl_proxy_marshal_constructor");
    __glfw.wl.client.proxy_marshal_constructor_versioned = (PFN_wl_proxy_marshal_constructor_versioned)
        __glfwPlatformGetModuleSymbol(__glfw.wl.client.handle, "wl_proxy_marshal_constructor_versioned");
    __glfw.wl.client.proxy_get_user_data = (PFN_wl_proxy_get_user_data)
        __glfwPlatformGetModuleSymbol(__glfw.wl.client.handle, "wl_proxy_get_user_data");
    __glfw.wl.client.proxy_set_user_data = (PFN_wl_proxy_set_user_data)
        __glfwPlatformGetModuleSymbol(__glfw.wl.client.handle, "wl_proxy_set_user_data");
    __glfw.wl.client.proxy_get_version = (PFN_wl_proxy_get_version)
        __glfwPlatformGetModuleSymbol(__glfw.wl.client.handle, "wl_proxy_get_version");
    __glfw.wl.client.proxy_marshal_flags = (PFN_wl_proxy_marshal_flags)
        __glfwPlatformGetModuleSymbol(__glfw.wl.client.handle, "wl_proxy_marshal_flags");

    if (!__glfw.wl.client.display_flush ||
        !__glfw.wl.client.display_cancel_read ||
        !__glfw.wl.client.display_dispatch_pending ||
        !__glfw.wl.client.display_read_events ||
        !__glfw.wl.client.display_disconnect ||
        !__glfw.wl.client.display_roundtrip ||
        !__glfw.wl.client.display_get_fd ||
        !__glfw.wl.client.display_prepare_read ||
        !__glfw.wl.client.proxy_marshal ||
        !__glfw.wl.client.proxy_add_listener ||
        !__glfw.wl.client.proxy_destroy ||
        !__glfw.wl.client.proxy_marshal_constructor ||
        !__glfw.wl.client.proxy_marshal_constructor_versioned ||
        !__glfw.wl.client.proxy_get_user_data ||
        !__glfw.wl.client.proxy_set_user_data)
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "Wayland: Failed to load libwayland-client entry point");
        return GLFW_FALSE;
    }

    __glfw.wl.cursor.handle = __glfwPlatformLoadModule("libwayland-cursor.so.0");
    if (!__glfw.wl.cursor.handle)
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "Wayland: Failed to load libwayland-cursor");
        return GLFW_FALSE;
    }

    __glfw.wl.cursor.theme_load = (PFN_wl_cursor_theme_load)
        __glfwPlatformGetModuleSymbol(__glfw.wl.cursor.handle, "wl_cursor_theme_load");
    __glfw.wl.cursor.theme_destroy = (PFN_wl_cursor_theme_destroy)
        __glfwPlatformGetModuleSymbol(__glfw.wl.cursor.handle, "wl_cursor_theme_destroy");
    __glfw.wl.cursor.theme_get_cursor = (PFN_wl_cursor_theme_get_cursor)
        __glfwPlatformGetModuleSymbol(__glfw.wl.cursor.handle, "wl_cursor_theme_get_cursor");
    __glfw.wl.cursor.image_get_buffer = (PFN_wl_cursor_image_get_buffer)
        __glfwPlatformGetModuleSymbol(__glfw.wl.cursor.handle, "wl_cursor_image_get_buffer");

    __glfw.wl.egl.handle = __glfwPlatformLoadModule("libwayland-egl.so.1");
    if (!__glfw.wl.egl.handle)
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "Wayland: Failed to load libwayland-egl");
        return GLFW_FALSE;
    }

    __glfw.wl.egl.window_create = (PFN_wl_egl_window_create)
        __glfwPlatformGetModuleSymbol(__glfw.wl.egl.handle, "wl_egl_window_create");
    __glfw.wl.egl.window_destroy = (PFN_wl_egl_window_destroy)
        __glfwPlatformGetModuleSymbol(__glfw.wl.egl.handle, "wl_egl_window_destroy");
    __glfw.wl.egl.window_resize = (PFN_wl_egl_window_resize)
        __glfwPlatformGetModuleSymbol(__glfw.wl.egl.handle, "wl_egl_window_resize");

    __glfw.wl.xkb.handle = __glfwPlatformLoadModule("libxkbcommon.so.0");
    if (!__glfw.wl.xkb.handle)
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "Wayland: Failed to load libxkbcommon");
        return GLFW_FALSE;
    }

    __glfw.wl.xkb.context_new = (PFN_xkb_context_new)
        __glfwPlatformGetModuleSymbol(__glfw.wl.xkb.handle, "xkb_context_new");
    __glfw.wl.xkb.context_unref = (PFN_xkb_context_unref)
        __glfwPlatformGetModuleSymbol(__glfw.wl.xkb.handle, "xkb_context_unref");
    __glfw.wl.xkb.keymap_new_from_string = (PFN_xkb_keymap_new_from_string)
        __glfwPlatformGetModuleSymbol(__glfw.wl.xkb.handle, "xkb_keymap_new_from_string");
    __glfw.wl.xkb.keymap_unref = (PFN_xkb_keymap_unref)
        __glfwPlatformGetModuleSymbol(__glfw.wl.xkb.handle, "xkb_keymap_unref");
    __glfw.wl.xkb.keymap_mod_get_index = (PFN_xkb_keymap_mod_get_index)
        __glfwPlatformGetModuleSymbol(__glfw.wl.xkb.handle, "xkb_keymap_mod_get_index");
    __glfw.wl.xkb.keymap_key_repeats = (PFN_xkb_keymap_key_repeats)
        __glfwPlatformGetModuleSymbol(__glfw.wl.xkb.handle, "xkb_keymap_key_repeats");
    __glfw.wl.xkb.keymap_key_get_syms_by_level = (PFN_xkb_keymap_key_get_syms_by_level)
        __glfwPlatformGetModuleSymbol(__glfw.wl.xkb.handle, "xkb_keymap_key_get_syms_by_level");
    __glfw.wl.xkb.state_new = (PFN_xkb_state_new)
        __glfwPlatformGetModuleSymbol(__glfw.wl.xkb.handle, "xkb_state_new");
    __glfw.wl.xkb.state_unref = (PFN_xkb_state_unref)
        __glfwPlatformGetModuleSymbol(__glfw.wl.xkb.handle, "xkb_state_unref");
    __glfw.wl.xkb.state_key_get_syms = (PFN_xkb_state_key_get_syms)
        __glfwPlatformGetModuleSymbol(__glfw.wl.xkb.handle, "xkb_state_key_get_syms");
    __glfw.wl.xkb.state_update_mask = (PFN_xkb_state_update_mask)
        __glfwPlatformGetModuleSymbol(__glfw.wl.xkb.handle, "xkb_state_update_mask");
    __glfw.wl.xkb.state_key_get_layout = (PFN_xkb_state_key_get_layout)
        __glfwPlatformGetModuleSymbol(__glfw.wl.xkb.handle, "xkb_state_key_get_layout");
    __glfw.wl.xkb.state_mod_index_is_active = (PFN_xkb_state_mod_index_is_active)
        __glfwPlatformGetModuleSymbol(__glfw.wl.xkb.handle, "xkb_state_mod_index_is_active");
    __glfw.wl.xkb.compose_table_new_from_locale = (PFN_xkb_compose_table_new_from_locale)
        __glfwPlatformGetModuleSymbol(__glfw.wl.xkb.handle, "xkb_compose_table_new_from_locale");
    __glfw.wl.xkb.compose_table_unref = (PFN_xkb_compose_table_unref)
        __glfwPlatformGetModuleSymbol(__glfw.wl.xkb.handle, "xkb_compose_table_unref");
    __glfw.wl.xkb.compose_state_new = (PFN_xkb_compose_state_new)
        __glfwPlatformGetModuleSymbol(__glfw.wl.xkb.handle, "xkb_compose_state_new");
    __glfw.wl.xkb.compose_state_unref = (PFN_xkb_compose_state_unref)
        __glfwPlatformGetModuleSymbol(__glfw.wl.xkb.handle, "xkb_compose_state_unref");
    __glfw.wl.xkb.compose_state_feed = (PFN_xkb_compose_state_feed)
        __glfwPlatformGetModuleSymbol(__glfw.wl.xkb.handle, "xkb_compose_state_feed");
    __glfw.wl.xkb.compose_state_get_status = (PFN_xkb_compose_state_get_status)
        __glfwPlatformGetModuleSymbol(__glfw.wl.xkb.handle, "xkb_compose_state_get_status");
    __glfw.wl.xkb.compose_state_get_one_sym = (PFN_xkb_compose_state_get_one_sym)
        __glfwPlatformGetModuleSymbol(__glfw.wl.xkb.handle, "xkb_compose_state_get_one_sym");

    __glfw.wl.registry = wl_display_get_registry(__glfw.wl.display);
    wl_registry_add_listener(__glfw.wl.registry, &registryListener, NULL);

    createKeyTables();

    __glfw.wl.xkb.context = xkb_context_new(0);
    if (!__glfw.wl.xkb.context)
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "Wayland: Failed to initialize xkb context");
        return GLFW_FALSE;
    }

    // Sync so we got all registry objects
    wl_display_roundtrip(__glfw.wl.display);

    // Sync so we got all initial output events
    wl_display_roundtrip(__glfw.wl.display);

#ifdef WL_KEYBOARD_REPEAT_INFO_SINCE_VERSION
    if (__glfw.wl.seatVersion >= WL_KEYBOARD_REPEAT_INFO_SINCE_VERSION)
    {
        __glfw.wl.keyRepeatTimerfd =
            timerfd_create(CLOCK_MONOTONIC, TFD_CLOEXEC | TFD_NONBLOCK);
    }
#endif

    if (!__glfw.wl.wmBase)
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "Wayland: Failed to find xdg-shell in your compositor");
        return GLFW_FALSE;
    }

    if (!__glfw.wl.shm)
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "Wayland: Failed to find wl_shm in your compositor");
        return GLFW_FALSE;
    }

    if (!loadCursorTheme())
        return GLFW_FALSE;

    if (__glfw.wl.seat && __glfw.wl.dataDeviceManager)
    {
        __glfw.wl.dataDevice =
            wl_data_device_manager_get_data_device(__glfw.wl.dataDeviceManager,
                                                   __glfw.wl.seat);
        __glfwAddDataDeviceListenerWayland(__glfw.wl.dataDevice);
    }

    return GLFW_TRUE;
}

void ___glfwTerminateWayland(void)
{
    ____glfwTerminateEGL();
    ____glfwTerminateOSMesa();

    if (__glfw.wl.egl.handle)
    {
        __glfwPlatformFreeModule(__glfw.wl.egl.handle);
        __glfw.wl.egl.handle = NULL;
    }

    if (__glfw.wl.xkb.composeState)
        xkb_compose_state_unref(__glfw.wl.xkb.composeState);
    if (__glfw.wl.xkb.keymap)
        xkb_keymap_unref(__glfw.wl.xkb.keymap);
    if (__glfw.wl.xkb.state)
        xkb_state_unref(__glfw.wl.xkb.state);
    if (__glfw.wl.xkb.context)
        xkb_context_unref(__glfw.wl.xkb.context);
    if (__glfw.wl.xkb.handle)
    {
        __glfwPlatformFreeModule(__glfw.wl.xkb.handle);
        __glfw.wl.xkb.handle = NULL;
    }

    if (__glfw.wl.cursorTheme)
        wl_cursor_theme_destroy(__glfw.wl.cursorTheme);
    if (__glfw.wl.cursorThemeHiDPI)
        wl_cursor_theme_destroy(__glfw.wl.cursorThemeHiDPI);
    if (__glfw.wl.cursor.handle)
    {
        __glfwPlatformFreeModule(__glfw.wl.cursor.handle);
        __glfw.wl.cursor.handle = NULL;
    }

    for (unsigned int i = 0; i < __glfw.wl.offerCount; i++)
        wl_data_offer_destroy(__glfw.wl.offers[i].offer);

    __glfw_free(__glfw.wl.offers);

    if (__glfw.wl.cursorSurface)
        wl_surface_destroy(__glfw.wl.cursorSurface);
    if (__glfw.wl.subcompositor)
        wl_subcompositor_destroy(__glfw.wl.subcompositor);
    if (__glfw.wl.compositor)
        wl_compositor_destroy(__glfw.wl.compositor);
    if (__glfw.wl.shm)
        wl_shm_destroy(__glfw.wl.shm);
    if (__glfw.wl.viewporter)
        wp_viewporter_destroy(__glfw.wl.viewporter);
    if (__glfw.wl.decorationManager)
        zxdg_decoration_manager_v1_destroy(__glfw.wl.decorationManager);
    if (__glfw.wl.wmBase)
        xdg_wm_base_destroy(__glfw.wl.wmBase);
    if (__glfw.wl.selectionOffer)
        wl_data_offer_destroy(__glfw.wl.selectionOffer);
    if (__glfw.wl.dragOffer)
        wl_data_offer_destroy(__glfw.wl.dragOffer);
    if (__glfw.wl.selectionSource)
        wl_data_source_destroy(__glfw.wl.selectionSource);
    if (__glfw.wl.dataDevice)
        wl_data_device_destroy(__glfw.wl.dataDevice);
    if (__glfw.wl.dataDeviceManager)
        wl_data_device_manager_destroy(__glfw.wl.dataDeviceManager);
    if (__glfw.wl.pointer)
        wl_pointer_destroy(__glfw.wl.pointer);
    if (__glfw.wl.keyboard)
        wl_keyboard_destroy(__glfw.wl.keyboard);
    if (__glfw.wl.seat)
        wl_seat_destroy(__glfw.wl.seat);
    if (__glfw.wl.relativePointerManager)
        zwp_relative_pointer_manager_v1_destroy(__glfw.wl.relativePointerManager);
    if (__glfw.wl.pointerConstraints)
        zwp_pointer_constraints_v1_destroy(__glfw.wl.pointerConstraints);
    if (__glfw.wl.idleInhibitManager)
        zwp_idle_inhibit_manager_v1_destroy(__glfw.wl.idleInhibitManager);
    if (__glfw.wl.registry)
        wl_registry_destroy(__glfw.wl.registry);
    if (__glfw.wl.display)
    {
        wl_display_flush(__glfw.wl.display);
        wl_display_disconnect(__glfw.wl.display);
    }

    if (__glfw.wl.keyRepeatTimerfd >= 0)
        close(__glfw.wl.keyRepeatTimerfd);
    if (__glfw.wl.cursorTimerfd >= 0)
        close(__glfw.wl.cursorTimerfd);

    __glfw_free(__glfw.wl.clipboardString);
}

