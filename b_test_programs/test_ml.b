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

# a scalar literal becomes a shape [1] tensor
assert(ml.tensor(5.0).shape == [1]);

# every optional parameter can be passed by keyword, in any order
assert(ml.tensor([[1.0]], datatype="float32", dev="cpu", requires_grad=false).shape == [1, 1]);

# --- 2. methods -------------------------------------------------------------

val c = a.matmul(b);                                                       # 2x2
assert(c.shape == [2, 2]);
assert(same(c, ml.tensor([[58.0, 64.0], [139.0, 154.0]])));

assert(same(c.add(c), ml.tensor([[116.0, 128.0], [278.0, 308.0]])));
assert(same(c.sub(c), ml.tensor([[0.0, 0.0], [0.0, 0.0]])));
assert(same(c.mul(c), ml.tensor([[3364.0, 4096.0], [19321.0, 23716.0]])));
assert(same(c.div(c), ml.tensor([[1.0, 1.0], [1.0, 1.0]])));

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
val y = x.matmul(w); # [[6.0]]

assert(y.item() == 6.0);
assert(w.grad == null);

y.backward();
assert(w.grad.item() == 3.0); # dy/dw = x = 3
assert(x.grad == null);       # x is not tracked

w.zero_grad();
assert(w.grad == null);

# --- 5. module-level functions ----------------------------------------------

val e = c.add(c);
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
assert(to_list(c == a.matmul(b)) == [[true, true], [true, true]]);
assert(to_list(c > 0.0) == [[true, true], [true, true]]);
assert(to_list(c < 0.0) == [[false, false], [false, false]]);
assert(to_list(c != c) == [[false, false], [false, false]]);

# the module functions and methods match the operators, and take scalars
assert(same(c > 0.0, ml.gt(c, 0.0)));
assert(same(c >= 0.0, ml.ge(c, 0.0)));
assert(same(c < 0.0, ml.lt(c, 0.0)));
assert(same(c <= 0.0, ml.le(c, 0.0)));
assert(same(c != c, ml.ne(c, c)));
assert(same(c == a.matmul(b), ml.eq(c, a.matmul(b))));
assert(same(c.gt(0.0), ml.gt(c, 0.0)));
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
assert(same(u.pow(ml.tensor([[2.0, 2.0, 2.0]])), ml.tensor([[1.0, 0.0, 16.0]])));
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
val ng = ml.no_grad(fun() { ml.tensor([[1.0]], requires_grad=true).mul(ml.tensor([[2.0]])) });
assert(ng.requires_grad == false);

# --- 14. TARGET: losses -----------------------------------------------------

val logits = ml.tensor([[2.0, 1.0, 0.1]]);
val target = ml.tensor([0.0]);
assert(ml.cross_entropy(logits, target).shape == [1]);
assert(ml.mse_loss(ml.tensor([[1.0, 2.0]]), ml.tensor([[0.0, 0.0]])).item() > 0.0);

# a loss is differentiable end to end
val lw = ml.tensor([[1.0], [2.0]], requires_grad=true);
val lx = ml.tensor([[1.0, 2.0]]);
val lloss = ml.mse_loss(lx.matmul(lw), ml.tensor([[1.0]]));
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
val C = A.matmul(B).relu();
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
