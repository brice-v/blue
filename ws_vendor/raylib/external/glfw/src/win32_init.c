//========================================================================
// GLFW 3.4 Win32 - www.glfw.org
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

#include <stdlib.h>

static const GUID __glfw_GUID_DEVINTERFACE_HID =
    {0x4d1e55b2,0xf16f,0x11cf,{0x88,0xcb,0x00,0x11,0x11,0x00,0x00,0x30}};

#define GUID_DEVINTERFACE_HID __glfw_GUID_DEVINTERFACE_HID

#if defined(_GLFW_USE_HYBRID_HPG) || defined(_GLFW_USE_OPTIMUS_HPG)

#if defined(_GLFW_BUILD_DLL)
 #pragma message("These symbols must be exported by the executable and have no effect in a DLL")
#endif

// Executables (but not DLLs) exporting this symbol with this value will be
// automatically directed to the high-performance GPU on Nvidia Optimus systems
// with up-to-date drivers
//
__declspec(dllexport) DWORD NvOptimusEnablement = 1;

// Executables (but not DLLs) exporting this symbol with this value will be
// automatically directed to the high-performance GPU on AMD PowerXpress systems
// with up-to-date drivers
//
__declspec(dllexport) int AmdPowerXpressRequestHighPerformance = 1;

#endif // _GLFW_USE_HYBRID_HPG

#if defined(_GLFW_BUILD_DLL)

// GLFW DLL entry point
//
BOOL WINAPI DllMain(HINSTANCE instance, DWORD reason, LPVOID reserved)
{
    return TRUE;
}

#endif // _GLFW_BUILD_DLL

