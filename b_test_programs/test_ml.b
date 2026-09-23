## End-state integration test for the `ml` module (see ML_PLAN.md).
##
## The VM stops at the first error, so the file lights up top to bottom as
## features land.
##
## Value checks use two helpers:
##   same(a, b)  -> ml.equal(a, b)     exact shape + dtype + element equality
##   close(a, b) -> ml.allclose(a, b)  equality within rtol/atol, for computed
##                                     results such as softmax, exp, sqrt
## Do not use `==` to compare two tensors in a test: it is the elementwise
## comparison operator and returns a bool tensor, which `assert` cannot take.
## Bool tensors can be checked with same()/close(), or read elementwise with
## to_list().

import ml

fun same(a, b) {
    ## exact value equality for two tensors
    ml.equal(a, b)
}

fun close(a, b) {
    ## value equality within tolerance, for computed results
    ml.allclose(a, b)
}

# --- 1. creation and properties ---------------------------------------------

val a = ml.tensor([[1.0, 2.0, 3.0], [4.0, 5.0, 6.0]]);      # 2x3
val b = ml.tensor([[7.0, 8.0], [9.0, 10.0], [11.0, 12.0]]); # 3x2

assert(type(a) == "TENSOR");
assert(a.shape == [2, 3]);
assert(a.ndim == 2);
assert(a.strides == [3, 1]);
assert(a.offset == 0);
assert(a.dtype == "float32");
assert(a.device == "cpu");
assert(a.requires_grad == false);
assert(a.grad == null);

# transpose is a metadata-only property
assert(a.T.shape == [3, 2]);
assert(a.T.strides == [1, 3]);

# a scalar literal becomes a 0-D tensor, like PyTorch
assert(ml.tensor(5.0).shape == []);
assert(ml.tensor(5.0).ndim == 0);

# dtype is inferred from the values, like torch.tensor
assert(ml.tensor([1, 2, 3]).dtype == "int64");
assert(ml.tensor(7).dtype == "int64");
assert(ml.tensor(7.0).dtype == "float32");
assert(ml.tensor([1, 2.5]).dtype == "float32");
assert(ml.tensor([true, false]).dtype == "bool");
assert(ml.tensor([true, 1]).dtype == "int64");

# the repr matches PyTorch's for the inferred dtype (bools use blue's spelling)
assert(str(ml.tensor([1, 2])) == "tensor([1, 2])");
assert(str(ml.tensor([1.0, 2.0])) == "tensor([1., 2.])");
assert(str(ml.tensor([true, false])) == "tensor([ true, false])");

# every optional parameter can be passed by keyword, in any order
assert(ml.tensor([[1.0]], datatype="float32", dev="cpu", requires_grad=false).shape == [1, 1]);

# --- 2. binary operators ----------------------------------------------------

# Binary arithmetic and comparisons have one spelling each: the operators and
# the ml.* functions. There are no a.add(b)-style methods for them.
val c = a @ b;                                                             # 2x2
assert(c.shape == [2, 2]);
assert(same(c, ml.tensor([[58.0, 64.0], [139.0, 154.0]])));

assert(same(c + c, ml.tensor([[116.0, 128.0], [278.0, 308.0]])));
assert(same(c - c, ml.tensor([[0.0, 0.0], [0.0, 0.0]])));
assert(same(c * c, ml.tensor([[3364.0, 4096.0], [19321.0, 23716.0]])));
assert(same(c / c, ml.tensor([[1.0, 1.0], [1.0, 1.0]])));

# unary methods
assert(same(a.relu(), a));
assert(same(ml.tensor([[-1.0, 2.0]]).relu(), ml.tensor([[0.0, 2.0]])));
assert(ml.tensor([0.0]).exp().item() == 1.0);
assert(ml.tensor([1.0]).log().item() == 0.0);
assert(ml.tensor([4.0]).sqrt().item() == 2.0);

# item on a single-element tensor
assert(ml.tensor([[6.0]]).item() == 6.0);

# --- 3. matmul operator `@` -------------------------------------------------

assert(same(a @ b, c));

# --- 4. autograd ------------------------------------------------------------

val w = ml.tensor([[2.0]], requires_grad=true);
val x = ml.tensor([[3.0]]);
val y = x @ w; # [[6.0]]

assert(y.item() == 6.0);
assert(w.grad == null);

y.backward();
assert(w.grad.item() == 3.0); # dy/dw = x = 3
assert(x.grad == null);       # x is not tracked

w.zero_grad();
assert(w.grad == null);

# --- 5. module-level functions ----------------------------------------------

val e = c + c;
val m = ml.matmul(a, b);
assert(same(m, c));
val n = ml.add(c, c);
assert(same(n, e));

# keyword arguments bind by name and may be reordered
assert(same(ml.add(a=c, b=c), e));
assert(same(ml.matmul(a=a, b=b), c));

# --- 6. TARGET: arithmetic operators ----------------------------------------

assert(same(c + c, e));
assert(same(c - c, ml.tensor([[0.0, 0.0], [0.0, 0.0]])));
assert(same(c * c, ml.tensor([[3364.0, 4096.0], [19321.0, 23716.0]])));
assert(same(c / c, ml.tensor([[1.0, 1.0], [1.0, 1.0]])));

