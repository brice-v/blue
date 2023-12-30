## `ui` is the module that contains ui related functions
## the ui is built using fyne.io and some functions are currently
## setup
##
## this allows the user to create a very basic ui for some scripts
## that need a graphical user interface and still interact with
## blue code
##
## when initalized via window() the _app will be created in this module
## to attach and show the window

val __new_app = _new_app;
var __app = null;
val _app = fun() {
    if (__app == null) {
        __app = __new_app();
    }
    return __app;
}();
val __window = _window;
# Layout
val __row = _row;
val __col = _col;
val __grid = _grid;
# Widgets
val label = _label;
val __button = _button;
val __entry = _entry;
val __entry_get_text = _entry_get_text;
val __entry_set_text = _entry_set_text;
val __checkbox = _check_box;
val __radio_group = _radio_group;
val __option_select = _option_select;
val __progress_bar = _progress_bar;
val __progress_bar_set_value = _progress_bar_set_value;
# Form
val __form = _form;
val __append_form = _append_form;

fun window(width=400, height=400, title="blue ui window", content=null) {
    ##std:this,__window
    ## `window` is the main method that should be called when declaring a ui
    ## it will create, show, and run a ui that has a root content layout/widget/form
    ##
    ## width and height is the size of the ui window to be displayed
    ## title is the title displayed at the appropriate area in the os
    ## content is a ui object {t: 'ui', v: uint}, all widgets/layouts will return
    ## this object, as well as a form
    ##
    ## window(width: int=400, height: int=400, title: str='blue ui window', content: {t: 'ui', v: uint}) ->
    ##        {t: 'ui', v: uint}
    if (content == null) {
        return error("`window` content was not given");
    }
    var id = null;
    if (type(content) == Type.GO_OBJ) {
        id = content;
    } else if (type(content) == Type.MAP) {
        if ('_form' in content) {
            id = content._form;
        } else if ('widget' in content) {
            id = content.widget;
        }
    } else {
        return error("`window` expects content to be a GO_OBJ[fyne.CanvasObject] or MAP, got=#{type(content)}");
    }
    return __window(_app, width, height, title, id);
}

fun row(children=[]) {
    ##std:this,__row
    ## `row` is a layout function for the ui that accepts a list of layouts/widgets/forms
    ##
    ## the layout for row is vertical such that the first item is on top of the second item
    ##
    ## row(children: list[{t: "ui*", v: uint}]=[]) -> {t: "ui", v: uint}
    if (children.len() == 0) {
        return error("ui row: children length should be greater than 0")
    }

    for (child in children) {
        if (type(child) == Type.GO_OBJ) {
            continue;
        } else if (type(child) == Type.MAP) {
            if ("widget" notin child) {
                return error("ui row: 'widget' not found in child #{child}");
            }
        } else {
            return error("ui row: Unexpected child #{child}");
        }
    }
    var ids = [];
    for (child in children) {
        if (type(child) == Type.GO_OBJ) {
            ids << child;
            continue;
        }
        if ("widget" in child) {
            ids << child.widget;
        }
    }
    # get the ids of all the child 'canvas object elements' to put into the row
    __row(ids)
}

fun col(children=[]) {
    ##std:this,__col
    ## `col` is a layout function for the ui that accepts a list of layouts/widgets/forms
    ##
    ## the layout for col is horizontal such that the first item is to the left of the second item
    ##
    ## col(children: list[{t: "ui*", v: uint}]=[]) -> {t: "ui", v: uint}
    if (children.len() == 0) {
        return error("ui col: children length should be greater than 0")
    }

    for (child in children) {
        if (type(child) == Type.GO_OBJ) {
            continue;
        } else if (type(child) == Type.MAP) {
            if ("widget" notin child) {
                return error("ui col: 'widget' not found in child #{child}");
            }
        } else {
            return error("ui col: Unexpected child #{child}");
        }
    }
    var ids = [];
    for (child in children) {
        if (type(child) == Type.GO_OBJ) {
            ids << child;
            continue;
        }
        if ("widget" in child) {
            ids << child.widget;
        }
    }
    # get the ids of all the child 'canvas object elements' to put into the col
    __col(ids)
}

val GridType = {
    COLS: 'COLS',
    ROWS: 'ROWS'
};
fun grid(rowcols, t=GridType.COLS, children=[]) {
    ##std:this,__grid
    ## `grid` is a layout function for the ui that accepts a list of layouts/widgets/forms
    ##
    ## the layout for children is dependent on the grid type [t] (either GridType.COLS or GridType.ROWS)
    ## as well as the rowcols value which determins the # of rows, or cols
    ##
    ## grid(rowcols: int, t: 'ROWS'|'COLS', children: list[{t: "ui*", v: uint}]=[]) -> {t: "ui", v: uint}
    if (children.len() == 0) {
        return error("ui grid: children length should be greater than 0")
    }

    for (child in children) {
        if (type(child) == Type.GO_OBJ) {
            continue;
        } else if (type(child) == Type.MAP) {
            if ("widget" notin child) {
                return error("ui grid: 'widget' not found in child #{child}");
            }
        } else {
            return error("ui grid: Unexpected child #{child}");
        }
    }
    var ids = [];
    for (child in children) {
        if (type(child) == Type.GO_OBJ) {
            ids << child;
            continue;
        }
        if ("widget" in child) {
            ids << child.widget;
        }
    }

    # get the ids of all the child 'canvas object elements' to put into the col
    __grid(rowcols, t, ids)
}

