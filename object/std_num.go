package object

import (
	"fmt"
	"math"
	rand "math/rand/v2"
	"strings"

	"blue/ml"
	"gonum.org/v1/gonum/dsp/fourier"
	"gonum.org/v1/gonum/integrate"
	"gonum.org/v1/gonum/interp"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/mathext"
	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"
)

// NumMatrix is the blue-facing handle for a gonum dense matrix. gonum has no
// matrix object type of its own, so blue wraps one in a GoObj, the same way ml
// wraps a module.
type NumMatrix = mat.Dense

// numMatrixArg converts a blue value into a *mat.Dense. It accepts, in this
// order:
//
//   - a num matrix handle, returned as is
//   - an ml tensor, converted through its row-major data (see tensorToMatrix)
//   - a nested list, whose shape is inferred like ml.tensor
//
// This is what makes passing tensors into num trivial: every num function that
// wants a matrix calls this, so `num.solve(t, b)` works when t and b are ml
// tensors, gonum matrices, or plain lists.
func numMatrixArg(name string, pos int, o Object) (*mat.Dense, Object) {
	switch v := o.(type) {
	case *GoObj[*NumMatrix]:
		if v.Value == nil {
			return nil, newError("`%s` error: argument %d is a nil matrix", name, pos)
		}
		return v.Value, nil
	case *Tensor:
		m, err := tensorToMatrix(v.T)
		if err != nil {
			return nil, newError("`%s` error: %s", name, err.Error())
		}
		return m, nil
	case *List:
		m, err := listToMatrix(v)
		if err != nil {
			return nil, newError("`%s` error: %s", name, err.Error())
		}
		return m, nil
	default:
		return nil, newPositionalTypeError(name, pos, "MATRIX, TENSOR, or LIST", o.Type())
	}
}

// numVectorArg converts a blue value into a []float64. It accepts a num matrix
// handle when the matrix is a single row or column, an ml tensor of any rank
// (read in row-major order), or a list of numbers.
func numVectorArg(name string, pos int, o Object) ([]float64, Object) {
	switch v := o.(type) {
	case *GoObj[*NumMatrix]:
		if v.Value == nil {
			return nil, newError("`%s` error: argument %d is a nil matrix", name, pos)
		}
		r, c := v.Value.Dims()
		if r != 1 && c != 1 {
			return nil, newError("`%s` error: argument %d is a %dx%d matrix, want a vector", name, pos, r, c)
		}
		out := make([]float64, r*c)
		for i := range out {
			if c == 1 {
				out[i] = v.Value.At(i, 0)
			} else {
				out[i] = v.Value.At(0, i)
			}
		}
		return out, nil
	case *Tensor:
		return tensorToFloat64(v.T), nil
	case *List:
		out := make([]float64, len(v.Elements))
		for i, e := range v.Elements {
			f, ok := objectToFloat64(e)
			if !ok {
				return nil, newPositionalTypeError(name, pos, "LIST of numbers", e.Type())
			}
			out[i] = f
		}
		return out, nil
	default:
		return nil, newPositionalTypeError(name, pos, "VECTOR, TENSOR, or LIST", o.Type())
	}
}

// numSeriesArg reads a series argument as []float64. It accepts a list, an ml
// tensor, or a num matrix, so plotting a tensor needs no conversion call.
func numSeriesArg(name string, pos int, o Object) ([]float64, Object) {
	return numVectorArg(name, pos, o)
}

func objectToFloat64(o Object) (float64, bool) {
	switch v := o.(type) {
	case *Float:
		return v.Value, true
	case *Integer:
		return float64(v.Value), true
	case *UInteger:
		return float64(v.Value), true
	case *Boolean:
		if v.Value {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

// tensorToFloat64 reads an ml tensor's elements in row-major order as float64.
// ContiguousData already handles dtype and strided views, so this is the one
// conversion point for the tensor to numerics direction.
func tensorToFloat64(t *ml.Tensor) []float64 {
	data := t.ContiguousData()
	out := make([]float64, len(data))
	for i, v := range data {
		out[i] = float64(v)
	}
	return out
}

// tensorToMatrix converts an ml tensor into a *mat.Dense.
//
// Shape rules, chosen to match what gonum solvers expect:
//   - rank 2: direct, rows and cols as they are
//   - rank 1 of length n: a 1 x n row
//   - rank 0 (a single element): a 1 x 1 matrix
//   - rank > 2: the trailing dims are flattened into the columns, so a
//     [a, b, c, d] tensor becomes [a, b*c*d]
func tensorToMatrix(t *ml.Tensor) (*mat.Dense, error) {
	shape := t.Shape()
	data := tensorToFloat64(t)
	switch len(shape) {
	case 0:
		return mat.NewDense(1, 1, data), nil
	case 1:
		return mat.NewDense(1, shape[0], data), nil
	case 2:
		return mat.NewDense(shape[0], shape[1], data), nil
	default:
		rows := shape[0]
		cols := 1
		for _, d := range shape[1:] {
			cols *= d
		}
		return mat.NewDense(rows, cols, data), nil
	}
}

// listToMatrix builds a matrix from a nested list, inferring the shape the same
// way ml.tensor does.
func listToMatrix(l *List) (*mat.Dense, error) {
	rows := len(l.Elements)
	if rows == 0 {
		return mat.NewDense(0, 0, nil), nil
	}
	cols := -1
	data := make([]float64, 0, rows)
	for _, rowObj := range l.Elements {
		switch row := rowObj.(type) {
		case *List:
			if cols == -1 {
				cols = len(row.Elements)
			} else if len(row.Elements) != cols {
				return nil, fmt.Errorf("ragged list: row has %d entries, expected %d", len(row.Elements), cols)
			}
			for _, e := range row.Elements {
				f, ok := objectToFloat64(e)
				if !ok {
					return nil, fmt.Errorf("expected a number in the list, found %s", e.Type())
				}
				data = append(data, f)
			}
		default:
			// A flat list is a single row.
			if cols == -1 {
				cols = rows
			}
			f, ok := objectToFloat64(rowObj)
			if !ok {
				return nil, fmt.Errorf("expected a number in the list, found %s", rowObj.Type())
			}
			data = append(data, f)
		}
	}
	if cols == -1 {
		cols = 0
	}
	if len(data) != rows*cols {
		return nil, fmt.Errorf("list has %d entries, shape [%d, %d] needs %d", len(data), rows, cols, rows*cols)
	}
	return mat.NewDense(rows, cols, data), nil
}

// matrixToFloat32Data returns a matrix's elements in row-major float32, for
// handing back to ml.
func matrixToFloat32Data(m *mat.Dense) []float32 {
	r, c := m.Dims()
	out := make([]float32, r*c)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			out[i*c+j] = float32(m.At(i, j))
		}
	}
	return out
}

// matrixToFloat64Data returns a matrix's elements in row-major float64.
func matrixToFloat64Data(m *mat.Dense) []float64 {
	r, c := m.Dims()
	out := make([]float64, r*c)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			out[i*c+j] = m.At(i, j)
		}
	}
	return out
}

// float64ObjectList converts a slice of float64 into a blue list of floats.
func float64ObjectList(vs []float64) Object {
	elems := make([]Object, len(vs))
	for i, v := range vs {
		elems[i] = &Float{Value: v}
	}
	return &List{Elements: elems}
}

// float64ObjectMatrix converts a matrix into a nested blue list, one inner list
// per row. This is the num.to_list path.
func float64ObjectMatrix(m *mat.Dense) Object {
	r, c := m.Dims()
	rows := make([]Object, r)
	for i := 0; i < r; i++ {
		row := make([]Object, c)
		for j := 0; j < c; j++ {
			row[j] = &Float{Value: m.At(i, j)}
		}
		rows[i] = &List{Elements: row}
	}
	return &List{Elements: rows}
}

// The matrix kernels used by the num builtins. Keeping them here rather than
// inline in std_num.go keeps the builtin table readable.

func denseAdd(a, b *mat.Dense, sign float64) (*mat.Dense, error) {
	ar, ac := a.Dims()
	br, bc := b.Dims()
	if ar != br || ac != bc {
		return nil, fmt.Errorf("shape mismatch: %dx%d vs %dx%d", ar, ac, br, bc)
	}
	out := mat.NewDense(ar, ac, nil)
	for i := 0; i < ar; i++ {
		for j := 0; j < ac; j++ {
			out.Set(i, j, a.At(i, j)+sign*b.At(i, j))
		}
	}
	return out, nil
}

func denseElemMul(a, b *mat.Dense) (*mat.Dense, error) {
	ar, ac := a.Dims()
	br, bc := b.Dims()
	if ar != br || ac != bc {
		return nil, fmt.Errorf("shape mismatch: %dx%d vs %dx%d", ar, ac, br, bc)
	}
	out := mat.NewDense(ar, ac, nil)
	for i := 0; i < ar; i++ {
		for j := 0; j < ac; j++ {
			out.Set(i, j, a.At(i, j)*b.At(i, j))
		}
	}
	return out, nil
}

func denseMatMul(a, b *mat.Dense) (*mat.Dense, error) {
	ar, ac := a.Dims()
	br, bc := b.Dims()
	if ac != br {
		return nil, fmt.Errorf("inner dimension mismatch: %d vs %d", ac, br)
	}
	out := mat.NewDense(ar, bc, nil)
	out.Mul(a, b)
	return out, nil
}

func denseScale(m *mat.Dense, f float64) *mat.Dense {
	r, c := m.Dims()
	out := mat.NewDense(r, c, nil)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			out.Set(i, j, m.At(i, j)*f)
		}
	}
	return out
}

func denseTranspose(m *mat.Dense) (*mat.Dense, error) {
	r, c := m.Dims()
	out := mat.NewDense(c, r, nil)
	out.Copy(m.T())
	return out, nil
}

func denseDet(m *mat.Dense) (float64, error) {
	r, c := m.Dims()
	if r != c {
		return 0, fmt.Errorf("matrix is %dx%d, want square", r, c)
	}
	var lu mat.LU
	lu.Factorize(m)
	det := lu.Det()
	if det != det {
		// gonum returns NaN from a singular factorization.
		return 0, fmt.Errorf("matrix is singular")
	}
	return det, nil
}

func denseInverse(m *mat.Dense) (*mat.Dense, error) {
	r, c := m.Dims()
	if r != c {
		return nil, fmt.Errorf("matrix is %dx%d, want square", r, c)
	}
	out := mat.NewDense(r, c, nil)
	if err := out.Inverse(m); err != nil {
		return nil, err
	}
	return out, nil
}

func denseSolve(a, b *mat.Dense) (*mat.Dense, error) {
	ar, ac := a.Dims()
	if ar != ac {
		return nil, fmt.Errorf("matrix is %dx%d, want square", ar, ac)
	}
	br, bc := b.Dims()
	if br != ar {
		return nil, fmt.Errorf("b has %d rows, a needs %d", br, ar)
	}
	out := mat.NewDense(ac, bc, nil)
	if err := out.Solve(a, b); err != nil {
		return nil, err
	}
	return out, nil
}

// denseCholesky returns the lower triangular Cholesky factor. gonum's Cholesky
// factorizes a Symmetric, which a *mat.Dense is not, so the matrix is wrapped
// first.
func denseCholesky(m *mat.Dense) (*mat.Dense, error) {
	r, c := m.Dims()
	if r != c {
		return nil, fmt.Errorf("matrix is %dx%d, want square", r, c)
	}
	sym := mat.NewSymDense(r, nil)
	// Copy the dense values into the symmetric wrapper by hand: gonum's CopySym
	// wants a Symmetric, which a *mat.Dense is not.
	for i := 0; i < r; i++ {
		for j := 0; j <= i; j++ {
			sym.SetSym(i, j, m.At(i, j))
		}
	}
	var col mat.Cholesky
	if !col.Factorize(sym) {
		return nil, fmt.Errorf("matrix is not positive definite")
	}
	var lower mat.TriDense
	col.LTo(&lower)
	out := mat.NewDense(r, c, nil)
	for i := 0; i < r; i++ {
		for j := 0; j <= i; j++ {
			out.Set(i, j, lower.At(i, j))
		}
	}
	return out, nil
}

