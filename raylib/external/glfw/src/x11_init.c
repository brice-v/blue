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

#include <stdlib.h>
#include <string.h>
#include <limits.h>
#include <stdio.h>
#include <locale.h>
#include <unistd.h>
#include <fcntl.h>
#include <errno.h>
#include <assert.h>


// Translate the X11 KeySyms for a key to a GLFW key code
// NOTE: This is only used as a fallback, in case the XKB method fails
//       It is layout-dependent and will fail partially on most non-US layouts
//
static int translateKeySyms(const KeySym* keysyms, int width)
{
    if (width > 1)
    {
        switch (keysyms[1])
        {
            case XK_KP_0:           return GLFW_KEY_KP_0;
            case XK_KP_1:           return GLFW_KEY_KP_1;
            case XK_KP_2:           return GLFW_KEY_KP_2;
            case XK_KP_3:           return GLFW_KEY_KP_3;
            case XK_KP_4:           return GLFW_KEY_KP_4;
            case XK_KP_5:           return GLFW_KEY_KP_5;
            case XK_KP_6:           return GLFW_KEY_KP_6;
            case XK_KP_7:           return GLFW_KEY_KP_7;
            case XK_KP_8:           return GLFW_KEY_KP_8;
            case XK_KP_9:           return GLFW_KEY_KP_9;
            case XK_KP_Separator:
            case XK_KP_Decimal:     return GLFW_KEY_KP_DECIMAL;
            case XK_KP_Equal:       return GLFW_KEY_KP_EQUAL;
            case XK_KP_Enter:       return GLFW_KEY_KP_ENTER;
            default:                break;
        }
    }

    switch (keysyms[0])
    {
        case XK_Escape:         return GLFW_KEY_ESCAPE;
        case XK_Tab:            return GLFW_KEY_TAB;
        case XK_Shift_L:        return GLFW_KEY_LEFT_SHIFT;
        case XK_Shift_R:        return GLFW_KEY_RIGHT_SHIFT;
        case XK_Control_L:      return GLFW_KEY_LEFT_CONTROL;
        case XK_Control_R:      return GLFW_KEY_RIGHT_CONTROL;
        case XK_Meta_L:
        case XK_Alt_L:          return GLFW_KEY_LEFT_ALT;
        case XK_Mode_switch: // Mapped to Alt_R on many keyboards
        case XK_ISO_Level3_Shift: // AltGr on at least some machines
        case XK_Meta_R:
        case XK_Alt_R:          return GLFW_KEY_RIGHT_ALT;
        case XK_Super_L:        return GLFW_KEY_LEFT_SUPER;
        case XK_Super_R:        return GLFW_KEY_RIGHT_SUPER;
        case XK_Menu:           return GLFW_KEY_MENU;
        case XK_Num_Lock:       return GLFW_KEY_NUM_LOCK;
        case XK_Caps_Lock:      return GLFW_KEY_CAPS_LOCK;
        case XK_Print:          return GLFW_KEY_PRINT_SCREEN;
        case XK_Scroll_Lock:    return GLFW_KEY_SCROLL_LOCK;
        case XK_Pause:          return GLFW_KEY_PAUSE;
        case XK_Delete:         return GLFW_KEY_DELETE;
        case XK_BackSpace:      return GLFW_KEY_BACKSPACE;
        case XK_Return:         return GLFW_KEY_ENTER;
        case XK_Home:           return GLFW_KEY_HOME;
        case XK_End:            return GLFW_KEY_END;
        case XK_Page_Up:        return GLFW_KEY_PAGE_UP;
        case XK_Page_Down:      return GLFW_KEY_PAGE_DOWN;
        case XK_Insert:         return GLFW_KEY_INSERT;
        case XK_Left:           return GLFW_KEY_LEFT;
        case XK_Right:          return GLFW_KEY_RIGHT;
        case XK_Down:           return GLFW_KEY_DOWN;
        case XK_Up:             return GLFW_KEY_UP;
        case XK_F1:             return GLFW_KEY_F1;
        case XK_F2:             return GLFW_KEY_F2;
        case XK_F3:             return GLFW_KEY_F3;
        case XK_F4:             return GLFW_KEY_F4;
        case XK_F5:             return GLFW_KEY_F5;
        case XK_F6:             return GLFW_KEY_F6;
        case XK_F7:             return GLFW_KEY_F7;
        case XK_F8:             return GLFW_KEY_F8;
        case XK_F9:             return GLFW_KEY_F9;
        case XK_F10:            return GLFW_KEY_F10;
        case XK_F11:            return GLFW_KEY_F11;
        case XK_F12:            return GLFW_KEY_F12;
        case XK_F13:            return GLFW_KEY_F13;
        case XK_F14:            return GLFW_KEY_F14;
        case XK_F15:            return GLFW_KEY_F15;
        case XK_F16:            return GLFW_KEY_F16;
        case XK_F17:            return GLFW_KEY_F17;
        case XK_F18:            return GLFW_KEY_F18;
        case XK_F19:            return GLFW_KEY_F19;
        case XK_F20:            return GLFW_KEY_F20;
        case XK_F21:            return GLFW_KEY_F21;
        case XK_F22:            return GLFW_KEY_F22;
        case XK_F23:            return GLFW_KEY_F23;
        case XK_F24:            return GLFW_KEY_F24;
        case XK_F25:            return GLFW_KEY_F25;

        // Numeric keypad
        case XK_KP_Divide:      return GLFW_KEY_KP_DIVIDE;
        case XK_KP_Multiply:    return GLFW_KEY_KP_MULTIPLY;
        case XK_KP_Subtract:    return GLFW_KEY_KP_SUBTRACT;
        case XK_KP_Add:         return GLFW_KEY_KP_ADD;

        // These should have been detected in secondary keysym test above!
        case XK_KP_Insert:      return GLFW_KEY_KP_0;
        case XK_KP_End:         return GLFW_KEY_KP_1;
        case XK_KP_Down:        return GLFW_KEY_KP_2;
        case XK_KP_Page_Down:   return GLFW_KEY_KP_3;
        case XK_KP_Left:        return GLFW_KEY_KP_4;
        case XK_KP_Right:       return GLFW_KEY_KP_6;
        case XK_KP_Home:        return GLFW_KEY_KP_7;
        case XK_KP_Up:          return GLFW_KEY_KP_8;
        case XK_KP_Page_Up:     return GLFW_KEY_KP_9;
        case XK_KP_Delete:      return GLFW_KEY_KP_DECIMAL;
        case XK_KP_Equal:       return GLFW_KEY_KP_EQUAL;
        case XK_KP_Enter:       return GLFW_KEY_KP_ENTER;

        // Last resort: Check for printable keys (should not happen if the XKB
        // extension is available). This will give a layout dependent mapping
        // (which is wrong, and we may miss some keys, especially on non-US
        // keyboards), but it's better than nothing...
        case XK_a:              return GLFW_KEY_A;
        case XK_b:              return GLFW_KEY_B;
        case XK_c:              return GLFW_KEY_C;
        case XK_d:              return GLFW_KEY_D;
        case XK_e:              return GLFW_KEY_E;
        case XK_f:              return GLFW_KEY_F;
        case XK_g:              return GLFW_KEY_G;
        case XK_h:              return GLFW_KEY_H;
        case XK_i:              return GLFW_KEY_I;
        case XK_j:              return GLFW_KEY_J;
        case XK_k:              return GLFW_KEY_K;
        case XK_l:              return GLFW_KEY_L;
        case XK_m:              return GLFW_KEY_M;
        case XK_n:              return GLFW_KEY_N;
        case XK_o:              return GLFW_KEY_O;
        case XK_p:              return GLFW_KEY_P;
        case XK_q:              return GLFW_KEY_Q;
        case XK_r:              return GLFW_KEY_R;
        case XK_s:              return GLFW_KEY_S;
        case XK_t:              return GLFW_KEY_T;
        case XK_u:              return GLFW_KEY_U;
        case XK_v:              return GLFW_KEY_V;
        case XK_w:              return GLFW_KEY_W;
        case XK_x:              return GLFW_KEY_X;
        case XK_y:              return GLFW_KEY_Y;
        case XK_z:              return GLFW_KEY_Z;
        case XK_1:              return GLFW_KEY_1;
        case XK_2:              return GLFW_KEY_2;
        case XK_3:              return GLFW_KEY_3;
        case XK_4:              return GLFW_KEY_4;
        case XK_5:              return GLFW_KEY_5;
        case XK_6:              return GLFW_KEY_6;
        case XK_7:              return GLFW_KEY_7;
        case XK_8:              return GLFW_KEY_8;
        case XK_9:              return GLFW_KEY_9;
        case XK_0:              return GLFW_KEY_0;
        case XK_space:          return GLFW_KEY_SPACE;
        case XK_minus:          return GLFW_KEY_MINUS;
        case XK_equal:          return GLFW_KEY_EQUAL;
        case XK_bracketleft:    return GLFW_KEY_LEFT_BRACKET;
        case XK_bracketright:   return GLFW_KEY_RIGHT_BRACKET;
        case XK_backslash:      return GLFW_KEY_BACKSLASH;
        case XK_semicolon:      return GLFW_KEY_SEMICOLON;
        case XK_apostrophe:     return GLFW_KEY_APOSTROPHE;
        case XK_grave:          return GLFW_KEY_GRAVE_ACCENT;
        case XK_comma:          return GLFW_KEY_COMMA;
        case XK_period:         return GLFW_KEY_PERIOD;
        case XK_slash:          return GLFW_KEY_SLASH;
        case XK_less:           return GLFW_KEY_WORLD_1; // At least in some layouts...
        default:                break;
    }

    // No matching translation was found
    return GLFW_KEY_UNKNOWN;
}