// Load necessary libraries (DLLs)
//
static GLFWbool loadLibraries(void)
{
    if (!GetModuleHandleExW(GET_MODULE_HANDLE_EX_FLAG_FROM_ADDRESS |
                                GET_MODULE_HANDLE_EX_FLAG_UNCHANGED_REFCOUNT,
                            (const WCHAR*) &__glfw,
                            (HMODULE*) &__glfw.win32.instance))
    {
        ___glfwInputErrorWin32(GLFW_PLATFORM_ERROR,
                             "Win32: Failed to retrieve own module handle");
        return GLFW_FALSE;
    }

    __glfw.win32.user32.instance = __glfwPlatformLoadModule("user32.dll");
    if (!__glfw.win32.user32.instance)
    {
        ___glfwInputErrorWin32(GLFW_PLATFORM_ERROR,
                             "Win32: Failed to load user32.dll");
        return GLFW_FALSE;
    }

    __glfw.win32.user32.SetProcessDPIAware_ = (PFN_SetProcessDPIAware)
        __glfwPlatformGetModuleSymbol(__glfw.win32.user32.instance, "SetProcessDPIAware");
    __glfw.win32.user32.ChangeWindowMessageFilterEx_ = (PFN_ChangeWindowMessageFilterEx)
        __glfwPlatformGetModuleSymbol(__glfw.win32.user32.instance, "ChangeWindowMessageFilterEx");
    __glfw.win32.user32.EnableNonClientDpiScaling_ = (PFN_EnableNonClientDpiScaling)
        __glfwPlatformGetModuleSymbol(__glfw.win32.user32.instance, "EnableNonClientDpiScaling");
    __glfw.win32.user32.SetProcessDpiAwarenessContext_ = (PFN_SetProcessDpiAwarenessContext)
        __glfwPlatformGetModuleSymbol(__glfw.win32.user32.instance, "SetProcessDpiAwarenessContext");
    __glfw.win32.user32.GetDpiForWindow_ = (PFN_GetDpiForWindow)
        __glfwPlatformGetModuleSymbol(__glfw.win32.user32.instance, "GetDpiForWindow");
    __glfw.win32.user32.AdjustWindowRectExForDpi_ = (PFN_AdjustWindowRectExForDpi)
        __glfwPlatformGetModuleSymbol(__glfw.win32.user32.instance, "AdjustWindowRectExForDpi");
    __glfw.win32.user32.GetSystemMetricsForDpi_ = (PFN_GetSystemMetricsForDpi)
        __glfwPlatformGetModuleSymbol(__glfw.win32.user32.instance, "GetSystemMetricsForDpi");

    __glfw.win32.dinput8.instance = __glfwPlatformLoadModule("dinput8.dll");
    if (__glfw.win32.dinput8.instance)
    {
        __glfw.win32.dinput8.Create = (PFN_DirectInput8Create)
            __glfwPlatformGetModuleSymbol(__glfw.win32.dinput8.instance, "DirectInput8Create");
    }

    {
        int i;
        const char* names[] =
        {
            "xinput1_4.dll",
            "xinput1_3.dll",
            "xinput9_1_0.dll",
            "xinput1_2.dll",
            "xinput1_1.dll",
            NULL
        };

        for (i = 0;  names[i];  i++)
        {
            __glfw.win32.xinput.instance = __glfwPlatformLoadModule(names[i]);
            if (__glfw.win32.xinput.instance)
            {
                __glfw.win32.xinput.GetCapabilities = (PFN_XInputGetCapabilities)
                    __glfwPlatformGetModuleSymbol(__glfw.win32.xinput.instance, "XInputGetCapabilities");
                __glfw.win32.xinput.GetState = (PFN_XInputGetState)
                    __glfwPlatformGetModuleSymbol(__glfw.win32.xinput.instance, "XInputGetState");

                break;
            }
        }
    }

    __glfw.win32.dwmapi.instance = __glfwPlatformLoadModule("dwmapi.dll");
    if (__glfw.win32.dwmapi.instance)
    {
        __glfw.win32.dwmapi.IsCompositionEnabled = (PFN_DwmIsCompositionEnabled)
            __glfwPlatformGetModuleSymbol(__glfw.win32.dwmapi.instance, "DwmIsCompositionEnabled");
        __glfw.win32.dwmapi.Flush = (PFN_DwmFlush)
            __glfwPlatformGetModuleSymbol(__glfw.win32.dwmapi.instance, "DwmFlush");
        __glfw.win32.dwmapi.EnableBlurBehindWindow = (PFN_DwmEnableBlurBehindWindow)
            __glfwPlatformGetModuleSymbol(__glfw.win32.dwmapi.instance, "DwmEnableBlurBehindWindow");
        __glfw.win32.dwmapi.GetColorizationColor = (PFN_DwmGetColorizationColor)
            __glfwPlatformGetModuleSymbol(__glfw.win32.dwmapi.instance, "DwmGetColorizationColor");
    }

    __glfw.win32.shcore.instance = __glfwPlatformLoadModule("shcore.dll");
    if (__glfw.win32.shcore.instance)
    {
        __glfw.win32.shcore.SetProcessDpiAwareness_ = (PFN_SetProcessDpiAwareness)
            __glfwPlatformGetModuleSymbol(__glfw.win32.shcore.instance, "SetProcessDpiAwareness");
        __glfw.win32.shcore.GetDpiForMonitor_ = (PFN_GetDpiForMonitor)
            __glfwPlatformGetModuleSymbol(__glfw.win32.shcore.instance, "GetDpiForMonitor");
    }

    __glfw.win32.ntdll.instance = __glfwPlatformLoadModule("ntdll.dll");
    if (__glfw.win32.ntdll.instance)
    {
        __glfw.win32.ntdll.RtlVerifyVersionInfo_ = (PFN_RtlVerifyVersionInfo)
            __glfwPlatformGetModuleSymbol(__glfw.win32.ntdll.instance, "RtlVerifyVersionInfo");
    }

    return GLFW_TRUE;
}

// Unload used libraries (DLLs)
//
static void freeLibraries(void)
{
    if (__glfw.win32.xinput.instance)
        __glfwPlatformFreeModule(__glfw.win32.xinput.instance);

    if (__glfw.win32.dinput8.instance)
        __glfwPlatformFreeModule(__glfw.win32.dinput8.instance);

    if (__glfw.win32.user32.instance)
        __glfwPlatformFreeModule(__glfw.win32.user32.instance);

    if (__glfw.win32.dwmapi.instance)
        __glfwPlatformFreeModule(__glfw.win32.dwmapi.instance);

    if (__glfw.win32.shcore.instance)
        __glfwPlatformFreeModule(__glfw.win32.shcore.instance);

    if (__glfw.win32.ntdll.instance)
        __glfwPlatformFreeModule(__glfw.win32.ntdll.instance);
}

