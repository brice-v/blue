import ui

var entry = ui.entry();
var entry_ml = ui.entry(is_multiline=true);

fun submit_handler() {
    println("entry text = `#{entry.get_text()}`");
    println("entry_ml text = `#{entry_ml.get_text()}`");
}

var f = ui.form(children=[{'label': 'Entry label', 'elem': entry}], on_submit=submit_handler);

ui.append_form(f, "ML Label", entry_ml);

ui.window(width=1000, height=800, title="blue ui form demo", content=f);