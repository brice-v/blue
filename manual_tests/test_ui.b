import ui

var entry = ui.entry();
var entry2 = ui.entry(placeholder="Test placeholder");

fun button_handler() {
    println(entry.get_text());
}

fun button2_handler() {
    var text = entry2.get_text();
    entry.set_text(text);
    println("Setting entry 1's text to '#{entry2.get_text()}'");
}

fun checkbox_handler(v) {
    println("checkbox_handler `#{v}`");
}

fun radio_handler(v) {
    println("radio_handler `#{v}`");
}

fun option_handler(v) {
    println("option_handler `#{v}`");
}

# Note these all had to be the same # of elements in order to get the button to show up
var ui_content = ui.col([
    ui.row(children=[ui.label("Hello World!"), ui.label("Should be 2")]),
    ui.row(children=[ui.label("column 2"), ui.label("under column 2?")]),
    ui.row([ui.button("Click Me!", button_handler), entry]),
    ui.row([ui.button("Button 2!", button2_handler), entry2]),
    ui.row([ui.checkbox("Optional", checkbox_handler), ui.radio_group(["1", "2"], radio_handler), ui.option_select(["a", "b"], option_handler)])
]);

println("ui_content = #{ui_content}");

ui.window(width=1000, height=800, title="blue ui demo", content=ui_content);
