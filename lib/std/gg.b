## `gg` is the module that contains functions needed to
## run the games graphics library (Raylib)


val __init_window = _init_window;
val __window_should_close = _window_should_close;
val __close_window = _close_window;
val __is_window_ready = _is_window_ready;
val __is_window_fullscreen = _is_window_fullscreen;
val __is_window_hidden = _is_window_hidden;
val __is_window_maximized = _is_window_maximized;
val __is_window_minimized = _is_window_minimized;
val __is_window_focused = _is_window_focused;
val __is_window_resized = _is_window_resized;
val __toggle_window_fullscreen = _toggle_window_fullscreen;
val __maximize_window = _maximize_window;
val __minimize_window = _minimize_window;
val __restore_window = _restore_window;
val __set_window_icon = _set_window_icon;
val __set_window_title = _set_window_title;
val __set_window_position = _set_window_position;
val __set_window_monitor = _set_window_monitor;
val __set_window_min_size = _set_window_min_size;
val __set_window_size = _set_window_size;
val __set_window_opacity = _set_window_opacity;
val __set_window_focused = _set_window_focused;
val get_screen_width = _get_screen_width;
val get_screen_height = _get_screen_height;
val __get_monitor_count = _get_monitor_count;
val __get_current_monitor = _get_current_monitor;
val __get_monitor_position = _get_monitor_position;
val __get_monitor_width = _get_monitor_width;
val __get_monitor_height = _get_monitor_height;
val __get_monitor_physical_width = _get_monitor_physical_width;
val __get_monitor_physical_height = _get_monitor_physical_height;
val __get_monitor_refresh_rate = _get_monitor_refresh_rate;
val __get_monitor_name = _get_monitor_name;

val __get_clipboard_text = _get_clipboard_text;
val __set_clipboard_text = _set_clipboard_text;

val __show_cursor = _show_cursor;
val __hide_cursor = _hide_cursor;
val __is_cursor_hidden = _is_cursor_hidden;
val __enable_cursor = _enable_cursor;
val __disable_cursor = _disable_cursor;
val __is_cursor_on_screen = _is_cursor_on_screen;

val __clear_background = _clear_background;
val color = _color_map();
val begin_drawing = _begin_drawing;
val end_drawing = _end_drawing;
val set_target_fps = _set_target_fps;
val get_fps = _get_fps;
val get_frame_time = _get_frame_time;
val get_time = _get_time;

val unload = _unload;
val __draw_text = _draw_text;
val __draw_texture = _draw_texture;
val __draw_texture_pro = _draw_texture_pro;
val load_texture = _load_texture;

# Audio/Music/Sound
val init_audio_device = _init_audio_device;
val close_audio_device = _close_audio_device;
val load_music = _load_music;
val update_music = _update_music;
val play_music = _play_music;
val stop_music = _stop_music;
val pause_music = _pause_music;
val resume_music = _resume_music;
val load_sound = _load_sound;
val play_sound = _play_sound;
val stop_sound = _stop_sound;
val resume_sound = _resume_sound;
val pause_sound = _pause_sound;

val set_exit_key = _set_exit_key;
val is_key_up = _is_key_up;
val is_key_down = _is_key_down;
val is_key_pressed = _is_key_pressed;
val is_key_released = _is_key_released;

val __is_mouse_button_pressed = _is_mouse_button_pressed;
val __is_mouse_button_down = _is_mouse_button_down;
val __is_mouse_button_released = _is_mouse_button_released;
val __is_mouse_button_up = _is_mouse_button_up;
val __get_mouse_x = _get_mouse_x;
val __get_mouse_y = _get_mouse_y;
val __get_mouse_position = _get_mouse_position;
val __get_mouse_delta = _get_mouse_delta;
val __set_mouse_position = _set_mouse_position;
val __set_mouse_offset = _set_mouse_offset;
val __set_mouse_scale = _set_mouse_scale;
val __get_mouse_wheel_move = _get_mouse_wheel_move;
val __get_mouse_wheel_move_v = _get_mouse_wheel_move_v;
val __set_mouse_cursor = _set_mouse_cursor;

