## csv will allow the user to parse a csv from a string and also write rows
## to a string

val __parse = _parse;
val __dump = _dump;

fun parse(data, delimeter=',', named_fields=false, comment=null, lazy_quotes=false, trim_leading_space=false) {
    ## `parse` will take data as a string and return a list of rows represented by
    ## the csv file
    ##
    ## if named_fields is true the first row will be used as a header to label each
    ## of the following rows (and it will be skipped)
    ##
    ## delimeter is defaulted to ',' but can be set to any valid rune
    ##
    ## comment is defaulted to null implying no comments are present but can be
    ## a valid rune used for comments in the csv
    ##
    ## lazy_quotes is defaulted to false but if true a quote may appear in an unquoted
    ## field and a non-doubled quote may appear in a quoted field
    ##
    ## trim_leading_space is defaulted to false but if true leading white space
    ## in a field is ignored - this is done even if the field delimiter is white space
    ##
    ##
    ## parse(data: str, delimeter: str=',', named_fields: bool=false, comment: str=null,
    ##       lazy_quotes: bool=false, trim_leading_space: bool=false)
    ##       -> list[list[str]] or list[map[str:str]]
    __parse(data, delimeter, named_fields, comment, lazy_quotes, trim_leading_space)
}

fun dump(data, comma=',', use_crlf=false) {
    ## `dump` will take the data and dump it to a string formatted as a csv
    ## the data must be in the form list[list[str]] or list[map[any:any]]
    ##
    ## the reason the list of maps can use any to any is due to the fact that
    ## inspect will be called for all the items - for strings this will pretty
    ## much leave it all as is
    ## the keys must all match in all the rows, this will be written as the header
    ## for the csv
    ## for other data not originally in a csv, this makes it easier to put it
    ## all into a csv format
    ##
    ## invalid data will return an error
    ##
    ## the string returned should be written to a file to save it
    ##
    ## dump(data: list[list[any]]|list[map[any:any]]) -> str
    __dump(data, comma, use_crlf)
}