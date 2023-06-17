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

var col_ui_content = ui.grid(3, children=[
    ui.label("Hello World!"), ui.label("Should be 2"), ui.label("column 2"), ui.label("under column 2?"),
    ui.button("Click Me!", button_handler), entry,
    ui.button("Button 2!", button2_handler), entry2,
    ui.checkbox("Optional", checkbox_handler), ui.radio_group(["1", "2"], radio_handler), ui.option_select(["a", "b"], option_handler)
]);

var row_ui_content = ui.grid(3, type=ui.GridType.ROWS, children=[
    ui.label("Hello World!"), ui.label("Should be 2"), ui.label("column 2"), ui.label("under column 2?"),
    ui.button("Click Me!", button_handler), entry,
    ui.button("Button 2!", button2_handler), entry2,
    ui.checkbox("Optional", checkbox_handler), ui.radio_group(["1", "2"], radio_handler), ui.option_select(["a", "b"], option_handler)
]);

println("row_ui_content = #{row_ui_content}");
println("col_ui_content = #{col_ui_content}");

var should_use_cols = input("Should use cols? ");
println("should_use_cols = #{should_use_cols}")
if (should_use_cols == "") {
    ui.window(width=1000, height=800, title="blue ui demo", content=col_ui_content);
} else {
    ui.window(width=1000, height=800, title="blue ui demo", content=row_ui_content);
}
