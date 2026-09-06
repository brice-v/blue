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
val __separator = _separator;
val __hyperlink = _hyperlink;
val __hyperlink_with_handler = _hyperlink_with_handler;
val __icon_widget = _icon_widget;
val __card = _card;
val __accordion_item = _accordion_item;
val __accordion = _accordion;
val __tab_item = _tab_item;
val __tab_item_with_icon = _tab_item_with_icon;
val __tabs = _tabs;
val __doc_tabs = _doc_tabs;
val __scroll = _scroll;
val __hscroll = _hscroll;
val __vscroll = _vscroll;
val __hsplit = _hsplit;
val __vsplit = _vsplit;
val __border = _border;
val __center = _center;
val __padded = _padded;
val __stack = _stack;
val __new_color = _new_color;
val __canvas_rectangle = _canvas_rectangle;
val __canvas_circle = _canvas_circle;
val __canvas_line = _canvas_line;
val __canvas_text = _canvas_text;
val __rich_text = _rich_text;
val __slider_get_value = _slider_get_value;
val __slider_set_value = _slider_set_value;
val __slider = _slider;
val __check_group = _check_group;
val __select_entry = _select_entry;
val __canvas_image = _canvas_image;
val __canvas_arc = _canvas_arc;
val __canvas_polygon = _canvas_polygon;
val __text_grid = _text_grid;
val __activity = _activity;
val __activity_start = _activity_start;
val __activity_stop = _activity_stop;
val __inner_window = _inner_window;
val __file_icon = _file_icon;
val __calendar = _calendar;
val __list = _list;
val __table = _table;
val __canvas_linear_gradient = _canvas_linear_gradient;
val __canvas_radial_gradient = _canvas_radial_gradient;
val __canvas_square = _canvas_square;
val __dialog_info = _dialog_info;
val __grid_wrap = _grid_wrap;
val __tree = _tree;
val __dialog_confirm = _dialog_confirm;
val __dialog_file_open = _dialog_file_open;
val __menu_item = _menu_item;
val __menu = _menu;
val __set_main_menu = _set_main_menu;
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
                id = _widget.v;
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

fun separator() {
    ##std:this,__separator
    ## `separator` returns a themed separator line
    ##
    ## separator() -> {t: 'ui', v: uint}
    __separator()
}

fun hyperlink(text, url) {
    ##std:this,__hyperlink
    ## `hyperlink` returns a clickable hyperlink that opens url in the default browser
    ##
    ## hyperlink(text: str, url: str) -> {t: 'ui', v: uint}
    __hyperlink(text, url)
}

fun hyperlink_with_handler(text, url, on_tapped) {
    ##std:this,__hyperlink_with_handler
    ## `hyperlink_with_handler` returns a hyperlink with a custom tap handler
    ##
    ## hyperlink_with_handler(text: str, url: str, on_tapped: fun()) -> {t: 'ui', v: uint}
    __hyperlink_with_handler(text, url, on_tapped)
}

fun icon_widget(res) {
    ##std:this,__icon_widget
    ## `icon_widget` returns a simple icon widget
    ##
    ## icon_widget(res: GoObj[fyne.Resource]) -> {t: 'ui', v: uint}
    __icon_widget(res)
}

fun card(title, subtitle, content) {
    ##std:this,__card
    ## `card` returns a card container with title, subtitle and content
    ##
    ## card(title: str, subtitle: str, content: {t: 'ui', v: uint}) -> {t: 'ui', v: uint}
    __card(title, subtitle, content)
}

fun accordion_item(title, detail) {
    ##std:this,__accordion_item
    ## `accordion_item` creates an accordion item
    ##
    ## accordion_item(title: str, detail: {t: 'ui', v: uint}) -> GoObj[*widget.AccordionItem]
    __accordion_item(title, detail)
}

fun accordion(items) {
    ##std:this,__accordion
    ## `accordion` returns an accordion widget
    ##
    ## accordion(items: list[GoObj[*widget.AccordionItem]]) -> {t: 'ui', v: uint}
    __accordion(items)
}

fun tab_item(text, content) {
    ##std:this,__tab_item
    ## `tab_item` creates a tab item for tabs/doc_tabs
    ##
    ## tab_item(text: str, content: {t: 'ui', v: uint}) -> GoObj[*container.TabItem]
    __tab_item(text, content)
}

