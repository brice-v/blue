//========================================================================
// GLFW 3.4 - www.glfw.org
//------------------------------------------------------------------------
// Copyright (c) 2002-2006 Marcus Geelnard
// Copyright (c) 2006-2018 Camilla Löwy <elmindreda@glfw.org>
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

#include <string.h>
#include <stdlib.h>
#include <stdio.h>
#include <stdarg.h>
#include <assert.h>


// NOTE: The global variables below comprise all mutable global data in GLFW
//       Any other mutable global variable is a bug

// This contains all mutable state shared between compilation units of GLFW
//
_GLFWlibrary __glfw = { GLFW_FALSE };

// These are outside of __glfw so they can be used before initialization and
// after termination without special handling when __glfw is cleared to zero
//
static _GLFWerror __glfwMainThreadError;
static GLFWerrorfun __glfwErrorCallback;
static GLFWallocator ___glfwInitAllocator;
static _GLFWinitconfig ____glfwInitHints =
{
    GLFW_TRUE,      // hat buttons
    GLFW_ANGLE_PLATFORM_TYPE_NONE, // ANGLE backend
    GLFW_ANY_PLATFORM, // preferred platform
    NULL,           // vkGetInstanceProcAddr function
    {
        GLFW_TRUE,  // macOS menu bar
        GLFW_TRUE   // macOS bundle chdir
    },
    {
        GLFW_TRUE,  // X11 XCB Vulkan surface
    },
};

// The allocation function used when no custom allocator is set
//
static void* defaultAllocate(size_t size, void* user)
{
    return malloc(size);
}

// The deallocation function used when no custom allocator is set
//
static void defaultDeallocate(void* block, void* user)
{
    free(block);
}

// The reallocation function used when no custom allocator is set
//
static void* defaultReallocate(void* block, size_t size, void* user)
{
    return realloc(block, size);
}

// Terminate the library
//
static void terminate(void)
{
    int i;

    memset(&__glfw.callbacks, 0, sizeof(__glfw.callbacks));

    while (__glfw.windowListHead)
        __glfwDestroyWindow((GLFWwindow*) __glfw.windowListHead);

    while (__glfw.cursorListHead)
        __glfwDestroyCursor((GLFWcursor*) __glfw.cursorListHead);

    for (i = 0;  i < __glfw.monitorCount;  i++)
    {
        _GLFWmonitor* monitor = __glfw.monitors[i];
        if (monitor->originalRamp.size)
            __glfw.platform.setGammaRamp(monitor, &monitor->originalRamp);
        ___glfwFreeMonitor(monitor);
    }

    __glfw_free(__glfw.monitors);
    __glfw.monitors = NULL;
    __glfw.monitorCount = 0;

    __glfw_free(__glfw.mappings);
    __glfw.mappings = NULL;
    __glfw.mappingCount = 0;

    ____glfwTerminateVulkan();
    __glfw.platform.terminateJoysticks();
    __glfw.platform.terminate();

    __glfw.initialized = GLFW_FALSE;

    while (__glfw.errorListHead)
    {
        _GLFWerror* error = __glfw.errorListHead;
        __glfw.errorListHead = error->next;
        __glfw_free(error);
    }

    ___glfwPlatformDestroyTls(&__glfw.contextSlot);
    ___glfwPlatformDestroyTls(&__glfw.errorSlot);
    ___glfwPlatformDestroyMutex(&__glfw.errorLock);

    memset(&__glfw, 0, sizeof(__glfw));
}


//////////////////////////////////////////////////////////////////////////
//////                       GLFW internal API                      //////
//////////////////////////////////////////////////////////////////////////

// Encode a Unicode code point to a UTF-8 stream
// Based on cutef8 by Jeff Bezanson (Public Domain)
//
size_t ___glfwEncodeUTF8(char* s, uint32_t codepoint)
{
    size_t count = 0;

    if (codepoint < 0x80)
        s[count++] = (char) codepoint;
    else if (codepoint < 0x800)
    {
        s[count++] = (codepoint >> 6) | 0xc0;
        s[count++] = (codepoint & 0x3f) | 0x80;
    }
    else if (codepoint < 0x10000)
    {
        s[count++] = (codepoint >> 12) | 0xe0;
        s[count++] = ((codepoint >> 6) & 0x3f) | 0x80;
        s[count++] = (codepoint & 0x3f) | 0x80;
    }
    else if (codepoint < 0x110000)
    {
        s[count++] = (codepoint >> 18) | 0xf0;
        s[count++] = ((codepoint >> 12) & 0x3f) | 0x80;
        s[count++] = ((codepoint >> 6) & 0x3f) | 0x80;
        s[count++] = (codepoint & 0x3f) | 0x80;
    }

    return count;
}