# scalar broadcast operators (float and int scalars both coerce)
assert(same(c + 1.0, ml.tensor([[59.0, 65.0], [140.0, 155.0]])));
assert(same(1.0 + c, ml.tensor([[59.0, 65.0], [140.0, 155.0]])));
assert(same(c * 2.0, ml.tensor([[116.0, 128.0], [278.0, 308.0]])));
assert(same(2.0 * c, ml.tensor([[116.0, 128.0], [278.0, 308.0]])));
assert(same(c + 1, ml.tensor([[59.0, 65.0], [140.0, 155.0]])));
assert(same(1 + c, ml.tensor([[59.0, 65.0], [140.0, 155.0]])));
assert(same(2 * c, ml.tensor([[116.0, 128.0], [278.0, 308.0]])));

# scalar on the left for the non-commutative operators: operand order matters
assert(same(1.0 - c, ml.tensor([[-57.0, -63.0], [-138.0, -153.0]])));
assert(same(c - 1.0, ml.tensor([[57.0, 63.0], [138.0, 153.0]])));
assert(close(2.0 / c, ml.tensor([[2.0 / 58.0, 2.0 / 64.0], [2.0 / 139.0, 2.0 / 154.0]])));
assert(close(c / 2.0, ml.tensor([[29.0, 32.0], [69.5, 77.0]])));

# TARGET: unary negation
assert(same(-c, ml.tensor([[-58.0, -64.0], [-139.0, -154.0]])));

# TARGET: pow works for a negative base with an integer exponent, and the
# gradient reaches the base (dz/da = b * a^(b-1))
assert(same(c ** 2.0, ml.tensor([[3364.0, 4096.0], [19321.0, 23716.0]])));
assert(same(ml.tensor([[-2.0, 3.0]]) ** 2.0, ml.tensor([[4.0, 9.0]])));
val pbase = ml.tensor([[-3.0]], requires_grad=true);
val pout = pbase ** 2.0;
assert(same(pout, ml.tensor([[9.0]])));
pout.backward();
assert(same(pbase.grad, ml.tensor([[-6.0]])));

# compound assignment desugars to the binary op
var acc = c;
acc += c;
assert(same(acc, ml.tensor([[116.0, 128.0], [278.0, 308.0]])));

# --- 7. TARGET: property assignment -----------------------------------------

var prop = ml.tensor([[1.0]]);
assert(prop.requires_grad == false);
prop.requires_grad = true;
assert(prop.requires_grad == true);

# --- 8. TARGET: comparisons return elementwise bool tensors -----------------

# to_list is the core builtin, extended to accept a tensor; the method form too
assert(to_list(c) == [[58.0, 64.0], [139.0, 154.0]]);
assert(c.to_list() == [[58.0, 64.0], [139.0, 154.0]]);
assert(to_list(c == (a @ b)) == [[true, true], [true, true]]);
assert(to_list(c > 0.0) == [[true, true], [true, true]]);
assert(to_list(c < 0.0) == [[false, false], [false, false]]);
assert(to_list(c != c) == [[false, false], [false, false]]);

# the module functions and methods match the operators, and take scalars
assert(same(c > 0.0, ml.gt(c, 0.0)));
assert(same(c >= 0.0, ml.ge(c, 0.0)));
assert(same(c < 0.0, ml.lt(c, 0.0)));
assert(same(c <= 0.0, ml.le(c, 0.0)));
assert(same(c != c, ml.ne(c, c)));
assert(same(c == (a @ b), ml.eq(c, a @ b)));
assert(same(c.neg(), -c));

# keyword arguments on comparisons and on the tolerance-based helpers
assert(same(ml.gt(a=c, b=0.0), ml.gt(c, 0.0)));
assert(same(ml.le(a=c, b=0.0), ml.le(c, 0.0)));
assert(ml.allclose(c, c, rtol=1e-5, atol=1e-8));
assert(ml.allclose(b=c, a=c, atol=1e-8, rtol=1e-5));

# --- 9. TARGET: reductions --------------------------------------------------

val r = ml.tensor([[1.0, 2.0, 3.0], [4.0, 5.0, 6.0]]);
assert(same(r.sum(0), ml.tensor([5.0, 7.0, 9.0])));
assert(same(r.sum(1), ml.tensor([6.0, 15.0])));
assert(same(r.sum(1, keepdim=true), ml.tensor([[6.0], [15.0]])));
# the reduction arguments also bind by keyword, on the method and the module form
assert(same(r.sum(dim=0), r.sum(0)));
assert(same(r.sum(dim=1, keepdim=true), r.sum(1, keepdim=true)));
assert(same(ml.sum(r, dim=0), ml.tensor([5.0, 7.0, 9.0])));
assert(same(ml.sum(a=r, dim=1, keepdim=true), ml.tensor([[6.0], [15.0]])));
assert(same(r.mean(0), ml.tensor([2.5, 3.5, 4.5])));
assert(same(r.max(1), ml.tensor([3.0, 6.0])));
assert(same(r.min(1), ml.tensor([1.0, 4.0])));
# argmax/argmin return integer tensors, like PyTorch
assert(to_list(r.argmax(1)) == [2, 2]);
assert(to_list(r.argmin(1)) == [0, 0]);
assert(r.argmax(1).dtype == "int32");
assert(ml.sum(r).item() == 21.0);
assert(ml.mean(r).item() == 3.5);

