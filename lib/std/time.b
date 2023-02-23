## `time` is the module that contains time related functions


val __now = _now;
val __sleep = _sleep;

fun now() {
    ## `now` returns the current unix timestamp as an int
    ##
    ## now() -> int
    __now()
}

fun sleep(ms) {
    ## `sleep` will pause execution of the current process for the given number of milliseconds
    ##
    ## sleep(ms: int) -> null
    __sleep(ms)
}