// Create key code translation tables
//
static void createKeyTables(void)
{
    int scancodeMin, scancodeMax;

    memset(__glfw.x11.keycodes, -1, sizeof(__glfw.x11.keycodes));
    memset(__glfw.x11.scancodes, -1, sizeof(__glfw.x11.scancodes));

    if (__glfw.x11.xkb.available)
    {
        // Use XKB to determine physical key locations independently of the
        // current keyboard layout

        XkbDescPtr desc = XkbGetMap(__glfw.x11.display, 0, XkbUseCoreKbd);
        XkbGetNames(__glfw.x11.display, XkbKeyNamesMask | XkbKeyAliasesMask, desc);

        scancodeMin = desc->min_key_code;
        scancodeMax = desc->max_key_code;

        const struct
        {
            int key;
            char* name;
        } keymap[] =
        {
            { GLFW_KEY_GRAVE_ACCENT, "TLDE" },
            { GLFW_KEY_1, "AE01" },
            { GLFW_KEY_2, "AE02" },
            { GLFW_KEY_3, "AE03" },
            { GLFW_KEY_4, "AE04" },
            { GLFW_KEY_5, "AE05" },
            { GLFW_KEY_6, "AE06" },
            { GLFW_KEY_7, "AE07" },
            { GLFW_KEY_8, "AE08" },
            { GLFW_KEY_9, "AE09" },
            { GLFW_KEY_0, "AE10" },
            { GLFW_KEY_MINUS, "AE11" },
            { GLFW_KEY_EQUAL, "AE12" },
            { GLFW_KEY_Q, "AD01" },
            { GLFW_KEY_W, "AD02" },
            { GLFW_KEY_E, "AD03" },
            { GLFW_KEY_R, "AD04" },
            { GLFW_KEY_T, "AD05" },
            { GLFW_KEY_Y, "AD06" },
            { GLFW_KEY_U, "AD07" },
            { GLFW_KEY_I, "AD08" },
            { GLFW_KEY_O, "AD09" },
            { GLFW_KEY_P, "AD10" },
            { GLFW_KEY_LEFT_BRACKET, "AD11" },
            { GLFW_KEY_RIGHT_BRACKET, "AD12" },
            { GLFW_KEY_A, "AC01" },
            { GLFW_KEY_S, "AC02" },
            { GLFW_KEY_D, "AC03" },
            { GLFW_KEY_F, "AC04" },
            { GLFW_KEY_G, "AC05" },
            { GLFW_KEY_H, "AC06" },
            { GLFW_KEY_J, "AC07" },
            { GLFW_KEY_K, "AC08" },
            { GLFW_KEY_L, "AC09" },
            { GLFW_KEY_SEMICOLON, "AC10" },
            { GLFW_KEY_APOSTROPHE, "AC11" },
            { GLFW_KEY_Z, "AB01" },
            { GLFW_KEY_X, "AB02" },
            { GLFW_KEY_C, "AB03" },
            { GLFW_KEY_V, "AB04" },
            { GLFW_KEY_B, "AB05" },
            { GLFW_KEY_N, "AB06" },
            { GLFW_KEY_M, "AB07" },
            { GLFW_KEY_COMMA, "AB08" },
            { GLFW_KEY_PERIOD, "AB09" },
            { GLFW_KEY_SLASH, "AB10" },
            { GLFW_KEY_BACKSLASH, "BKSL" },
            { GLFW_KEY_WORLD_1, "LSGT" },
            { GLFW_KEY_SPACE, "SPCE" },
            { GLFW_KEY_ESCAPE, "ESC" },
            { GLFW_KEY_ENTER, "RTRN" },
            { GLFW_KEY_TAB, "TAB" },
            { GLFW_KEY_BACKSPACE, "BKSP" },
            { GLFW_KEY_INSERT, "INS" },
            { GLFW_KEY_DELETE, "DELE" },
            { GLFW_KEY_RIGHT, "RGHT" },
            { GLFW_KEY_LEFT, "LEFT" },
            { GLFW_KEY_DOWN, "DOWN" },
            { GLFW_KEY_UP, "UP" },
            { GLFW_KEY_PAGE_UP, "PGUP" },
            { GLFW_KEY_PAGE_DOWN, "PGDN" },
            { GLFW_KEY_HOME, "HOME" },
            { GLFW_KEY_END, "END" },
            { GLFW_KEY_CAPS_LOCK, "CAPS" },
            { GLFW_KEY_SCROLL_LOCK, "SCLK" },
            { GLFW_KEY_NUM_LOCK, "NMLK" },
            { GLFW_KEY_PRINT_SCREEN, "PRSC" },
            { GLFW_KEY_PAUSE, "PAUS" },
            { GLFW_KEY_F1, "FK01" },
            { GLFW_KEY_F2, "FK02" },
            { GLFW_KEY_F3, "FK03" },
            { GLFW_KEY_F4, "FK04" },
            { GLFW_KEY_F5, "FK05" },
            { GLFW_KEY_F6, "FK06" },
            { GLFW_KEY_F7, "FK07" },
            { GLFW_KEY_F8, "FK08" },
            { GLFW_KEY_F9, "FK09" },
            { GLFW_KEY_F10, "FK10" },
            { GLFW_KEY_F11, "FK11" },
            { GLFW_KEY_F12, "FK12" },
            { GLFW_KEY_F13, "FK13" },
            { GLFW_KEY_F14, "FK14" },
            { GLFW_KEY_F15, "FK15" },
            { GLFW_KEY_F16, "FK16" },
            { GLFW_KEY_F17, "FK17" },
            { GLFW_KEY_F18, "FK18" },
            { GLFW_KEY_F19, "FK19" },
            { GLFW_KEY_F20, "FK20" },
            { GLFW_KEY_F21, "FK21" },
            { GLFW_KEY_F22, "FK22" },
            { GLFW_KEY_F23, "FK23" },
            { GLFW_KEY_F24, "FK24" },
            { GLFW_KEY_F25, "FK25" },
            { GLFW_KEY_KP_0, "KP0" },
            { GLFW_KEY_KP_1, "KP1" },
            { GLFW_KEY_KP_2, "KP2" },
            { GLFW_KEY_KP_3, "KP3" },
            { GLFW_KEY_KP_4, "KP4" },
            { GLFW_KEY_KP_5, "KP5" },
            { GLFW_KEY_KP_6, "KP6" },
            { GLFW_KEY_KP_7, "KP7" },
            { GLFW_KEY_KP_8, "KP8" },
            { GLFW_KEY_KP_9, "KP9" },
            { GLFW_KEY_KP_DECIMAL, "KPDL" },
            { GLFW_KEY_KP_DIVIDE, "KPDV" },
            { GLFW_KEY_KP_MULTIPLY, "KPMU" },
            { GLFW_KEY_KP_SUBTRACT, "KPSU" },
            { GLFW_KEY_KP_ADD, "KPAD" },
            { GLFW_KEY_KP_ENTER, "KPEN" },
            { GLFW_KEY_KP_EQUAL, "KPEQ" },
            { GLFW_KEY_LEFT_SHIFT, "LFSH" },
            { GLFW_KEY_LEFT_CONTROL, "LCTL" },
            { GLFW_KEY_LEFT_ALT, "LALT" },
            { GLFW_KEY_LEFT_SUPER, "LWIN" },
            { GLFW_KEY_RIGHT_SHIFT, "RTSH" },
            { GLFW_KEY_RIGHT_CONTROL, "RCTL" },
            { GLFW_KEY_RIGHT_ALT, "RALT" },
            { GLFW_KEY_RIGHT_ALT, "LVL3" },
            { GLFW_KEY_RIGHT_ALT, "MDSW" },
            { GLFW_KEY_RIGHT_SUPER, "RWIN" },
            { GLFW_KEY_MENU, "MENU" }
        };

        // Find the X11 key code -> GLFW key code mapping
        for (int scancode = scancodeMin;  scancode <= scancodeMax;  scancode++)
        {
            int key = GLFW_KEY_UNKNOWN;

            // Map the key name to a GLFW key code. Note: We use the US
            // keyboard layout. Because function keys aren't mapped correctly
            // when using traditional KeySym translations, they are mapped
            // here instead.
            for (int i = 0;  i < sizeof(keymap) / sizeof(keymap[0]);  i++)
            {
                if (strncmp(desc->names->keys[scancode].name,
                            keymap[i].name,
                            XkbKeyNameLength) == 0)
                {
                    key = keymap[i].key;
                    break;
                }
            }

            // Fall back to key aliases in case the key name did not match
            for (int i = 0;  i < desc->names->num_key_aliases;  i++)
            {
                if (key != GLFW_KEY_UNKNOWN)
                    break;

                if (strncmp(desc->names->key_aliases[i].real,
                            desc->names->keys[scancode].name,
                            XkbKeyNameLength) != 0)
                {
                    continue;
                }

                for (int j = 0;  j < sizeof(keymap) / sizeof(keymap[0]);  j++)
                {
                    if (strncmp(desc->names->key_aliases[i].alias,
                                keymap[j].name,
                                XkbKeyNameLength) == 0)
                    {
                        key = keymap[j].key;
                        break;
                    }
                }
            }

            __glfw.x11.keycodes[scancode] = key;
        }

        XkbFreeNames(desc, XkbKeyNamesMask, True);
        XkbFreeKeyboard(desc, 0, True);
    }
    else
        XDisplayKeycodes(__glfw.x11.display, &scancodeMin, &scancodeMax);

    int width;
    KeySym* keysyms = XGetKeyboardMapping(__glfw.x11.display,
                                          scancodeMin,
                                          scancodeMax - scancodeMin + 1,
                                          &width);

    for (int scancode = scancodeMin;  scancode <= scancodeMax;  scancode++)
    {
        // Translate the un-translated key codes using traditional X11 KeySym
        // lookups
        if (__glfw.x11.keycodes[scancode] < 0)
        {
            const size_t base = (scancode - scancodeMin) * width;
            __glfw.x11.keycodes[scancode] = translateKeySyms(&keysyms[base], width);
        }

        // Store the reverse translation for faster key name lookup
        if (__glfw.x11.keycodes[scancode] > 0)
            __glfw.x11.scancodes[__glfw.x11.keycodes[scancode]] = scancode;
    }

    XFree(keysyms);
}

