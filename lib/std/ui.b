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
# Widgets
val label = _label;
val __button = _button;
val __entry = _entry;
val __entry_get_text = _entry_get_text;
val __checkbox = _check_box;
val __radio_group = _radio_group;
val __option_select = _option_select;
# Form
val __form = _form;
val __append_form = _append_form;

fun window(width=400, height=400, title="blue ui window", content) {
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
    return match content {
        {t: "ui", v: _} => {
            __window(_app, width, height, title, content.v)
        },
        _ => {
            error("ui window: content type was not 'ui', got=#{content}")
        },
    };
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
        if ("ui" notin child.t) {
            return error("ui row: found child without 'ui' type, got=#{child}");
        }
    }

    var ids = [child.v for (child in children)];
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
        if ("ui" notin child.t) {
            return error("ui col: found child without 'ui' type, got=#{child}");
        }
    }

    var ids = [child.v for (child in children)];
    # get the ids of all the child 'canvas object elements' to put into the col
    __col(ids)
}

# Note: Cant use this function as the it will be called first when we do child.label
# instead of using the string value (for form())
###fun label(label_str) {
    ## `label` will create a label widget with the given string
    ##
    ## label(label_str: str) -> {t: 'ui', v: uint}
    __label(label_str)
}###

fun button(button_label_str, on_click_fun) {
    ## `button` will create a button widget with a label and function that responds on click
    ##
    ## button(button_label_str: str, on_click_fun: fun) -> {t: 'ui', v: uint}
    __button(button_label_str, on_click_fun)
}

fun entry(is_multiline=false) {
    ## `entry` is a ui widget that returns an input
    ##
    ## this input can be used with the core method get_text to retrieve the string
    ## value inside of it
    ##
    ## is_multiline is a boolean to determine if the entry should support multiline
    ##
    ## entry(is_multiline: bool=false) -> {t: "ui/entry", v: uint}
    __entry(is_multiline)
}

fun entry_get_text(entry_id) {
    ## `entry_get_text` gets the text from an entry widget
    ## note: this function should mostly be called with the core function 'get_text'
    ##
    ## NOTE: this can only be called on an entry_id belonging to an entry object
    ## ie. {t: 'ui/entry', v: _}
    ##
    ## entry_get_text(entry_id: uint) -> str
    __entry_get_text(entry_id)
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

fun form(children=[], on_submit) {
    ## `form` is a ui object that can be used to group together labels with ui elements
    ## with an on_submit function
    ##
    ## children should be a list of objects with the shape {'label': _, 'widget': _}
    ## label will just be a string, widget should be a widget object
    ##
    ## on_submit is just a regular function that will be called when submitted
    ##
    ## form(children: list[{label: str, widget: {t: 'ui*', v: uint}}]=[]) -> {t: 'ui', v: uint}
    for (child in children) {
        match child {
            {'label': _, 'elem': _} => {
                if ("ui" notin child.elem.t) {
                    return error("`form` children elements should all be {t: '*ui*', v: _}. got=`#{child.elem}`")
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
            labels = labels.append(child.label);
            widgets = widgets.append(child.elem.v);
        }
    }
    __form(labels, widgets, on_submit)
}

fun append_form(form_obj, label, widget) {
    ## `append_form` will take a form object and append a label and widget
    ##
    ## form_obj is an object that represents the form with shape {t: 'ui', v: uint}
    ## label is a string
    ## widget is a widget object with the shape {t: 'ui*', v: uint}
    ##
    ## append_form(form_obj: {t: 'ui', v: uint}, label: str, widget: {t: 'ui*', v: uint}) -> null
    if ("ui" notin form_obj.t) {
        return error("`append_form` error: form_obj must be {t: '*ui*', v: _}. got=`#{form_obj}`")
    }
    if ("ui" notin widget.t) {
        return error("`append_form` error: widget must be {t: '*ui*', v: _}. got=`#{widget}`")
    }
    __append_form(form_obj.v, label, widget.v)
}