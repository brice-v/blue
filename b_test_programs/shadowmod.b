val tz = {'UTC': 'UTC'}
var counter = 0

fun fmt(unix_ts, tz=null) {
    ## fmt will use the passed tz or fall back to UTC when it is null
    ##
    ## fmt(unix_ts: int, tz: null|str=null) -> str
    if tz == null {
        tz = 'UTC'
    }
    'ts=' + str(unix_ts) + ' tz=' + tz
}

fun idx(tz=null) {
    ## idx will use the passed list or fall back to a default list when it is null
    ##
    ## idx(tz: null|list=null) -> list
    if tz == null {
        tz = [1, 2, 3]
    }
    tz[0] = 99
    tz
}

fun bump(by=1) {
    ## bump increments the module level counter by the given amount
    ##
    ## bump(by: int=1) -> int
    counter += by
    counter
}
