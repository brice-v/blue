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

#include <X11/cursorfont.h>
#include <X11/Xmd.h>

#include <poll.h>

#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <limits.h>
#include <errno.h>
#include <assert.h>

// Action for EWMH client messages
#define _NET_WM_STATE_REMOVE        0
#define _NET_WM_STATE_ADD           1
#define _NET_WM_STATE_TOGGLE        2

// Additional mouse button names for XButtonEvent
#define Button6            6
#define Button7            7

// Motif WM hints flags
#define MWM_HINTS_DECORATIONS   2
#define MWM_DECOR_ALL           1

#define _GLFW_XDND_VERSION 5

// Wait for event data to arrive on the X11 display socket
// This avoids blocking other threads via the per-display Xlib lock that also
// covers GLX functions
//
static GLFWbool waitForX11Event(double* timeout)
{
    struct pollfd fd = { ConnectionNumber(__glfw.x11.display), POLLIN };

    while (!XPending(__glfw.x11.display))
    {
        if (!__glfwPollPOSIX(&fd, 1, timeout))
            return GLFW_FALSE;
    }

    return GLFW_TRUE;
}

// Wait for event data to arrive on any event file descriptor
// This avoids blocking other threads via the per-display Xlib lock that also
// covers GLX functions
//
static GLFWbool waitForAnyEvent(double* timeout)
{
    nfds_t count = 2;
    struct pollfd fds[3] =
    {
        { ConnectionNumber(__glfw.x11.display), POLLIN },
        { __glfw.x11.emptyEventPipe[0], POLLIN }
    };

#if defined(__linux__)
    if (__glfw.joysticksInitialized)
        fds[count++] = (struct pollfd) { __glfw.linjs.inotify, POLLIN };
#endif

    while (!XPending(__glfw.x11.display))
    {
        if (!__glfwPollPOSIX(fds, count, timeout))
            return GLFW_FALSE;

        for (int i = 1; i < count; i++)
        {
            if (fds[i].revents & POLLIN)
                return GLFW_TRUE;
        }
    }

    return GLFW_TRUE;
}

// Writes a byte to the empty event pipe
//
static void writeEmptyEvent(void)
{
    for (;;)
    {
        const char byte = 0;
        const ssize_t result = write(__glfw.x11.emptyEventPipe[1], &byte, 1);
        if (result == 1 || (result == -1 && errno != EINTR))
            break;
    }
}

// Drains available data from the empty event pipe
//
static void drainEmptyEvents(void)
{
    for (;;)
    {
        char dummy[64];
        const ssize_t result = read(__glfw.x11.emptyEventPipe[0], dummy, sizeof(dummy));
        if (result == -1 && errno != EINTR)
            break;
    }
}

// Waits until a VisibilityNotify event arrives for the specified window or the
// timeout period elapses (ICCCM section 4.2.2)
//
static GLFWbool waitForVisibilityNotify(_GLFWwindow* window)
{
    XEvent dummy;
    double timeout = 0.1;

    while (!XCheckTypedWindowEvent(__glfw.x11.display,
                                   window->x11.handle,
                                   VisibilityNotify,
                                   &dummy))
    {
        if (!waitForX11Event(&timeout))
            return GLFW_FALSE;
    }

    return GLFW_TRUE;
}

// Returns whether the window is iconified
//
static int getWindowState(_GLFWwindow* window)
{
    int result = WithdrawnState;
    struct {
        CARD32 state;
        Window icon;
    } *state = NULL;

    if (___glfwGetWindowPropertyX11(window->x11.handle,
                                  __glfw.x11.WM_STATE,
                                  __glfw.x11.WM_STATE,
                                  (unsigned char**) &state) >= 2)
    {
        result = state->state;
    }

    if (state)
        XFree(state);

    return result;
}

// Returns whether the event is a selection event
//
static Bool isSelectionEvent(Display* display, XEvent* event, XPointer pointer)
{
    if (event->xany.window != __glfw.x11.helperWindowHandle)
        return False;

    return event->type == SelectionRequest ||
           event->type == SelectionNotify ||
           event->type == SelectionClear;
}

// Returns whether it is a _NET_FRAME_EXTENTS event for the specified window
//
static Bool isFrameExtentsEvent(Display* display, XEvent* event, XPointer pointer)
{
    _GLFWwindow* window = (_GLFWwindow*) pointer;
    return event->type == PropertyNotify &&
           event->xproperty.state == PropertyNewValue &&
           event->xproperty.window == window->x11.handle &&
           event->xproperty.atom == __glfw.x11.NET_FRAME_EXTENTS;
}

// Returns whether it is a property event for the specified selection transfer
//
static Bool isSelPropNewValueNotify(Display* display, XEvent* event, XPointer pointer)
{
    XEvent* notification = (XEvent*) pointer;
    return event->type == PropertyNotify &&
           event->xproperty.state == PropertyNewValue &&
           event->xproperty.window == notification->xselection.requestor &&
           event->xproperty.atom == notification->xselection.property;
}

// Translates an X event modifier state mask
//
static int translateState(int state)
{
    int mods = 0;

    if (state & ShiftMask)
        mods |= GLFW_MOD_SHIFT;
    if (state & ControlMask)
        mods |= GLFW_MOD_CONTROL;
    if (state & Mod1Mask)
        mods |= GLFW_MOD_ALT;
    if (state & Mod4Mask)
        mods |= GLFW_MOD_SUPER;
    if (state & LockMask)
        mods |= GLFW_MOD_CAPS_LOCK;
    if (state & Mod2Mask)
        mods |= GLFW_MOD_NUM_LOCK;

    return mods;
}

// Translates an X11 key code to a GLFW key token
//
static int translateKey(int scancode)
{
    // Use the pre-filled LUT (see createKeyTables() in x11_init.c)
    if (scancode < 0 || scancode > 255)
        return GLFW_KEY_UNKNOWN;

    return __glfw.x11.keycodes[scancode];
}

// Sends an EWMH or ICCCM event to the window manager
//
static void sendEventToWM(_GLFWwindow* window, Atom type,
                          long a, long b, long c, long d, long e)
{
    XEvent event = { ClientMessage };
    event.xclient.window = window->x11.handle;
    event.xclient.format = 32; // Data is 32-bit longs
    event.xclient.message_type = type;
    event.xclient.data.l[0] = a;
    event.xclient.data.l[1] = b;
    event.xclient.data.l[2] = c;
    event.xclient.data.l[3] = d;
    event.xclient.data.l[4] = e;

    XSendEvent(__glfw.x11.display, __glfw.x11.root,
               False,
               SubstructureNotifyMask | SubstructureRedirectMask,
               &event);
}

// Updates the normal hints according to the window settings
//
static void updateNormalHints(_GLFWwindow* window, int width, int height)
{
    XSizeHints* hints = XAllocSizeHints();

    long supplied;
    XGetWMNormalHints(__glfw.x11.display, window->x11.handle, hints, &supplied);

    hints->flags &= ~(PMinSize | PMaxSize | PAspect);

    if (!window->monitor)
    {
        if (window->resizable)
        {
            if (window->minwidth != GLFW_DONT_CARE &&
                window->minheight != GLFW_DONT_CARE)
            {
                hints->flags |= PMinSize;
                hints->min_width = window->minwidth;
                hints->min_height = window->minheight;
            }

            if (window->maxwidth != GLFW_DONT_CARE &&
                window->maxheight != GLFW_DONT_CARE)
            {
                hints->flags |= PMaxSize;
                hints->max_width = window->maxwidth;
                hints->max_height = window->maxheight;
            }

            if (window->numer != GLFW_DONT_CARE &&
                window->denom != GLFW_DONT_CARE)
            {
                hints->flags |= PAspect;
                hints->min_aspect.x = hints->max_aspect.x = window->numer;
                hints->min_aspect.y = hints->max_aspect.y = window->denom;
            }
        }
        else
        {
            hints->flags |= (PMinSize | PMaxSize);
            hints->min_width  = hints->max_width  = width;
            hints->min_height = hints->max_height = height;
        }
    }

    XSetWMNormalHints(__glfw.x11.display, window->x11.handle, hints);
    XFree(hints);
}

// Updates the full screen status of the window
//
static void updateWindowMode(_GLFWwindow* window)
{
    if (window->monitor)
    {
        if (__glfw.x11.xinerama.available &&
            __glfw.x11.NET_WM_FULLSCREEN_MONITORS)
        {
            sendEventToWM(window,
                          __glfw.x11.NET_WM_FULLSCREEN_MONITORS,
                          window->monitor->x11.index,
                          window->monitor->x11.index,
                          window->monitor->x11.index,
                          window->monitor->x11.index,
                          0);
        }

        if (__glfw.x11.NET_WM_STATE && __glfw.x11.NET_WM_STATE_FULLSCREEN)
        {
            sendEventToWM(window,
                          __glfw.x11.NET_WM_STATE,
                          _NET_WM_STATE_ADD,
                          __glfw.x11.NET_WM_STATE_FULLSCREEN,
                          0, 1, 0);
        }
        else
        {
            // This is the butcher's way of removing window decorations
            // Setting the override-redirect attribute on a window makes the
            // window manager ignore the window completely (ICCCM, section 4)
            // The good thing is that this makes undecorated full screen windows
            // easy to do; the bad thing is that we have to do everything
            // manually and some things (like iconify/restore) won't work at
            // all, as those are tasks usually performed by the window manager

            XSetWindowAttributes attributes;
            attributes.override_redirect = True;
            XChangeWindowAttributes(__glfw.x11.display,
                                    window->x11.handle,
                                    CWOverrideRedirect,
                                    &attributes);

            window->x11.overrideRedirect = GLFW_TRUE;
        }

        // Enable compositor bypass
        if (!window->x11.transparent)
        {
            const unsigned long value = 1;

            XChangeProperty(__glfw.x11.display,  window->x11.handle,
                            __glfw.x11.NET_WM_BYPASS_COMPOSITOR, XA_CARDINAL, 32,
                            PropModeReplace, (unsigned char*) &value, 1);
        }
    }
    else
    {
        if (__glfw.x11.xinerama.available &&
            __glfw.x11.NET_WM_FULLSCREEN_MONITORS)
        {
            XDeleteProperty(__glfw.x11.display, window->x11.handle,
                            __glfw.x11.NET_WM_FULLSCREEN_MONITORS);
        }

        if (__glfw.x11.NET_WM_STATE && __glfw.x11.NET_WM_STATE_FULLSCREEN)
        {
            sendEventToWM(window,
                          __glfw.x11.NET_WM_STATE,
                          _NET_WM_STATE_REMOVE,
                          __glfw.x11.NET_WM_STATE_FULLSCREEN,
                          0, 1, 0);
        }
        else
        {
            XSetWindowAttributes attributes;
            attributes.override_redirect = False;
            XChangeWindowAttributes(__glfw.x11.display,
                                    window->x11.handle,
                                    CWOverrideRedirect,
                                    &attributes);

            window->x11.overrideRedirect = GLFW_FALSE;
        }

        // Disable compositor bypass
        if (!window->x11.transparent)
        {
            XDeleteProperty(__glfw.x11.display, window->x11.handle,
                            __glfw.x11.NET_WM_BYPASS_COMPOSITOR);
        }
    }
}

// Decode a Unicode code point from a UTF-8 stream
// Based on cutef8 by Jeff Bezanson (Public Domain)
//
static uint32_t decodeUTF8(const char** s)
{
    uint32_t codepoint = 0, count = 0;
    static const uint32_t offsets[] =
    {
        0x00000000u, 0x00003080u, 0x000e2080u,
        0x03c82080u, 0xfa082080u, 0x82082080u
    };

    do
    {
        codepoint = (codepoint << 6) + (unsigned char) **s;
        (*s)++;
        count++;
    } while ((**s & 0xc0) == 0x80);

    assert(count <= 6);
    return codepoint - offsets[count - 1];
}

// Convert the specified Latin-1 string to UTF-8
//
static char* convertLatin1toUTF8(const char* source)
{
    size_t size = 1;
    const char* sp;

    for (sp = source;  *sp;  sp++)
        size += (*sp & 0x80) ? 2 : 1;

    char* target = __glfw_calloc(size, 1);
    char* tp = target;

    for (sp = source;  *sp;  sp++)
        tp += ___glfwEncodeUTF8(tp, *sp);

    return target;
}

