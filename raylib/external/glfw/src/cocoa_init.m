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
// It is fine to use C99 in this file because it will not be built with VS
//========================================================================

#include "internal.h"
#include <sys/param.h> // For MAXPATHLEN

// Needed for _NSGetProgname
#include <crt_externs.h>

// Change to our application bundle's resources directory, if present
//
static void changeToResourcesDirectory(void)
{
    char resourcesPath[MAXPATHLEN];

    CFBundleRef bundle = CFBundleGetMainBundle();
    if (!bundle)
        return;

    CFURLRef resourcesURL = CFBundleCopyResourcesDirectoryURL(bundle);

    CFStringRef last = CFURLCopyLastPathComponent(resourcesURL);
    if (CFStringCompare(CFSTR("Resources"), last, 0) != kCFCompareEqualTo)
    {
        CFRelease(last);
        CFRelease(resourcesURL);
        return;
    }

    CFRelease(last);

    if (!CFURLGetFileSystemRepresentation(resourcesURL,
                                          true,
                                          (UInt8*) resourcesPath,
                                          MAXPATHLEN))
    {
        CFRelease(resourcesURL);
        return;
    }

    CFRelease(resourcesURL);

    chdir(resourcesPath);
}

// Set up the menu bar (manually)
// This is nasty, nasty stuff -- calls to undocumented semi-private APIs that
// could go away at any moment, lots of stuff that really should be
// localize(d|able), etc.  Add a nib to save us this horror.
//
static void createMenuBar(void)
{
    NSString* appName = nil;
    NSDictionary* bundleInfo = [[NSBundle mainBundle] infoDictionary];
    NSString* nameKeys[] =
    {
        @"CFBundleDisplayName",
        @"CFBundleName",
        @"CFBundleExecutable",
    };

    // Try to figure out what the calling application is called

    for (size_t i = 0;  i < sizeof(nameKeys) / sizeof(nameKeys[0]);  i++)
    {
        id name = bundleInfo[nameKeys[i]];
        if (name &&
            [name isKindOfClass:[NSString class]] &&
            ![name isEqualToString:@""])
        {
            appName = name;
            break;
        }
    }

    if (!appName)
    {
        char** progname = _NSGetProgname();
        if (progname && *progname)
            appName = @(*progname);
        else
            appName = @"GLFW Application";
    }

    NSMenu* bar = [[NSMenu alloc] init];
    [NSApp setMainMenu:bar];

    NSMenuItem* appMenuItem =
        [bar addItemWithTitle:@"" action:NULL keyEquivalent:@""];
    NSMenu* appMenu = [[NSMenu alloc] init];
    [appMenuItem setSubmenu:appMenu];

    [appMenu addItemWithTitle:[NSString stringWithFormat:@"About %@", appName]
                       action:@selector(orderFrontStandardAboutPanel:)
                keyEquivalent:@""];
    [appMenu addItem:[NSMenuItem separatorItem]];
    NSMenu* servicesMenu = [[NSMenu alloc] init];
    [NSApp setServicesMenu:servicesMenu];
    [[appMenu addItemWithTitle:@"Services"
                       action:NULL
                keyEquivalent:@""] setSubmenu:servicesMenu];
    [servicesMenu release];
    [appMenu addItem:[NSMenuItem separatorItem]];
    [appMenu addItemWithTitle:[NSString stringWithFormat:@"Hide %@", appName]
                       action:@selector(hide:)
                keyEquivalent:@"h"];
    [[appMenu addItemWithTitle:@"Hide Others"
                       action:@selector(hideOtherApplications:)
                keyEquivalent:@"h"]
        setKeyEquivalentModifierMask:NSEventModifierFlagOption | NSEventModifierFlagCommand];
    [appMenu addItemWithTitle:@"Show All"
                       action:@selector(unhideAllApplications:)
                keyEquivalent:@""];
    [appMenu addItem:[NSMenuItem separatorItem]];
    [appMenu addItemWithTitle:[NSString stringWithFormat:@"Quit %@", appName]
                       action:@selector(terminate:)
                keyEquivalent:@"q"];

    NSMenuItem* windowMenuItem =
        [bar addItemWithTitle:@"" action:NULL keyEquivalent:@""];
    [bar release];
    NSMenu* windowMenu = [[NSMenu alloc] initWithTitle:@"Window"];
    [NSApp setWindowsMenu:windowMenu];
    [windowMenuItem setSubmenu:windowMenu];

    [windowMenu addItemWithTitle:@"Minimize"
                          action:@selector(performMiniaturize:)
                   keyEquivalent:@"m"];
    [windowMenu addItemWithTitle:@"Zoom"
                          action:@selector(performZoom:)
                   keyEquivalent:@""];
    [windowMenu addItem:[NSMenuItem separatorItem]];
    [windowMenu addItemWithTitle:@"Bring All to Front"
                          action:@selector(arrangeInFront:)
                   keyEquivalent:@""];

    // TODO: Make this appear at the bottom of the menu (for consistency)
    [windowMenu addItem:[NSMenuItem separatorItem]];
    [[windowMenu addItemWithTitle:@"Enter Full Screen"
                           action:@selector(toggleFullScreen:)
                    keyEquivalent:@"f"]
     setKeyEquivalentModifierMask:NSEventModifierFlagControl | NSEventModifierFlagCommand];

    // Prior to Snow Leopard, we need to use this oddly-named semi-private API
    // to get the application menu working properly.
    SEL setAppleMenuSelector = NSSelectorFromString(@"setAppleMenu:");
    [NSApp performSelector:setAppleMenuSelector withObject:appMenu];
}

