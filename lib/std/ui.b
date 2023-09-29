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
    }
    return __window(_app, width, height, title, id);
}

fun row(children=[]) {
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
    ## `button` will create a button widget with a label and function that responds on click
    ##
    ## button(button_label_str: str, on_click_fun: fun) -> {t: 'ui', v: uint}
    __button(button_label_str, on_click_fun)
}

fun entry(is_multiline=false, placeholder="") {
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
