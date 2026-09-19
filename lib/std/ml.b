## `ml` is the module that contains functions needed to
## train ml models

val __tensor = _tensor;
val __matmul = _matmul;
val __add = _add;
val __sub = _sub;
val __mul = _mul;
val __div = _div;
val __pow = _pow;
val __neg = _neg;
val __relu = _relu;
val __exp = _exp;
val __log = _log;
val __sqrt = _sqrt;
val __reshape = _reshape;
val __transpose = _transpose;
val __eq = _eq;
val __ne = _ne;
val __gt = _gt;
val __ge = _ge;
val __lt = _lt;
val __le = _le;
val __equal = _equal;
val __allclose = _allclose;
val __sum = _sum;
val __mean = _mean;
val __max = _max;
val __min = _min;
val __argmax = _argmax;
val __argmin = _argmin;
val __softmax = _softmax;

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
    ## `tensor` builds a tensor from nested lists, inferring the shape from the nesting
    ##
    ## tensor(data: list, datatype: str='float32', dev: str='cpu', requires_grad: bool=false) -> tensor
    __tensor(data, datatype, dev, requires_grad)
}

fun matmul(a, b) {
    ##std:this,__matmul
    ## `matmul` returns the matrix product of two tensors
    ##
    ## matmul(a: tensor, b: tensor) -> tensor
    __matmul(a, b)
}

fun add(a, b) {
    ##std:this,__add
    ## `add` returns the elementwise sum of two tensors, with scalar broadcast
    ##
    ## add(a: tensor, b: tensor) -> tensor
    __add(a, b)
}

fun sub(a, b) {
    ##std:this,__sub
    ## `sub` returns the elementwise difference of two tensors, with scalar broadcast
    ##
    ## sub(a: tensor, b: tensor) -> tensor
    __sub(a, b)
}

fun mul(a, b) {
    ##std:this,__mul
    ## `mul` returns the elementwise product of two tensors, with scalar broadcast
    ##
    ## mul(a: tensor, b: tensor) -> tensor
    __mul(a, b)
}

fun div(a, b) {
    ##std:this,__div
    ## `div` returns the elementwise quotient of two tensors, with scalar broadcast
    ##
    ## div(a: tensor, b: tensor) -> tensor
    __div(a, b)
}

fun pow(a, b) {
    ##std:this,__pow
    ## `pow` returns the elementwise power of two tensors, with scalar broadcast.
    ## A negative base is allowed when the exponent is an integer.
    ##
    ## pow(a: tensor, b: tensor) -> tensor
    __pow(a, b)
}

fun neg(a) {
    ##std:this,__neg
    ## `neg` returns the negation of each element
    ##
    ## neg(a: tensor) -> tensor
    __neg(a)
}

fun relu(a) {
    ##std:this,__relu
    ## `relu` returns max(0, x) elementwise
    ##
    ## relu(a: tensor) -> tensor
    __relu(a)
}

fun exp(a) {
    ##std:this,__exp
    ## `exp` returns e to the power of each element
    ##
    ## exp(a: tensor) -> tensor
    __exp(a)
}

fun log(a) {
    ##std:this,__log
    ## `log` returns the natural logarithm of each element
    ##
    ## log(a: tensor) -> tensor
    __log(a)
}

fun sqrt(a) {
    ##std:this,__sqrt
    ## `sqrt` returns the square root of each element
    ##
    ## sqrt(a: tensor) -> tensor
    __sqrt(a)
}

fun reshape(a, shape) {
    ##std:this,__reshape
    ## `reshape` returns a view of the tensor with a new shape; the element count must match
    ##
    ## reshape(a: tensor, shape: list[int]) -> tensor
    __reshape(a, shape)
}

fun transpose(a, dim0=0, dim1=1) {
    ##std:this,__transpose
    ## `transpose` returns a view with dim0 and dim1 swapped (metadata only)
    ##
    ## transpose(a: tensor, dim0: int=0, dim1: int=1) -> tensor
    __transpose(a, dim0, dim1)
}

fun eq(a, b) {
    ##std:this,__eq
    ## `eq` returns a bool tensor that is true where a == b, with scalar broadcast.
    ## Comparisons are not differentiable, and `==` is the operator form.
    ##
    ## eq(a: tensor, b: tensor) -> tensor
    __eq(a, b)
}