// Splits and translates a text/uri-list into separate file paths
// NOTE: This function destroys the provided string
//
char** ___glfwParseUriList(char* text, int* count)
{
    const char* prefix = "file://";
    char** paths = NULL;
    char* line;

    *count = 0;

    while ((line = strtok(text, "\r\n")))
    {
        char* path;

        text = NULL;

        if (line[0] == '#')
            continue;

        if (strncmp(line, prefix, strlen(prefix)) == 0)
        {
            line += strlen(prefix);
            // TODO: Validate hostname
            while (*line != '/')
                line++;
        }

        (*count)++;

        path = __glfw_calloc(strlen(line) + 1, 1);
        paths = __glfw_realloc(paths, *count * sizeof(char*));
        paths[*count - 1] = path;

        while (*line)
        {
            if (line[0] == '%' && line[1] && line[2])
            {
                const char digits[3] = { line[1], line[2], '\0' };
                *path = (char) strtol(digits, NULL, 16);
                line += 2;
            }
            else
                *path = *line;

            path++;
            line++;
        }
    }

    return paths;
}

char* ___glfw_strdup(const char* source)
{
    const size_t length = strlen(source);
    char* result = __glfw_calloc(length + 1, 1);
    strcpy(result, source);
    return result;
}

int ___glfw_min(int a, int b)
{
    return a < b ? a : b;
}

int ___glfw_max(int a, int b)
{
    return a > b ? a : b;
}

float ___glfw_fminf(float a, float b)
{
    if (a != a)
        return b;
    else if (b != b)
        return a;
    else if (a < b)
        return a;
    else
        return b;
}

float ___glfw_fmaxf(float a, float b)
{
    if (a != a)
        return b;
    else if (b != b)
        return a;
    else if (a > b)
        return a;
    else
        return b;
}

void* __glfw_calloc(size_t count, size_t size)
{
    if (count && size)
    {
        void* block;

        if (count > SIZE_MAX / size)
        {
            ___glfwInputError(GLFW_INVALID_VALUE, "Allocation size overflow");
            return NULL;
        }

        block = __glfw.allocator.allocate(count * size, __glfw.allocator.user);
        if (block)
            return memset(block, 0, count * size);
        else
        {
            ___glfwInputError(GLFW_OUT_OF_MEMORY, NULL);
            return NULL;
        }
    }
    else
        return NULL;
}

void* __glfw_realloc(void* block, size_t size)
{
    if (block && size)
    {
        void* resized = __glfw.allocator.reallocate(block, size, __glfw.allocator.user);
        if (resized)
            return resized;
        else
        {
            ___glfwInputError(GLFW_OUT_OF_MEMORY, NULL);
            return NULL;
        }
    }
    else if (block)
    {
        __glfw_free(block);
        return NULL;
    }
    else
        return __glfw_calloc(1, size);
}

void __glfw_free(void* block)
{
    if (block)
        __glfw.allocator.deallocate(block, __glfw.allocator.user);
}


//////////////////////////////////////////////////////////////////////////
//////                         GLFW event API                       //////
//////////////////////////////////////////////////////////////////////////

// Notifies shared code of an error
//
void ___glfwInputError(int code, const char* format, ...)
{
    _GLFWerror* error;
    char description[_GLFW_MESSAGE_SIZE];

    if (format)
    {
        va_list vl;

        va_start(vl, format);
        vsnprintf(description, sizeof(description), format, vl);
        va_end(vl);

        description[sizeof(description) - 1] = '\0';
    }
    else
    {
        if (code == GLFW_NOT_INITIALIZED)
            strcpy(description, "The GLFW library is not initialized");
        else if (code == GLFW_NO_CURRENT_CONTEXT)
            strcpy(description, "There is no current context");
        else if (code == GLFW_INVALID_ENUM)
            strcpy(description, "Invalid argument for enum parameter");
        else if (code == GLFW_INVALID_VALUE)
            strcpy(description, "Invalid value for parameter");
        else if (code == GLFW_OUT_OF_MEMORY)
            strcpy(description, "Out of memory");
        else if (code == GLFW_API_UNAVAILABLE)
            strcpy(description, "The requested API is unavailable");
        else if (code == GLFW_VERSION_UNAVAILABLE)
            strcpy(description, "The requested API version is unavailable");
        else if (code == GLFW_PLATFORM_ERROR)
            strcpy(description, "A platform-specific error occurred");
        else if (code == GLFW_FORMAT_UNAVAILABLE)
            strcpy(description, "The requested format is unavailable");
        else if (code == GLFW_NO_WINDOW_CONTEXT)
            strcpy(description, "The specified window has no context");
        else if (code == GLFW_CURSOR_UNAVAILABLE)
            strcpy(description, "The specified cursor shape is unavailable");
        else if (code == GLFW_FEATURE_UNAVAILABLE)
            strcpy(description, "The requested feature cannot be implemented for this platform");
        else if (code == GLFW_FEATURE_UNIMPLEMENTED)
            strcpy(description, "The requested feature has not yet been implemented for this platform");
        else if (code == GLFW_PLATFORM_UNAVAILABLE)
            strcpy(description, "The requested platform is unavailable");
        else
            strcpy(description, "ERROR: UNKNOWN GLFW ERROR");
    }

    if (__glfw.initialized)
    {
        error = ___glfwPlatformGetTls(&__glfw.errorSlot);
        if (!error)
        {
            error = __glfw_calloc(1, sizeof(_GLFWerror));
            ___glfwPlatformSetTls(&__glfw.errorSlot, error);
            ___glfwPlatformLockMutex(&__glfw.errorLock);
            error->next = __glfw.errorListHead;
            __glfw.errorListHead = error;
            ___glfwPlatformUnlockMutex(&__glfw.errorLock);
        }
    }
    else
        error = &__glfwMainThreadError;

    error->code = code;
    strcpy(error->description, description);

    if (__glfwErrorCallback)
        __glfwErrorCallback(code, description);
}


