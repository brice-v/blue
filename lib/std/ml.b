## `ml` is the module that contains functions needed to
## train ml models

val __tensor = _tensor;

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