// Create key code translation tables
//
static void createKeyTables(void)
{
    memset(__glfw.ns.keycodes, -1, sizeof(__glfw.ns.keycodes));
    memset(__glfw.ns.scancodes, -1, sizeof(__glfw.ns.scancodes));

    __glfw.ns.keycodes[0x1D] = GLFW_KEY_0;
    __glfw.ns.keycodes[0x12] = GLFW_KEY_1;
    __glfw.ns.keycodes[0x13] = GLFW_KEY_2;
    __glfw.ns.keycodes[0x14] = GLFW_KEY_3;
    __glfw.ns.keycodes[0x15] = GLFW_KEY_4;
    __glfw.ns.keycodes[0x17] = GLFW_KEY_5;
    __glfw.ns.keycodes[0x16] = GLFW_KEY_6;
    __glfw.ns.keycodes[0x1A] = GLFW_KEY_7;
    __glfw.ns.keycodes[0x1C] = GLFW_KEY_8;
    __glfw.ns.keycodes[0x19] = GLFW_KEY_9;
    __glfw.ns.keycodes[0x00] = GLFW_KEY_A;
    __glfw.ns.keycodes[0x0B] = GLFW_KEY_B;
    __glfw.ns.keycodes[0x08] = GLFW_KEY_C;
    __glfw.ns.keycodes[0x02] = GLFW_KEY_D;
    __glfw.ns.keycodes[0x0E] = GLFW_KEY_E;
    __glfw.ns.keycodes[0x03] = GLFW_KEY_F;
    __glfw.ns.keycodes[0x05] = GLFW_KEY_G;
    __glfw.ns.keycodes[0x04] = GLFW_KEY_H;
    __glfw.ns.keycodes[0x22] = GLFW_KEY_I;
    __glfw.ns.keycodes[0x26] = GLFW_KEY_J;
    __glfw.ns.keycodes[0x28] = GLFW_KEY_K;
    __glfw.ns.keycodes[0x25] = GLFW_KEY_L;
    __glfw.ns.keycodes[0x2E] = GLFW_KEY_M;
    __glfw.ns.keycodes[0x2D] = GLFW_KEY_N;
    __glfw.ns.keycodes[0x1F] = GLFW_KEY_O;
    __glfw.ns.keycodes[0x23] = GLFW_KEY_P;
    __glfw.ns.keycodes[0x0C] = GLFW_KEY_Q;
    __glfw.ns.keycodes[0x0F] = GLFW_KEY_R;
    __glfw.ns.keycodes[0x01] = GLFW_KEY_S;
    __glfw.ns.keycodes[0x11] = GLFW_KEY_T;
    __glfw.ns.keycodes[0x20] = GLFW_KEY_U;
    __glfw.ns.keycodes[0x09] = GLFW_KEY_V;
    __glfw.ns.keycodes[0x0D] = GLFW_KEY_W;
    __glfw.ns.keycodes[0x07] = GLFW_KEY_X;
    __glfw.ns.keycodes[0x10] = GLFW_KEY_Y;
    __glfw.ns.keycodes[0x06] = GLFW_KEY_Z;

    __glfw.ns.keycodes[0x27] = GLFW_KEY_APOSTROPHE;
    __glfw.ns.keycodes[0x2A] = GLFW_KEY_BACKSLASH;
    __glfw.ns.keycodes[0x2B] = GLFW_KEY_COMMA;
    __glfw.ns.keycodes[0x18] = GLFW_KEY_EQUAL;
    __glfw.ns.keycodes[0x32] = GLFW_KEY_GRAVE_ACCENT;
    __glfw.ns.keycodes[0x21] = GLFW_KEY_LEFT_BRACKET;
    __glfw.ns.keycodes[0x1B] = GLFW_KEY_MINUS;
    __glfw.ns.keycodes[0x2F] = GLFW_KEY_PERIOD;
    __glfw.ns.keycodes[0x1E] = GLFW_KEY_RIGHT_BRACKET;
    __glfw.ns.keycodes[0x29] = GLFW_KEY_SEMICOLON;
    __glfw.ns.keycodes[0x2C] = GLFW_KEY_SLASH;
    __glfw.ns.keycodes[0x0A] = GLFW_KEY_WORLD_1;

    __glfw.ns.keycodes[0x33] = GLFW_KEY_BACKSPACE;
    __glfw.ns.keycodes[0x39] = GLFW_KEY_CAPS_LOCK;
    __glfw.ns.keycodes[0x75] = GLFW_KEY_DELETE;
    __glfw.ns.keycodes[0x7D] = GLFW_KEY_DOWN;
    __glfw.ns.keycodes[0x77] = GLFW_KEY_END;
    __glfw.ns.keycodes[0x24] = GLFW_KEY_ENTER;
    __glfw.ns.keycodes[0x35] = GLFW_KEY_ESCAPE;
    __glfw.ns.keycodes[0x7A] = GLFW_KEY_F1;
    __glfw.ns.keycodes[0x78] = GLFW_KEY_F2;
    __glfw.ns.keycodes[0x63] = GLFW_KEY_F3;
    __glfw.ns.keycodes[0x76] = GLFW_KEY_F4;
    __glfw.ns.keycodes[0x60] = GLFW_KEY_F5;
    __glfw.ns.keycodes[0x61] = GLFW_KEY_F6;
    __glfw.ns.keycodes[0x62] = GLFW_KEY_F7;
    __glfw.ns.keycodes[0x64] = GLFW_KEY_F8;
    __glfw.ns.keycodes[0x65] = GLFW_KEY_F9;
    __glfw.ns.keycodes[0x6D] = GLFW_KEY_F10;
    __glfw.ns.keycodes[0x67] = GLFW_KEY_F11;
    __glfw.ns.keycodes[0x6F] = GLFW_KEY_F12;
    __glfw.ns.keycodes[0x69] = GLFW_KEY_PRINT_SCREEN;
    __glfw.ns.keycodes[0x6B] = GLFW_KEY_F14;
    __glfw.ns.keycodes[0x71] = GLFW_KEY_F15;
    __glfw.ns.keycodes[0x6A] = GLFW_KEY_F16;
    __glfw.ns.keycodes[0x40] = GLFW_KEY_F17;
    __glfw.ns.keycodes[0x4F] = GLFW_KEY_F18;
    __glfw.ns.keycodes[0x50] = GLFW_KEY_F19;
    __glfw.ns.keycodes[0x5A] = GLFW_KEY_F20;
    __glfw.ns.keycodes[0x73] = GLFW_KEY_HOME;
    __glfw.ns.keycodes[0x72] = GLFW_KEY_INSERT;
    __glfw.ns.keycodes[0x7B] = GLFW_KEY_LEFT;
    __glfw.ns.keycodes[0x3A] = GLFW_KEY_LEFT_ALT;
    __glfw.ns.keycodes[0x3B] = GLFW_KEY_LEFT_CONTROL;
    __glfw.ns.keycodes[0x38] = GLFW_KEY_LEFT_SHIFT;
    __glfw.ns.keycodes[0x37] = GLFW_KEY_LEFT_SUPER;
    __glfw.ns.keycodes[0x6E] = GLFW_KEY_MENU;
    __glfw.ns.keycodes[0x47] = GLFW_KEY_NUM_LOCK;
    __glfw.ns.keycodes[0x79] = GLFW_KEY_PAGE_DOWN;
    __glfw.ns.keycodes[0x74] = GLFW_KEY_PAGE_UP;
    __glfw.ns.keycodes[0x7C] = GLFW_KEY_RIGHT;
    __glfw.ns.keycodes[0x3D] = GLFW_KEY_RIGHT_ALT;
    __glfw.ns.keycodes[0x3E] = GLFW_KEY_RIGHT_CONTROL;
    __glfw.ns.keycodes[0x3C] = GLFW_KEY_RIGHT_SHIFT;
    __glfw.ns.keycodes[0x36] = GLFW_KEY_RIGHT_SUPER;
    __glfw.ns.keycodes[0x31] = GLFW_KEY_SPACE;
    __glfw.ns.keycodes[0x30] = GLFW_KEY_TAB;
    __glfw.ns.keycodes[0x7E] = GLFW_KEY_UP;

    __glfw.ns.keycodes[0x52] = GLFW_KEY_KP_0;
    __glfw.ns.keycodes[0x53] = GLFW_KEY_KP_1;
    __glfw.ns.keycodes[0x54] = GLFW_KEY_KP_2;
    __glfw.ns.keycodes[0x55] = GLFW_KEY_KP_3;
    __glfw.ns.keycodes[0x56] = GLFW_KEY_KP_4;
    __glfw.ns.keycodes[0x57] = GLFW_KEY_KP_5;
    __glfw.ns.keycodes[0x58] = GLFW_KEY_KP_6;
    __glfw.ns.keycodes[0x59] = GLFW_KEY_KP_7;
    __glfw.ns.keycodes[0x5B] = GLFW_KEY_KP_8;
    __glfw.ns.keycodes[0x5C] = GLFW_KEY_KP_9;
    __glfw.ns.keycodes[0x45] = GLFW_KEY_KP_ADD;
    __glfw.ns.keycodes[0x41] = GLFW_KEY_KP_DECIMAL;
    __glfw.ns.keycodes[0x4B] = GLFW_KEY_KP_DIVIDE;
    __glfw.ns.keycodes[0x4C] = GLFW_KEY_KP_ENTER;
    __glfw.ns.keycodes[0x51] = GLFW_KEY_KP_EQUAL;
    __glfw.ns.keycodes[0x43] = GLFW_KEY_KP_MULTIPLY;
    __glfw.ns.keycodes[0x4E] = GLFW_KEY_KP_SUBTRACT;

    for (int scancode = 0;  scancode < 256;  scancode++)
    {
        // Store the reverse translation for faster key name lookup
        if (__glfw.ns.keycodes[scancode] >= 0)
            __glfw.ns.scancodes[__glfw.ns.keycodes[scancode]] = scancode;
    }
}

