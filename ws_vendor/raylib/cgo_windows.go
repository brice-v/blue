//go:build windows
// +build windows

package rl

/*
#include "external/glfw/src/context.c"
#include "external/glfw/src/init.c"
#include "external/glfw/src/input.c"
#include "external/glfw/src/monitor.c"
#include "external/glfw/src/platform.c"
#include "external/glfw/src/vulkan.c"
#include "external/glfw/src/window.c"

#include "external/glfw/src/win32_init.c"
#include "external/glfw/src/win32_joystick.c"
#include "external/glfw/src/win32_module.c"
#include "external/glfw/src/win32_monitor.c"
#include "external/glfw/src/win32_thread.c"
#include "external/glfw/src/win32_time.c"
#include "external/glfw/src/win32_window.c"
#include "external/glfw/src/wgl_context.c"
#include "external/glfw/src/egl_context.c"
#include "external/glfw/src/osmesa_context.c"

#cgo windows LDFLAGS: -lgdi32 -lwinmm -lole32
#cgo windows CFLAGS: -D_GLFW_WIN32 -Iexternal -Iexternal/glfw/include -Iexternal/glfw/deps/mingw -DPLATFORM_DESKTOP

#cgo windows,!angle LDFLAGS: -lopengl32

#cgo windows,opengl11,!angle CFLAGS: -DGRAPHICS_API_OPENGL_11
#cgo windows,opengl21,!angle CFLAGS: -DGRAPHICS_API_OPENGL_21
#cgo windows,opengl43,!angle CFLAGS: -DGRAPHICS_API_OPENGL_43
#cgo windows,!opengl11,!opengl21,!opengl43,!angle CFLAGS: -DGRAPHICS_API_OPENGL_33
#cgo windows,angle CFLAGS: -DGRAPHICS_API_OPENGL_ES2
*/
import "C"
