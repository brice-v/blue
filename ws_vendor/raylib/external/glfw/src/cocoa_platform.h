//========================================================================
// GLFW 3.4 macOS - www.glfw.org
//------------------------------------------------------------------------
// Copyright (c) 2009-2019 Camilla Löwy <elmindreda@glfw.org>
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

#include <stdint.h>

#include <Carbon/Carbon.h>
#include <IOKit/hid/IOHIDLib.h>

// NOTE: All of NSGL was deprecated in the 10.14 SDK
//       This disables the pointless warnings for every symbol we use
#ifndef GL_SILENCE_DEPRECATION
#define GL_SILENCE_DEPRECATION
#endif

#if defined(__OBJC__)
#import <Cocoa/Cocoa.h>
#else
typedef void* id;
#endif

// NOTE: Many Cocoa enum values have been renamed and we need to build across
//       SDK versions where one is unavailable or deprecated.
//       We use the newer names in code and replace them with the older names if
//       the base SDK does not provide the newer names.

#if MAC_OS_X_VERSION_MAX_ALLOWED < 101400
 #define NSOpenGLContextParameterSwapInterval NSOpenGLCPSwapInterval
 #define NSOpenGLContextParameterSurfaceOpacity NSOpenGLCPSurfaceOpacity
#endif

#if MAC_OS_X_VERSION_MAX_ALLOWED < 101200
 #define NSBitmapFormatAlphaNonpremultiplied NSAlphaNonpremultipliedBitmapFormat
 #define NSEventMaskAny NSAnyEventMask
 #define NSEventMaskKeyUp NSKeyUpMask
 #define NSEventModifierFlagCapsLock NSAlphaShiftKeyMask
 #define NSEventModifierFlagCommand NSCommandKeyMask
 #define NSEventModifierFlagControl NSControlKeyMask
 #define NSEventModifierFlagDeviceIndependentFlagsMask NSDeviceIndependentModifierFlagsMask
 #define NSEventModifierFlagOption NSAlternateKeyMask
 #define NSEventModifierFlagShift NSShiftKeyMask
 #define NSEventTypeApplicationDefined NSApplicationDefined
 #define NSWindowStyleMaskBorderless NSBorderlessWindowMask
 #define NSWindowStyleMaskClosable NSClosableWindowMask
 #define NSWindowStyleMaskMiniaturizable NSMiniaturizableWindowMask
 #define NSWindowStyleMaskResizable NSResizableWindowMask
 #define NSWindowStyleMaskTitled NSTitledWindowMask
#endif

// NOTE: Many Cocoa dynamically linked constants have been renamed and we need
//       to build across SDK versions where one is unavailable or deprecated.
//       We use the newer names in code and replace them with the older names if
//       the deployment target is older than the newer names.

#if MAC_OS_X_VERSION_MIN_REQUIRED < 101300
 #define NSPasteboardTypeURL NSURLPboardType
#endif

typedef VkFlags VkMacOSSurfaceCreateFlagsMVK;
typedef VkFlags VkMetalSurfaceCreateFlagsEXT;

typedef struct VkMacOSSurfaceCreateInfoMVK
{
    VkStructureType                 sType;
    const void*                     pNext;
    VkMacOSSurfaceCreateFlagsMVK    flags;
    const void*                     pView;
} VkMacOSSurfaceCreateInfoMVK;

typedef struct VkMetalSurfaceCreateInfoEXT
{
    VkStructureType                 sType;
    const void*                     pNext;
    VkMetalSurfaceCreateFlagsEXT    flags;
    const void*                     pLayer;
} VkMetalSurfaceCreateInfoEXT;

typedef VkResult (APIENTRY *PFN_vkCreateMacOSSurfaceMVK)(VkInstance,const VkMacOSSurfaceCreateInfoMVK*,const VkAllocationCallbacks*,VkSurfaceKHR*);
typedef VkResult (APIENTRY *PFN_vkCreateMetalSurfaceEXT)(VkInstance,const VkMetalSurfaceCreateInfoEXT*,const VkAllocationCallbacks*,VkSurfaceKHR*);

#define GLFW_COCOA_WINDOW_STATE         _GLFWwindowNS  ns;
#define GLFW_COCOA_LIBRARY_WINDOW_STATE _GLFWlibraryNS ns;
#define GLFW_COCOA_MONITOR_STATE        _GLFWmonitorNS ns;
#define GLFW_COCOA_CURSOR_STATE         _GLFWcursorNS  ns;

#define GLFW_NSGL_CONTEXT_STATE         _GLFWcontextNSGL nsgl;
#define GLFW_NSGL_LIBRARY_CONTEXT_STATE _GLFWlibraryNSGL nsgl;

// HIToolbox.framework pointer typedefs
#define kTISPropertyUnicodeKeyLayoutData __glfw.ns.tis.kPropertyUnicodeKeyLayoutData
typedef TISInputSourceRef (*PFN_TISCopyCurrentKeyboardLayoutInputSource)(void);
#define TISCopyCurrentKeyboardLayoutInputSource __glfw.ns.tis.CopyCurrentKeyboardLayoutInputSource
typedef void* (*PFN_TISGetInputSourceProperty)(TISInputSourceRef,CFStringRef);
#define TISGetInputSourceProperty __glfw.ns.tis.GetInputSourceProperty
typedef UInt8 (*PFN_LMGetKbdType)(void);
#define LMGetKbdType __glfw.ns.tis.GetKbdType


