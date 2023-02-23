## `search` is the module that contains search related functions
## its mostly used by core 'find_one' and 'find_all' functions


val __by_xpath = _by_xpath;
val __by_regex = _by_regex;

fun by_xpath(str_to_search, str_query, should_find_one) {
    ## `by_xpath` is the search method for xpaths
    ## this is only usable on an HTML document
    ## it will return a list of strings if should_find_one is false
    ##
    ## by_xpath(str_to_search: str, str_query: str, should_find_one: bool) -> str|list[str]
    __by_xpath(str_to_search, str_query, should_find_one)
}

fun by_regex(str_to_search, str_query, should_find_one) {
    ## `by_regex` is the search method for regular expressions
    ## the underlying engine is the go standard library
    ## it will return a list of strings if should_find_one is false
    ##
    ## by_regex(str_to_search: str, str_query: str, should_find_one: bool) -> str|list[str]
    __by_regex(str_to_search, str_query, should_find_one)
}