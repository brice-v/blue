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
## Bool tensors from `>`/`<` are still read with `ml.to_list` until there is a
## way to compare bool tensors directly.

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

# --- 9. TARGET: reductions --------------------------------------------------

val r = ml.tensor([[1.0, 2.0, 3.0], [4.0, 5.0, 6.0]]);
assert(same(r.sum(0), ml.tensor([5.0, 7.0, 9.0])));
assert(same(r.sum(1), ml.tensor([6.0, 15.0])));
assert(same(r.sum(1, keepdim=true), ml.tensor([[6.0], [15.0]])));
assert(same(r.mean(0), ml.tensor([2.5, 3.5, 4.5])));
assert(same(r.max(1), ml.tensor([3.0, 6.0])));
assert(same(r.min(1), ml.tensor([1.0, 4.0])));
assert(same(r.argmax(1), ml.tensor([2.0, 2.0])));
assert(same(r.argmin(1), ml.tensor([0.0, 0.0])));
assert(ml.sum(r).item() == 21.0);
assert(ml.mean(r).item() == 3.5);

# --- 10. TARGET: shape ops and remaining elementwise ------------------------

val s = ml.tensor([[1.0, 2.0, 3.0], [4.0, 5.0, 6.0]]);
assert(s.reshape([3, 2]).shape == [3, 2]);
assert(same(s.reshape([6]), ml.tensor([1.0, 2.0, 3.0, 4.0, 5.0, 6.0])));
assert(same(ml.transpose(s, 0, 1), s.T));
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
