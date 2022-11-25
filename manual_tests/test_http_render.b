import http


#val html_resp = http.render('example_html_tmpl.html', {'title': 'My title Here!'});

#println(html_resp);

val style_sheet_html_snippet = """
	<html>
<head>
<title>Hello World!</title>
<link rel="stylesheet" href="https://cdn.simplecss.org/simple.min.css">
</head>
""";

val file_content = read('example_md.md');
val html_content = http.md_to_html(file_content);
#println("html_content = #{html_content}");

var c = style_sheet_html_snippet+html_content+"</html>";
val small_content = http.sanitize_and_minify(c);
#println("small_content = #{small_content}");


val html_file_content = read('example_html.html');
val rendered_template = eval_template(html_file_content+"</html>", {'stylesheet': style_sheet_html_snippet});
#println("rendered_template = #{rendered_template}");
# Note: sanitize_and_minify gets rid of the <head> stuff
val c1 = http.sanitize_and_minify(rendered_template, should_sanitize=false);
#println("c1 = #{c1}");

http.handle("/md", fun() { return c; });
http.handle("/", fun() { return c1; });
http.handle("/html", fun() { return rendered_template; });

http.serve();
