val tz1 = {'UTC': 'UTC'}
var counter1 = 0

fun fmt1(unix_ts, tz1=null) {
    ## fmt1 will use the passed tz or fall back to UTC when it is null
    ##
    ## fmt1(unix_ts: int, tz1: null|str=null) -> str
    if tz1 == null {
        tz1 = 'UTC'
    }
    'ts=' + str(unix_ts) + ' tz=' + tz1
}

fun bump1(by=1) {
    ## bump1 increments the module level counter by the given amount
    ##
    ## bump1(by: int=1) -> int
    counter1 += by
    counter1
}