// Check whether the IM has a usable style
//
static GLFWbool hasUsableInputMethodStyle(void)
{
    GLFWbool found = GLFW_FALSE;
    XIMStyles* styles = NULL;

    if (XGetIMValues(__glfw.x11.im, XNQueryInputStyle, &styles, NULL) != NULL)
        return GLFW_FALSE;

    for (unsigned int i = 0;  i < styles->count_styles;  i++)
    {
        if (styles->supported_styles[i] == (XIMPreeditNothing | XIMStatusNothing))
        {
            found = GLFW_TRUE;
            break;
        }
    }

    XFree(styles);
    return found;
}

static void inputMethodDestroyCallback(XIM im, XPointer clientData, XPointer callData)
{
    __glfw.x11.im = NULL;
}

static void inputMethodInstantiateCallback(Display* display,
                                           XPointer clientData,
                                           XPointer callData)
{
    if (__glfw.x11.im)
        return;

    __glfw.x11.im = XOpenIM(__glfw.x11.display, 0, NULL, NULL);
    if (__glfw.x11.im)
    {
        if (!hasUsableInputMethodStyle())
        {
            XCloseIM(__glfw.x11.im);
            __glfw.x11.im = NULL;
        }
    }

    if (__glfw.x11.im)
    {
        XIMCallback callback;
        callback.callback = (XIMProc) inputMethodDestroyCallback;
        callback.client_data = NULL;
        XSetIMValues(__glfw.x11.im, XNDestroyCallback, &callback, NULL);

        for (_GLFWwindow* window = __glfw.windowListHead;  window;  window = window->next)
            __glfwCreateInputContextX11(window);
    }
}

// Return the atom ID only if it is listed in the specified array
//
static Atom getAtomIfSupported(Atom* supportedAtoms,
                               unsigned long atomCount,
                               const char* atomName)
{
    const Atom atom = XInternAtom(__glfw.x11.display, atomName, False);

    for (unsigned long i = 0;  i < atomCount;  i++)
    {
        if (supportedAtoms[i] == atom)
            return atom;
    }

    return None;
}

// Check whether the running window manager is EWMH-compliant
//
static void detectEWMH(void)
{
    // First we read the _NET_SUPPORTING_WM_CHECK property on the root window

    Window* windowFromRoot = NULL;
    if (!___glfwGetWindowPropertyX11(__glfw.x11.root,
                                   __glfw.x11.NET_SUPPORTING_WM_CHECK,
                                   XA_WINDOW,
                                   (unsigned char**) &windowFromRoot))
    {
        return;
    }

    ___glfwGrabErrorHandlerX11();

    // If it exists, it should be the XID of a top-level window
    // Then we look for the same property on that window

    Window* windowFromChild = NULL;
    if (!___glfwGetWindowPropertyX11(*windowFromRoot,
                                   __glfw.x11.NET_SUPPORTING_WM_CHECK,
                                   XA_WINDOW,
                                   (unsigned char**) &windowFromChild))
    {
        XFree(windowFromRoot);
        return;
    }

    ___glfwReleaseErrorHandlerX11();

    // If the property exists, it should contain the XID of the window

    if (*windowFromRoot != *windowFromChild)
    {
        XFree(windowFromRoot);
        XFree(windowFromChild);
        return;
    }

    XFree(windowFromRoot);
    XFree(windowFromChild);

    // We are now fairly sure that an EWMH-compliant WM is currently running
    // We can now start querying the WM about what features it supports by
    // looking in the _NET_SUPPORTED property on the root window
    // It should contain a list of supported EWMH protocol and state atoms

    Atom* supportedAtoms = NULL;
    const unsigned long atomCount =
        ___glfwGetWindowPropertyX11(__glfw.x11.root,
                                  __glfw.x11.NET_SUPPORTED,
                                  XA_ATOM,
                                  (unsigned char**) &supportedAtoms);

    // See which of the atoms we support that are supported by the WM

    __glfw.x11.NET_WM_STATE =
        getAtomIfSupported(supportedAtoms, atomCount, "_NET_WM_STATE");
    __glfw.x11.NET_WM_STATE_ABOVE =
        getAtomIfSupported(supportedAtoms, atomCount, "_NET_WM_STATE_ABOVE");
    __glfw.x11.NET_WM_STATE_FULLSCREEN =
        getAtomIfSupported(supportedAtoms, atomCount, "_NET_WM_STATE_FULLSCREEN");
    __glfw.x11.NET_WM_STATE_MAXIMIZED_VERT =
        getAtomIfSupported(supportedAtoms, atomCount, "_NET_WM_STATE_MAXIMIZED_VERT");
    __glfw.x11.NET_WM_STATE_MAXIMIZED_HORZ =
        getAtomIfSupported(supportedAtoms, atomCount, "_NET_WM_STATE_MAXIMIZED_HORZ");
    __glfw.x11.NET_WM_STATE_DEMANDS_ATTENTION =
        getAtomIfSupported(supportedAtoms, atomCount, "_NET_WM_STATE_DEMANDS_ATTENTION");
    __glfw.x11.NET_WM_FULLSCREEN_MONITORS =
        getAtomIfSupported(supportedAtoms, atomCount, "_NET_WM_FULLSCREEN_MONITORS");
    __glfw.x11.NET_WM_WINDOW_TYPE =
        getAtomIfSupported(supportedAtoms, atomCount, "_NET_WM_WINDOW_TYPE");
    __glfw.x11.NET_WM_WINDOW_TYPE_NORMAL =
        getAtomIfSupported(supportedAtoms, atomCount, "_NET_WM_WINDOW_TYPE_NORMAL");
    __glfw.x11.NET_WORKAREA =
        getAtomIfSupported(supportedAtoms, atomCount, "_NET_WORKAREA");
    __glfw.x11.NET_CURRENT_DESKTOP =
        getAtomIfSupported(supportedAtoms, atomCount, "_NET_CURRENT_DESKTOP");
    __glfw.x11.NET_ACTIVE_WINDOW =
        getAtomIfSupported(supportedAtoms, atomCount, "_NET_ACTIVE_WINDOW");
    __glfw.x11.NET_FRAME_EXTENTS =
        getAtomIfSupported(supportedAtoms, atomCount, "_NET_FRAME_EXTENTS");
    __glfw.x11.NET_REQUEST_FRAME_EXTENTS =
        getAtomIfSupported(supportedAtoms, atomCount, "_NET_REQUEST_FRAME_EXTENTS");

    if (supportedAtoms)
        XFree(supportedAtoms);
}

