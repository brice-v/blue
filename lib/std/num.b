## `num` is the module that contains numerical routines.
##
## It wraps the pure-Go gonum library: dense matrices, linear algebra,
## decompositions, distributions, statistics, and special functions. The
## matrices and statistics are gonum float64 values, while ml tensors are
## float32, so every crossing loses a little precision. Use `to_tensor_f64` and
## `from_matrix_f64` when an exact round trip matters.
##
## Every function that takes a matrix or a sample also takes an ml tensor, a num
## matrix, or a nested list, so `num.solve(ml_tensor, ml_b)` and
## `num.mean(ml_tensor)` work with no conversion call.
##
## A distribution is a handle with `dist_prob`, `dist_cdf`, `dist_quantile`,
## `dist_mean`, `dist_variance`, `dist_rand`, and `dist_sample`. Pass a `seed` to
## make sampling reproducible. Poisson and binomial have no closed-form quantile,
## so `dist_quantile` is not meaningful for them.

val __dense = _num_dense;
val __zeros = _num_zeros;
val __eye = _num_eye;
val __diag = _num_diag;
val __rows = _num_rows;
val __cols = _num_cols;
val __at = _num_at;
val __to_list = _num_to_list;
val __str = _num_str;
val __add = _num_add;
val __sub = _num_sub;
val __mul = _num_mul;
val __matmul = _num_matmul;
val __scale = _num_scale;
val __transpose = _num_transpose;
val __det = _num_det;
val __trace = _num_trace;
val __inverse = _num_inverse;
val __solve = _num_solve;
val __rank = _num_rank;
val __norm = _num_norm;
val __eig = _num_eig;
val __svd = _num_svd;
val __cholesky = _num_cholesky;
val __qr = _num_qr;
val __lu = _num_lu;
val __from_tensor = _num_from_tensor;
val __to_tensor = _num_to_tensor;
val __to_tensor_f64 = _num_to_tensor_f64;

val __normal = _num_normal;
val __student_t = _num_student_t;
val __chi_squared_dist = _num_chi_squared_dist;
val __uniform = _num_uniform;
val __exponential = _num_exponential;
val __gamma_dist = _num_gamma_dist;
val __beta_dist = _num_beta_dist;
val __poisson = _num_poisson;
val __binomial = _num_binomial;
val __dist_prob = _num_dist_prob;
val __dist_cdf = _num_dist_cdf;
val __dist_quantile = _num_dist_quantile;
val __dist_mean = _num_dist_mean;
val __dist_variance = _num_dist_variance;
val __dist_rand = _num_dist_rand;
val __dist_sample = _num_dist_sample;
val __rand_normal = _num_rand_normal;
val __rand_uniform = _num_rand_uniform;
val __mean = _num_mean;
val __variance = _num_variance;
val __std_dev = _num_std_dev;
val __median = _num_median;
val __quantile = _num_quantile;
val __correlation = _num_correlation;
val __covariance = _num_covariance;
val __entropy = _num_entropy;
val __histogram = _num_histogram;
val __t_test = _num_t_test;
val __chi_square_test = _num_chi_square_test;
val __erfinv = _num_erfinv;
val __gamma_fn = _num_gamma_fn;
val __lgamma = _num_lgamma;
val __digamma = _num_digamma;
val __beta_inc = _num_beta_inc;
val __gamma_inc = _num_gamma_inc;

val __quad = _num_quad;
val __minimize = _num_minimize;
val __trapezoidal = _num_trapezoidal;
val __simpsons = _num_simpsons;
val __romberg = _num_romberg;
val __fft = _num_fft;
val __ifft = _num_ifft;
val __dct = _num_dct;
val __fft_freqs = _num_fft_freqs;
val __interp_linear = _num_interp_linear;
val __interp_cubic = _num_interp_cubic;
val __predict = _num_predict;

fun dense(data) {
    ##std:this,__dense
    ## `dense` builds a matrix from a nested list, an ml tensor, or another
    ## matrix. A rank 1 tensor becomes a 1 by n row; a rank above 2 flattens the
    ## trailing dims into the columns.
    ##
    ## dense(data: list|tensor|matrix) -> matrix
    __dense(data)
}