// Updates the cursor image according to its cursor mode
//
static void updateCursorImage(_GLFWwindow* window)
{
    if (window->cursorMode == GLFW_CURSOR_NORMAL ||
        window->cursorMode == GLFW_CURSOR_CAPTURED)
    {
        if (window->cursor)
        {
            XDefineCursor(__glfw.x11.display, window->x11.handle,
                          window->cursor->x11.handle);
        }
        else
            XUndefineCursor(__glfw.x11.display, window->x11.handle);
    }
    else
    {
        XDefineCursor(__glfw.x11.display, window->x11.handle,
                      __glfw.x11.hiddenCursorHandle);
    }
}

// Grabs the cursor and confines it to the window
//
static void captureCursor(_GLFWwindow* window)
{
    XGrabPointer(__glfw.x11.display, window->x11.handle, True,
                 ButtonPressMask | ButtonReleaseMask | PointerMotionMask,
                 GrabModeAsync, GrabModeAsync,
                 window->x11.handle,
                 None,
                 CurrentTime);
}

// Ungrabs the cursor
//
static void releaseCursor(void)
{
    XUngrabPointer(__glfw.x11.display, CurrentTime);
}

// Enable XI2 raw mouse motion events
//
static void enableRawMouseMotion(_GLFWwindow* window)
{
    XIEventMask em;
    unsigned char mask[XIMaskLen(XI_RawMotion)] = { 0 };

    em.deviceid = XIAllMasterDevices;
    em.mask_len = sizeof(mask);
    em.mask = mask;
    XISetMask(mask, XI_RawMotion);

    XISelectEvents(__glfw.x11.display, __glfw.x11.root, &em, 1);
}

// Disable XI2 raw mouse motion events
//
static void disableRawMouseMotion(_GLFWwindow* window)
{
    XIEventMask em;
    unsigned char mask[] = { 0 };

    em.deviceid = XIAllMasterDevices;
    em.mask_len = sizeof(mask);
    em.mask = mask;

    XISelectEvents(__glfw.x11.display, __glfw.x11.root, &em, 1);
}

// Apply disabled cursor mode to a focused window
//
static void disableCursor(_GLFWwindow* window)
{
    if (window->rawMouseMotion)
        enableRawMouseMotion(window);

    __glfw.x11.disabledCursorWindow = window;
    ___glfwGetCursorPosX11(window,
                         &__glfw.x11.restoreCursorPosX,
                         &__glfw.x11.restoreCursorPosY);
    updateCursorImage(window);
    ___glfwCenterCursorInContentArea(window);
    captureCursor(window);
}

// Exit disabled cursor mode for the specified window
//
static void enableCursor(_GLFWwindow* window)
{
    if (window->rawMouseMotion)
        disableRawMouseMotion(window);

    __glfw.x11.disabledCursorWindow = NULL;
    releaseCursor();
    ____glfwSetCursorPosX11(window,
                         __glfw.x11.restoreCursorPosX,
                         __glfw.x11.restoreCursorPosY);
    updateCursorImage(window);
}

// Clear its handle when the input context has been destroyed
//
static void inputContextDestroyCallback(XIC ic, XPointer clientData, XPointer callData)
{
    _GLFWwindow* window = (_GLFWwindow*) clientData;
    window->x11.ic = NULL;
}

// Create the X11 window (and its colormap)
//
static GLFWbool createNativeWindow(_GLFWwindow* window,
                                   const _GLFWwndconfig* wndconfig,
                                   Visual* visual, int depth)
{
    int width = wndconfig->width;
    int height = wndconfig->height;

    if (wndconfig->scaleToMonitor)
    {
        width *= __glfw.x11.contentScaleX;
        height *= __glfw.x11.contentScaleY;
    }

    int xpos = 0, ypos = 0;

    if (wndconfig->xpos != GLFW_ANY_POSITION && wndconfig->ypos != GLFW_ANY_POSITION)
    {
        xpos = wndconfig->xpos;
        ypos = wndconfig->ypos;
    }

    // Create a colormap based on the visual used by the current context
    window->x11.colormap = XCreateColormap(__glfw.x11.display,
                                           __glfw.x11.root,
                                           visual,
                                           AllocNone);

    window->x11.transparent = ___glfwIsVisualTransparentX11(visual);

    XSetWindowAttributes wa = { 0 };
    wa.colormap = window->x11.colormap;
    wa.event_mask = StructureNotifyMask | KeyPressMask | KeyReleaseMask |
                    PointerMotionMask | ButtonPressMask | ButtonReleaseMask |
                    ExposureMask | FocusChangeMask | VisibilityChangeMask |
                    EnterWindowMask | LeaveWindowMask | PropertyChangeMask;

    ___glfwGrabErrorHandlerX11();

    window->x11.parent = __glfw.x11.root;
    window->x11.handle = XCreateWindow(__glfw.x11.display,
                                       __glfw.x11.root,
                                       xpos, ypos,
                                       width, height,
                                       0,      // Border width
                                       depth,  // Color depth
                                       InputOutput,
                                       visual,
                                       CWBorderPixel | CWColormap | CWEventMask,
                                       &wa);

    ___glfwReleaseErrorHandlerX11();

    if (!window->x11.handle)
    {
        ____glfwInputErrorX11(GLFW_PLATFORM_ERROR,
                           "X11: Failed to create window");
        return GLFW_FALSE;
    }

    XSaveContext(__glfw.x11.display,
                 window->x11.handle,
                 __glfw.x11.context,
                 (XPointer) window);

    if (!wndconfig->decorated)
        __glfwSetWindowDecoratedX11(window, GLFW_FALSE);

    if (__glfw.x11.NET_WM_STATE && !window->monitor)
    {
        Atom states[3];
        int count = 0;

        if (wndconfig->floating)
        {
            if (__glfw.x11.NET_WM_STATE_ABOVE)
                states[count++] = __glfw.x11.NET_WM_STATE_ABOVE;
        }

        if (wndconfig->maximized)
        {
            if (__glfw.x11.NET_WM_STATE_MAXIMIZED_VERT &&
                __glfw.x11.NET_WM_STATE_MAXIMIZED_HORZ)
            {
                states[count++] = __glfw.x11.NET_WM_STATE_MAXIMIZED_VERT;
                states[count++] = __glfw.x11.NET_WM_STATE_MAXIMIZED_HORZ;
                window->x11.maximized = GLFW_TRUE;
            }
        }

        if (count)
        {
            XChangeProperty(__glfw.x11.display, window->x11.handle,
                            __glfw.x11.NET_WM_STATE, XA_ATOM, 32,
                            PropModeReplace, (unsigned char*) states, count);
        }
    }

    // Declare the WM protocols supported by GLFW
    {
        Atom protocols[] =
        {
            __glfw.x11.WM_DELETE_WINDOW,
            __glfw.x11.NET_WM_PING
        };

        XSetWMProtocols(__glfw.x11.display, window->x11.handle,
                        protocols, sizeof(protocols) / sizeof(Atom));
    }

    // Declare our PID
    {
        const long pid = getpid();

        XChangeProperty(__glfw.x11.display,  window->x11.handle,
                        __glfw.x11.NET_WM_PID, XA_CARDINAL, 32,
                        PropModeReplace,
                        (unsigned char*) &pid, 1);
    }

    if (__glfw.x11.NET_WM_WINDOW_TYPE && __glfw.x11.NET_WM_WINDOW_TYPE_NORMAL)
    {
        Atom type = __glfw.x11.NET_WM_WINDOW_TYPE_NORMAL;
        XChangeProperty(__glfw.x11.display,  window->x11.handle,
                        __glfw.x11.NET_WM_WINDOW_TYPE, XA_ATOM, 32,
                        PropModeReplace, (unsigned char*) &type, 1);
    }

    // Set ICCCM WM_HINTS property
    {
        XWMHints* hints = XAllocWMHints();
        if (!hints)
        {
            ___glfwInputError(GLFW_OUT_OF_MEMORY,
                            "X11: Failed to allocate WM hints");
            return GLFW_FALSE;
        }

        hints->flags = StateHint;
        hints->initial_state = NormalState;

        XSetWMHints(__glfw.x11.display, window->x11.handle, hints);
        XFree(hints);
    }

    // Set ICCCM WM_NORMAL_HINTS property
    {
        XSizeHints* hints = XAllocSizeHints();
        if (!hints)
        {
            ___glfwInputError(GLFW_OUT_OF_MEMORY, "X11: Failed to allocate size hints");
            return GLFW_FALSE;
        }

        if (!wndconfig->resizable)
        {
            hints->flags |= (PMinSize | PMaxSize);
            hints->min_width  = hints->max_width  = width;
            hints->min_height = hints->max_height = height;
        }

        // HACK: Explicitly setting PPosition to any value causes some WMs, notably
        //       Compiz and Metacity, to honor the position of unmapped windows
        if (wndconfig->xpos != GLFW_ANY_POSITION && wndconfig->ypos != GLFW_ANY_POSITION)
        {
            hints->flags |= PPosition;
            hints->x = 0;
            hints->y = 0;
        }

        hints->flags |= PWinGravity;
        hints->win_gravity = StaticGravity;

        XSetWMNormalHints(__glfw.x11.display, window->x11.handle, hints);
        XFree(hints);
    }

    // Set ICCCM WM_CLASS property
    {
        XClassHint* hint = XAllocClassHint();

        if (strlen(wndconfig->x11.instanceName) &&
            strlen(wndconfig->x11.className))
        {
            hint->res_name = (char*) wndconfig->x11.instanceName;
            hint->res_class = (char*) wndconfig->x11.className;
        }
        else
        {
            const char* resourceName = getenv("RESOURCE_NAME");
            if (resourceName && strlen(resourceName))
                hint->res_name = (char*) resourceName;
            else if (strlen(wndconfig->title))
                hint->res_name = (char*) wndconfig->title;
            else
                hint->res_name = (char*) "glfw-application";

            if (strlen(wndconfig->title))
                hint->res_class = (char*) wndconfig->title;
            else
                hint->res_class = (char*) "GLFW-Application";
        }

        XSetClassHint(__glfw.x11.display, window->x11.handle, hint);
        XFree(hint);
    }

    // Announce support for Xdnd (drag and drop)
    {
        const Atom version = _GLFW_XDND_VERSION;
        XChangeProperty(__glfw.x11.display, window->x11.handle,
                        __glfw.x11.XdndAware, XA_ATOM, 32,
                        PropModeReplace, (unsigned char*) &version, 1);
    }

    if (__glfw.x11.im)
        __glfwCreateInputContextX11(window);

    ___glfwSetWindowTitleX11(window, wndconfig->title);
    ___glfwGetWindowPosX11(window, &window->x11.xpos, &window->x11.ypos);
    ___glfwGetWindowSizeX11(window, &window->x11.width, &window->x11.height);

    return GLFW_TRUE;
}

// Set the specified property to the selection converted to the requested target
//
static Atom writeTargetToProperty(const XSelectionRequestEvent* request)
{
    char* selectionString = NULL;
    const Atom formats[] = { __glfw.x11.UTF8_STRING, XA_STRING };
    const int formatCount = sizeof(formats) / sizeof(formats[0]);

    if (request->selection == __glfw.x11.PRIMARY)
        selectionString = __glfw.x11.primarySelectionString;
    else
        selectionString = __glfw.x11.clipboardString;

    if (request->property == None)
    {
        // The requester is a legacy client (ICCCM section 2.2)
        // We don't support legacy clients, so fail here
        return None;
    }

    if (request->target == __glfw.x11.TARGETS)
    {
        // The list of supported targets was requested

        const Atom targets[] = { __glfw.x11.TARGETS,
                                 __glfw.x11.MULTIPLE,
                                 __glfw.x11.UTF8_STRING,
                                 XA_STRING };

        XChangeProperty(__glfw.x11.display,
                        request->requestor,
                        request->property,
                        XA_ATOM,
                        32,
                        PropModeReplace,
                        (unsigned char*) targets,
                        sizeof(targets) / sizeof(targets[0]));

        return request->property;
    }

    if (request->target == __glfw.x11.MULTIPLE)
    {
        // Multiple conversions were requested

        Atom* targets;
        const unsigned long count =
            ___glfwGetWindowPropertyX11(request->requestor,
                                      request->property,
                                      __glfw.x11.ATOM_PAIR,
                                      (unsigned char**) &targets);

        for (unsigned long i = 0;  i < count;  i += 2)
        {
            int j;

            for (j = 0;  j < formatCount;  j++)
            {
                if (targets[i] == formats[j])
                    break;
            }

            if (j < formatCount)
            {
                XChangeProperty(__glfw.x11.display,
                                request->requestor,
                                targets[i + 1],
                                targets[i],
                                8,
                                PropModeReplace,
                                (unsigned char *) selectionString,
                                strlen(selectionString));
            }
            else
                targets[i + 1] = None;
        }

        XChangeProperty(__glfw.x11.display,
                        request->requestor,
                        request->property,
                        __glfw.x11.ATOM_PAIR,
                        32,
                        PropModeReplace,
                        (unsigned char*) targets,
                        count);

        XFree(targets);

        return request->property;
    }

    if (request->target == __glfw.x11.SAVE_TARGETS)
    {
        // The request is a check whether we support SAVE_TARGETS
        // It should be handled as a no-op side effect target

        XChangeProperty(__glfw.x11.display,
                        request->requestor,
                        request->property,
                        __glfw.x11.NULL_,
                        32,
                        PropModeReplace,
                        NULL,
                        0);

        return request->property;
    }

    // Conversion to a data target was requested

    for (int i = 0;  i < formatCount;  i++)
    {
        if (request->target == formats[i])
        {
            // The requested target is one we support

            XChangeProperty(__glfw.x11.display,
                            request->requestor,
                            request->property,
                            request->target,
                            8,
                            PropModeReplace,
                            (unsigned char *) selectionString,
                            strlen(selectionString));

            return request->property;
        }
    }

    // The requested target is not supported

    return None;
}

