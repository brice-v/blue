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

#include <unistd.h>
#include <signal.h>
#include <stdint.h>

#include <X11/Xlib.h>
#include <X11/keysym.h>
#include <X11/Xatom.h>
#include <X11/Xresource.h>
#include <X11/Xcursor/Xcursor.h>

// The XRandR extension provides mode setting and gamma control
#include <X11/extensions/Xrandr.h>

// The Xkb extension provides improved keyboard support
#include <X11/XKBlib.h>

// The Xinerama extension provides legacy monitor indices
#include <X11/extensions/Xinerama.h>

// The XInput extension provides raw mouse motion input
#include <X11/extensions/XInput2.h>

// The Shape extension provides custom window shapes
#include <X11/extensions/shape.h>

#define GLX_VENDOR 1
#define GLX_RGBA_BIT 0x00000001
#define GLX_WINDOW_BIT 0x00000001
#define GLX_DRAWABLE_TYPE 0x8010
#define GLX_RENDER_TYPE 0x8011
#define GLX_RGBA_TYPE 0x8014
#define GLX_DOUBLEBUFFER 5
#define GLX_STEREO 6
#define GLX_AUX_BUFFERS 7
#define GLX_RED_SIZE 8
#define GLX_GREEN_SIZE 9
#define GLX_BLUE_SIZE 10
#define GLX_ALPHA_SIZE 11
#define GLX_DEPTH_SIZE 12
#define GLX_STENCIL_SIZE 13
#define GLX_ACCUM_RED_SIZE 14
#define GLX_ACCUM_GREEN_SIZE 15
#define GLX_ACCUM_BLUE_SIZE 16
#define GLX_ACCUM_ALPHA_SIZE 17
#define GLX_SAMPLES 0x186a1
#define GLX_VISUAL_ID 0x800b

#define GLX_FRAMEBUFFER_SRGB_CAPABLE_ARB 0x20b2
#define GLX_CONTEXT_DEBUG_BIT_ARB 0x00000001
#define GLX_CONTEXT_COMPATIBILITY_PROFILE_BIT_ARB 0x00000002
#define GLX_CONTEXT_CORE_PROFILE_BIT_ARB 0x00000001
#define GLX_CONTEXT_PROFILE_MASK_ARB 0x9126
#define GLX_CONTEXT_FORWARD_COMPATIBLE_BIT_ARB 0x00000002
#define GLX_CONTEXT_MAJOR_VERSION_ARB 0x2091
#define GLX_CONTEXT_MINOR_VERSION_ARB 0x2092
#define GLX_CONTEXT_FLAGS_ARB 0x2094
#define GLX_CONTEXT_ES2_PROFILE_BIT_EXT 0x00000004
#define GLX_CONTEXT_ROBUST_ACCESS_BIT_ARB 0x00000004
#define GLX_LOSE_CONTEXT_ON_RESET_ARB 0x8252
#define GLX_CONTEXT_RESET_NOTIFICATION_STRATEGY_ARB 0x8256
#define GLX_NO_RESET_NOTIFICATION_ARB 0x8261
#define GLX_CONTEXT_RELEASE_BEHAVIOR_ARB 0x2097
#define GLX_CONTEXT_RELEASE_BEHAVIOR_NONE_ARB 0
#define GLX_CONTEXT_RELEASE_BEHAVIOR_FLUSH_ARB 0x2098
#define GLX_CONTEXT_OPENGL_NO_ERROR_ARB 0x31b3

typedef XID GLXWindow;
typedef XID GLXDrawable;
typedef struct __GLXFBConfig* GLXFBConfig;
typedef struct __GLXcontext* GLXContext;
typedef void (*__GLXextproc)(void);