// Create key code translation tables
//
static void createKeyTables(void)
{
    int scancode;

    memset(__glfw.win32.keycodes, -1, sizeof(__glfw.win32.keycodes));
    memset(__glfw.win32.scancodes, -1, sizeof(__glfw.win32.scancodes));

    __glfw.win32.keycodes[0x00B] = GLFW_KEY_0;
    __glfw.win32.keycodes[0x002] = GLFW_KEY_1;
    __glfw.win32.keycodes[0x003] = GLFW_KEY_2;
    __glfw.win32.keycodes[0x004] = GLFW_KEY_3;
    __glfw.win32.keycodes[0x005] = GLFW_KEY_4;
    __glfw.win32.keycodes[0x006] = GLFW_KEY_5;
    __glfw.win32.keycodes[0x007] = GLFW_KEY_6;
    __glfw.win32.keycodes[0x008] = GLFW_KEY_7;
    __glfw.win32.keycodes[0x009] = GLFW_KEY_8;
    __glfw.win32.keycodes[0x00A] = GLFW_KEY_9;
    __glfw.win32.keycodes[0x01E] = GLFW_KEY_A;
    __glfw.win32.keycodes[0x030] = GLFW_KEY_B;
    __glfw.win32.keycodes[0x02E] = GLFW_KEY_C;
    __glfw.win32.keycodes[0x020] = GLFW_KEY_D;
    __glfw.win32.keycodes[0x012] = GLFW_KEY_E;
    __glfw.win32.keycodes[0x021] = GLFW_KEY_F;
    __glfw.win32.keycodes[0x022] = GLFW_KEY_G;
    __glfw.win32.keycodes[0x023] = GLFW_KEY_H;
    __glfw.win32.keycodes[0x017] = GLFW_KEY_I;
    __glfw.win32.keycodes[0x024] = GLFW_KEY_J;
    __glfw.win32.keycodes[0x025] = GLFW_KEY_K;
    __glfw.win32.keycodes[0x026] = GLFW_KEY_L;
    __glfw.win32.keycodes[0x032] = GLFW_KEY_M;
    __glfw.win32.keycodes[0x031] = GLFW_KEY_N;
    __glfw.win32.keycodes[0x018] = GLFW_KEY_O;
    __glfw.win32.keycodes[0x019] = GLFW_KEY_P;
    __glfw.win32.keycodes[0x010] = GLFW_KEY_Q;
    __glfw.win32.keycodes[0x013] = GLFW_KEY_R;
    __glfw.win32.keycodes[0x01F] = GLFW_KEY_S;
    __glfw.win32.keycodes[0x014] = GLFW_KEY_T;
    __glfw.win32.keycodes[0x016] = GLFW_KEY_U;
    __glfw.win32.keycodes[0x02F] = GLFW_KEY_V;
    __glfw.win32.keycodes[0x011] = GLFW_KEY_W;
    __glfw.win32.keycodes[0x02D] = GLFW_KEY_X;
    __glfw.win32.keycodes[0x015] = GLFW_KEY_Y;
    __glfw.win32.keycodes[0x02C] = GLFW_KEY_Z;

    __glfw.win32.keycodes[0x028] = GLFW_KEY_APOSTROPHE;
    __glfw.win32.keycodes[0x02B] = GLFW_KEY_BACKSLASH;
    __glfw.win32.keycodes[0x033] = GLFW_KEY_COMMA;
    __glfw.win32.keycodes[0x00D] = GLFW_KEY_EQUAL;
    __glfw.win32.keycodes[0x029] = GLFW_KEY_GRAVE_ACCENT;
    __glfw.win32.keycodes[0x01A] = GLFW_KEY_LEFT_BRACKET;
    __glfw.win32.keycodes[0x00C] = GLFW_KEY_MINUS;
    __glfw.win32.keycodes[0x034] = GLFW_KEY_PERIOD;
    __glfw.win32.keycodes[0x01B] = GLFW_KEY_RIGHT_BRACKET;
    __glfw.win32.keycodes[0x027] = GLFW_KEY_SEMICOLON;
    __glfw.win32.keycodes[0x035] = GLFW_KEY_SLASH;
    __glfw.win32.keycodes[0x056] = GLFW_KEY_WORLD_2;

    __glfw.win32.keycodes[0x00E] = GLFW_KEY_BACKSPACE;
    __glfw.win32.keycodes[0x153] = GLFW_KEY_DELETE;
    __glfw.win32.keycodes[0x14F] = GLFW_KEY_END;
    __glfw.win32.keycodes[0x01C] = GLFW_KEY_ENTER;
    __glfw.win32.keycodes[0x001] = GLFW_KEY_ESCAPE;
    __glfw.win32.keycodes[0x147] = GLFW_KEY_HOME;
    __glfw.win32.keycodes[0x152] = GLFW_KEY_INSERT;
    __glfw.win32.keycodes[0x15D] = GLFW_KEY_MENU;
    __glfw.win32.keycodes[0x151] = GLFW_KEY_PAGE_DOWN;
    __glfw.win32.keycodes[0x149] = GLFW_KEY_PAGE_UP;
    __glfw.win32.keycodes[0x045] = GLFW_KEY_PAUSE;
    __glfw.win32.keycodes[0x039] = GLFW_KEY_SPACE;
    __glfw.win32.keycodes[0x00F] = GLFW_KEY_TAB;
    __glfw.win32.keycodes[0x03A] = GLFW_KEY_CAPS_LOCK;
    __glfw.win32.keycodes[0x145] = GLFW_KEY_NUM_LOCK;
    __glfw.win32.keycodes[0x046] = GLFW_KEY_SCROLL_LOCK;
    __glfw.win32.keycodes[0x03B] = GLFW_KEY_F1;
    __glfw.win32.keycodes[0x03C] = GLFW_KEY_F2;
    __glfw.win32.keycodes[0x03D] = GLFW_KEY_F3;
    __glfw.win32.keycodes[0x03E] = GLFW_KEY_F4;
    __glfw.win32.keycodes[0x03F] = GLFW_KEY_F5;
    __glfw.win32.keycodes[0x040] = GLFW_KEY_F6;
    __glfw.win32.keycodes[0x041] = GLFW_KEY_F7;
    __glfw.win32.keycodes[0x042] = GLFW_KEY_F8;
    __glfw.win32.keycodes[0x043] = GLFW_KEY_F9;
    __glfw.win32.keycodes[0x044] = GLFW_KEY_F10;
    __glfw.win32.keycodes[0x057] = GLFW_KEY_F11;
    __glfw.win32.keycodes[0x058] = GLFW_KEY_F12;
    __glfw.win32.keycodes[0x064] = GLFW_KEY_F13;
    __glfw.win32.keycodes[0x065] = GLFW_KEY_F14;
    __glfw.win32.keycodes[0x066] = GLFW_KEY_F15;
    __glfw.win32.keycodes[0x067] = GLFW_KEY_F16;
    __glfw.win32.keycodes[0x068] = GLFW_KEY_F17;
    __glfw.win32.keycodes[0x069] = GLFW_KEY_F18;
    __glfw.win32.keycodes[0x06A] = GLFW_KEY_F19;
    __glfw.win32.keycodes[0x06B] = GLFW_KEY_F20;
    __glfw.win32.keycodes[0x06C] = GLFW_KEY_F21;
    __glfw.win32.keycodes[0x06D] = GLFW_KEY_F22;
    __glfw.win32.keycodes[0x06E] = GLFW_KEY_F23;
    __glfw.win32.keycodes[0x076] = GLFW_KEY_F24;
    __glfw.win32.keycodes[0x038] = GLFW_KEY_LEFT_ALT;
    __glfw.win32.keycodes[0x01D] = GLFW_KEY_LEFT_CONTROL;
    __glfw.win32.keycodes[0x02A] = GLFW_KEY_LEFT_SHIFT;
    __glfw.win32.keycodes[0x15B] = GLFW_KEY_LEFT_SUPER;
    __glfw.win32.keycodes[0x137] = GLFW_KEY_PRINT_SCREEN;
    __glfw.win32.keycodes[0x138] = GLFW_KEY_RIGHT_ALT;
    __glfw.win32.keycodes[0x11D] = GLFW_KEY_RIGHT_CONTROL;
    __glfw.win32.keycodes[0x036] = GLFW_KEY_RIGHT_SHIFT;
    __glfw.win32.keycodes[0x15C] = GLFW_KEY_RIGHT_SUPER;
    __glfw.win32.keycodes[0x150] = GLFW_KEY_DOWN;
    __glfw.win32.keycodes[0x14B] = GLFW_KEY_LEFT;
    __glfw.win32.keycodes[0x14D] = GLFW_KEY_RIGHT;
    __glfw.win32.keycodes[0x148] = GLFW_KEY_UP;

    __glfw.win32.keycodes[0x052] = GLFW_KEY_KP_0;
    __glfw.win32.keycodes[0x04F] = GLFW_KEY_KP_1;
    __glfw.win32.keycodes[0x050] = GLFW_KEY_KP_2;
    __glfw.win32.keycodes[0x051] = GLFW_KEY_KP_3;
    __glfw.win32.keycodes[0x04B] = GLFW_KEY_KP_4;
    __glfw.win32.keycodes[0x04C] = GLFW_KEY_KP_5;
    __glfw.win32.keycodes[0x04D] = GLFW_KEY_KP_6;
    __glfw.win32.keycodes[0x047] = GLFW_KEY_KP_7;
    __glfw.win32.keycodes[0x048] = GLFW_KEY_KP_8;
    __glfw.win32.keycodes[0x049] = GLFW_KEY_KP_9;
    __glfw.win32.keycodes[0x04E] = GLFW_KEY_KP_ADD;
    __glfw.win32.keycodes[0x053] = GLFW_KEY_KP_DECIMAL;
    __glfw.win32.keycodes[0x135] = GLFW_KEY_KP_DIVIDE;
    __glfw.win32.keycodes[0x11C] = GLFW_KEY_KP_ENTER;
    __glfw.win32.keycodes[0x059] = GLFW_KEY_KP_EQUAL;
    __glfw.win32.keycodes[0x037] = GLFW_KEY_KP_MULTIPLY;
    __glfw.win32.keycodes[0x04A] = GLFW_KEY_KP_SUBTRACT;

    for (scancode = 0;  scancode < 512;  scancode++)
    {
        if (__glfw.win32.keycodes[scancode] > 0)
            __glfw.win32.scancodes[__glfw.win32.keycodes[scancode]] = scancode;
    }
}

