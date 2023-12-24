val x = help(sort);
println(x);
val xx = """`sort` sorts the given list, if its ints, floats, or strings no custom key is needed, otherwise a function returning the key to sort should be returned (ie. a str, float, or int)
    Signature:  sort(l: list[int|float|str|any], reverse: bool=false, key: null|fun(e: list[any])=>int|str|float=null) -> list[int|float|str|any] (sorted)
    Error(s):   InvalidArgCount,PositionalType,CustomError
    Example(s): sort(['c','b','a']) => ['a','b','c']

    type = 'BUILTIN'
    inspect = 'builtin function'""".replace("\r", "");
assert(x == xx);

val y = help(fetch);
println(y);
val yy = """`fetch` allows the user to send GET, POST, PUT, PATCH, and DELETE
http methods to a various resource

there are other specific methods that populate these
options appropriately. user-agent in header is always
set to one specific to blue.

example option to send get request:
{method: 'GET', headers: {}, body: null}

example option to send post request:
{method: 'POST', body: str, headers: {'content-type': mime_type}}

fetch(resource: str, options: map[any:str]=null, full_resp: bool=true) -> any
`fetch` returns the body or full response of a network request
    Signature:  fetch(resource: str, method: str('POST'|'PUT'|'GET'|'HEAD'|'DELETE')='GET', headers: map[str]str, body: null|str|bytes, full_resp: bool)
    Error(s):   InvalidArgCount,PositionalType,CustomError
    Example(s): fetch('https://danluu.com',full_resp=false) => <html>...</html>

    type = 'BUILTIN'
    inspect = 'builtin function'
""".replace("\r", "");
assert(y == yy);


val z = help(replace);
println(z);
val zz = """`replace` will take the string, and replace the replacer with replaced

replacer can be a string or regex
replaced should be a string

is_regex can be used to convert a replacer string to a regex object

replace(str_to_replace: str, replacer: str|regex, replaced: str, is_regex: bool=false) -> str
`replace` will return a STRING with all occurrences of the given replacer STRING replaced by the next given STRING
    Signature:  replace(arg: str, replacer: str, replaced: str) -> str
    Error(s):   InvalidArgCount,PositionalType
    Example(s): replace('Hello', 'l', 'X') => 'HeXXo'

    type = 'BUILTIN'
    inspect = 'builtin function'
`replace_regex` will return a STRING with all occurrences of the given replacer REGEX STRING replaced by the next given STRING
    Signature:  replace_regex(arg: str, replacer: str, replaced: str) -> str
    Error(s):   InvalidArgCount,PositionalType
    Example(s): replace_regex('Hello', 'l', 'X') => 'HeXXo'

    type = 'BUILTIN'
    inspect = 'builtin function'
""".replace("\r", "");
assert(z == zz);