// denseEigObject runs the eigen decomposition. gonum returns complex values even
// for real input, so the eigenvalues are returned as an [n, 2] matrix of real
// and imaginary parts (columns 0 and 1), which keeps them inside the float64
// matrix type instead of needing a complex object type.
func denseEigObject(m *mat.Dense) Object {
	r, c := m.Dims()
	if r != c {
		return newError("`eig` error: matrix is %dx%d, want square", r, c)
	}
	var eig mat.Eigen
	if !eig.Factorize(m, mat.EigenRight) {
		return newError("`eig` error: factorization did not converge")
	}
	vals := eig.Values(nil)
	v := mat.NewDense(r, 2, nil)
	for i := 0; i < r; i++ {
		v.Set(i, 0, real(vals[i]))
		v.Set(i, 1, imag(vals[i]))
	}
	var vectors mat.CDense
	eig.VectorsTo(&vectors)
	// Return the eigenvectors as two stacked real matrices, since blue matrices
	// are float64: `vectors_real` and `vectors_imag`.
	vr := mat.NewDense(r, r, nil)
	vi := mat.NewDense(r, r, nil)
	for i := 0; i < r; i++ {
		for j := 0; j < r; j++ {
			z := vectors.At(i, j)
			vr.Set(i, j, real(z))
			vi.Set(i, j, imag(z))
		}
	}
	return numMapObject(
		"values", numMatrixObject(v),
		"vectors", numMatrixObject(vr),
		"vectors_imag", numMatrixObject(vi),
	)
}

func denseSVDObject(m *mat.Dense) Object {
	var svd mat.SVD
	if !svd.Factorize(m, mat.SVDThin) {
		return newError("`svd` error: factorization did not converge")
	}
	var u, v mat.Dense
	svd.UTo(&u)
	svd.VTo(&v)
	values := svd.Values(nil)
	return numMapObject(
		"u", numMatrixObject(&u),
		"values", float64ObjectList(values),
		"v", numMatrixObject(&v),
	)
}

func denseQRObject(m *mat.Dense) Object {
	var qr mat.QR
	qr.Factorize(m)
	var q, r mat.Dense
	qr.QTo(&q)
	qr.RTo(&r)
	return numMapObject(
		"q", numMatrixObject(&q),
		"r", numMatrixObject(&r),
	)
}

// denseLUObject returns L and U. gonum's LTo and UTo fill TriDense, so the
// factors are copied into dense matrices, which is the only matrix type blue
// hands back.
func denseLUObject(m *mat.Dense) Object {
	r, c := m.Dims()
	if r != c {
		return newError("`lu` error: matrix is %dx%d, want square", r, c)
	}
	var lu mat.LU
	lu.Factorize(m)
	var lt, ut mat.TriDense
	lu.LTo(&lt)
	lu.UTo(&ut)
	l := mat.NewDense(r, c, nil)
	u := mat.NewDense(r, c, nil)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			l.Set(i, j, lt.At(i, j))
			u.Set(i, j, ut.At(i, j))
		}
	}
	return numMapObject(
		"l", numMatrixObject(l),
		"u", numMatrixObject(u),
	)
}

// mlMatrixTensor builds an ml tensor from a gonum matrix, for the conversion
// helpers that hand a num result back to ml.
func mlMatrixTensor(m *mat.Dense, dtype ml.DType, device ml.Device) (*ml.Tensor, error) {
	r, c := m.Dims()
	return ml.NewTensor(matrixToFloat32Data(m), []int{r, c}, dtype, device)
}

// NumDist is the blue-facing handle for a univariate distribution. gonum's
// distributions are values with no common interface beyond the methods they
// share, so blue stores one behind a small wrapper that names the operations the
// builtins need.
type NumDist struct {
	Name string

	// Prob is the probability density or mass function.
	Prob func(x float64) float64
	// CDF is the cumulative distribution function.
	CDF func(x float64) float64
	// Quantile is the inverse CDF.
	Quantile func(p float64) float64
	// Mean and Variance are the first two moments.
	Mean     func() float64
	Variance func() float64
	// Rand draws one sample, and Sample draws n.
	Rand   func() float64
	Sample func(n int) []float64
}

// newNumDist builds a distribution handle from gonum's method set. gonum's
// distributions share these method names by convention rather than by
// interface, so the wrappers are explicit per distribution.
func newNumDist(name string, src rand.Source,
	prob, cdf, quantile func(float64) float64,
	mean, variance func() float64,
	randFn func() float64) *NumDist {
	return &NumDist{
		Name:     name,
		Prob:     prob,
		CDF:      cdf,
		Quantile: quantile,
		Mean:     mean,
		Variance: variance,
		Rand:     randFn,
		Sample: func(n int) []float64 {
			out := make([]float64, n)
			for i := range out {
				out[i] = randFn()
			}
			return out
		},
	}
}

// numDistArg accepts a distribution handle.
func numDistArg(name string, pos int, o Object) (*NumDist, Object) {
	g, ok := o.(*GoObj[*NumDist])
	if !ok || g.Value == nil {
		return nil, newPositionalTypeError(name, pos, "DISTRIBUTION", o.Type())
	}
	return g.Value, nil
}

// numFloatArg accepts a number, so distribution queries take an int or a float.
func numFloatArg(name string, pos int, o Object) (float64, Object) {
	f, ok := objectToFloat64(o)
	if !ok {
		return 0, newPositionalTypeError(name, pos, "number", o.Type())
	}
	return f, nil
}

// numDistScalarBuiltin builds a distribution method taking one number.
func numDistScalarBuiltin(name string, f func(d *NumDist, x float64) float64) func(...Object) Object {
	return func(args ...Object) Object {
		if err := checkArgCount(name, 2, args); err != nil {
			return err
		}
		d, errObj := numDistArg(name, 1, args[0])
		if errObj != nil {
			return errObj
		}
		x, errObj := numFloatArg(name, 2, args[1])
		if errObj != nil {
			return errObj
		}
		return &Float{Value: f(d, x)}
	}
}

// numDistMomentBuiltin builds a distribution method taking no argument.
func numDistMomentBuiltin(name string, f func(d *NumDist) float64) func(...Object) Object {
	return func(args ...Object) Object {
		if err := checkArgCount(name, 1, args); err != nil {
			return err
		}
		d, errObj := numDistArg(name, 1, args[0])
		if errObj != nil {
			return errObj
		}
		return &Float{Value: f(d)}
	}
}

// numSourceArg turns an optional seed into a rand.Source. A nil object means "use
// the global source", so unseeded sampling still varies between runs.
func numSourceArg(name string, o Object) (rand.Source, Object) {
	if o == nil {
		return nil, nil
	}
	if _, isNull := o.(*Null); isNull {
		return nil, nil
	}
	n, ok := o.(*Integer)
	if !ok {
		return nil, newPositionalTypeError(name, 2, "INTEGER or NULL", o.Type())
	}
	return rand.NewPCG(uint64(n.Value), 0), nil
}

// makeNormal builds a normal distribution, using the global source when no seed
// is given.
func makeNormal(mu, sigma float64, src rand.Source) (*NumDist, error) {
	if sigma <= 0 {
		return nil, fmt.Errorf("sigma must be positive, got %g", sigma)
	}
	if src == nil {
		src = rand.NewPCG(rand.Uint64(), rand.Uint64())
	}
	d := distuv.Normal{Mu: mu, Sigma: sigma, Src: src}
	return newNumDist("normal", src, d.Prob, d.CDF, d.Quantile, d.Mean, d.Variance, d.Rand), nil
}

func makeUniform(lo, hi float64, src rand.Source) (*NumDist, error) {
	if hi <= lo {
		return nil, fmt.Errorf("max must be greater than min, got %g and %g", lo, hi)
	}
	if src == nil {
		src = rand.NewPCG(rand.Uint64(), rand.Uint64())
	}
	d := distuv.Uniform{Min: lo, Max: hi, Src: src}
	return newNumDist("uniform", src, d.Prob, d.CDF, d.Quantile, d.Mean, d.Variance, d.Rand), nil
}

func makeExponential(rate float64, src rand.Source) (*NumDist, error) {
	if rate <= 0 {
		return nil, fmt.Errorf("rate must be positive, got %g", rate)
	}
	if src == nil {
		src = rand.NewPCG(rand.Uint64(), rand.Uint64())
	}
	d := distuv.Exponential{Rate: rate, Src: src}
	return newNumDist("exponential", src, d.Prob, d.CDF, d.Quantile, d.Mean, d.Variance, d.Rand), nil
}

func makeGamma(alpha, beta float64, src rand.Source) (*NumDist, error) {
	if alpha <= 0 || beta <= 0 {
		return nil, fmt.Errorf("alpha and beta must be positive, got %g and %g", alpha, beta)
	}
	if src == nil {
		src = rand.NewPCG(rand.Uint64(), rand.Uint64())
	}
	d := distuv.Gamma{Alpha: alpha, Beta: beta, Src: src}
	return newNumDist("gamma", src, d.Prob, d.CDF, d.Quantile, d.Mean, d.Variance, d.Rand), nil
}

func makeBeta(alpha, beta float64, src rand.Source) (*NumDist, error) {
	if alpha <= 0 || beta <= 0 {
		return nil, fmt.Errorf("alpha and beta must be positive, got %g and %g", alpha, beta)
	}
	if src == nil {
		src = rand.NewPCG(rand.Uint64(), rand.Uint64())
	}
	d := distuv.Beta{Alpha: alpha, Beta: beta, Src: src}
	return newNumDist("beta", src, d.Prob, d.CDF, d.Quantile, d.Mean, d.Variance, d.Rand), nil
}

func makePoisson(lambda float64, src rand.Source) (*NumDist, error) {
	if lambda <= 0 {
		return nil, fmt.Errorf("lambda must be positive, got %g", lambda)
	}
	if src == nil {
		src = rand.NewPCG(rand.Uint64(), rand.Uint64())
	}
	d := distuv.Poisson{Lambda: lambda, Src: src}
	return newNumDist("poisson", src, d.Prob, d.CDF, nil, d.Mean, d.Variance, d.Rand), nil
}

// makeBinomial builds a binomial distribution. gonum models it as a float N,
// matching the discrete count.
func makeBinomial(n int, p float64, src rand.Source) (*NumDist, error) {
	if n <= 0 {
		return nil, fmt.Errorf("n must be positive, got %d", n)
	}
	if p < 0 || p > 1 {
		return nil, fmt.Errorf("p must be in [0, 1], got %g", p)
	}
	if src == nil {
		src = rand.NewPCG(rand.Uint64(), rand.Uint64())
	}
	d := distuv.Binomial{N: float64(n), P: p, Src: src}
	return newNumDist("binomial", src, d.Prob, d.CDF, nil, d.Mean, d.Variance, d.Rand), nil
}