# --- 10. TARGET: shape ops and remaining elementwise ------------------------

val s = ml.tensor([[1.0, 2.0, 3.0], [4.0, 5.0, 6.0]]);
assert(s.reshape([3, 2]).shape == [3, 2]);
assert(same(s.reshape([6]), ml.tensor([1.0, 2.0, 3.0, 4.0, 5.0, 6.0])));
assert(same(ml.transpose(s, 0, 1), s.T));
assert(same(ml.transpose(s, dim0=0, dim1=1), s.T));
assert(s.permute([1, 0]).shape == [3, 2]);
assert(same(s.flatten(), ml.tensor([1.0, 2.0, 3.0, 4.0, 5.0, 6.0])));
assert(ml.tensor([[1.0, 2.0]]).unsqueeze(0).shape == [1, 1, 2]);
assert(ml.tensor([[[1.0, 2.0]]]).squeeze().shape == [2]);
assert(ml.tensor([[1.0, 2.0]]).broadcast_to([3, 2]).shape == [3, 2]);

val u = ml.tensor([[-1.0, 0.0, 4.0]]);
assert(same(u.abs(), ml.tensor([[1.0, 0.0, 4.0]])));
assert(same(u.neg(), ml.tensor([[1.0, 0.0, -4.0]])));
assert(same(u ** ml.tensor([[2.0, 2.0, 2.0]]), ml.tensor([[1.0, 0.0, 16.0]])));
assert(ml.tensor([0.0]).sigmoid().item() == 0.5);
assert(ml.tensor([0.0]).tanh().item() == 0.0);
assert(close(ml.softmax(ml.tensor([[1.0, 2.0, 3.0]]), 1),
             ml.tensor([[0.09003057, 0.24472847, 0.66524096]])));

# --- 11. TARGET: creation ops -----------------------------------------------

assert(same(ml.zeros([2, 2]), ml.tensor([[0.0, 0.0], [0.0, 0.0]])));
assert(same(ml.ones([2, 2]), ml.tensor([[1.0, 1.0], [1.0, 1.0]])));
assert(same(ml.full([2, 2], 5.0), ml.tensor([[5.0, 5.0], [5.0, 5.0]])));
assert(ml.randn([2, 3]).shape == [2, 3]);
assert(same(ml.arange(0.0, 5.0, 1.0), ml.tensor([0.0, 1.0, 2.0, 3.0, 4.0])));
assert(same(ml.eye(2), ml.tensor([[1.0, 0.0], [0.0, 1.0]])));
assert(ml.tensor([[1.0]], datatype="float64").dtype == "float64");
assert(ml.randn([2, 2], requires_grad=true).requires_grad == true);
ml.manual_seed(0);

# --- 12. TARGET: indexing and selection -------------------------------------

assert(same(ml.slice(r, 0, 0, 1), ml.tensor([[1.0, 2.0, 3.0]])));
assert(same(ml.clamp(ml.tensor([[-5.0, 0.5, 5.0]]), 0.0, 1.0), ml.tensor([[0.0, 0.5, 1.0]])));
assert(same(ml.where(ml.tensor([[1.0, 0.0]]), ml.tensor([[10.0, 20.0]]), ml.tensor([[30.0, 40.0]])),
            ml.tensor([[10.0, 40.0]])));
assert(same(ml.onehot(ml.tensor([0.0, 2.0]), 3),
            ml.tensor([[1.0, 0.0, 0.0], [0.0, 0.0, 1.0]])));

# --- 13. TARGET: autograd control builtins ----------------------------------

val g = ml.tensor([[1.0, 2.0]], requires_grad=true);
ml.set_requires_grad(g, true);
assert(ml.requires_grad(g) == true);
ml.backward(g.sum());
assert(same(ml.grad(g), ml.tensor([[1.0, 1.0]])));
ml.zero_grad(g);
assert(ml.grad(g) == null);

val det = ml.tensor([[1.0]], requires_grad=true).detach();
assert(det.requires_grad == false);

# no_grad disables graph building inside the closure
val ng = ml.no_grad(fun() { ml.mul(ml.tensor([[1.0]], requires_grad=true), ml.tensor([[2.0]])) });
assert(ng.requires_grad == false);

# --- 14. TARGET: losses -----------------------------------------------------

val logits = ml.tensor([[2.0, 1.0, 0.1]]);
val target = ml.tensor([0.0]);
assert(ml.cross_entropy(logits, target).shape == [1]);
assert(ml.mse_loss(ml.tensor([[1.0, 2.0]]), ml.tensor([[0.0, 0.0]])).item() > 0.0);

# a loss is differentiable end to end
val lw = ml.tensor([[1.0], [2.0]], requires_grad=true);
val lx = ml.tensor([[1.0, 2.0]]);
val lloss = ml.mse_loss(lx @ lw, ml.tensor([[1.0]]));
lloss.backward();
assert(type(lw.grad) == "TENSOR");

# --- 15. TARGET: in-place functions -----------------------------------------

val ip = ml.tensor([[1.0, 2.0]]);
ml.add_(ip, ip);
assert(same(ip, ml.tensor([[2.0, 4.0]])));
ml.mul_(ip, 2.0);
assert(same(ip, ml.tensor([[4.0, 8.0]])));
ml.sub_(ip, 1.0);
assert(same(ip, ml.tensor([[3.0, 7.0]])));
ml.div_(ip, 1.0);
assert(same(ip, ml.tensor([[3.0, 7.0]])));