// Retrieve Unicode data for the current keyboard layout
//
static GLFWbool updateUnicodeData(void)
{
    if (__glfw.ns.inputSource)
    {
        CFRelease(__glfw.ns.inputSource);
        __glfw.ns.inputSource = NULL;
        __glfw.ns.unicodeData = nil;
    }

    __glfw.ns.inputSource = TISCopyCurrentKeyboardLayoutInputSource();
    if (!__glfw.ns.inputSource)
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "Cocoa: Failed to retrieve keyboard layout input source");
        return GLFW_FALSE;
    }

    __glfw.ns.unicodeData =
        TISGetInputSourceProperty(__glfw.ns.inputSource,
                                  kTISPropertyUnicodeKeyLayoutData);
    if (!__glfw.ns.unicodeData)
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "Cocoa: Failed to retrieve keyboard layout Unicode data");
        return GLFW_FALSE;
    }

    return GLFW_TRUE;
}

// Load HIToolbox.framework and the TIS symbols we need from it
//
static GLFWbool initializeTIS(void)
{
    // This works only because Cocoa has already loaded it properly
    __glfw.ns.tis.bundle =
        CFBundleGetBundleWithIdentifier(CFSTR("com.apple.HIToolbox"));
    if (!__glfw.ns.tis.bundle)
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "Cocoa: Failed to load HIToolbox.framework");
        return GLFW_FALSE;
    }

    CFStringRef* kPropertyUnicodeKeyLayoutData =
        CFBundleGetDataPointerForName(__glfw.ns.tis.bundle,
                                      CFSTR("kTISPropertyUnicodeKeyLayoutData"));
    __glfw.ns.tis.CopyCurrentKeyboardLayoutInputSource =
        CFBundleGetFunctionPointerForName(__glfw.ns.tis.bundle,
                                          CFSTR("TISCopyCurrentKeyboardLayoutInputSource"));
    __glfw.ns.tis.GetInputSourceProperty =
        CFBundleGetFunctionPointerForName(__glfw.ns.tis.bundle,
                                          CFSTR("TISGetInputSourceProperty"));
    __glfw.ns.tis.GetKbdType =
        CFBundleGetFunctionPointerForName(__glfw.ns.tis.bundle,
                                          CFSTR("LMGetKbdType"));

    if (!kPropertyUnicodeKeyLayoutData ||
        !TISCopyCurrentKeyboardLayoutInputSource ||
        !TISGetInputSourceProperty ||
        !LMGetKbdType)
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "Cocoa: Failed to load TIS API symbols");
        return GLFW_FALSE;
    }

    __glfw.ns.tis.kPropertyUnicodeKeyLayoutData =
        *kPropertyUnicodeKeyLayoutData;

    return updateUnicodeData();
}