typedef XClassHint* (* PFN_XAllocClassHint)(void);
typedef XSizeHints* (* PFN_XAllocSizeHints)(void);
typedef XWMHints* (* PFN_XAllocWMHints)(void);
typedef int (* PFN_XChangeProperty)(Display*,Window,Atom,Atom,int,int,const unsigned char*,int);
typedef int (* PFN_XChangeWindowAttributes)(Display*,Window,unsigned long,XSetWindowAttributes*);
typedef Bool (* PFN_XCheckIfEvent)(Display*,XEvent*,Bool(*)(Display*,XEvent*,XPointer),XPointer);
typedef Bool (* PFN_XCheckTypedWindowEvent)(Display*,Window,int,XEvent*);
typedef int (* PFN_XCloseDisplay)(Display*);
typedef Status (* PFN_XCloseIM)(XIM);
typedef int (* PFN_XConvertSelection)(Display*,Atom,Atom,Atom,Window,Time);
typedef Colormap (* PFN_XCreateColormap)(Display*,Window,Visual*,int);
typedef Cursor (* PFN_XCreateFontCursor)(Display*,unsigned int);
typedef XIC (* PFN_XCreateIC)(XIM,...);
typedef Region (* PFN_XCreateRegion)(void);
typedef Window (* PFN_XCreateWindow)(Display*,Window,int,int,unsigned int,unsigned int,unsigned int,int,unsigned int,Visual*,unsigned long,XSetWindowAttributes*);
typedef int (* PFN_XDefineCursor)(Display*,Window,Cursor);
typedef int (* PFN_XDeleteContext)(Display*,XID,XContext);
typedef int (* PFN_XDeleteProperty)(Display*,Window,Atom);
typedef void (* PFN_XDestroyIC)(XIC);
typedef int (* PFN_XDestroyRegion)(Region);
typedef int (* PFN_XDestroyWindow)(Display*,Window);
typedef int (* PFN_XDisplayKeycodes)(Display*,int*,int*);
typedef int (* PFN_XEventsQueued)(Display*,int);
typedef Bool (* PFN_XFilterEvent)(XEvent*,Window);
typedef int (* PFN_XFindContext)(Display*,XID,XContext,XPointer*);
typedef int (* PFN_XFlush)(Display*);
typedef int (* PFN_XFree)(void*);
typedef int (* PFN_XFreeColormap)(Display*,Colormap);
typedef int (* PFN_XFreeCursor)(Display*,Cursor);
typedef void (* PFN_XFreeEventData)(Display*,XGenericEventCookie*);
typedef int (* PFN_XGetErrorText)(Display*,int,char*,int);
typedef Bool (* PFN_XGetEventData)(Display*,XGenericEventCookie*);
typedef char* (* PFN_XGetICValues)(XIC,...);
typedef char* (* PFN_XGetIMValues)(XIM,...);
typedef int (* PFN_XGetInputFocus)(Display*,Window*,int*);
typedef KeySym* (* PFN_XGetKeyboardMapping)(Display*,KeyCode,int,int*);
typedef int (* PFN_XGetScreenSaver)(Display*,int*,int*,int*,int*);
typedef Window (* PFN_XGetSelectionOwner)(Display*,Atom);
typedef XVisualInfo* (* PFN_XGetVisualInfo)(Display*,long,XVisualInfo*,int*);
typedef Status (* PFN_XGetWMNormalHints)(Display*,Window,XSizeHints*,long*);
typedef Status (* PFN_XGetWindowAttributes)(Display*,Window,XWindowAttributes*);
typedef int (* PFN_XGetWindowProperty)(Display*,Window,Atom,long,long,Bool,Atom,Atom*,int*,unsigned long*,unsigned long*,unsigned char**);
typedef int (* PFN_XGrabPointer)(Display*,Window,Bool,unsigned int,int,int,Window,Cursor,Time);
typedef Status (* PFN_XIconifyWindow)(Display*,Window,int);
typedef Status (* PFN_XInitThreads)(void);
typedef Atom (* PFN_XInternAtom)(Display*,const char*,Bool);
typedef int (* PFN_XLookupString)(XKeyEvent*,char*,int,KeySym*,XComposeStatus*);
typedef int (* PFN_XMapRaised)(Display*,Window);
typedef int (* PFN_XMapWindow)(Display*,Window);
typedef int (* PFN_XMoveResizeWindow)(Display*,Window,int,int,unsigned int,unsigned int);
typedef int (* PFN_XMoveWindow)(Display*,Window,int,int);
typedef int (* PFN_XNextEvent)(Display*,XEvent*);
typedef Display* (* PFN_XOpenDisplay)(const char*);
typedef XIM (* PFN_XOpenIM)(Display*,XrmDatabase*,char*,char*);
typedef int (* PFN_XPeekEvent)(Display*,XEvent*);
typedef int (* PFN_XPending)(Display*);
typedef Bool (* PFN_XQueryExtension)(Display*,const char*,int*,int*,int*);
typedef Bool (* PFN_XQueryPointer)(Display*,Window,Window*,Window*,int*,int*,int*,int*,unsigned int*);
typedef int (* PFN_XRaiseWindow)(Display*,Window);
typedef Bool (* PFN_XRegisterIMInstantiateCallback)(Display*,void*,char*,char*,XIDProc,XPointer);
typedef int (* PFN_XResizeWindow)(Display*,Window,unsigned int,unsigned int);
typedef char* (* PFN_XResourceManagerString)(Display*);
typedef int (* PFN_XSaveContext)(Display*,XID,XContext,const char*);
typedef int (* PFN_XSelectInput)(Display*,Window,long);
typedef Status (* PFN_XSendEvent)(Display*,Window,Bool,long,XEvent*);
typedef int (* PFN_XSetClassHint)(Display*,Window,XClassHint*);
typedef XErrorHandler (* PFN_XSetErrorHandler)(XErrorHandler);
typedef void (* PFN_XSetICFocus)(XIC);
typedef char* (* PFN_XSetIMValues)(XIM,...);
typedef int (* PFN_XSetInputFocus)(Display*,Window,int,Time);
typedef char* (* PFN_XSetLocaleModifiers)(const char*);
typedef int (* PFN_XSetScreenSaver)(Display*,int,int,int,int);
typedef int (* PFN_XSetSelectionOwner)(Display*,Atom,Window,Time);
typedef int (* PFN_XSetWMHints)(Display*,Window,XWMHints*);
typedef void (* PFN_XSetWMNormalHints)(Display*,Window,XSizeHints*);
typedef Status (* PFN_XSetWMProtocols)(Display*,Window,Atom*,int);
typedef Bool (* PFN_XSupportsLocale)(void);
typedef int (* PFN_XSync)(Display*,Bool);
typedef Bool (* PFN_XTranslateCoordinates)(Display*,Window,Window,int,int,int*,int*,Window*);
typedef int (* PFN_XUndefineCursor)(Display*,Window);
typedef int (* PFN_XUngrabPointer)(Display*,Time);
typedef int (* PFN_XUnmapWindow)(Display*,Window);
typedef void (* PFN_XUnsetICFocus)(XIC);
typedef VisualID (* PFN_XVisualIDFromVisual)(Visual*);
typedef int (* PFN_XWarpPointer)(Display*,Window,Window,int,int,unsigned int,unsigned int,int,int);
typedef void (* PFN_XkbFreeKeyboard)(XkbDescPtr,unsigned int,Bool);
typedef void (* PFN_XkbFreeNames)(XkbDescPtr,unsigned int,Bool);
typedef XkbDescPtr (* PFN_XkbGetMap)(Display*,unsigned int,unsigned int);
typedef Status (* PFN_XkbGetNames)(Display*,unsigned int,XkbDescPtr);
typedef Status (* PFN_XkbGetState)(Display*,unsigned int,XkbStatePtr);
typedef KeySym (* PFN_XkbKeycodeToKeysym)(Display*,KeyCode,int,int);
typedef Bool (* PFN_XkbQueryExtension)(Display*,int*,int*,int*,int*,int*);
typedef Bool (* PFN_XkbSelectEventDetails)(Display*,unsigned int,unsigned int,unsigned long,unsigned long);
typedef Bool (* PFN_XkbSetDetectableAutoRepeat)(Display*,Bool,Bool*);
typedef void (* PFN_XrmDestroyDatabase)(XrmDatabase);
typedef Bool (* PFN_XrmGetResource)(XrmDatabase,const char*,const char*,char**,XrmValue*);
typedef XrmDatabase (* PFN_XrmGetStringDatabase)(const char*);
typedef void (* PFN_XrmInitialize)(void);
typedef XrmQuark (* PFN_XrmUniqueQuark)(void);
typedef Bool (* PFN_XUnregisterIMInstantiateCallback)(Display*,void*,char*,char*,XIDProc,XPointer);
typedef int (* PFN_Xutf8LookupString)(XIC,XKeyPressedEvent*,char*,int,KeySym*,Status*);
typedef void (* PFN_Xutf8SetWMProperties)(Display*,Window,const char*,const char*,char**,int,XSizeHints*,XWMHints*,XClassHint*);
#define XAllocClassHint __glfw.x11.xlib.AllocClassHint
#define XAllocSizeHints __glfw.x11.xlib.AllocSizeHints
#define XAllocWMHints __glfw.x11.xlib.AllocWMHints
#define XChangeProperty __glfw.x11.xlib.ChangeProperty
#define XChangeWindowAttributes __glfw.x11.xlib.ChangeWindowAttributes
#define XCheckIfEvent __glfw.x11.xlib.CheckIfEvent
#define XCheckTypedWindowEvent __glfw.x11.xlib.CheckTypedWindowEvent
#define XCloseDisplay __glfw.x11.xlib.CloseDisplay
#define XCloseIM __glfw.x11.xlib.CloseIM
#define XConvertSelection __glfw.x11.xlib.ConvertSelection
#define XCreateColormap __glfw.x11.xlib.CreateColormap
#define XCreateFontCursor __glfw.x11.xlib.CreateFontCursor
#define XCreateIC __glfw.x11.xlib.CreateIC
#define XCreateRegion __glfw.x11.xlib.CreateRegion
#define XCreateWindow __glfw.x11.xlib.CreateWindow
#define XDefineCursor __glfw.x11.xlib.DefineCursor
#define XDeleteContext __glfw.x11.xlib.DeleteContext
#define XDeleteProperty __glfw.x11.xlib.DeleteProperty
#define XDestroyIC __glfw.x11.xlib.DestroyIC
#define XDestroyRegion __glfw.x11.xlib.DestroyRegion
#define XDestroyWindow __glfw.x11.xlib.DestroyWindow
#define XDisplayKeycodes __glfw.x11.xlib.DisplayKeycodes
#define XEventsQueued __glfw.x11.xlib.EventsQueued
#define XFilterEvent __glfw.x11.xlib.FilterEvent
#define XFindContext __glfw.x11.xlib.FindContext
#define XFlush __glfw.x11.xlib.Flush
#define XFree __glfw.x11.xlib.Free
#define XFreeColormap __glfw.x11.xlib.FreeColormap
#define XFreeCursor __glfw.x11.xlib.FreeCursor
#define XFreeEventData __glfw.x11.xlib.FreeEventData
#define XGetErrorText __glfw.x11.xlib.GetErrorText
#define XGetEventData __glfw.x11.xlib.GetEventData
#define XGetICValues __glfw.x11.xlib.GetICValues
#define XGetIMValues __glfw.x11.xlib.GetIMValues
#define XGetInputFocus __glfw.x11.xlib.GetInputFocus
#define XGetKeyboardMapping __glfw.x11.xlib.GetKeyboardMapping
#define XGetScreenSaver __glfw.x11.xlib.GetScreenSaver
#define XGetSelectionOwner __glfw.x11.xlib.GetSelectionOwner
#define XGetVisualInfo __glfw.x11.xlib.GetVisualInfo
#define XGetWMNormalHints __glfw.x11.xlib.GetWMNormalHints
#define XGetWindowAttributes __glfw.x11.xlib.GetWindowAttributes
#define XGetWindowProperty __glfw.x11.xlib.GetWindowProperty
#define XGrabPointer __glfw.x11.xlib.GrabPointer
#define XIconifyWindow __glfw.x11.xlib.IconifyWindow
#define XInternAtom __glfw.x11.xlib.InternAtom
#define XLookupString __glfw.x11.xlib.LookupString
#define XMapRaised __glfw.x11.xlib.MapRaised
#define XMapWindow __glfw.x11.xlib.MapWindow
#define XMoveResizeWindow __glfw.x11.xlib.MoveResizeWindow
#define XMoveWindow __glfw.x11.xlib.MoveWindow
#define XNextEvent __glfw.x11.xlib.NextEvent
#define XOpenIM __glfw.x11.xlib.OpenIM
#define XPeekEvent __glfw.x11.xlib.PeekEvent
#define XPending __glfw.x11.xlib.Pending
#define XQueryExtension __glfw.x11.xlib.QueryExtension
#define XQueryPointer __glfw.x11.xlib.QueryPointer
#define XRaiseWindow __glfw.x11.xlib.RaiseWindow
#define XRegisterIMInstantiateCallback __glfw.x11.xlib.RegisterIMInstantiateCallback
#define XResizeWindow __glfw.x11.xlib.ResizeWindow
#define XResourceManagerString __glfw.x11.xlib.ResourceManagerString
#define XSaveContext __glfw.x11.xlib.SaveContext
#define XSelectInput __glfw.x11.xlib.SelectInput
#define XSendEvent __glfw.x11.xlib.SendEvent
#define XSetClassHint __glfw.x11.xlib.SetClassHint
#define XSetErrorHandler __glfw.x11.xlib.SetErrorHandler
#define XSetICFocus __glfw.x11.xlib.SetICFocus
#define XSetIMValues __glfw.x11.xlib.SetIMValues
#define XSetInputFocus __glfw.x11.xlib.SetInputFocus
#define XSetLocaleModifiers __glfw.x11.xlib.SetLocaleModifiers
#define XSetScreenSaver __glfw.x11.xlib.SetScreenSaver
#define XSetSelectionOwner __glfw.x11.xlib.SetSelectionOwner
#define XSetWMHints __glfw.x11.xlib.SetWMHints
#define XSetWMNormalHints __glfw.x11.xlib.SetWMNormalHints
#define XSetWMProtocols __glfw.x11.xlib.SetWMProtocols
#define XSupportsLocale __glfw.x11.xlib.SupportsLocale
#define XSync __glfw.x11.xlib.Sync
#define XTranslateCoordinates __glfw.x11.xlib.TranslateCoordinates
#define XUndefineCursor __glfw.x11.xlib.UndefineCursor
#define XUngrabPointer __glfw.x11.xlib.UngrabPointer
#define XUnmapWindow __glfw.x11.xlib.UnmapWindow
#define XUnsetICFocus __glfw.x11.xlib.UnsetICFocus
#define XVisualIDFromVisual __glfw.x11.xlib.VisualIDFromVisual
#define XWarpPointer __glfw.x11.xlib.WarpPointer
#define XkbFreeKeyboard __glfw.x11.xkb.FreeKeyboard
#define XkbFreeNames __glfw.x11.xkb.FreeNames
#define XkbGetMap __glfw.x11.xkb.GetMap
#define XkbGetNames __glfw.x11.xkb.GetNames
#define XkbGetState __glfw.x11.xkb.GetState
#define XkbKeycodeToKeysym __glfw.x11.xkb.KeycodeToKeysym
#define XkbQueryExtension __glfw.x11.xkb.QueryExtension
#define XkbSelectEventDetails __glfw.x11.xkb.SelectEventDetails
#define XkbSetDetectableAutoRepeat __glfw.x11.xkb.SetDetectableAutoRepeat
#define XrmDestroyDatabase __glfw.x11.xrm.DestroyDatabase
#define XrmGetResource __glfw.x11.xrm.GetResource
#define XrmGetStringDatabase __glfw.x11.xrm.GetStringDatabase
#define XrmUniqueQuark __glfw.x11.xrm.UniqueQuark
#define XUnregisterIMInstantiateCallback __glfw.x11.xlib.UnregisterIMInstantiateCallback
#define Xutf8LookupString __glfw.x11.xlib.utf8LookupString
#define Xutf8SetWMProperties __glfw.x11.xlib.utf8SetWMProperties

