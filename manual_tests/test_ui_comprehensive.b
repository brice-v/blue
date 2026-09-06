import ui

# =============================================================
#  Comprehensive manual UI test - covers every public ui.* name
#  Run with:  blue manual_tests/test_ui_comprehensive.b
#  Close the window to exit. Watch stdout for handler prints.
# =============================================================

println("=== BLUE COMPREHENSIVE UI TEST ===")
println("This window exercises the entire expanded fyne api surface.")
println("Interact with widgets, watch stdout, verify visually.")

# ---------- shared handlers ----------

fun on_button() {
    println("[handler] button clicked")
}

fun on_checkbox(v) {
    println("[handler] checkbox => #{v}")
}

fun on_radio(v) {
    println("[handler] radio => #{v}")
}

fun on_select(v) {
    println("[handler] option_select => #{v}")
}

fun on_check_group(v) {
    println("[handler] check_group => #{v} len=#{v.len()}")
}

fun on_select_entry(v) {
    println("[handler] select_entry => #{v}")
}

fun on_slider(v) {
    println("[handler] slider => #{v}")
}

fun on_hyperlink() {
    println("[handler] hyperlink_with_handler tapped!")
}

fun on_submit() {
    println("[handler] form submitted")
}

fun on_calendar(d) {
    println("[handler] calendar => #{d}")
}

fun on_list(i, v) {
    println("[handler] list => index #{i} value #{v}")
}

fun on_table(r, c, v) {
    println("[handler] table => row #{r} col #{c} = #{v}")
}

# ---------- entries (old coverage) ----------

var e1 = ui.entry(placeholder="single line")
var e2 = ui.entry(is_multiline=true, placeholder="multiline")
e1.set_text("hello blue")
println("entry get_text: #{e1.get_text()}")

var entry_row = ui.row([ui.label("Entries:"), e1.widget, e2.widget])

# ---------- buttons / checkbox / radio / select ----------

var btn = ui.button("Click Me!", on_button)
var chk = ui.checkbox("Check me", on_checkbox)
var rad = ui.radio_group(["Alpha", "Beta", "Gamma"], on_radio)
var sel = ui.option_select(["Red", "Green", "Blue"], on_select)

# ---------- new selector widgets ----------

var slider_obj = ui.slider(0, 100, 42, on_slider)
var slider_label = ui.label("Slider (0-100):")
# progress bar for slider feedback
var prog = ui.progress_bar(false)
# update prog when slider moves - wrap original handler
# slider already has handler; additional demo: set value button
fun slider_set_80() {
    slider_obj.set_value(80)
    prog.set_value(0.8)
    println("[action] slider set to 80, prog 0.8")
}
var slider_set_btn = ui.button("Set slider->80", slider_set_80)
var slider_row = ui.row([slider_label, slider_obj.widget, slider_set_btn])
println("slider initial value: #{slider_obj.get_value()}")

var cg = ui.check_group(["opt1", "opt2", "opt3", "opt4"], on_check_group)
var se = ui.select_entry(["apple", "banana", "cherry"], on_select_entry)

# ---------- hyperlink / separator / icon ----------

var hl_plain = ui.hyperlink("Fyne.io (opens browser)", "https://fyne.io")
var hl_handled = ui.hyperlink_with_handler("Custom tap", "https://example.com", on_hyperlink)
var sep = ui.separator()
var icon_lab = ui.icon_widget(ui.icon.home)
var icon_row = ui.row([ui.label("Icon:"), icon_lab, hl_plain, hl_handled])

# ---------- card / accordion ----------

var card_inner = ui.label("Card content - hello from inside a Card!")
var my_card = ui.card("Card Title", "Subtitle goes here", card_inner)

var acc_item1 = ui.accordion_item("Section 1", ui.label("Detail for section 1"))
var acc_item2 = ui.accordion_item("Section 2", ui.row([ui.label("Row in accordion"), ui.button("Btn", on_button)]))
var acc = ui.accordion([acc_item1, acc_item2])

# ---------- canvas primitives ----------

var red   = ui.new_color(220, 40, 40, 255)
var green = ui.new_color(40, 180, 40, 255)
var bluec  = ui.new_color(40, 80, 220, 255)
var dark  = ui.new_color(20, 20, 20, 255)