fun button(button_label_str, on_click_fun) {
    ##std:this,__button
    ## `button` will create a button widget with a label and function that responds on click
    ##
    ## button(button_label_str: str, on_click_fun: fun) -> {t: 'ui', v: uint}
    __button(button_label_str, on_click_fun)
}

fun entry(is_multiline=false, placeholder="") {
    ##std:this,__entry
    ## `entry` is a ui widget that returns an input
    ##
    ## this input can be used with the core method get_text to retrieve the string
    ## value inside of it
    ##
    ## is_multiline is a boolean to determine if the entry should support multiline
    ##
    ## entry(is_multiline: bool=false) -> {t: "ui/entry", v: uint}
    var this = {};
    this.widget = __entry(is_multiline, placeholder);

    this.set_text = fun(value) {
        return __entry_set_text(this.widget, value);
    };
    this.get_text = fun() {
        return __entry_get_text(this.widget);
    };

    return this;
}

fun checkbox(checkbox_label, on_change_fun) {
    ##std:this,__checkbox
    ## `checkbox` will create a checkbox widget with the given label and a 
    ## function thats called on change
    ##
    ## note: the on_change_fun handler should take 1 arg which is true or false
    ## depending on the checkbox state
    ##
    ## checkbox(checkbox_label: str, on_change_fun: fun) -> {t: 'ui/check', v: uint}
    __checkbox(checkbox_label, on_change_fun)
}

fun radio_group(options, on_change_fun) {
    ##std:this,__radio_group
    ## `radio_group` will create a radio_group widget with the given options and
    ## a function thats called on change
    ##
    ## note: the on_change_fun handler should take 1 arg which is the string value
    ## of the option selected in the radio group
    ##
    ## radio_group(options: list[str], on_change_fun: fun) -> {t: 'ui/radio', v: uint}
    __radio_group(options, on_change_fun)
}

fun option_select(options, on_change_fun) {
    ##std:this,__option_select
    ## `option_select` will create a option_select widget with the given options
    ## and a function thats called on change
    ##
    ## note: the on_change_fun handler should take 1 arg which is the string value
    ## of the option selected in the option select
    ##
    ## option_select(options: list[str], on_change_fun: fun) -> {t: 'ui/option', v: uint}
    __option_select(options, on_change_fun)
}

fun form(children=[], on_submit=null) {
    ##std:this,__form
    ## `form` is a ui object that can be used to group together labels with ui elements
    ## with an on_submit function
    ##
    ## children should be a list of objects with the shape {'label': _, 'widget': _}
    ## label will just be a string, widget should be a widget object
    ##
    ## on_submit is just a regular function that will be called when submitted
    ##
    ## form(children: list[{label: str, widget: {t: 'ui*', v: uint}}]=[]) -> {t: 'ui', v: uint}
    var this = {};
    if (on_submit == null) {
        return error("`form` on_submit handler was not given");
    }
    for (child in children) {
        match child {
            {'label': _, 'elem': _} => {
                if (type(child.elem) == Type.GO_OBJ) {
                    continue;
                } else if (type(child.elem) == Type.MAP) {
                    if ('t' in child.elem) {
                        if ("ui" notin child.elem.t) {
                            return error("`form` children elements should all be {t: '*ui*', v: _}. got=`#{child.elem}`");
                        }
                    } else if ('widget' notin child.elem) {
                        return error("`form` children elements should all have a 'widget'");
                    }
                }
            },
            _ => {
                return error("`form` children should match {label: _, elem: {t: '*ui*', v: _}}. got=`#{child}`");
            },
        };
    }
    var labels = [];
    var widgets = [];
    if (children.len() > 0) {
        for (child in children) {
            labels << child.label;
            if (type(child.elem) == Type.GO_OBJ) {
                widgets << child.elem;
            } else {
                if ('v' in child.elem) {
                    widgets << child.elem.v;
                } else if ('widget' in child.elem) {
                    widgets << child.elem.widget;
                }
            }
        }
    }
    this._form = __form(labels, widgets, on_submit);
    this.append_form = fun(label, _widget) {
        var id = null;
        if (type(_widget) == Type.GO_OBJ) {
            id = _widget;
        } else {
            if ('v' in _widget) {
                id = widget.v;
            } else if ('widget' in _widget) {
                id = _widget.widget;
            }
        }
        return __append_form(this._form, label, id);
    }
    return this;
}