typedef XRRCrtcGamma* (* PFN_XRRAllocGamma)(int);
typedef void (* PFN_XRRFreeCrtcInfo)(XRRCrtcInfo*);
typedef void (* PFN_XRRFreeGamma)(XRRCrtcGamma*);
typedef void (* PFN_XRRFreeOutputInfo)(XRROutputInfo*);
typedef void (* PFN_XRRFreeScreenResources)(XRRScreenResources*);
typedef XRRCrtcGamma* (* PFN_XRRGetCrtcGamma)(Display*,RRCrtc);
typedef int (* PFN_XRRGetCrtcGammaSize)(Display*,RRCrtc);
typedef XRRCrtcInfo* (* PFN_XRRGetCrtcInfo) (Display*,XRRScreenResources*,RRCrtc);
typedef XRROutputInfo* (* PFN_XRRGetOutputInfo)(Display*,XRRScreenResources*,RROutput);
typedef RROutput (* PFN_XRRGetOutputPrimary)(Display*,Window);
typedef XRRScreenResources* (* PFN_XRRGetScreenResourcesCurrent)(Display*,Window);
typedef Bool (* PFN_XRRQueryExtension)(Display*,int*,int*);
typedef Status (* PFN_XRRQueryVersion)(Display*,int*,int*);
typedef void (* PFN_XRRSelectInput)(Display*,Window,int);
typedef Status (* PFN_XRRSetCrtcConfig)(Display*,XRRScreenResources*,RRCrtc,Time,int,int,RRMode,Rotation,RROutput*,int);
typedef void (* PFN_XRRSetCrtcGamma)(Display*,RRCrtc,XRRCrtcGamma*);
typedef int (* PFN_XRRUpdateConfiguration)(XEvent*);
#define XRRAllocGamma __glfw.x11.randr.AllocGamma
#define XRRFreeCrtcInfo __glfw.x11.randr.FreeCrtcInfo
#define XRRFreeGamma __glfw.x11.randr.FreeGamma
#define XRRFreeOutputInfo __glfw.x11.randr.FreeOutputInfo
#define XRRFreeScreenResources __glfw.x11.randr.FreeScreenResources
#define XRRGetCrtcGamma __glfw.x11.randr.GetCrtcGamma
#define XRRGetCrtcGammaSize __glfw.x11.randr.GetCrtcGammaSize
#define XRRGetCrtcInfo __glfw.x11.randr.GetCrtcInfo
#define XRRGetOutputInfo __glfw.x11.randr.GetOutputInfo
#define XRRGetOutputPrimary __glfw.x11.randr.GetOutputPrimary
#define XRRGetScreenResourcesCurrent __glfw.x11.randr.GetScreenResourcesCurrent
#define XRRQueryExtension __glfw.x11.randr.QueryExtension
#define XRRQueryVersion __glfw.x11.randr.QueryVersion
#define XRRSelectInput __glfw.x11.randr.SelectInput
#define XRRSetCrtcConfig __glfw.x11.randr.SetCrtcConfig
#define XRRSetCrtcGamma __glfw.x11.randr.SetCrtcGamma
#define XRRUpdateConfiguration __glfw.x11.randr.UpdateConfiguration