var rect = ui.canvas_rectangle(red)
var circ = ui.canvas_circle(green)
var line = ui.canvas_line(bluec)
var ctext = ui.canvas_text("Canvas Text", dark, 18)
var canvas_row = ui.grid(2, children=[rect, circ, line, ctext])
var canvas_demo = ui.padded(canvas_row)

# ---------- rich text (markdown) ----------

var rt = ui.rich_text("# RichText\n**Bold** *italic* `code`\n- item 1\n- item 2\n[Link](https://example.com)")

# ---------- next biggest: canvas extras + data widgets (NEW) ----------

var act = ui.activity()
fun toggle_act() {
    act.stop()
    println("[action] activity stopped")
}
var act_row = ui.row([ui.label("Activity:"), act.widget, ui.button("Stop activity", toggle_act), ui.button("Start", || => { act.start(); println("activity started") })])

var text_grid = ui.text_grid("TextGrid demo\nLine 2: monospace grid\nCol\tA\tB\tC\nRow1\t123\t456\t789\nRow2\tabc\tdef\tghi")

var file_icon_w = ui.file_icon("file:///tmp/example.png")
var file_icon_w2 = ui.file_icon("file:///tmp/photo.jpg")
var file_icon_row = ui.row([ui.label("FileIcons:"), file_icon_w, file_icon_w2, ui.label("file.png / photo.jpg")])

var inner_win = ui.inner_window("Inner Window", ui.col([ui.label("Draggable inner window"), ui.button("Inner Btn", on_button)]))
var inner_demo = ui.stack([ui.label("Behind inner"), inner_win])

# canvas extras - fixed width via padded containers (arc/poly/image each ~200px)
var arc_demo = ui.canvas_arc(ui.new_color(255, 128, 0, 255), 0, 270, 0.4)
var poly_demo = ui.canvas_polygon(6, ui.new_color(0, 128, 255, 255))
var img_demo = ui.canvas_image("manual_tests/test_img.png")
# give image a fixed width by wrapping in padded/center; row distributes equally
var canvas_extras_row = ui.grid(3, children=[arc_demo, poly_demo, img_demo])
var canvas_extras = ui.padded(canvas_extras_row)

# data widgets
var cal = ui.calendar("2026-09-06", on_calendar)
var lst = ui.list_items(["Item A", "Item B", "Item C", "Item D", "Item E", "Item F", "Long item to test scrolling", "Item H", "Item I", "Item J"], on_list)
var tbl = ui.table([["R1C1","R1C2","R1C3","R1C4"], ["R2C1","R2C2","R2C3","R2C4"], ["R3C1","R3C2","R3C3","R3C4"], ["R4C1","R4C2","R4C3","R4C4"]], on_table)

# ---------- next next biggest: gradients, grid_wrap, tree, dialogs, menu ----------
fun on_grid_wrap(i, v) {
    println("[handler] grid_wrap => #{i} #{v}")
}
fun on_tree(uid) {
    println("[handler] tree => #{uid}")
}
fun on_dialog_confirm(v) {
    println("[handler] dialog_confirm => #{v}")
}
fun on_file_open(p) {
    println("[handler] file_open => '#{p}'")
}

var grad_linear = ui.canvas_linear_gradient(ui.new_color(255,0,0,255), ui.new_color(0,0,255,255), 90)
var grad_radial = ui.canvas_radial_gradient(ui.new_color(255,255,0,255), ui.new_color(255,0,0,0))
var square_demo = ui.canvas_square(ui.new_color(128,0,128,255))
var grad_row = ui.grid(3, children=[grad_linear, grad_radial, square_demo])
var grad_demo = ui.padded(grad_row)

var gw = ui.grid_wrap(["GW A", "GW B", "GW C", "GW D", "GW E", "GW F", "GW G", "GW H", "GW I", "GW J", "GW K", "GW L"], on_grid_wrap)

var tree_data = {
    '': ['Projects', 'Docs'],
    'Projects': ['blue', 'app'],
    'Docs': ['README', 'Guide'],
    'blue': ['main.go', 'lib'],
    'app': ['window', 'canvas']
}
var tree_demo = ui.tree(tree_data, '', on_tree)

var btn_info = ui.button("Info Dialog", || => { ui.dialog_info("Info", "Hello from dialog_info!\nThis is a non-blocking info window with OK."); println("dialog_info shown") })
var btn_confirm = ui.button("Confirm Dialog", || => { ui.dialog_confirm("Confirm", "Are you sure you want to proceed?", on_dialog_confirm); println("dialog_confirm shown") })
var btn_file = ui.button("File Open Dialog", || => { ui.dialog_file_open(on_file_open); println("file_open shown") })
var dialog_row = ui.row([btn_info, btn_confirm, btn_file])