@interface GLFWHelper : NSObject
@end

@implementation GLFWHelper

- (void)selectedKeyboardInputSourceChanged:(NSObject* )object
{
    updateUnicodeData();
}

- (void)doNothing:(id)object
{
}

@end // GLFWHelper

@interface GLFWApplicationDelegate : NSObject <NSApplicationDelegate>
@end

@implementation GLFWApplicationDelegate

- (NSApplicationTerminateReply)applicationShouldTerminate:(NSApplication *)sender
{
    for (_GLFWwindow* window = __glfw.windowListHead;  window;  window = window->next)
        ___glfwInputWindowCloseRequest(window);

    return NSTerminateCancel;
}

- (void)applicationDidChangeScreenParameters:(NSNotification *) notification
{
    for (_GLFWwindow* window = __glfw.windowListHead;  window;  window = window->next)
    {
        if (window->context.client != GLFW_NO_API)
            [window->context.nsgl.object update];
    }

    __glfwPollMonitorsCocoa();
}

- (void)applicationWillFinishLaunching:(NSNotification *)notification
{
    if (__glfw.hints.init.ns.menubar)
    {
        // Menu bar setup must go between sharedApplication and finishLaunching
        // in order to properly emulate the behavior of NSApplicationMain

        if ([[NSBundle mainBundle] pathForResource:@"MainMenu" ofType:@"nib"])
        {
            [[NSBundle mainBundle] loadNibNamed:@"MainMenu"
                                          owner:NSApp
                                topLevelObjects:&__glfw.ns.nibObjects];
        }
        else
            createMenuBar();
    }
}