fun zeros(rows, cols) {
    ##std:this,__zeros
    ## `zeros` returns a rows by cols matrix of zeros.
    ##
    ## zeros(rows: int, cols: int) -> matrix
    __zeros(rows, cols)
}

fun eye(n) {
    ##std:this,__eye
    ## `eye` returns an n by n identity matrix.
    ##
    ## eye(n: int) -> matrix
    __eye(n)
}

fun diag(v) {
    ##std:this,__diag
    ## `diag` returns a square matrix with `v` on its diagonal.
    ##
    ## diag(v: list|tensor) -> matrix
    __diag(v)
}

fun rows(m) {
    ##std:this,__rows
    ## `rows` returns the number of rows.
    ##
    ## rows(m: matrix|tensor|list) -> int
    __rows(m)
}

fun cols(m) {
    ##std:this,__cols
    ## `cols` returns the number of columns.
    ##
    ## cols(m: matrix|tensor|list) -> int
    __cols(m)
}

fun at(m, i, j) {
    ##std:this,__at
    ## `at` returns the element at row i, column j.
    ##
    ## at(m: matrix|tensor|list, i: int, j: int) -> float
    __at(m, i, j)
}

fun to_list(m) {
    ##std:this,__to_list
    ## `to_list` returns a matrix as a nested list of rows.
    ##
    ## to_list(m: matrix|tensor|list) -> list
    __to_list(m)
}

fun str(m) {
    ##std:this,__str
    ## `str` returns a readable string form of a matrix.
    ##
    ## str(m: matrix|tensor|list) -> str
    __str(m)
}

fun add(a, b) {
    ##std:this,__add
    ## `add` returns the elementwise sum of two matrices.
    ##
    ## add(a: matrix|tensor|list, b: matrix|tensor|list) -> matrix
    __add(a, b)
}

fun sub(a, b) {
    ##std:this,__sub
    ## `sub` returns the elementwise difference of two matrices.
    ##
    ## sub(a: matrix|tensor|list, b: matrix|tensor|list) -> matrix
    __sub(a, b)
}

fun mul(a, b) {
    ##std:this,__mul
    ## `mul` returns the elementwise product of two matrices.
    ##
    ## mul(a: matrix|tensor|list, b: matrix|tensor|list) -> matrix
    __mul(a, b)
}

fun matmul(a, b) {
    ##std:this,__matmul
    ## `matmul` returns the matrix product a @ b.
    ##
    ## matmul(a: matrix|tensor|list, b: matrix|tensor|list) -> matrix
    __matmul(a, b)
}

fun scale(m, factor) {
    ##std:this,__scale
    ## `scale` multiplies every element by a scalar.
    ##
    ## scale(m: matrix|tensor|list, factor: float) -> matrix
    __scale(m, factor)
}

fun transpose(m) {
    ##std:this,__transpose
    ## `transpose` returns the transpose of a matrix.
    ##
    ## transpose(m: matrix|tensor|list) -> matrix
    __transpose(m)
}

fun det(m) {
    ##std:this,__det
    ## `det` returns the determinant of a square matrix.
    ##
    ## det(m: matrix|tensor|list) -> float
    __det(m)
}

fun trace(m) {
    ##std:this,__trace
    ## `trace` returns the sum of the diagonal of a square matrix.
    ##
    ## trace(m: matrix|tensor|list) -> float
    __trace(m)
}

fun inverse(m) {
    ##std:this,__inverse
    ## `inverse` returns the inverse of a square matrix.
    ##
    ## inverse(m: matrix|tensor|list) -> matrix
    __inverse(m)
}

fun solve(a, b) {
    ##std:this,__solve
    ## `solve` returns x such that a @ x = b. This is the numerically stable way
    ## to solve a linear system; prefer it over `inverse(a)` times b.
    ##
    ## solve(a: matrix|tensor|list, b: matrix|tensor|list) -> matrix
    __solve(a, b)
}