# menu demo - set main menu via button (requires window exists)
var m_file = ui.menu("File", [ui.menu_item("Info", || => { ui.dialog_info("Menu", "File->Info clicked"); println("menu File->Info") }), ui.menu_item("Quit", on_button)])
var m_help = ui.menu("Help", [ui.menu_item("About", || => { println("menu Help->About") })])
var btn_set_menu = ui.button("Set Main Menu (File, Help)", || => { ui.set_main_menu([m_file, m_help]); println("main menu set") })

# ---------- containers: scroll, split, border, center, padded, stack ----------

var long_col = ui.col([
    ui.label("Line 1 - scroll demo"),
    ui.label("Line 2"),
    ui.label("Line 3 - keep scrolling"),
    ui.label("Line 4"),
    ui.label("Line 5"),
    ui.label("Line 6"),
    ui.button("Inside scroll", on_button)
])
var scroll_demo = ui.scroll(long_col)
var hscroll_demo = ui.hscroll(ui.row([ui.label("H-Scroll ->"), ui.label("item 1"), ui.label("item 2"), ui.label("item 3"), ui.label("item 4"), ui.label("item 5"), ui.label("very wide content to force horizontal scrolling")]))
var vscroll_demo = ui.vscroll(ui.col([ui.label("V-Scroll down"), ui.label("a"), ui.label("b"), ui.label("c"), ui.label("d"), ui.label("e"), ui.label("f"), ui.label("g"), ui.label("h")]))

var left_pane = ui.padded(ui.label("Left pane (HSplit)"))
var right_pane = ui.padded(ui.label("Right pane"))
var hsplit_demo = ui.hsplit(left_pane, right_pane)

var top_pane = ui.label("Top (VSplit)")
var bottom_pane = ui.label("Bottom")
var vsplit_demo = ui.vsplit(top_pane, bottom_pane)

var border_demo = ui.border(
    ui.label("Top (border)"),
    ui.label("Bottom"),
    ui.label("Left"),
    ui.label("Right"),
    ui.center(ui.label("Center"))
)

var stacked = ui.stack([ui.canvas_rectangle(ui.new_color(255,0,0,80)), ui.center(ui.label("Stacked!"))])
var stack_demo = ui.padded(stacked)

# ---------- form (old) + toolbar ----------

var f_entry = ui.entry(placeholder="form field")
var f_ml = ui.entry(is_multiline=true, placeholder="form multiline")
var my_form = ui.form(children=[{'label': 'Field', 'elem': f_entry}, {'label': 'Multiline', 'elem': f_ml}], on_submit=on_submit)
my_form.append_form("Extra", ui.label("appended label"))

var tb = ui.toolbar.new(
    ui.toolbar.action(ui.icon.home, on_button),
    ui.toolbar.spacer(),
    ui.toolbar.action(ui.icon.settings, on_button),
    ui.toolbar.separator(),
    ui.toolbar.action(ui.icon.info, on_button)
)

# ---------- progress bars ----------

var prog_inf = ui.progress_bar(true)
var prog_row = ui.row([ui.label("Progress:"), prog.widget, prog_inf.widget, ui.button("Inc prog", || => { prog.set_value(0.5); println("prog set 0.5") })])

# =============================================================
#  Build tab structure - each tab exercises a family of widgets
# =============================================================

var tab_basics = ui.tab_item("Basics", ui.vscroll(ui.col([
    ui.label("== BASICS: label / button / entry / checkbox =="),
    ui.row([ui.label("Hello World!"), btn]),
    entry_row,
    ui.row([chk, rad]),
    ui.row([sel]),
    sep,
    icon_row,
    prog_row
])))

var tab_selectors = ui.tab_item_with_icon("Selectors", ui.icon.settings, ui.vscroll(ui.col([
    ui.label("== SELECTORS: radio / option_select / check_group / select_entry / slider =="),
    ui.label("Radio + OptionSelect (old):"),
    ui.row([rad, sel]),
    ui.label("CheckGroup (new, multi-select):"),
    cg,
    ui.label("SelectEntry (new, editable combo):"),
    se,
    ui.label("Slider (new, float callback):"),
    slider_row,
    ui.label("Hyperlinks:"),
    ui.row([hl_plain, hl_handled])
])))