static void handleSelectionRequest(XEvent* event)
{
    const XSelectionRequestEvent* request = &event->xselectionrequest;

    XEvent reply = { SelectionNotify };
    reply.xselection.property = writeTargetToProperty(request);
    reply.xselection.display = request->display;
    reply.xselection.requestor = request->requestor;
    reply.xselection.selection = request->selection;
    reply.xselection.target = request->target;
    reply.xselection.time = request->time;

    XSendEvent(__glfw.x11.display, request->requestor, False, 0, &reply);
}

static const char* getSelectionString(Atom selection)
{
    char** selectionString = NULL;
    const Atom targets[] = { __glfw.x11.UTF8_STRING, XA_STRING };
    const size_t targetCount = sizeof(targets) / sizeof(targets[0]);

    if (selection == __glfw.x11.PRIMARY)
        selectionString = &__glfw.x11.primarySelectionString;
    else
        selectionString = &__glfw.x11.clipboardString;

    if (XGetSelectionOwner(__glfw.x11.display, selection) ==
        __glfw.x11.helperWindowHandle)
    {
        // Instead of doing a large number of X round-trips just to put this
        // string into a window property and then read it back, just return it
        return *selectionString;
    }

    __glfw_free(*selectionString);
    *selectionString = NULL;

    for (size_t i = 0;  i < targetCount;  i++)
    {
        char* data;
        Atom actualType;
        int actualFormat;
        unsigned long itemCount, bytesAfter;
        XEvent notification, dummy;

        XConvertSelection(__glfw.x11.display,
                          selection,
                          targets[i],
                          __glfw.x11.GLFW_SELECTION,
                          __glfw.x11.helperWindowHandle,
                          CurrentTime);

        while (!XCheckTypedWindowEvent(__glfw.x11.display,
                                       __glfw.x11.helperWindowHandle,
                                       SelectionNotify,
                                       &notification))
        {
            waitForX11Event(NULL);
        }

        if (notification.xselection.property == None)
            continue;

        XCheckIfEvent(__glfw.x11.display,
                      &dummy,
                      isSelPropNewValueNotify,
                      (XPointer) &notification);

        XGetWindowProperty(__glfw.x11.display,
                           notification.xselection.requestor,
                           notification.xselection.property,
                           0,
                           LONG_MAX,
                           True,
                           AnyPropertyType,
                           &actualType,
                           &actualFormat,
                           &itemCount,
                           &bytesAfter,
                           (unsigned char**) &data);

        if (actualType == __glfw.x11.INCR)
        {
            size_t size = 1;
            char* string = NULL;

            for (;;)
            {
                while (!XCheckIfEvent(__glfw.x11.display,
                                      &dummy,
                                      isSelPropNewValueNotify,
                                      (XPointer) &notification))
                {
                    waitForX11Event(NULL);
                }

                XFree(data);
                XGetWindowProperty(__glfw.x11.display,
                                   notification.xselection.requestor,
                                   notification.xselection.property,
                                   0,
                                   LONG_MAX,
                                   True,
                                   AnyPropertyType,
                                   &actualType,
                                   &actualFormat,
                                   &itemCount,
                                   &bytesAfter,
                                   (unsigned char**) &data);

                if (itemCount)
                {
                    size += itemCount;
                    string = __glfw_realloc(string, size);
                    string[size - itemCount - 1] = '\0';
                    strcat(string, data);
                }

                if (!itemCount)
                {
                    if (string)
                    {
                        if (targets[i] == XA_STRING)
                        {
                            *selectionString = convertLatin1toUTF8(string);
                            __glfw_free(string);
                        }
                        else
                            *selectionString = string;
                    }

                    break;
                }
            }
        }
        else if (actualType == targets[i])
        {
            if (targets[i] == XA_STRING)
                *selectionString = convertLatin1toUTF8(data);
            else
                *selectionString = ___glfw_strdup(data);
        }

        XFree(data);

        if (*selectionString)
            break;
    }

    if (!*selectionString)
    {
        ___glfwInputError(GLFW_FORMAT_UNAVAILABLE,
                        "X11: Failed to convert selection to string");
    }

    return *selectionString;
}

// Make the specified window and its video mode active on its monitor
//
static void acquireMonitor(_GLFWwindow* window)
{
    if (__glfw.x11.saver.count == 0)
    {
        // Remember old screen saver settings
        XGetScreenSaver(__glfw.x11.display,
                        &__glfw.x11.saver.timeout,
                        &__glfw.x11.saver.interval,
                        &__glfw.x11.saver.blanking,
                        &__glfw.x11.saver.exposure);

        // Disable screen saver
        XSetScreenSaver(__glfw.x11.display, 0, 0, DontPreferBlanking,
                        DefaultExposures);
    }

    if (!window->monitor->windowww)
        __glfw.x11.saver.count++;

    ___glfwSetVideoModeX11(window->monitor, &window->videoMode);

    if (window->x11.overrideRedirect)
    {
        int xpos, ypos;
        GLFWvidmode mode;

        // Manually position the window over its monitor
        ___glfwGetMonitorPosX11(window->monitor, &xpos, &ypos);
        ___glfwGetVideoModeX11(window->monitor, &mode);

        XMoveResizeWindow(__glfw.x11.display, window->x11.handle,
                          xpos, ypos, mode.width, mode.height);
    }

    ____glfwInputMonitorWindow(window->monitor, window);
}

// Remove the window and restore the original video mode
//
static void releaseMonitor(_GLFWwindow* window)
{
    if (window->monitor->windowww != window)
        return;

    ____glfwInputMonitorWindow(window->monitor, NULL);
    ___glfwRestoreVideoModeX11(window->monitor);

    __glfw.x11.saver.count--;

    if (__glfw.x11.saver.count == 0)
    {
        // Restore old screen saver settings
        XSetScreenSaver(__glfw.x11.display,
                        __glfw.x11.saver.timeout,
                        __glfw.x11.saver.interval,
                        __glfw.x11.saver.blanking,
                        __glfw.x11.saver.exposure);
    }
}