fun rank(m) {
    ##std:this,__rank
    ## `rank` returns the rank of a matrix.
    ##
    ## rank(m: matrix|tensor|list) -> int
    __rank(m)
}

fun norm(m, kind='fro') {
    ##std:this,__norm
    ## `norm` returns a matrix norm: 'fro' (Frobenius), '1', 'inf', or '2'
    ## (spectral).
    ##
    ## norm(m: matrix|tensor|list, kind: str='fro') -> float
    __norm(m, kind)
}

fun eig(m) {
    ##std:this,__eig
    ## `eig` returns {values: matrix, vectors: matrix}. gonum returns complex
    ## eigenvalues even for real input, so `values` is an [n, 2] matrix of real
    ## and imaginary parts, column 0 and column 1.
    ##
    ## eig(m: matrix|tensor|list) -> map
    __eig(m)
}

fun svd(m) {
    ##std:this,__svd
    ## `svd` returns {u: matrix, values: list, v: matrix}.
    ##
    ## svd(m: matrix|tensor|list) -> map
    __svd(m)
}

fun cholesky(m) {
    ##std:this,__cholesky
    ## `cholesky` returns the lower triangular Cholesky factor of a symmetric
    ## positive definite matrix.
    ##
    ## cholesky(m: matrix|tensor|list) -> matrix
    __cholesky(m)
}

fun qr(m) {
    ##std:this,__qr
    ## `qr` returns {q: matrix, r: matrix}.
    ##
    ## qr(m: matrix|tensor|list) -> map
    __qr(m)
}

fun lu(m) {
    ##std:this,__lu
    ## `lu` returns {l: matrix, u: matrix}.
    ##
    ## lu(m: matrix|tensor|list) -> map
    __lu(m)
}

fun from_tensor(t) {
    ##std:this,__from_tensor
    ## `from_tensor` converts an ml tensor to a matrix. This is rarely needed:
    ## every num function accepts a tensor directly.
    ##
    ## from_tensor(t: tensor) -> matrix
    __from_tensor(t)
}

fun to_tensor(m) {
    ##std:this,__to_tensor
    ## `to_tensor` converts a matrix to an ml float32 tensor on the cpu.
    ## `ml.from_matrix` does the same thing from the ml side.
    ##
    ## to_tensor(m: matrix|list) -> tensor
    __to_tensor(m)
}

fun to_tensor_f64(m) {
    ##std:this,__to_tensor_f64
    ## `to_tensor_f64` converts a matrix to an ml float64 tensor, so a round
    ## trip through num is exact.
    ##
    ## to_tensor_f64(m: matrix|list) -> tensor
    __to_tensor_f64(m)
}

## --- distributions ----------------------------------------------------------
##
## A distribution is a handle with `dist_prob`, `dist_cdf`, `dist_quantile`,
## `dist_mean`, `dist_variance`, `dist_rand`, and `dist_sample`. A seed makes
## sampling reproducible. Poisson and binomial have no closed-form quantile, so
## `dist_quantile` returns NUll-ish garbage for them; use `dist_cdf` to search.

fun normal(mu=0.0, sigma=1.0, seed=null) {
    ##std:this,__normal
    ## `normal` returns a normal distribution.
    ##
    ## normal(mu: float=0.0, sigma: float=1.0, seed: int|null=null) -> distribution
    __normal(mu, sigma, seed)
}

fun uniform(min_val=0.0, max_val=1.0, seed=null) {
    ##std:this,__uniform
    ## `uniform` returns a uniform distribution over [min, max).
    ##
    ## uniform(min: float=0.0, max: float=1.0, seed: int|null=null) -> distribution
    __uniform(min_val, max_val, seed)
}

fun exponential(rate, seed=null) {
    ##std:this,__exponential
    ## `exponential` returns an exponential distribution with the given rate.
    ##
    ## exponential(rate: float, seed: int|null=null) -> distribution
    __exponential(rate, seed)
}