typedef XcursorImage* (* PFN_XcursorImageCreate)(int,int);
typedef void (* PFN_XcursorImageDestroy)(XcursorImage*);
typedef Cursor (* PFN_XcursorImageLoadCursor)(Display*,const XcursorImage*);
typedef char* (* PFN_XcursorGetTheme)(Display*);
typedef int (* PFN_XcursorGetDefaultSize)(Display*);
typedef XcursorImage* (* PFN_XcursorLibraryLoadImage)(const char*,const char*,int);
#define XcursorImageCreate __glfw.x11.xcursor.ImageCreate
#define XcursorImageDestroy __glfw.x11.xcursor.ImageDestroy
#define XcursorImageLoadCursor __glfw.x11.xcursor.ImageLoadCursor
#define XcursorGetTheme __glfw.x11.xcursor.GetTheme
#define XcursorGetDefaultSize __glfw.x11.xcursor.GetDefaultSize
#define XcursorLibraryLoadImage __glfw.x11.xcursor.LibraryLoadImage

typedef Bool (* PFN_XineramaIsActive)(Display*);
typedef Bool (* PFN_XineramaQueryExtension)(Display*,int*,int*);
typedef XineramaScreenInfo* (* PFN_XineramaQueryScreens)(Display*,int*);
#define XineramaIsActive __glfw.x11.xinerama.IsActive
#define XineramaQueryExtension __glfw.x11.xinerama.QueryExtension
#define XineramaQueryScreens __glfw.x11.xinerama.QueryScreens

typedef XID xcb_window_t;
typedef XID xcb_visualid_t;
typedef struct xcb_connection_t xcb_connection_t;
typedef xcb_connection_t* (* PFN_XGetXCBConnection)(Display*);
#define XGetXCBConnection __glfw.x11.x11xcb.GetXCBConnection

typedef Bool (* PFN_XF86VidModeQueryExtension)(Display*,int*,int*);
typedef Bool (* PFN_XF86VidModeGetGammaRamp)(Display*,int,int,unsigned short*,unsigned short*,unsigned short*);
typedef Bool (* PFN_XF86VidModeSetGammaRamp)(Display*,int,int,unsigned short*,unsigned short*,unsigned short*);
typedef Bool (* PFN_XF86VidModeGetGammaRampSize)(Display*,int,int*);
#define XF86VidModeQueryExtension __glfw.x11.vidmode.QueryExtension
#define XF86VidModeGetGammaRamp __glfw.x11.vidmode.GetGammaRamp
#define XF86VidModeSetGammaRamp __glfw.x11.vidmode.SetGammaRamp
#define XF86VidModeGetGammaRampSize __glfw.x11.vidmode.GetGammaRampSize

typedef Status (* PFN_XIQueryVersion)(Display*,int*,int*);
typedef int (* PFN_XISelectEvents)(Display*,Window,XIEventMask*,int);
#define XIQueryVersion __glfw.x11.xi.QueryVersion
#define XISelectEvents __glfw.x11.xi.SelectEvents

typedef Bool (* PFN_XRenderQueryExtension)(Display*,int*,int*);
typedef Status (* PFN_XRenderQueryVersion)(Display*dpy,int*,int*);
typedef XRenderPictFormat* (* PFN_XRenderFindVisualFormat)(Display*,Visual const*);
#define XRenderQueryExtension __glfw.x11.xrender.QueryExtension
#define XRenderQueryVersion __glfw.x11.xrender.QueryVersion
#define XRenderFindVisualFormat __glfw.x11.xrender.FindVisualFormat

typedef Bool (* PFN_XShapeQueryExtension)(Display*,int*,int*);
typedef Status (* PFN_XShapeQueryVersion)(Display*dpy,int*,int*);
typedef void (* PFN_XShapeCombineRegion)(Display*,Window,int,int,int,Region,int);
typedef void (* PFN_XShapeCombineMask)(Display*,Window,int,int,int,Pixmap,int);

#define XShapeQueryExtension __glfw.x11.xshape.QueryExtension
#define XShapeQueryVersion __glfw.x11.xshape.QueryVersion
#define XShapeCombineRegion __glfw.x11.xshape.ShapeCombineRegion
#define XShapeCombineMask __glfw.x11.xshape.ShapeCombineMask

typedef int (*PFNGLXGETFBCONFIGATTRIBPROC)(Display*,GLXFBConfig,int,int*);
typedef const char* (*PFNGLXGETCLIENTSTRINGPROC)(Display*,int);
typedef Bool (*PFNGLXQUERYEXTENSIONPROC)(Display*,int*,int*);
typedef Bool (*PFNGLXQUERYVERSIONPROC)(Display*,int*,int*);
typedef void (*PFNGLXDESTROYCONTEXTPROC)(Display*,GLXContext);
typedef Bool (*PFNGLXMAKECURRENTPROC)(Display*,GLXDrawable,GLXContext);
typedef void (*PFNGLXSWAPBUFFERSPROC)(Display*,GLXDrawable);
typedef const char* (*PFNGLXQUERYEXTENSIONSSTRINGPROC)(Display*,int);
typedef GLXFBConfig* (*PFNGLXGETFBCONFIGSPROC)(Display*,int,int*);
typedef GLXContext (*PFNGLXCREATENEWCONTEXTPROC)(Display*,GLXFBConfig,int,GLXContext,Bool);
typedef __GLXextproc (* PFNGLXGETPROCADDRESSPROC)(const GLubyte *procName);
typedef void (*PFNGLXSWAPINTERVALEXTPROC)(Display*,GLXDrawable,int);
typedef XVisualInfo* (*PFNGLXGETVISUALFROMFBCONFIGPROC)(Display*,GLXFBConfig);
typedef GLXWindow (*PFNGLXCREATEWINDOWPROC)(Display*,GLXFBConfig,Window,const int*);
typedef void (*PFNGLXDESTROYWINDOWPROC)(Display*,GLXWindow);