// Process the specified X event
//
static void processEvent(XEvent *event)
{
    int keycode = 0;
    Bool filtered = False;

    // HACK: Save scancode as some IMs clear the field in XFilterEvent
    if (event->type == KeyPress || event->type == KeyRelease)
        keycode = event->xkey.keycode;

    filtered = XFilterEvent(event, None);

    if (__glfw.x11.randr.available)
    {
        if (event->type == __glfw.x11.randr.eventBase + RRNotify)
        {
            XRRUpdateConfiguration(event);
            ___glfwPollMonitorsX11();
            return;
        }
    }

    if (__glfw.x11.xkb.available)
    {
        if (event->type == __glfw.x11.xkb.eventBase + XkbEventCode)
        {
            if (((XkbEvent*) event)->any.xkb_type == XkbStateNotify &&
                (((XkbEvent*) event)->state.changed & XkbGroupStateMask))
            {
                __glfw.x11.xkb.group = ((XkbEvent*) event)->state.group;
            }

            return;
        }
    }

    if (event->type == GenericEvent)
    {
        if (__glfw.x11.xi.available)
        {
            _GLFWwindow* window = __glfw.x11.disabledCursorWindow;

            if (window &&
                window->rawMouseMotion &&
                event->xcookie.extension == __glfw.x11.xi.majorOpcode &&
                XGetEventData(__glfw.x11.display, &event->xcookie) &&
                event->xcookie.evtype == XI_RawMotion)
            {
                XIRawEvent* re = event->xcookie.data;
                if (re->valuators.mask_len)
                {
                    const double* values = re->raw_values;
                    double xpos = window->virtualCursorPosX;
                    double ypos = window->virtualCursorPosY;

                    if (XIMaskIsSet(re->valuators.mask, 0))
                    {
                        xpos += *values;
                        values++;
                    }

                    if (XIMaskIsSet(re->valuators.mask, 1))
                        ypos += *values;

                    ___glfwInputCursorPos(window, xpos, ypos);
                }
            }

            XFreeEventData(__glfw.x11.display, &event->xcookie);
        }

        return;
    }

    if (event->type == SelectionRequest)
    {
        handleSelectionRequest(event);
        return;
    }

    _GLFWwindow* window = NULL;
    if (XFindContext(__glfw.x11.display,
                     event->xany.window,
                     __glfw.x11.context,
                     (XPointer*) &window) != 0)
    {
        // This is an event for a window that has already been destroyed
        return;
    }

    switch (event->type)
    {
        case ReparentNotify:
        {
            window->x11.parent = event->xreparent.parent;
            return;
        }

        case KeyPress:
        {
            const int key = translateKey(keycode);
            const int mods = translateState(event->xkey.state);
            const int plain = !(mods & (GLFW_MOD_CONTROL | GLFW_MOD_ALT));

            if (window->x11.ic)
            {
                // HACK: Do not report the key press events duplicated by XIM
                //       Duplicate key releases are filtered out implicitly by
                //       the GLFW key repeat logic in ___glfwInputKey
                //       A timestamp per key is used to handle simultaneous keys
                // NOTE: Always allow the first event for each key through
                //       (the server never sends a timestamp of zero)
                // NOTE: Timestamp difference is compared to handle wrap-around
                Time diff = event->xkey.time - window->x11.keyPressTimes[keycode];
                if (diff == event->xkey.time || (diff > 0 && diff < ((Time)1 << 31)))
                {
                    if (keycode)
                        ___glfwInputKey(window, key, keycode, GLFW_PRESS, mods);

                    window->x11.keyPressTimes[keycode] = event->xkey.time;
                }

                if (!filtered)
                {
                    int count;
                    Status status;
                    char buffer[100];
                    char* chars = buffer;

                    count = Xutf8LookupString(window->x11.ic,
                                              &event->xkey,
                                              buffer, sizeof(buffer) - 1,
                                              NULL, &status);

                    if (status == XBufferOverflow)
                    {
                        chars = __glfw_calloc(count + 1, 1);
                        count = Xutf8LookupString(window->x11.ic,
                                                  &event->xkey,
                                                  chars, count,
                                                  NULL, &status);
                    }

                    if (status == XLookupChars || status == XLookupBoth)
                    {
                        const char* c = chars;
                        chars[count] = '\0';
                        while (c - chars < count)
                            ___glfwInputChar(window, decodeUTF8(&c), mods, plain);
                    }

                    if (chars != buffer)
                        __glfw_free(chars);
                }
            }
            else
            {
                KeySym keysym;
                XLookupString(&event->xkey, NULL, 0, &keysym, NULL);

                ___glfwInputKey(window, key, keycode, GLFW_PRESS, mods);

                const uint32_t codepoint = ___glfwKeySym2Unicode(keysym);
                if (codepoint != GLFW_INVALID_CODEPOINT)
                    ___glfwInputChar(window, codepoint, mods, plain);
            }

            return;
        }

        case KeyRelease:
        {
            const int key = translateKey(keycode);
            const int mods = translateState(event->xkey.state);

            if (!__glfw.x11.xkb.detectable)
            {
                // HACK: Key repeat events will arrive as KeyRelease/KeyPress
                //       pairs with similar or identical time stamps
                //       The key repeat logic in ___glfwInputKey expects only key
                //       presses to repeat, so detect and discard release events
                if (XEventsQueued(__glfw.x11.display, QueuedAfterReading))
                {
                    XEvent next;
                    XPeekEvent(__glfw.x11.display, &next);

                    if (next.type == KeyPress &&
                        next.xkey.window == event->xkey.window &&
                        next.xkey.keycode == keycode)
                    {
                        // HACK: The time of repeat events sometimes doesn't
                        //       match that of the press event, so add an
                        //       epsilon
                        //       Toshiyuki Takahashi can press a button
                        //       16 times per second so it's fairly safe to
                        //       assume that no human is pressing the key 50
                        //       times per second (value is ms)
                        if ((next.xkey.time - event->xkey.time) < 20)
                        {
                            // This is very likely a server-generated key repeat
                            // event, so ignore it
                            return;
                        }
                    }
                }
            }

            ___glfwInputKey(window, key, keycode, GLFW_RELEASE, mods);
            return;
        }

        case ButtonPress:
        {
            const int mods = translateState(event->xbutton.state);

            if (event->xbutton.button == Button1)
                ___glfwInputMouseClick(window, GLFW_MOUSE_BUTTON_LEFT, GLFW_PRESS, mods);
            else if (event->xbutton.button == Button2)
                ___glfwInputMouseClick(window, GLFW_MOUSE_BUTTON_MIDDLE, GLFW_PRESS, mods);
            else if (event->xbutton.button == Button3)
                ___glfwInputMouseClick(window, GLFW_MOUSE_BUTTON_RIGHT, GLFW_PRESS, mods);

            // Modern X provides scroll events as mouse button presses
            else if (event->xbutton.button == Button4)
                ___glfwInputScroll(window, 0.0, 1.0);
            else if (event->xbutton.button == Button5)
                ___glfwInputScroll(window, 0.0, -1.0);
            else if (event->xbutton.button == Button6)
                ___glfwInputScroll(window, 1.0, 0.0);
            else if (event->xbutton.button == Button7)
                ___glfwInputScroll(window, -1.0, 0.0);

            else
            {
                // Additional buttons after 7 are treated as regular buttons
                // We subtract 4 to fill the gap left by scroll input above
                ___glfwInputMouseClick(window,
                                     event->xbutton.button - Button1 - 4,
                                     GLFW_PRESS,
                                     mods);
            }

            return;
        }

        case ButtonRelease:
        {
            const int mods = translateState(event->xbutton.state);

            if (event->xbutton.button == Button1)
            {
                ___glfwInputMouseClick(window,
                                     GLFW_MOUSE_BUTTON_LEFT,
                                     GLFW_RELEASE,
                                     mods);
            }
            else if (event->xbutton.button == Button2)
            {
                ___glfwInputMouseClick(window,
                                     GLFW_MOUSE_BUTTON_MIDDLE,
                                     GLFW_RELEASE,
                                     mods);
            }
            else if (event->xbutton.button == Button3)
            {
                ___glfwInputMouseClick(window,
                                     GLFW_MOUSE_BUTTON_RIGHT,
                                     GLFW_RELEASE,
                                     mods);
            }
            else if (event->xbutton.button > Button7)
            {
                // Additional buttons after 7 are treated as regular buttons
                // We subtract 4 to fill the gap left by scroll input above
                ___glfwInputMouseClick(window,
                                     event->xbutton.button - Button1 - 4,
                                     GLFW_RELEASE,
                                     mods);
            }

            return;
        }

        case EnterNotify:
        {
            // XEnterWindowEvent is XCrossingEvent
            const int x = event->xcrossing.x;
            const int y = event->xcrossing.y;

            // HACK: This is a workaround for WMs (KWM, Fluxbox) that otherwise
            //       ignore the defined cursor for hidden cursor mode
            if (window->cursorMode == GLFW_CURSOR_HIDDEN)
                updateCursorImage(window);

            ___glfwInputCursorEnter(window, GLFW_TRUE);
            ___glfwInputCursorPos(window, x, y);

            window->x11.lastCursorPosX = x;
            window->x11.lastCursorPosY = y;
            return;
        }

        case LeaveNotify:
        {
            ___glfwInputCursorEnter(window, GLFW_FALSE);
            return;
        }

        case MotionNotify:
        {
            const int x = event->xmotion.x;
            const int y = event->xmotion.y;

            if (x != window->x11.warpCursorPosX ||
                y != window->x11.warpCursorPosY)
            {
                // The cursor was moved by something other than GLFW

                if (window->cursorMode == GLFW_CURSOR_DISABLED)
                {
                    if (__glfw.x11.disabledCursorWindow != window)
                        return;
                    if (window->rawMouseMotion)
                        return;

                    const int dx = x - window->x11.lastCursorPosX;
                    const int dy = y - window->x11.lastCursorPosY;

                    ___glfwInputCursorPos(window,
                                        window->virtualCursorPosX + dx,
                                        window->virtualCursorPosY + dy);
                }
                else
                    ___glfwInputCursorPos(window, x, y);
            }

            window->x11.lastCursorPosX = x;
            window->x11.lastCursorPosY = y;
            return;
        }

        case ConfigureNotify:
        {
            if (event->xconfigure.width != window->x11.width ||
                event->xconfigure.height != window->x11.height)
            {
                ___glfwInputFramebufferSize(window,
                                          event->xconfigure.width,
                                          event->xconfigure.height);

                ___glfwInputWindowSize(window,
                                     event->xconfigure.width,
                                     event->xconfigure.height);

                window->x11.width = event->xconfigure.width;
                window->x11.height = event->xconfigure.height;
            }

            int xpos = event->xconfigure.x;
            int ypos = event->xconfigure.y;

            // NOTE: ConfigureNotify events from the server are in local
            //       coordinates, so if we are reparented we need to translate
            //       the position into root (screen) coordinates
            if (!event->xany.send_event && window->x11.parent != __glfw.x11.root)
            {
                ___glfwGrabErrorHandlerX11();

                Window dummy;
                XTranslateCoordinates(__glfw.x11.display,
                                      window->x11.parent,
                                      __glfw.x11.root,
                                      xpos, ypos,
                                      &xpos, &ypos,
                                      &dummy);

                ___glfwReleaseErrorHandlerX11();
                if (__glfw.x11.errorCode == BadWindow)
                    return;
            }

            if (xpos != window->x11.xpos || ypos != window->x11.ypos)
            {
                ___glfwInputWindowPos(window, xpos, ypos);
                window->x11.xpos = xpos;
                window->x11.ypos = ypos;
            }

            return;
        }

        case ClientMessage:
        {
            // Custom client message, probably from the window manager

            if (filtered)
                return;

            if (event->xclient.message_type == None)
                return;

            if (event->xclient.message_type == __glfw.x11.WM_PROTOCOLS)
            {
                const Atom protocol = event->xclient.data.l[0];
                if (protocol == None)
                    return;

                if (protocol == __glfw.x11.WM_DELETE_WINDOW)
                {
                    // The window manager was asked to close the window, for
                    // example by the user pressing a 'close' window decoration
                    // button
                    ___glfwInputWindowCloseRequest(window);
                }
                else if (protocol == __glfw.x11.NET_WM_PING)
                {
                    // The window manager is pinging the application to ensure
                    // it's still responding to events

                    XEvent reply = *event;
                    reply.xclient.window = __glfw.x11.root;

                    XSendEvent(__glfw.x11.display, __glfw.x11.root,
                               False,
                               SubstructureNotifyMask | SubstructureRedirectMask,
                               &reply);
                }
            }
            else if (event->xclient.message_type == __glfw.x11.XdndEnter)
            {
                // A drag operation has entered the window
                unsigned long count;
                Atom* formats = NULL;
                const GLFWbool list = event->xclient.data.l[1] & 1;

                __glfw.x11.xdnd.source  = event->xclient.data.l[0];
                __glfw.x11.xdnd.version = event->xclient.data.l[1] >> 24;
                __glfw.x11.xdnd.format  = None;

                if (__glfw.x11.xdnd.version > _GLFW_XDND_VERSION)
                    return;

                if (list)
                {
                    count = ___glfwGetWindowPropertyX11(__glfw.x11.xdnd.source,
                                                      __glfw.x11.XdndTypeList,
                                                      XA_ATOM,
                                                      (unsigned char**) &formats);
                }
                else
                {
                    count = 3;
                    formats = (Atom*) event->xclient.data.l + 2;
                }

                for (unsigned int i = 0;  i < count;  i++)
                {
                    if (formats[i] == __glfw.x11.text_uri_list)
                    {
                        __glfw.x11.xdnd.format = __glfw.x11.text_uri_list;
                        break;
                    }
                }

                if (list && formats)
                    XFree(formats);
            }
            else if (event->xclient.message_type == __glfw.x11.XdndDrop)
            {
                // The drag operation has finished by dropping on the window
                Time time = CurrentTime;

                if (__glfw.x11.xdnd.version > _GLFW_XDND_VERSION)
                    return;

                if (__glfw.x11.xdnd.format)
                {
                    if (__glfw.x11.xdnd.version >= 1)
                        time = event->xclient.data.l[2];

                    // Request the chosen format from the source window
                    XConvertSelection(__glfw.x11.display,
                                      __glfw.x11.XdndSelection,
                                      __glfw.x11.xdnd.format,
                                      __glfw.x11.XdndSelection,
                                      window->x11.handle,
                                      time);
                }
                else if (__glfw.x11.xdnd.version >= 2)
                {
                    XEvent reply = { ClientMessage };
                    reply.xclient.window = __glfw.x11.xdnd.source;
                    reply.xclient.message_type = __glfw.x11.XdndFinished;
                    reply.xclient.format = 32;
                    reply.xclient.data.l[0] = window->x11.handle;
                    reply.xclient.data.l[1] = 0; // The drag was rejected
                    reply.xclient.data.l[2] = None;

                    XSendEvent(__glfw.x11.display, __glfw.x11.xdnd.source,
                               False, NoEventMask, &reply);
                    XFlush(__glfw.x11.display);
                }
            }
            else if (event->xclient.message_type == __glfw.x11.XdndPosition)
            {
                // The drag operation has moved over the window
                const int xabs = (event->xclient.data.l[2] >> 16) & 0xffff;
                const int yabs = (event->xclient.data.l[2]) & 0xffff;
                Window dummy;
                int xpos, ypos;

                if (__glfw.x11.xdnd.version > _GLFW_XDND_VERSION)
                    return;

                XTranslateCoordinates(__glfw.x11.display,
                                      __glfw.x11.root,
                                      window->x11.handle,
                                      xabs, yabs,
                                      &xpos, &ypos,
                                      &dummy);

                ___glfwInputCursorPos(window, xpos, ypos);

                XEvent reply = { ClientMessage };
                reply.xclient.window = __glfw.x11.xdnd.source;
                reply.xclient.message_type = __glfw.x11.XdndStatus;
                reply.xclient.format = 32;
                reply.xclient.data.l[0] = window->x11.handle;
                reply.xclient.data.l[2] = 0; // Specify an empty rectangle
                reply.xclient.data.l[3] = 0;

                if (__glfw.x11.xdnd.format)
                {
                    // Reply that we are ready to copy the dragged data
                    reply.xclient.data.l[1] = 1; // Accept with no rectangle
                    if (__glfw.x11.xdnd.version >= 2)
                        reply.xclient.data.l[4] = __glfw.x11.XdndActionCopy;
                }

                XSendEvent(__glfw.x11.display, __glfw.x11.xdnd.source,
                           False, NoEventMask, &reply);
                XFlush(__glfw.x11.display);
            }

            return;
        }

        case SelectionNotify:
        {
            if (event->xselection.property == __glfw.x11.XdndSelection)
            {
                // The converted data from the drag operation has arrived
                char* data;
                const unsigned long result =
                    ___glfwGetWindowPropertyX11(event->xselection.requestor,
                                              event->xselection.property,
                                              event->xselection.target,
                                              (unsigned char**) &data);

                if (result)
                {
                    int count;
                    char** paths = ___glfwParseUriList(data, &count);

                    ___glfwInputDrop(window, count, (const char**) paths);

                    for (int i = 0;  i < count;  i++)
                        __glfw_free(paths[i]);
                    __glfw_free(paths);
                }

                if (data)
                    XFree(data);

                if (__glfw.x11.xdnd.version >= 2)
                {
                    XEvent reply = { ClientMessage };
                    reply.xclient.window = __glfw.x11.xdnd.source;
                    reply.xclient.message_type = __glfw.x11.XdndFinished;
                    reply.xclient.format = 32;
                    reply.xclient.data.l[0] = window->x11.handle;
                    reply.xclient.data.l[1] = result;
                    reply.xclient.data.l[2] = __glfw.x11.XdndActionCopy;

                    XSendEvent(__glfw.x11.display, __glfw.x11.xdnd.source,
                               False, NoEventMask, &reply);
                    XFlush(__glfw.x11.display);
                }
            }

            return;
        }

        case FocusIn:
        {
            if (event->xfocus.mode == NotifyGrab ||
                event->xfocus.mode == NotifyUngrab)
            {
                // Ignore focus events from popup indicator windows, window menu
                // key chords and window dragging
                return;
            }

            if (window->cursorMode == GLFW_CURSOR_DISABLED)
                disableCursor(window);
            else if (window->cursorMode == GLFW_CURSOR_CAPTURED)
                captureCursor(window);

            if (window->x11.ic)
                XSetICFocus(window->x11.ic);

            ___glfwInputWindowFocus(window, GLFW_TRUE);
            return;
        }

        case FocusOut:
        {
            if (event->xfocus.mode == NotifyGrab ||
                event->xfocus.mode == NotifyUngrab)
            {
                // Ignore focus events from popup indicator windows, window menu
                // key chords and window dragging
                return;
            }

            if (window->cursorMode == GLFW_CURSOR_DISABLED)
                enableCursor(window);
            else if (window->cursorMode == GLFW_CURSOR_CAPTURED)
                releaseCursor();

            if (window->x11.ic)
                XUnsetICFocus(window->x11.ic);

            if (window->monitor && window->autoIconify)
                ___glfwIconifyWindowX11(window);

            ___glfwInputWindowFocus(window, GLFW_FALSE);
            return;
        }

        case Expose:
        {
            ___glfwInputWindowDamage(window);
            return;
        }

        case PropertyNotify:
        {
            if (event->xproperty.state != PropertyNewValue)
                return;

            if (event->xproperty.atom == __glfw.x11.WM_STATE)
            {
                const int state = getWindowState(window);
                if (state != IconicState && state != NormalState)
                    return;

                const GLFWbool iconified = (state == IconicState);
                if (window->x11.iconified != iconified)
                {
                    if (window->monitor)
                    {
                        if (iconified)
                            releaseMonitor(window);
                        else
                            acquireMonitor(window);
                    }

                    window->x11.iconified = iconified;
                    ___glfwInputWindowIconify(window, iconified);
                }
            }
            else if (event->xproperty.atom == __glfw.x11.NET_WM_STATE)
            {
                const GLFWbool maximized = __glfwWindowMaximizedX11(window);
                if (window->x11.maximized != maximized)
                {
                    window->x11.maximized = maximized;
                    ___glfwInputWindowMaximize(window, maximized);
                }
            }

            return;
        }

        case DestroyNotify:
            return;
    }
}