// Window procedure for the hidden helper window
//
static LRESULT CALLBACK helperWindowProc(HWND hWnd, UINT uMsg, WPARAM wParam, LPARAM lParam)
{
    switch (uMsg)
    {
        case WM_DISPLAYCHANGE:
            __glfwPollMonitorsWin32();
            break;

        case WM_DEVICECHANGE:
        {
            if (!__glfw.joysticksInitialized)
                break;

            if (wParam == DBT_DEVICEARRIVAL)
            {
                DEV_BROADCAST_HDR* dbh = (DEV_BROADCAST_HDR*) lParam;
                if (dbh && dbh->dbch_devicetype == DBT_DEVTYP_DEVICEINTERFACE)
                    __glfwDetectJoystickConnectionWin32();
            }
            else if (wParam == DBT_DEVICEREMOVECOMPLETE)
            {
                DEV_BROADCAST_HDR* dbh = (DEV_BROADCAST_HDR*) lParam;
                if (dbh && dbh->dbch_devicetype == DBT_DEVTYP_DEVICEINTERFACE)
                    __glfwDetectJoystickDisconnectionWin32();
            }

            break;
        }
    }

    return DefWindowProcW(hWnd, uMsg, wParam, lParam);
}

// Creates a dummy window for behind-the-scenes work
//
static GLFWbool createHelperWindow(void)
{
    MSG msg;
    WNDCLASSEXW wc = { sizeof(wc) };

    wc.style         = CS_OWNDC;
    wc.lpfnWndProc   = (WNDPROC) helperWindowProc;
    wc.hInstance     = __glfw.win32.instance;
    wc.lpszClassName = L"GLFW3 Helper";

    __glfw.win32.helperWindowClass = RegisterClassExW(&wc);
    if (!__glfw.win32.helperWindowClass)
    {
        ___glfwInputErrorWin32(GLFW_PLATFORM_ERROR,
                             "WIn32: Failed to register helper window class");
        return GLFW_FALSE;
    }

    __glfw.win32.helperWindowHandle =
        CreateWindowExW(WS_EX_OVERLAPPEDWINDOW,
                        MAKEINTATOM(__glfw.win32.helperWindowClass),
                        L"GLFW message window",
                        WS_CLIPSIBLINGS | WS_CLIPCHILDREN,
                        0, 0, 1, 1,
                        NULL, NULL,
                        __glfw.win32.instance,
                        NULL);

    if (!__glfw.win32.helperWindowHandle)
    {
        ___glfwInputErrorWin32(GLFW_PLATFORM_ERROR,
                             "Win32: Failed to create helper window");
        return GLFW_FALSE;
    }

    // HACK: The command to the first ShowWindow call is ignored if the parent
    //       process passed along a STARTUPINFO, so clear that with a no-op call
    ShowWindow(__glfw.win32.helperWindowHandle, SW_HIDE);

    // Register for HID device notifications
    {
        DEV_BROADCAST_DEVICEINTERFACE_W dbi;
        ZeroMemory(&dbi, sizeof(dbi));
        dbi.dbcc_size = sizeof(dbi);
        dbi.dbcc_devicetype = DBT_DEVTYP_DEVICEINTERFACE;
        dbi.dbcc_classguid = GUID_DEVINTERFACE_HID;

        __glfw.win32.deviceNotificationHandle =
            RegisterDeviceNotificationW(__glfw.win32.helperWindowHandle,
                                        (DEV_BROADCAST_HDR*) &dbi,
                                        DEVICE_NOTIFY_WINDOW_HANDLE);
    }

    while (PeekMessageW(&msg, __glfw.win32.helperWindowHandle, 0, 0, PM_REMOVE))
    {
        TranslateMessage(&msg);
        DispatchMessageW(&msg);
    }

   return GLFW_TRUE;
}