fun tab_item_with_icon(text, icon_res, content) {
    ##std:this,__tab_item_with_icon
    ## `tab_item_with_icon` creates a tab item with an icon
    ##
    ## tab_item_with_icon(text: str, icon: GoObj[fyne.Resource], content: {t: 'ui', v: uint}) -> GoObj[*container.TabItem]
    __tab_item_with_icon(text, icon_res, content)
}

fun tabs(items) {
    ##std:this,__tabs
    ## `tabs` returns an app tabs container (top tab bar)
    ##
    ## tabs(items: list[GoObj[*container.TabItem]]) -> {t: 'ui', v: uint}
    __tabs(items)
}

fun doc_tabs(items) {
    ##std:this,__doc_tabs
    ## `doc_tabs` returns a document tabs container (closable)
    ##
    ## doc_tabs(items: list[GoObj[*container.TabItem]]) -> {t: 'ui', v: uint}
    __doc_tabs(items)
}

fun scroll(content) {
    ##std:this,__scroll
    ## `scroll` wraps content in a scroll container (both axes)
    ##
    ## scroll(content: {t: 'ui', v: uint}) -> {t: 'ui', v: uint}
    __scroll(content)
}

fun hscroll(content) {
    ##std:this,__hscroll
    ## `hscroll` wraps content in a horizontal scroll container
    ##
    ## hscroll(content: {t: 'ui', v: uint}) -> {t: 'ui', v: uint}
    __hscroll(content)
}

fun vscroll(content) {
    ##std:this,__vscroll
    ## `vscroll` wraps content in a vertical scroll container
    ##
    ## vscroll(content: {t: 'ui', v: uint}) -> {t: 'ui', v: uint}
    __vscroll(content)
}

fun hsplit(left, right) {
    ##std:this,__hsplit
    ## `hsplit` returns a horizontal split container with draggable divider
    ##
    ## hsplit(left: {t: 'ui', v: uint}, right: {t: 'ui', v: uint}) -> {t: 'ui', v: uint}
    __hsplit(left, right)
}

fun vsplit(top, bottom) {
    ##std:this,__vsplit
    ## `vsplit` returns a vertical split container with draggable divider
    ##
    ## vsplit(top: {t: 'ui', v: uint}, bottom: {t: 'ui', v: uint}) -> {t: 'ui', v: uint}
    __vsplit(top, bottom)
}

fun border(top, bottom, left, right, center_content) {
    ##std:this,__border
    ## `border` returns a border layout (nullable top/bottom/left/right + mandatory center)
    ##
    ## border(top: {t: 'ui', v: uint}|null, bottom: {t: 'ui', v: uint}|null, left: {t: 'ui', v: uint}|null, right: {t: 'ui', v: uint}|null, center: {t: 'ui', v: uint}) -> {t: 'ui', v: uint}
    __border(top, bottom, left, right, center_content)
}

fun center(content) {
    ##std:this,__center
    ## `center` returns a centered container
    ##
    ## center(content: {t: 'ui', v: uint}) -> {t: 'ui', v: uint}
    __center(content)
}

fun padded(content) {
    ##std:this,__padded
    ## `padded` returns a padded container
    ##
    ## padded(content: {t: 'ui', v: uint}) -> {t: 'ui', v: uint}
    __padded(content)
}

fun stack(children) {
    ##std:this,__stack
    ## `stack` returns a stacked container (overlapping)
    ##
    ## stack(children: list[{t: 'ui', v: uint}]) -> {t: 'ui', v: uint}
    __stack(children)
}

fun new_color(r, g, b, a=255) {
    ##std:this,__new_color
    ## `new_color` creates a color for canvas primitives
    ##
    ## new_color(r: int, g: int, b: int, a: int=255) -> GoObj[color.Color]
    __new_color(r, g, b, a)
}

fun canvas_rectangle(col) {
    ##std:this,__canvas_rectangle
    ## `canvas_rectangle` creates a canvas rectangle primitive
    ##
    ## canvas_rectangle(color: GoObj[color.Color]) -> {t: 'ui', v: uint}
    __canvas_rectangle(col)
}

fun canvas_circle(col) {
    ##std:this,__canvas_circle
    ## `canvas_circle` creates a canvas circle primitive
    ##
    ## canvas_circle(color: GoObj[color.Color]) -> {t: 'ui', v: uint}
    __canvas_circle(col)
}