# --- 16. TARGET: nn and optim -----------------------------------------------

val lin = ml.nn.Linear(3, 2);
assert(lin.forward(ml.tensor([[1.0, 2.0, 3.0]])).shape == [1, 2]);
assert(ml.nn.parameters(lin).len() == 2); # weight and bias

val model = ml.nn.Sequential([ml.nn.Linear(2, 8), ml.nn.ReLU(), ml.nn.Linear(8, 1), ml.nn.Sigmoid()]);
val opt = ml.optim.SGD(ml.nn.parameters(model), lr=0.5);

# a tiny synthetic problem: the loss must drop over training
val X = ml.tensor([[0.0, 0.0], [1.0, 1.0], [1.0, 0.0], [0.0, 1.0]]);
val Y = ml.tensor([[0.0], [1.0], [1.0], [1.0]]);

var first_loss = 0.0;
var last_loss = 0.0;
for (i in 1..200) {
    opt.zero_grad();
    val pred = model.forward(X);
    val loss = ml.mse_loss(pred, Y);
    loss.backward();
    opt.step();
    if (i == 1) {
        first_loss = loss.item();
    }
    last_loss = loss.item();
}
assert(last_loss < first_loss);

# Adam is available too
val opt2 = ml.optim.Adam(ml.nn.parameters(ml.nn.Linear(2, 1)), lr=0.01);
opt2.zero_grad();

# --- 17. TARGET: save and load ----------------------------------------------

ml.save(a, "test_ml_tensor.bin");
assert(same(ml.load("test_ml_tensor.bin"), a));

# --- 18. canonical usage snippet (ML_PLAN.md Phase 12) ----------------------

val A = ml.tensor([[1.0, 2.0], [3.0, 4.0]]);
val B = ml.randn([2, 2], requires_grad=true);
val C = (A @ B).relu();
val L = C.sum();
L.backward();
assert(type(B.grad) == "TENSOR");

# --- 19. TARGET: graph compiler (ml.compile) --------------------------------

val cmodel = ml.nn.Sequential([ml.nn.Linear(4, 8), ml.nn.ReLU(), ml.nn.Linear(8, 3)]);
val cx = ml.tensor([[0.1, 0.2, 0.3, 0.4], [0.5, -0.6, 0.7, -0.8], [1.0, 0.9, -0.8, 0.7]]);
val eager_out = cmodel.forward(cx);

# compile captures the forward, fuses and prunes it, then replays it
val compiled = ml.compile(cmodel, cx);
assert(close(compiled.forward(cx), eager_out));
assert(type(compiled.stats()) == "STRING");

# replay runs real ops, so the tape still records and training works; this also
# checks that parameters stay rebound across optimizer steps
val cy = ml.tensor([0, 2, 1], datatype=ml.dtype.int32);
val copt = ml.optim.Adam(compiled.parameters(), lr=0.05);
var cfirst = 0.0;
var clast = 0.0;
for (i in 1..100) {
    copt.zero_grad();
    val closs = ml.cross_entropy(compiled.forward(cx), cy);
    closs.backward();
    copt.step();
    if (i == 1) { cfirst = closs.item(); }
    clast = closs.item();
}
assert(clast < cfirst);

# --- 20. TARGET: cat / gather --------------------------------------

val ca = ml.tensor([[1.0, 2.0, 3.0], [4.0, 5.0, 6.0]]); # [2, 3]
val cb = ml.tensor([[7.0, 8.0, 9.0]]);                    # [1, 3]

assert(ml.cat([ca, cb], 0).shape == [3, 3]);
assert(same(ml.cat([ca, cb], 0), ml.tensor([[1.0, 2.0, 3.0], [4.0, 5.0, 6.0], [7.0, 8.0, 9.0]])));

# cat is differentiable: each operand gets a gradient of ones
val cp = ml.tensor([[1.0, 2.0]], requires_grad=true);
val cq = ml.tensor([[3.0, 4.0]], requires_grad=true);
ml.sum(ml.cat([cp, cq], 1)).backward();
assert(same(cp.grad, ml.tensor([[1.0, 1.0]])));
assert(same(cq.grad, ml.tensor([[1.0, 1.0]])));

# gather selects along a dim, torch.gather style
val gidx = ml.tensor([[0, 2], [2, 0]], datatype=ml.dtype.int32);
assert(same(ml.gather(ca, 1, gidx), ml.tensor([[1.0, 3.0], [6.0, 4.0]])));

# gather scatters its gradient back to the selected positions
val gx = ml.tensor([[1.0, 2.0, 3.0]], requires_grad=true);
val gsel = ml.tensor([[2, 0]], datatype=ml.dtype.int32);
ml.sum(ml.gather(gx, 1, gsel)).backward();
assert(same(gx.grad, ml.tensor([[1.0, 0.0, 1.0]])));

# --- 21. TARGET: compiled model recompiles across shapes --------------------

val rmodel = ml.nn.Sequential([ml.nn.Linear(4, 8), ml.nn.ReLU(), ml.nn.Linear(8, 2)]);
val rx2 = ml.tensor([[0.1, 0.2, 0.3, 0.4], [0.5, -0.6, 0.7, -0.8]]);
val rcompiled = ml.compile(rmodel, rx2);
assert(rcompiled.forward(rx2).shape == [2, 2]);