val __rectangle = _rectangle;
val __vector2 = _vector2;
val __vector3 = _vector3;
val __vector4 = _vector4;
val __camera2d = _camera2d;
val __camera3d = _camera3d;

val __begin_mode2d = _begin_mode2d;
val end_mode2d = _end_mode2d;
val __begin_mode3d = _begin_mode3d;
val end_mode3d = _end_mode3d;

val __draw_rectangle = _draw_rectangle;
val __draw_rectangle_gradient = _draw_rectangle_gradient;
val __draw_rectangle_lines = _draw_rectangle_lines;
val __draw_rectangle_rounded = _draw_rectangle_rounded;
val __draw_rectangle_rounded_lines = _draw_rectangle_rounded_lines;
val __draw_triangle = _draw_triangle;
val __draw_triangle_fan = _draw_triangle_fan;
val __draw_triangle_strip = _draw_triangle_strip;
val __draw_poly = _draw_poly;

# TODO: Once we have more check_collision functions, just make it standalone
val __rectangle_check_collision = _rectangle_check_collision;

# Drawing
val draw = {
    'rectangle': __draw_rectangle,
    'rectangle_gradient': fun(a,b,c,d,e,f=null,is_vertical=true) {
        __draw_rectangle_gradient(a,b,c,d,e,f,is_vertical);
    },
    'rectangle_lines': __draw_rectangle_lines,
    'rectangle_rounded': __draw_rectangle_rounded,
    'rectangle_rounded_lines': __draw_rectangle_rounded_lines,
    'triangle': fun(a,b,c,d,with_lines=false) {
        __draw_triangle(a,b,c,d,with_lines)
    },
    'triangle_fan': __draw_triangle_fan,
    'triangle_strip': __draw_triangle_strip,
    'poly': fun(a,b,c,d,e,f=null,with_lines=false) {
        __draw_poly(a,b,c,d,e,f,with_lines)
    },
}

# Input Constants

val Key = {
    # Keyboard Function Keys
    'Space': 32,
    'Escape': 256,
    'Enter': 257,
    'Tab': 258,
    'Backspace': 259,
    'Insert': 260,
    'Delete': 261,
    'Right': 262,
    'Left': 263,
    'Down': 264,
    'Up': 265,
    'PageUp': 266,
    'PageDown': 267,
    'Home': 268,
    'End': 269,
    'CapsLock': 280,
    'ScrollLock': 281,
    'NumLock': 282,
    'PrintScreen': 283,
    'Pause': 284,
    'F1': 290,
    'F2': 291,
    'F3': 292,
    'F4': 293,
    'F5': 294,
    'F6': 295,
    'F7': 296,
    'F8': 297,
    'F9': 298,
    'F10': 299,
    'F11': 300,
    'F12': 301,
    'LeftShift': 340,
    'LeftControl': 341,
    'LeftAlt': 342,
    'LeftSuper': 343,
    'RightShift': 344,
    'RightControl': 345,
    'RightAlt': 346,
    'RightSuper': 347,
    'KbMenu': 348,
    'LeftBracket': 91,
    'BackSlash': 92,
    'RightBracket': 93,
    'Grave': 96,

    # Keyboard Number Pad Keys
    'Kp0': 320,
    'Kp1': 321,
    'Kp2': 322,
    'Kp3': 323,
    'Kp4': 324,
    'Kp5': 325,
    'Kp6': 326,
    'Kp7': 327,
    'Kp8': 328,
    'Kp9': 329,
    'KpDecimal': 330,
    'KpDivide': 331,
    'KpMultiply': 332,
    'KpSubtract': 333,
    'KpAdd': 334,
    'KpEnter': 335,
    'KpEqual': 336,

    # Keyboard Alpha Numeric Keys
    'Apostrophe': 39,
    'Comma': 44,
    'Minus': 45,
    'Period': 46,
    'Slash': 47,
    'Zero': 48,
    'One': 49,
    'Two': 50,
    'Three': 51,
    'Four': 52,
    'Five': 53,
    'Six': 54,
    'Seven': 55,
    'Eight': 56,
    'Nine': 57,
    'Semicolon': 59,
    'Equal': 61,
    'A': 65,
    'B': 66,
    'C': 67,
    'D': 68,
    'E': 69,
    'F': 70,
    'G': 71,
    'H': 72,
    'I': 73,
    'J': 74,
    'K': 75,
    'L': 76,
    'M': 77,
    'N': 78,
    'O': 79,
    'P': 80,
    'Q': 81,
    'R': 82,
    'S': 83,
    'T': 84,
    'U': 85,
    'V': 86,
    'W': 87,
    'X': 88,
    'Y': 89,
    'Z': 90,

    # Android keys
    'Back': 4,
    'Menu': 82,
    'VolumeUp': 24,
    'VolumeDown': 25,
};