typedef int (*PFNGLXSWAPINTERVALMESAPROC)(int);
typedef int (*PFNGLXSWAPINTERVALSGIPROC)(int);
typedef GLXContext (*PFNGLXCREATECONTEXTATTRIBSARBPROC)(Display*,GLXFBConfig,GLXContext,Bool,const int*);

// libGL.so function pointer typedefs
#define glXGetFBConfigs __glfw.glx.GetFBConfigs
#define glXGetFBConfigAttrib __glfw.glx.GetFBConfigAttrib
#define glXGetClientString __glfw.glx.GetClientString
#define glXQueryExtension __glfw.glx.QueryExtension
#define glXQueryVersion __glfw.glx.QueryVersion
#define glXDestroyContext __glfw.glx.DestroyContext
#define glXMakeCurrent __glfw.glx.MakeCurrent
#define glXSwapBuffers __glfw.glx.SwapBuffers
#define glXQueryExtensionsString __glfw.glx.QueryExtensionsString
#define glXCreateNewContext __glfw.glx.CreateNewContext
#define glXGetVisualFromFBConfig __glfw.glx.GetVisualFromFBConfig
#define glXCreateWindow __glfw.glx.CreateWindow
#define glXDestroyWindow __glfw.glx.DestroyWindow

typedef VkFlags VkXlibSurfaceCreateFlagsKHR;
typedef VkFlags VkXcbSurfaceCreateFlagsKHR;

typedef struct VkXlibSurfaceCreateInfoKHR
{
    VkStructureType             sType;
    const void*                 pNext;
    VkXlibSurfaceCreateFlagsKHR flags;
    Display*                    dpy;
    Window                      window;
} VkXlibSurfaceCreateInfoKHR;

typedef struct VkXcbSurfaceCreateInfoKHR
{
    VkStructureType             sType;
    const void*                 pNext;
    VkXcbSurfaceCreateFlagsKHR  flags;
    xcb_connection_t*           connection;
    xcb_window_t                window;
} VkXcbSurfaceCreateInfoKHR;

typedef VkResult (APIENTRY *PFN_vkCreateXlibSurfaceKHR)(VkInstance,const VkXlibSurfaceCreateInfoKHR*,const VkAllocationCallbacks*,VkSurfaceKHR*);
typedef VkBool32 (APIENTRY *PFN_vkGetPhysicalDeviceXlibPresentationSupportKHR)(VkPhysicalDevice,uint32_t,Display*,VisualID);
typedef VkResult (APIENTRY *PFN_vkCreateXcbSurfaceKHR)(VkInstance,const VkXcbSurfaceCreateInfoKHR*,const VkAllocationCallbacks*,VkSurfaceKHR*);
typedef VkBool32 (APIENTRY *PFN_vkGetPhysicalDeviceXcbPresentationSupportKHR)(VkPhysicalDevice,uint32_t,xcb_connection_t*,xcb_visualid_t);

#include "xkb_unicode.h"
#include "posix_poll.h"

#define GLFW_X11_WINDOW_STATE           _GLFWwindowX11 x11;
#define GLFW_X11_LIBRARY_WINDOW_STATE   _GLFWlibraryX11 x11;
#define GLFW_X11_MONITOR_STATE          _GLFWmonitorX11 x11;
#define GLFW_X11_CURSOR_STATE           _GLFWcursorX11 x11;

#define GLFW_GLX_CONTEXT_STATE          _GLFWcontextGLX glx;
#define GLFW_GLX_LIBRARY_CONTEXT_STATE  _GLFWlibraryGLX glx;


// GLX-specific per-context data
//
typedef struct _GLFWcontextGLX
{
    GLXContext      handle;
    GLXWindow       window;
} _GLFWcontextGLX;

// GLX-specific global data
//
typedef struct _GLFWlibraryGLX
{
    int             major, minor;
    int             eventBase;
    int             errorBase;

    void*           handle;

    // GLX 1.3 functions
    PFNGLXGETFBCONFIGSPROC              GetFBConfigs;
    PFNGLXGETFBCONFIGATTRIBPROC         GetFBConfigAttrib;
    PFNGLXGETCLIENTSTRINGPROC           GetClientString;
    PFNGLXQUERYEXTENSIONPROC            QueryExtension;
    PFNGLXQUERYVERSIONPROC              QueryVersion;
    PFNGLXDESTROYCONTEXTPROC            DestroyContext;
    PFNGLXMAKECURRENTPROC               MakeCurrent;
    PFNGLXSWAPBUFFERSPROC               SwapBuffers;
    PFNGLXQUERYEXTENSIONSSTRINGPROC     QueryExtensionsString;
    PFNGLXCREATENEWCONTEXTPROC          CreateNewContext;
    PFNGLXGETVISUALFROMFBCONFIGPROC     GetVisualFromFBConfig;
    PFNGLXCREATEWINDOWPROC              CreateWindow;
    PFNGLXDESTROYWINDOWPROC             DestroyWindow;

    // GLX 1.4 and extension functions
    PFNGLXGETPROCADDRESSPROC            GetProcAddress;
    PFNGLXGETPROCADDRESSPROC            GetProcAddressARB;
    PFNGLXSWAPINTERVALSGIPROC           SwapIntervalSGI;
    PFNGLXSWAPINTERVALEXTPROC           SwapIntervalEXT;
    PFNGLXSWAPINTERVALMESAPROC          SwapIntervalMESA;
    PFNGLXCREATECONTEXTATTRIBSARBPROC   CreateContextAttribsARB;
    GLFWbool        SGI_swap_control;
    GLFWbool        EXT_swap_control;
    GLFWbool        MESA_swap_control;
    GLFWbool        ARB_multisample;
    GLFWbool        ARB_framebuffer_sRGB;
    GLFWbool        EXT_framebuffer_sRGB;
    GLFWbool        ARB_create_context;
    GLFWbool        ARB_create_context_profile;
    GLFWbool        ARB_create_context_robustness;
    GLFWbool        EXT_create_context_es2_profile;
    GLFWbool        ARB_create_context_no_error;
    GLFWbool        ARB_context_flush_control;
} _GLFWlibraryGLX;

// X11-specific per-window data
//
typedef struct _GLFWwindowX11
{
    Colormap        colormap;
    Window          handle;
    Window          parent;
    XIC             ic;

    GLFWbool        overrideRedirect;
    GLFWbool        iconified;
    GLFWbool        maximized;

    // Whether the visual supports framebuffer transparency
    GLFWbool        transparent;

    // Cached position and size used to filter out duplicate events
    int             width, height;
    int             xpos, ypos;

    // The last received cursor position, regardless of source
    int             lastCursorPosX, lastCursorPosY;
    // The last position the cursor was warped to by GLFW
    int             warpCursorPosX, warpCursorPosY;

    // The time of the last KeyPress event per keycode, for discarding
    // duplicate key events generated for some keys by ibus
    Time            keyPressTimes[256];
} _GLFWwindowX11;