fun gamma_dist(alpha, beta, seed=null) {
    ##std:this,__gamma_dist
    ## `gamma_dist` returns a gamma distribution with shape alpha and rate beta.
    ## The name is `gamma_dist` because `gamma_fn` is the gamma function.
    ##
    ## gamma_dist(alpha: float, beta: float, seed: int|null=null) -> distribution
    __gamma_dist(alpha, beta, seed)
}

fun beta_dist(alpha, beta, seed=null) {
    ##std:this,__beta_dist
    ## `beta_dist` returns a beta distribution.
    ##
    ## beta_dist(alpha: float, beta: float, seed: int|null=null) -> distribution
    __beta_dist(alpha, beta, seed)
}

fun student_t(nu, mu=0.0, sigma=1.0, seed=null) {
    ##std:this,__student_t
    ## `student_t` returns a Student's t distribution with `nu` degrees of
    ## freedom. `sigma` is a scale, so the standard deviation is
    ## sigma*sqrt(nu/(nu-2)), which needs nu greater than 2.
    ##
    ## student_t(nu: float, mu: float=0.0, sigma: float=1.0, seed: int|null=null) -> distribution
    __student_t(nu, mu, sigma, seed)
}

fun chi_squared_dist(k, seed=null) {
    ##std:this,__chi_squared_dist
    ## `chi_squared_dist` returns a chi-squared distribution with `k` degrees of
    ## freedom. The name is suffixed so it does not clash with `chi_square_test`.
    ##
    ## chi_squared_dist(k: float, seed: int|null=null) -> distribution
    __chi_squared_dist(k, seed)
}

fun poisson(lambda, seed=null) {
    ##std:this,__poisson
    ## `poisson` returns a Poisson distribution.
    ##
    ## poisson(lambda: float, seed: int|null=null) -> distribution
    __poisson(lambda, seed)
}

fun binomial(n, p=0.5, seed=null) {
    ##std:this,__binomial
    ## `binomial` returns a binomial distribution over n trials.
    ##
    ## binomial(n: int, p: float=0.5, seed: int|null=null) -> distribution
    __binomial(n, p, seed)
}

fun dist_prob(d, x) {
    ##std:this,__dist_prob
    ## `dist_prob` is the probability density or mass at x.
    ##
    ## dist_prob(d: distribution, x: float) -> float
    __dist_prob(d, x)
}

fun dist_cdf(d, x) {
    ##std:this,__dist_cdf
    ## `dist_cdf` is the cumulative probability up to x.
    ##
    ## dist_cdf(d: distribution, x: float) -> float
    __dist_cdf(d, x)
}

fun dist_quantile(d, p) {
    ##std:this,__dist_quantile
    ## `dist_quantile` is the inverse CDF.
    ##
    ## dist_quantile(d: distribution, p: float) -> float
    __dist_quantile(d, p)
}

fun dist_mean(d) {
    ##std:this,__dist_mean
    ## `dist_mean` is the distribution mean.
    ##
    ## dist_mean(d: distribution) -> float
    __dist_mean(d)
}

fun dist_variance(d) {
    ##std:this,__dist_variance
    ## `dist_variance` is the distribution variance.
    ##
    ## dist_variance(d: distribution) -> float
    __dist_variance(d)
}

fun dist_rand(d) {
    ##std:this,__dist_rand
    ## `dist_rand` draws one sample.
    ##
    ## dist_rand(d: distribution) -> float
    __dist_rand(d)
}

fun dist_sample(d, n) {
    ##std:this,__dist_sample
    ## `dist_sample` draws n samples as a float64 1d tensor, ready for ml.
    ##
    ## dist_sample(d: distribution, n: int) -> tensor
    __dist_sample(d, n)
}

fun rand_normal(n, seed=null) {
    ##std:this,__rand_normal
    ## `rand_normal` draws n standard normal samples as a float64 tensor.
    ##
    ## rand_normal(n: int, seed: int|null=null) -> tensor
    __rand_normal(n, seed)
}