val mouse = {
    'button': {
        'LeftButton': 0,
        'Left': 0,
        'RightButton': 1,
        'Right': 1,
        'MiddleButton': 2,
        'Middle': 2,
        'SideButton': 3,
        'Side': 3,
        'ExtraButton': 4,
        'Extra': 4,
        'ForwardButton': 5,
        'Forward': 5,
        'BackButton': 6,
        'Back': 6,
    },
    'is_button_pressed': __is_mouse_button_pressed,
    'is_button_released': __is_mouse_button_released,
    'is_button_down': __is_mouse_button_down,
    'is_button_up': __is_mouse_button_up,
    'get_x': __get_mouse_x,
    'get_y': __get_mouse_y,
    'get_position': __get_mouse_position,
    'get_delta': __get_mouse_delta,
    'set_position': __set_mouse_position,
    'set_offset': __set_mouse_offset,
    'set_scale': __set_mouse_scale,
    'get_wheel_move': fun(as_vector=false) {
        if as_vector {
            return __get_mouse_wheel_move_v;
        }
        return __get_mouse_wheel_move;
    },
    'set_cursor': __set_mouse_cursor,
};

# Touch points registered
val MaxTouchPoints = 2;

val Gamepad = {
    # Gamepad Number
    'Player1': 0,
    'Player2': 1,
    'Player3': 2,
    'Player4': 3,
    
    # Gamepad Buttons/Axis
    
    # PS3 USB Controller Buttons
    'Ps3ButtonTriangle': 0,
    'Ps3ButtonCircle': 1,
    'Ps3ButtonCross': 2,
    'Ps3ButtonSquare': 3,
    'Ps3ButtonL1': 6,
    'Ps3ButtonR1': 7,
    'Ps3ButtonL2': 4,
    'Ps3ButtonR2': 5,
    'Ps3ButtonStart': 8,
    'Ps3ButtonSelect': 9,
    'Ps3ButtonUp': 24,
    'Ps3ButtonRight': 25,
    'Ps3ButtonDown': 26,
    'Ps3ButtonLeft': 27,
    'Ps3ButtonPs': 12,
    
    # PS3 USB Controller Axis
    'Ps3AxisLeftX': 0,
    'Ps3AxisLeftY': 1,
    'Ps3AxisRightX': 2,
    'Ps3AxisRightY': 5,
    # [1..-1] (pressure-level)
    'Ps3AxisL2': 3,
    # [1..-1] (pressure-level)
    'Ps3AxisR2': 4,
    
    # Xbox360 USB Controller Buttons
    'XboxButtonA': 0,
    'XboxButtonB': 1,
    'XboxButtonX': 2,
    'XboxButtonY': 3,
    'XboxButtonLb': 4,
    'XboxButtonRb': 5,
    'XboxButtonSelect': 6,
    'XboxButtonStart': 7,
    'XboxButtonUp': 10,
    'XboxButtonRight': 11,
    'XboxButtonDown': 12,
    'XboxButtonLeft': 13,
    'XboxButtonHome': 8,
    
    # Android Gamepad Controller (SNES CLASSIC)
    'AndroidDpadUp': 19,
    'AndroidDpadDown': 20,
    'AndroidDpadLeft': 21,
    'AndroidDpadRight': 22,
    'AndroidDpadCenter': 23,
    
    'AndroidButtonA': 96,
    'AndroidButtonB': 97,
    'AndroidButtonC': 98,
    'AndroidButtonX': 99,
    'AndroidButtonY': 100,
    'AndroidButtonZ': 101,
    'AndroidButtonL1': 102,
    'AndroidButtonR1': 103,
    'AndroidButtonL2': 104,
    'AndroidButtonR2': 105,
    
    # Xbox360 USB Controller Axis
    # [-1..1] (left->right)
    'XboxAxisLeftX': 0,
    # [1..-1] (up->down)
    'XboxAxisLeftY': 1,
    # [-1..1] (left->right)
    'XboxAxisRightX': 2,
    # [1..-1] (up->down)
    'XboxAxisRightY': 3,
    # [-1..1] (pressure-level)
    'XboxAxisLt': 4,
    # [-1..1] (pressure-level)
    'XboxAxisRt': 5,
};