// X11-specific global data
//
typedef struct _GLFWlibraryX11
{
    Display*        display;
    int             screen;
    Window          root;

    // System content scale
    float           contentScaleX, contentScaleY;
    // Helper window for IPC
    Window          helperWindowHandle;
    // Invisible cursor for hidden cursor mode
    Cursor          hiddenCursorHandle;
    // Context for mapping window XIDs to _GLFWwindow pointers
    XContext        context;
    // XIM input method
    XIM             im;
    // The previous X error handler, to be restored later
    XErrorHandler   errorHandler;
    // Most recent error code received by X error handler
    int             errorCode;
    // Primary selection string (while the primary selection is owned)
    char*           primarySelectionString;
    // Clipboard string (while the selection is owned)
    char*           clipboardString;
    // Key name string
    char            keynames[GLFW_KEY_LAST + 1][5];
    // X11 keycode to GLFW key LUT
    short int       keycodes[256];
    // GLFW key to X11 keycode LUT
    short int       scancodes[GLFW_KEY_LAST + 1];
    // Where to place the cursor when re-enabled
    double          restoreCursorPosX, restoreCursorPosY;
    // The window whose disabled cursor mode is active
    _GLFWwindow*    disabledCursorWindow;
    int             emptyEventPipe[2];

    // Window manager atoms
    Atom            NET_SUPPORTED;
    Atom            NET_SUPPORTING_WM_CHECK;
    Atom            WM_PROTOCOLS;
    Atom            WM_STATE;
    Atom            WM_DELETE_WINDOW;
    Atom            NET_WM_NAME;
    Atom            NET_WM_ICON_NAME;
    Atom            NET_WM_ICON;
    Atom            NET_WM_PID;
    Atom            NET_WM_PING;
    Atom            NET_WM_WINDOW_TYPE;
    Atom            NET_WM_WINDOW_TYPE_NORMAL;
    Atom            NET_WM_STATE;
    Atom            NET_WM_STATE_ABOVE;
    Atom            NET_WM_STATE_FULLSCREEN;
    Atom            NET_WM_STATE_MAXIMIZED_VERT;
    Atom            NET_WM_STATE_MAXIMIZED_HORZ;
    Atom            NET_WM_STATE_DEMANDS_ATTENTION;
    Atom            NET_WM_BYPASS_COMPOSITOR;
    Atom            NET_WM_FULLSCREEN_MONITORS;
    Atom            NET_WM_WINDOW_OPACITY;
    Atom            NET_WM_CM_Sx;
    Atom            NET_WORKAREA;
    Atom            NET_CURRENT_DESKTOP;
    Atom            NET_ACTIVE_WINDOW;
    Atom            NET_FRAME_EXTENTS;
    Atom            NET_REQUEST_FRAME_EXTENTS;
    Atom            MOTIF_WM_HINTS;

    // Xdnd (drag and drop) atoms
    Atom            XdndAware;
    Atom            XdndEnter;
    Atom            XdndPosition;
    Atom            XdndStatus;
    Atom            XdndActionCopy;
    Atom            XdndDrop;
    Atom            XdndFinished;
    Atom            XdndSelection;
    Atom            XdndTypeList;
    Atom            text_uri_list;

    // Selection (clipboard) atoms
    Atom            TARGETS;
    Atom            MULTIPLE;
    Atom            INCR;
    Atom            CLIPBOARD;
    Atom            PRIMARY;
    Atom            CLIPBOARD_MANAGER;
    Atom            SAVE_TARGETS;
    Atom            NULL_;
    Atom            UTF8_STRING;
    Atom            COMPOUND_STRING;
    Atom            ATOM_PAIR;
    Atom            GLFW_SELECTION;

    struct {
        void*       handle;
        GLFWbool    utf8;
        PFN_XAllocClassHint AllocClassHint;
        PFN_XAllocSizeHints AllocSizeHints;
        PFN_XAllocWMHints AllocWMHints;
        PFN_XChangeProperty ChangeProperty;
        PFN_XChangeWindowAttributes ChangeWindowAttributes;
        PFN_XCheckIfEvent CheckIfEvent;
        PFN_XCheckTypedWindowEvent CheckTypedWindowEvent;
        PFN_XCloseDisplay CloseDisplay;
        PFN_XCloseIM CloseIM;
        PFN_XConvertSelection ConvertSelection;
        PFN_XCreateColormap CreateColormap;
        PFN_XCreateFontCursor CreateFontCursor;
        PFN_XCreateIC CreateIC;
        PFN_XCreateRegion CreateRegion;
        PFN_XCreateWindow CreateWindow;
        PFN_XDefineCursor DefineCursor;
        PFN_XDeleteContext DeleteContext;
        PFN_XDeleteProperty DeleteProperty;
        PFN_XDestroyIC DestroyIC;
        PFN_XDestroyRegion DestroyRegion;
        PFN_XDestroyWindow DestroyWindow;
        PFN_XDisplayKeycodes DisplayKeycodes;
        PFN_XEventsQueued EventsQueued;
        PFN_XFilterEvent FilterEvent;
        PFN_XFindContext FindContext;
        PFN_XFlush Flush;
        PFN_XFree Free;
        PFN_XFreeColormap FreeColormap;
        PFN_XFreeCursor FreeCursor;
        PFN_XFreeEventData FreeEventData;
        PFN_XGetErrorText GetErrorText;
        PFN_XGetEventData GetEventData;
        PFN_XGetICValues GetICValues;
        PFN_XGetIMValues GetIMValues;
        PFN_XGetInputFocus GetInputFocus;
        PFN_XGetKeyboardMapping GetKeyboardMapping;
        PFN_XGetScreenSaver GetScreenSaver;
        PFN_XGetSelectionOwner GetSelectionOwner;
        PFN_XGetVisualInfo GetVisualInfo;
        PFN_XGetWMNormalHints GetWMNormalHints;
        PFN_XGetWindowAttributes GetWindowAttributes;
        PFN_XGetWindowProperty GetWindowProperty;
        PFN_XGrabPointer GrabPointer;
        PFN_XIconifyWindow IconifyWindow;
        PFN_XInternAtom InternAtom;
        PFN_XLookupString LookupString;
        PFN_XMapRaised MapRaised;
        PFN_XMapWindow MapWindow;
        PFN_XMoveResizeWindow MoveResizeWindow;
        PFN_XMoveWindow MoveWindow;
        PFN_XNextEvent NextEvent;
        PFN_XOpenIM OpenIM;
        PFN_XPeekEvent PeekEvent;
        PFN_XPending Pending;
        PFN_XQueryExtension QueryExtension;
        PFN_XQueryPointer QueryPointer;
        PFN_XRaiseWindow RaiseWindow;
        PFN_XRegisterIMInstantiateCallback RegisterIMInstantiateCallback;
        PFN_XResizeWindow ResizeWindow;
        PFN_XResourceManagerString ResourceManagerString;
        PFN_XSaveContext SaveContext;
        PFN_XSelectInput SelectInput;
        PFN_XSendEvent SendEvent;
        PFN_XSetClassHint SetClassHint;
        PFN_XSetErrorHandler SetErrorHandler;
        PFN_XSetICFocus SetICFocus;
        PFN_XSetIMValues SetIMValues;
        PFN_XSetInputFocus SetInputFocus;
        PFN_XSetLocaleModifiers SetLocaleModifiers;
        PFN_XSetScreenSaver SetScreenSaver;
        PFN_XSetSelectionOwner SetSelectionOwner;
        PFN_XSetWMHints SetWMHints;
        PFN_XSetWMNormalHints SetWMNormalHints;
        PFN_XSetWMProtocols SetWMProtocols;
        PFN_XSupportsLocale SupportsLocale;
        PFN_XSync Sync;
        PFN_XTranslateCoordinates TranslateCoordinates;
        PFN_XUndefineCursor UndefineCursor;
        PFN_XUngrabPointer UngrabPointer;
        PFN_XUnmapWindow UnmapWindow;
        PFN_XUnsetICFocus UnsetICFocus;
        PFN_XVisualIDFromVisual VisualIDFromVisual;
        PFN_XWarpPointer WarpPointer;
        PFN_XUnregisterIMInstantiateCallback UnregisterIMInstantiateCallback;
        PFN_Xutf8LookupString utf8LookupString;
        PFN_Xutf8SetWMProperties utf8SetWMProperties;
    } xlib;

    struct {
        PFN_XrmDestroyDatabase DestroyDatabase;
        PFN_XrmGetResource GetResource;
        PFN_XrmGetStringDatabase GetStringDatabase;
        PFN_XrmUniqueQuark UniqueQuark;
    } xrm;

    struct {
        GLFWbool    available;
        void*       handle;
        int         eventBase;
        int         errorBase;
        int         major;
        int         minor;
        GLFWbool    gammaBroken;
        GLFWbool    monitorBroken;
        PFN_XRRAllocGamma AllocGamma;
        PFN_XRRFreeCrtcInfo FreeCrtcInfo;
        PFN_XRRFreeGamma FreeGamma;
        PFN_XRRFreeOutputInfo FreeOutputInfo;
        PFN_XRRFreeScreenResources FreeScreenResources;
        PFN_XRRGetCrtcGamma GetCrtcGamma;
        PFN_XRRGetCrtcGammaSize GetCrtcGammaSize;
        PFN_XRRGetCrtcInfo GetCrtcInfo;
        PFN_XRRGetOutputInfo GetOutputInfo;
        PFN_XRRGetOutputPrimary GetOutputPrimary;
        PFN_XRRGetScreenResourcesCurrent GetScreenResourcesCurrent;
        PFN_XRRQueryExtension QueryExtension;
        PFN_XRRQueryVersion QueryVersion;
        PFN_XRRSelectInput SelectInput;
        PFN_XRRSetCrtcConfig SetCrtcConfig;
        PFN_XRRSetCrtcGamma SetCrtcGamma;
        PFN_XRRUpdateConfiguration UpdateConfiguration;
    } randr;

    struct {
        GLFWbool     available;
        GLFWbool     detectable;
        int          majorOpcode;
        int          eventBase;
        int          errorBase;
        int          major;
        int          minor;
        unsigned int group;
        PFN_XkbFreeKeyboard FreeKeyboard;
        PFN_XkbFreeNames FreeNames;
        PFN_XkbGetMap GetMap;
        PFN_XkbGetNames GetNames;
        PFN_XkbGetState GetState;
        PFN_XkbKeycodeToKeysym KeycodeToKeysym;
        PFN_XkbQueryExtension QueryExtension;
        PFN_XkbSelectEventDetails SelectEventDetails;
        PFN_XkbSetDetectableAutoRepeat SetDetectableAutoRepeat;
    } xkb;

    struct {
        int         count;
        int         timeout;
        int         interval;
        int         blanking;
        int         exposure;
    } saver;

    struct {
        int         version;
        Window      source;
        Atom        format;
    } xdnd;

    struct {
        void*       handle;
        PFN_XcursorImageCreate ImageCreate;
        PFN_XcursorImageDestroy ImageDestroy;
        PFN_XcursorImageLoadCursor ImageLoadCursor;
        PFN_XcursorGetTheme GetTheme;
        PFN_XcursorGetDefaultSize GetDefaultSize;
        PFN_XcursorLibraryLoadImage LibraryLoadImage;
    } xcursor;

    struct {
        GLFWbool    available;
        void*       handle;
        int         major;
        int         minor;
        PFN_XineramaIsActive IsActive;
        PFN_XineramaQueryExtension QueryExtension;
        PFN_XineramaQueryScreens QueryScreens;
    } xinerama;

    struct {
        void*       handle;
        PFN_XGetXCBConnection GetXCBConnection;
    } x11xcb;

    struct {
        GLFWbool    available;
        void*       handle;
        int         eventBase;
        int         errorBase;
        PFN_XF86VidModeQueryExtension QueryExtension;
        PFN_XF86VidModeGetGammaRamp GetGammaRamp;
        PFN_XF86VidModeSetGammaRamp SetGammaRamp;
        PFN_XF86VidModeGetGammaRampSize GetGammaRampSize;
    } vidmode;

    struct {
        GLFWbool    available;
        void*       handle;
        int         majorOpcode;
        int         eventBase;
        int         errorBase;
        int         major;
        int         minor;
        PFN_XIQueryVersion QueryVersion;
        PFN_XISelectEvents SelectEvents;
    } xi;

    struct {
        GLFWbool    available;
        void*       handle;
        int         major;
        int         minor;
        int         eventBase;
        int         errorBase;
        PFN_XRenderQueryExtension QueryExtension;
        PFN_XRenderQueryVersion QueryVersion;
        PFN_XRenderFindVisualFormat FindVisualFormat;
    } xrender;

    struct {
        GLFWbool    available;
        void*       handle;
        int         major;
        int         minor;
        int         eventBase;
        int         errorBase;
        PFN_XShapeQueryExtension QueryExtension;
        PFN_XShapeCombineRegion ShapeCombineRegion;
        PFN_XShapeQueryVersion QueryVersion;
        PFN_XShapeCombineMask ShapeCombineMask;
    } xshape;
} _GLFWlibraryX11;