//////////////////////////////////////////////////////////////////////////
//////                       GLFW internal API                      //////
//////////////////////////////////////////////////////////////////////////

// Returns a wide string version of the specified UTF-8 string
//
WCHAR* __glfwCreateWideStringFromUTF8Win32(const char* source)
{
    WCHAR* target;
    int count;

    count = MultiByteToWideChar(CP_UTF8, 0, source, -1, NULL, 0);
    if (!count)
    {
        ___glfwInputErrorWin32(GLFW_PLATFORM_ERROR,
                             "Win32: Failed to convert string from UTF-8");
        return NULL;
    }

    target = __glfw_calloc(count, sizeof(WCHAR));

    if (!MultiByteToWideChar(CP_UTF8, 0, source, -1, target, count))
    {
        ___glfwInputErrorWin32(GLFW_PLATFORM_ERROR,
                             "Win32: Failed to convert string from UTF-8");
        __glfw_free(target);
        return NULL;
    }

    return target;
}

// Returns a UTF-8 string version of the specified wide string
//
char* __glfwCreateUTF8FromWideStringWin32(const WCHAR* source)
{
    char* target;
    int size;

    size = WideCharToMultiByte(CP_UTF8, 0, source, -1, NULL, 0, NULL, NULL);
    if (!size)
    {
        ___glfwInputErrorWin32(GLFW_PLATFORM_ERROR,
                             "Win32: Failed to convert string to UTF-8");
        return NULL;
    }

    target = __glfw_calloc(size, 1);

    if (!WideCharToMultiByte(CP_UTF8, 0, source, -1, target, size, NULL, NULL))
    {
        ___glfwInputErrorWin32(GLFW_PLATFORM_ERROR,
                             "Win32: Failed to convert string to UTF-8");
        __glfw_free(target);
        return NULL;
    }

    return target;
}