//////////////////////////////////////////////////////////////////////////
//////                        GLFW public API                       //////
//////////////////////////////////////////////////////////////////////////

GLFWAPI int __glfwInit(void)
{
    if (__glfw.initialized)
        return GLFW_TRUE;

    memset(&__glfw, 0, sizeof(__glfw));
    __glfw.hints.init = ____glfwInitHints;

    __glfw.allocator = ___glfwInitAllocator;
    if (!__glfw.allocator.allocate)
    {
        __glfw.allocator.allocate   = defaultAllocate;
        __glfw.allocator.reallocate = defaultReallocate;
        __glfw.allocator.deallocate = defaultDeallocate;
    }

    if (!__glfwSelectPlatform(__glfw.hints.init.platformID, &__glfw.platform))
        return GLFW_FALSE;

    if (!__glfw.platform.init())
    {
        terminate();
        return GLFW_FALSE;
    }

    if (!___glfwPlatformCreateMutex(&__glfw.errorLock) ||
        !___glfwPlatformCreateTls(&__glfw.errorSlot) ||
        !___glfwPlatformCreateTls(&__glfw.contextSlot))
    {
        terminate();
        return GLFW_FALSE;
    }

    ___glfwPlatformSetTls(&__glfw.errorSlot, &__glfwMainThreadError);

    ____glfwInitGamepadMappings();

    __glfwPlatformInitTimer();
    __glfw.timer.offset = ___glfwPlatformGetTimerValue();

    __glfw.initialized = GLFW_TRUE;

    __glfwDefaultWindowHints();
    return GLFW_TRUE;
}

GLFWAPI void __glfwTerminate(void)
{
    if (!__glfw.initialized)
        return;

    terminate();
}

GLFWAPI void ___glfwInitHint(int hint, int value)
{
    switch (hint)
    {
        case GLFW_JOYSTICK_HAT_BUTTONS:
            ____glfwInitHints.hatButtons = value;
            return;
        case GLFW_ANGLE_PLATFORM_TYPE:
            ____glfwInitHints.angleType = value;
            return;
        case GLFW_PLATFORM:
            ____glfwInitHints.platformID = value;
            return;
        case GLFW_COCOA_CHDIR_RESOURCES:
            ____glfwInitHints.ns.chdir = value;
            return;
        case GLFW_COCOA_MENUBAR:
            ____glfwInitHints.ns.menubar = value;
            return;
        case GLFW_X11_XCB_VULKAN_SURFACE:
            ____glfwInitHints.x11.xcbVulkanSurface = value;
            return;
    }

    ___glfwInputError(GLFW_INVALID_ENUM,
                    "Invalid init hint 0x%08X", hint);
}

GLFWAPI void __glfwInitAllocator(const GLFWallocator* allocator)
{
    if (allocator)
    {
        if (allocator->allocate && allocator->reallocate && allocator->deallocate)
            ___glfwInitAllocator = *allocator;
        else
            ___glfwInputError(GLFW_INVALID_VALUE, "Missing function in allocator");
    }
    else
        memset(&___glfwInitAllocator, 0, sizeof(GLFWallocator));
}

GLFWAPI void __glfwInitVulkanLoader(PFN_vkGetInstanceProcAddr loader)
{
    ____glfwInitHints.vulkanLoader = loader;
}

GLFWAPI void __glfwGetVersion(int* major, int* minor, int* rev)
{
    if (major != NULL)
        *major = GLFW_VERSION_MAJOR;
    if (minor != NULL)
        *minor = GLFW_VERSION_MINOR;
    if (rev != NULL)
        *rev = GLFW_VERSION_REVISION;
}

GLFWAPI int __glfwGetError(const char** description)
{
    _GLFWerror* error;
    int code = GLFW_NO_ERROR;

    if (description)
        *description = NULL;

    if (__glfw.initialized)
        error = ___glfwPlatformGetTls(&__glfw.errorSlot);
    else
        error = &__glfwMainThreadError;

    if (error)
    {
        code = error->code;
        error->code = GLFW_NO_ERROR;
        if (description && code)
            *description = error->description;
    }

    return code;
}

GLFWAPI GLFWerrorfun __glfwSetErrorCallback(GLFWerrorfun cbfun)
{
    _GLFW_SWAP(GLFWerrorfun, __glfwErrorCallback, cbfun);
    return cbfun;
}