# A different batch size recompiles rather than erroring, and matches eager.
val rx1 = ml.tensor([[0.1, 0.2, 0.3, 0.4]]);
assert(rcompiled.forward(rx1).shape == [1, 2]);
assert(close(rcompiled.forward(rx1), rmodel.forward(rx1)));

# Calling the same shape again is a cache hit (no growth in plan count).
val before = rcompiled.stats();
rcompiled.forward(rx1);
assert(rcompiled.stats() == before);

# Training through the recompiled model still works: parameters stay bound
# across optimizer steps even after a recompile.
val ry = ml.tensor([0, 1], datatype=ml.dtype.int32);
val ropt = ml.optim.Adam(rcompiled.parameters(), lr=0.05);
var rfirst = 0.0;
var rlast = 0.0;
for (i in 1..60) {
    ropt.zero_grad();
    val rloss = ml.cross_entropy(rcompiled.forward(rx2), ry);
    rloss.backward();
    ropt.step();
    if (i == 1) { rfirst = rloss.item(); }
    rlast = rloss.item();
}
assert(rlast < rfirst);

# --- 22. TARGET: compiled backward ------------------------------------------

# The compiled backward differentiates the captured forward graph, fusing the
# backward's own elementwise chains. With an explicit seed (the loss gradient
# w.r.t. the output) it trains the same objective as the eager backward.
val bwmodel = ml.nn.Sequential([ml.nn.Linear(4, 12), ml.nn.ReLU(), ml.nn.Linear(12, 1)]);
val bwX = ml.tensor([[0.1, 0.2, 0.3, 0.4], [0.5, -0.6, 0.7, -0.8], [1.0, 0.9, -0.8, 0.7]]);
val bwY = ml.full([3, 1], 0.0);
val bwCompiled = ml.compile(bwmodel, bwX);
val bwOpt = ml.optim.Adam(bwCompiled.parameters(), lr=0.05);
var bwFirst = 0.0;
var bwLast = 0.0;
for (i in 1..200) {
    val out = bwCompiled.forward(bwX);
    val diff = ml.sub(out, bwY);
    val loss = ml.mean(ml.mul(diff, diff));
    # d(mean squared error)/d(out) = 2*(out - y)/n
    val seed = ml.div(ml.mul(diff, 2.0), 3.0);
    bwOpt.zero_grad();
    bwCompiled.backward(bwX, seed);
    bwOpt.step();
    if (i == 1) { bwFirst = loss.item(); }
    bwLast = loss.item();
}
assert(bwLast < bwFirst);

# without a seed it differentiates the sum of the output
bwCompiled.backward(bwX);

# --- 23. TARGET: compiled cross-entropy training ----------------------------

# cross_entropy_grad is d(mean cross-entropy)/d(logits). Passing it as the
# compiled backward seed trains a classifier through the compiled backward, so
# the loss stays outside the captured graph while the backward is compiled.
val cemodel = ml.nn.Sequential([ml.nn.Linear(4, 12), ml.nn.ReLU(), ml.nn.Linear(12, 3)]);
val ceX = ml.tensor([[0.1, 0.2, 0.3, 0.4], [0.5, -0.6, 0.7, -0.8], [1.0, 0.9, -0.8, 0.7], [0.2, 0.3, -0.4, 0.5]]);
val ceY = ml.tensor([0, 2, 1, 0], datatype=ml.dtype.int32);
val ceCompiled = ml.compile(cemodel, ceX);
val ceOpt = ml.optim.Adam(ceCompiled.parameters(), lr=0.05);
var ceFirst = 0.0;
var ceLast = 0.0;
for (i in 1..200) {
    val logits = ceCompiled.forward(ceX);
    val loss = ml.cross_entropy(logits, ceY);
    val seed = ml.cross_entropy_grad(logits, ceY);
    ceOpt.zero_grad();
    ceCompiled.backward(ceX, seed);
    ceOpt.step();
    if (i == 1) { ceFirst = loss.item(); }
    ceLast = loss.item();
}
assert(ceLast < ceFirst);

# --- 24. TARGET: compile eager fallback -------------------------------------

# A forward that reads host data (here .item()) cannot be captured. ml.compile
# must fall back to eager execution rather than failing or mis-compiling.
var fbmodel = {};
fbmodel.forward = fun(x) { return ml.add(x, ml.mean(x).item()); };
fbmodel.parameters = fun() { return []; };
fbmodel.to = fun(dev) { return fbmodel; };
val fbx = ml.tensor([[1.0, 2.0, 3.0, 4.0]]);
val fbc = ml.compile(fbmodel, fbx);
assert(fbc.is_eager());
assert(same(fbc.forward(fbx), ml.add(fbx, 2.5)));
assert(fbc.stats() == "eager-fallback");

# A traceable model stays compiled (no fallback).
val okc = ml.compile(ml.nn.Sequential([ml.nn.Linear(4, 4), ml.nn.ReLU(), ml.nn.Linear(4, 2)]), fbx);
assert(!okc.is_eager());

# --- 25. TARGET: indexing ---------------------------------------------------