//////////////////////////////////////////////////////////////////////////
//////                       GLFW internal API                      //////
//////////////////////////////////////////////////////////////////////////

// Retrieve a single window property of the specified type
// Inspired by fghGetWindowProperty from freeglut
//
unsigned long ___glfwGetWindowPropertyX11(Window window,
                                        Atom property,
                                        Atom type,
                                        unsigned char** value)
{
    Atom actualType;
    int actualFormat;
    unsigned long itemCount, bytesAfter;

    XGetWindowProperty(__glfw.x11.display,
                       window,
                       property,
                       0,
                       LONG_MAX,
                       False,
                       type,
                       &actualType,
                       &actualFormat,
                       &itemCount,
                       &bytesAfter,
                       value);

    return itemCount;
}

GLFWbool ___glfwIsVisualTransparentX11(Visual* visual)
{
    if (!__glfw.x11.xrender.available)
        return GLFW_FALSE;

    XRenderPictFormat* pf = XRenderFindVisualFormat(__glfw.x11.display, visual);
    return pf && pf->direct.alphaMask;
}

// Push contents of our selection to clipboard manager
//
void ___glfwPushSelectionToManagerX11(void)
{
    XConvertSelection(__glfw.x11.display,
                      __glfw.x11.CLIPBOARD_MANAGER,
                      __glfw.x11.SAVE_TARGETS,
                      None,
                      __glfw.x11.helperWindowHandle,
                      CurrentTime);

    for (;;)
    {
        XEvent event;

        while (XCheckIfEvent(__glfw.x11.display, &event, isSelectionEvent, NULL))
        {
            switch (event.type)
            {
                case SelectionRequest:
                    handleSelectionRequest(&event);
                    break;

                case SelectionNotify:
                {
                    if (event.xselection.target == __glfw.x11.SAVE_TARGETS)
                    {
                        // This means one of two things; either the selection
                        // was not owned, which means there is no clipboard
                        // manager, or the transfer to the clipboard manager has
                        // completed
                        // In either case, it means we are done here
                        return;
                    }

                    break;
                }
            }
        }

        waitForX11Event(NULL);
    }
}

void __glfwCreateInputContextX11(_GLFWwindow* window)
{
    XIMCallback callback;
    callback.callback = (XIMProc) inputContextDestroyCallback;
    callback.client_data = (XPointer) window;

    window->x11.ic = XCreateIC(__glfw.x11.im,
                               XNInputStyle,
                               XIMPreeditNothing | XIMStatusNothing,
                               XNClientWindow,
                               window->x11.handle,
                               XNFocusWindow,
                               window->x11.handle,
                               XNDestroyCallback,
                               &callback,
                               NULL);

    if (window->x11.ic)
    {
        XWindowAttributes attribs;
        XGetWindowAttributes(__glfw.x11.display, window->x11.handle, &attribs);

        unsigned long filter = 0;
        if (XGetICValues(window->x11.ic, XNFilterEvents, &filter, NULL) == NULL)
        {
            XSelectInput(__glfw.x11.display,
                         window->x11.handle,
                         attribs.your_event_mask | filter);
        }
    }
}


//////////////////////////////////////////////////////////////////////////
//////                       GLFW platform API                      //////
//////////////////////////////////////////////////////////////////////////

GLFWbool ___glfwCreateWindowX11(_GLFWwindow* window,
                              const _GLFWwndconfig* wndconfig,
                              const _GLFWctxconfig* ctxconfig,
                              const _GLFWfbconfig* fbconfig)
{
    Visual* visual = NULL;
    int depth;

    if (ctxconfig->client != GLFW_NO_API)
    {
        if (ctxconfig->source == GLFW_NATIVE_CONTEXT_API)
        {
            if (!____glfwInitGLX())
                return GLFW_FALSE;
            if (!___glfwChooseVisualGLX(wndconfig, ctxconfig, fbconfig, &visual, &depth))
                return GLFW_FALSE;
        }
        else if (ctxconfig->source == GLFW_EGL_CONTEXT_API)
        {
            if (!____glfwInitEGL())
                return GLFW_FALSE;
            if (!___glfwChooseVisualEGL(wndconfig, ctxconfig, fbconfig, &visual, &depth))
                return GLFW_FALSE;
        }
        else if (ctxconfig->source == GLFW_OSMESA_CONTEXT_API)
        {
            if (!____glfwInitOSMesa())
                return GLFW_FALSE;
        }
    }

    if (!visual)
    {
        visual = DefaultVisual(__glfw.x11.display, __glfw.x11.screen);
        depth = DefaultDepth(__glfw.x11.display, __glfw.x11.screen);
    }

    if (!createNativeWindow(window, wndconfig, visual, depth))
        return GLFW_FALSE;

    if (ctxconfig->client != GLFW_NO_API)
    {
        if (ctxconfig->source == GLFW_NATIVE_CONTEXT_API)
        {
            if (!___glfwCreateContextGLX(window, ctxconfig, fbconfig))
                return GLFW_FALSE;
        }
        else if (ctxconfig->source == GLFW_EGL_CONTEXT_API)
        {
            if (!___glfwCreateContextEGL(window, ctxconfig, fbconfig))
                return GLFW_FALSE;
        }
        else if (ctxconfig->source == GLFW_OSMESA_CONTEXT_API)
        {
            if (!___glfwCreateContextOSMesa(window, ctxconfig, fbconfig))
                return GLFW_FALSE;
        }

        if (!___glfwRefreshContextAttribs(window, ctxconfig))
            return GLFW_FALSE;
    }

    if (wndconfig->mousePassthrough)
        __glfwSetWindowMousePassthroughX11(window, GLFW_TRUE);

    if (window->monitor)
    {
        ___glfwShowWindowX11(window);
        updateWindowMode(window);
        acquireMonitor(window);

        if (wndconfig->centerCursor)
            ___glfwCenterCursorInContentArea(window);
    }
    else
    {
        if (wndconfig->visible)
        {
            ___glfwShowWindowX11(window);
            if (wndconfig->focused)
                ___glfwFocusWindowX11(window);
        }
    }

    XFlush(__glfw.x11.display);
    return GLFW_TRUE;
}

void ___glfwDestroyWindowX11(_GLFWwindow* window)
{
    if (__glfw.x11.disabledCursorWindow == window)
        enableCursor(window);

    if (window->monitor)
        releaseMonitor(window);

    if (window->x11.ic)
    {
        XDestroyIC(window->x11.ic);
        window->x11.ic = NULL;
    }

    if (window->context.destroy)
        window->context.destroy(window);

    if (window->x11.handle)
    {
        XDeleteContext(__glfw.x11.display, window->x11.handle, __glfw.x11.context);
        XUnmapWindow(__glfw.x11.display, window->x11.handle);
        XDestroyWindow(__glfw.x11.display, window->x11.handle);
        window->x11.handle = (Window) 0;
    }

    if (window->x11.colormap)
    {
        XFreeColormap(__glfw.x11.display, window->x11.colormap);
        window->x11.colormap = (Colormap) 0;
    }

    XFlush(__glfw.x11.display);
}

void ___glfwSetWindowTitleX11(_GLFWwindow* window, const char* title)
{
    if (__glfw.x11.xlib.utf8)
    {
        Xutf8SetWMProperties(__glfw.x11.display,
                             window->x11.handle,
                             title, title,
                             NULL, 0,
                             NULL, NULL, NULL);
    }

    XChangeProperty(__glfw.x11.display,  window->x11.handle,
                    __glfw.x11.NET_WM_NAME, __glfw.x11.UTF8_STRING, 8,
                    PropModeReplace,
                    (unsigned char*) title, strlen(title));

    XChangeProperty(__glfw.x11.display,  window->x11.handle,
                    __glfw.x11.NET_WM_ICON_NAME, __glfw.x11.UTF8_STRING, 8,
                    PropModeReplace,
                    (unsigned char*) title, strlen(title));

    XFlush(__glfw.x11.display);
}

void ___glfwSetWindowIconX11(_GLFWwindow* window, int count, const GLFWimage* images)
{
    if (count)
    {
        int longCount = 0;

        for (int i = 0;  i < count;  i++)
            longCount += 2 + images[i].width * images[i].height;

        unsigned long* icon = __glfw_calloc(longCount, sizeof(unsigned long));
        unsigned long* target = icon;

        for (int i = 0;  i < count;  i++)
        {
            *target++ = images[i].width;
            *target++ = images[i].height;

            for (int j = 0;  j < images[i].width * images[i].height;  j++)
            {
                *target++ = (((unsigned long) images[i].pixels[j * 4 + 0]) << 16) |
                            (((unsigned long) images[i].pixels[j * 4 + 1]) <<  8) |
                            (((unsigned long) images[i].pixels[j * 4 + 2]) <<  0) |
                            (((unsigned long) images[i].pixels[j * 4 + 3]) << 24);
            }
        }

        // NOTE: XChangeProperty expects 32-bit values like the image data above to be
        //       placed in the 32 least significant bits of individual longs.  This is
        //       true even if long is 64-bit and a WM protocol calls for "packed" data.
        //       This is because of a historical mistake that then became part of the Xlib
        //       ABI.  Xlib will pack these values into a regular array of 32-bit values
        //       before sending it over the wire.
        XChangeProperty(__glfw.x11.display, window->x11.handle,
                        __glfw.x11.NET_WM_ICON,
                        XA_CARDINAL, 32,
                        PropModeReplace,
                        (unsigned char*) icon,
                        longCount);

        __glfw_free(icon);
    }
    else
    {
        XDeleteProperty(__glfw.x11.display, window->x11.handle,
                        __glfw.x11.NET_WM_ICON);
    }

    XFlush(__glfw.x11.display);
}

void ___glfwGetWindowPosX11(_GLFWwindow* window, int* xpos, int* ypos)
{
    Window dummy;
    int x, y;

    XTranslateCoordinates(__glfw.x11.display, window->x11.handle, __glfw.x11.root,
                          0, 0, &x, &y, &dummy);

    if (xpos)
        *xpos = x;
    if (ypos)
        *ypos = y;
}

