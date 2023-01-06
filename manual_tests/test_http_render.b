import http


#val html_resp = http.render('example_html_tmpl.html', {'title': 'My title Here!'});

#println(html_resp);

# Note: When rendered this will have escaped '<' signs
val top_html = """
	<h1>Hello World!</h1>
""";

val file_content = read('example_md.md');
val html_content = http.md_to_html(file_content);
#println("html_content = #{html_content}");

val html_stylesheet = """<html>
<head>
<title>Hello World!</title>
<link rel="stylesheet" href="https://cdn.simplecss.org/simple.min.css">
</head>""";
var c = html_stylesheet+top_html+html_content+"</html>";
val small_content = http.sanitize_and_minify(c);
#println("small_content = #{small_content}");


val html_file_content = read('example_html.html');
val list_thing = ['some', 'things', 'here', 'to', 'list'];
val map_thing = {'1': 1234, 'list': [1,3,4,true]};
val rendered_template = http.sanitize_and_minify(eval_template(html_file_content+"</html>", {'top_html': top_html, 'thing': 12345, 'list_thing': list_thing, 'map_thing': map_thing}), should_sanitize=false);
println("rendered_template = #{rendered_template}");
exit(1);
# Note: sanitize_and_minify gets rid of the <head> stuff
val c1 = http.sanitize_and_minify(rendered_template, should_sanitize=false);
#println("c1 = #{c1}");

http.handle("/md", fun() { return c; });
http.handle("/", fun() { return c1; });
http.handle("/html", fun() { return rendered_template; });

http.serve();
