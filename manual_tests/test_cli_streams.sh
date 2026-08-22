#!/bin/bash
# Manual test that builds the blue binary and exercises the CLI
# STDIN/STDOUT/STDERR behavior end to end.
#
# Usage:
#     ./manual_tests/test_cli_streams.sh
#
# Exits 0 when every check passes, 1 otherwise.

set -u
cd "$(dirname "$0")/.." || exit 1

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT
BIN="$TMP_DIR/blue"

echo "== building blue to $BIN"
if ! go build -o "$BIN" .; then
    echo "FAIL: go build"
    exit 1
fi

PASS=0
FAIL=0

pass() { PASS=$((PASS + 1)); echo "PASS: $1"; }
fail() {
    FAIL=$((FAIL + 1))
    echo "FAIL: $1"
    [ $# -gt 1 ] && printf '      detail: %s\n' "$2"
}

# check_eq NAME EXPECTED ACTUAL
check_eq() {
    if [ "$2" = "$3" ]; then pass "$1"; else fail "$1" "expected [$2] got [$3]"; fi
}

# check_contains NAME SUBSTRING ACTUAL
check_contains() {
    case "$3" in
        *"$2"*) pass "$1" ;;
        *) fail "$1" "expected output containing [$2], got [$3]" ;;
    esac
}

echo "== stdin program execution"
check_eq "piped stdin with no command runs the program" \
    "3" "$(echo 'println(1 + 2)' | "$BIN")"

check_eq "'-' reads the program from stdin" \
    "6" "$(printf 'val x = 2\nprintln(x * 3)' | "$BIN" -)"

check_eq "'blue vm -' reads the program from stdin" \
    "v
null" "$(printf 'println("v")' | "$BIN" vm -)"

printf 'println("redirect works")\n' > "$TMP_DIR/prog.b"
check_eq "'blue < file' runs the redirected program" \
    "redirect works" "$("$BIN" < "$TMP_DIR/prog.b")"

check_contains "'blue compile -' emits bytecode for piped source" \
    "OpAdd" "$(echo '1 + 2' | "$BIN" compile -)"

check_eq "exit code is 0 on success" \
    "0" "$(echo 'println(1)' | "$BIN" > /dev/null 2>&1; echo $?)"

check_eq "exit code is 1 on compile error" \
    "1" "$(printf 'x = \n' | "$BIN" > /dev/null 2>&1; echo $?)"

check_eq "exit code is 1 for a missing file" \
    "1" "$("$BIN" "$TMP_DIR/does_not_exist.b" > /dev/null 2>&1; echo $?)"

echo "== stderr and stdout separation"
STDOUT_OUT="$(printf 'x = \n' | "$BIN" 2> /dev/null)"
if [ -z "$STDOUT_OUT" ]; then pass "parse errors do not pollute stdout"; else fail "parse errors do not pollute stdout" "stdout was [$STDOUT_OUT]"; fi

STDERR_OUT="$(printf 'val x = 1\nx[0]' | "$BIN" 2>&1 > /dev/null)"
check_contains "runtime errors are written to stderr" "VMError" "$STDERR_OUT"

PROGRAM_STDOUT="$(printf 'println("only here")\nval x = 1\nx[0]' | "$BIN" 2> /dev/null)"
check_eq "program output still lands on stdout when it later errors" \
    "only here" "$PROGRAM_STDOUT"

echo "== stdin from inside blue programs"
FILTER_OUT="$(printf 'b\na\n' | "$BIN" manual_tests/stdin_filter.b)"
check_eq "input() reads consecutive lines without loss (filter script)" \
    "1: b
2: a
read 2 line(s)" "$FILTER_OUT"

check_eq "input() returns null at eof so loops can terminate cleanly" \
    "read 0 line(s)" "$(printf '' | "$BIN" manual_tests/stdin_filter.b)"

check_eq "STDIN.read() slurp script computes counts" \
    "words: 4
lines: 2
chars: 23" "$(printf 'hello blue\nstdin world\n' | "$BIN" manual_tests/stdin_wc.b)"

check_eq "STDIN.read() slurp script handles empty stdin" \
    "words: 0
lines: 0
chars: 0" "$(printf '' | "$BIN" manual_tests/stdin_wc.b)"

check_eq "STDIN.read() slurp script handles tabs and runs of spaces" \
    "words: 4
lines: 2
chars: 9" "$(printf 'a\tb  c\nd\n' | "$BIN" manual_tests/stdin_wc.b)"

check_eq "STDIN.read() slurps all of stdin as a string" \
    "a
b" "$(printf 'a\nb\n' | "$BIN" vm 'println(STDIN.read(false))' | head -2)"

BYTE_OUT="$(printf 'abc' | "$BIN" vm 'println(STDIN.read(true))')"
check_contains "STDIN.read(true) returns bytes" "0x61" "$BYTE_OUT"

WRITE_FILE="$TMP_DIR/write_out.txt"
printf 'STDOUT.write("hello")\nSTDOUT.write(" world\\n")\n' > "$TMP_DIR/wr.b"
"$BIN" "$TMP_DIR/wr.b" > "$WRITE_FILE" 2> /dev/null
check_eq "STDOUT.write lands on stdout" "hello world" "$(cat "$WRITE_FILE")"

echo
echo "results: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ]