- (void)applicationDidFinishLaunching:(NSNotification *)notification
{
    ___glfwPostEmptyEventCocoa();
    [NSApp stop:nil];
}

- (void)applicationDidHide:(NSNotification *)notification
{
    for (int i = 0;  i < __glfw.monitorCount;  i++)
        __glfwRestoreVideoModeCocoa(__glfw.monitors[i]);
}

@end // GLFWApplicationDelegate


//////////////////////////////////////////////////////////////////////////
//////                       GLFW internal API                      //////
//////////////////////////////////////////////////////////////////////////

void* __glfwLoadLocalVulkanLoaderCocoa(void)
{
    CFBundleRef bundle = CFBundleGetMainBundle();
    if (!bundle)
        return NULL;

    CFURLRef frameworksUrl = CFBundleCopyPrivateFrameworksURL(bundle);
    if (!frameworksUrl)
        return NULL;

    CFURLRef loaderUrl = CFURLCreateCopyAppendingPathComponent(
        kCFAllocatorDefault, frameworksUrl, CFSTR("libvulkan.1.dylib"), false);
    if (!loaderUrl)
    {
        CFRelease(frameworksUrl);
        return NULL;
    }

    char path[PATH_MAX];
    void* handle = NULL;

    if (CFURLGetFileSystemRepresentation(loaderUrl, true, (UInt8*) path, sizeof(path) - 1))
        handle = __glfwPlatformLoadModule(path);

    CFRelease(loaderUrl);
    CFRelease(frameworksUrl);
    return handle;
}


