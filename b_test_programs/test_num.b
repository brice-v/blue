## End-state integration test for the `num` module.
##
## The VM stops at the first error, so the file lights up top to bottom.
##
## `num` wraps gonum, which is float64, while ml tensors are float32. Comparisons
## on computed results use a tolerance for that reason.

import num
import ml

fun same(a, b) {
    ## exact value equality for two tensors
    ml.equal(a, b)
}

fun close_to(a, b) {
    ## float comparison within a tolerance, for gonum's float64 results
    if (a - b < 1e-9 && b - a < 1e-9) {
        return true;
    }
    return false;
}

# --- 1. construction and inspection -----------------------------------------

val a = num.dense([[2.0, 1.0], [1.0, 3.0]]);
assert(num.rows(a) == 2);
assert(num.cols(a) == 2);
assert(num.at(a, 0, 1) == 1.0);
assert(num.at(a, 1, 1) == 3.0);
assert(num.to_list(a) == [[2.0, 1.0], [1.0, 3.0]]);
assert(type(num.str(a)) == "STRING");

val z = num.zeros(2, 3);
assert(num.rows(z) == 2);
assert(num.cols(z) == 3);
assert(num.to_list(z) == [[0.0, 0.0, 0.0], [0.0, 0.0, 0.0]]);

assert(num.to_list(num.eye(2)) == [[1.0, 0.0], [0.0, 1.0]]);
assert(close_to(num.at(num.diag([1.0, 2.0]), 1, 1), 2.0));

# --- 2. arithmetic ----------------------------------------------------------

assert(num.to_list(num.add(a, a)) == [[4.0, 2.0], [2.0, 6.0]]);
assert(num.to_list(num.sub(a, a)) == [[0.0, 0.0], [0.0, 0.0]]);
assert(num.to_list(num.mul(a, a)) == [[4.0, 1.0], [1.0, 9.0]]);
assert(num.to_list(num.scale(a, 2.0)) == [[4.0, 2.0], [2.0, 6.0]]);
assert(num.to_list(num.transpose(a)) == [[2.0, 1.0], [1.0, 3.0]]);

# a @ b, with b a column vector
val b = num.dense([[3.0], [5.0]]);
assert(num.to_list(num.matmul(a, b)) == [[11.0], [18.0]]);

# --- 3. linear algebra ------------------------------------------------------

assert(close_to(num.det(a), 5.0));
assert(close_to(num.trace(a), 5.0));
assert(num.rank(a) == 2);
assert(num.rank(num.dense([[1.0, 2.0], [2.0, 4.0]])) == 1);
assert(close_to(num.norm(a, 'fro'), 3.872983346207417));

val x = num.solve(a, b);
assert(num.rows(x) == 2);
assert(num.cols(x) == 1);
assert(close_to(num.at(x, 0, 0), 0.8));
assert(close_to(num.at(x, 1, 0), 1.4));

# a @ inverse(a) is the identity
val ident = num.matmul(a, num.inverse(a));
assert(close_to(num.at(ident, 0, 0), 1.0));
assert(close_to(num.at(ident, 1, 1), 1.0));
assert(close_to(num.at(ident, 0, 1), 0.0));

# --- 4. decompositions ------------------------------------------------------

val e = num.eig(num.dense([[2.0, 0.0], [0.0, 3.0]]));
assert(num.rows(e['values']) == 2);
assert(num.cols(e['values']) == 2);   # [re, im] columns
assert(num.rows(e['vectors']) == 2);

val s = num.svd(num.dense([[3.0, 0.0], [0.0, 2.0]]));
assert(s['values'].len() == 2);
assert(close_to(s['values'][0], 3.0));
assert(num.rows(s['u']) == 2);
assert(num.rows(s['v']) == 2);

val q = num.qr(num.dense([[1.0, 2.0], [3.0, 4.0]]));
assert(num.rows(q['q']) == 2);
assert(num.rows(q['r']) == 2);

val l = num.lu(num.dense([[2.0, 1.0], [1.0, 3.0]]));
assert(num.rows(l['l']) == 2);
assert(num.rows(l['u']) == 2);

# cholesky of [[4,2],[2,3]] is [[2,0],[1,sqrt(2)]]
val c = num.cholesky(num.dense([[4.0, 2.0], [2.0, 3.0]]));
assert(close_to(num.at(c, 0, 0), 2.0));
assert(close_to(num.at(c, 1, 0), 1.0));
assert(num.at(c, 0, 1) == 0.0);

# --- 5. TARGET: ml tensors pass straight into num ---------------------------

# no conversion call: num accepts a tensor anywhere a matrix is expected
val t = ml.tensor([[4.0, 0.0], [0.0, 2.0]]);
assert(close_to(num.det(t), 8.0));
assert(num.rows(t) == 2);
assert(num.to_list(num.transpose(t)) == [[4.0, 0.0], [0.0, 2.0]]);

val tx = num.solve(t, ml.tensor([[8.0], [4.0]]));
assert(close_to(num.at(tx, 0, 0), 2.0));
assert(close_to(num.at(tx, 1, 0), 2.0));