// X11-specific per-monitor data
//
typedef struct _GLFWmonitorX11
{
    RROutput        output;
    RRCrtc          crtc;
    RRMode          oldMode;

    // Index of corresponding Xinerama screen,
    // for EWMH full screen window placement
    int             index;
} _GLFWmonitorX11;

// X11-specific per-cursor data
//
typedef struct _GLFWcursorX11
{
    Cursor handle;
} _GLFWcursorX11;


GLFWbool __glfwConnectX11(int platformID, _GLFWplatform* platform);
int ___glfwInitX11(void);
void ___glfwTerminateX11(void);

GLFWbool ___glfwCreateWindowX11(_GLFWwindow* window, const _GLFWwndconfig* wndconfig, const _GLFWctxconfig* ctxconfig, const _GLFWfbconfig* fbconfig);
void ___glfwDestroyWindowX11(_GLFWwindow* window);
void ___glfwSetWindowTitleX11(_GLFWwindow* window, const char* title);
void ___glfwSetWindowIconX11(_GLFWwindow* window, int count, const GLFWimage* images);
void ___glfwGetWindowPosX11(_GLFWwindow* window, int* xpos, int* ypos);
void ___glfwSetWindowPosX11(_GLFWwindow* window, int xpos, int ypos);
void ___glfwGetWindowSizeX11(_GLFWwindow* window, int* width, int* height);
void ___glfwSetWindowSizeX11(_GLFWwindow* window, int width, int height);
void ____glfwSetWindowSizeLimitsX11(_GLFWwindow* window, int minwidth, int minheight, int maxwidth, int maxheight);
void ___glfwSetWindowAspectRatioX11(_GLFWwindow* window, int numer, int denom);
void ___glfwGetFramebufferSizeX11(_GLFWwindow* window, int* width, int* height);
void ___glfwGetWindowFrameSizeX11(_GLFWwindow* window, int* left, int* top, int* right, int* bottom);
void ___glfwGetWindowContentScaleX11(_GLFWwindow* window, float* xscale, float* yscale);
void ___glfwIconifyWindowX11(_GLFWwindow* window);
void ___glfwRestoreWindowX11(_GLFWwindow* window);
void ___glfwMaximizeWindowX11(_GLFWwindow* window);
void ___glfwShowWindowX11(_GLFWwindow* window);
void ___glfwHideWindowX11(_GLFWwindow* window);
void ___glfwRequestWindowAttentionX11(_GLFWwindow* window);
void ___glfwFocusWindowX11(_GLFWwindow* window);
void ___glfwSetWindowMonitorX11(_GLFWwindow* window, _GLFWmonitor* monitor, int xpos, int ypos, int width, int height, int refreshRate);
GLFWbool __glfwWindowFocusedX11(_GLFWwindow* window);
GLFWbool __glfwWindowIconifiedX11(_GLFWwindow* window);
GLFWbool __glfwWindowVisibleX11(_GLFWwindow* window);
GLFWbool __glfwWindowMaximizedX11(_GLFWwindow* window);
GLFWbool __glfwWindowHoveredX11(_GLFWwindow* window);
GLFWbool __glfwFramebufferTransparentX11(_GLFWwindow* window);
void __glfwSetWindowResizableX11(_GLFWwindow* window, GLFWbool enabled);
void __glfwSetWindowDecoratedX11(_GLFWwindow* window, GLFWbool enabled);
void __glfwSetWindowFloatingX11(_GLFWwindow* window, GLFWbool enabled);
float ___glfwGetWindowOpacityX11(_GLFWwindow* window);
void ___glfwSetWindowOpacityX11(_GLFWwindow* window, float opacity);
void __glfwSetWindowMousePassthroughX11(_GLFWwindow* window, GLFWbool enabled);