// makeStudentT builds a Student's t distribution. Nu is the degrees of freedom;
// Sigma is a scale, so the standard deviation is Sigma*sqrt(Nu/(Nu-2)).
func makeStudentT(nu, mu, sigma float64, src rand.Source) (*NumDist, error) {
	if nu <= 0 {
		return nil, fmt.Errorf("degrees of freedom must be positive, got %g", nu)
	}
	if sigma <= 0 {
		return nil, fmt.Errorf("sigma must be positive, got %g", sigma)
	}
	if src == nil {
		src = rand.NewPCG(rand.Uint64(), rand.Uint64())
	}
	d := distuv.StudentsT{Mu: mu, Sigma: sigma, Nu: nu, Src: src}
	return newNumDist("student_t", src, d.Prob, d.CDF, d.Quantile, d.Mean, d.Variance, d.Rand), nil
}

// makeChiSquared builds a chi-squared distribution with k degrees of freedom.
func makeChiSquared(k float64, src rand.Source) (*NumDist, error) {
	if k <= 0 {
		return nil, fmt.Errorf("degrees of freedom must be positive, got %g", k)
	}
	if src == nil {
		src = rand.NewPCG(rand.Uint64(), rand.Uint64())
	}
	d := distuv.ChiSquared{K: k, Src: src}
	return newNumDist("chi_squared", src, d.Prob, d.CDF, d.Quantile, d.Mean, d.Variance, d.Rand), nil
}

// numListStatsArg reads a list argument as []float64, accepting an ml tensor too
// so `num.mean(tensor)` works with no conversion call.
func numListStatsArg(name string, pos int, o Object) ([]float64, Object) {
	return numVectorArg(name, pos, o)
}

// numStatBuiltin builds a list statistic taking one variable.
func numStatBuiltin(name string, f func(x []float64) float64) func(...Object) Object {
	return func(args ...Object) Object {
		if err := checkArgCount(name, 1, args); err != nil {
			return err
		}
		xs, errObj := numListStatsArg(name, 1, args[0])
		if errObj != nil {
			return errObj
		}
		if len(xs) == 0 {
			return newError("`%s` error: empty input", name)
		}
		return &Float{Value: f(xs)}
	}
}

// numStat2Builtin builds a two-variable statistic.
func numStat2Builtin(name string, f func(x, y []float64) (float64, error)) func(...Object) Object {
	return func(args ...Object) Object {
		if err := checkArgCount(name, 2, args); err != nil {
			return err
		}
		xs, errObj := numListStatsArg(name, 1, args[0])
		if errObj != nil {
			return errObj
		}
		ys, errObj := numListStatsArg(name, 2, args[1])
		if errObj != nil {
			return errObj
		}
		out, err := f(xs, ys)
		if err != nil {
			return newError("`%s` error: %s", name, err.Error())
		}
		return &Float{Value: out}
	}
}

// mlFloat64Tensor wraps a []float64 as a 1d ml tensor, for the sampling
// functions so their output feeds straight into ml.
func mlFloat64Tensor(vs []float64) (*ml.Tensor, error) {
	return ml.NewFloat64Tensor(vs, []int{len(vs)}, ml.CPU)
}

// histogramOf bins values into nbins equal-width buckets between min and max
// (or the data range when either is NaN), returning the counts and the edges.
func histogramOf(xs []float64, nbins int) (counts, edges []float64, err error) {
	if nbins <= 0 {
		return nil, nil, fmt.Errorf("bins must be positive, got %d", nbins)
	}
	if len(xs) == 0 {
		return nil, nil, fmt.Errorf("empty input")
	}
	lo, hi := xs[0], xs[0]
	for _, v := range xs {
		if v < lo {
			lo = v
		}
		if v > hi {
			hi = v
		}
	}
	if hi == lo {
		// A constant series still deserves one occupied bin.
		hi = lo + 1
	}
	width := (hi - lo) / float64(nbins)
	counts = make([]float64, nbins)
	edges = make([]float64, nbins+1)
	for i := range edges {
		edges[i] = lo + float64(i)*width
	}
	for _, v := range xs {
		idx := int((v - lo) / width)
		if idx >= nbins {
			idx = nbins - 1
		}
		if idx < 0 {
			idx = 0
		}
		counts[idx]++
	}
	return counts, edges, nil
}

// tTest is a one sample t-test of the sample mean against mu, returning the
// statistic and the two-sided p-value.
func tTest(xs []float64, mu float64) (tStat, p float64, err error) {
	n := float64(len(xs))
	if n < 2 {
		return 0, 0, fmt.Errorf("need at least 2 samples, got %d", len(xs))
	}
	mean := stat.Mean(xs, nil)
	variance := stat.Variance(xs, nil)
	if variance == 0 {
		return 0, 0, fmt.Errorf("sample variance is zero")
	}
	se := math.Sqrt(variance / n)
	tStat = (mean - mu) / se
	// Two-sided p from Student's t with n-1 degrees of freedom.
	dist := distuv.StudentsT{Mu: 0, Sigma: 1, Nu: n - 1, Src: rand.NewPCG(1, 0)}
	p = 2 * dist.CDF(-math.Abs(tStat))
	return tStat, p, nil
}

// chiSquareTest is a chi-squared goodness of fit test of observed counts against
// expected counts.
func chiSquareTest(observed, expected []float64) (chiSq, p float64, err error) {
	if len(observed) != len(expected) {
		return 0, 0, fmt.Errorf("observed has %d bins, expected has %d", len(observed), len(expected))
	}
	if len(observed) < 2 {
		return 0, 0, fmt.Errorf("need at least 2 bins, got %d", len(observed))
	}
	for i, e := range expected {
		if e <= 0 {
			return 0, 0, fmt.Errorf("expected count %d is not positive", i)
		}
		diff := observed[i] - e
		chiSq += diff * diff / e
	}
	dist := distuv.ChiSquared{K: float64(len(observed) - 1), Src: rand.NewPCG(1, 0)}
	p = 1 - dist.CDF(chiSq)
	return chiSq, p, nil
}

// The pure-array numerics for the num module: integration over sampled values,
// Fourier transforms, and interpolation.
//
// Everything that calls back into blue (quad over a function, minimize) lives in
// vm/builtins_num_vm.go instead, because only the vm can invoke a blue closure.

// integrateTrapezoidal integrates sampled values with the trapezoidal rule.
func integrateTrapezoidal(xs, ys []float64) (float64, error) {
	if len(xs) != len(ys) {
		return 0, fmt.Errorf("x has %d values, y has %d", len(xs), len(ys))
	}
	if len(xs) < 2 {
		return 0, fmt.Errorf("need at least 2 points, got %d", len(xs))
	}
	return integrate.Trapezoidal(xs, ys), nil
}

func integrateSimpsons(xs, ys []float64) (float64, error) {
	if len(xs) != len(ys) {
		return 0, fmt.Errorf("x has %d values, y has %d", len(xs), len(ys))
	}
	if len(xs) < 2 {
		return 0, fmt.Errorf("need at least 2 points, got %d", len(xs))
	}
	return integrate.Simpsons(xs, ys), nil
}

func integrateRomberg(ys []float64, dx float64) (float64, error) {
	if len(ys) < 2 {
		return 0, fmt.Errorf("need at least 2 values, got %d", len(ys))
	}
	if dx == 0 {
		return 0, fmt.Errorf("dx must be nonzero")
	}
	return integrate.Romberg(ys, dx), nil
}

// interpLinear fits a piecewise linear interpolant through the points.
func interpLinear(xs, ys []float64) (*interp.PiecewiseLinear, error) {
	if len(xs) != len(ys) {
		return nil, fmt.Errorf("x has %d values, y has %d", len(xs), len(ys))
	}
	if len(xs) < 2 {
		return nil, fmt.Errorf("need at least 2 points, got %d", len(xs))
	}
	f := &interp.PiecewiseLinear{}
	if err := f.Fit(xs, ys); err != nil {
		return nil, err
	}
	return f, nil
}

// interpCubic fits a natural cubic spline through the points.
func interpCubic(xs, ys []float64) (*interp.NaturalCubic, error) {
	if len(xs) != len(ys) {
		return nil, fmt.Errorf("x has %d values, y has %d", len(xs), len(ys))
	}
	if len(xs) < 2 {
		return nil, fmt.Errorf("need at least 2 points, got %d", len(xs))
	}
	f := &interp.NaturalCubic{}
	if err := f.Fit(xs, ys); err != nil {
		return nil, err
	}
	return f, nil
}

// fftTransform returns the real and imaginary parts of the discrete Fourier
// transform. blue has no complex type, so the result is two parallel slices.
//
// gonum's Coefficients returns the half spectrum, n/2+1 values, for a length n
// input: the second half is the conjugate mirror of the first for real input.
// that is enough to reconstruct the signal, and ifftInverse knows to expand it.
func fftTransform(seq []float64) (re, im []float64, err error) {
	if len(seq) == 0 {
		return nil, nil, fmt.Errorf("empty sequence")
	}
	t := fourier.NewFFT(len(seq))
	c := t.Coefficients(nil, seq)
	re = make([]float64, len(c))
	im = make([]float64, len(c))
	for i, v := range c {
		re[i] = real(v)
		im[i] = imag(v)
	}
	return re, im, nil
}

// ifftInverse rebuilds a real sequence from the half spectrum fftTransform
// returned. gonum's Sequence takes that same half spectrum, n/2+1 complex
// values, so no expansion is needed. The result is unnormalized: the forward
// and inverse pair multiplies the input by n, matching gonum.
func ifftInverse(re, im []float64) ([]float64, error) {
	if len(re) != len(im) {
		return nil, fmt.Errorf("real has %d values, imaginary has %d", len(re), len(im))
	}
	if len(re) == 0 {
		return nil, fmt.Errorf("empty spectrum")
	}
	half := len(re)
	// half = n/2+1 for a length n transform, so n = 2*(half-1).
	if half == 1 {
		// One bin is a constant signal of length 1.
		return []float64{re[0]}, nil
	}
	n := 2 * (half - 1)
	coeff := make([]complex128, half)
	for i := 0; i < half; i++ {
		coeff[i] = complex(re[i], im[i])
	}
	t := fourier.NewFFT(n)
	return t.Sequence(nil, coeff), nil
}

// dctTransform returns the discrete cosine transform of the sequence.
func dctTransform(seq []float64) ([]float64, error) {
	if len(seq) == 0 {
		return nil, fmt.Errorf("empty sequence")
	}
	t := fourier.NewDCT(len(seq))
	return t.Transform(nil, seq), nil
}

// fftFreqs returns the sample frequencies that pair with fftTransform's output,
// one per half-spectrum bin, so len(fft(x)) == len(fft_freqs(len(x))).
func fftFreqs(n int) ([]float64, error) {
	if n <= 0 {
		return nil, fmt.Errorf("n must be positive, got %d", n)
	}
	t := fourier.NewFFT(n)
	out := make([]float64, t.Len()/2+1)
	for i := range out {
		out[i] = t.Freq(i)
	}
	return out, nil
}

// numMatrixBuiltin wraps a matrix-returning function so it accepts a matrix, an
// ml tensor, or a nested list, and always returns a matrix handle.
func numMatrixBuiltin(name string, f func(a *mat.Dense) (*mat.Dense, error)) func(...Object) Object {
	return func(args ...Object) Object {
		if err := checkArgCount(name, 1, args); err != nil {
			return err
		}
		a, errObj := numMatrixArg(name, 1, args[0])
		if errObj != nil {
			return errObj
		}
		out, err := f(a)
		if err != nil {
			return newError("`%s` error: %s", name, err.Error())
		}
		return &GoObj[*NumMatrix]{Value: out}
	}
}