val ixs = ml.tensor([0.0, 1.0, 2.0, 3.0, 4.0, 5.0]);
assert(same(ml.slice(ixs, 0, 1, 6, 2), ml.tensor([1.0, 3.0, 5.0])));
assert(same(ml.slice(ixs, 0, 5, -1, -1), ml.tensor([5.0, 4.0, 3.0, 2.0, 1.0, 0.0])));
assert(same(ml.slice(ixs, 0, -2, 6), ml.tensor([4.0, 5.0])));
assert(same(ml.batch(ml.tensor([[1.0, 2.0], [3.0, 4.0], [5.0, 6.0]]), 1, 3),
    ml.tensor([[3.0, 4.0], [5.0, 6.0]])));

val ixm = ml.tensor([[1.0, 2.0, 3.0], [4.0, 5.0, 6.0]]);
assert(same(ml.select(ixm, 0, 1), ml.tensor([4.0, 5.0, 6.0])));
assert(same(ml.select(ixm, 1, -1), ml.tensor([3.0, 6.0])));
assert(same(ml.index_select(ixm, 0, ml.tensor([1, 0], datatype=ml.dtype.int32)),
    ml.tensor([[4.0, 5.0, 6.0], [1.0, 2.0, 3.0]])));
assert(same(ml.masked_fill(ixm, ml.tensor([[0.0, 1.0, 0.0], [1.0, 0.0, 1.0]]), -1.0),
    ml.tensor([[1.0, -1.0, 3.0], [-1.0, 5.0, -1.0]])));

# select is differentiable
val selx = ml.tensor([[1.0, 2.0], [3.0, 4.0]], requires_grad=true);
ml.sum(ml.select(selx, 0, 1)).backward();
assert(same(selx.grad, ml.tensor([[0.0, 0.0], [1.0, 1.0]])));

# --- 26. TARGET: embedding / bmm / silu -------------------------------------

val emb_w = ml.tensor([[1.0, 2.0], [3.0, 4.0], [5.0, 6.0]]);
assert(same(ml.embedding(emb_w, ml.tensor([2, 0], datatype=ml.dtype.int32)),
    ml.tensor([[5.0, 6.0], [1.0, 2.0]])));

val bm_a = ml.full([2, 2, 3], 1.0);
val bm_b = ml.full([2, 3, 2], 1.0);
assert(ml.bmm(bm_a, bm_b).shape == [2, 2, 2]);
assert(same(ml.bmm(bm_a, bm_b), ml.full([2, 2, 2], 3.0)));

assert(close(ml.silu(ml.tensor([0.0, 1.0])), ml.tensor([0.0, 0.7310586])));

# --- 27. TARGET: randperm / shuffle -----------------------------------------

ml.manual_seed(5);
val perm = ml.randperm(5);
assert(perm.shape == [5]);
# 0..4 sums to 10 and the squares sum to 30, which a permutation must satisfy
assert(ml.sum(perm).item() == 10.0);
assert(ml.sum(ml.mul(perm, perm)).item() == 30.0);

val sh = ml.shuffle(ml.tensor([[0.0, 1.0], [2.0, 3.0], [4.0, 5.0]]), 0);
assert(sh.shape == [3, 2]);

# --- 28. TARGET: losses and stats -------------------------------------------

val lg = ml.tensor([[1.0, 2.0, 3.0]]);
val lsm = ml.log_softmax(lg, 1);
assert(close(ml.exp(lsm), ml.softmax(lg, 1)));

val tgt = ml.tensor([2], datatype=ml.dtype.int32);
assert(close(ml.nll_loss(lsm, tgt), ml.neg(ml.mean(ml.select(lsm, 1, 2)))));

val vx = ml.tensor([[1.0, 2.0, 3.0], [4.0, 5.0, 6.0]]);
assert(close(ml.variance(vx, 1), ml.tensor([0.6666667, 0.6666667])));
assert(close(ml.std(vx, 1), ml.sqrt(ml.variance(vx, 1))));
assert(close(ml.norm(ml.tensor([[3.0, 4.0]]), 1), ml.tensor([5.0])));

# --- 29. TARGET: nn layers --------------------------------------------------

val emb = ml.nn.Embedding(10, 4);
assert(emb.forward(ml.tensor([1, 2], datatype=ml.dtype.int32)).shape == [2, 4]);
assert(ml.nn.parameters(emb).len() == 1);

val ln = ml.nn.LayerNorm(4);
assert(ln.forward(ml.tensor([[1.0, 2.0, 3.0, 4.0]])).shape == [1, 4]);
assert(ml.nn.parameters(ln).len() == 2);

val rn = ml.nn.RMSNorm(4);
assert(rn.forward(ml.tensor([[1.0, 2.0, 3.0, 4.0]])).shape == [1, 4]);
assert(ml.nn.parameters(rn).len() == 1);

val dr = ml.nn.Dropout(0.5);
dr.set_training(false);
assert(same(dr.forward(ml.tensor([1.0, 2.0])), ml.tensor([1.0, 2.0])));
dr.set_training(true);
assert(dr.forward(ml.tensor([1.0, 2.0])).shape == [2]);

val sq = ml.nn.Sequential([ml.nn.Linear(2, 4), ml.nn.SiLU(), ml.nn.Linear(4, 1)]);
assert(sq.forward(ml.tensor([[1.0, 2.0]])).shape == [1, 1]);