val monitor = {
    'get_count': __get_monitor_count,
    'get_current': __get_current_monitor,
    'get_position': __get_monitor_position,
    'get_width': __get_monitor_width,
    'get_height': __get_monitor_height,
    'get_physical_width': __get_monitor_physical_width,
    'get_physical_height': __get_monitor_physical_height,
    'get_monitor_refresh_rate': __get_monitor_refresh_rate,
    'get_name': __get_monitor_name
};

val clipboard = {
    'get_text': __get_clipboard_text,
    'set_text': __set_clipboard_text
};

val cursor = {
    'show': __show_cursor,
    'hide': __hide_cursor,
    'is_hidden': __is_cursor_hidden,
    'enable': __enable_cursor,
    'disable': __disable_cursor,
    'is_on_screen': __is_cursor_on_screen
};

# GG Objects

fun Rectangle(x=0.0, y=0.0, width=0.0, height=0.0) {
    ##std:this,__rectangle
    ## `Rectangle` is an object constructor that represents the GO_OBJ for rectangles
    ##
    ## the object can be modified using the x,y,width,height variables and the GO_OBJ
    ## can be retrieved with this.obj()
    ##
    ## Rectangle(x: float=0.0, y: float=0.0, width: float=0.0, height: float=0.0) ->
    ## {'x':float,'y':float,'width':float,'height':float,'obj':|this|=>GO_OBJ[rl.Rectangle]}
    var this = {
        'x': x,
        'y': y,
        'width': width,
        'height': height,
    };
    this.obj = fun() {
        return __rectangle(float(this['x']), float(this['y']), float(this['width']), float(this['height']));
    }
    this.draw = fun(c=color.red) {
        __draw_rectangle(int(this.x), int(this.y), int(this.width), int(this.height), c);
    }
    this.check_collision = fun(rec) {
        if ('obj' notin rec) {
            return error("rec must have obj() function on its map");
        }
        return __rectangle_check_collision(this.obj(), rec.obj());
    }
    return this;
}

fun Vector(x=0.0, y=0.0, z=null, w=null) {
    ##std:this,__vector2,__vector3,__vector4
    ## `Vector` is an object constructor that represents the GO_OBJ for a 2, 3, or 4 Point Vector
    ## currently there is no way to edit it so any vec will just need to be recreated with .obj()
    ##
    ## Vector() -> will return a 2 point vector
    ## Vector(z=0.0) -> will return a 3 point vector
    ## Vector(w=0.0) -> returns an error (invalid args [needs a z as well])
    ##
    ## Vector(x: float=0.0, y: float=0.0, z: float=null, w: float=null)
    ##     -> {'x':float,'y':float,'obj':|this|=>GO_OBJ[rl.Vector2]}
    ##      | {'x':float,'y':float,'z':float,'obj':|this|=>GO_OBJ[rl.Vector3]}
    ##      | {'x':float,'y':float,'z':float,'w':float,'obj':|this|=>GO_OBJ[rl.Vector4]}
    var this = {
        'x': float(x),
        'y': float(y),
    };
    if (z == null && w == null) {
        this.obj = fun() {
            return __vector2(float(this['x']), float(this['y']))
        };
    } else if (z != null && w == null) {
        this['z'] = float(z);
        this.obj = fun() {
            return __vector3(float(this['x']), float(this['y']), float(this['z']))
        };
    } else if (z != null && w != null) {
        this['z'] = float(z);
        this['w'] = float(w);
        this.obj = fun() {
            return __vector4(float(this['x']), float(this['y']), float(this['z']), float(this['w']))
        };
    } else {
        return error("gg.Vector has invalid arguments");
    }
    return this;
}