//////////////////////////////////////////////////////////////////////////
//////                       GLFW platform API                      //////
//////////////////////////////////////////////////////////////////////////

GLFWbool __glfwConnectCocoa(int platformID, _GLFWplatform* platform)
{
    const _GLFWplatform cocoa =
    {
        GLFW_PLATFORM_COCOA,
        ___glfwInitCocoa,
        ___glfwTerminateCocoa,
        ___glfwGetCursorPosCocoa,
        ____glfwSetCursorPosCocoa,
        ___glfwSetCursorModeCocoa,
        __glfwSetRawMouseMotionCocoa,
        ___glfwRawMouseMotionSupportedCocoa,
        ___glfwCreateCursorCocoa,
        ___glfwCreateStandardCursorCocoa,
        ___glfwDestroyCursorCocoa,
        ___glfwSetCursorCocoa,
        __glfwGetScancodeNameCocoa,
        ____glfwGetKeyScancodeCocoa,
        ___glfwSetClipboardStringCocoa,
        ___glfwGetClipboardStringCocoa,
        ___glfwInitJoysticksCocoa,
        ___glfwTerminateJoysticksCocoa,
        __glfwPollJoystickCocoa,
        __glfwGetMappingNameCocoa,
        __glfwUpdateGamepadGUIDCocoa,
        ___glfwFreeMonitorCocoa,
        ___glfwGetMonitorPosCocoa,
        ___glfwGetMonitorContentScaleCocoa,
        ___glfwGetMonitorWorkareaCocoa,
        ____glfwGetVideoModesCocoa,
        ___glfwGetVideoModeCocoa,
        ___glfwGetGammaRampCocoa,
        ____glfwSetGammaRampCocoa,
        ___glfwCreateWindowCocoa,
        ___glfwDestroyWindowCocoa,
        ___glfwSetWindowTitleCocoa,
        ___glfwSetWindowIconCocoa,
        ___glfwGetWindowPosCocoa,
        ___glfwSetWindowPosCocoa,
        ___glfwGetWindowSizeCocoa,
        ___glfwSetWindowSizeCocoa,
        ____glfwSetWindowSizeLimitsCocoa,
        ___glfwSetWindowAspectRatioCocoa,
        ___glfwGetFramebufferSizeCocoa,
        ___glfwGetWindowFrameSizeCocoa,
        ___glfwGetWindowContentScaleCocoa,
        ___glfwIconifyWindowCocoa,
        ___glfwRestoreWindowCocoa,
        ___glfwMaximizeWindowCocoa,
        ___glfwShowWindowCocoa,
        ___glfwHideWindowCocoa,
        ___glfwRequestWindowAttentionCocoa,
        ___glfwFocusWindowCocoa,
        ___glfwSetWindowMonitorCocoa,
        __glfwWindowFocusedCocoa,
        __glfwWindowIconifiedCocoa,
        __glfwWindowVisibleCocoa,
        __glfwWindowMaximizedCocoa,
        __glfwWindowHoveredCocoa,
        __glfwFramebufferTransparentCocoa,
        ___glfwGetWindowOpacityCocoa,
        __glfwSetWindowResizableCocoa,
        __glfwSetWindowDecoratedCocoa,
        __glfwSetWindowFloatingCocoa,
        ___glfwSetWindowOpacityCocoa,
        __glfwSetWindowMousePassthroughCocoa,
        ___glfwPollEventsCocoa,
        ___glfwWaitEventsCocoa,
        ____glfwWaitEventsTimeoutCocoa,
        ___glfwPostEmptyEventCocoa,
        __glfwGetEGLPlatformCocoa,
        __glfwGetEGLNativeDisplayCocoa,
        __glfwGetEGLNativeWindowCocoa,
        ___glfwGetRequiredInstanceExtensionsCocoa,
        ___glfwGetPhysicalDevicePresentationSupportCocoa,
        ____glfwCreateWindowSurfaceCocoa,
    };

    *platform = cocoa;
    return GLFW_TRUE;
}