// Look for and initialize supported X11 extensions
//
static GLFWbool initExtensions(void)
{
#if defined(__OpenBSD__) || defined(__NetBSD__)
    __glfw.x11.vidmode.handle = __glfwPlatformLoadModule("libXxf86vm.so");
#else
    __glfw.x11.vidmode.handle = __glfwPlatformLoadModule("libXxf86vm.so.1");
#endif
    if (__glfw.x11.vidmode.handle)
    {
        __glfw.x11.vidmode.QueryExtension = (PFN_XF86VidModeQueryExtension)
            __glfwPlatformGetModuleSymbol(__glfw.x11.vidmode.handle, "XF86VidModeQueryExtension");
        __glfw.x11.vidmode.GetGammaRamp = (PFN_XF86VidModeGetGammaRamp)
            __glfwPlatformGetModuleSymbol(__glfw.x11.vidmode.handle, "XF86VidModeGetGammaRamp");
        __glfw.x11.vidmode.SetGammaRamp = (PFN_XF86VidModeSetGammaRamp)
            __glfwPlatformGetModuleSymbol(__glfw.x11.vidmode.handle, "XF86VidModeSetGammaRamp");
        __glfw.x11.vidmode.GetGammaRampSize = (PFN_XF86VidModeGetGammaRampSize)
            __glfwPlatformGetModuleSymbol(__glfw.x11.vidmode.handle, "XF86VidModeGetGammaRampSize");

        __glfw.x11.vidmode.available =
            XF86VidModeQueryExtension(__glfw.x11.display,
                                      &__glfw.x11.vidmode.eventBase,
                                      &__glfw.x11.vidmode.errorBase);
    }

#if defined(__CYGWIN__)
    __glfw.x11.xi.handle = __glfwPlatformLoadModule("libXi-6.so");
#elif defined(__OpenBSD__) || defined(__NetBSD__)
    __glfw.x11.xi.handle = __glfwPlatformLoadModule("libXi.so");
#else
    __glfw.x11.xi.handle = __glfwPlatformLoadModule("libXi.so.6");
#endif
    if (__glfw.x11.xi.handle)
    {
        __glfw.x11.xi.QueryVersion = (PFN_XIQueryVersion)
            __glfwPlatformGetModuleSymbol(__glfw.x11.xi.handle, "XIQueryVersion");
        __glfw.x11.xi.SelectEvents = (PFN_XISelectEvents)
            __glfwPlatformGetModuleSymbol(__glfw.x11.xi.handle, "XISelectEvents");

        if (XQueryExtension(__glfw.x11.display,
                            "XInputExtension",
                            &__glfw.x11.xi.majorOpcode,
                            &__glfw.x11.xi.eventBase,
                            &__glfw.x11.xi.errorBase))
        {
            __glfw.x11.xi.major = 2;
            __glfw.x11.xi.minor = 0;

            if (XIQueryVersion(__glfw.x11.display,
                               &__glfw.x11.xi.major,
                               &__glfw.x11.xi.minor) == Success)
            {
                __glfw.x11.xi.available = GLFW_TRUE;
            }
        }
    }

#if defined(__CYGWIN__)
    __glfw.x11.randr.handle = __glfwPlatformLoadModule("libXrandr-2.so");
#elif defined(__OpenBSD__) || defined(__NetBSD__)
    __glfw.x11.randr.handle = __glfwPlatformLoadModule("libXrandr.so");
#else
    __glfw.x11.randr.handle = __glfwPlatformLoadModule("libXrandr.so.2");
#endif
    if (__glfw.x11.randr.handle)
    {
        __glfw.x11.randr.AllocGamma = (PFN_XRRAllocGamma)
            __glfwPlatformGetModuleSymbol(__glfw.x11.randr.handle, "XRRAllocGamma");
        __glfw.x11.randr.FreeGamma = (PFN_XRRFreeGamma)
            __glfwPlatformGetModuleSymbol(__glfw.x11.randr.handle, "XRRFreeGamma");
        __glfw.x11.randr.FreeCrtcInfo = (PFN_XRRFreeCrtcInfo)
            __glfwPlatformGetModuleSymbol(__glfw.x11.randr.handle, "XRRFreeCrtcInfo");
        __glfw.x11.randr.FreeGamma = (PFN_XRRFreeGamma)
            __glfwPlatformGetModuleSymbol(__glfw.x11.randr.handle, "XRRFreeGamma");
        __glfw.x11.randr.FreeOutputInfo = (PFN_XRRFreeOutputInfo)
            __glfwPlatformGetModuleSymbol(__glfw.x11.randr.handle, "XRRFreeOutputInfo");
        __glfw.x11.randr.FreeScreenResources = (PFN_XRRFreeScreenResources)
            __glfwPlatformGetModuleSymbol(__glfw.x11.randr.handle, "XRRFreeScreenResources");
        __glfw.x11.randr.GetCrtcGamma = (PFN_XRRGetCrtcGamma)
            __glfwPlatformGetModuleSymbol(__glfw.x11.randr.handle, "XRRGetCrtcGamma");
        __glfw.x11.randr.GetCrtcGammaSize = (PFN_XRRGetCrtcGammaSize)
            __glfwPlatformGetModuleSymbol(__glfw.x11.randr.handle, "XRRGetCrtcGammaSize");
        __glfw.x11.randr.GetCrtcInfo = (PFN_XRRGetCrtcInfo)
            __glfwPlatformGetModuleSymbol(__glfw.x11.randr.handle, "XRRGetCrtcInfo");
        __glfw.x11.randr.GetOutputInfo = (PFN_XRRGetOutputInfo)
            __glfwPlatformGetModuleSymbol(__glfw.x11.randr.handle, "XRRGetOutputInfo");
        __glfw.x11.randr.GetOutputPrimary = (PFN_XRRGetOutputPrimary)
            __glfwPlatformGetModuleSymbol(__glfw.x11.randr.handle, "XRRGetOutputPrimary");
        __glfw.x11.randr.GetScreenResourcesCurrent = (PFN_XRRGetScreenResourcesCurrent)
            __glfwPlatformGetModuleSymbol(__glfw.x11.randr.handle, "XRRGetScreenResourcesCurrent");
        __glfw.x11.randr.QueryExtension = (PFN_XRRQueryExtension)
            __glfwPlatformGetModuleSymbol(__glfw.x11.randr.handle, "XRRQueryExtension");
        __glfw.x11.randr.QueryVersion = (PFN_XRRQueryVersion)
            __glfwPlatformGetModuleSymbol(__glfw.x11.randr.handle, "XRRQueryVersion");
        __glfw.x11.randr.SelectInput = (PFN_XRRSelectInput)
            __glfwPlatformGetModuleSymbol(__glfw.x11.randr.handle, "XRRSelectInput");
        __glfw.x11.randr.SetCrtcConfig = (PFN_XRRSetCrtcConfig)
            __glfwPlatformGetModuleSymbol(__glfw.x11.randr.handle, "XRRSetCrtcConfig");
        __glfw.x11.randr.SetCrtcGamma = (PFN_XRRSetCrtcGamma)
            __glfwPlatformGetModuleSymbol(__glfw.x11.randr.handle, "XRRSetCrtcGamma");
        __glfw.x11.randr.UpdateConfiguration = (PFN_XRRUpdateConfiguration)
            __glfwPlatformGetModuleSymbol(__glfw.x11.randr.handle, "XRRUpdateConfiguration");

        if (XRRQueryExtension(__glfw.x11.display,
                              &__glfw.x11.randr.eventBase,
                              &__glfw.x11.randr.errorBase))
        {
            if (XRRQueryVersion(__glfw.x11.display,
                                &__glfw.x11.randr.major,
                                &__glfw.x11.randr.minor))
            {
                // The GLFW RandR path requires at least version 1.3
                if (__glfw.x11.randr.major > 1 || __glfw.x11.randr.minor >= 3)
                    __glfw.x11.randr.available = GLFW_TRUE;
            }
            else
            {
                ___glfwInputError(GLFW_PLATFORM_ERROR,
                                "X11: Failed to query RandR version");
            }
        }
    }

    if (__glfw.x11.randr.available)
    {
        XRRScreenResources* sr = XRRGetScreenResourcesCurrent(__glfw.x11.display,
                                                              __glfw.x11.root);

        if (!sr->ncrtc || !XRRGetCrtcGammaSize(__glfw.x11.display, sr->crtcs[0]))
        {
            // This is likely an older Nvidia driver with broken gamma support
            // Flag it as useless and fall back to xf86vm gamma, if available
            __glfw.x11.randr.gammaBroken = GLFW_TRUE;
        }

        if (!sr->ncrtc)
        {
            // A system without CRTCs is likely a system with broken RandR
            // Disable the RandR monitor path and fall back to core functions
            __glfw.x11.randr.monitorBroken = GLFW_TRUE;
        }

        XRRFreeScreenResources(sr);
    }

    if (__glfw.x11.randr.available && !__glfw.x11.randr.monitorBroken)
    {
        XRRSelectInput(__glfw.x11.display, __glfw.x11.root,
                       RROutputChangeNotifyMask);
    }

#if defined(__CYGWIN__)
    __glfw.x11.xcursor.handle = __glfwPlatformLoadModule("libXcursor-1.so");
#elif defined(__OpenBSD__) || defined(__NetBSD__)
    __glfw.x11.xcursor.handle = __glfwPlatformLoadModule("libXcursor.so");
#else
    __glfw.x11.xcursor.handle = __glfwPlatformLoadModule("libXcursor.so.1");
#endif
    if (__glfw.x11.xcursor.handle)
    {
        __glfw.x11.xcursor.ImageCreate = (PFN_XcursorImageCreate)
            __glfwPlatformGetModuleSymbol(__glfw.x11.xcursor.handle, "XcursorImageCreate");
        __glfw.x11.xcursor.ImageDestroy = (PFN_XcursorImageDestroy)
            __glfwPlatformGetModuleSymbol(__glfw.x11.xcursor.handle, "XcursorImageDestroy");
        __glfw.x11.xcursor.ImageLoadCursor = (PFN_XcursorImageLoadCursor)
            __glfwPlatformGetModuleSymbol(__glfw.x11.xcursor.handle, "XcursorImageLoadCursor");
        __glfw.x11.xcursor.GetTheme = (PFN_XcursorGetTheme)
            __glfwPlatformGetModuleSymbol(__glfw.x11.xcursor.handle, "XcursorGetTheme");
        __glfw.x11.xcursor.GetDefaultSize = (PFN_XcursorGetDefaultSize)
            __glfwPlatformGetModuleSymbol(__glfw.x11.xcursor.handle, "XcursorGetDefaultSize");
        __glfw.x11.xcursor.LibraryLoadImage = (PFN_XcursorLibraryLoadImage)
            __glfwPlatformGetModuleSymbol(__glfw.x11.xcursor.handle, "XcursorLibraryLoadImage");
    }

#if defined(__CYGWIN__)
    __glfw.x11.xinerama.handle = __glfwPlatformLoadModule("libXinerama-1.so");
#elif defined(__OpenBSD__) || defined(__NetBSD__)
    __glfw.x11.xinerama.handle = __glfwPlatformLoadModule("libXinerama.so");
#else
    __glfw.x11.xinerama.handle = __glfwPlatformLoadModule("libXinerama.so.1");
#endif
    if (__glfw.x11.xinerama.handle)
    {
        __glfw.x11.xinerama.IsActive = (PFN_XineramaIsActive)
            __glfwPlatformGetModuleSymbol(__glfw.x11.xinerama.handle, "XineramaIsActive");
        __glfw.x11.xinerama.QueryExtension = (PFN_XineramaQueryExtension)
            __glfwPlatformGetModuleSymbol(__glfw.x11.xinerama.handle, "XineramaQueryExtension");
        __glfw.x11.xinerama.QueryScreens = (PFN_XineramaQueryScreens)
            __glfwPlatformGetModuleSymbol(__glfw.x11.xinerama.handle, "XineramaQueryScreens");

        if (XineramaQueryExtension(__glfw.x11.display,
                                   &__glfw.x11.xinerama.major,
                                   &__glfw.x11.xinerama.minor))
        {
            if (XineramaIsActive(__glfw.x11.display))
                __glfw.x11.xinerama.available = GLFW_TRUE;
        }
    }

    __glfw.x11.xkb.major = 1;
    __glfw.x11.xkb.minor = 0;
    __glfw.x11.xkb.available =
        XkbQueryExtension(__glfw.x11.display,
                          &__glfw.x11.xkb.majorOpcode,
                          &__glfw.x11.xkb.eventBase,
                          &__glfw.x11.xkb.errorBase,
                          &__glfw.x11.xkb.major,
                          &__glfw.x11.xkb.minor);

    if (__glfw.x11.xkb.available)
    {
        Bool supported;

        if (XkbSetDetectableAutoRepeat(__glfw.x11.display, True, &supported))
        {
            if (supported)
                __glfw.x11.xkb.detectable = GLFW_TRUE;
        }

        XkbStateRec state;
        if (XkbGetState(__glfw.x11.display, XkbUseCoreKbd, &state) == Success)
            __glfw.x11.xkb.group = (unsigned int)state.group;

        XkbSelectEventDetails(__glfw.x11.display, XkbUseCoreKbd, XkbStateNotify,
                              XkbGroupStateMask, XkbGroupStateMask);
    }

    if (__glfw.hints.init.x11.xcbVulkanSurface)
    {
#if defined(__CYGWIN__)
        __glfw.x11.x11xcb.handle = __glfwPlatformLoadModule("libX11-xcb-1.so");
#elif defined(__OpenBSD__) || defined(__NetBSD__)
        __glfw.x11.x11xcb.handle = __glfwPlatformLoadModule("libX11-xcb.so");
#else
        __glfw.x11.x11xcb.handle = __glfwPlatformLoadModule("libX11-xcb.so.1");
#endif
    }

    if (__glfw.x11.x11xcb.handle)
    {
        __glfw.x11.x11xcb.GetXCBConnection = (PFN_XGetXCBConnection)
            __glfwPlatformGetModuleSymbol(__glfw.x11.x11xcb.handle, "XGetXCBConnection");
    }

#if defined(__CYGWIN__)
    __glfw.x11.xrender.handle = __glfwPlatformLoadModule("libXrender-1.so");
#elif defined(__OpenBSD__) || defined(__NetBSD__)
    __glfw.x11.xrender.handle = __glfwPlatformLoadModule("libXrender.so");
#else
    __glfw.x11.xrender.handle = __glfwPlatformLoadModule("libXrender.so.1");
#endif
    if (__glfw.x11.xrender.handle)
    {
        __glfw.x11.xrender.QueryExtension = (PFN_XRenderQueryExtension)
            __glfwPlatformGetModuleSymbol(__glfw.x11.xrender.handle, "XRenderQueryExtension");
        __glfw.x11.xrender.QueryVersion = (PFN_XRenderQueryVersion)
            __glfwPlatformGetModuleSymbol(__glfw.x11.xrender.handle, "XRenderQueryVersion");
        __glfw.x11.xrender.FindVisualFormat = (PFN_XRenderFindVisualFormat)
            __glfwPlatformGetModuleSymbol(__glfw.x11.xrender.handle, "XRenderFindVisualFormat");

        if (XRenderQueryExtension(__glfw.x11.display,
                                  &__glfw.x11.xrender.errorBase,
                                  &__glfw.x11.xrender.eventBase))
        {
            if (XRenderQueryVersion(__glfw.x11.display,
                                    &__glfw.x11.xrender.major,
                                    &__glfw.x11.xrender.minor))
            {
                __glfw.x11.xrender.available = GLFW_TRUE;
            }
        }
    }

#if defined(__CYGWIN__)
    __glfw.x11.xshape.handle = __glfwPlatformLoadModule("libXext-6.so");
#elif defined(__OpenBSD__) || defined(__NetBSD__)
    __glfw.x11.xshape.handle = __glfwPlatformLoadModule("libXext.so");
#else
    __glfw.x11.xshape.handle = __glfwPlatformLoadModule("libXext.so.6");
#endif
    if (__glfw.x11.xshape.handle)
    {
        __glfw.x11.xshape.QueryExtension = (PFN_XShapeQueryExtension)
            __glfwPlatformGetModuleSymbol(__glfw.x11.xshape.handle, "XShapeQueryExtension");
        __glfw.x11.xshape.ShapeCombineRegion = (PFN_XShapeCombineRegion)
            __glfwPlatformGetModuleSymbol(__glfw.x11.xshape.handle, "XShapeCombineRegion");
        __glfw.x11.xshape.QueryVersion = (PFN_XShapeQueryVersion)
            __glfwPlatformGetModuleSymbol(__glfw.x11.xshape.handle, "XShapeQueryVersion");
        __glfw.x11.xshape.ShapeCombineMask = (PFN_XShapeCombineMask)
            __glfwPlatformGetModuleSymbol(__glfw.x11.xshape.handle, "XShapeCombineMask");

        if (XShapeQueryExtension(__glfw.x11.display,
            &__glfw.x11.xshape.errorBase,
            &__glfw.x11.xshape.eventBase))
        {
            if (XShapeQueryVersion(__glfw.x11.display,
                &__glfw.x11.xshape.major,
                &__glfw.x11.xshape.minor))
            {
                __glfw.x11.xshape.available = GLFW_TRUE;
            }
        }
    }

    // Update the key code LUT
    // FIXME: We should listen to XkbMapNotify events to track changes to
    // the keyboard mapping.
    createKeyTables();

    // String format atoms
    __glfw.x11.NULL_ = XInternAtom(__glfw.x11.display, "NULL", False);
    __glfw.x11.UTF8_STRING = XInternAtom(__glfw.x11.display, "UTF8_STRING", False);
    __glfw.x11.ATOM_PAIR = XInternAtom(__glfw.x11.display, "ATOM_PAIR", False);

    // Custom selection property atom
    __glfw.x11.GLFW_SELECTION =
        XInternAtom(__glfw.x11.display, "GLFW_SELECTION", False);

    // ICCCM standard clipboard atoms
    __glfw.x11.TARGETS = XInternAtom(__glfw.x11.display, "TARGETS", False);
    __glfw.x11.MULTIPLE = XInternAtom(__glfw.x11.display, "MULTIPLE", False);
    __glfw.x11.PRIMARY = XInternAtom(__glfw.x11.display, "PRIMARY", False);
    __glfw.x11.INCR = XInternAtom(__glfw.x11.display, "INCR", False);
    __glfw.x11.CLIPBOARD = XInternAtom(__glfw.x11.display, "CLIPBOARD", False);

    // Clipboard manager atoms
    __glfw.x11.CLIPBOARD_MANAGER =
        XInternAtom(__glfw.x11.display, "CLIPBOARD_MANAGER", False);
    __glfw.x11.SAVE_TARGETS =
        XInternAtom(__glfw.x11.display, "SAVE_TARGETS", False);

    // Xdnd (drag and drop) atoms
    __glfw.x11.XdndAware = XInternAtom(__glfw.x11.display, "XdndAware", False);
    __glfw.x11.XdndEnter = XInternAtom(__glfw.x11.display, "XdndEnter", False);
    __glfw.x11.XdndPosition = XInternAtom(__glfw.x11.display, "XdndPosition", False);
    __glfw.x11.XdndStatus = XInternAtom(__glfw.x11.display, "XdndStatus", False);
    __glfw.x11.XdndActionCopy = XInternAtom(__glfw.x11.display, "XdndActionCopy", False);
    __glfw.x11.XdndDrop = XInternAtom(__glfw.x11.display, "XdndDrop", False);
    __glfw.x11.XdndFinished = XInternAtom(__glfw.x11.display, "XdndFinished", False);
    __glfw.x11.XdndSelection = XInternAtom(__glfw.x11.display, "XdndSelection", False);
    __glfw.x11.XdndTypeList = XInternAtom(__glfw.x11.display, "XdndTypeList", False);
    __glfw.x11.text_uri_list = XInternAtom(__glfw.x11.display, "text/uri-list", False);

    // ICCCM, EWMH and Motif window property atoms
    // These can be set safely even without WM support
    // The EWMH atoms that require WM support are handled in detectEWMH
    __glfw.x11.WM_PROTOCOLS =
        XInternAtom(__glfw.x11.display, "WM_PROTOCOLS", False);
    __glfw.x11.WM_STATE =
        XInternAtom(__glfw.x11.display, "WM_STATE", False);
    __glfw.x11.WM_DELETE_WINDOW =
        XInternAtom(__glfw.x11.display, "WM_DELETE_WINDOW", False);
    __glfw.x11.NET_SUPPORTED =
        XInternAtom(__glfw.x11.display, "_NET_SUPPORTED", False);
    __glfw.x11.NET_SUPPORTING_WM_CHECK =
        XInternAtom(__glfw.x11.display, "_NET_SUPPORTING_WM_CHECK", False);
    __glfw.x11.NET_WM_ICON =
        XInternAtom(__glfw.x11.display, "_NET_WM_ICON", False);
    __glfw.x11.NET_WM_PING =
        XInternAtom(__glfw.x11.display, "_NET_WM_PING", False);
    __glfw.x11.NET_WM_PID =
        XInternAtom(__glfw.x11.display, "_NET_WM_PID", False);
    __glfw.x11.NET_WM_NAME =
        XInternAtom(__glfw.x11.display, "_NET_WM_NAME", False);
    __glfw.x11.NET_WM_ICON_NAME =
        XInternAtom(__glfw.x11.display, "_NET_WM_ICON_NAME", False);
    __glfw.x11.NET_WM_BYPASS_COMPOSITOR =
        XInternAtom(__glfw.x11.display, "_NET_WM_BYPASS_COMPOSITOR", False);
    __glfw.x11.NET_WM_WINDOW_OPACITY =
        XInternAtom(__glfw.x11.display, "_NET_WM_WINDOW_OPACITY", False);
    __glfw.x11.MOTIF_WM_HINTS =
        XInternAtom(__glfw.x11.display, "_MOTIF_WM_HINTS", False);

    // The compositing manager selection name contains the screen number
    {
        char name[32];
        snprintf(name, sizeof(name), "_NET_WM_CM_S%u", __glfw.x11.screen);
        __glfw.x11.NET_WM_CM_Sx = XInternAtom(__glfw.x11.display, name, False);
    }

    // Detect whether an EWMH-conformant window manager is running
    detectEWMH();

    return GLFW_TRUE;
}