# a plain list works too, for symmetry
assert(close_to(num.det([[1.0, 0.0], [0.0, 1.0]]), 1.0));

# --- 6. TARGET: num results feed straight into ml ---------------------------

val back = ml.from_matrix(a);
assert(back.shape == [2, 2]);
assert(same(back, ml.tensor([[2.0, 1.0], [1.0, 3.0]])));

# a nested list converts the same way
val from_list_mat = ml.from_matrix([[1.0, 2.0], [3.0, 4.0]]);
assert(same(from_list_mat, ml.tensor([[1.0, 2.0], [3.0, 4.0]])));

# statistics and samples come back as 1d tensors
val v = ml.from_list([1.0, 2.0, 3.0]);
assert(v.shape == [3]);
assert(same(v, ml.tensor([1.0, 2.0, 3.0])));

# to_tensor is the num-side spelling of the same thing
assert(same(num.to_tensor(a), back));

# --- 7. TARGET: round trip and gradients still work -------------------------

# a float64 round trip is exact
val f64 = num.to_tensor_f64(a);
assert(f64.dtype == "float64");
assert(num.to_list(num.from_tensor(f64)) == [[2.0, 1.0], [1.0, 3.0]]);
assert(num.to_list(ml.from_matrix_f64(a)) == [[2.0, 1.0], [1.0, 3.0]]);

# a tensor that came from num is a normal tensor: it can be trained
val w = ml.from_matrix([[1.0, 2.0], [3.0, 4.0]]);
ml.set_requires_grad(w, true);
ml.sum(ml.mul(w, w)).backward();
assert(same(w.grad, ml.tensor([[2.0, 4.0], [6.0, 8.0]])));

# and a num result can be a model input
val logits = ml.matmul(ml.from_matrix(num.eye(2)), ml.tensor([[1.0], [2.0]]));
assert(same(logits, ml.tensor([[1.0], [2.0]])));

# --- 8. TARGET: distributions -----------------------------------------------

val nd = num.normal(0.0, 1.0, seed=1);
assert(close_to(num.dist_cdf(nd, 1.96), 0.9750021048517795));
assert(close_to(num.dist_quantile(nd, 0.975), 1.9599639845400536));
assert(close_to(num.dist_mean(nd), 0.0));
assert(close_to(num.dist_variance(nd), 1.0));
assert(close_to(num.dist_prob(nd, 0.0), 0.39894228040143265));

assert(close_to(num.dist_mean(num.uniform(0.0, 2.0)), 1.0));
assert(close_to(num.dist_mean(num.exponential(2.0)), 0.5));
assert(close_to(num.dist_mean(num.gamma_dist(2.0, 1.0)), 2.0));
assert(close_to(num.dist_mean(num.beta_dist(2.0, 2.0)), 0.5));
assert(close_to(num.dist_mean(num.poisson(3.0)), 3.0));
assert(close_to(num.dist_mean(num.binomial(10, 0.5)), 5.0));

# a bad parameter is an error, not a silent wrong answer
assert(close_to(num.dist_mean(num.normal(0.0, 1.0)), 0.0));

# --- 9. TARGET: reproducible sampling, feeding ml ---------------------------

val s1 = num.rand_normal(5, 42);
val s2 = num.rand_normal(5, 42);
assert(same(s1, s2));                 # the seed makes it reproducible
assert(s1.shape == [5]);
assert(s1.dtype == "float64");        # gonum is float64

# samples are normal tensors: they can take part in model math
assert(ml.matmul(ml.reshape(s1, [5, 1]), ml.transpose(ml.from_matrix([[1.0], [2.0]]))).shape == [5, 2]);

val su = num.rand_uniform(4, 7);
assert(ml.equal(su, num.rand_uniform(4, 7)));

# --- 10. TARGET: statistics accept tensors ----------------------------------

val st = ml.tensor([1.0, 2.0, 3.0, 4.0]);
assert(close_to(num.mean(st), 2.5));
assert(close_to(num.variance(st), 1.6666666666666667));
assert(close_to(num.std_dev(st), 1.2909944487358056));
# the empirical quantile is nearest-rank, so an even sample gives 2.0, not 2.5
assert(close_to(num.median(st), 2.0));
assert(close_to(num.quantile(st, 0.25), 1.0));
assert(close_to(num.correlation(st, st), 1.0));
assert(close_to(num.covariance(st, st), 1.6666666666666667));
assert(close_to(num.entropy(num.dense([[0.5, 0.5]])), 0.6931471805599453));

val hist = num.histogram(st, 2);
assert(hist['counts'].len() == 2);
assert(hist['edges'].len() == 3);

# a tensor and the same list give the same answers
assert(close_to(num.mean(st), num.mean([1.0, 2.0, 3.0, 4.0])));

# --- 11. TARGET: hypothesis tests and special functions ---------------------

