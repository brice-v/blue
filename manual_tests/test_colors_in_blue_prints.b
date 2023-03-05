import color

var s = color.style(text=color.bold, fg_color=color.red, bg_color=color.white);

println(s, "Hello World");

s = color.style(text=color.underlined, fg_color=color.cyan, bg_color=color.white)
println(s, "Some other string")

s = color.style(text=color.normal, fg_color=color.magenta, bg_color=color.normal)
println(s, "With default styling")


s = color.style(color.italic, color.magenta, color.green)
println(s, "With weird styling")

println("No Styling")

assert(true);