// Retrieve system content scale via folklore heuristics
//
static void getSystemContentScale(float* xscale, float* yscale)
{
    // Start by assuming the default X11 DPI
    // NOTE: Some desktop environments (KDE) may remove the Xft.dpi field when it
    //       would be set to 96, so assume that is the case if we cannot find it
    float xdpi = 96.f, ydpi = 96.f;

    // NOTE: Basing the scale on Xft.dpi where available should provide the most
    //       consistent user experience (matches Qt, Gtk, etc), although not
    //       always the most accurate one
    char* rms = XResourceManagerString(__glfw.x11.display);
    if (rms)
    {
        XrmDatabase db = XrmGetStringDatabase(rms);
        if (db)
        {
            XrmValue value;
            char* type = NULL;

            if (XrmGetResource(db, "Xft.dpi", "Xft.Dpi", &type, &value))
            {
                if (type && strcmp(type, "String") == 0)
                    xdpi = ydpi = atof(value.addr);
            }

            XrmDestroyDatabase(db);
        }
    }

    *xscale = xdpi / 96.f;
    *yscale = ydpi / 96.f;
}

// Create a blank cursor for hidden and disabled cursor modes
//
static Cursor createHiddenCursor(void)
{
    unsigned char pixels[16 * 16 * 4] = { 0 };
    GLFWimage image = { 16, 16, pixels };
    return __glfwCreateNativeCursorX11(&image, 0, 0);
}