fun ne(a, b) {
    ##std:this,__ne
    ## `ne` returns a bool tensor that is true where a != b, with scalar broadcast.
    ## Comparisons are not differentiable, and `!=` is the operator form.
    ##
    ## ne(a: tensor, b: tensor) -> tensor
    __ne(a, b)
}

fun gt(a, b) {
    ##std:this,__gt
    ## `gt` returns a bool tensor that is true where a > b, with scalar broadcast.
    ## Comparisons are not differentiable, and `>` is the operator form.
    ##
    ## gt(a: tensor, b: tensor) -> tensor
    __gt(a, b)
}

fun ge(a, b) {
    ##std:this,__ge
    ## `ge` returns a bool tensor that is true where a >= b, with scalar broadcast.
    ## Comparisons are not differentiable, and `>=` is the operator form.
    ##
    ## ge(a: tensor, b: tensor) -> tensor
    __ge(a, b)
}

fun lt(a, b) {
    ##std:this,__lt
    ## `lt` returns a bool tensor that is true where a < b, with scalar broadcast.
    ## Comparisons are not differentiable, and `<` is the operator form.
    ##
    ## lt(a: tensor, b: tensor) -> tensor
    __lt(a, b)
}

fun le(a, b) {
    ##std:this,__le
    ## `le` returns a bool tensor that is true where a <= b, with scalar broadcast.
    ## Comparisons are not differentiable, and `<=` is the operator form.
    ##
    ## le(a: tensor, b: tensor) -> tensor
    __le(a, b)
}

fun equal(a, b) {
    ##std:this,__equal
    ## `equal` returns true when two tensors have the same shape, dtype, and elements.
    ## Use this to compare tensors in tests, since `==` is the elementwise comparison
    ## operator and returns a bool tensor.
    ##
    ## equal(a: tensor, b: tensor) -> bool
    __equal(a, b)
}

fun allclose(a, b, rtol=1e-5, atol=1e-8) {
    ##std:this,__allclose
    ## `allclose` returns true when two tensors have the same shape and every element
    ## satisfies |a - b| <= atol + rtol*|b|. Use it for computed results such as
    ## softmax, exp, and sqrt.
    ##
    ## allclose(a: tensor, b: tensor, rtol: float=1e-5, atol: float=1e-8) -> bool
    __allclose(a, b, rtol, atol)
}

fun sum(a, dim=null, keepdim=false) {
    ##std:this,__sum
    ## `sum` returns the sum over `dim`, or over every element when `dim` is null.
    ## `dim` may be an int or a list of ints; `keepdim` keeps reduced dims as size 1.
    ##
    ## sum(a: tensor, dim: int|list[int]|null=null, keepdim: bool=false) -> tensor
    __sum(a, dim, keepdim)
}

fun mean(a, dim=null, keepdim=false) {
    ##std:this,__mean
    ## `mean` returns the average over `dim`, or over every element when `dim` is null.
    ##
    ## mean(a: tensor, dim: int|list[int]|null=null, keepdim: bool=false) -> tensor
    __mean(a, dim, keepdim)
}

fun max(a, dim=null, keepdim=false) {
    ##std:this,__max
    ## `max` returns the maximum over `dim`, or over every element when `dim` is null.
    ##
    ## max(a: tensor, dim: int|list[int]|null=null, keepdim: bool=false) -> tensor
    __max(a, dim, keepdim)
}

fun min(a, dim=null, keepdim=false) {
    ##std:this,__min
    ## `min` returns the minimum over `dim`, or over every element when `dim` is null.
    ##
    ## min(a: tensor, dim: int|list[int]|null=null, keepdim: bool=false) -> tensor
    __min(a, dim, keepdim)
}

fun argmax(a, dim) {
    ##std:this,__argmax
    ## `argmax` returns the index of the maximum along `dim`.
    ##
    ## argmax(a: tensor, dim: int) -> tensor
    __argmax(a, dim)
}

fun argmin(a, dim) {
    ##std:this,__argmin
    ## `argmin` returns the index of the minimum along `dim`.
    ##
    ## argmin(a: tensor, dim: int) -> tensor
    __argmin(a, dim)
}

fun softmax(a, dim) {
    ##std:this,__softmax
    ## `softmax` returns a numerically stable softmax along `dim`.
    ##
    ## softmax(a: tensor, dim: int) -> tensor
    __softmax(a, dim)
}