void ___glfwSetWindowPosX11(_GLFWwindow* window, int xpos, int ypos)
{
    // HACK: Explicitly setting PPosition to any value causes some WMs, notably
    //       Compiz and Metacity, to honor the position of unmapped windows
    if (!__glfwWindowVisibleX11(window))
    {
        long supplied;
        XSizeHints* hints = XAllocSizeHints();

        if (XGetWMNormalHints(__glfw.x11.display, window->x11.handle, hints, &supplied))
        {
            hints->flags |= PPosition;
            hints->x = hints->y = 0;

            XSetWMNormalHints(__glfw.x11.display, window->x11.handle, hints);
        }

        XFree(hints);
    }

    XMoveWindow(__glfw.x11.display, window->x11.handle, xpos, ypos);
    XFlush(__glfw.x11.display);
}

void ___glfwGetWindowSizeX11(_GLFWwindow* window, int* width, int* height)
{
    XWindowAttributes attribs;
    XGetWindowAttributes(__glfw.x11.display, window->x11.handle, &attribs);

    if (width)
        *width = attribs.width;
    if (height)
        *height = attribs.height;
}

void ___glfwSetWindowSizeX11(_GLFWwindow* window, int width, int height)
{
    if (window->monitor)
    {
        if (window->monitor->windowww == window)
            acquireMonitor(window);
    }
    else
    {
        if (!window->resizable)
            updateNormalHints(window, width, height);

        XResizeWindow(__glfw.x11.display, window->x11.handle, width, height);
    }

    XFlush(__glfw.x11.display);
}

void ____glfwSetWindowSizeLimitsX11(_GLFWwindow* window,
                                 int minwidth, int minheight,
                                 int maxwidth, int maxheight)
{
    int width, height;
    ___glfwGetWindowSizeX11(window, &width, &height);
    updateNormalHints(window, width, height);
    XFlush(__glfw.x11.display);
}

void ___glfwSetWindowAspectRatioX11(_GLFWwindow* window, int numer, int denom)
{
    int width, height;
    ___glfwGetWindowSizeX11(window, &width, &height);
    updateNormalHints(window, width, height);
    XFlush(__glfw.x11.display);
}

void ___glfwGetFramebufferSizeX11(_GLFWwindow* window, int* width, int* height)
{
    ___glfwGetWindowSizeX11(window, width, height);
}

void ___glfwGetWindowFrameSizeX11(_GLFWwindow* window,
                                int* left, int* top,
                                int* right, int* bottom)
{
    long* extents = NULL;

    if (window->monitor || !window->decorated)
        return;

    if (__glfw.x11.NET_FRAME_EXTENTS == None)
        return;

    if (!__glfwWindowVisibleX11(window) &&
        __glfw.x11.NET_REQUEST_FRAME_EXTENTS)
    {
        XEvent event;
        double timeout = 0.5;

        // Ensure _NET_FRAME_EXTENTS is set, allowing __glfwGetWindowFrameSize to
        // function before the window is mapped
        sendEventToWM(window, __glfw.x11.NET_REQUEST_FRAME_EXTENTS,
                      0, 0, 0, 0, 0);

        // HACK: Use a timeout because earlier versions of some window managers
        //       (at least Unity, Fluxbox and Xfwm) failed to send the reply
        //       They have been fixed but broken versions are still in the wild
        //       If you are affected by this and your window manager is NOT
        //       listed above, PLEASE report it to their and our issue trackers
        while (!XCheckIfEvent(__glfw.x11.display,
                              &event,
                              isFrameExtentsEvent,
                              (XPointer) window))
        {
            if (!waitForX11Event(&timeout))
            {
                ___glfwInputError(GLFW_PLATFORM_ERROR,
                                "X11: The window manager has a broken _NET_REQUEST_FRAME_EXTENTS implementation; please report this issue");
                return;
            }
        }
    }

    if (___glfwGetWindowPropertyX11(window->x11.handle,
                                  __glfw.x11.NET_FRAME_EXTENTS,
                                  XA_CARDINAL,
                                  (unsigned char**) &extents) == 4)
    {
        if (left)
            *left = extents[0];
        if (top)
            *top = extents[2];
        if (right)
            *right = extents[1];
        if (bottom)
            *bottom = extents[3];
    }

    if (extents)
        XFree(extents);
}

void ___glfwGetWindowContentScaleX11(_GLFWwindow* window, float* xscale, float* yscale)
{
    if (xscale)
        *xscale = __glfw.x11.contentScaleX;
    if (yscale)
        *yscale = __glfw.x11.contentScaleY;
}

void ___glfwIconifyWindowX11(_GLFWwindow* window)
{
    if (window->x11.overrideRedirect)
    {
        // Override-redirect windows cannot be iconified or restored, as those
        // tasks are performed by the window manager
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "X11: Iconification of full screen windows requires a WM that supports EWMH full screen");
        return;
    }

    XIconifyWindow(__glfw.x11.display, window->x11.handle, __glfw.x11.screen);
    XFlush(__glfw.x11.display);
}

void ___glfwRestoreWindowX11(_GLFWwindow* window)
{
    if (window->x11.overrideRedirect)
    {
        // Override-redirect windows cannot be iconified or restored, as those
        // tasks are performed by the window manager
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "X11: Iconification of full screen windows requires a WM that supports EWMH full screen");
        return;
    }

    if (__glfwWindowIconifiedX11(window))
    {
        XMapWindow(__glfw.x11.display, window->x11.handle);
        waitForVisibilityNotify(window);
    }
    else if (__glfwWindowVisibleX11(window))
    {
        if (__glfw.x11.NET_WM_STATE &&
            __glfw.x11.NET_WM_STATE_MAXIMIZED_VERT &&
            __glfw.x11.NET_WM_STATE_MAXIMIZED_HORZ)
        {
            sendEventToWM(window,
                          __glfw.x11.NET_WM_STATE,
                          _NET_WM_STATE_REMOVE,
                          __glfw.x11.NET_WM_STATE_MAXIMIZED_VERT,
                          __glfw.x11.NET_WM_STATE_MAXIMIZED_HORZ,
                          1, 0);
        }
    }

    XFlush(__glfw.x11.display);
}

void ___glfwMaximizeWindowX11(_GLFWwindow* window)
{
    if (!__glfw.x11.NET_WM_STATE ||
        !__glfw.x11.NET_WM_STATE_MAXIMIZED_VERT ||
        !__glfw.x11.NET_WM_STATE_MAXIMIZED_HORZ)
    {
        return;
    }

    if (__glfwWindowVisibleX11(window))
    {
        sendEventToWM(window,
                    __glfw.x11.NET_WM_STATE,
                    _NET_WM_STATE_ADD,
                    __glfw.x11.NET_WM_STATE_MAXIMIZED_VERT,
                    __glfw.x11.NET_WM_STATE_MAXIMIZED_HORZ,
                    1, 0);
    }
    else
    {
        Atom* states = NULL;
        unsigned long count =
            ___glfwGetWindowPropertyX11(window->x11.handle,
                                      __glfw.x11.NET_WM_STATE,
                                      XA_ATOM,
                                      (unsigned char**) &states);

        // NOTE: We don't check for failure as this property may not exist yet
        //       and that's fine (and we'll create it implicitly with append)

        Atom missing[2] =
        {
            __glfw.x11.NET_WM_STATE_MAXIMIZED_VERT,
            __glfw.x11.NET_WM_STATE_MAXIMIZED_HORZ
        };
        unsigned long missingCount = 2;

        for (unsigned long i = 0;  i < count;  i++)
        {
            for (unsigned long j = 0;  j < missingCount;  j++)
            {
                if (states[i] == missing[j])
                {
                    missing[j] = missing[missingCount - 1];
                    missingCount--;
                }
            }
        }

        if (states)
            XFree(states);

        if (!missingCount)
            return;

        XChangeProperty(__glfw.x11.display, window->x11.handle,
                        __glfw.x11.NET_WM_STATE, XA_ATOM, 32,
                        PropModeAppend,
                        (unsigned char*) missing,
                        missingCount);
    }

    XFlush(__glfw.x11.display);
}

void ___glfwShowWindowX11(_GLFWwindow* window)
{
    if (__glfwWindowVisibleX11(window))
        return;

    XMapWindow(__glfw.x11.display, window->x11.handle);
    waitForVisibilityNotify(window);
}

void ___glfwHideWindowX11(_GLFWwindow* window)
{
    XUnmapWindow(__glfw.x11.display, window->x11.handle);
    XFlush(__glfw.x11.display);
}

void ___glfwRequestWindowAttentionX11(_GLFWwindow* window)
{
    if (!__glfw.x11.NET_WM_STATE || !__glfw.x11.NET_WM_STATE_DEMANDS_ATTENTION)
        return;

    sendEventToWM(window,
                  __glfw.x11.NET_WM_STATE,
                  _NET_WM_STATE_ADD,
                  __glfw.x11.NET_WM_STATE_DEMANDS_ATTENTION,
                  0, 1, 0);
}

void ___glfwFocusWindowX11(_GLFWwindow* window)
{
    if (__glfw.x11.NET_ACTIVE_WINDOW)
        sendEventToWM(window, __glfw.x11.NET_ACTIVE_WINDOW, 1, 0, 0, 0, 0);
    else if (__glfwWindowVisibleX11(window))
    {
        XRaiseWindow(__glfw.x11.display, window->x11.handle);
        XSetInputFocus(__glfw.x11.display, window->x11.handle,
                       RevertToParent, CurrentTime);
    }

    XFlush(__glfw.x11.display);
}

void ___glfwSetWindowMonitorX11(_GLFWwindow* window,
                              _GLFWmonitor* monitor,
                              int xpos, int ypos,
                              int width, int height,
                              int refreshRate)
{
    if (window->monitor == monitor)
    {
        if (monitor)
        {
            if (monitor->windowww == window)
                acquireMonitor(window);
        }
        else
        {
            if (!window->resizable)
                updateNormalHints(window, width, height);

            XMoveResizeWindow(__glfw.x11.display, window->x11.handle,
                              xpos, ypos, width, height);
        }

        XFlush(__glfw.x11.display);
        return;
    }

    if (window->monitor)
    {
        __glfwSetWindowDecoratedX11(window, window->decorated);
        __glfwSetWindowFloatingX11(window, window->floating);
        releaseMonitor(window);
    }

    ___glfwInputWindowMonitor(window, monitor);
    updateNormalHints(window, width, height);

    if (window->monitor)
    {
        if (!__glfwWindowVisibleX11(window))
        {
            XMapRaised(__glfw.x11.display, window->x11.handle);
            waitForVisibilityNotify(window);
        }

        updateWindowMode(window);
        acquireMonitor(window);
    }
    else
    {
        updateWindowMode(window);
        XMoveResizeWindow(__glfw.x11.display, window->x11.handle,
                          xpos, ypos, width, height);
    }

    XFlush(__glfw.x11.display);
}

GLFWbool __glfwWindowFocusedX11(_GLFWwindow* window)
{
    Window focused;
    int state;

    XGetInputFocus(__glfw.x11.display, &focused, &state);
    return window->x11.handle == focused;
}

GLFWbool __glfwWindowIconifiedX11(_GLFWwindow* window)
{
    return getWindowState(window) == IconicState;
}

GLFWbool __glfwWindowVisibleX11(_GLFWwindow* window)
{
    XWindowAttributes wa;
    XGetWindowAttributes(__glfw.x11.display, window->x11.handle, &wa);
    return wa.map_state == IsViewable;
}

GLFWbool __glfwWindowMaximizedX11(_GLFWwindow* window)
{
    Atom* states;
    GLFWbool maximized = GLFW_FALSE;

    if (!__glfw.x11.NET_WM_STATE ||
        !__glfw.x11.NET_WM_STATE_MAXIMIZED_VERT ||
        !__glfw.x11.NET_WM_STATE_MAXIMIZED_HORZ)
    {
        return maximized;
    }

    const unsigned long count =
        ___glfwGetWindowPropertyX11(window->x11.handle,
                                  __glfw.x11.NET_WM_STATE,
                                  XA_ATOM,
                                  (unsigned char**) &states);

    for (unsigned long i = 0;  i < count;  i++)
    {
        if (states[i] == __glfw.x11.NET_WM_STATE_MAXIMIZED_VERT ||
            states[i] == __glfw.x11.NET_WM_STATE_MAXIMIZED_HORZ)
        {
            maximized = GLFW_TRUE;
            break;
        }
    }

    if (states)
        XFree(states);

    return maximized;
}