// Create a helper window for IPC
//
static Window createHelperWindow(void)
{
    XSetWindowAttributes wa;
    wa.event_mask = PropertyChangeMask;

    return XCreateWindow(__glfw.x11.display, __glfw.x11.root,
                         0, 0, 1, 1, 0, 0,
                         InputOnly,
                         DefaultVisual(__glfw.x11.display, __glfw.x11.screen),
                         CWEventMask, &wa);
}

// Create the pipe for empty events without assumuing the OS has pipe2(2)
//
static GLFWbool createEmptyEventPipe(void)
{
    if (pipe(__glfw.x11.emptyEventPipe) != 0)
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "X11: Failed to create empty event pipe: %s",
                        strerror(errno));
        return GLFW_FALSE;
    }

    for (int i = 0; i < 2; i++)
    {
        const int sf = fcntl(__glfw.x11.emptyEventPipe[i], F_GETFL, 0);
        const int df = fcntl(__glfw.x11.emptyEventPipe[i], F_GETFD, 0);

        if (sf == -1 || df == -1 ||
            fcntl(__glfw.x11.emptyEventPipe[i], F_SETFL, sf | O_NONBLOCK) == -1 ||
            fcntl(__glfw.x11.emptyEventPipe[i], F_SETFD, df | FD_CLOEXEC) == -1)
        {
            ___glfwInputError(GLFW_PLATFORM_ERROR,
                            "X11: Failed to set flags for empty event pipe: %s",
                            strerror(errno));
            return GLFW_FALSE;
        }
    }

    return GLFW_TRUE;
}

// X error handler
//
static int errorHandler(Display *display, XErrorEvent* event)
{
    if (__glfw.x11.display != display)
        return 0;

    __glfw.x11.errorCode = event->error_code;
    return 0;
}


//////////////////////////////////////////////////////////////////////////
//////                       GLFW internal API                      //////
//////////////////////////////////////////////////////////////////////////

// Sets the X error handler callback
//
void ___glfwGrabErrorHandlerX11(void)
{
    assert(__glfw.x11.errorHandler == NULL);
    __glfw.x11.errorCode = Success;
    __glfw.x11.errorHandler = XSetErrorHandler(errorHandler);
}

// Clears the X error handler callback
//
void ___glfwReleaseErrorHandlerX11(void)
{
    // Synchronize to make sure all commands are processed
    XSync(__glfw.x11.display, False);
    XSetErrorHandler(__glfw.x11.errorHandler);
    __glfw.x11.errorHandler = NULL;
}

// Reports the specified error, appending information about the last X error
//
void ____glfwInputErrorX11(int error, const char* message)
{
    char buffer[_GLFW_MESSAGE_SIZE];
    XGetErrorText(__glfw.x11.display, __glfw.x11.errorCode,
                  buffer, sizeof(buffer));

    ___glfwInputError(error, "%s: %s", message, buffer);
}

// Creates a native cursor object from the specified image and hotspot
//
Cursor __glfwCreateNativeCursorX11(const GLFWimage* image, int xhot, int yhot)
{
    Cursor cursor;

    if (!__glfw.x11.xcursor.handle)
        return None;

    XcursorImage* native = XcursorImageCreate(image->width, image->height);
    if (native == NULL)
        return None;

    native->xhot = xhot;
    native->yhot = yhot;

    unsigned char* source = (unsigned char*) image->pixels;
    XcursorPixel* target = native->pixels;

    for (int i = 0;  i < image->width * image->height;  i++, target++, source += 4)
    {
        unsigned int alpha = source[3];

        *target = (alpha << 24) |
                  ((unsigned char) ((source[0] * alpha) / 255) << 16) |
                  ((unsigned char) ((source[1] * alpha) / 255) <<  8) |
                  ((unsigned char) ((source[2] * alpha) / 255) <<  0);
    }

    cursor = XcursorImageLoadCursor(__glfw.x11.display, native);
    XcursorImageDestroy(native);

    return cursor;
}


//////////////////////////////////////////////////////////////////////////
//////                       GLFW platform API                      //////
//////////////////////////////////////////////////////////////////////////

