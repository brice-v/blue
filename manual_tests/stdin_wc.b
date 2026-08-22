## Slurps ALL of stdin in one call using STDIN.read() and prints simple
## counts (a mini wc). Contrast with stdin_filter.b which consumes stdin
## line by line with input()
##
## Why STDIN and not FSTDIN: STDIN holds the path '/dev/stdin' so
## STDIN.read() goes through the same file read builtin used for any path.
## FSTDIN is a wrapped Go *os.File handle which the read builtin does not
## accept (it expects a STRING path), it exists for subsystems that take
## real handles such as wasm module init
##
## Note: stdin can only be consumed once per run so this script reads a
## single time and works from that buffer
val data = STDIN.read(false)

var word_count = 0
for word in data.replace("\n", " ").replace("\t", " ").split(" ") {
    if word != "" {
        word_count += 1
    }
}

var line_count = 0
for line in data.split("\n") {
    if line != "" {
        line_count += 1
    }
}

println("words: #{word_count}")
println("lines: #{line_count}")
println("chars: #{data.len()}")
