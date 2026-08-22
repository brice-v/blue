## Reads every line from STDIN and echoes it back with a line number,
## demonstrating filter style scripts that consume piped input
var count = 1
for true {
    val line = input()
    if line == null {
        break
    }
    println("#{count}: #{line}")
    count += 1
}
println("read #{count - 1} line(s)")
