## `gg` is the module that contains functions needed to
## run the games graphics library (Raylib)


val __init_window = _init_window;
val __clear_background = _clear_background;
val color = _color_map();
val begin_drawing = _begin_drawing;
val end_drawing = _end_drawing;
val set_target_fps = _set_target_fps;
val window_should_close = _window_should_close;
val close_window = _close_window;
val __draw_text = _draw_text;
val __draw_texture = _draw_texture;
val __draw_texture_pro = _draw_texture_pro;
val load_texture = _load_texture;
val unload_texture = _unload_texture;
val init_audio_device = _init_audio_device;
val close_audio_device = _close_audio_device;
val load_music = _load_music;
val unload_music = _unload_music;
val update_music = _update_music;
val play_music = _play_music;
val stop_music = _stop_music;
val pause_music = _pause_music;
val resume_music = _resume_music;

val set_exit_key = _set_exit_key;
val is_key_up = _is_key_up;
val is_key_down = _is_key_down;
val is_key_pressed = _is_key_pressed;
val is_key_released = _is_key_released;

val __rectangle = _rectangle;
val __vector2 = _vector2;
val __vector3 = _vector3;
val __vector4 = _vector4;


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

val Mouse = {
    # Mouse Buttons'
    'LeftButton': 0,
    'RightButton': 1,
    'MiddleButton': 2,
    'SideButton': 3,
    'ExtraButton': 4,
    'ForwardButton': 5,
    'BackButton': 6,
};

# Touch points registered
val MaxTouchPoints = 2;

val Gamepad = {
    # Gamepad Number'
    'Player1': 0,
    'Player2': 1,
    'Player3': 2,
    'Player4': 3,
    
    # Gamepad Buttons/Axis'
    
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

# GG Objects

fun Rectangle(x=0.0, y=0.0, width=0.0, height=0.0) {
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
    return this;
}

fun Vector(x=0.0, y=0.0, z=null, w=null) {
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

# Specialized Public Functions

fun init_window(width=800, height=600, title="gg - example app") {
    ## `init_window` will initialize the raylib window
    ##
    ## init_window(width: int=800, height: int=600, title: str="gg - example app") -> null
    __init_window(width, height, title)
}

fun clear_background(clear_color=color.white) {
    ## `clear_background` will use the clear color to clear the background with that color
    ##
    ## clear_background(clear_color: GO_OBJ[rl.Color]=color.white) -> null
    __clear_background(clear_color)
}

fun draw_text(text, pos_x=0, pos_y=0, font_size=20, text_color=color.black) {
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
