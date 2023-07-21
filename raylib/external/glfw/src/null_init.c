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
// It is fine to use C99 in this file because it will not be built with VS
//========================================================================

#include "internal.h"

#include <stdlib.h>


//////////////////////////////////////////////////////////////////////////
//////                       GLFW platform API                      //////
//////////////////////////////////////////////////////////////////////////

GLFWbool __glfwConnectNull(int platformID, _GLFWplatform* platform)
{
    const _GLFWplatform null =
    {
        GLFW_PLATFORM_NULL,
        ___glfwInitNull,
        ___glfwTerminateNull,
        ___glfwGetCursorPosNull,
        ____glfwSetCursorPosNull,
        ___glfwSetCursorModeNull,
        __glfwSetRawMouseMotionNull,
        ___glfwRawMouseMotionSupportedNull,
        ___glfwCreateCursorNull,
        ___glfwCreateStandardCursorNull,
        ___glfwDestroyCursorNull,
        ___glfwSetCursorNull,
        __glfwGetScancodeNameNull,
        ____glfwGetKeyScancodeNull,
        ___glfwSetClipboardStringNull,
        ___glfwGetClipboardStringNull,
        ___glfwInitJoysticksNull,
        ___glfwTerminateJoysticksNull,
        __glfwPollJoystickNull,
        __glfwGetMappingNameNull,
        __glfwUpdateGamepadGUIDNull,
        ___glfwFreeMonitorNull,
        ___glfwGetMonitorPosNull,
        ___glfwGetMonitorContentScaleNull,
        ___glfwGetMonitorWorkareaNull,
        ____glfwGetVideoModesNull,
        ___glfwGetVideoModeNull,
        ___glfwGetGammaRampNull,
        ____glfwSetGammaRampNull,
        ___glfwCreateWindowNull,
        ___glfwDestroyWindowNull,
        ___glfwSetWindowTitleNull,
        ___glfwSetWindowIconNull,
        ___glfwGetWindowPosNull,
        ___glfwSetWindowPosNull,
        ___glfwGetWindowSizeNull,
        ___glfwSetWindowSizeNull,
        ____glfwSetWindowSizeLimitsNull,
        ___glfwSetWindowAspectRatioNull,
        ___glfwGetFramebufferSizeNull,
        ___glfwGetWindowFrameSizeNull,
        ___glfwGetWindowContentScaleNull,
        ___glfwIconifyWindowNull,
        ___glfwRestoreWindowNull,
        ___glfwMaximizeWindowNull,
        ___glfwShowWindowNull,
        ___glfwHideWindowNull,
        ___glfwRequestWindowAttentionNull,
        ___glfwFocusWindowNull,
        ___glfwSetWindowMonitorNull,
        __glfwWindowFocusedNull,
        __glfwWindowIconifiedNull,
        __glfwWindowVisibleNull,
        __glfwWindowMaximizedNull,
        __glfwWindowHoveredNull,
        __glfwFramebufferTransparentNull,
        ___glfwGetWindowOpacityNull,
        __glfwSetWindowResizableNull,
        __glfwSetWindowDecoratedNull,
        __glfwSetWindowFloatingNull,
        ___glfwSetWindowOpacityNull,
        __glfwSetWindowMousePassthroughNull,
        ___glfwPollEventsNull,
        ___glfwWaitEventsNull,
        ____glfwWaitEventsTimeoutNull,
        ___glfwPostEmptyEventNull,
        __glfwGetEGLPlatformNull,
        __glfwGetEGLNativeDisplayNull,
        __glfwGetEGLNativeWindowNull,
        ___glfwGetRequiredInstanceExtensionsNull,
        ___glfwGetPhysicalDevicePresentationSupportNull,
        ____glfwCreateWindowSurfaceNull,
    };

    *platform = null;
    return GLFW_TRUE;
}

int ___glfwInitNull(void)
{
    __glfwPollMonitorsNull();
    return GLFW_TRUE;
}

void ___glfwTerminateNull(void)
{
    free(__glfw.null.clipboardString);
    ____glfwTerminateOSMesa();
    ____glfwTerminateEGL();
}

