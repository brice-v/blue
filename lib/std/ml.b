## `ml` is the module that contains functions needed to
## train ml models

val __tensor = _tensor;
val __matmul = _matmul;
val __add = _add;
val __sub = _sub;
val __mul = _mul;
val __div = _div;
val __relu = _relu;
val __exp = _exp;
val __log = _log;
val __sqrt = _sqrt;
val __reshape = _reshape;
val __transpose = _transpose;

val dtype = {
    'float32': 'float32',
    'float64': 'float64',
    'int32': 'int32',
    'bool': 'bool',
};

val device = {
    'cpu': 'cpu',
    'gpu': 'gpu',
};

fun tensor(data, datatype=dtype.float32, dev=device.cpu, requires_grad=false) {
    ##std:this,__tensor
    ## TODO: provide docs for tensor
    __tensor(data, datatype, dev, requires_grad)
}

fun matmul(a, b) {
    ##std:this,__matmul
    ## TODO: provide docs for matmul
    __matmul(a, b)
}

fun add(a, b) {
    ##std:this,__add
    ## TODO: provide docs for add
    __add(a, b)
}

fun sub(a, b) {
    ##std:this,__sub
    ## TODO: provide docs for sub
    __sub(a, b)
}

fun mul(a, b) {
    ##std:this,__mul
    ## TODO: provide docs for mul
    __mul(a, b)
}

fun div(a, b) {
    ##std:this,__div
    ## TODO: provide docs for div
    __div(a, b)
}

fun relu(a) {
    ##std:this,__relu
    ## TODO: provide docs for relu
    __relu(a)
}

fun exp(a) {
    ##std:this,__exp
    ## TODO: provide docs for exp
    __exp(a)
}

fun log(a) {
    ##std:this,__log
    ## TODO: provide docs for log
    __log(a)
}

fun sqrt(a) {
    ##std:this,__sqrt
    ## TODO: provide docs for sqrt
    __sqrt(a)
}

fun reshape(a, shape) {
    ##std:this,__reshape
    ## TODO: provide docs for reshape
    __reshape(a, shape)
}

fun transpose(a, dim0=0, dim1=1) {
    ##std:this,__transpose
    ## TODO: provide docs for transpose
    __transpose(a, dim0, dim1)
}