// numBinaryMatrixBuiltin wraps a two-matrix function with the same conversions.
func numBinaryMatrixBuiltin(name string, f func(a, b *mat.Dense) (*mat.Dense, error)) func(...Object) Object {
	return func(args ...Object) Object {
		if err := checkArgCount(name, 2, args); err != nil {
			return err
		}
		a, errObj := numMatrixArg(name, 1, args[0])
		if errObj != nil {
			return errObj
		}
		b, errObj := numMatrixArg(name, 2, args[1])
		if errObj != nil {
			return errObj
		}
		out, err := f(a, b)
		if err != nil {
			return newError("`%s` error: %s", name, err.Error())
		}
		return &GoObj[*NumMatrix]{Value: out}
	}
}

// numScalarBuiltin wraps a matrix-to-float64 reduction.
func numScalarBuiltin(name string, f func(a *mat.Dense) (float64, error)) func(...Object) Object {
	return func(args ...Object) Object {
		if err := checkArgCount(name, 1, args); err != nil {
			return err
		}
		a, errObj := numMatrixArg(name, 1, args[0])
		if errObj != nil {
			return errObj
		}
		out, err := f(a)
		if err != nil {
			return newError("`%s` error: %s", name, err.Error())
		}
		return &Float{Value: out}
	}
}