fun rand_uniform(n, seed=null) {
    ##std:this,__rand_uniform
    ## `rand_uniform` draws n uniform samples in [0, 1) as a float64 tensor.
    ##
    ## rand_uniform(n: int, seed: int|null=null) -> tensor
    __rand_uniform(n, seed)
}

## --- statistics -------------------------------------------------------------
##
## Every function here accepts a list or an ml tensor.

fun mean(x) {
    ##std:this,__mean
    ## `mean` is the arithmetic mean.
    ##
    ## mean(x: list|tensor) -> float
    __mean(x)
}

fun variance(x) {
    ##std:this,__variance
    ## `variance` is the unbiased sample variance, dividing by n-1.
    ##
    ## variance(x: list|tensor) -> float
    __variance(x)
}

fun std_dev(x) {
    ##std:this,__std_dev
    ## `std_dev` is the square root of the unbiased sample variance.
    ##
    ## std_dev(x: list|tensor) -> float
    __std_dev(x)
}

fun median(x) {
    ##std:this,__median
    ## `median` is the 0.5 empirical quantile. Note that the empirical quantile
    ## is nearest-rank, not interpolated, so an even-length sample returns one of
    ## the two middle values rather than their average.
    ##
    ## median(x: list|tensor) -> float
    __median(x)
}

fun quantile(x, p) {
    ##std:this,__quantile
    ## `quantile` is the empirical quantile at probability p. It is nearest-rank,
    ## not interpolated.
    ##
    ## quantile(x: list|tensor, p: float) -> float
    __quantile(x, p)
}

fun correlation(x, y) {
    ##std:this,__correlation
    ## `correlation` is the Pearson correlation coefficient.
    ##
    ## correlation(x: list|tensor, y: list|tensor) -> float
    __correlation(x, y)
}

fun covariance(x, y) {
    ##std:this,__covariance
    ## `covariance` is the unbiased sample covariance.
    ##
    ## covariance(x: list|tensor, y: list|tensor) -> float
    __covariance(x, y)
}

fun entropy(p) {
    ##std:this,__entropy
    ## `entropy` is the Shannon entropy of a probability distribution.
    ##
    ## entropy(p: list|tensor) -> float
    __entropy(p)
}

fun histogram(x, bins=10) {
    ##std:this,__histogram
    ## `histogram` bins values into equal-width buckets and returns
    ## {counts: list, edges: list}.
    ##
    ## histogram(x: list|tensor, bins: int=10) -> map
    __histogram(x, bins)
}

fun t_test(x, mu=0.0) {
    ##std:this,__t_test
    ## `t_test` is a one sample t-test of the mean against mu, returning
    ## {statistic, p}.
    ##
    ## t_test(x: list|tensor, mu: float=0.0) -> map
    __t_test(x, mu)
}

fun chi_square_test(observed, expected) {
    ##std:this,__chi_square_test
    ## `chi_square_test` is a goodness of fit test of observed counts against
    ## expected counts, returning {statistic, p}.
    ##
    ## chi_square_test(observed: list|tensor, expected: list|tensor) -> map
    __chi_square_test(observed, expected)
}

## --- special functions ------------------------------------------------------

fun erfinv(x) {
    ##std:this,__erfinv
    ## `erfinv` is the inverse error function.
    ##
    ## erfinv(x: float) -> float
    __erfinv(x)
}

fun gamma_fn(x) {
    ##std:this,__gamma_fn
    ## `gamma_fn` is the gamma function, named so it does not clash with the
    ## gamma distribution constructor.
    ##
    ## gamma_fn(x: float) -> float
    __gamma_fn(x)
}

fun lgamma(x) {
    ##std:this,__lgamma
    ## `lgamma` is the natural log of the absolute value of the gamma function.
    ##
    ## lgamma(x: float) -> float
    __lgamma(x)
}

fun digamma(x) {
    ##std:this,__digamma
    ## `digamma` is the logarithmic derivative of the gamma function.
    ##
    ## digamma(x: float) -> float
    __digamma(x)
}