fun Camera2D(offset=Vector(), target=Vector(), rotation=0.0, zoom=1.0) {
    ##std:this,__camera2d
    ## `Camera2D` is an object constructor that represents the GO_OBJ for 2D Cameras
    ##
    ## offset (displacement from target)
	## target (rotation and zoom origin)
	## rotation in degrees
	## zoom (scaling), should be 1.0f by default
    ##
    ## Camera2D(offset: Vector2, target: Vector2, rotation: float, zoom: float) 
    ##  -> {'offset':Vector(x,y)=Vector(),'target':Vector(x,y)=Vector(),'rotation':float=0.0,'zoom':float=1.0,'obj':|this|=>GO_OBJ[rl.Camera2D]}
    if ('obj' notin offset) {
        return error("offset must have obj() function on its map");
    }
    if ('obj' notin target) {
        return error("target must have obj() function on its map");
    }
    var this = {
        'offset': offset,
        'target': target,
        'rotation': float(rotation),
        'zoom': float(zoom),
    };
    this.obj = fun() {
        return __camera2d(this['offset']['obj'](), this['target']['obj'](), float(this['rotation']), float(this['zoom']));
    }
    return this;
}

val CameraProjection = {
    'Perspective': 0,
    'Orthographic': 1,
};

fun Camera3D(position=Vector(z=0.0), target=Vector(z=0.0), up=Vector(z=0.0), fovy=0.0, projection=CameraProjection.Perspective) {
    ##std:this,__camera3d
    ## `Camera3D` is an object constructor that represents the GO_OBJ for 2D Cameras
    ##
    ## position - cameras position
	## target - where it looks-at
	## up vector (rotation over its axis)
	## fovy - field-of-view apperture in Y (degrees) in perspective, used as near plane width in orthographic
    ## projection - Camera type, controlling projection type, either CameraProjection.Perspective or CameraProjection.Orthographic
    ##
    ## Camera3D(position: Vector3, target: Vector3, up: Vector3, fovy: float, projection: int[CameraProjection.Perspective|Orthographic]) 
    ##  -> {'position':Vector(x,y,z)=Vector(z=0.0),
    ##      'target':Vector(x,y,z)=Vector(z=0.0),
    ##      'up':Vector(x,y,z)=Vector(z=0.0),
    ##      'fovy':float=0.0,
    ##      'projection':int=CameraProjection.Perspective,
    ##      'obj':|this|=>GO_OBJ[rl.Camera3D]}
    if ('obj' notin position) {
        return error("position must have obj() function on its map");
    }
    if ('obj' notin target) {
        return error("target must have obj() function on its map");
    }
    if ('obj' notin up) {
        return error("up must have obj() function on its map");
    }
    var this = {
        'position': position,
        'target': target,
        'up': up,
        'fovy': float(fovy),
        'projection': int(projection),
    };
    this.obj = fun() {
        return __camera3d(this['position']['obj'](), this['target']['obj'](), this['up']['obj'](), float(this['fovy']), int(this['projection']));
    }
    return this;
}

# Specialized Public Functions

fun new_window(width=800, height=600, title="gg - example app") {
    ##std:this,__init_window
    ## `new_window` will initialize the raylib window
    ##
    ## new_window(width: int=800, height: int=600, title: str="gg - example app") -> null
    __init_window(width, height, title);
    return {
        'is_ready': __is_window_ready,
        'is_fullscreen': __is_window_fullscreen,
        'is_hidden': __is_window_hidden,
        'is_maximized': __is_window_maximized,
        'is_minimized': __is_window_minimized,
        'is_focused': __is_window_focused,
        'is_resized': __is_window_resized,
        'toggle_fullscreen': __toggle_window_fullscreen,
        'maximize': __maximize_window,
        'minimize': __minimize_window,
        'restore': __restore_window,
        'close': __close_window,
        'should_close': __window_should_close,
        'set_icon': __set_window_icon,
        'set_icons': __set_window_icon,
        'set_title': __set_window_title,
        'set_position': __set_window_position,
        'set_monitor': __set_window_monitor,
        'set_min_size': __set_window_min_size,
        'set_size': __set_window_size,
        'set_opacity': __set_window_opacity,
        'set_focused': __set_window_focused
    };
}

