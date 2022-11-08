val __new_app = _new_app;
val _app = __new_app();
val __window = _window;
val __row = _row;
val __col = _col;
val label = _label;
val button = _button;
val entry = _entry;
val entry_get_text = _entry_get_text;

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