## `math` is the module that deals with most math related
## functions and constants
##
## e, pi, phi, sqrt2, sqrt_e, sqrt_pi, sqrt_phi, ln2, log2e,
## ln10, log10e are all stored as constants currently
##
## Note: all math functions implemented in go only really use floats


val e = 2.71828182845904523536028747135266249775724709369995957496696763; # https://oeis.org/A001113
val pi = 3.14159265358979323846264338327950288419716939937510582097494459; # https://oeis.org/A000796
val phi = 1.61803398874989484820458683436563811772030917980576286213544862; # https://oeis.org/A001622

val sqrt2 = 1.41421356237309504880168872420969807856967187537694807317667974; # https://oeis.org/A002193
val sqrt_e = 1.64872127070012814684865078781416357165377610071014801157507931; # https://oeis.org/A019774
val sqrt_pi = 1.77245385090551602729816748334114518279754945612238712821380779; # https://oeis.org/A002161
val sqrt_phi = 1.27201964951406896425242246173749149171560804184009624861664038; # https://oeis.org/A139339

val ln2 = 0.693147180559945309417232121458176568075500134360255254120680009; # https://oeis.org/A002162
val log2e = 1 / ln2;
val ln10 = 2.30258509299404568401799145468436420760110148862877297603332790; # https://oeis.org/A002392
val log10e = 1 / ln10;

val __rand = _rand;

val NaN = _NaN();

fun max(x, y) {
    ## `max` will return the max of the 2 numbers passed in
    ##
    ## max(x: num, y: num) -> num
    if (x > y) {
        x
    } else {
        y
    }
}

fun min(x, y) {
    ## `min` will return the min of the 2 numbers passed in
    ##
    ## min(x: num, y: num) -> num
    if (x < y) {
        x
    } else {
        y
    }
}

fun abs(x) {
    ## `abs` will return the absolute value of the number passed in
    ##
    ## abs(x: num) -> num
    if (x < 0) {
        x * -1
    } else {
        x
    }
}

fun sqrt(x) {
    ## `sqrt` will return the square root of the number passed in
    ##
    ## sqrt(x: num) -> num
    x ** 0.5
}

fun sum(x) {
    ## `sum` will add up all the numbers of the list passed in
    ##
    ## sum(x: list[num]) -> num
    var _sum = 0;
    for (i in x) {
        _sum += i;
    }
    return _sum;
}

fun rand() {
    ## `rand` returns a random float between 0 and 1
    ##
    ## rand() -> float
    __rand()
}