var NumBuiltins = []*Builtin{
	{
		Name: "_num_dense",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("dense", 1, args); err != nil {
				return err
			}
			m, errObj := numMatrixArg("dense", 1, args[0])
			if errObj != nil {
				return errObj
			}
			return &GoObj[*NumMatrix]{Value: m}
		},
		HelpStr: helpStrArgs{
			explanation: "`dense` builds a matrix from a nested list, an ml tensor, or another matrix",
			signature:   "dense(data: list|tensor|matrix) -> matrix",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "dense([[1.0, 2.0], [3.0, 4.0]]) => matrix",
		}.String(),
	},
	{
		Name: "_num_zeros",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("zeros", 2, args); err != nil {
				return err
			}
			r, ok := args[0].(*Integer)
			if !ok {
				return newPositionalTypeError("zeros", 1, INTEGER_OBJ, args[0].Type())
			}
			c, ok := args[1].(*Integer)
			if !ok {
				return newPositionalTypeError("zeros", 2, INTEGER_OBJ, args[1].Type())
			}
			return &GoObj[*NumMatrix]{Value: mat.NewDense(int(r.Value), int(c.Value), nil)}
		},
		HelpStr: helpStrArgs{
			explanation: "`zeros` returns a rows by cols matrix of zeros",
			signature:   "zeros(rows: int, cols: int) -> matrix",
			errors:      "InvalidArgCount,PositionalType",
			example:     "zeros(2, 3) => matrix",
		}.String(),
	},
	{
		Name: "_num_eye",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("eye", 1, args); err != nil {
				return err
			}
			n, ok := args[0].(*Integer)
			if !ok {
				return newPositionalTypeError("eye", 1, INTEGER_OBJ, args[0].Type())
			}
			sz := int(n.Value)
			m := mat.NewDense(sz, sz, nil)
			for i := 0; i < sz; i++ {
				m.Set(i, i, 1)
			}
			return &GoObj[*NumMatrix]{Value: m}
		},
		HelpStr: helpStrArgs{
			explanation: "`eye` returns an n by n identity matrix",
			signature:   "eye(n: int) -> matrix",
			errors:      "InvalidArgCount,PositionalType",
			example:     "eye(3) => matrix",
		}.String(),
	},
	{
		Name: "_num_diag",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("diag", 1, args); err != nil {
				return err
			}
			vs, errObj := numVectorArg("diag", 1, args[0])
			if errObj != nil {
				return errObj
			}
			d := mat.NewDiagDense(len(vs), vs)
			// Convert to a dense matrix so every num value has one backing type.
			out := mat.NewDense(len(vs), len(vs), nil)
			out.Copy(d)
			return &GoObj[*NumMatrix]{Value: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`diag` returns a square matrix with the given vector on its diagonal",
			signature:   "diag(v: list|tensor) -> matrix",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "diag([1.0, 2.0, 3.0]) => matrix",
		}.String(),
	},
	{
		Name: "_num_rows",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("rows", 1, args); err != nil {
				return err
			}
			m, errObj := numMatrixArg("rows", 1, args[0])
			if errObj != nil {
				return errObj
			}
			r, _ := m.Dims()
			return NewInteger(int64(r))
		},
		HelpStr: helpStrArgs{
			explanation: "`rows` returns the number of rows in a matrix",
			signature:   "rows(m: matrix|tensor|list) -> int",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "rows([[1.0, 2.0]]) => 1",
		}.String(),
	},
	{
		Name: "_num_cols",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("cols", 1, args); err != nil {
				return err
			}
			m, errObj := numMatrixArg("cols", 1, args[0])
			if errObj != nil {
				return errObj
			}
			_, c := m.Dims()
			return NewInteger(int64(c))
		},
		HelpStr: helpStrArgs{
			explanation: "`cols` returns the number of columns in a matrix",
			signature:   "cols(m: matrix|tensor|list) -> int",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "cols([[1.0, 2.0]]) => 2",
		}.String(),
	},
	{
		Name: "_num_at",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("at", 3, args); err != nil {
				return err
			}
			m, errObj := numMatrixArg("at", 1, args[0])
			if errObj != nil {
				return errObj
			}
			i, ok := args[1].(*Integer)
			if !ok {
				return newPositionalTypeError("at", 2, INTEGER_OBJ, args[1].Type())
			}
			j, ok := args[2].(*Integer)
			if !ok {
				return newPositionalTypeError("at", 3, INTEGER_OBJ, args[2].Type())
			}
			r, c := m.Dims()
			if i.Value < 0 || i.Value >= int64(r) || j.Value < 0 || j.Value >= int64(c) {
				return newError("`at` error: index (%d, %d) out of range for a %dx%d matrix", i.Value, j.Value, r, c)
			}
			return &Float{Value: m.At(int(i.Value), int(j.Value))}
		},
		HelpStr: helpStrArgs{
			explanation: "`at` returns the element at row i, column j",
			signature:   "at(m: matrix|tensor|list, i: int, j: int) -> float",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "at([[1.0, 2.0]], 0, 1) => 2.0",
		}.String(),
	},
	{
		Name: "_num_to_list",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("to_list", 1, args); err != nil {
				return err
			}
			m, errObj := numMatrixArg("to_list", 1, args[0])
			if errObj != nil {
				return errObj
			}
			return float64ObjectMatrix(m)
		},
		HelpStr: helpStrArgs{
			explanation: "`to_list` returns a matrix as a nested list of lists",
			signature:   "to_list(m: matrix|tensor|list) -> list",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "to_list([[1.0, 2.0]]) => [[1.0, 2.0]]",
		}.String(),
	},
	{
		Name: "_num_str",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("str", 1, args); err != nil {
				return err
			}
			m, errObj := numMatrixArg("str", 1, args[0])
			if errObj != nil {
				return errObj
			}
			return &Stringo{Value: fmt.Sprintf("%v", mat.Formatted(m))}
		},
		HelpStr: helpStrArgs{
			explanation: "`str` returns a readable string form of a matrix",
			signature:   "str(m: matrix|tensor|list) -> str",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "str(eye(2)) => string form",
		}.String(),
	},
	{
		Name: "_num_add",
		Fun:  numBinaryMatrixBuiltin("add", func(a, b *mat.Dense) (*mat.Dense, error) { return denseAdd(a, b, 1) }),
		HelpStr: helpStrArgs{
			explanation: "`add` returns the elementwise sum of two matrices, which must have the same shape",
			signature:   "add(a: matrix|tensor|list, b: matrix|tensor|list) -> matrix",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "add([[1.0]], [[2.0]]) => [[3.0]]",
		}.String(),
	},
	{
		Name: "_num_sub",
		Fun:  numBinaryMatrixBuiltin("sub", func(a, b *mat.Dense) (*mat.Dense, error) { return denseAdd(a, b, -1) }),
		HelpStr: helpStrArgs{
			explanation: "`sub` returns the elementwise difference of two matrices",
			signature:   "sub(a: matrix|tensor|list, b: matrix|tensor|list) -> matrix",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "sub([[3.0]], [[2.0]]) => [[1.0]]",
		}.String(),
	},
	{
		Name: "_num_mul",
		Fun:  numBinaryMatrixBuiltin("mul", denseElemMul),
		HelpStr: helpStrArgs{
			explanation: "`mul` returns the elementwise product of two matrices, which must have the same shape",
			signature:   "mul(a: matrix|tensor|list, b: matrix|tensor|list) -> matrix",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "mul([[2.0]], [[3.0]]) => [[6.0]]",
		}.String(),
	},
	{
		Name: "_num_matmul",
		Fun:  numBinaryMatrixBuiltin("matmul", denseMatMul),
		HelpStr: helpStrArgs{
			explanation: "`matmul` returns the matrix product a @ b",
			signature:   "matmul(a: matrix|tensor|list, b: matrix|tensor|list) -> matrix",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "matmul(eye(2), [[1.0], [2.0]]) => [[1.0], [2.0]]",
		}.String(),
	},
	{
		Name: "_num_scale",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("scale", 2, args); err != nil {
				return err
			}
			m, errObj := numMatrixArg("scale", 1, args[0])
			if errObj != nil {
				return errObj
			}
			f, ok := objectToFloat64(args[1])
			if !ok {
				return newPositionalTypeError("scale", 2, FLOAT_OBJ, args[1].Type())
			}
			return &GoObj[*NumMatrix]{Value: denseScale(m, f)}
		},
		HelpStr: helpStrArgs{
			explanation: "`scale` multiplies every element by a scalar",
			signature:   "scale(m: matrix|tensor|list, factor: float) -> matrix",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "scale([[1.0, 2.0]], 2.0) => [[2.0, 4.0]]",
		}.String(),
	},
	{
		Name: "_num_transpose",
		Fun:  numMatrixBuiltin("transpose", denseTranspose),
		HelpStr: helpStrArgs{
			explanation: "`transpose` returns the transpose of a matrix",
			signature:   "transpose(m: matrix|tensor|list) -> matrix",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "transpose([[1.0, 2.0]]) => [[1.0], [2.0]]",
		}.String(),
	},
	{
		Name: "_num_det",
		Fun:  numScalarBuiltin("det", denseDet),
		HelpStr: helpStrArgs{
			explanation: "`det` returns the determinant of a square matrix",
			signature:   "det(m: matrix|tensor|list) -> float",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "det([[1.0, 2.0], [3.0, 4.0]]) => -2.0",
		}.String(),
	},
	{
		Name: "_num_trace",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("trace", 1, args); err != nil {
				return err
			}
			m, errObj := numMatrixArg("trace", 1, args[0])
			if errObj != nil {
				return errObj
			}
			r, c := m.Dims()
			if r != c {
				return newError("`trace` error: matrix is %dx%d, want square", r, c)
			}
			sum := 0.0
			for i := 0; i < r; i++ {
				sum += m.At(i, i)
			}
			return &Float{Value: sum}
		},
		HelpStr: helpStrArgs{
			explanation: "`trace` returns the sum of the diagonal of a square matrix",
			signature:   "trace(m: matrix|tensor|list) -> float",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "trace(eye(3)) => 3.0",
		}.String(),
	},
	{
		Name: "_num_inverse",
		Fun:  numMatrixBuiltin("inverse", denseInverse),
		HelpStr: helpStrArgs{
			explanation: "`inverse` returns the inverse of a square matrix",
			signature:   "inverse(m: matrix|tensor|list) -> matrix",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "inverse(eye(3)) => identity",
		}.String(),
	},
	{
		Name: "_num_solve",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("solve", 2, args); err != nil {
				return err
			}
			a, errObj := numMatrixArg("solve", 1, args[0])
			if errObj != nil {
				return errObj
			}
			b, errObj := numMatrixArg("solve", 2, args[1])
			if errObj != nil {
				return errObj
			}
			out, err := denseSolve(a, b)
			if err != nil {
				return newError("`solve` error: %s", err.Error())
			}
			return &GoObj[*NumMatrix]{Value: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`solve` returns x such that a @ x = b, for a square a",
			signature:   "solve(a: matrix|tensor|list, b: matrix|tensor|list) -> matrix",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "solve([[2.0, 0.0], [0.0, 4.0]], [[2.0], [8.0]]) => [[1.0], [2.0]]",
		}.String(),
	},
	{
		Name: "_num_rank",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("rank", 1, args); err != nil {
				return err
			}
			m, errObj := numMatrixArg("rank", 1, args[0])
			if errObj != nil {
				return errObj
			}
			r, ok := denseRank(m)
			if !ok {
				return newError("`rank` error: singular value decomposition did not converge")
			}
			return NewInteger(int64(r))
		},
		HelpStr: helpStrArgs{
			explanation: "`rank` returns the rank of a matrix",
			signature:   "rank(m: matrix|tensor|list) -> int",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "rank(eye(3)) => 3",
		}.String(),
	},
	{
		Name: "_num_norm",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("norm", 2, args); err != nil {
				return err
			}
			m, errObj := numMatrixArg("norm", 1, args[0])
			if errObj != nil {
				return errObj
			}
			kind, ok := args[1].(*Stringo)
			if !ok {
				return newPositionalTypeError("norm", 2, STRING_OBJ, args[1].Type())
			}
			which, ok := parseNormKind(kind.Value)
			if !ok {
				return newError("`norm` error: unknown norm %q, want 'fro', '1', 'inf', or '2'", kind.Value)
			}
			return &Float{Value: denseNorm(m, which)}
		},
		HelpStr: helpStrArgs{
			explanation: "`norm` returns a matrix norm: 'fro' (Frobenius), '1' (max absolute column sum), or 'inf' (max absolute row sum)",
			signature:   "norm(m: matrix|tensor|list, kind: str='fro') -> float",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "norm(eye(3), 'fro') => 1.7320508",
		}.String(),
	},
	{
		Name: "_num_eig",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("eig", 1, args); err != nil {
				return err
			}
			m, errObj := numMatrixArg("eig", 1, args[0])
			if errObj != nil {
				return errObj
			}
			return denseEigObject(m)
		},
		HelpStr: helpStrArgs{
			explanation: "`eig` returns {values: matrix, vectors: matrix} for a square matrix; values are complex, stored as a two column [re, im] matrix",
			signature:   "eig(m: matrix|tensor|list) -> map",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "eig([[2.0, 0.0], [0.0, 3.0]]) => {values: ..., vectors: ...}",
		}.String(),
	},
	{
		Name: "_num_svd",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("svd", 1, args); err != nil {
				return err
			}
			m, errObj := numMatrixArg("svd", 1, args[0])
			if errObj != nil {
				return errObj
			}
			return denseSVDObject(m)
		},
		HelpStr: helpStrArgs{
			explanation: "`svd` returns {u: matrix, values: list, v: matrix}, the singular value decomposition",
			signature:   "svd(m: matrix|tensor|list) -> map",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "svd([[1.0, 0.0], [0.0, 2.0]]) => {u: ..., values: [2.0, 1.0], v: ...}",
		}.String(),
	},
	{
		Name: "_num_cholesky",
		Fun:  numMatrixBuiltin("cholesky", denseCholesky),
		HelpStr: helpStrArgs{
			explanation: "`cholesky` returns the lower triangular Cholesky factor of a symmetric positive definite matrix",
			signature:   "cholesky(m: matrix|tensor|list) -> matrix",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "cholesky([[1.0, 0.0], [0.0, 1.0]]) => identity",
		}.String(),
	},
	{
		Name: "_num_qr",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("qr", 1, args); err != nil {
				return err
			}
			m, errObj := numMatrixArg("qr", 1, args[0])
			if errObj != nil {
				return errObj
			}
			return denseQRObject(m)
		},
		HelpStr: helpStrArgs{
			explanation: "`qr` returns {q: matrix, r: matrix}, the QR decomposition",
			signature:   "qr(m: matrix|tensor|list) -> map",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "qr([[1.0, 2.0], [3.0, 4.0]]) => {q: ..., r: ...}",
		}.String(),
	},
	{
		Name: "_num_lu",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("lu", 1, args); err != nil {
				return err
			}
			m, errObj := numMatrixArg("lu", 1, args[0])
			if errObj != nil {
				return errObj
			}
			return denseLUObject(m)
		},
		HelpStr: helpStrArgs{
			explanation: "`lu` returns {l: matrix, u: matrix}, the LU decomposition",
			signature:   "lu(m: matrix|tensor|list) -> map",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "lu([[2.0, 1.0], [1.0, 3.0]]) => {l: ..., u: ...}",
		}.String(),
	},
	{
		Name: "_num_from_tensor",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("from_tensor", 1, args); err != nil {
				return err
			}
			t, ok := args[0].(*Tensor)
			if !ok {
				return newPositionalTypeError("from_tensor", 1, TENSOR_OBJ, args[0].Type())
			}
			m, err := tensorToMatrix(t.T)
			if err != nil {
				return newError("`from_tensor` error: %s", err.Error())
			}
			return &GoObj[*NumMatrix]{Value: m}
		},
		HelpStr: helpStrArgs{
			explanation: "`from_tensor` converts an ml tensor to a matrix; rank 1 becomes a 1 by n row, rank > 2 flattens trailing dims into columns",
			signature:   "from_tensor(t: tensor) -> matrix",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "from_tensor(ml.zeros([2, 3])) => matrix",
		}.String(),
	},
	{
		Name: "_num_to_tensor",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("to_tensor", 1, args); err != nil {
				return err
			}
			m, errObj := numMatrixArg("to_tensor", 1, args[0])
			if errObj != nil {
				return errObj
			}
			t, err := mlMatrixTensor(m, ml.Float32, ml.CPU)
			if err != nil {
				return newError("`to_tensor` error: %s", err.Error())
			}
			return &Tensor{T: t}
		},
		HelpStr: helpStrArgs{
			explanation: "`to_tensor` converts a matrix to an ml float32 tensor on the cpu",
			signature:   "to_tensor(m: matrix|list) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "to_tensor(eye(2)) => Tensor{shape: [2 2]}",
		}.String(),
	},
	{
		Name: "_num_to_tensor_f64",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("to_tensor_f64", 1, args); err != nil {
				return err
			}
			m, errObj := numMatrixArg("to_tensor_f64", 1, args[0])
			if errObj != nil {
				return errObj
			}
			r, c := m.Dims()
			t, err := ml.NewFloat64Tensor(matrixToFloat64Data(m), []int{r, c}, ml.CPU)
			if err != nil {
				return newError("`to_tensor_f64` error: %s", err.Error())
			}
			return &Tensor{T: t}
		},
		HelpStr: helpStrArgs{
			explanation: "`to_tensor_f64` converts a matrix to an ml float64 tensor, so a round trip through num is exact",
			signature:   "to_tensor_f64(m: matrix|list) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "to_tensor_f64(eye(2)) => Tensor{shape: [2 2]}",
		}.String(),
	},
	{
		Name: "_num_normal",
		Fun: numDistBuiltin("normal", func(args []Object) (*NumDist, Object) {
			if err := checkArgsCount("normal", []int{0, 1, 2, 3}, args); err != nil {
				return nil, err
			}
			mu, sigma := 0.0, 1.0
			if len(args) >= 1 {
				v, errObj := numFloatArg("normal", 1, args[0])
				if errObj != nil {
					return nil, errObj
				}
				mu = v
			}
			if len(args) >= 2 {
				v, errObj := numFloatArg("normal", 2, args[1])
				if errObj != nil {
					return nil, errObj
				}
				sigma = v
			}
			var src rand.Source
			if len(args) == 3 {
				s, errObj := numSourceArg("normal", args[2])
				if errObj != nil {
					return nil, errObj
				}
				src = s
			}
			d, err := makeNormal(mu, sigma, src)
			if err != nil {
				return nil, newError("`normal` error: %s", err.Error())
			}
			return d, nil
		}),
		HelpStr: helpStrArgs{
			explanation: "`normal` returns a normal distribution; the seed makes sampling reproducible",
			signature:   "normal(mu: float=0.0, sigma: float=1.0, seed: int|null=null) -> distribution",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "normal(0.0, 1.0).cdf(1.96) => 0.975",
		}.String(),
	},
	{
		Name: "_num_uniform",
		Fun: numDistBuiltin("uniform", func(args []Object) (*NumDist, Object) {
			if err := checkArgsCount("uniform", []int{0, 1, 2, 3}, args); err != nil {
				return nil, err
			}
			lo, hi := 0.0, 1.0
			if len(args) >= 1 {
				v, errObj := numFloatArg("uniform", 1, args[0])
				if errObj != nil {
					return nil, errObj
				}
				lo = v
			}
			if len(args) >= 2 {
				v, errObj := numFloatArg("uniform", 2, args[1])
				if errObj != nil {
					return nil, errObj
				}
				hi = v
			}
			var src rand.Source
			if len(args) == 3 {
				s, errObj := numSourceArg("uniform", args[2])
				if errObj != nil {
					return nil, errObj
				}
				src = s
			}
			d, err := makeUniform(lo, hi, src)
			if err != nil {
				return nil, newError("`uniform` error: %s", err.Error())
			}
			return d, nil
		}),
		HelpStr: helpStrArgs{
			explanation: "`uniform` returns a uniform distribution over [min, max)",
			signature:   "uniform(min: float=0.0, max: float=1.0, seed: int|null=null) -> distribution",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "uniform(0.0, 1.0).mean() => 0.5",
		}.String(),
	},
	{
		Name: "_num_exponential",
		Fun: numDistBuiltin("exponential", func(args []Object) (*NumDist, Object) {
			if err := checkArgsCount("exponential", []int{1, 2}, args); err != nil {
				return nil, err
			}
			rate, errObj := numFloatArg("exponential", 1, args[0])
			if errObj != nil {
				return nil, errObj
			}
			var src rand.Source
			if len(args) == 2 {
				s, errObj := numSourceArg("exponential", args[1])
				if errObj != nil {
					return nil, errObj
				}
				src = s
			}
			d, err := makeExponential(rate, src)
			if err != nil {
				return nil, newError("`exponential` error: %s", err.Error())
			}
			return d, nil
		}),
		HelpStr: helpStrArgs{
			explanation: "`exponential` returns an exponential distribution with the given rate",
			signature:   "exponential(rate: float, seed: int|null=null) -> distribution",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "exponential(2.0).mean() => 0.5",
		}.String(),
	},
	{
		Name: "_num_gamma_dist",
		Fun: numDistBuiltin("gamma", func(args []Object) (*NumDist, Object) {
			if err := checkArgsCount("gamma", []int{2, 3}, args); err != nil {
				return nil, err
			}
			alpha, errObj := numFloatArg("gamma", 1, args[0])
			if errObj != nil {
				return nil, errObj
			}
			beta, errObj := numFloatArg("gamma", 2, args[1])
			if errObj != nil {
				return nil, errObj
			}
			var src rand.Source
			if len(args) == 3 {
				s, errObj := numSourceArg("gamma", args[2])
				if errObj != nil {
					return nil, errObj
				}
				src = s
			}
			d, err := makeGamma(alpha, beta, src)
			if err != nil {
				return nil, newError("`gamma` error: %s", err.Error())
			}
			return d, nil
		}),
		HelpStr: helpStrArgs{
			explanation: "`gamma` returns a gamma distribution with shape alpha and rate beta",
			signature:   "gamma(alpha: float, beta: float, seed: int|null=null) -> distribution",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "gamma(2.0, 1.0).mean() => 2.0",
		}.String(),
	},
	{
		Name: "_num_beta_dist",
		Fun: numDistBuiltin("beta", func(args []Object) (*NumDist, Object) {
			if err := checkArgsCount("beta", []int{2, 3}, args); err != nil {
				return nil, err
			}
			alpha, errObj := numFloatArg("beta", 1, args[0])
			if errObj != nil {
				return nil, errObj
			}
			beta, errObj := numFloatArg("beta", 2, args[1])
			if errObj != nil {
				return nil, errObj
			}
			var src rand.Source
			if len(args) == 3 {
				s, errObj := numSourceArg("beta", args[2])
				if errObj != nil {
					return nil, errObj
				}
				src = s
			}
			d, err := makeBeta(alpha, beta, src)
			if err != nil {
				return nil, newError("`beta` error: %s", err.Error())
			}
			return d, nil
		}),
		HelpStr: helpStrArgs{
			explanation: "`beta` returns a beta distribution with shape parameters alpha and beta",
			signature:   "beta(alpha: float, beta: float, seed: int|null=null) -> distribution",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "beta(2.0, 2.0).mean() => 0.5",
		}.String(),
	},
	{
		Name: "_num_poisson",
		Fun: numDistBuiltin("poisson", func(args []Object) (*NumDist, Object) {
			if err := checkArgsCount("poisson", []int{1, 2}, args); err != nil {
				return nil, err
			}
			lambda, errObj := numFloatArg("poisson", 1, args[0])
			if errObj != nil {
				return nil, errObj
			}
			var src rand.Source
			if len(args) == 2 {
				s, errObj := numSourceArg("poisson", args[1])
				if errObj != nil {
					return nil, errObj
				}
				src = s
			}
			d, err := makePoisson(lambda, src)
			if err != nil {
				return nil, newError("`poisson` error: %s", err.Error())
			}
			return d, nil
		}),
		HelpStr: helpStrArgs{
			explanation: "`poisson` returns a Poisson distribution with the given rate",
			signature:   "poisson(lambda: float, seed: int|null=null) -> distribution",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "poisson(3.0).mean() => 3.0",
		}.String(),
	},
	{
		Name: "_num_binomial",
		Fun: numDistBuiltin("binomial", func(args []Object) (*NumDist, Object) {
			if err := checkArgsCount("binomial", []int{1, 2, 3}, args); err != nil {
				return nil, err
			}
			n, ok := args[0].(*Integer)
			if !ok {
				return nil, newPositionalTypeError("binomial", 1, INTEGER_OBJ, args[0].Type())
			}
			p := 0.5
			if len(args) >= 2 {
				v, errObj := numFloatArg("binomial", 2, args[1])
				if errObj != nil {
					return nil, errObj
				}
				p = v
			}
			var src rand.Source
			if len(args) == 3 {
				s, errObj := numSourceArg("binomial", args[2])
				if errObj != nil {
					return nil, errObj
				}
				src = s
			}
			d, err := makeBinomial(int(n.Value), p, src)
			if err != nil {
				return nil, newError("`binomial` error: %s", err.Error())
			}
			return d, nil
		}),
		HelpStr: helpStrArgs{
			explanation: "`binomial` returns a binomial distribution over n trials with success probability p",
			signature:   "binomial(n: int, p: float=0.5, seed: int|null=null) -> distribution",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "binomial(10, 0.5).mean() => 5.0",
		}.String(),
	},
	{
		Name: "_num_dist_prob",
		Fun:  numDistScalarBuiltin("dist_prob", func(d *NumDist, x float64) float64 { return d.Prob(x) }),
		HelpStr: helpStrArgs{
			explanation: "`dist_prob` is the probability density or mass at x",
			signature:   "dist_prob(d: distribution, x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "dist_prob(normal(0.0, 1.0), 0.0) => 0.3989",
		}.String(),
	},
	{
		Name: "_num_dist_cdf",
		Fun:  numDistScalarBuiltin("dist_cdf", func(d *NumDist, x float64) float64 { return d.CDF(x) }),
		HelpStr: helpStrArgs{
			explanation: "`dist_cdf` is the cumulative probability up to x",
			signature:   "dist_cdf(d: distribution, x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "dist_cdf(normal(0.0, 1.0), 1.96) => 0.975",
		}.String(),
	},
	{
		Name: "_num_dist_quantile",
		Fun:  numDistScalarBuiltin("dist_quantile", func(d *NumDist, p float64) float64 { return d.Quantile(p) }),
		HelpStr: helpStrArgs{
			explanation: "`dist_quantile` is the inverse CDF",
			signature:   "dist_quantile(d: distribution, p: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "dist_quantile(normal(0.0, 1.0), 0.975) => 1.96",
		}.String(),
	},
	{
		Name: "_num_dist_mean",
		Fun:  numDistMomentBuiltin("dist_mean", func(d *NumDist) float64 { return d.Mean() }),
		HelpStr: helpStrArgs{
			explanation: "`dist_mean` is the distribution mean",
			signature:   "dist_mean(d: distribution) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "dist_mean(normal(2.0, 1.0)) => 2.0",
		}.String(),
	},
	{
		Name: "_num_dist_variance",
		Fun:  numDistMomentBuiltin("dist_variance", func(d *NumDist) float64 { return d.Variance() }),
		HelpStr: helpStrArgs{
			explanation: "`dist_variance` is the distribution variance",
			signature:   "dist_variance(d: distribution) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "dist_variance(normal(0.0, 2.0)) => 4.0",
		}.String(),
	},
	{
		Name: "_num_dist_rand",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("dist_rand", 1, args); err != nil {
				return err
			}
			d, errObj := numDistArg("dist_rand", 1, args[0])
			if errObj != nil {
				return errObj
			}
			return &Float{Value: d.Rand()}
		},
		HelpStr: helpStrArgs{
			explanation: "`dist_rand` draws one sample from a distribution",
			signature:   "dist_rand(d: distribution) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "dist_rand(normal(0.0, 1.0)) => a sample",
		}.String(),
	},
	{
		Name: "_num_dist_sample",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("dist_sample", 2, args); err != nil {
				return err
			}
			d, errObj := numDistArg("dist_sample", 1, args[0])
			if errObj != nil {
				return errObj
			}
			n, ok := args[1].(*Integer)
			if !ok {
				return newPositionalTypeError("dist_sample", 2, INTEGER_OBJ, args[1].Type())
			}
			if n.Value < 0 {
				return newError("`dist_sample` error: n must be non-negative")
			}
			t, err := mlFloat64Tensor(d.Sample(int(n.Value)))
			if err != nil {
				return newError("`dist_sample` error: %s", err.Error())
			}
			return &Tensor{T: t}
		},
		HelpStr: helpStrArgs{
			explanation: "`dist_sample` draws n samples as a float64 1d tensor, ready for ml",
			signature:   "dist_sample(d: distribution, n: int) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "dist_sample(normal(0.0, 1.0), 100) => Tensor{shape: [100]}",
		}.String(),
	},
	{
		Name: "_num_rand_normal",
		Fun: numSampleBuiltin("rand_normal", func(d *NumDist, n int) []float64 { return d.Sample(n) },
			func(src rand.Source) (*NumDist, error) { return makeNormal(0, 1, src) }),
		HelpStr: helpStrArgs{
			explanation: "`rand_normal` draws n standard normal samples as a float64 tensor; the seed makes it reproducible",
			signature:   "rand_normal(n: int, seed: int|null=null) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "rand_normal(10, 1) => Tensor{shape: [10]}",
		}.String(),
	},
	{
		Name: "_num_rand_uniform",
		Fun: numSampleBuiltin("rand_uniform", func(d *NumDist, n int) []float64 { return d.Sample(n) },
			func(src rand.Source) (*NumDist, error) { return makeUniform(0, 1, src) }),
		HelpStr: helpStrArgs{
			explanation: "`rand_uniform` draws n uniform samples in [0, 1) as a float64 tensor; the seed makes it reproducible",
			signature:   "rand_uniform(n: int, seed: int|null=null) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "rand_uniform(10, 1) => Tensor{shape: [10]}",
		}.String(),
	},
	{
		Name: "_num_mean",
		Fun:  numStatBuiltin("mean", func(x []float64) float64 { return stat.Mean(x, nil) }),
		HelpStr: helpStrArgs{
			explanation: "`mean` is the arithmetic mean; it accepts a list or an ml tensor",
			signature:   "mean(x: list|tensor) -> float",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "mean([1.0, 2.0, 3.0]) => 2.0",
		}.String(),
	},
	{
		Name: "_num_variance",
		Fun:  numStatBuiltin("variance", func(x []float64) float64 { return stat.Variance(x, nil) }),
		HelpStr: helpStrArgs{
			explanation: "`variance` is the unbiased sample variance, dividing by n-1",
			signature:   "variance(x: list|tensor) -> float",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "variance([1.0, 2.0, 3.0]) => 1.0",
		}.String(),
	},
	{
		Name: "_num_std_dev",
		Fun:  numStatBuiltin("std_dev", func(x []float64) float64 { return stat.StdDev(x, nil) }),
		HelpStr: helpStrArgs{
			explanation: "`std_dev` is the square root of the unbiased sample variance",
			signature:   "std_dev(x: list|tensor) -> float",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "std_dev([1.0, 2.0, 3.0]) => 1.0",
		}.String(),
	},
	{
		Name: "_num_median",
		Fun:  numStatBuiltin("median", func(x []float64) float64 { return stat.Quantile(0.5, stat.Empirical, x, nil) }),
		HelpStr: helpStrArgs{
			explanation: "`median` is the 0.5 empirical quantile",
			signature:   "median(x: list|tensor) -> float",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "median([1.0, 2.0, 3.0]) => 2.0",
		}.String(),
	},
	{
		Name: "_num_quantile",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("quantile", 2, args); err != nil {
				return err
			}
			xs, errObj := numListStatsArg("quantile", 1, args[0])
			if errObj != nil {
				return errObj
			}
			p, errObj := numFloatArg("quantile", 2, args[1])
			if errObj != nil {
				return errObj
			}
			if p < 0 || p > 1 {
				return newError("`quantile` error: p must be in [0, 1], got %g", p)
			}
			if len(xs) == 0 {
				return newError("`quantile` error: empty input")
			}
			return &Float{Value: stat.Quantile(p, stat.Empirical, xs, nil)}
		},
		HelpStr: helpStrArgs{
			explanation: "`quantile` is the empirical quantile at probability p",
			signature:   "quantile(x: list|tensor, p: float) -> float",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "quantile([1.0, 2.0, 3.0], 0.5) => 2.0",
		}.String(),
	},
	{
		Name: "_num_correlation",
		Fun: numStat2Builtin("correlation", func(x, y []float64) (float64, error) {
			if len(x) != len(y) {
				return 0, fmt.Errorf("x has %d values, y has %d", len(x), len(y))
			}
			return stat.Correlation(x, y, nil), nil
		}),
		HelpStr: helpStrArgs{
			explanation: "`correlation` is the Pearson correlation coefficient of two samples",
			signature:   "correlation(x: list|tensor, y: list|tensor) -> float",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "correlation([1.0, 2.0], [2.0, 4.0]) => 1.0",
		}.String(),
	},
	{
		Name: "_num_covariance",
		Fun: numStat2Builtin("covariance", func(x, y []float64) (float64, error) {
			if len(x) != len(y) {
				return 0, fmt.Errorf("x has %d values, y has %d", len(x), len(y))
			}
			return stat.Covariance(x, y, nil), nil
		}),
		HelpStr: helpStrArgs{
			explanation: "`covariance` is the unbiased sample covariance of two samples",
			signature:   "covariance(x: list|tensor, y: list|tensor) -> float",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "covariance([1.0, 2.0], [2.0, 4.0]) => 1.0",
		}.String(),
	},
	{
		Name: "_num_entropy",
		Fun: func(args ...Object) Object {
			if err := checkArgsCount("entropy", []int{1, 2}, args); err != nil {
				return err
			}
			xs, errObj := numListStatsArg("entropy", 1, args[0])
			if errObj != nil {
				return errObj
			}
			if len(args) == 2 {
				// gonum's Entropy takes no weights, so weighted entropy is not
				// silently ignored: it is rejected.
				return newError("`entropy` error: weighted entropy is not supported")
			}
			if len(xs) == 0 {
				return newError("`entropy` error: empty input")
			}
			return &Float{Value: stat.Entropy(xs)}
		},
		HelpStr: helpStrArgs{
			explanation: "`entropy` is the Shannon entropy of a probability distribution",
			signature:   "entropy(p: list|tensor) -> float",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "entropy([0.5, 0.5]) => 0.693",
		}.String(),
	},
	{
		Name: "_num_histogram",
		Fun: func(args ...Object) Object {
			if err := checkArgsCount("histogram", []int{1, 2}, args); err != nil {
				return err
			}
			xs, errObj := numListStatsArg("histogram", 1, args[0])
			if errObj != nil {
				return errObj
			}
			bins := 10
			if len(args) == 2 {
				n, ok := args[1].(*Integer)
				if !ok {
					return newPositionalTypeError("histogram", 2, INTEGER_OBJ, args[1].Type())
				}
				bins = int(n.Value)
			}
			counts, edges, err := histogramOf(xs, bins)
			if err != nil {
				return newError("`histogram` error: %s", err.Error())
			}
			return numMapObject(
				"counts", float64ObjectList(counts),
				"edges", float64ObjectList(edges),
			)
		},
		HelpStr: helpStrArgs{
			explanation: "`histogram` bins values into equal-width buckets and returns {counts: list, edges: list}",
			signature:   "histogram(x: list|tensor, bins: int=10) -> map",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "histogram([1.0, 2.0, 3.0], 2) => {counts: [...], edges: [...]}",
		}.String(),
	},
	{
		Name: "_num_t_test",
		Fun: func(args ...Object) Object {
			if err := checkArgsCount("t_test", []int{1, 2}, args); err != nil {
				return err
			}
			xs, errObj := numListStatsArg("t_test", 1, args[0])
			if errObj != nil {
				return errObj
			}
			mu := 0.0
			if len(args) == 2 {
				v, errObj := numFloatArg("t_test", 2, args[1])
				if errObj != nil {
					return errObj
				}
				mu = v
			}
			tStat, p, err := tTest(xs, mu)
			if err != nil {
				return newError("`t_test` error: %s", err.Error())
			}
			return numMapObject(
				"statistic", &Float{Value: tStat},
				"p", &Float{Value: p},
			)
		},
		HelpStr: helpStrArgs{
			explanation: "`t_test` is a one sample t-test of the mean against mu, returning {statistic, p}",
			signature:   "t_test(x: list|tensor, mu: float=0.0) -> map",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "t_test([1.0, 2.0, 3.0], 0.0) => {statistic: ..., p: ...}",
		}.String(),
	},
	{
		Name: "_num_chi_square_test",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("chi_square_test", 2, args); err != nil {
				return err
			}
			obs, errObj := numListStatsArg("chi_square_test", 1, args[0])
			if errObj != nil {
				return errObj
			}
			exp, errObj := numListStatsArg("chi_square_test", 2, args[1])
			if errObj != nil {
				return errObj
			}
			chiSq, p, err := chiSquareTest(obs, exp)
			if err != nil {
				return newError("`chi_square_test` error: %s", err.Error())
			}
			return numMapObject(
				"statistic", &Float{Value: chiSq},
				"p", &Float{Value: p},
			)
		},
		HelpStr: helpStrArgs{
			explanation: "`chi_square_test` is a goodness of fit test of observed counts against expected counts, returning {statistic, p}",
			signature:   "chi_square_test(observed: list|tensor, expected: list|tensor) -> map",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "chi_square_test([10.0, 20.0], [15.0, 15.0]) => {statistic: ..., p: ...}",
		}.String(),
	},
	{
		Name: "_num_erfinv",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("erfinv", 1, args); err != nil {
				return err
			}
			x, errObj := numFloatArg("erfinv", 1, args[0])
			if errObj != nil {
				return errObj
			}
			return &Float{Value: math.Erfinv(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`erfinv` is the inverse error function",
			signature:   "erfinv(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "erfinv(0.0) => 0.0",
		}.String(),
	},
	{
		Name: "_num_gamma_fn",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("gamma_fn", 1, args); err != nil {
				return err
			}
			x, errObj := numFloatArg("gamma_fn", 1, args[0])
			if errObj != nil {
				return errObj
			}
			return &Float{Value: math.Gamma(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`gamma_fn` is the gamma function, named so it does not clash with the gamma distribution constructor",
			signature:   "gamma_fn(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "gamma_fn(5.0) => 24.0",
		}.String(),
	},
	{
		Name: "_num_lgamma",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("lgamma", 1, args); err != nil {
				return err
			}
			x, errObj := numFloatArg("lgamma", 1, args[0])
			if errObj != nil {
				return errObj
			}
			lg, _ := math.Lgamma(x)
			return &Float{Value: lg}
		},
		HelpStr: helpStrArgs{
			explanation: "`lgamma` is the natural log of the absolute value of the gamma function",
			signature:   "lgamma(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "lgamma(5.0) => 3.178",
		}.String(),
	},
	{
		Name: "_num_digamma",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("digamma", 1, args); err != nil {
				return err
			}
			x, errObj := numFloatArg("digamma", 1, args[0])
			if errObj != nil {
				return errObj
			}
			return &Float{Value: mathext.Digamma(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`digamma` is the logarithmic derivative of the gamma function, which Go's math package does not provide",
			signature:   "digamma(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "digamma(1.0) => -0.5772",
		}.String(),
	},
	{
		Name: "_num_beta_inc",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("beta_inc", 3, args); err != nil {
				return err
			}
			a, errObj := numFloatArg("beta_inc", 1, args[0])
			if errObj != nil {
				return errObj
			}
			b, errObj := numFloatArg("beta_inc", 2, args[1])
			if errObj != nil {
				return errObj
			}
			x, errObj := numFloatArg("beta_inc", 3, args[2])
			if errObj != nil {
				return errObj
			}
			return &Float{Value: mathext.RegIncBeta(a, b, x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`beta_inc` is the regularized incomplete beta function, which Go's math package does not provide",
			signature:   "beta_inc(a: float, b: float, x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "beta_inc(1.0, 1.0, 0.5) => 0.5",
		}.String(),
	},
	{
		Name: "_num_gamma_inc",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("gamma_inc", 2, args); err != nil {
				return err
			}
			a, errObj := numFloatArg("gamma_inc", 1, args[0])
			if errObj != nil {
				return errObj
			}
			x, errObj := numFloatArg("gamma_inc", 2, args[1])
			if errObj != nil {
				return errObj
			}
			return &Float{Value: mathext.GammaIncReg(a, x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`gamma_inc` is the regularized lower incomplete gamma function, which Go's math package does not provide",
			signature:   "gamma_inc(a: float, x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "gamma_inc(1.0, 1.0) => 0.6321",
		}.String(),
	},
	{
		Name: "_num_student_t",
		Fun: numDistBuiltin("student_t", func(args []Object) (*NumDist, Object) {
			if err := checkArgsCount("student_t", []int{1, 2, 3, 4}, args); err != nil {
				return nil, err
			}
			nu, errObj := numFloatArg("student_t", 1, args[0])
			if errObj != nil {
				return nil, errObj
			}
			mu, sigma := 0.0, 1.0
			if len(args) >= 2 {
				v, errObj := numFloatArg("student_t", 2, args[1])
				if errObj != nil {
					return nil, errObj
				}
				mu = v
			}
			if len(args) >= 3 {
				v, errObj := numFloatArg("student_t", 3, args[2])
				if errObj != nil {
					return nil, errObj
				}
				sigma = v
			}
			var src rand.Source
			if len(args) == 4 {
				s, errObj := numSourceArg("student_t", args[3])
				if errObj != nil {
					return nil, errObj
				}
				src = s
			}
			d, err := makeStudentT(nu, mu, sigma, src)
			if err != nil {
				return nil, newError("`student_t` error: %s", err.Error())
			}
			return d, nil
		}),
		HelpStr: helpStrArgs{
			explanation: "`student_t` returns a Student's t distribution with nu degrees of freedom; sigma is a scale, so the standard deviation is sigma*sqrt(nu/(nu-2))",
			signature:   "student_t(nu: float, mu: float=0.0, sigma: float=1.0, seed: int|null=null) -> distribution",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "student_t(10.0).cdf(0.0) => 0.5",
		}.String(),
	},
	{
		Name: "_num_chi_squared_dist",
		Fun: numDistBuiltin("chi_squared_dist", func(args []Object) (*NumDist, Object) {
			if err := checkArgsCount("chi_squared_dist", []int{1, 2}, args); err != nil {
				return nil, err
			}
			k, errObj := numFloatArg("chi_squared_dist", 1, args[0])
			if errObj != nil {
				return nil, errObj
			}
			var src rand.Source
			if len(args) == 2 {
				s, errObj := numSourceArg("chi_squared_dist", args[1])
				if errObj != nil {
					return nil, errObj
				}
				src = s
			}
			d, err := makeChiSquared(k, src)
			if err != nil {
				return nil, newError("`chi_squared_dist` error: %s", err.Error())
			}
			return d, nil
		}),
		HelpStr: helpStrArgs{
			explanation: "`chi_squared_dist` returns a chi-squared distribution with k degrees of freedom; the name is suffixed so it does not clash with chi_square_test",
			signature:   "chi_squared_dist(k: float, seed: int|null=null) -> distribution",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "chi_squared_dist(2.0).mean() => 2.0",
		}.String(),
	},
	// The following two have no Fun here on purpose: they call back into blue,
	// so the vm resolves them through GetStdBuiltinWithVm. A nil Fun is exactly
	// the marker the vm uses for a vm-bound std builtin.
	{
		Name: "_num_quad",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`quad` integrates a function over [min, max] with Gauss-Legendre quadrature",
			signature:   "quad(f: fun, min: float, max: float, n: int=100) -> float",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "quad(fun(x) { return x * x; }, 0.0, 1.0, 100) => 0.3333",
		}.String(),
	},
	{
		Name: "_num_minimize",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`minimize` minimizes a function from a starting point, returning {x, value, iterations}",
			signature:   "minimize(f: fun, x0: list, method: str='nelder-mead', max_iter: int=0) -> map",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "minimize(fun(x) { return (x[0] - 2.0) ** 2; }, [0.0]) => {x: [2.0], ...}",
		}.String(),
	},
	{
		Name: "_num_trapezoidal",
		Fun:  numIntegrateBuiltin("trapezoidal", integrateTrapezoidal),
		HelpStr: helpStrArgs{
			explanation: "`trapezoidal` integrates sampled values with the trapezoidal rule",
			signature:   "trapezoidal(x: list|tensor, y: list|tensor) -> float",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "trapezoidal([0.0, 1.0, 2.0], [0.0, 1.0, 4.0]) => 4.0",
		}.String(),
	},
	{
		Name: "_num_simpsons",
		Fun:  numIntegrateBuiltin("simpsons", integrateSimpsons),
		HelpStr: helpStrArgs{
			explanation: "`simpsons` integrates sampled values with Simpson's rule",
			signature:   "simpsons(x: list|tensor, y: list|tensor) -> float",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "simpsons([0.0, 1.0, 2.0], [0.0, 1.0, 4.0]) => 2.6666667",
		}.String(),
	},
	{
		Name: "_num_romberg",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("romberg", 2, args); err != nil {
				return err
			}
			ys, errObj := numSeriesArg("romberg", 1, args[0])
			if errObj != nil {
				return errObj
			}
			dx, errObj := numFloatArg("romberg", 2, args[1])
			if errObj != nil {
				return errObj
			}
			v, err := integrateRomberg(ys, dx)
			if err != nil {
				return newError("`romberg` error: %s", err.Error())
			}
			return &Float{Value: v}
		},
		HelpStr: helpStrArgs{
			explanation: "`romberg` integrates evenly spaced samples with Romberg's method, which needs the step size",
			signature:   "romberg(y: list|tensor, dx: float) -> float",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "romberg([0.0, 1.0, 4.0], 1.0) => 2.6666667",
		}.String(),
	},
	{
		Name: "_num_fft",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("fft", 1, args); err != nil {
				return err
			}
			seq, errObj := numSeriesArg("fft", 1, args[0])
			if errObj != nil {
				return errObj
			}
			re, im, err := fftTransform(seq)
			if err != nil {
				return newError("`fft` error: %s", err.Error())
			}
			return numMapObject(
				"real", float64ObjectList(re),
				"imag", float64ObjectList(im),
			)
		},
		HelpStr: helpStrArgs{
			explanation: "`fft` is the discrete Fourier transform, returned as {real: list, imag: list} because blue has no complex type",
			signature:   "fft(x: list|tensor) -> map",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "fft([1.0, 0.0, 0.0, 0.0]) => {real: [...], imag: [...]}",
		}.String(),
	},
	{
		Name: "_num_ifft",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("ifft", 2, args); err != nil {
				return err
			}
			re, errObj := numSeriesArg("ifft", 1, args[0])
			if errObj != nil {
				return errObj
			}
			im, errObj := numSeriesArg("ifft", 2, args[1])
			if errObj != nil {
				return errObj
			}
			out, err := ifftInverse(re, im)
			if err != nil {
				return newError("`ifft` error: %s", err.Error())
			}
			t, err := mlFloat64Tensor(out)
			if err != nil {
				return newError("`ifft` error: %s", err.Error())
			}
			return &Tensor{T: t}
		},
		HelpStr: helpStrArgs{
			explanation: "`ifft` inverts `fft`, taking the real and imaginary parts and returning the reconstructed sequence",
			signature:   "ifft(real: list|tensor, imag: list|tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "ifft(r, i) => Tensor{shape: [n]}",
		}.String(),
	},
	{
		Name: "_num_dct",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("dct", 1, args); err != nil {
				return err
			}
			seq, errObj := numSeriesArg("dct", 1, args[0])
			if errObj != nil {
				return errObj
			}
			out, err := dctTransform(seq)
			if err != nil {
				return newError("`dct` error: %s", err.Error())
			}
			t, err := mlFloat64Tensor(out)
			if err != nil {
				return newError("`dct` error: %s", err.Error())
			}
			return &Tensor{T: t}
		},
		HelpStr: helpStrArgs{
			explanation: "`dct` is the discrete cosine transform",
			signature:   "dct(x: list|tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "dct([1.0, 2.0, 3.0]) => Tensor{shape: [3]}",
		}.String(),
	},
	{
		Name: "_num_fft_freqs",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("fft_freqs", 1, args); err != nil {
				return err
			}
			n, ok := args[0].(*Integer)
			if !ok {
				return newPositionalTypeError("fft_freqs", 1, INTEGER_OBJ, args[0].Type())
			}
			out, err := fftFreqs(int(n.Value))
			if err != nil {
				return newError("`fft_freqs` error: %s", err.Error())
			}
			t, err := mlFloat64Tensor(out)
			if err != nil {
				return newError("`fft_freqs` error: %s", err.Error())
			}
			return &Tensor{T: t}
		},
		HelpStr: helpStrArgs{
			explanation: "`fft_freqs` returns the sample frequencies for a length n transform, which pair with `fft` output",
			signature:   "fft_freqs(n: int) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "fft_freqs(8) => Tensor{shape: [5]}",
		}.String(),
	},
	{
		Name: "_num_interp_linear",
		Fun: func(args ...Object) Object {
			f, errObj := numInterpFit("interp_linear", args, func(xs, ys []float64) (numInterpolator, error) {
				return interpLinear(xs, ys)
			})
			if errObj != nil {
				return errObj
			}
			return &GoObj[*numInterpHandle]{Value: f}
		},
		HelpStr: helpStrArgs{
			explanation: "`interp_linear` fits a piecewise linear interpolant; call `predict` on the result",
			signature:   "interp_linear(x: list|tensor, y: list|tensor) -> interpolant",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "interp_linear([0.0, 1.0], [0.0, 2.0]).predict(0.5) => 1.0",
		}.String(),
	},
	{
		Name: "_num_interp_cubic",
		Fun: func(args ...Object) Object {
			f, errObj := numInterpFit("interp_cubic", args, func(xs, ys []float64) (numInterpolator, error) {
				return interpCubic(xs, ys)
			})
			if errObj != nil {
				return errObj
			}
			return &GoObj[*numInterpHandle]{Value: f}
		},
		HelpStr: helpStrArgs{
			explanation: "`interp_cubic` fits a natural cubic spline; call `predict` on the result",
			signature:   "interp_cubic(x: list|tensor, y: list|tensor) -> interpolant",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "interp_cubic([0.0, 1.0, 2.0], [0.0, 1.0, 0.0]).predict(1.5) => 0.75",
		}.String(),
	},
	{
		Name: "_num_predict",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("predict", 2, args); err != nil {
				return err
			}
			g, ok := args[0].(*GoObj[*numInterpHandle])
			if !ok || g.Value == nil {
				return newPositionalTypeErrorForGoObj("predict", 1, "*numInterpHandle", args[0])
			}
			x, errObj := numFloatArg("predict", 2, args[1])
			if errObj != nil {
				return errObj
			}
			return &Float{Value: g.Value.predict(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`predict` evaluates an interpolant at x",
			signature:   "predict(f: interpolant, x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "predict(f, 0.5) => 1.0",
		}.String(),
	},
}

