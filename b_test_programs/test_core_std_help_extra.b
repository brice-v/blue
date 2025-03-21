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
    Signature:  fetch(resource: str, method: str('POST'|'PUT'|'PATCH'|'GET'|'HEAD'|'DELETE')='GET', headers: map[str]str, body: null|str|bytes, full_resp: bool)
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


import http

val handle_h = help(http.handle);
println(handle_h);
val expected_handle_h = """`handle` takes a pattern, function, and method
and attaches itself to the _server http object

handler functions should return a string or bytes or null
they can take parameters which correspond to the pattern
string.

example: with a pattern string of '/hello/:a/:b'
the handler function should have a signature such as
fun(a, b) {} which allows the params to be used

example: with a 'POST' method the handler function
should have a signature such as fun(post_values=['a', 'b']) {}
where 'a' and 'b' are values received in the POST request

query_params operates in a similar fashion to the post_values
accepting a list of strings for the query_params passed to the
request

headers is also reserved in the function signature
to allow the user to retrieve the headers of the request
passed in

handle(pattern: str, fn: fun, method: str='GET') -> null
`handle` puts a handler on the server for a given pattern and method, `handle_use` also can use this function if no method is provided
    Signature:  handle(server: GoObj[*fiber.App], pattern: str, fn: fun, method: str='GET') -> null
    Error(s):   InvalidArgCount,PositionalType,CustomError
    Example(s): handle(s, '/', fn) => null

    type = 'BUILTIN'
    inspect = 'builtin function'
""".replace("\r", "");
assert(handle_h == expected_handle_h);
val handle_use_h = help(http.handle_use);
println(handle_use_h);
val expected_handle_use_h = """`handle_use` takes an optional pattern, and a function, and method
and attaches itself to the _server http object

example: with a pattern string of '/hello/:a/:b'
the handler function should have a signature such as
fun(a, b) {} which allows the params to be used

example: with a 'POST' method the handler function
should have a signature such as fun(post_values=['a', 'b']) {}
where 'a' and 'b' are values received in the POST request

query_params operates in a similar fashion to the post_values
accepting a list of strings for the query_params passed to the
request

headers is also reserved in the function signature
to allow the user to retrieve the headers of the request
passed in

handle(pattern: str, fn: fun, method: str='GET') -> null
`handle` puts a handler on the server for a given pattern and method, `handle_use` also can use this function if no method is provided
    Signature:  handle(server: GoObj[*fiber.App], pattern: str, fn: fun, method: str='GET') -> null
    Error(s):   InvalidArgCount,PositionalType,CustomError
    Example(s): handle(s, '/', fn) => null

    type = 'BUILTIN'
    inspect = 'builtin function'
""".replace("\r", "");
assert(handle_use_h == expected_handle_use_h);

import math

val hypot_h = help(math.hypot);
println(hypot_h);
val expected_hypot_h = """`hypot` returns sqrt(p*p + q*q), taking care to avoid unnecessary overflow and underflow
    Signature:  hypot(p: float, q: float) -> float
    Error(s):   InvalidArgCount,PositionalType
    Example(s): hypot(3.0,4.0) => 5.0

    type = 'BUILTIN'
    inspect = 'builtin function'""".replace("\r", "");
assert(hypot_h == expected_hypot_h);



import psutil
val cpu_percent_help = help(psutil.cpu.percent);
println(cpu_percent_help);
val cpu_percent_expected = """`cpu_usage_percent` returns a list of cpu usages as floats per core
    Signature:  cpu_usage_percent() -> list[float]
    Error(s):   InvalidArgCount,CustomError
    Example(s): cpu_usage_percent() => [1.0,0.4,0.2,0.6]

    type = 'BUILTIN'
    inspect = 'builtin function'""".replace("\r","");
assert(cpu_percent_help == cpu_percent_expected)
val host_temps_help = help(psutil.host.temps);
println(host_temps_help);
val host_temps_expected = """`host.temps()`: `psutil_host_temps_info_to_map` returns the mapped version of host_temps_info json

host.temps() -> list[map[str:any]]
`host_temps_info` returns a list of json strings of host sensor temperature info
    Signature:  host_temps_info() -> list[str]
    Error(s):   InvalidArgCount,CustomError
    Example(s): host_temps_info() => [json_with_keys('sensorKey','temperature','sensorHigh','sensorCritical')]

    type = 'BUILTIN'
    inspect = 'builtin function'
""".replace("\r","");
assert(host_temps_help == host_temps_expected);
val mem_virtual_help = help(psutil.mem.virtual);
println(mem_virtual_help);
val mem_virtual_expected = """`mem.virtual()`: `psutil_mem_info_to_map` returns the mapped version of mem_virt_info json

mem.virtual() -> map[str:int]
`mem_virt_info` returns a json string of virtual memory info
    Signature:  mem_virt_info() -> str
    Error(s):   InvalidArgCount,CustomError
    Example(s): mem_virt_info() => json_with_keys('total','available','used','usedPercent','free','active','inactive','wired','laundry','buffers','cached','writeBack','dirty','writeBackTmp','shared','slab','sreclaimable','sunreclaim','pageTables','swapCached','commitLimit','committedAS','highTotal','highFree','lowTotal','lowFree','swapTotal','swapFree','mapped','vmallocTotal','vmallocUsed','vmallocChunk','hugePagesTotal','hugePagesFree','hugePagesRsvd','hugePagesSurp','hugePageSize')

    type = 'BUILTIN'
    inspect = 'builtin function'
""".replace("\r", "");
assert(mem_virtual_help == mem_virtual_expected);