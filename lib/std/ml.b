## `ml` is the module that contains functions needed to
## train ml models

val __tensor = _tensor;
val __binary = _binary;
val __compare = _compare;
val __unary = _unary;
val __reduce = _reduce;
val __argreduce = _argreduce;
val __reshape = _reshape;
val __transpose = _transpose;
val __equal = _equal;
val __allclose = _allclose;
val __softmax = _softmax;

## The named helpers below are thin wrappers over the generic dispatchers, so
## there is one implementation per family and the public functions stay named.
fun __matmul(a, b) { __binary('matmul', a, b); }
fun __add(a, b) { __binary('add', a, b); }
fun __sub(a, b) { __binary('sub', a, b); }
fun __mul(a, b) { __binary('mul', a, b); }
fun __div(a, b) { __binary('div', a, b); }
fun __pow(a, b) { __binary('pow', a, b); }
fun __neg(a) { __unary('neg', a); }
fun __relu(a) { __unary('relu', a); }
fun __exp(a) { __unary('exp', a); }
fun __log(a) { __unary('log', a); }
fun __sqrt(a) { __unary('sqrt', a); }
fun __abs(a) { __unary('abs', a); }
fun __sigmoid(a) { __unary('sigmoid', a); }
fun __tanh(a) { __unary('tanh', a); }
fun __eq(a, b) { __compare('eq', a, b); }
fun __ne(a, b) { __compare('ne', a, b); }
fun __gt(a, b) { __compare('gt', a, b); }
fun __ge(a, b) { __compare('ge', a, b); }
fun __lt(a, b) { __compare('lt', a, b); }
fun __le(a, b) { __compare('le', a, b); }
fun __sum(a, dim, keepdim) { __reduce('sum', a, dim, keepdim); }
fun __mean(a, dim, keepdim) { __reduce('mean', a, dim, keepdim); }
fun __max(a, dim, keepdim) { __reduce('max', a, dim, keepdim); }
fun __min(a, dim, keepdim) { __reduce('min', a, dim, keepdim); }
fun __argmax(a, dim) { __argreduce('argmax', a, dim); }
fun __argmin(a, dim) { __argreduce('argmin', a, dim); }
val __flatten = _flatten;
val __unsqueeze = _unsqueeze;
val __permute = _permute;
val __broadcast_to = _broadcast_to;
val __squeeze = _squeeze;
val __zeros = _zeros;
val __ones = _ones;
val __randn = _randn;
val __full = _full;
val __arange = _arange;
val __eye = _eye;
val __manual_seed = _manual_seed;
val __slice = _slice;
val __clamp = _clamp;
val __where = _where;
val __cat = _cat;
val __gather = _gather;
val __onehot = _onehot;
val __set_requires_grad = _set_requires_grad;
val __requires_grad = _requires_grad;
val __backward = _backward;
val __grad = _grad;
val __zero_grad = _zero_grad;
val __set_grad_enabled = _set_grad_enabled;
val __cross_entropy = _cross_entropy;
val __cross_entropy_grad = _cross_entropy_grad;
val __mse_loss = _mse_loss;
val __add_ = _add_;
val __sub_ = _sub_;
val __mul_ = _mul_;
val __div_ = _div_;
val __save = _save;
val __load = _load;
val __optim_sgd = _optim_sgd;
val __optim_adam = _optim_adam;
val __optim_step = _optim_step;
val __optim_zero_grad = _optim_zero_grad;
val __nn_linear = _nn_linear;
val __nn_relu = _nn_relu;
val __nn_sigmoid = _nn_sigmoid;
val __nn_forward = _nn_forward;
val __nn_parameters = _nn_parameters;
val __nn_linear_act = _nn_linear_act;
val __nn_to = _nn_to;
val __gpu_is_available = _gpu_is_available;
val __cast = _cast;
val __trace_begin = _trace_begin;
val __trace_input = _trace_input;
val __trace_end = _trace_end;
val __graph_optimize = _graph_optimize;
val __compiled_new = _compiled_new;
val __compiled_has_plan = _compiled_has_plan;
val __compiled_install = _compiled_install;
val __compiled_run = _compiled_run;
val __compiled_stats = _compiled_stats;
val __compiled_backward = _compiled_backward;
val __embedding = _embedding;
fun __bmm(a, b) { __binary('bmm', a, b); }
fun __silu(a) { __unary('silu', a); }
val __rand = _rand;
val __randperm = _randperm;
val __shuffle = _shuffle;
val __select = _select;
val __index_select = _index_select;
val __masked_fill = _masked_fill;
val __optim_set_lr = _optim_set_lr;
val __optim_clip_grad_norm = _optim_clip_grad_norm;
val __lr_schedule = _lr_schedule;
val __clone = _clone;
val __state_dict = _state_dict;
val __save_state = _save_state;
val __load_state = _load_state;
val __flip = _flip;
val __masked_select = _masked_select;