void __glfwSetRawMouseMotionX11(_GLFWwindow *window, GLFWbool enabled);
GLFWbool ___glfwRawMouseMotionSupportedX11(void);

void ___glfwPollEventsX11(void);
void ___glfwWaitEventsX11(void);
void ____glfwWaitEventsTimeoutX11(double timeout);
void ___glfwPostEmptyEventX11(void);

void ___glfwGetCursorPosX11(_GLFWwindow* window, double* xpos, double* ypos);
void ____glfwSetCursorPosX11(_GLFWwindow* window, double xpos, double ypos);
void ___glfwSetCursorModeX11(_GLFWwindow* window, int mode);
const char* __glfwGetScancodeNameX11(int scancode);
int ____glfwGetKeyScancodeX11(int key);
GLFWbool ____glfwCreateCursorX11(_GLFWcursor* cursor, const GLFWimage* image, int xhot, int yhot);
GLFWbool ___glfwCreateStandardCursorX11(_GLFWcursor* cursor, int shape);
void ___glfwDestroyCursorX11(_GLFWcursor* cursor);
void ___glfwSetCursorX11(_GLFWwindow* window, _GLFWcursor* cursor);
void ___glfwSetClipboardStringX11(const char* string);
const char* ___glfwGetClipboardStringX11(void);

EGLenum __glfwGetEGLPlatformX11(EGLint** attribs);
EGLNativeDisplayType __glfwGetEGLNativeDisplayX11(void);
EGLNativeWindowType __glfwGetEGLNativeWindowX11(_GLFWwindow* window);

void ___glfwGetRequiredInstanceExtensionsX11(char** extensions);
GLFWbool ___glfwGetPhysicalDevicePresentationSupportX11(VkInstance instance, VkPhysicalDevice device, uint32_t queuefamily);
VkResult ____glfwCreateWindowSurfaceX11(VkInstance instance, _GLFWwindow* window, const VkAllocationCallbacks* allocator, VkSurfaceKHR* surface);

void ___glfwFreeMonitorX11(_GLFWmonitor* monitor);
void ___glfwGetMonitorPosX11(_GLFWmonitor* monitor, int* xpos, int* ypos);
void ___glfwGetMonitorContentScaleX11(_GLFWmonitor* monitor, float* xscale, float* yscale);
void ___glfwGetMonitorWorkareaX11(_GLFWmonitor* monitor, int* xpos, int* ypos, int* width, int* height);
GLFWvidmode* ____glfwGetVideoModesX11(_GLFWmonitor* monitor, int* count);
void ___glfwGetVideoModeX11(_GLFWmonitor* monitor, GLFWvidmode* mode);
GLFWbool ___glfwGetGammaRampX11(_GLFWmonitor* monitor, GLFWgammaramp* ramp);
void ____glfwSetGammaRampX11(_GLFWmonitor* monitor, const GLFWgammaramp* ramp);

void ___glfwPollMonitorsX11(void);
void ___glfwSetVideoModeX11(_GLFWmonitor* monitor, const GLFWvidmode* desired);
void ___glfwRestoreVideoModeX11(_GLFWmonitor* monitor);

Cursor __glfwCreateNativeCursorX11(const GLFWimage* image, int xhot, int yhot);

unsigned long ___glfwGetWindowPropertyX11(Window window,
                                        Atom property,
                                        Atom type,
                                        unsigned char** value);
GLFWbool ___glfwIsVisualTransparentX11(Visual* visual);

void ___glfwGrabErrorHandlerX11(void);
void ___glfwReleaseErrorHandlerX11(void);
void ____glfwInputErrorX11(int error, const char* message);

void ___glfwPushSelectionToManagerX11(void);
void __glfwCreateInputContextX11(_GLFWwindow* window);

GLFWbool ____glfwInitGLX(void);
void ____glfwTerminateGLX(void);
GLFWbool ___glfwCreateContextGLX(_GLFWwindow* window,
                               const _GLFWctxconfig* ctxconfig,
                               const _GLFWfbconfig* fbconfig);
void __glfwDestroyContextGLX(_GLFWwindow* window);
GLFWbool ___glfwChooseVisualGLX(const _GLFWwndconfig* wndconfig,
                              const _GLFWctxconfig* ctxconfig,
                              const _GLFWfbconfig* fbconfig,
                              Visual** visual, int* depth);