fun beta_inc(a, b, x) {
    ##std:this,__beta_inc
    ## `beta_inc` is the regularized incomplete beta function.
    ##
    ## beta_inc(a: float, b: float, x: float) -> float
    __beta_inc(a, b, x)
}

fun gamma_inc(a, x) {
    ##std:this,__gamma_inc
    ## `gamma_inc` is the regularized lower incomplete gamma function.
    ##
    ## gamma_inc(a: float, x: float) -> float
    __gamma_inc(a, x)
}

## --- integration, transforms, and interpolation -----------------------------

fun quad(f, min_val, max_val, n=100) {
    ##std:this,__quad
    ## `quad` integrates a function over [min, max] with Gauss-Legendre
    ## quadrature at n points. It is the only integration form that needs a
    ## function rather than samples.
    ##
    ## quad(f: fun, min: float, max: float, n: int=100) -> float
    __quad(f, min_val, max_val, n)
}

fun minimize(f, x0, method='nelder-mead', max_iter=0) {
    ##std:this,__minimize
    ## `minimize` minimizes f from the starting point x0, returning
    ## {x: list, value: float, iterations: int}. Methods are 'nelder-mead'
    ## (default, no gradient), 'bfgs', and 'cg'. The gradient methods use a
    ## numerical gradient, since a blue function has no analytic one.
    ##
    ## minimize(f: fun, x0: list, method: str='nelder-mead', max_iter: int=0) -> map
    __minimize(f, x0, method, max_iter)
}

fun trapezoidal(x, y) {
    ##std:this,__trapezoidal
    ## `trapezoidal` integrates sampled values with the trapezoidal rule.
    ##
    ## trapezoidal(x: list|tensor, y: list|tensor) -> float
    __trapezoidal(x, y)
}

fun simpsons(x, y) {
    ##std:this,__simpsons
    ## `simpsons` integrates sampled values with Simpson's rule, which is more
    ## accurate than trapezoidal for smooth data.
    ##
    ## simpsons(x: list|tensor, y: list|tensor) -> float
    __simpsons(x, y)
}

fun romberg(y, dx) {
    ##std:this,__romberg
    ## `romberg` integrates evenly spaced samples with Romberg's method.
    ##
    ## romberg(y: list|tensor, dx: float) -> float
    __romberg(y, dx)
}

fun fft(x) {
    ##std:this,__fft
    ## `fft` is the discrete Fourier transform, returned as
    ## {real: list, imag: list} because blue has no complex type. Pair it with
    ## `fft_freqs` to get the frequencies.
    ##
    ## fft(x: list|tensor) -> map
    __fft(x)
}

fun ifft(real_part, imag_part) {
    ##std:this,__ifft
    ## `ifft` inverts `fft`, taking the real and imaginary parts.
    ##
    ## ifft(real: list|tensor, imag: list|tensor) -> tensor
    __ifft(real_part, imag_part)
}

fun dct(x) {
    ##std:this,__dct
    ## `dct` is the discrete cosine transform.
    ##
    ## dct(x: list|tensor) -> tensor
    __dct(x)
}

fun fft_freqs(n) {
    ##std:this,__fft_freqs
    ## `fft_freqs` returns the sample frequencies for a length n transform.
    ##
    ## fft_freqs(n: int) -> tensor
    __fft_freqs(n)
}

fun interp_linear(x, y) {
    ##std:this,__interp_linear
    ## `interp_linear` fits a piecewise linear interpolant. Call `predict` on the
    ## result to evaluate it.
    ##
    ## interp_linear(x: list|tensor, y: list|tensor) -> interpolant
    __interp_linear(x, y)
}

fun interp_cubic(x, y) {
    ##std:this,__interp_cubic
    ## `interp_cubic` fits a natural cubic spline.
    ##
    ## interp_cubic(x: list|tensor, y: list|tensor) -> interpolant
    __interp_cubic(x, y)
}

fun predict(f, x) {
    ##std:this,__predict
    ## `predict` evaluates an interpolant at x.
    ##
    ## predict(f: interpolant, x: float) -> float
    __predict(f, x)
}