fun canvas_line(col) {
    ##std:this,__canvas_line
    ## `canvas_line` creates a canvas line primitive
    ##
    ## canvas_line(color: GoObj[color.Color]) -> {t: 'ui', v: uint}
    __canvas_line(col)
}

fun canvas_text(txt, col, size=null) {
    ##std:this,__canvas_text
    ## `canvas_text` creates a canvas text primitive
    ##
    ## canvas_text(text: str, color: GoObj[color.Color], size: int|float|null=null) -> {t: 'ui', v: uint}
    if (size == null) {
        return __canvas_text(txt, col);
    }
    return __canvas_text(txt, col, size);
}

fun rich_text(markdown) {
    ##std:this,__rich_text
    ## `rich_text` creates a rich text widget rendering markdown
    ##
    ## rich_text(markdown: str) -> {t: 'ui', v: uint}
    __rich_text(markdown)
}

fun slider(min, max, value, on_change) {
    ##std:this,__slider,__slider_get_value,__slider_set_value
    ## `slider` returns a slider widget plus helpers get/set
    ##
    ## slider(min: int|float, max: int|float, value: int|float, on_change: fun(value: float)) -> {t: 'ui', v: uint}
    var this = {};
    this.widget = __slider(min, max, value, on_change);
    this.get_value = fun() {
        return __slider_get_value(this.widget);
    };
    this.set_value = fun(v) {
        return __slider_set_value(this.widget, v);
    };
    return this;
}

fun check_group(options, on_change) {
    ##std:this,__check_group
    ## `check_group` returns a multi-select check group
    ##
    ## check_group(options: list[str], on_change: fun(selected: list[str])) -> {t: 'ui', v: uint}
    __check_group(options, on_change)
}

fun select_entry(options, on_change) {
    ##std:this,__select_entry
    ## `select_entry` returns an editable combo box
    ##
    ## select_entry(options: list[str], on_change: fun(value: str)) -> {t: 'ui', v: uint}
    __select_entry(options, on_change)
}

fun canvas_image(path) {
    ##std:this,__canvas_image
    ## `canvas_image` creates an image canvas object from a file path
    ##
    ## canvas_image(path: str) -> {t: 'ui', v: uint}
    __canvas_image(path)
}

fun canvas_arc(col, start_angle, end_angle, cutout=0.0) {
    ##std:this,__canvas_arc
    ## `canvas_arc` creates an arc/circle-sector primitive
    ##
    ## canvas_arc(color: GoObj[color.Color], startAngle: float|int, endAngle: float|int, cutout: float|int=0.0) -> {t: 'ui', v: uint}
    __canvas_arc(col, start_angle, end_angle, cutout)
}

fun canvas_polygon(sides, col) {
    ##std:this,__canvas_polygon
    ## `canvas_polygon` creates a regular polygon primitive
    ##
    ## canvas_polygon(sides: int, color: GoObj[color.Color]) -> {t: 'ui', v: uint}
    __canvas_polygon(sides, col)
}

fun text_grid(content) {
    ##std:this,__text_grid
    ## `text_grid` creates a monospace text grid widget
    ##
    ## text_grid(content: str) -> {t: 'ui', v: uint}
    __text_grid(content)
}

fun activity() {
    ##std:this,__activity,__activity_start,__activity_stop
    ## `activity` creates an activity indicator (auto-started) with start/stop helpers
    ##
    ## activity() -> {widget: {t: 'ui', v: uint}, start: fun(), stop: fun()}
    var this = {};
    this.widget = __activity();
    this.start = fun() {
        return __activity_start(this.widget);
    };
    this.stop = fun() {
        return __activity_stop(this.widget);
    };
    return this;
}

fun inner_window(title, content) {
    ##std:this,__inner_window
    ## `inner_window` creates a draggable inner window
    ##
    ## inner_window(title: str, content: {t: 'ui', v: uint}) -> {t: 'ui', v: uint}
    __inner_window(title, content)
}

fun file_icon(uri) {
    ##std:this,__file_icon
    ## `file_icon` creates a file icon widget for a URI/path
    ##
    ## file_icon(uri: str) -> {t: 'ui', v: uint}
    __file_icon(uri)
}