int ___glfwInitCocoa(void)
{
    @autoreleasepool {

    __glfw.ns.helper = [[GLFWHelper alloc] init];

    [NSThread detachNewThreadSelector:@selector(doNothing:)
                             toTarget:__glfw.ns.helper
                           withObject:nil];

    [NSApplication sharedApplication];

    __glfw.ns.delegate = [[GLFWApplicationDelegate alloc] init];
    if (__glfw.ns.delegate == nil)
    {
        ___glfwInputError(GLFW_PLATFORM_ERROR,
                        "Cocoa: Failed to create application delegate");
        return GLFW_FALSE;
    }

    [NSApp setDelegate:__glfw.ns.delegate];

    NSEvent* (^block)(NSEvent*) = ^ NSEvent* (NSEvent* event)
    {
        if ([event modifierFlags] & NSEventModifierFlagCommand)
            [[NSApp keyWindow] sendEvent:event];

        return event;
    };

    __glfw.ns.keyUpMonitor =
        [NSEvent addLocalMonitorForEventsMatchingMask:NSEventMaskKeyUp
                                              handler:block];

    if (__glfw.hints.init.ns.chdir)
        changeToResourcesDirectory();

    // Press and Hold prevents some keys from emitting repeated characters
    NSDictionary* defaults = @{@"ApplePressAndHoldEnabled":@NO};
    [[NSUserDefaults standardUserDefaults] registerDefaults:defaults];

    [[NSNotificationCenter defaultCenter]
        addObserver:__glfw.ns.helper
           selector:@selector(selectedKeyboardInputSourceChanged:)
               name:NSTextInputContextKeyboardSelectionDidChangeNotification
             object:nil];

    createKeyTables();

    __glfw.ns.eventSource = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
    if (!__glfw.ns.eventSource)
        return GLFW_FALSE;

    CGEventSourceSetLocalEventsSuppressionInterval(__glfw.ns.eventSource, 0.0);

    if (!initializeTIS())
        return GLFW_FALSE;

    __glfwPollMonitorsCocoa();

    if (![[NSRunningApplication currentApplication] isFinishedLaunching])
        [NSApp run];

    // In case we are unbundled, make us a proper UI application
    if (__glfw.hints.init.ns.menubar)
        [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];

    return GLFW_TRUE;

    } // autoreleasepool
}

void ___glfwTerminateCocoa(void)
{
    @autoreleasepool {

    if (__glfw.ns.inputSource)
    {
        CFRelease(__glfw.ns.inputSource);
        __glfw.ns.inputSource = NULL;
        __glfw.ns.unicodeData = nil;
    }

    if (__glfw.ns.eventSource)
    {
        CFRelease(__glfw.ns.eventSource);
        __glfw.ns.eventSource = NULL;
    }

    if (__glfw.ns.delegate)
    {
        [NSApp setDelegate:nil];
        [__glfw.ns.delegate release];
        __glfw.ns.delegate = nil;
    }

    if (__glfw.ns.helper)
    {
        [[NSNotificationCenter defaultCenter]
            removeObserver:__glfw.ns.helper
                      name:NSTextInputContextKeyboardSelectionDidChangeNotification
                    object:nil];
        [[NSNotificationCenter defaultCenter]
            removeObserver:__glfw.ns.helper];
        [__glfw.ns.helper release];
        __glfw.ns.helper = nil;
    }

    if (__glfw.ns.keyUpMonitor)
        [NSEvent removeMonitor:__glfw.ns.keyUpMonitor];

    __glfw_free(__glfw.ns.clipboardString);

    ___glfwTerminateNSGL();
    ____glfwTerminateEGL();
    ____glfwTerminateOSMesa();

    } // autoreleasepool
}