val dtype = {
    'float32': 'float32',
    'float64': 'float64',
    'int32': 'int32',
    'int64': 'int64',
    'uint8': 'uint8',
    'bool': 'bool',
};

val device = {
    'cpu': 'cpu',
    'gpu': 'gpu',
};

## `gpu.is_available()` mirrors torch.cuda.is_available().
val gpu = {
    'is_available': fun() { __gpu_is_available(); },
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

fun abs(a) {
    ##std:this,__abs
    ## `abs` returns the absolute value of each element.
    ##
    ## abs(a: tensor) -> tensor
    __abs(a)
}

fun sigmoid(a) {
    ##std:this,__sigmoid
    ## `sigmoid` returns 1/(1+exp(-x)) elementwise.
    ##
    ## sigmoid(a: tensor) -> tensor
    __sigmoid(a)
}

fun tanh(a) {
    ##std:this,__tanh
    ## `tanh` returns the hyperbolic tangent elementwise.
    ##
    ## tanh(a: tensor) -> tensor
    __tanh(a)
}

fun flatten(a) {
    ##std:this,__flatten
    ## `flatten` returns a 1d view of the tensor.
    ##
    ## flatten(a: tensor) -> tensor
    __flatten(a)
}

fun unsqueeze(a, dim) {
    ##std:this,__unsqueeze
    ## `unsqueeze` inserts a size-1 dimension at `dim`.
    ##
    ## unsqueeze(a: tensor, dim: int) -> tensor
    __unsqueeze(a, dim)
}

fun permute(a, dims) {
    ##std:this,__permute
    ## `permute` reorders the dimensions.
    ##
    ## permute(a: tensor, dims: list[int]) -> tensor
    __permute(a, dims)
}

fun broadcast_to(a, shape) {
    ##std:this,__broadcast_to
    ## `broadcast_to` expands size-1 dims to the given shape.
    ##
    ## broadcast_to(a: tensor, shape: list[int]) -> tensor
    __broadcast_to(a, shape)
}

fun squeeze(a, dim=null) {
    ##std:this,__squeeze
    ## `squeeze` removes size-1 dimensions; `dim` null removes every size-1 dim.
    ##
    ## squeeze(a: tensor, dim: int|list[int]|null=null) -> tensor
    __squeeze(a, dim)
}

fun zeros(shape, datatype=dtype.float32, dev=device.cpu, requires_grad=false) {
    ##std:this,__zeros
    ## `zeros` returns a tensor of zeros with the given shape.
    ##
    ## zeros(shape: list[int], datatype: str='float32', dev: str='cpu', requires_grad: bool=false) -> tensor
    __zeros(shape, datatype, dev, requires_grad)
}

fun ones(shape, datatype=dtype.float32, dev=device.cpu, requires_grad=false) {
    ##std:this,__ones
    ## `ones` returns a tensor of ones with the given shape.
    ##
    ## ones(shape: list[int], datatype: str='float32', dev: str='cpu', requires_grad: bool=false) -> tensor
    __ones(shape, datatype, dev, requires_grad)
}

fun randn(shape, datatype=dtype.float32, dev=device.cpu, requires_grad=false) {
    ##std:this,__randn
    ## `randn` returns a tensor of standard normal samples.
    ##
    ## randn(shape: list[int], datatype: str='float32', dev: str='cpu', requires_grad: bool=false) -> tensor
    __randn(shape, datatype, dev, requires_grad)
}

fun full(shape, fill_value, datatype=dtype.float32, dev=device.cpu, requires_grad=false) {
    ##std:this,__full
    ## `full` returns a tensor filled with a value.
    ##
    ## full(shape: list[int], fill_value: float, datatype: str='float32', dev: str='cpu', requires_grad: bool=false) -> tensor
    __full(shape, fill_value, datatype, dev, requires_grad)
}

fun arange(start, end, step=1.0, datatype=dtype.float32, dev=device.cpu) {
    ##std:this,__arange
    ## `arange` returns evenly spaced values in [start, end).
    ##
    ## arange(start: float, end: float, step: float=1.0, datatype: str='float32', dev: str='cpu') -> tensor
    __arange(start, end, step, datatype, dev)
}

fun eye(n, datatype=dtype.float32, dev=device.cpu) {
    ##std:this,__eye
    ## `eye` returns an n by n identity matrix.
    ##
    ## eye(n: int, datatype: str='float32', dev: str='cpu') -> tensor
    __eye(n, datatype, dev)
}

fun manual_seed(seed) {
    ##std:this,__manual_seed
    ## `manual_seed` seeds the shared random generator.
    ##
    ## manual_seed(seed: int) -> null
    __manual_seed(seed)
}

fun slice(a, dim, start, end, step=1) {
    ##std:this,__slice
    ## `slice` selects `start` to `end` with `step` along `dim`, like
    ## a[start:end:step]. Negative start and end count from the end of the dim.
    ## A negative step reverses, and end=-1 means through index 0.
    ##
    ## slice(a: tensor, dim: int, start: int, end: int, step: int=1) -> tensor
    __slice(a, dim, start, end, step)
}

fun select(a, dim, index) {
    ##std:this,__select
    ## `select` returns the entries at `index` along `dim` with that dim removed,
    ## covering x[i] and x[:, j].
    ##
    ## select(a: tensor, dim: int, index: int) -> tensor
    __select(a, dim, index)
}

fun batch(a, start, end) {
    ##std:this,__slice
    ## `batch` returns rows start:end along dim 0, a convenience over `slice`
    ## for minibatch training.
    ##
    ## batch(a: tensor, start: int, end: int) -> tensor
    __slice(a, 0, start, end, 1)
}

fun index_select(a, dim, index) {
    ##std:this,__index_select
    ## `index_select` gathers rows or slices along `dim` using a 1d index tensor,
    ## matching torch.index_select. It is differentiable.
    ##
    ## index_select(a: tensor, dim: int, index: tensor) -> tensor
    __index_select(a, dim, index)
}

fun masked_fill(a, mask, value) {
    ##std:this,__masked_fill
    ## `masked_fill` replaces elements where `mask` is true with `value`.
    ##
    ## masked_fill(a: tensor, mask: tensor, value: float) -> tensor
    __masked_fill(a, mask, value)
}

fun masked_select(a, mask) {
    ##std:this,__masked_select
    ## `masked_select` returns the elements where `mask` is nonzero as a 1d
    ## tensor. The output length depends on the data, so it is inference-only.
    ##
    ## masked_select(a: tensor, mask: tensor) -> tensor
    __masked_select(a, mask)
}

fun flip(a, dims) {
    ##std:this,__flip
    ## `flip` reverses elements along each dim in `dims`, matching torch.flip.
    ## It is differentiable.
    ##
    ## flip(a: tensor, dims: list[int]) -> tensor
    __flip(a, dims)
}

fun embedding(weight, indices) {
    ##std:this,__embedding
    ## `embedding` looks up rows of a [vocab, dim] weight tensor by an index
    ## tensor, matching torch.nn.functional.embedding. It is differentiable.
    ##
    ## embedding(weight: tensor, indices: tensor) -> tensor
    __embedding(weight, indices)
}

fun bmm(a, b) {
    ##std:this,__bmm
    ## `bmm` is batched matrix multiplication over the last two dims.
    ##
    ## bmm(a: tensor, b: tensor) -> tensor
    __bmm(a, b)
}

fun silu(a) {
    ##std:this,__silu
    ## `silu` is the Sigmoid Linear Unit, x * sigmoid(x).
    ##
    ## silu(a: tensor) -> tensor
    __silu(a)
}

fun rand(shape, datatype=dtype.float32, dev=device.cpu, requires_grad=false) {
    ##std:this,__rand
    ## `rand` returns uniform samples in [0, 1) from the shared engine RNG.
    ##
    ## rand(shape: list[int], datatype: str='float32', dev: str='cpu', requires_grad: bool=false) -> tensor
    __rand(shape, datatype, dev, requires_grad)
}

fun randperm(n, dev=device.cpu) {
    ##std:this,__randperm
    ## `randperm` returns a random permutation of 0..n-1, reproducible after
    ## manual_seed.
    ##
    ## randperm(n: int, dev: str='cpu') -> tensor
    __randperm(n, dev)
}

fun shuffle(a, dim=0) {
    ##std:this,__shuffle
    ## `shuffle` returns a copy of `a` with `dim` permuted by a random
    ## permutation, for minibatch training.
    ##
    ## shuffle(a: tensor, dim: int=0) -> tensor
    __shuffle(a, dim)
}

fun log_softmax(a, dim=-1) {
    ##std:this,__max,__sum,__exp,__log,__sub
    ## `log_softmax` returns log(softmax(a, dim)) computed stably by subtracting
    ## the max before the log-sum-exp. It is differentiable.
    ##
    ## log_softmax(a: tensor, dim: int=-1) -> tensor
    val m = __max(a, dim, true);
    val shifted = __sub(a, m);
    val lse = __log(__sum(__exp(shifted), dim, true));
    return __sub(shifted, lse);
}

fun nll_loss(log_probs, target) {
    ##std:this,__gather,__unsqueeze,__mean,__neg
    ## `nll_loss` is the mean negative log likelihood of `target` class indices
    ## under `log_probs` [batch, classes]. Pair it with log_softmax.
    ##
    ## nll_loss(log_probs: tensor, target: tensor) -> tensor
    val picked = __gather(log_probs, 1, __unsqueeze(target, 1));
    return __neg(__mean(picked, null, false));
}

fun variance(a, dim=null, keepdim=false) {
    ##std:this,__mean,__sub,__mul
    ## `variance` is the population variance over `dim` (or every element).
    ##
    ## variance(a: tensor, dim: int|list[int]|null=null, keepdim: bool=false) -> tensor
    val m = __mean(a, dim, true);
    val d = __sub(a, m);
    return __mean(__mul(d, d), dim, keepdim);
}

fun std(a, dim=null, keepdim=false) {
    ##std:this,__sqrt
    ## `std` is the square root of `variance`.
    ##
    ## std(a: tensor, dim: int|list[int]|null=null, keepdim: bool=false) -> tensor
    return __sqrt(variance(a, dim, keepdim));
}

fun norm(a, dim=null, keepdim=false) {
    ##std:this,__sqrt,__sum,__mul
    ## `norm` is the L2 norm over `dim` (or every element).
    ##
    ## norm(a: tensor, dim: int|list[int]|null=null, keepdim: bool=false) -> tensor
    return __sqrt(__sum(__mul(a, a), dim, keepdim));
}

fun clip_grad_norm_(params, max_norm) {
    ##std:this,__optim_clip_grad_norm
    ## `clip_grad_norm_` scales the last backward's gradients in place so their
    ## global L2 norm over `params` is at most `max_norm`, matching
    ## torch.nn.utils.clip_grad_norm_. It returns the norm before clipping.
    ##
    ## clip_grad_norm_(params: list, max_norm: float) -> float
    __optim_clip_grad_norm(params, max_norm)
}

fun clone(a) {
    ##std:this,__clone
    ## `clone` returns a deep copy that stays in the autograd graph, like
    ## torch.Tensor.clone. Use `a.detach()` for a copy cut out of the graph.
    ##
    ## clone(a: tensor) -> tensor
    __clone(a)
}

fun state_dict(model) {
    ##std:this,__state_dict
    ## `state_dict` returns a map of dotted names to the tensors in a model,
    ## walking maps, lists, tensors, and module parameters.
    ##
    ## state_dict(model: map|list|tensor|module) -> map[str]tensor
    __state_dict(model)
}

fun save_state(model, path) {
    ##std:this,__save_state
    ## `save_state` writes a model's named tensors to one file, so it can be
    ## rebuilt and loaded by name.
    ##
    ## save_state(model: map|list|tensor|module, path: str) -> null
    __save_state(model, path)
}

fun load_state(model, path) {
    ##std:this,__load_state
    ## `load_state` loads named tensors from a file into a model in place,
    ## matching names. Missing names are skipped and shape mismatches error.
    ##
    ## load_state(model: map|list|tensor|module, path: str) -> null
    __load_state(model, path)
}

fun clamp(a, min_val, max_val) {
    ##std:this,__clamp
    ## `clamp` limits each element to [min, max].
    ##
    ## clamp(a: tensor, min: float, max: float) -> tensor
    __clamp(a, min_val, max_val)
}

fun where(condition, a, b) {
    ##std:this,__where
    ## `where` selects a where condition is true and b otherwise.
    ##
    ## where(condition: tensor, a: tensor, b: tensor) -> tensor
    __where(condition, a, b)
}

fun cat(tensors, dim=0) {
    ##std:this,__cat
    ## `cat` joins a list of tensors along `dim`, like torch.cat. Every tensor
    ## must share all dimensions except `dim`.
    ##
    ## cat(tensors: list[tensor], dim: int=0) -> tensor
    __cat(tensors, dim)
}

fun gather(a, dim, index) {
    ##std:this,__gather
    ## `gather` selects entries along `dim` using an int index tensor, matching
    ## torch.gather (out[i][j] = a[i][index[i][j]] for dim=1). It is
    ## differentiable, so it can be used inside models.
    ##
    ## gather(a: tensor, dim: int, index: tensor) -> tensor
    __gather(a, dim, index)
}

fun onehot(labels, classes) {
    ##std:this,__onehot
    ## `onehot` turns a 1d label tensor into a [n, classes] matrix.
    ##
    ## onehot(labels: tensor, classes: int) -> tensor
    __onehot(labels, classes)
}

fun set_requires_grad(a, flag) {
    ##std:this,__set_requires_grad
    ## `set_requires_grad` turns gradient tracking on or off for a tensor.
    ##
    ## set_requires_grad(a: tensor, flag: bool) -> null
    __set_requires_grad(a, flag)
}

fun requires_grad(a) {
    ##std:this,__requires_grad
    ## `requires_grad` reports whether a tensor is tracked.
    ##
    ## requires_grad(a: tensor) -> bool
    __requires_grad(a)
}

fun backward(a) {
    ##std:this,__backward
    ## `backward` runs reverse-mode autograd from a scalar tensor.
    ##
    ## backward(a: tensor) -> null
    __backward(a)
}

fun grad(a) {
    ##std:this,__grad
    ## `grad` returns the accumulated gradient or null.
    ##
    ## grad(a: tensor) -> tensor|null
    __grad(a)
}

fun zero_grad(a) {
    ##std:this,__zero_grad
    ## `zero_grad` clears a tensor's gradient.
    ##
    ## zero_grad(a: tensor) -> null
    __zero_grad(a)
}

fun no_grad(f) {
    ##std:this,__set_grad_enabled
    ## `no_grad` runs a closure with graph building disabled.
    ##
    ## no_grad(f: fun) -> any
    val prev = __set_grad_enabled(false);
    val out = f();
    __set_grad_enabled(prev);
    return out;
}

fun cross_entropy(logits, target) {
    ##std:this,__cross_entropy
    ## `cross_entropy` returns the mean cross-entropy loss between logits
    ## [batch, classes] and class indices [batch].
    ##
    ## cross_entropy(logits: tensor, target: tensor) -> tensor
    __cross_entropy(logits, target)
}

fun cross_entropy_grad(logits, target) {
    ##std:this,__cross_entropy_grad
    ## `cross_entropy_grad` returns d(mean cross-entropy)/d(logits), the seed a
    ## compiled backward needs when the loss is kept outside the compiled graph.
    ##
    ## cross_entropy_grad(logits: tensor, target: tensor) -> tensor
    __cross_entropy_grad(logits, target)
}

fun mse_loss(predictions, targets) {
    ##std:this,__mse_loss
    ## `mse_loss` returns the mean squared error between two tensors of equal shape.
    ##
    ## mse_loss(predictions: tensor, targets: tensor) -> tensor
    __mse_loss(predictions, targets)
}

fun run_eager(m, x) { m.forward(x); }

fun compile(model, example) {
    ##std:this,__compiled_new,__trace_begin,__trace_input,__trace_end,__graph_optimize,__compiled_has_plan,__compiled_install,__compiled_run,__compiled_stats,__compiled_backward
    ## `compile` captures a model's forward pass, fuses and prunes the recorded
    ## graph, and returns a module whose forward replays it.
    ##
    ## The graph is specialised for the example's shape, dtype and device. A
    ## call with a different signature recompiles automatically (torch.compile
    ## guards work the same way), so the same compiled model works across batch
    ## sizes.
    ##
    ## A forward that uses an op the tracer cannot record, or that reads host
    ## data, falls back to eager execution rather than failing, so ml.compile
    ## never makes a working model stop working.
    ##
    ## compile(model: module, example: tensor) -> module
    val cache = __compiled_new();

    var this = {};
    this.__cache = cache;
    this.__model = model;
    # Mutable state lives on the map, not in a captured local: blue closures
    # capture captured variables by value, so a local flag would not propagate.
    this.__fellback = false;
    this.is_eager = fun() { return this.__fellback; };
    this.stats = fun() {
        if (this.__fellback) { return "eager-fallback"; }
        return __compiled_stats(cache);
    };
    this.forward = fun(x) {
        val xin = x;
        if (this.__fellback) { return run_eager(model, xin); }
        if (!__compiled_has_plan(cache, xin)) {
            val tracer = __trace_begin(xin);
            val out = model.forward(__trace_input(tracer));
            val graph = __trace_end(tracer, out);
            if (graph == null) {
                # The forward uses an op the tracer cannot record, or reads host
                # data. Run it eagerly from here on instead of failing.
                this.__fellback = true;
                return run_eager(model, xin);
            }
            __graph_optimize(graph);
            __compiled_install(cache, xin, graph);
        }
        return __compiled_run(cache, xin);
    };
    # backward runs the compiled (differentiated) graph, so the backward's own
    # elementwise chains are fused the same way the forward's are. An optional
    # seed is the loss gradient with respect to the forward output; without one
    # it differentiates the sum of the output. Under the eager fallback the
    # eager tape backward runs instead.
    this.backward = fun(x, seed=null) {
        if (this.__fellback) {
            val out = run_eager(model, x);
            if (seed == null) {
                backward(sum(out));
            } else {
                backward(sum(mul(out, seed)));
            }
        } else {
            if (seed == null) {
                __compiled_backward(cache, x);
            } else {
                __compiled_backward(cache, x, seed);
            }
        }
    };
    this.parameters = fun() { return model.parameters(); };
    this.to = fun(dev) { model.to(dev); return this; };

    # compile the example shape up front, so a model that cannot be traced is
    # switched to eager before the first real call
    this.forward(example);
    return this;
}

fun add_(a, b) {
    ##std:this,__add_
    ## `add_` adds b (or a scalar) into a in place.
    ##
    ## add_(a: tensor, b: tensor) -> null
    __add_(a, b)
}

fun sub_(a, b) {
    ##std:this,__sub_
    ## `sub_` subtracts b (or a scalar) from a in place.
    ##
    ## sub_(a: tensor, b: tensor) -> null
    __sub_(a, b)
}

fun mul_(a, b) {
    ##std:this,__mul_
    ## `mul_` multiplies a by b (or a scalar) in place.
    ##
    ## mul_(a: tensor, b: tensor) -> null
    __mul_(a, b)
}

fun div_(a, b) {
    ##std:this,__div_
    ## `div_` divides a by b (or a scalar) in place.
    ##
    ## div_(a: tensor, b: tensor) -> null
    __div_(a, b)
}

fun save(a, path) {
    ##std:this,__save
    ## `save` writes a tensor to a file.
    ##
    ## save(a: tensor, path: str) -> null
    __save(a, path)
}

fun load(path, dev=device.cpu) {
    ##std:this,__load
    ## `load` reads a tensor written by `save` onto a device.
    ##
    ## load(path: str, dev: str='cpu') -> tensor
    __load(path, dev)
}

fun cast(a, datatype) {
    ##std:this,__cast
    ## `cast` converts a tensor to another dtype.
    ##
    ## cast(a: tensor, datatype: str) -> tensor
    __cast(a, datatype)
}

## `nn` holds the small PyTorch-style layer set. Each layer is a map with
## `forward(x)` and `parameters()`. The math runs through the ml ops, which
## delegate to the borncgo engine.
## `ml.parameters` collects trainable leaves from a model. A map or list is
## walked recursively, tensors with requires_grad are kept, and nn module
## parameters are returned as handles. This is the PyTorch model.parameters()
## replacement for models that are plain maps.
fun parameters(model) {
    ##std:this,__nn_parameters
    ## `parameters` collects trainable leaves from a model. A map with a
    ## `parameters` method (an nn module) uses it; a plain map or list is walked
    ## recursively; a tensor with requires_grad is returned; a module handle uses
    ## the engine's parameter list.
    ##
    ## parameters(model: map|list|tensor|module) -> list[tensor|param]
    return __nn_parameters(model);
}

## `nn` holds the small PyTorch-style layer set. A module is a plain map, not a
## type: it has `forward(x)` and `parameters()`, and optionally `to(dev)` and
## `set_training(flag)`. Layers built from raw tensors (Embedding, LayerNorm,
## RMSNorm) store those tensors on the map, so `ml.parameters` finds them without
## calling closures. `nn.Sequential` stores its module list for the same reason.
## `nn.Linear`/`ReLU`/`Sigmoid` wrap an engine module in a handle.
val nn = {
    'Linear': fun(in_features, out_features, dev=device.cpu) {
        val h = __nn_linear(in_features, out_features, dev);
        var this = {};
        this.__handle = h;
        this.kind = 'linear';
        this.forward = fun(x) { return __nn_forward(h, x); };
        # fused matmul + bias (+ relu), one dispatch
        this.forward_act = fun(x, relu) { return __nn_linear_act(h, x, relu); };
        this.parameters = fun() { return __nn_parameters(h); };
        this.to = fun(dev) { __nn_to(h, dev); return this; };
        return this;
    },
    'ReLU': fun(dev=device.cpu) {
        val h = __nn_relu(dev);
        var this = {};
        this.__handle = h;
        this.kind = 'relu';
        this.forward = fun(x) { return __nn_forward(h, x); };
        this.parameters = fun() { return __nn_parameters(h); };
        this.to = fun(dev) { __nn_to(h, dev); return this; };
        return this;
    },
    'Sigmoid': fun(dev=device.cpu) {
        val h = __nn_sigmoid(dev);
        var this = {};
        this.__handle = h;
        this.kind = 'sigmoid';
        this.forward = fun(x) { return __nn_forward(h, x); };
        this.parameters = fun() { return __nn_parameters(h); };
        this.to = fun(dev) { __nn_to(h, dev); return this; };
        return this;
    },
    'SiLU': fun() {
        var this = {};
        this.kind = 'silu';
        this.forward = fun(x) { return __silu(x); };
        this.parameters = fun() { return []; };
        this.to = fun(dev) { return this; };
        return this;
    },
    'Embedding': fun(num_embeddings, embedding_dim, dev=device.cpu) {
        var this = {};
        this.kind = 'embedding';
        # weight is a plain tensor, so ml.parameters finds it without a module
        this.weight = __randn([num_embeddings, embedding_dim], 'float32', dev, true);
        this.forward = fun(indices) { return __embedding(this.weight, indices); };
        this.parameters = fun() { return [this.weight]; };
        this.to = fun(dev) { this.weight = this.weight.to(dev); return this; };
        return this;
    },
    'LayerNorm': fun(d_model, eps=1e-5, dev=device.cpu) {
        var this = {};
        this.kind = 'layernorm';
        this.weight = __ones([d_model], 'float32', dev, true);
        this.bias = __zeros([d_model], 'float32', dev, true);
        this.forward = fun(x) {
            val mu = __mean(x, -1, true);
            val xc = __sub(x, mu);
            val v = __mean(__mul(xc, xc), -1, true);
            val inv = __pow(__add(v, eps), -0.5);
            return __add(__mul(__mul(xc, inv), this.weight), this.bias);
        };
        this.parameters = fun() { return [this.weight, this.bias]; };
        this.to = fun(dev) {
            this.weight = this.weight.to(dev);
            this.bias = this.bias.to(dev);
            return this;
        };
        return this;
    },
    'RMSNorm': fun(d_model, eps=1e-5, dev=device.cpu) {
        var this = {};
        this.kind = 'rmsnorm';
        this.weight = __ones([d_model], 'float32', dev, true);
        this.forward = fun(x) {
            val ms = __mean(__mul(x, x), -1, true);
            val scale = __pow(__add(ms, eps), -0.5);
            return __mul(__mul(x, scale), this.weight);
        };
        this.parameters = fun() { return [this.weight]; };
        this.to = fun(dev) { this.weight = this.weight.to(dev); return this; };
        return this;
    },
    'Dropout': fun(p=0.5) {
        var this = {};
        this.kind = 'dropout';
        this.p = p;
        this.training = true;
        this.forward = fun(x) {
            if (!this.training || this.p <= 0.0) { return x; }
            # x['device'] (not x.device) because `device` is a val in this module
            val keep = __gt(__rand(x.shape, 'float32', x['device'], false), this.p);
            val m = __cast(keep, 'float32');
            return __div(__mul(x, m), 1.0 - this.p);
        };
        this.parameters = fun() { return []; };
        this.set_training = fun(flag) { this.training = flag; return this; };
        this.to = fun(dev) { return this; };
        return this;
    },
    'Sequential': fun(modules) {
        var this = {};
        # The module list is stored on the map so parameter discovery can walk
        # it without calling closures. Fusion lives in ml.compile and in the
        # Linear fused kernel, so the container is a plain fold over the modules.
        this.modules = modules;
        this.forward = fun(x) {
            var out = x;
            for (m in modules) { out = m.forward(out); }
            return out;
        };
        this.parameters = fun() {
            var ps = [];
            for (m in modules) {
                for (p in m.parameters()) { push(ps, p); }
            }
            return ps;
        };
        this.to = fun(dev) {
            for (m in modules) { m.to(dev); }
            return this;
        };
        return this;
    },
    'parameters': fun(model) { return parameters(model); },
};

## `optim` holds the optimizers. State lives on the optimizer object because
## closures capture scalars by value.
val optim = {
    'SGD': fun(params, lr=0.01, momentum=0.0, weight_decay=0.0) {
        val h = __optim_sgd(params, lr, momentum, weight_decay);
        var this = {};
        this.step = fun() { __optim_step(h); };
        this.zero_grad = fun() { __optim_zero_grad(h); };
        this.set_lr = fun(new_lr) { __optim_set_lr(h, new_lr); return this; };
        return this;
    },
    'Adam': fun(params, lr=0.001, beta1=0.9, beta2=0.999, eps=1e-8, weight_decay=0.0) {
        val h = __optim_adam(params, lr, beta1, beta2, eps, weight_decay);
        var this = {};
        this.step = fun() { __optim_step(h); };
        this.zero_grad = fun() { __optim_zero_grad(h); };
        this.set_lr = fun(new_lr) { __optim_set_lr(h, new_lr); return this; };
        return this;
    },
    # AdamW is Adam with decoupled weight decay on by default.
    'AdamW': fun(params, lr=0.001, beta1=0.9, beta2=0.999, eps=1e-8, weight_decay=0.01) {
        val h = __optim_adam(params, lr, beta1, beta2, eps, weight_decay);
        var this = {};
        this.step = fun() { __optim_step(h); };
        this.zero_grad = fun() { __optim_zero_grad(h); };
        this.set_lr = fun(new_lr) { __optim_set_lr(h, new_lr); return this; };
        return this;
    },
    # Learning-rate schedules are pure functions of the step. Call
    # opt.set_lr(ml.optim.cosine_lr(base, step, total)) each step.
    'step_lr': fun(base_lr, step, step_size, gamma=0.1) {
        __lr_schedule('step', base_lr, step, step_size, gamma);
    },
    'cosine_lr': fun(base_lr, step, total_steps, min_lr=0.0) {
        __lr_schedule('cosine', base_lr, step, total_steps, min_lr);
    },
    'warmup_lr': fun(base_lr, step, warmup_steps) {
        __lr_schedule('warmup', base_lr, step, warmup_steps, 0.0);
    },
};