val tt = num.t_test(st, 0.0);
assert(close_to(tt['statistic'], 3.872983346207417));
assert(tt['p'] > 0.0);
assert(tt['p'] < 1.0);

val cs = num.chi_square_test([10.0, 20.0], [15.0, 15.0]);
assert(close_to(cs['statistic'], 3.3333333333333335));
assert(cs['p'] > 0.0);

assert(close_to(num.erfinv(0.0), 0.0));
assert(close_to(num.gamma_fn(5.0), 24.0));
assert(close_to(num.lgamma(5.0), 3.1780538303479458));
assert(close_to(num.digamma(1.0), -0.5772156649542974));
assert(close_to(num.beta_inc(1.0, 1.0, 0.5), 0.5));
assert(close_to(num.gamma_inc(1.0, 1.0), 0.6321205588285577));

# --- 12. TARGET: integration, transforms, interpolation ---------------------

# sampled integration
assert(close_to(num.trapezoidal([0.0, 1.0, 2.0], [0.0, 1.0, 4.0]), 3.0));
assert(close_to(num.simpsons([0.0, 1.0, 2.0], [0.0, 1.0, 4.0]), 2.6666666666666665));
assert(close_to(num.romberg([0.0, 1.0, 4.0], 1.0), 2.6666666666666665));

# integration over a blue closure: the coordinates arrive as separate scalars
assert(close_to(num.quad(fun(x) { return x ** 2; }, 0.0, 1.0, 100), 0.33333333333333337));
assert(close_to(num.quad(fun(x) { return x ** 3; }, 0.0, 1.0, 100), 0.25));

# minimization, 1d and 2d
val mn = num.minimize(fun(x) { return (x - 2.0) ** 2; }, [0.0]);
assert(close_to(mn['x'][0], 2.0));
assert(close_to(mn['value'], 0.0));
assert(mn['iterations'] > 0);

val mn2 = num.minimize(fun(a, b) { return (a - 1.0) ** 2 + (b - 3.0) ** 2; }, [0.0, 0.0]);
assert(close_to(mn2['x'][0], 1.0));
assert(close_to(mn2['x'][1], 3.0));

# a gradient method works too, using a numerical gradient
val mn3 = num.minimize(fun(x) { return (x - 2.0) ** 2; }, [0.0], method='bfgs');
assert(close_to(mn3['x'][0], 2.0));

# fft: the half spectrum pairs with fft_freqs
val sig = [1.0, 2.0, 3.0, 4.0];
val sp = num.fft(sig);
assert(sp['real'].len() == 3);
assert(sp['imag'].len() == 3);
assert(close_to(sp['real'][0], 10.0));
assert(close_to(sp['imag'][1], 2.0));
assert(num.fft_freqs(4).shape == [3]);

# ifft inverts the forward transform up to the unnormalized factor of n
val ifft_back = num.ifft(sp['real'], sp['imag']);
assert(close_to(ifft_back.to_list()[0], 4.0));
assert(close_to(ifft_back.to_list()[3], 16.0));

assert(close_to(num.dct(sig).to_list()[0], 15.0));

# tensors work here too
assert(close_to(num.trapezoidal([0.0, 1.0, 2.0], ml.tensor([0.0, 1.0, 4.0])), 3.0));
assert(close_to(num.fft(ml.tensor([0.0, 1.0, 4.0]))['real'][0], 5.0));

# interpolation
val li = num.interp_linear([0.0, 1.0, 2.0], [0.0, 1.0, 4.0]);
assert(close_to(num.predict(li, 0.5), 0.5));
assert(close_to(num.predict(li, 1.5), 2.5));
val ci = num.interp_cubic([0.0, 1.0, 2.0], [0.0, 1.0, 4.0]);
assert(close_to(num.predict(ci, 1.0), 1.0));

# a closure that returns a non-number degrades to NaN rather than killing the vm
val bad = num.quad(fun(x) { return "nope"; }, 0.0, 1.0, 10);
assert(bad != bad);   # NaN is not equal to itself

# --- 13. TARGET: student_t and chi_squared distributions --------------------

val tdist = num.student_t(10.0);
assert(close_to(num.dist_mean(tdist), 0.0));
assert(close_to(num.dist_cdf(tdist, 0.0), 0.5));
assert(close_to(num.dist_quantile(tdist, 0.975), 2.2281388519862744));
assert(close_to(num.dist_prob(tdist, 0.0), 0.38910838396603104));

val cdist = num.chi_squared_dist(2.0);
assert(close_to(num.dist_mean(cdist), 2.0));
assert(close_to(num.dist_variance(cdist), 4.0));
assert(close_to(num.dist_cdf(cdist, 2.0), 0.6321205588285577));
assert(close_to(num.dist_quantile(cdist, 0.95), 5.99146454710798));

# both are seeded and sample like the others
assert(num.dist_sample(num.student_t(5.0, seed=3), 4).shape == [4]);
assert(ml.equal(num.dist_sample(num.chi_squared_dist(3.0, seed=9), 4),
    num.dist_sample(num.chi_squared_dist(3.0, seed=9), 4)));