// NSGL-specific per-context data
//
typedef struct _GLFWcontextNSGL
{
    id                pixelFormat;
    id                object;
} _GLFWcontextNSGL;

// NSGL-specific global data
//
typedef struct _GLFWlibraryNSGL
{
    // dlopen handle for OpenGL.framework (for __glfwGetProcAddress)
    CFBundleRef     framework;
} _GLFWlibraryNSGL;

// Cocoa-specific per-window data
//
typedef struct _GLFWwindowNS
{
    id              object;
    id              delegate;
    id              view;
    id              layer;

    GLFWbool        maximized;
    GLFWbool        occluded;
    GLFWbool        retina;

    // Cached window properties to filter out duplicate events
    int             width, height;
    int             fbWidth, fbHeight;
    float           xscale, yscale;

    // The total sum of the distances the cursor has been warped
    // since the last cursor motion event was processed
    // This is kept to counteract Cocoa doing the same internally
    double          cursorWarpDeltaX, cursorWarpDeltaY;
} _GLFWwindowNS;

// Cocoa-specific global data
//
typedef struct _GLFWlibraryNS
{
    CGEventSourceRef    eventSource;
    id                  delegate;
    GLFWbool            cursorHidden;
    TISInputSourceRef   inputSource;
    IOHIDManagerRef     hidManager;
    id                  unicodeData;
    id                  helper;
    id                  keyUpMonitor;
    id                  nibObjects;

    char                keynames[GLFW_KEY_LAST + 1][17];
    short int           keycodes[256];
    short int           scancodes[GLFW_KEY_LAST + 1];
    char*               clipboardString;
    CGPoint             cascadePoint;
    // Where to place the cursor when re-enabled
    double              restoreCursorPosX, restoreCursorPosY;
    // The window whose disabled cursor mode is active
    _GLFWwindow*        disabledCursorWindow;

    struct {
        CFBundleRef     bundle;
        PFN_TISCopyCurrentKeyboardLayoutInputSource CopyCurrentKeyboardLayoutInputSource;
        PFN_TISGetInputSourceProperty GetInputSourceProperty;
        PFN_LMGetKbdType GetKbdType;
        CFStringRef     kPropertyUnicodeKeyLayoutData;
    } tis;
} _GLFWlibraryNS;

// Cocoa-specific per-monitor data
//
typedef struct _GLFWmonitorNS
{
    CGDirectDisplayID   displayID;
    CGDisplayModeRef    previousMode;
    uint32_t            unitNumber;
    id                  screen;
    double              fallbackRefreshRate;
} _GLFWmonitorNS;

// Cocoa-specific per-cursor data
//
typedef struct _GLFWcursorNS
{
    id              object;
} _GLFWcursorNS;


GLFWbool __glfwConnectCocoa(int platformID, _GLFWplatform* platform);
int ___glfwInitCocoa(void);
void ___glfwTerminateCocoa(void);

GLFWbool ___glfwCreateWindowCocoa(_GLFWwindow* window, const _GLFWwndconfig* wndconfig, const _GLFWctxconfig* ctxconfig, const _GLFWfbconfig* fbconfig);
void ___glfwDestroyWindowCocoa(_GLFWwindow* window);
void ___glfwSetWindowTitleCocoa(_GLFWwindow* window, const char* title);
void ___glfwSetWindowIconCocoa(_GLFWwindow* window, int count, const GLFWimage* images);
void ___glfwGetWindowPosCocoa(_GLFWwindow* window, int* xpos, int* ypos);
void ___glfwSetWindowPosCocoa(_GLFWwindow* window, int xpos, int ypos);
void ___glfwGetWindowSizeCocoa(_GLFWwindow* window, int* width, int* height);
void ___glfwSetWindowSizeCocoa(_GLFWwindow* window, int width, int height);
void ____glfwSetWindowSizeLimitsCocoa(_GLFWwindow* window, int minwidth, int minheight, int maxwidth, int maxheight);
void ___glfwSetWindowAspectRatioCocoa(_GLFWwindow* window, int numer, int denom);
void ___glfwGetFramebufferSizeCocoa(_GLFWwindow* window, int* width, int* height);
void ___glfwGetWindowFrameSizeCocoa(_GLFWwindow* window, int* left, int* top, int* right, int* bottom);
void ___glfwGetWindowContentScaleCocoa(_GLFWwindow* window, float* xscale, float* yscale);
void ___glfwIconifyWindowCocoa(_GLFWwindow* window);
void ___glfwRestoreWindowCocoa(_GLFWwindow* window);
void ___glfwMaximizeWindowCocoa(_GLFWwindow* window);
void ___glfwShowWindowCocoa(_GLFWwindow* window);
void ___glfwHideWindowCocoa(_GLFWwindow* window);
void ___glfwRequestWindowAttentionCocoa(_GLFWwindow* window);
void ___glfwFocusWindowCocoa(_GLFWwindow* window);
void ___glfwSetWindowMonitorCocoa(_GLFWwindow* window, _GLFWmonitor* monitor, int xpos, int ypos, int width, int height, int refreshRate);
GLFWbool __glfwWindowFocusedCocoa(_GLFWwindow* window);
GLFWbool __glfwWindowIconifiedCocoa(_GLFWwindow* window);
GLFWbool __glfwWindowVisibleCocoa(_GLFWwindow* window);
GLFWbool __glfwWindowMaximizedCocoa(_GLFWwindow* window);
GLFWbool __glfwWindowHoveredCocoa(_GLFWwindow* window);
GLFWbool __glfwFramebufferTransparentCocoa(_GLFWwindow* window);
void __glfwSetWindowResizableCocoa(_GLFWwindow* window, GLFWbool enabled);
void __glfwSetWindowDecoratedCocoa(_GLFWwindow* window, GLFWbool enabled);
void __glfwSetWindowFloatingCocoa(_GLFWwindow* window, GLFWbool enabled);
float ___glfwGetWindowOpacityCocoa(_GLFWwindow* window);
void ___glfwSetWindowOpacityCocoa(_GLFWwindow* window, float opacity);
void __glfwSetWindowMousePassthroughCocoa(_GLFWwindow* window, GLFWbool enabled);

