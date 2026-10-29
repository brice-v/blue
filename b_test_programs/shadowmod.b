val tz = {'UTC': 'UTC'}
val dtype = {'float32': 'float32'}
val scale_factor = 3
var counter = 0

fun tensor(data, dtype=dtype.float32) {
    ## tensor uses the module level dtype as the default for its dtype parameter
    ##
    ## tensor(data: list, dtype: str='float32') -> list
    [data, dtype]
}

fun scale(x, scale_factor=scale_factor) {
    ## scale uses the module level scale_factor as the default for its parameter
    ##
    ## scale(x: int, scale_factor: int=3) -> int
    x * scale_factor
}

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