GLFWbool __glfwWindowHoveredX11(_GLFWwindow* window)
{
    Window w = __glfw.x11.root;
    while (w)
    {
        Window root;
        int rootX, rootY, childX, childY;
        unsigned int mask;

        ___glfwGrabErrorHandlerX11();

        const Bool result = XQueryPointer(__glfw.x11.display, w,
                                          &root, &w, &rootX, &rootY,
                                          &childX, &childY, &mask);

        ___glfwReleaseErrorHandlerX11();

        if (__glfw.x11.errorCode == BadWindow)
            w = __glfw.x11.root;
        else if (!result)
            return GLFW_FALSE;
        else if (w == window->x11.handle)
            return GLFW_TRUE;
    }

    return GLFW_FALSE;
}

GLFWbool __glfwFramebufferTransparentX11(_GLFWwindow* window)
{
    if (!window->x11.transparent)
        return GLFW_FALSE;

    return XGetSelectionOwner(__glfw.x11.display, __glfw.x11.NET_WM_CM_Sx) != None;
}

void __glfwSetWindowResizableX11(_GLFWwindow* window, GLFWbool enabled)
{
    int width, height;
    ___glfwGetWindowSizeX11(window, &width, &height);
    updateNormalHints(window, width, height);
}

void __glfwSetWindowDecoratedX11(_GLFWwindow* window, GLFWbool enabled)
{
    struct
    {
        unsigned long flags;
        unsigned long functions;
        unsigned long decorations;
        long input_mode;
        unsigned long status;
    } hints = {0};

    hints.flags = MWM_HINTS_DECORATIONS;
    hints.decorations = enabled ? MWM_DECOR_ALL : 0;

    XChangeProperty(__glfw.x11.display, window->x11.handle,
                    __glfw.x11.MOTIF_WM_HINTS,
                    __glfw.x11.MOTIF_WM_HINTS, 32,
                    PropModeReplace,
                    (unsigned char*) &hints,
                    sizeof(hints) / sizeof(long));
}

void __glfwSetWindowFloatingX11(_GLFWwindow* window, GLFWbool enabled)
{
    if (!__glfw.x11.NET_WM_STATE || !__glfw.x11.NET_WM_STATE_ABOVE)
        return;

    if (__glfwWindowVisibleX11(window))
    {
        const long action = enabled ? _NET_WM_STATE_ADD : _NET_WM_STATE_REMOVE;
        sendEventToWM(window,
                      __glfw.x11.NET_WM_STATE,
                      action,
                      __glfw.x11.NET_WM_STATE_ABOVE,
                      0, 1, 0);
    }
    else
    {
        Atom* states = NULL;
        const unsigned long count =
            ___glfwGetWindowPropertyX11(window->x11.handle,
                                      __glfw.x11.NET_WM_STATE,
                                      XA_ATOM,
                                      (unsigned char**) &states);

        // NOTE: We don't check for failure as this property may not exist yet
        //       and that's fine (and we'll create it implicitly with append)

        if (enabled)
        {
            unsigned long i;

            for (i = 0;  i < count;  i++)
            {
                if (states[i] == __glfw.x11.NET_WM_STATE_ABOVE)
                    break;
            }

            if (i == count)
            {
                XChangeProperty(__glfw.x11.display, window->x11.handle,
                                __glfw.x11.NET_WM_STATE, XA_ATOM, 32,
                                PropModeAppend,
                                (unsigned char*) &__glfw.x11.NET_WM_STATE_ABOVE,
                                1);
            }
        }
        else if (states)
        {
            for (unsigned long i = 0;  i < count;  i++)
            {
                if (states[i] == __glfw.x11.NET_WM_STATE_ABOVE)
                {
                    states[i] = states[count - 1];
                    XChangeProperty(__glfw.x11.display, window->x11.handle,
                                    __glfw.x11.NET_WM_STATE, XA_ATOM, 32,
                                    PropModeReplace, (unsigned char*) states, count - 1);
                    break;
                }
            }
        }

        if (states)
            XFree(states);
    }

    XFlush(__glfw.x11.display);
}

void __glfwSetWindowMousePassthroughX11(_GLFWwindow* window, GLFWbool enabled)
{
    if (!__glfw.x11.xshape.available)
        return;

    if (enabled)
    {
        Region region = XCreateRegion();
        XShapeCombineRegion(__glfw.x11.display, window->x11.handle,
                            ShapeInput, 0, 0, region, ShapeSet);
        XDestroyRegion(region);
    }
    else
    {
        XShapeCombineMask(__glfw.x11.display, window->x11.handle,
                          ShapeInput, 0, 0, None, ShapeSet);
    }
}

float ___glfwGetWindowOpacityX11(_GLFWwindow* window)
{
    float opacity = 1.f;

    if (XGetSelectionOwner(__glfw.x11.display, __glfw.x11.NET_WM_CM_Sx))
    {
        CARD32* value = NULL;

        if (___glfwGetWindowPropertyX11(window->x11.handle,
                                      __glfw.x11.NET_WM_WINDOW_OPACITY,
                                      XA_CARDINAL,
                                      (unsigned char**) &value))
        {
            opacity = (float) (*value / (double) 0xffffffffu);
        }

        if (value)
            XFree(value);
    }

    return opacity;
}

void ___glfwSetWindowOpacityX11(_GLFWwindow* window, float opacity)
{
    const CARD32 value = (CARD32) (0xffffffffu * (double) opacity);
    XChangeProperty(__glfw.x11.display, window->x11.handle,
                    __glfw.x11.NET_WM_WINDOW_OPACITY, XA_CARDINAL, 32,
                    PropModeReplace, (unsigned char*) &value, 1);
}

void __glfwSetRawMouseMotionX11(_GLFWwindow *window, GLFWbool enabled)
{
    if (!__glfw.x11.xi.available)
        return;

    if (__glfw.x11.disabledCursorWindow != window)
        return;

    if (enabled)
        enableRawMouseMotion(window);
    else
        disableRawMouseMotion(window);
}

GLFWbool ___glfwRawMouseMotionSupportedX11(void)
{
    return __glfw.x11.xi.available;
}

void ___glfwPollEventsX11(void)
{
    drainEmptyEvents();

#if defined(__linux__)
    if (__glfw.joysticksInitialized)
        ___glfwDetectJoystickConnectionLinux();
#endif
    XPending(__glfw.x11.display);

    while (QLength(__glfw.x11.display))
    {
        XEvent event;
        XNextEvent(__glfw.x11.display, &event);
        processEvent(&event);
    }

    _GLFWwindow* window = __glfw.x11.disabledCursorWindow;
    if (window)
    {
        int width, height;
        ___glfwGetWindowSizeX11(window, &width, &height);

        // NOTE: Re-center the cursor only if it has moved since the last call,
        //       to avoid breaking __glfwWaitEvents with MotionNotify
        if (window->x11.lastCursorPosX != width / 2 ||
            window->x11.lastCursorPosY != height / 2)
        {
            ____glfwSetCursorPosX11(window, width / 2, height / 2);
        }
    }

    XFlush(__glfw.x11.display);
}

void ___glfwWaitEventsX11(void)
{
    waitForAnyEvent(NULL);
    ___glfwPollEventsX11();
}

void ____glfwWaitEventsTimeoutX11(double timeout)
{
    waitForAnyEvent(&timeout);
    ___glfwPollEventsX11();
}

void ___glfwPostEmptyEventX11(void)
{
    writeEmptyEvent();
}

void ___glfwGetCursorPosX11(_GLFWwindow* window, double* xpos, double* ypos)
{
    Window root, child;
    int rootX, rootY, childX, childY;
    unsigned int mask;

    XQueryPointer(__glfw.x11.display, window->x11.handle,
                  &root, &child,
                  &rootX, &rootY, &childX, &childY,
                  &mask);

    if (xpos)
        *xpos = childX;
    if (ypos)
        *ypos = childY;
}

void ____glfwSetCursorPosX11(_GLFWwindow* window, double x, double y)
{
    // Store the new position so it can be recognized later
    window->x11.warpCursorPosX = (int) x;
    window->x11.warpCursorPosY = (int) y;

    XWarpPointer(__glfw.x11.display, None, window->x11.handle,
                 0,0,0,0, (int) x, (int) y);
    XFlush(__glfw.x11.display);
}

void ___glfwSetCursorModeX11(_GLFWwindow* window, int mode)
{
    if (__glfwWindowFocusedX11(window))
    {
        if (mode == GLFW_CURSOR_DISABLED)
        {
            ___glfwGetCursorPosX11(window,
                                 &__glfw.x11.restoreCursorPosX,
                                 &__glfw.x11.restoreCursorPosY);
            ___glfwCenterCursorInContentArea(window);
            if (window->rawMouseMotion)
                enableRawMouseMotion(window);
        }
        else if (__glfw.x11.disabledCursorWindow == window)
        {
            if (window->rawMouseMotion)
                disableRawMouseMotion(window);
        }

        if (mode == GLFW_CURSOR_DISABLED || mode == GLFW_CURSOR_CAPTURED)
            captureCursor(window);
        else
            releaseCursor();

        if (mode == GLFW_CURSOR_DISABLED)
            __glfw.x11.disabledCursorWindow = window;
        else if (__glfw.x11.disabledCursorWindow == window)
        {
            __glfw.x11.disabledCursorWindow = NULL;
            ____glfwSetCursorPosX11(window,
                                 __glfw.x11.restoreCursorPosX,
                                 __glfw.x11.restoreCursorPosY);
        }
    }

    updateCursorImage(window);
    XFlush(__glfw.x11.display);
}

const char* __glfwGetScancodeNameX11(int scancode)
{
    if (!__glfw.x11.xkb.available)
        return NULL;

    if (scancode < 0 || scancode > 0xff ||
        __glfw.x11.keycodes[scancode] == GLFW_KEY_UNKNOWN)
    {
        ___glfwInputError(GLFW_INVALID_VALUE, "Invalid scancode %i", scancode);
        return NULL;
    }

    const int key = __glfw.x11.keycodes[scancode];
    const KeySym keysym = XkbKeycodeToKeysym(__glfw.x11.display,
                                             scancode, __glfw.x11.xkb.group, 0);
    if (keysym == NoSymbol)
        return NULL;

    const uint32_t codepoint = ___glfwKeySym2Unicode(keysym);
    if (codepoint == GLFW_INVALID_CODEPOINT)
        return NULL;

    const size_t count = ___glfwEncodeUTF8(__glfw.x11.keynames[key], codepoint);
    if (count == 0)
        return NULL;

    __glfw.x11.keynames[key][count] = '\0';
    return __glfw.x11.keynames[key];
}

int ____glfwGetKeyScancodeX11(int key)
{
    return __glfw.x11.scancodes[key];
}

GLFWbool ____glfwCreateCursorX11(_GLFWcursor* cursor,
                              const GLFWimage* image,
                              int xhot, int yhot)
{
    cursor->x11.handle = __glfwCreateNativeCursorX11(image, xhot, yhot);
    if (!cursor->x11.handle)
        return GLFW_FALSE;

    return GLFW_TRUE;
}

GLFWbool ___glfwCreateStandardCursorX11(_GLFWcursor* cursor, int shape)
{
    if (__glfw.x11.xcursor.handle)
    {
        char* theme = XcursorGetTheme(__glfw.x11.display);
        if (theme)
        {
            const int size = XcursorGetDefaultSize(__glfw.x11.display);
            const char* name = NULL;

            switch (shape)
            {
                case GLFW_ARROW_CURSOR:
                    name = "default";
                    break;
                case GLFW_IBEAM_CURSOR:
                    name = "text";
                    break;
                case GLFW_CROSSHAIR_CURSOR:
                    name = "crosshair";
                    break;
                case GLFW_POINTING_HAND_CURSOR:
                    name = "pointer";
                    break;
                case GLFW_RESIZE_EW_CURSOR:
                    name = "ew-resize";
                    break;
                case GLFW_RESIZE_NS_CURSOR:
                    name = "ns-resize";
                    break;
                case GLFW_RESIZE_NWSE_CURSOR:
                    name = "nwse-resize";
                    break;
                case GLFW_RESIZE_NESW_CURSOR:
                    name = "nesw-resize";
                    break;
                case GLFW_RESIZE_ALL_CURSOR:
                    name = "all-scroll";
                    break;
                case GLFW_NOT_ALLOWED_CURSOR:
                    name = "not-allowed";
                    break;
            }

            XcursorImage* image = XcursorLibraryLoadImage(name, theme, size);
            if (image)
            {
                cursor->x11.handle = XcursorImageLoadCursor(__glfw.x11.display, image);
                XcursorImageDestroy(image);
            }
        }
    }

    if (!cursor->x11.handle)
    {
        unsigned int native = 0;

        switch (shape)
        {
            case GLFW_ARROW_CURSOR:
                native = XC_left_ptr;
                break;
            case GLFW_IBEAM_CURSOR:
                native = XC_xterm;
                break;
            case GLFW_CROSSHAIR_CURSOR:
                native = XC_crosshair;
                break;
            case GLFW_POINTING_HAND_CURSOR:
                native = XC_hand2;
                break;
            case GLFW_RESIZE_EW_CURSOR:
                native = XC_sb_h_double_arrow;
                break;
            case GLFW_RESIZE_NS_CURSOR:
                native = XC_sb_v_double_arrow;
                break;
            case GLFW_RESIZE_ALL_CURSOR:
                native = XC_fleur;
                break;
            default:
                ___glfwInputError(GLFW_CURSOR_UNAVAILABLE,
                                "X11: Standard cursor shape unavailable");
                return GLFW_FALSE;
        }

        cursor->x11.handle = XCreateFontCursor(__glfw.x11.display, native);
        if (!cursor->x11.handle)
        {
            ___glfwInputError(GLFW_PLATFORM_ERROR,
                            "X11: Failed to create standard cursor");
            return GLFW_FALSE;
        }
    }

    return GLFW_TRUE;
}