GLFWbool __glfwConnectX11(int platformID, _GLFWplatform* platform)
{
    const _GLFWplatform x11 =
    {
        GLFW_PLATFORM_X11,
        ___glfwInitX11,
        ___glfwTerminateX11,
        ___glfwGetCursorPosX11,
        ____glfwSetCursorPosX11,
        ___glfwSetCursorModeX11,
        __glfwSetRawMouseMotionX11,
        ___glfwRawMouseMotionSupportedX11,
        ____glfwCreateCursorX11,
        ___glfwCreateStandardCursorX11,
        ___glfwDestroyCursorX11,
        ___glfwSetCursorX11,
        __glfwGetScancodeNameX11,
        ____glfwGetKeyScancodeX11,
        ___glfwSetClipboardStringX11,
        ___glfwGetClipboardStringX11,
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
        ___glfwFreeMonitorX11,
        ___glfwGetMonitorPosX11,
        ___glfwGetMonitorContentScaleX11,
        ___glfwGetMonitorWorkareaX11,
        ____glfwGetVideoModesX11,
        ___glfwGetVideoModeX11,
        ___glfwGetGammaRampX11,
        ____glfwSetGammaRampX11,
        ___glfwCreateWindowX11,
        ___glfwDestroyWindowX11,
        ___glfwSetWindowTitleX11,
        ___glfwSetWindowIconX11,
        ___glfwGetWindowPosX11,
        ___glfwSetWindowPosX11,
        ___glfwGetWindowSizeX11,
        ___glfwSetWindowSizeX11,
        ____glfwSetWindowSizeLimitsX11,
        ___glfwSetWindowAspectRatioX11,
        ___glfwGetFramebufferSizeX11,
        ___glfwGetWindowFrameSizeX11,
        ___glfwGetWindowContentScaleX11,
        ___glfwIconifyWindowX11,
        ___glfwRestoreWindowX11,
        ___glfwMaximizeWindowX11,
        ___glfwShowWindowX11,
        ___glfwHideWindowX11,
        ___glfwRequestWindowAttentionX11,
        ___glfwFocusWindowX11,
        ___glfwSetWindowMonitorX11,
        __glfwWindowFocusedX11,
        __glfwWindowIconifiedX11,
        __glfwWindowVisibleX11,
        __glfwWindowMaximizedX11,
        __glfwWindowHoveredX11,
        __glfwFramebufferTransparentX11,
        ___glfwGetWindowOpacityX11,
        __glfwSetWindowResizableX11,
        __glfwSetWindowDecoratedX11,
        __glfwSetWindowFloatingX11,
        ___glfwSetWindowOpacityX11,
        __glfwSetWindowMousePassthroughX11,
        ___glfwPollEventsX11,
        ___glfwWaitEventsX11,
        ____glfwWaitEventsTimeoutX11,
        ___glfwPostEmptyEventX11,
        __glfwGetEGLPlatformX11,
        __glfwGetEGLNativeDisplayX11,
        __glfwGetEGLNativeWindowX11,
        ___glfwGetRequiredInstanceExtensionsX11,
        ___glfwGetPhysicalDevicePresentationSupportX11,
        ____glfwCreateWindowSurfaceX11,
    };

    // HACK: If the application has left the locale as "C" then both wide
    //       character text input and explicit UTF-8 input via XIM will break
    //       This sets the CTYPE part of the current locale from the environment
    //       in the hope that it is set to something more sane than "C"
    if (strcmp(setlocale(LC_CTYPE, NULL), "C") == 0)
        setlocale(LC_CTYPE, "");

#if defined(__CYGWIN__)
    void* module = __glfwPlatformLoadModule("libX11-6.so");
#elif defined(__OpenBSD__) || defined(__NetBSD__)
    void* module = __glfwPlatformLoadModule("libX11.so");
#else
    void* module = __glfwPlatformLoadModule("libX11.so.6");
#endif
    if (!module)
    {
        if (platformID == GLFW_PLATFORM_X11)
            ___glfwInputError(GLFW_PLATFORM_ERROR, "X11: Failed to load Xlib");

        return GLFW_FALSE;
    }

    PFN_XInitThreads XInitThreads = (PFN_XInitThreads)
        __glfwPlatformGetModuleSymbol(module, "XInitThreads");
    PFN_XrmInitialize XrmInitialize = (PFN_XrmInitialize)
        __glfwPlatformGetModuleSymbol(module, "XrmInitialize");
    PFN_XOpenDisplay XOpenDisplay = (PFN_XOpenDisplay)
        __glfwPlatformGetModuleSymbol(module, "XOpenDisplay");
    if (!XInitThreads || !XrmInitialize || !XOpenDisplay)
    {
        if (platformID == GLFW_PLATFORM_X11)
            ___glfwInputError(GLFW_PLATFORM_ERROR, "X11: Failed to load Xlib entry point");

        __glfwPlatformFreeModule(module);
        return GLFW_FALSE;
    }

    XInitThreads();
    XrmInitialize();

    Display* display = XOpenDisplay(NULL);
    if (!display)
    {
        if (platformID == GLFW_PLATFORM_X11)
        {
            const char* name = getenv("DISPLAY");
            if (name)
            {
                ___glfwInputError(GLFW_PLATFORM_UNAVAILABLE,
                                "X11: Failed to open display %s", name);
            }
            else
            {
                ___glfwInputError(GLFW_PLATFORM_UNAVAILABLE,
                                "X11: The DISPLAY environment variable is missing");
            }
        }

        __glfwPlatformFreeModule(module);
        return GLFW_FALSE;
    }

    __glfw.x11.display = display;
    __glfw.x11.xlib.handle = module;

    *platform = x11;
    return GLFW_TRUE;
}

int ___glfwInitX11(void)
{
    __glfw.x11.xlib.AllocClassHint = (PFN_XAllocClassHint)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XAllocClassHint");
    __glfw.x11.xlib.AllocSizeHints = (PFN_XAllocSizeHints)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XAllocSizeHints");
    __glfw.x11.xlib.AllocWMHints = (PFN_XAllocWMHints)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XAllocWMHints");
    __glfw.x11.xlib.ChangeProperty = (PFN_XChangeProperty)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XChangeProperty");
    __glfw.x11.xlib.ChangeWindowAttributes = (PFN_XChangeWindowAttributes)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XChangeWindowAttributes");
    __glfw.x11.xlib.CheckIfEvent = (PFN_XCheckIfEvent)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XCheckIfEvent");
    __glfw.x11.xlib.CheckTypedWindowEvent = (PFN_XCheckTypedWindowEvent)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XCheckTypedWindowEvent");
    __glfw.x11.xlib.CloseDisplay = (PFN_XCloseDisplay)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XCloseDisplay");
    __glfw.x11.xlib.CloseIM = (PFN_XCloseIM)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XCloseIM");
    __glfw.x11.xlib.ConvertSelection = (PFN_XConvertSelection)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XConvertSelection");
    __glfw.x11.xlib.CreateColormap = (PFN_XCreateColormap)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XCreateColormap");
    __glfw.x11.xlib.CreateFontCursor = (PFN_XCreateFontCursor)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XCreateFontCursor");
    __glfw.x11.xlib.CreateIC = (PFN_XCreateIC)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XCreateIC");
    __glfw.x11.xlib.CreateRegion = (PFN_XCreateRegion)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XCreateRegion");
    __glfw.x11.xlib.CreateWindow = (PFN_XCreateWindow)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XCreateWindow");
    __glfw.x11.xlib.DefineCursor = (PFN_XDefineCursor)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XDefineCursor");
    __glfw.x11.xlib.DeleteContext = (PFN_XDeleteContext)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XDeleteContext");
    __glfw.x11.xlib.DeleteProperty = (PFN_XDeleteProperty)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XDeleteProperty");
    __glfw.x11.xlib.DestroyIC = (PFN_XDestroyIC)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XDestroyIC");
    __glfw.x11.xlib.DestroyRegion = (PFN_XDestroyRegion)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XDestroyRegion");
    __glfw.x11.xlib.DestroyWindow = (PFN_XDestroyWindow)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XDestroyWindow");
    __glfw.x11.xlib.DisplayKeycodes = (PFN_XDisplayKeycodes)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XDisplayKeycodes");
    __glfw.x11.xlib.EventsQueued = (PFN_XEventsQueued)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XEventsQueued");
    __glfw.x11.xlib.FilterEvent = (PFN_XFilterEvent)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XFilterEvent");
    __glfw.x11.xlib.FindContext = (PFN_XFindContext)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XFindContext");
    __glfw.x11.xlib.Flush = (PFN_XFlush)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XFlush");
    __glfw.x11.xlib.Free = (PFN_XFree)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XFree");
    __glfw.x11.xlib.FreeColormap = (PFN_XFreeColormap)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XFreeColormap");
    __glfw.x11.xlib.FreeCursor = (PFN_XFreeCursor)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XFreeCursor");
    __glfw.x11.xlib.FreeEventData = (PFN_XFreeEventData)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XFreeEventData");
    __glfw.x11.xlib.GetErrorText = (PFN_XGetErrorText)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XGetErrorText");
    __glfw.x11.xlib.GetEventData = (PFN_XGetEventData)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XGetEventData");
    __glfw.x11.xlib.GetICValues = (PFN_XGetICValues)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XGetICValues");
    __glfw.x11.xlib.GetIMValues = (PFN_XGetIMValues)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XGetIMValues");
    __glfw.x11.xlib.GetInputFocus = (PFN_XGetInputFocus)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XGetInputFocus");
    __glfw.x11.xlib.GetKeyboardMapping = (PFN_XGetKeyboardMapping)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XGetKeyboardMapping");
    __glfw.x11.xlib.GetScreenSaver = (PFN_XGetScreenSaver)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XGetScreenSaver");
    __glfw.x11.xlib.GetSelectionOwner = (PFN_XGetSelectionOwner)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XGetSelectionOwner");
    __glfw.x11.xlib.GetVisualInfo = (PFN_XGetVisualInfo)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XGetVisualInfo");
    __glfw.x11.xlib.GetWMNormalHints = (PFN_XGetWMNormalHints)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XGetWMNormalHints");
    __glfw.x11.xlib.GetWindowAttributes = (PFN_XGetWindowAttributes)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XGetWindowAttributes");
    __glfw.x11.xlib.GetWindowProperty = (PFN_XGetWindowProperty)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XGetWindowProperty");
    __glfw.x11.xlib.GrabPointer = (PFN_XGrabPointer)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XGrabPointer");
    __glfw.x11.xlib.IconifyWindow = (PFN_XIconifyWindow)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XIconifyWindow");
    __glfw.x11.xlib.InternAtom = (PFN_XInternAtom)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XInternAtom");
    __glfw.x11.xlib.LookupString = (PFN_XLookupString)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XLookupString");
    __glfw.x11.xlib.MapRaised = (PFN_XMapRaised)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XMapRaised");
    __glfw.x11.xlib.MapWindow = (PFN_XMapWindow)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XMapWindow");
    __glfw.x11.xlib.MoveResizeWindow = (PFN_XMoveResizeWindow)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XMoveResizeWindow");
    __glfw.x11.xlib.MoveWindow = (PFN_XMoveWindow)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XMoveWindow");
    __glfw.x11.xlib.NextEvent = (PFN_XNextEvent)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XNextEvent");
    __glfw.x11.xlib.OpenIM = (PFN_XOpenIM)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XOpenIM");
    __glfw.x11.xlib.PeekEvent = (PFN_XPeekEvent)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XPeekEvent");
    __glfw.x11.xlib.Pending = (PFN_XPending)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XPending");
    __glfw.x11.xlib.QueryExtension = (PFN_XQueryExtension)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XQueryExtension");
    __glfw.x11.xlib.QueryPointer = (PFN_XQueryPointer)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XQueryPointer");
    __glfw.x11.xlib.RaiseWindow = (PFN_XRaiseWindow)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XRaiseWindow");
    __glfw.x11.xlib.RegisterIMInstantiateCallback = (PFN_XRegisterIMInstantiateCallback)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XRegisterIMInstantiateCallback");
    __glfw.x11.xlib.ResizeWindow = (PFN_XResizeWindow)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XResizeWindow");
    __glfw.x11.xlib.ResourceManagerString = (PFN_XResourceManagerString)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XResourceManagerString");
    __glfw.x11.xlib.SaveContext = (PFN_XSaveContext)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XSaveContext");
    __glfw.x11.xlib.SelectInput = (PFN_XSelectInput)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XSelectInput");
    __glfw.x11.xlib.SendEvent = (PFN_XSendEvent)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XSendEvent");
    __glfw.x11.xlib.SetClassHint = (PFN_XSetClassHint)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XSetClassHint");
    __glfw.x11.xlib.SetErrorHandler = (PFN_XSetErrorHandler)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XSetErrorHandler");
    __glfw.x11.xlib.SetICFocus = (PFN_XSetICFocus)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XSetICFocus");
    __glfw.x11.xlib.SetIMValues = (PFN_XSetIMValues)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XSetIMValues");
    __glfw.x11.xlib.SetInputFocus = (PFN_XSetInputFocus)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XSetInputFocus");
    __glfw.x11.xlib.SetLocaleModifiers = (PFN_XSetLocaleModifiers)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XSetLocaleModifiers");
    __glfw.x11.xlib.SetScreenSaver = (PFN_XSetScreenSaver)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XSetScreenSaver");
    __glfw.x11.xlib.SetSelectionOwner = (PFN_XSetSelectionOwner)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XSetSelectionOwner");
    __glfw.x11.xlib.SetWMHints = (PFN_XSetWMHints)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XSetWMHints");
    __glfw.x11.xlib.SetWMNormalHints = (PFN_XSetWMNormalHints)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XSetWMNormalHints");
    __glfw.x11.xlib.SetWMProtocols = (PFN_XSetWMProtocols)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XSetWMProtocols");
    __glfw.x11.xlib.SupportsLocale = (PFN_XSupportsLocale)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XSupportsLocale");
    __glfw.x11.xlib.Sync = (PFN_XSync)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XSync");
    __glfw.x11.xlib.TranslateCoordinates = (PFN_XTranslateCoordinates)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XTranslateCoordinates");
    __glfw.x11.xlib.UndefineCursor = (PFN_XUndefineCursor)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XUndefineCursor");
    __glfw.x11.xlib.UngrabPointer = (PFN_XUngrabPointer)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XUngrabPointer");
    __glfw.x11.xlib.UnmapWindow = (PFN_XUnmapWindow)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XUnmapWindow");
    __glfw.x11.xlib.UnsetICFocus = (PFN_XUnsetICFocus)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XUnsetICFocus");
    __glfw.x11.xlib.VisualIDFromVisual = (PFN_XVisualIDFromVisual)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XVisualIDFromVisual");
    __glfw.x11.xlib.WarpPointer = (PFN_XWarpPointer)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XWarpPointer");
    __glfw.x11.xkb.FreeKeyboard = (PFN_XkbFreeKeyboard)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XkbFreeKeyboard");
    __glfw.x11.xkb.FreeNames = (PFN_XkbFreeNames)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XkbFreeNames");
    __glfw.x11.xkb.GetMap = (PFN_XkbGetMap)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XkbGetMap");
    __glfw.x11.xkb.GetNames = (PFN_XkbGetNames)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XkbGetNames");
    __glfw.x11.xkb.GetState = (PFN_XkbGetState)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XkbGetState");
    __glfw.x11.xkb.KeycodeToKeysym = (PFN_XkbKeycodeToKeysym)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XkbKeycodeToKeysym");
    __glfw.x11.xkb.QueryExtension = (PFN_XkbQueryExtension)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XkbQueryExtension");
    __glfw.x11.xkb.SelectEventDetails = (PFN_XkbSelectEventDetails)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XkbSelectEventDetails");
    __glfw.x11.xkb.SetDetectableAutoRepeat = (PFN_XkbSetDetectableAutoRepeat)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XkbSetDetectableAutoRepeat");
    __glfw.x11.xrm.DestroyDatabase = (PFN_XrmDestroyDatabase)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XrmDestroyDatabase");
    __glfw.x11.xrm.GetResource = (PFN_XrmGetResource)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XrmGetResource");
    __glfw.x11.xrm.GetStringDatabase = (PFN_XrmGetStringDatabase)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XrmGetStringDatabase");
    __glfw.x11.xrm.UniqueQuark = (PFN_XrmUniqueQuark)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XrmUniqueQuark");
    __glfw.x11.xlib.UnregisterIMInstantiateCallback = (PFN_XUnregisterIMInstantiateCallback)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "XUnregisterIMInstantiateCallback");
    __glfw.x11.xlib.utf8LookupString = (PFN_Xutf8LookupString)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "Xutf8LookupString");
    __glfw.x11.xlib.utf8SetWMProperties = (PFN_Xutf8SetWMProperties)
        __glfwPlatformGetModuleSymbol(__glfw.x11.xlib.handle, "Xutf8SetWMProperties");

    if (__glfw.x11.xlib.utf8LookupString && __glfw.x11.xlib.utf8SetWMProperties)
        __glfw.x11.xlib.utf8 = GLFW_TRUE;

    __glfw.x11.screen = DefaultScreen(__glfw.x11.display);
    __glfw.x11.root = RootWindow(__glfw.x11.display, __glfw.x11.screen);
    __glfw.x11.context = XUniqueContext();

    getSystemContentScale(&__glfw.x11.contentScaleX, &__glfw.x11.contentScaleY);

    if (!createEmptyEventPipe())
        return GLFW_FALSE;

    if (!initExtensions())
        return GLFW_FALSE;

    __glfw.x11.helperWindowHandle = createHelperWindow();
    __glfw.x11.hiddenCursorHandle = createHiddenCursor();

    if (XSupportsLocale() && __glfw.x11.xlib.utf8)
    {
        XSetLocaleModifiers("");

        // If an IM is already present our callback will be called right away
        XRegisterIMInstantiateCallback(__glfw.x11.display,
                                       NULL, NULL, NULL,
                                       inputMethodInstantiateCallback,
                                       NULL);
    }

    ___glfwPollMonitorsX11();
    return GLFW_TRUE;
}