// Reports the specified error, appending information about the last Win32 error
//
void ___glfwInputErrorWin32(int error, const char* description)
{
    WCHAR buffer[_GLFW_MESSAGE_SIZE] = L"";
    char message[_GLFW_MESSAGE_SIZE] = "";

    FormatMessageW(FORMAT_MESSAGE_FROM_SYSTEM |
                       FORMAT_MESSAGE_IGNORE_INSERTS |
                       FORMAT_MESSAGE_MAX_WIDTH_MASK,
                   NULL,
                   GetLastError() & 0xffff,
                   MAKELANGID(LANG_NEUTRAL, SUBLANG_DEFAULT),
                   buffer,
                   sizeof(buffer) / sizeof(WCHAR),
                   NULL);
    WideCharToMultiByte(CP_UTF8, 0, buffer, -1, message, sizeof(message), NULL, NULL);

    ___glfwInputError(error, "%s: %s", description, message);
}

// Updates key names according to the current keyboard layout
//
void __glfwUpdateKeyNamesWin32(void)
{
    int key;
    BYTE state[256] = {0};

    memset(__glfw.win32.keynames, 0, sizeof(__glfw.win32.keynames));

    for (key = GLFW_KEY_SPACE;  key <= GLFW_KEY_LAST;  key++)
    {
        UINT vk;
        int scancode, length;
        WCHAR chars[16];

        scancode = __glfw.win32.scancodes[key];
        if (scancode == -1)
            continue;

        if (key >= GLFW_KEY_KP_0 && key <= GLFW_KEY_KP_ADD)
        {
            const UINT vks[] = {
                VK_NUMPAD0,  VK_NUMPAD1,  VK_NUMPAD2, VK_NUMPAD3,
                VK_NUMPAD4,  VK_NUMPAD5,  VK_NUMPAD6, VK_NUMPAD7,
                VK_NUMPAD8,  VK_NUMPAD9,  VK_DECIMAL, VK_DIVIDE,
                VK_MULTIPLY, VK_SUBTRACT, VK_ADD
            };

            vk = vks[key - GLFW_KEY_KP_0];
        }
        else
            vk = MapVirtualKeyW(scancode, MAPVK_VSC_TO_VK);

        length = ToUnicode(vk, scancode, state,
                           chars, sizeof(chars) / sizeof(WCHAR),
                           0);

        if (length == -1)
        {
            // This is a dead key, so we need a second simulated key press
            // to make it output its own character (usually a diacritic)
            length = ToUnicode(vk, scancode, state,
                               chars, sizeof(chars) / sizeof(WCHAR),
                               0);
        }

        if (length < 1)
            continue;

        WideCharToMultiByte(CP_UTF8, 0, chars, 1,
                            __glfw.win32.keynames[key],
                            sizeof(__glfw.win32.keynames[key]),
                            NULL, NULL);
    }
}