// The remaining helpers live in std_num_ops.go.

// parseNormKind maps a blue norm name to the numeric constant gonum expects.
//
// gonum's Norm takes a float64, where 1 is the maximum absolute column sum, 2 is
// the Frobenius norm, and +Inf is the maximum absolute row sum. There is no
// spectral norm, so '2' is documented as Frobenius rather than claimed as
// spectral.
func parseNormKind(s string) (float64, bool) {
	switch strings.ToLower(s) {
	case "fro", "frobenius", "2":
		return 2, true
	case "1":
		return 1, true
	case "inf", "infinity":
		return math.Inf(1), true
	default:
		return 0, false
	}
}

// denseNorm wraps mat.Norm, which panics on a zero-length matrix.
func denseNorm(m *mat.Dense, kind float64) float64 {
	r, c := m.Dims()
	if r == 0 || c == 0 {
		return 0
	}
	return mat.Norm(m, kind)
}

// denseRank computes the matrix rank from the singular values, counting values
// above a tolerance relative to the largest. gonum has no exported Rank.
func denseRank(m *mat.Dense) (int, bool) {
	var svd mat.SVD
	if !svd.Factorize(m, mat.SVDNone) {
		return 0, false
	}
	values := svd.Values(nil)
	if len(values) == 0 {
		return 0, true
	}
	tol := values[0] * float64(max(m.Dims())) * 2.220446049250313e-16
	rank := 0
	for _, v := range values {
		if v > tol {
			rank++
		}
	}
	return rank, true
}