void __glfwSetRawMouseMotionCocoa(_GLFWwindow *window, GLFWbool enabled);
GLFWbool ___glfwRawMouseMotionSupportedCocoa(void);

void ___glfwPollEventsCocoa(void);
void ___glfwWaitEventsCocoa(void);
void ____glfwWaitEventsTimeoutCocoa(double timeout);
void ___glfwPostEmptyEventCocoa(void);

void ___glfwGetCursorPosCocoa(_GLFWwindow* window, double* xpos, double* ypos);
void ____glfwSetCursorPosCocoa(_GLFWwindow* window, double xpos, double ypos);
void ___glfwSetCursorModeCocoa(_GLFWwindow* window, int mode);
const char* __glfwGetScancodeNameCocoa(int scancode);
int ____glfwGetKeyScancodeCocoa(int key);
GLFWbool ___glfwCreateCursorCocoa(_GLFWcursor* cursor, const GLFWimage* image, int xhot, int yhot);
GLFWbool ___glfwCreateStandardCursorCocoa(_GLFWcursor* cursor, int shape);
void ___glfwDestroyCursorCocoa(_GLFWcursor* cursor);
void ___glfwSetCursorCocoa(_GLFWwindow* window, _GLFWcursor* cursor);
void ___glfwSetClipboardStringCocoa(const char* string);
const char* ___glfwGetClipboardStringCocoa(void);

EGLenum __glfwGetEGLPlatformCocoa(EGLint** attribs);
EGLNativeDisplayType __glfwGetEGLNativeDisplayCocoa(void);
EGLNativeWindowType __glfwGetEGLNativeWindowCocoa(_GLFWwindow* window);

void ___glfwGetRequiredInstanceExtensionsCocoa(char** extensions);
GLFWbool ___glfwGetPhysicalDevicePresentationSupportCocoa(VkInstance instance, VkPhysicalDevice device, uint32_t queuefamily);
VkResult ____glfwCreateWindowSurfaceCocoa(VkInstance instance, _GLFWwindow* window, const VkAllocationCallbacks* allocator, VkSurfaceKHR* surface);

void ___glfwFreeMonitorCocoa(_GLFWmonitor* monitor);
void ___glfwGetMonitorPosCocoa(_GLFWmonitor* monitor, int* xpos, int* ypos);
void ___glfwGetMonitorContentScaleCocoa(_GLFWmonitor* monitor, float* xscale, float* yscale);
void ___glfwGetMonitorWorkareaCocoa(_GLFWmonitor* monitor, int* xpos, int* ypos, int* width, int* height);
GLFWvidmode* ____glfwGetVideoModesCocoa(_GLFWmonitor* monitor, int* count);
void ___glfwGetVideoModeCocoa(_GLFWmonitor* monitor, GLFWvidmode* mode);
GLFWbool ___glfwGetGammaRampCocoa(_GLFWmonitor* monitor, GLFWgammaramp* ramp);
void ____glfwSetGammaRampCocoa(_GLFWmonitor* monitor, const GLFWgammaramp* ramp);

void __glfwPollMonitorsCocoa(void);
void __glfwSetVideoModeCocoa(_GLFWmonitor* monitor, const GLFWvidmode* desired);
void __glfwRestoreVideoModeCocoa(_GLFWmonitor* monitor);

float __glfwTransformYCocoa(float y);

void* __glfwLoadLocalVulkanLoaderCocoa(void);

GLFWbool ___glfwInitNSGL(void);
void ___glfwTerminateNSGL(void);
GLFWbool __glfwCreateContextNSGL(_GLFWwindow* window,
                                const _GLFWctxconfig* ctxconfig,
                                const _GLFWfbconfig* fbconfig);
void __glfwDestroyContextNSGL(_GLFWwindow* window);

