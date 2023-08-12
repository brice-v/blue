//========================================================================
// GLFW 3.4 - www.glfw.org
//------------------------------------------------------------------------
// Copyright (c) 2016 Google Inc.
// Copyright (c) 2016-2017 Camilla Löwy <elmindreda@glfw.org>
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

#define GLFW_NULL_WINDOW_STATE          _GLFWwindowNull null;
#define GLFW_NULL_LIBRARY_WINDOW_STATE  _GLFWlibraryNull null;
#define GLFW_NULL_MONITOR_STATE         _GLFWmonitorNull null;

#define GLFW_NULL_CONTEXT_STATE
#define GLFW_NULL_CURSOR_STATE
#define GLFW_NULL_LIBRARY_CONTEXT_STATE


// Null-specific per-window data
//
typedef struct _GLFWwindowNull
{
    int             xpos;
    int             ypos;
    int             width;
    int             height;
    char*           title;
    GLFWbool        visible;
    GLFWbool        iconified;
    GLFWbool        maximized;
    GLFWbool        resizable;
    GLFWbool        decorated;
    GLFWbool        floating;
    GLFWbool        transparent;
    float           opacity;
} _GLFWwindowNull;

// Null-specific per-monitor data
//
typedef struct _GLFWmonitorNull
{
    GLFWgammaramp   ramp;
} _GLFWmonitorNull;

// Null-specific global data
//
typedef struct _GLFWlibraryNull
{
    int             xcursor;
    int             ycursor;
    char*           clipboardString;
    _GLFWwindow*    focusedWindow;
} _GLFWlibraryNull;

void __glfwPollMonitorsNull(void);

GLFWbool __glfwConnectNull(int platformID, _GLFWplatform* platform);
int ___glfwInitNull(void);
void ___glfwTerminateNull(void);

void ___glfwFreeMonitorNull(_GLFWmonitor* monitor);
void ___glfwGetMonitorPosNull(_GLFWmonitor* monitor, int* xpos, int* ypos);
void ___glfwGetMonitorContentScaleNull(_GLFWmonitor* monitor, float* xscale, float* yscale);
void ___glfwGetMonitorWorkareaNull(_GLFWmonitor* monitor, int* xpos, int* ypos, int* width, int* height);
GLFWvidmode* ____glfwGetVideoModesNull(_GLFWmonitor* monitor, int* found);
void ___glfwGetVideoModeNull(_GLFWmonitor* monitor, GLFWvidmode* mode);
GLFWbool ___glfwGetGammaRampNull(_GLFWmonitor* monitor, GLFWgammaramp* ramp);
void ____glfwSetGammaRampNull(_GLFWmonitor* monitor, const GLFWgammaramp* ramp);