// numMatrixObject builds a matrix handle.
func numMatrixObject(m *mat.Dense) Object {
	return &GoObj[*NumMatrix]{Value: m}
}

// numMapObject builds an ordered blue map from alternating key, value pairs.
func numMapObject(pairs ...any) Object {
	if len(pairs)%2 != 0 {
		return newError("numMapObject: odd number of arguments")
	}
	out := NewOrderedMap[string, Object]()
	for i := 0; i < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			return newError("numMapObject: keys must be strings")
		}
		out.Set(key, pairs[i+1].(Object))
	}
	return CreateMapObjectForGoMap(*out)
}

// numDistBuiltin builds a distribution constructor.
func numDistBuiltin(name string, f func(args []Object) (*NumDist, Object)) func(...Object) Object {
	return func(args ...Object) Object {
		d, errObj := f(args)
		if errObj != nil {
			return errObj
		}
		return &GoObj[*NumDist]{Value: d}
	}
}

// numSampleBuiltin builds a sampling function that returns a float64 1d tensor,
// so samples feed straight into ml. It accepts a count and an optional seed.
func numSampleBuiltin(name string, draw func(d *NumDist, n int) []float64, make func(src rand.Source) (*NumDist, error)) func(...Object) Object {
	return func(args ...Object) Object {
		if err := checkArgsCount(name, []int{1, 2}, args); err != nil {
			return err
		}
		n, ok := args[0].(*Integer)
		if !ok {
			return newPositionalTypeError(name, 1, INTEGER_OBJ, args[0].Type())
		}
		var src rand.Source
		if len(args) == 2 {
			s, errObj := numSourceArg(name, args[1])
			if errObj != nil {
				return errObj
			}
			src = s
		}
		d, err := make(src)
		if err != nil {
			return newError("`%s` error: %s", name, err.Error())
		}
		vs := draw(d, int(n.Value))
		t, err := mlFloat64Tensor(vs)
		if err != nil {
			return newError("`%s` error: %s", name, err.Error())
		}
		return &Tensor{T: t}
	}
}

