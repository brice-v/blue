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
val button = _button;
val __entry = _entry;
val entry_get_text = _entry_get_text;
val checkbox = _check_box;
val radio_group = _radio_group;
val option_select = _option_select;
# Form
val __form = _form;
val __append_form = _append_form;

fun window(width=400, height=400, title="blue ui window", content) {
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

fun entry(is_multiline=false) {
    __entry(is_multiline)
}

fun form(children=[], on_submit) {
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
    if ("ui" notin form_obj.t) {
        return error("`append_form` error: form_obj must be {t: '*ui*', v: _}. got=`#{form_obj}`")
    }
    if ("ui" notin widget.t) {
        return error("`append_form` error: widget must be {t: '*ui*', v: _}. got=`#{widget}`")
    }
    __append_form(form_obj.v, label, widget.v)
}