# --- 30. TARGET: custom model training with ml.parameters -------------------

# A model is a plain map of tensors plus a forward. ml.parameters finds the
# leaves, and the optimizer updates them in place, so no nn module is needed.
val cw1 = ml.randn([2, 16], requires_grad=true);
val cw2 = ml.randn([16, 1], requires_grad=true);
fun cforward(x) {
    return ml.matmul(ml.relu(ml.matmul(x, cw1)), cw2);
}
val cparams = ml.parameters({'w1': cw1, 'w2': cw2});
assert(cparams.len() == 2);
val cmopt = ml.optim.Adam(cparams, lr=0.05);
var cm0 = 0.0;
var cm1 = 0.0;
for (i in 1..300) {
    cmopt.zero_grad();
    val pred = cforward(X);
    val loss = ml.mse_loss(pred, Y);
    loss.backward();
    cmopt.step();
    if (i == 1) { cm0 = loss.item(); }
    cm1 = loss.item();
}
assert(cm1 < cm0);

# AdamW is Adam with decoupled weight decay
val wopt = ml.optim.AdamW(ml.parameters({'w': ml.randn([2, 2], requires_grad=true)}), lr=0.01);
wopt.zero_grad();

# --- 31. TARGET: grad clipping and LR schedules -----------------------------

val cgx = ml.tensor([[1.0, 2.0]], requires_grad=true);
val cgw = ml.tensor([[3.0, 4.0]]);
ml.sum(ml.mul(cgx, cgw)).backward();
val cgnorm = ml.clip_grad_norm_([cgx], 1.0);
assert(close(ml.tensor([cgnorm]), ml.tensor([5.0])));
assert(close(ml.norm(cgx.grad, 1), ml.tensor([1.0])));

assert(close(ml.tensor([ml.optim.step_lr(1.0, 0, 10, 0.1)]), ml.tensor([1.0])));
assert(close(ml.tensor([ml.optim.step_lr(1.0, 25, 10, 0.1)]), ml.tensor([0.01])));
assert(close(ml.tensor([ml.optim.warmup_lr(1.0, 0, 10)]), ml.tensor([0.1])));
assert(close(ml.tensor([ml.optim.cosine_lr(1.0, 0, 10)]), ml.tensor([1.0])));
assert(close(ml.tensor([ml.optim.cosine_lr(1.0, 10, 10)]), ml.tensor([0.0])));

# --- 32. TARGET: dtype honesty and clone ------------------------------------

val z32 = ml.zeros([2, 2], datatype=ml.dtype.int32);
assert(z32.dtype == "int32");
assert(z32.to_list()[0][0] == 0);

val o64 = ml.ones([2], datatype=ml.dtype.float64);
assert(o64.dtype == "float64");
assert(o64.to_list()[0] == 1.0);

# clone stays in the graph, detach does not
val clx = ml.tensor([[1.0, 2.0]], requires_grad=true);
ml.sum(ml.clone(clx)).backward();
assert(same(clx.grad, ml.tensor([[1.0, 1.0]])));

val dlx = ml.tensor([[1.0, 2.0]], requires_grad=true);
ml.sum(dlx.detach()).backward();
assert(dlx.grad == null);

# --- 33. TARGET: state_dict save/load ---------------------------------------

val sd_w = ml.randn([3, 4], requires_grad=true);
val sd_b = ml.zeros([4], requires_grad=true);
val sd_model = {'w': sd_w, 'b': sd_b};
val sd = ml.state_dict(sd_model);
assert(sd.len() == 2);
ml.save_state(sd_model, "test_ml_state.bin");

# a freshly initialized model loads the saved values in place
val sd_model2 = {'w': ml.randn([3, 4], requires_grad=true), 'b': ml.randn([4], requires_grad=true)};
ml.load_state(sd_model2, "test_ml_state.bin");
assert(same(sd_model2['w'], sd_model['w']));
assert(same(sd_model2['b'], sd_model['b']));

# --- 34. TARGET: flip / masked_select ---------------------------------------

val fx = ml.tensor([[1.0, 2.0, 3.0], [4.0, 5.0, 6.0]]);
assert(same(ml.flip(fx, [1]), ml.tensor([[3.0, 2.0, 1.0], [6.0, 5.0, 4.0]])));
assert(same(ml.masked_select(fx, ml.tensor([[1.0, 0.0, 1.0], [0.0, 1.0, 0.0]])),
    ml.tensor([1.0, 3.0, 5.0])));

# --- 35. TARGET: conv2d / maxpool2d -----------------------------------------

val cv = ml.nn.Conv2d(1, 2, 3, stride=1, padding=1);
assert(cv.forward(ml.randn([1, 1, 4, 4])).shape == [1, 2, 4, 4]);
assert(ml.parameters(cv).len() == 2); # weight and bias

val mp = ml.nn.MaxPool2d(2);
assert(mp.forward(ml.randn([1, 2, 4, 4])).shape == [1, 2, 2, 2]);
assert(ml.parameters(mp).len() == 0);

# --- 36. TARGET: max_dim, stats, tile, cumsum, sort, topk, nonzero ----------