void ___glfwTerminateX11(void)
{
    if (__glfw.x11.helperWindowHandle)
    {
        if (XGetSelectionOwner(__glfw.x11.display, __glfw.x11.CLIPBOARD) ==
            __glfw.x11.helperWindowHandle)
        {
            ___glfwPushSelectionToManagerX11();
        }

        XDestroyWindow(__glfw.x11.display, __glfw.x11.helperWindowHandle);
        __glfw.x11.helperWindowHandle = None;
    }

    if (__glfw.x11.hiddenCursorHandle)
    {
        XFreeCursor(__glfw.x11.display, __glfw.x11.hiddenCursorHandle);
        __glfw.x11.hiddenCursorHandle = (Cursor) 0;
    }

    __glfw_free(__glfw.x11.primarySelectionString);
    __glfw_free(__glfw.x11.clipboardString);

    XUnregisterIMInstantiateCallback(__glfw.x11.display,
                                     NULL, NULL, NULL,
                                     inputMethodInstantiateCallback,
                                     NULL);

    if (__glfw.x11.im)
    {
        XCloseIM(__glfw.x11.im);
        __glfw.x11.im = NULL;
    }

    if (__glfw.x11.display)
    {
        XCloseDisplay(__glfw.x11.display);
        __glfw.x11.display = NULL;
    }

    if (__glfw.x11.x11xcb.handle)
    {
        __glfwPlatformFreeModule(__glfw.x11.x11xcb.handle);
        __glfw.x11.x11xcb.handle = NULL;
    }

    if (__glfw.x11.xcursor.handle)
    {
        __glfwPlatformFreeModule(__glfw.x11.xcursor.handle);
        __glfw.x11.xcursor.handle = NULL;
    }

    if (__glfw.x11.randr.handle)
    {
        __glfwPlatformFreeModule(__glfw.x11.randr.handle);
        __glfw.x11.randr.handle = NULL;
    }

    if (__glfw.x11.xinerama.handle)
    {
        __glfwPlatformFreeModule(__glfw.x11.xinerama.handle);
        __glfw.x11.xinerama.handle = NULL;
    }

    if (__glfw.x11.xrender.handle)
    {
        __glfwPlatformFreeModule(__glfw.x11.xrender.handle);
        __glfw.x11.xrender.handle = NULL;
    }

    if (__glfw.x11.vidmode.handle)
    {
        __glfwPlatformFreeModule(__glfw.x11.vidmode.handle);
        __glfw.x11.vidmode.handle = NULL;
    }

    if (__glfw.x11.xi.handle)
    {
        __glfwPlatformFreeModule(__glfw.x11.xi.handle);
        __glfw.x11.xi.handle = NULL;
    }

    ____glfwTerminateOSMesa();
    // NOTE: These need to be unloaded after XCloseDisplay, as they register
    //       cleanup callbacks that get called by that function
    ____glfwTerminateEGL();
    ____glfwTerminateGLX();

    if (__glfw.x11.xlib.handle)
    {
        __glfwPlatformFreeModule(__glfw.x11.xlib.handle);
        __glfw.x11.xlib.handle = NULL;
    }

    if (__glfw.x11.emptyEventPipe[0] || __glfw.x11.emptyEventPipe[1])
    {
        close(__glfw.x11.emptyEventPipe[0]);
        close(__glfw.x11.emptyEventPipe[1]);
    }
}