fun progress_bar(is_infinite=false) {
    ##std:this,__progress_bar,__progress_bar_set_value
    ## `progress_bar` is a ui object that returns a progress_bar object map
    ## that has a function to set its value if its not infinite
    ##
    ## is_infinite: bool that determines whether this is an infinite progress bar
    ## set_value: sets this progress bar to the given float value
    ##
    ## progress_bar(is_infinite: bool=false) -> {'widget': this, set_value: fun(v: float)->null}
    var this = {};

    this.widget = __progress_bar(is_infinite);

    this.set_value = fun(value) {
        if (is_infinite) {
            return null;
        }
        return __progress_bar_set_value(this.widget.v, value);
    }
    return this;
}

val toolbar = {
    'new': _toolbar,
    'spacer': _toolbar_spacer,
    'separator': _toolbar_separator,
    'action': _toolbar_action
};

val icon = {
    'account': _icon_account(),
    'cancel': _icon_cancel(),
    'check_button_checked': _icon_check_button_checked(),
    'check_button': _icon_check_button(),
    'color_achromatic': _icon_color_achromatic(),
    'color_chromatic': _icon_color_chromatic(),
    'color_palette': _icon_color_palette(),
    'computer': _icon_computer(),
    'confirm': _icon_confirm(),
    'content_add': _icon_content_add(),
    'content_clear': _icon_content_clear(),
    'content_copy': _icon_content_copy(),
    'content_cut': _icon_content_cut(),
    'content_paste': _icon_content_paste(),
    'content_redo': _icon_content_redo(),
    'content_remove': _icon_content_remove(),
    'content_undo': _icon_content_undo(),
    'delete': _icon_delete(),
    'document_create': _icon_document_create(),
    'document': _icon_document(),
    'document_print': _icon_document_print(),
    'document_save': _icon_document_save(),
    'download': _icon_download(),
    'error': _icon_error(),
    'file_application': _icon_file_application(),
    'file_audio': _icon_file_audio(),
    'file': _icon_file(),
    'file_image': _icon_file_image(),
    'file_text': _icon_file_text(),
    'file_video': _icon_file_video(),
    'folder': _icon_folder(),
    'folder_new': _icon_folder_new(),
    'folder_open': _icon_folder_open(),
    'grid': _icon_grid(),
    'help': _icon_help(),
    'history': _icon_history(),
    'home': _icon_home(),
    'info': _icon_info(),
    'list': _icon_list(),
    'login': _icon_login(),
    'logout': _icon_logout(),
    'mail_attachment': _icon_mail_attachment(),
    'mail_compose': _icon_mail_compose(),
    'mail_forward': _icon_mail_forward(),
    'mail_reply_all': _icon_mail_reply_all(),
    'mail_reply': _icon_mail_reply(),
    'mail_send': _icon_mail_send(),
    'media_fast_forward': _icon_media_fast_forward(),
    'media_fast_rewind': _icon_media_fast_rewind(),
    'media_music': _icon_media_music(),
    'media_pause': _icon_media_pause(),
    'media_photo': _icon_media_photo(),
    'media_play': _icon_media_play(),
    'media_record': _icon_media_record(),
    'media_replay': _icon_media_replay(),
    'media_skip_next': _icon_media_skip_next(),
    'media_skip_previous': _icon_media_skip_previous(),
    'media_stop': _icon_media_stop(),
    'media_video': _icon_media_video(),
    'menu_drop_down': _icon_menu_drop_down(),
    'menu_drop_up': _icon_menu_drop_up(),
    'menu_expand': _icon_menu_expand(),
    'menu': _icon_menu(),
    'more_horizontal': _icon_more_horizontal(),
    'more_vertical': _icon_more_vertical(),
    'move_down': _icon_move_down(),
    'move_up': _icon_move_up(),
    'navigate_back': _icon_navigate_back(),
    'navigate_next': _icon_navigate_next(),
    'question': _icon_question(),
    'radio_button_checked': _icon_radio_button_checked(),
    'radio_button': _icon_radio_button(),
    'search': _icon_search(),
    'search_replace': _icon_search_replace(),
    'settings': _icon_settings(),
    'storage': _icon_storage(),
    'upload': _icon_upload(),
    'view_full_screen': _icon_view_full_screen(),
    'view_refresh': _icon_view_refresh(),
    'view_restore': _icon_view_restore(),
    'visibility': _icon_visibility(),
    'visibility_off': _icon_visibility_off(),
    'volume_down': _icon_volume_down(),
    'volume_mute': _icon_volume_mute(),
    'volume_up': _icon_volume_up(),
    'warning': _icon_warning(),
    'zoom_fit': _icon_zoom_fit(),
    'zoom_in': _icon_zoom_in(),
    'zoom_out': _icon_zoom_out(),
};