val ox = ml.tensor([[1.0, 3.0, 2.0], [4.0, 0.0, 5.0]]);
val mdv = ml.max_dim(ox, 1);
assert(same(mdv[0], ml.tensor([3.0, 5.0])));
assert(same(mdv[1], ml.tensor([1, 2], datatype=ml.dtype.int32)));
val mnv = ml.min_dim(ox, 1);
assert(same(mnv[0], ml.tensor([1.0, 0.0])));
assert(same(mnv[1], ml.tensor([0, 1], datatype=ml.dtype.int32)));

assert(close(ml.variance(ox, 1), ml.tensor([0.6666667, 4.6666665])));
assert(close(ml.variance(ox, 1, unbiased=true), ml.tensor([1.0, 7.0])));
assert(close(ml.std(ox, 1, unbiased=true), ml.tensor([1.0, 2.6457512])));

assert(same(ml.tile(ml.tensor([[1.0, 2.0]]), [2, 2]),
    ml.tensor([[1.0, 2.0, 1.0, 2.0], [1.0, 2.0, 1.0, 2.0]])));
assert(same(ml.cumsum(ml.tensor([1.0, 2.0, 3.0]), 0), ml.tensor([1.0, 3.0, 6.0])));

val srt = ml.sort(ml.tensor([3.0, 1.0, 2.0]), 0, false);
assert(same(srt[0], ml.tensor([1.0, 2.0, 3.0])));
assert(same(srt[1], ml.tensor([1, 2, 0], datatype=ml.dtype.int32)));

val tk = ml.topk(ml.tensor([3.0, 1.0, 2.0]), 2, 0, true);
assert(same(tk[0], ml.tensor([3.0, 2.0])));
assert(same(tk[1], ml.tensor([0, 2], datatype=ml.dtype.int32)));

assert(same(ml.nonzero(ml.tensor([0.0, 2.0, 0.0, 3.0])),
    ml.tensor([[1], [3]], datatype=ml.dtype.int32)));
assert(same(ml.scatter_add(ml.zeros([3, 2]), 0,
        ml.tensor([[0, 1]], datatype=ml.dtype.int32), ml.ones([1, 2])),
    ml.tensor([[1.0, 0.0], [0.0, 1.0], [0.0, 0.0]])));

# --- 37. TARGET: cross_entropy options --------------------------------------

val cel = ml.tensor([[1.0, 2.0, 3.0]]);
val cet = ml.tensor([2], datatype=ml.dtype.int32);
assert(close(ml.cross_entropy(cel, cet), ml.tensor([0.40760595])));
assert(close(ml.cross_entropy(cel, cet, reduction='sum'), ml.tensor([0.40760595])));
assert(close(ml.cross_entropy(cel, cet, reduction='none'), ml.tensor([0.40760595])));
assert(close(ml.cross_entropy(cel, cet, weight=ml.tensor([1.0, 1.0, 2.0])),
    ml.tensor([0.8152119])));
val cei = ml.cross_entropy(ml.tensor([[1.0, 2.0, 3.0], [1.0, 2.0, 3.0]]),
    ml.tensor([2, 0], datatype=ml.dtype.int32), ignore_index=2);
assert(close(cei, ml.tensor([2.4076059])));

# --- 38. TARGET: autograd extras, optimizer state, views --------------------

# retain_grad keeps a non-leaf gradient
val rgx = ml.tensor([[2.0]], requires_grad=true);
val rgy = ml.mul(rgx, rgx);
ml.retain_grad(rgy);
ml.sum(rgy).backward();
assert(same(rgx.grad, ml.tensor([[4.0]])));
assert(same(rgy.grad, ml.tensor([[1.0]])));

# autograd_grad returns gradients without storing them
val aga = ml.tensor([[3.0]], requires_grad=true);
val agb = ml.tensor([[4.0]], requires_grad=true);
val ags = ml.autograd_grad(ml.sum(ml.mul(aga, agb)), [aga, agb]);
assert(same(ags[0], ml.tensor([[4.0]])));
assert(same(ags[1], ml.tensor([[3.0]])));
assert(aga.grad == null);

# optimizer state dict shares storage and round-trips through save/load
val ow = ml.randn([2, 2], requires_grad=true);
val oopt = ml.optim.Adam([ow], lr=0.01);
ml.sum(ml.mul(ow, ow)).backward();
oopt.step();
val osd = oopt.state_dict();
assert(osd.len() == 2); # m and v
ml.save_state(osd, "test_ml_optim.bin");
val om0 = ml.clone(osd['m.0']);
ml.sum(ml.mul(ow, ow)).backward();
oopt.step();
ml.load_state(osd, "test_ml_optim.bin");
assert(same(osd['m.0'], om0));

# storage-sharing views under no_grad
ml.no_grad(fun() {
    val vx = ml.tensor([[1.0, 2.0, 3.0], [4.0, 5.0, 6.0]]);
    val vt = ml.transpose(vx, 0, 1);
    assert(vt.shape == [3, 2]);
    assert(same(vt, ml.tensor([[1.0, 4.0], [2.0, 5.0], [3.0, 6.0]])));
    assert(same(ml.slice(vx, 1, 1, 3), ml.tensor([[2.0, 3.0], [5.0, 6.0]])));
    assert(same(ml.contiguous(vt), vt));
    assert(ml.view(ml.tensor([1.0, 2.0, 3.0, 4.0]), [2, 2]).shape == [2, 2]);
    return null;
});