fun calendar(initial="now", on_change) {
    ##std:this,__calendar
    ## `calendar` returns a calendar widget with onChanged handler (date str YYYY-MM-DD)
    ##
    ## calendar(initial: str='now', on_change: fun(date: str)) -> {t: 'ui', v: uint}
    __calendar(initial, on_change)
}

fun list_items(items, on_selected) {
    ##std:this,__list
    ## `list_items` returns a virtualized scrolling list of strings (alias for list to avoid keyword conflict)
    ##
    ## list_items(items: list[str], on_selected: fun(index: int, value: str)) -> {t: 'ui', v: uint}
    __list(items, on_selected)
}

fun list(items, on_selected) {
    ##std:this,__list
    ## `list` alias for list_items
    __list(items, on_selected)
}

fun table(data, on_selected) {
    ##std:this,__table
    ## `table` returns a table widget from 2D string array
    ##
    ## table(data: list[list[str]], on_selected: fun(row: int, col: int, value: str)) -> {t: 'ui', v: uint}
    __table(data, on_selected)
}

fun canvas_linear_gradient(start, end, angle) {
    ##std:this,__canvas_linear_gradient
    ## `canvas_linear_gradient` creates a linear gradient (angle degrees)
    ##
    ## canvas_linear_gradient(start: GoObj[color.Color], end: GoObj[color.Color], angle: float|int) -> {t: 'ui', v: uint}
    __canvas_linear_gradient(start, end, angle)
}

fun canvas_radial_gradient(start, end) {
    ##std:this,__canvas_radial_gradient
    ## `canvas_radial_gradient` creates a radial gradient (center outward)
    ##
    ## canvas_radial_gradient(start: GoObj[color.Color], end: GoObj[color.Color]) -> {t: 'ui', v: uint}
    __canvas_radial_gradient(start, end)
}

fun canvas_square(col) {
    ##std:this,__canvas_square
    ## `canvas_square` creates a square canvas primitive
    ##
    ## canvas_square(color: GoObj[color.Color]) -> {t: 'ui', v: uint}
    __canvas_square(col)
}

fun dialog_info(title, message) {
    ##std:this,__dialog_info
    ## `dialog_info` shows an information dialog with OK button
    ##
    ## dialog_info(title: str, message: str) -> null
    __dialog_info(title, message)
}

fun dialog_confirm(title, message, on_confirm) {
    ##std:this,__dialog_confirm
    ## `dialog_confirm` shows a confirm dialog calling handler with bool
    ##
    ## dialog_confirm(title: str, message: str, on_confirm: fun(confirmed: bool)) -> null
    __dialog_confirm(title, message, on_confirm)
}

fun dialog_file_open(on_chosen) {
    ##std:this,__dialog_file_open
    ## `dialog_file_open` shows a file-open entry dialog, handler receives path or empty if cancelled
    ##
    ## dialog_file_open(on_chosen: fun(path: str)) -> null
    __dialog_file_open(on_chosen)
}

fun grid_wrap(items, on_selected) {
    ##std:this,__grid_wrap
    ## `grid_wrap` returns a wrapping grid of strings
    ##
    ## grid_wrap(items: list[str], on_selected: fun(index: int, value: str)) -> {t: 'ui', v: uint}
    __grid_wrap(items, on_selected)
}

fun tree(data, root="", on_selected=null) {
    ##std:this,__tree
    ## `tree` returns a hierarchical tree from map[str]list[str]
    ##
    ## tree(data: map[str]list[str], root: str='', on_selected: fun(uid: str)) -> {t: 'ui', v: uint}
    if (on_selected == null) {
        return error("tree requires on_selected handler");
    }
    __tree(data, root, on_selected)
}

fun menu_item(label, on_click) {
    ##std:this,__menu_item
    ## `menu_item` creates a menu item with action
    ##
    ## menu_item(label: str, on_click: fun()) -> GoObj[*fyne.MenuItem]
    __menu_item(label, on_click)
}

fun menu(label, items) {
    ##std:this,__menu
    ## `menu` creates a menu with label and items
    ##
    ## menu(label: str, items: list[GoObj[*fyne.MenuItem]]) -> GoObj[*fyne.Menu]
    __menu(label, items)
}

fun set_main_menu(menus) {
    ##std:this,__set_main_menu
    ## `set_main_menu` sets the application main menu
    ##
    ## set_main_menu(menus: list[GoObj[*fyne.Menu]]) -> null
    __set_main_menu(menus)
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