fun clear_background(clear_color=color.white) {
    ##std:this,__clear_background
    ## `clear_background` will use the clear color to clear the background with that color
    ##
    ## clear_background(clear_color: GO_OBJ[rl.Color]=color.white) -> null
    __clear_background(clear_color)
}

fun draw_text(text, pos_x=0, pos_y=0, font_size=20, text_color=color.black) {
    ##std:this,__draw_text
    ## `draw_text` will draw the given text to the raylib window with the given parameters
    ##
    ## pos_x the x position of the text
    ## pos_y the y position of the text
    ## font_size the size of the font
    ## text_color the color of the text
    ##
    ## draw_text(text: str, pos_x: int=0, pos_y: int=0, font_size: int=20, text_color: GO_OBJ[rl.Color]=color.black) -> null
    __draw_text(text, pos_x, pos_y, font_size, text_color)
}

fun draw_texture(texture, pos_x=0, pos_y=0, tint=color.white) {
    ##std:this,__draw_texture
    ## `draw_texture` will draw the given 2D texture to the raylib window with the given parameters
    ##
    ## texture is the 2D texture, likely a file loaded with `load_texture`
    ## pos_x the x position of the texture
    ## pos_y the y position of the texture
    ## tint is the color shading to give to the texture (color.white should leave it as default)
    ##
    ## draw_texture(texture: GO_OBJ[rl.Texture2D], pos_x: int=0, pos_y: int=0, tint: GO_OBJ[rl.Color]=color.white) -> null
    __draw_texture(texture, pos_x, pos_y, tint)
}

fun draw_texture_pro(texture, source_rec=Rectangle().obj(), dest_rec=Rectangle().obj(), origin=Vector().obj(), rotation=0.0, tint=color.white) {
    ##std:this,__draw_texture_pro
    ## `draw_texture_pro` will draw the given 2D texture to the raylib window with the given parameters defined
    ## by a rectangle with 'pro' parameters
    ##
    ## texture is the 2D texture, likely a file loaded with `load_texture`
    ## source_rec is the rectangle position of the source 
    ## dest_rec is the rectangle position of the destination
    ## origin is the point where to get the texture from
    ## rotation is the amount the texture should be rotated by
    ## tint is the color shading to give to the texture (color.white should leave it as default)
    ##
    ## draw_texture_pro(texture: GO_OBJ[rl.Texture2D],
    ##                  source_rec: GO_OBJ[rl.Rectangle]=Rectangle().obj(),
    ##                  dest_rec: GO_OBJ[rl.Rectangle]=Rectangle().obj(),
    ##                  origin: GO_OBJ[rl.Vector2]=Vector().obj(),
    ##                  rotation: float=0.0, tint=color.white)
    var src_rec = if (type(source_rec) == Type.MAP && source_rec['obj'] != null) { source_rec.obj() } else { source_rec };
    var dst_rec = if (type(dest_rec) == Type.MAP && dest_rec['obj'] != null) { dest_rec.obj() } else { dest_rec };
    var org_vec2 = if (type(origin) == Type.MAP && origin['obj'] != null) { origin.obj() } else { origin };
    __draw_texture_pro(texture, src_rec, dst_rec, org_vec2, float(rotation), tint)
}

fun begin_mode2d(cam=Camera2D()) {
    ##std:this,__begin_mode2d
    ## `begin_mode2d` will start the 2d mode for an initialized 2d camera
    ##
    ## begin_mode2d(cam: Camera2D=Camera2D()) -> null
    if ('obj' notin cam) {
        return error("cam must have obj() function on its map");
    }
    __begin_mode2d(cam.obj())
}

fun begin_mode3d(cam=Camera3D()) {
    ##std:this,__begin_mode3d
    ## `begin_mode3d` will start the 3d mode for an initialized 3d camera
    ##
    ## begin_mode3d(cam: Camera3D=Camera3D()) -> null
    if ('obj' notin cam) {
        return error("cam must have obj() function on its map");
    }
    __begin_mode3d(cam.obj())
}