// Replacement for IsWindowsVersionOrGreater, as we cannot rely on the
// application having a correct embedded manifest
//
BOOL __glfwIsWindowsVersionOrGreaterWin32(WORD major, WORD minor, WORD sp)
{
    OSVERSIONINFOEXW osvi = { sizeof(osvi), major, minor, 0, 0, {0}, sp };
    DWORD mask = VER_MAJORVERSION | VER_MINORVERSION | VER_SERVICEPACKMAJOR;
    ULONGLONG cond = VerSetConditionMask(0, VER_MAJORVERSION, VER_GREATER_EQUAL);
    cond = VerSetConditionMask(cond, VER_MINORVERSION, VER_GREATER_EQUAL);
    cond = VerSetConditionMask(cond, VER_SERVICEPACKMAJOR, VER_GREATER_EQUAL);
    // HACK: Use RtlVerifyVersionInfo instead of VerifyVersionInfoW as the
    //       latter lies unless the user knew to embed a non-default manifest
    //       announcing support for Windows 10 via supportedOS GUID
    return RtlVerifyVersionInfo(&osvi, mask, cond) == 0;
}

// Checks whether we are on at least the specified build of Windows 10
//
BOOL __glfwIsWindows10BuildOrGreaterWin32(WORD build)
{
    OSVERSIONINFOEXW osvi = { sizeof(osvi), 10, 0, build };
    DWORD mask = VER_MAJORVERSION | VER_MINORVERSION | VER_BUILDNUMBER;
    ULONGLONG cond = VerSetConditionMask(0, VER_MAJORVERSION, VER_GREATER_EQUAL);
    cond = VerSetConditionMask(cond, VER_MINORVERSION, VER_GREATER_EQUAL);
    cond = VerSetConditionMask(cond, VER_BUILDNUMBER, VER_GREATER_EQUAL);
    // HACK: Use RtlVerifyVersionInfo instead of VerifyVersionInfoW as the
    //       latter lies unless the user knew to embed a non-default manifest
    //       announcing support for Windows 10 via supportedOS GUID
    return RtlVerifyVersionInfo(&osvi, mask, cond) == 0;
}