var tab_containers = ui.tab_item("Containers", ui.vscroll(ui.col([
    ui.label("== CONTAINERS =="),
    ui.label("Row / Col / Grid (old):"),
    ui.row([ui.label("R1"), ui.label("R2"), ui.label("R3")]),
    ui.col([ui.label("C1"), ui.label("C2")]),
    ui.grid(2, children=[ui.label("G1"), ui.label("G2"), ui.label("G3"), ui.label("G4")]),
    sep,
    ui.label("Scroll variants:"),
    ui.row([scroll_demo, vscroll_demo]),
    hscroll_demo,
    sep,
    ui.label("Split:"),
    ui.row([hsplit_demo, vsplit_demo]),
    sep,
    ui.label("Border / Center / Padded / Stack:"),
    border_demo,
    sep,
    ui.label("Stack (overlapping):"),
    stack_demo
])))

var tab_advanced = ui.tab_item("Advanced", ui.vscroll(ui.col([
    ui.label("== ADVANCED: card / accordion / rich_text / canvas =="),
    my_card,
    sep,
    ui.label("Accordion:"),
    acc,
    sep,
    ui.label("RichText (markdown):"),
    rt,
    sep,
    ui.label("Canvas primitives + colors:"),
    canvas_demo,
    sep,
    ui.label("Toolbar:"),
    tb
])))

var tab_data = ui.tab_item_with_icon("Data", ui.icon.storage, ui.vscroll(ui.col([
    ui.label("== NEXT: calendar / list / table / textgrid / activity / inner / fileicon / canvas =="),
    ui.label("Calendar (select date):"),
    cal,
    sep,
    ui.label("List (virtualized, click row):"),
    lst,
    sep,
    ui.label("Table (click cell):"),
    tbl,
    sep,
    ui.label("TextGrid (monospace):"),
    text_grid,
    sep,
    ui.label("Activity indicator + FileIcons + InnerWindow:"),
    act_row,
    file_icon_row,
    inner_demo,
    sep,
    ui.label("Canvas extras: Arc / Polygon / Image:"),
    canvas_extras
])))

var tab_next = ui.tab_item_with_icon("Extra", ui.icon.help, ui.vscroll(ui.col([
    ui.label("== EXTRA: gradients / grid_wrap / tree / dialogs / menu =="),
    ui.label("Canvas gradients + square:"),
    grad_demo,
    sep,
    ui.label("GridWrap (click item):"),
    gw,
    sep,
    ui.label("Tree (hierarchical, click node):"),
    tree_demo,
    sep,
    ui.label("Dialogs (click to show):"),
    dialog_row,
    sep,
    ui.label("Main Menu (click to set File/Help menu):"),
    btn_set_menu,
    ui.label("(Menu appears at top of window after click)")
])))

var tab_form = ui.tab_item_with_icon("Form", ui.icon.document, ui.vscroll(ui.col([
    ui.label("== FORM + toolbar demo =="),
    ui.label("Form (with append_form):"),
    my_form._form,
    sep,
    ui.label("Toolbar with icons:"),
    tb
])))

# Inner doc_tabs demo (closable tabs) as a nested example
var inner_doc = ui.doc_tabs([
    ui.tab_item("Doc1", ui.label("Document tab 1 content")),
    ui.tab_item("Doc2", ui.label("Document tab 2 content")),
    ui.tab_item_with_icon("Doc3", ui.icon.file, ui.label("Doc3 with icon"))
])
var tab_docs = ui.tab_item("DocTabs", ui.vscroll(ui.col([
    ui.label("== DocTabs (closable, inner) =="),
    inner_doc
])))

# ---------- top-level tabs ----------

var main_tabs = ui.tabs([tab_basics, tab_selectors, tab_containers, tab_advanced, tab_data, tab_next, tab_form, tab_docs])

println("Launching window with 8 top-level tabs + toolbar + all new widgets...")
println("Try: click buttons, toggle checks, move slider, pick options, open accordion, pick calendar/list/table, grid_wrap/tree, dialogs/menu, switch tabs.")

ui.window(width=640, height=480, title="Blue - Comprehensive UI Test (expanded fyne coverage)", content=main_tabs)

println("Window closed - manual verification complete.")
