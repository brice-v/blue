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
val __abs = _abs;
val __sigmoid = _sigmoid;
val __tanh = _tanh;
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

fun slice(a, dim, start, end) {
    ##std:this,__slice
    ## `slice` returns a view of `dim` from `start` to `end`.
    ##
    ## slice(a: tensor, dim: int, start: int, end: int) -> tensor
    __slice(a, dim, start, end)
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
    'Sequential': fun(modules) {
        var this = {};
        # Fuse a Linear with the ReLU that follows it into one kernel. This is a
        # small, structural "compile": the module list is the graph, and the
        # Linear -> ReLU pair is the pattern that matters.
        this.forward = fun(x) {
            var out = x;
            var next = 0;
            val n = modules.len();
            for (i in 0..(n - 1)) {
                if (i >= next) {
                    val m = modules[i];
                    if (m.kind == 'linear') {
                        var fuse = false;
                        if (i + 1 <= n - 1) {
                            if (modules[i + 1].kind == 'relu') {
                                fuse = true;
                                next = i + 2;
                            } else {
                                next = i + 1;
                            }
                        } else {
                            next = i + 1;
                        }
                        out = m.forward_act(out, fuse);
                    } else {
                        out = m.forward(out);
                        next = i + 1;
                    }
                }
            }
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
    'parameters': fun(model) { return model.parameters(); },
};

## `optim` holds the optimizers. State lives on the optimizer object because
## closures capture scalars by value.
val optim = {
    'SGD': fun(params, lr=0.01, momentum=0.0) {
        val h = __optim_sgd(params, lr, momentum);
        var this = {};
        this.step = fun() { __optim_step(h); };
        this.zero_grad = fun() { __optim_zero_grad(h); };
        return this;
    },
    'Adam': fun(params, lr=0.001, beta1=0.9, beta2=0.999, eps=1e-8) {
        val h = __optim_adam(params, lr, beta1, beta2, eps);
        var this = {};
        this.step = fun() { __optim_step(h); };
        this.zero_grad = fun() { __optim_zero_grad(h); };
        return this;
    },
};