GLFWbool ___glfwCreateWindowNull(_GLFWwindow* window, const _GLFWwndconfig* wndconfig, const _GLFWctxconfig* ctxconfig, const _GLFWfbconfig* fbconfig);
void ___glfwDestroyWindowNull(_GLFWwindow* window);
void ___glfwSetWindowTitleNull(_GLFWwindow* window, const char* title);
void ___glfwSetWindowIconNull(_GLFWwindow* window, int count, const GLFWimage* images);
void ___glfwSetWindowMonitorNull(_GLFWwindow* window, _GLFWmonitor* monitor, int xpos, int ypos, int width, int height, int refreshRate);
void ___glfwGetWindowPosNull(_GLFWwindow* window, int* xpos, int* ypos);
void ___glfwSetWindowPosNull(_GLFWwindow* window, int xpos, int ypos);
void ___glfwGetWindowSizeNull(_GLFWwindow* window, int* width, int* height);
void ___glfwSetWindowSizeNull(_GLFWwindow* window, int width, int height);
void ____glfwSetWindowSizeLimitsNull(_GLFWwindow* window, int minwidth, int minheight, int maxwidth, int maxheight);
void ___glfwSetWindowAspectRatioNull(_GLFWwindow* window, int n, int d);
void ___glfwGetFramebufferSizeNull(_GLFWwindow* window, int* width, int* height);
void ___glfwGetWindowFrameSizeNull(_GLFWwindow* window, int* left, int* top, int* right, int* bottom);
void ___glfwGetWindowContentScaleNull(_GLFWwindow* window, float* xscale, float* yscale);
void ___glfwIconifyWindowNull(_GLFWwindow* window);
void ___glfwRestoreWindowNull(_GLFWwindow* window);
void ___glfwMaximizeWindowNull(_GLFWwindow* window);
GLFWbool __glfwWindowMaximizedNull(_GLFWwindow* window);
GLFWbool __glfwWindowHoveredNull(_GLFWwindow* window);
GLFWbool __glfwFramebufferTransparentNull(_GLFWwindow* window);
void __glfwSetWindowResizableNull(_GLFWwindow* window, GLFWbool enabled);
void __glfwSetWindowDecoratedNull(_GLFWwindow* window, GLFWbool enabled);
void __glfwSetWindowFloatingNull(_GLFWwindow* window, GLFWbool enabled);
void __glfwSetWindowMousePassthroughNull(_GLFWwindow* window, GLFWbool enabled);
float ___glfwGetWindowOpacityNull(_GLFWwindow* window);
void ___glfwSetWindowOpacityNull(_GLFWwindow* window, float opacity);
void __glfwSetRawMouseMotionNull(_GLFWwindow *window, GLFWbool enabled);
GLFWbool ___glfwRawMouseMotionSupportedNull(void);
void ___glfwShowWindowNull(_GLFWwindow* window);
void ___glfwRequestWindowAttentionNull(_GLFWwindow* window);
void ___glfwRequestWindowAttentionNull(_GLFWwindow* window);
void ___glfwHideWindowNull(_GLFWwindow* window);
void ___glfwFocusWindowNull(_GLFWwindow* window);
GLFWbool __glfwWindowFocusedNull(_GLFWwindow* window);
GLFWbool __glfwWindowIconifiedNull(_GLFWwindow* window);
GLFWbool __glfwWindowVisibleNull(_GLFWwindow* window);
void ___glfwPollEventsNull(void);
void ___glfwWaitEventsNull(void);
void ____glfwWaitEventsTimeoutNull(double timeout);
void ___glfwPostEmptyEventNull(void);
void ___glfwGetCursorPosNull(_GLFWwindow* window, double* xpos, double* ypos);
void ____glfwSetCursorPosNull(_GLFWwindow* window, double x, double y);
void ___glfwSetCursorModeNull(_GLFWwindow* window, int mode);
GLFWbool ___glfwCreateCursorNull(_GLFWcursor* cursor, const GLFWimage* image, int xhot, int yhot);
GLFWbool ___glfwCreateStandardCursorNull(_GLFWcursor* cursor, int shape);
void ___glfwDestroyCursorNull(_GLFWcursor* cursor);
void ___glfwSetCursorNull(_GLFWwindow* window, _GLFWcursor* cursor);
void ___glfwSetClipboardStringNull(const char* string);
const char* ___glfwGetClipboardStringNull(void);
const char* __glfwGetScancodeNameNull(int scancode);
int ____glfwGetKeyScancodeNull(int key);

EGLenum __glfwGetEGLPlatformNull(EGLint** attribs);
EGLNativeDisplayType __glfwGetEGLNativeDisplayNull(void);
EGLNativeWindowType __glfwGetEGLNativeWindowNull(_GLFWwindow* window);

void ___glfwGetRequiredInstanceExtensionsNull(char** extensions);
GLFWbool ___glfwGetPhysicalDevicePresentationSupportNull(VkInstance instance, VkPhysicalDevice device, uint32_t queuefamily);
VkResult ____glfwCreateWindowSurfaceNull(VkInstance instance, _GLFWwindow* window, const VkAllocationCallbacks* allocator, VkSurfaceKHR* surface);

void __glfwPollMonitorsNull(void);