GLFWbool __glfwConnectWin32(int platformID, _GLFWplatform* platform)
{
    const _GLFWplatform win32 =
    {
        GLFW_PLATFORM_WIN32,
        ___glfwInitWin32,
        ___glfwTerminateWin32,
        ___glfwGetCursorPosWin32,
        ____glfwSetCursorPosWin32,
        ___glfwSetCursorModeWin32,
        __glfwSetRawMouseMotionWin32,
        ___glfwRawMouseMotionSupportedWin32,
        ___glfwCreateCursorWin32,
        ___glfwCreateStandardCursorWin32,
        ___glfwDestroyCursorWin32,
        ___glfwSetCursorWin32,
        __glfwGetScancodeNameWin32,
        ____glfwGetKeyScancodeWin32,
        ___glfwSetClipboardStringWin32,
        ___glfwGetClipboardStringWin32,
        ___glfwInitJoysticksWin32,
        ___glfwTerminateJoysticksWin32,
        __glfwPollJoystickWin32,
        __glfwGetMappingNameWin32,
        __glfwUpdateGamepadGUIDWin32,
        ___glfwFreeMonitorWin32,
        ___glfwGetMonitorPosWin32,
        ___glfwGetMonitorContentScaleWin32,
        ___glfwGetMonitorWorkareaWin32,
        ____glfwGetVideoModesWin32,
        ___glfwGetVideoModeWin32,
        ___glfwGetGammaRampWin32,
        ____glfwSetGammaRampWin32,
        ___glfwCreateWindowWin32,
        ___glfwDestroyWindowWin32,
        ___glfwSetWindowTitleWin32,
        ___glfwSetWindowIconWin32,
        ___glfwGetWindowPosWin32,
        ___glfwSetWindowPosWin32,
        ___glfwGetWindowSizeWin32,
        ___glfwSetWindowSizeWin32,
        ____glfwSetWindowSizeLimitsWin32,
        ___glfwSetWindowAspectRatioWin32,
        ___glfwGetFramebufferSizeWin32,
        ___glfwGetWindowFrameSizeWin32,
        ___glfwGetWindowContentScaleWin32,
        ___glfwIconifyWindowWin32,
        ___glfwRestoreWindowWin32,
        ___glfwMaximizeWindowWin32,
        ___glfwShowWindowWin32,
        ___glfwHideWindowWin32,
        ___glfwRequestWindowAttentionWin32,
        ___glfwFocusWindowWin32,
        ___glfwSetWindowMonitorWin32,
        __glfwWindowFocusedWin32,
        __glfwWindowIconifiedWin32,
        __glfwWindowVisibleWin32,
        __glfwWindowMaximizedWin32,
        __glfwWindowHoveredWin32,
        __glfwFramebufferTransparentWin32,
        ___glfwGetWindowOpacityWin32,
        __glfwSetWindowResizableWin32,
        __glfwSetWindowDecoratedWin32,
        __glfwSetWindowFloatingWin32,
        ___glfwSetWindowOpacityWin32,
        __glfwSetWindowMousePassthroughWin32,
        ___glfwPollEventsWin32,
        ___glfwWaitEventsWin32,
        ____glfwWaitEventsTimeoutWin32,
        ___glfwPostEmptyEventWin32,
        __glfwGetEGLPlatformWin32,
        __glfwGetEGLNativeDisplayWin32,
        __glfwGetEGLNativeWindowWin32,
        ___glfwGetRequiredInstanceExtensionsWin32,
        ___glfwGetPhysicalDevicePresentationSupportWin32,
        ____glfwCreateWindowSurfaceWin32,
    };

    *platform = win32;
    return GLFW_TRUE;
}

int ___glfwInitWin32(void)
{
    if (!loadLibraries())
        return GLFW_FALSE;

    createKeyTables();
    __glfwUpdateKeyNamesWin32();

    if (__glfwIsWindows10Version1703OrGreaterWin32())
        SetProcessDpiAwarenessContext(DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2);
    else if (IsWindows8Point1OrGreater())
        SetProcessDpiAwareness(PROCESS_PER_MONITOR_DPI_AWARE);
    else if (IsWindowsVistaOrGreater())
        SetProcessDPIAware();

    if (!createHelperWindow())
        return GLFW_FALSE;

    __glfwPollMonitorsWin32();
    return GLFW_TRUE;
}

void ___glfwTerminateWin32(void)
{
    if (__glfw.win32.deviceNotificationHandle)
        UnregisterDeviceNotification(__glfw.win32.deviceNotificationHandle);

    if (__glfw.win32.helperWindowHandle)
        DestroyWindow(__glfw.win32.helperWindowHandle);
    if (__glfw.win32.helperWindowClass)
        UnregisterClassW(MAKEINTATOM(__glfw.win32.helperWindowClass), __glfw.win32.instance);
    if (__glfw.win32.mainWindowClass)
        UnregisterClassW(MAKEINTATOM(__glfw.win32.mainWindowClass), __glfw.win32.instance);

    __glfw_free(__glfw.win32.clipboardString);
    __glfw_free(__glfw.win32.rawInput);

    ___glfwTerminateWGL();
    ____glfwTerminateEGL();
    ____glfwTerminateOSMesa();

    freeLibraries();
}