void ___glfwDestroyCursorX11(_GLFWcursor* cursor)
{
    if (cursor->x11.handle)
        XFreeCursor(__glfw.x11.display, cursor->x11.handle);
}

void ___glfwSetCursorX11(_GLFWwindow* window, _GLFWcursor* cursor)
{
    if (window->cursorMode == GLFW_CURSOR_NORMAL ||
        window->cursorMode == GLFW_CURSOR_CAPTURED)
    {
        updateCursorImage(window);
        XFlush(__glfw.x11.display);
    }
}

void ___glfwSetClipboardStringX11(const char* string)
{
    char* copy = ___glfw_strdup(string);
    __glfw_free(__glfw.x11.clipboardString);
    __glfw.x11.clipboardString = copy;

    XSetSelectionOwner(__glfw.x11.display,
                       __glfw.x11.CLIPBOARD,
                       __glfw.x11.helperWindowHandle,
                       CurrentTime);

    if (XGetSelectionOwner(__glfw.x11.display, __glfw.x11.CLIPBOARD) !=
        __glfw.x11.helperWindowHandle)
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "X11: Failed to become owner of clipboard selection");
    }
}

const char* ___glfwGetClipboardStringX11(void)
{
    return getSelectionString(__glfw.x11.CLIPBOARD);
}

EGLenum __glfwGetEGLPlatformX11(EGLint** attribs)
{
    if (__glfw.egl.ANGLE_platform_angle)
    {
        int type = 0;

        if (__glfw.egl.ANGLE_platform_angle_opengl)
        {
            if (__glfw.hints.init.angleType == GLFW_ANGLE_PLATFORM_TYPE_OPENGL)
                type = EGL_PLATFORM_ANGLE_TYPE_OPENGL_ANGLE;
        }

        if (__glfw.egl.ANGLE_platform_angle_vulkan)
        {
            if (__glfw.hints.init.angleType == GLFW_ANGLE_PLATFORM_TYPE_VULKAN)
                type = EGL_PLATFORM_ANGLE_TYPE_VULKAN_ANGLE;
        }

        if (type)
        {
            *attribs = __glfw_calloc(5, sizeof(EGLint));
            (*attribs)[0] = EGL_PLATFORM_ANGLE_TYPE_ANGLE;
            (*attribs)[1] = type;
            (*attribs)[2] = EGL_PLATFORM_ANGLE_NATIVE_PLATFORM_TYPE_ANGLE;
            (*attribs)[3] = EGL_PLATFORM_X11_EXT;
            (*attribs)[4] = EGL_NONE;
            return EGL_PLATFORM_ANGLE_ANGLE;
        }
    }

    if (__glfw.egl.EXT_platform_base && __glfw.egl.EXT_platform_x11)
        return EGL_PLATFORM_X11_EXT;

    return 0;
}

EGLNativeDisplayType __glfwGetEGLNativeDisplayX11(void)
{
    return __glfw.x11.display;
}

EGLNativeWindowType __glfwGetEGLNativeWindowX11(_GLFWwindow* window)
{
    if (__glfw.egl.platform)
        return &window->x11.handle;
    else
        return (EGLNativeWindowType) window->x11.handle;
}

void ___glfwGetRequiredInstanceExtensionsX11(char** extensions)
{
    if (!__glfw.vk.KHR_surface)
        return;

    if (!__glfw.vk.KHR_xcb_surface || !__glfw.x11.x11xcb.handle)
    {
        if (!__glfw.vk.KHR_xlib_surface)
            return;
    }

    extensions[0] = "VK_KHR_surface";

    // NOTE: VK_KHR_xcb_surface is preferred due to some early ICDs exposing but
    //       not correctly implementing VK_KHR_xlib_surface
    if (__glfw.vk.KHR_xcb_surface && __glfw.x11.x11xcb.handle)
        extensions[1] = "VK_KHR_xcb_surface";
    else
        extensions[1] = "VK_KHR_xlib_surface";
}

GLFWbool ___glfwGetPhysicalDevicePresentationSupportX11(VkInstance instance,
                                                      VkPhysicalDevice device,
                                                      uint32_t queuefamily)
{
    VisualID visualID = XVisualIDFromVisual(DefaultVisual(__glfw.x11.display,
                                                          __glfw.x11.screen));

    if (__glfw.vk.KHR_xcb_surface && __glfw.x11.x11xcb.handle)
    {
        PFN_vkGetPhysicalDeviceXcbPresentationSupportKHR
            vkGetPhysicalDeviceXcbPresentationSupportKHR =
            (PFN_vkGetPhysicalDeviceXcbPresentationSupportKHR)
            vkGetInstanceProcAddr(instance, "vkGetPhysicalDeviceXcbPresentationSupportKHR");
        if (!vkGetPhysicalDeviceXcbPresentationSupportKHR)
        {
            ___glfwInputError(GLFW_API_UNAVAILABLE,
                            "X11: Vulkan instance missing VK_KHR_xcb_surface extension");
            return GLFW_FALSE;
        }

        xcb_connection_t* connection = XGetXCBConnection(__glfw.x11.display);
        if (!connection)
        {
            ___glfwInputError(GLFW_PLATFORM_ERROR,
                            "X11: Failed to retrieve XCB connection");
            return GLFW_FALSE;
        }

        return vkGetPhysicalDeviceXcbPresentationSupportKHR(device,
                                                            queuefamily,
                                                            connection,
                                                            visualID);
    }
    else
    {
        PFN_vkGetPhysicalDeviceXlibPresentationSupportKHR
            vkGetPhysicalDeviceXlibPresentationSupportKHR =
            (PFN_vkGetPhysicalDeviceXlibPresentationSupportKHR)
            vkGetInstanceProcAddr(instance, "vkGetPhysicalDeviceXlibPresentationSupportKHR");
        if (!vkGetPhysicalDeviceXlibPresentationSupportKHR)
        {
            ___glfwInputError(GLFW_API_UNAVAILABLE,
                            "X11: Vulkan instance missing VK_KHR_xlib_surface extension");
            return GLFW_FALSE;
        }

        return vkGetPhysicalDeviceXlibPresentationSupportKHR(device,
                                                             queuefamily,
                                                             __glfw.x11.display,
                                                             visualID);
    }
}

VkResult ____glfwCreateWindowSurfaceX11(VkInstance instance,
                                     _GLFWwindow* window,
                                     const VkAllocationCallbacks* allocator,
                                     VkSurfaceKHR* surface)
{
    if (__glfw.vk.KHR_xcb_surface && __glfw.x11.x11xcb.handle)
    {
        VkResult err;
        VkXcbSurfaceCreateInfoKHR sci;
        PFN_vkCreateXcbSurfaceKHR vkCreateXcbSurfaceKHR;

        xcb_connection_t* connection = XGetXCBConnection(__glfw.x11.display);
        if (!connection)
        {
            ___glfwInputError(GLFW_PLATFORM_ERROR,
                            "X11: Failed to retrieve XCB connection");
            return VK_ERROR_EXTENSION_NOT_PRESENT;
        }

        vkCreateXcbSurfaceKHR = (PFN_vkCreateXcbSurfaceKHR)
            vkGetInstanceProcAddr(instance, "vkCreateXcbSurfaceKHR");
        if (!vkCreateXcbSurfaceKHR)
        {
            ___glfwInputError(GLFW_API_UNAVAILABLE,
                            "X11: Vulkan instance missing VK_KHR_xcb_surface extension");
            return VK_ERROR_EXTENSION_NOT_PRESENT;
        }

        memset(&sci, 0, sizeof(sci));
        sci.sType = VK_STRUCTURE_TYPE_XCB_SURFACE_CREATE_INFO_KHR;
        sci.connection = connection;
        sci.window = window->x11.handle;

        err = vkCreateXcbSurfaceKHR(instance, &sci, allocator, surface);
        if (err)
        {
            ___glfwInputError(GLFW_PLATFORM_ERROR,
                            "X11: Failed to create Vulkan XCB surface: %s",
                            ___glfwGetVulkanResultString(err));
        }

        return err;
    }
    else
    {
        VkResult err;
        VkXlibSurfaceCreateInfoKHR sci;
        PFN_vkCreateXlibSurfaceKHR vkCreateXlibSurfaceKHR;

        vkCreateXlibSurfaceKHR = (PFN_vkCreateXlibSurfaceKHR)
            vkGetInstanceProcAddr(instance, "vkCreateXlibSurfaceKHR");
        if (!vkCreateXlibSurfaceKHR)
        {
            ___glfwInputError(GLFW_API_UNAVAILABLE,
                            "X11: Vulkan instance missing VK_KHR_xlib_surface extension");
            return VK_ERROR_EXTENSION_NOT_PRESENT;
        }

        memset(&sci, 0, sizeof(sci));
        sci.sType = VK_STRUCTURE_TYPE_XLIB_SURFACE_CREATE_INFO_KHR;
        sci.dpy = __glfw.x11.display;
        sci.window = window->x11.handle;

        err = vkCreateXlibSurfaceKHR(instance, &sci, allocator, surface);
        if (err)
        {
            ___glfwInputError(GLFW_PLATFORM_ERROR,
                            "X11: Failed to create Vulkan X11 surface: %s",
                            ___glfwGetVulkanResultString(err));
        }

        return err;
    }
}


//////////////////////////////////////////////////////////////////////////
//////                        GLFW native API                       //////
//////////////////////////////////////////////////////////////////////////

GLFWAPI Display* __glfwGetX11Display(void)
{
    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);

    if (__glfw.platform.platformID != GLFW_PLATFORM_X11)
    {
        ___glfwInputError(GLFW_PLATFORM_UNAVAILABLE, "X11: Platform not initialized");
        return NULL;
    }

    return __glfw.x11.display;
}

GLFWAPI Window __glfwGetX11Window(GLFWwindow* handle)
{
    _GLFWwindow* window = (_GLFWwindow*) handle;
    _GLFW_REQUIRE_INIT_OR_RETURN(None);

    if (__glfw.platform.platformID != GLFW_PLATFORM_X11)
    {
        ___glfwInputError(GLFW_PLATFORM_UNAVAILABLE, "X11: Platform not initialized");
        return None;
    }

    return window->x11.handle;
}

GLFWAPI void __glfwSetX11SelectionString(const char* string)
{
    _GLFW_REQUIRE_INIT();

    if (__glfw.platform.platformID != GLFW_PLATFORM_X11)
    {
        ___glfwInputError(GLFW_PLATFORM_UNAVAILABLE, "X11: Platform not initialized");
        return;
    }

    __glfw_free(__glfw.x11.primarySelectionString);
    __glfw.x11.primarySelectionString = ___glfw_strdup(string);

    XSetSelectionOwner(__glfw.x11.display,
                       __glfw.x11.PRIMARY,
                       __glfw.x11.helperWindowHandle,
                       CurrentTime);

    if (XGetSelectionOwner(__glfw.x11.display, __glfw.x11.PRIMARY) !=
        __glfw.x11.helperWindowHandle)
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "X11: Failed to become owner of primary selection");
    }
}

GLFWAPI const char* __glfwGetX11SelectionString(void)
{
    _GLFW_REQUIRE_INIT_OR_RETURN(NULL);

    if (__glfw.platform.platformID != GLFW_PLATFORM_X11)
    {
        ___glfwInputError(GLFW_PLATFORM_UNAVAILABLE, "X11: Platform not initialized");
        return NULL;
    }

    return getSelectionString(__glfw.x11.PRIMARY);
}