// numInterpolator is the shared method set of gonum's interpolants.
type numInterpolator interface {
	Predict(float64) float64
}

// numInterpHandle wraps a fitted interpolant so the builtins can hold it.
type numInterpHandle struct {
	predict func(float64) float64
}

// numInterpFit validates the arguments and fits an interpolant.
func numInterpFit(name string, args []Object, fit func(xs, ys []float64) (numInterpolator, error)) (*numInterpHandle, Object) {
	if err := checkArgCount(name, 2, args); err != nil {
		return nil, err
	}
	xs, errObj := numSeriesArg(name, 1, args[0])
	if errObj != nil {
		return nil, errObj
	}
	ys, errObj := numSeriesArg(name, 2, args[1])
	if errObj != nil {
		return nil, errObj
	}
	f, err := fit(xs, ys)
	if err != nil {
		return nil, newError("`%s` error: %s", name, err.Error())
	}
	return &numInterpHandle{predict: f.Predict}, nil
}

// numIntegrateBuiltin builds a two-series integration builtin.
func numIntegrateBuiltin(name string, f func(xs, ys []float64) (float64, error)) func(...Object) Object {
	return func(args ...Object) Object {
		if err := checkArgCount(name, 2, args); err != nil {
			return err
		}
		xs, errObj := numSeriesArg(name, 1, args[0])
		if errObj != nil {
			return errObj
		}
		ys, errObj := numSeriesArg(name, 2, args[1])
		if errObj != nil {
			return errObj
		}
		v, err := f(xs, ys)
		if err != nil {
			return newError("`%s` error: %s", name, err.Error())
		}
		return &Float{Value: v}
	}
}

var _ = interp.PiecewiseLinear{}
