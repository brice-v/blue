package object

import (
	"blue/ml"
	"gonum.org/v1/gonum/mat"
	"math"
	rand "math/rand/v2"
	"testing"
)

func testDense(t *testing.T, rows [][]float64) *mat.Dense {
	t.Helper()
	r := len(rows)
	if r == 0 {
		return mat.NewDense(0, 0, nil)
	}
	c := len(rows[0])
	data := make([]float64, 0, r*c)
	for _, row := range rows {
		if len(row) != c {
			t.Fatalf("ragged row: %v", rows)
		}
		data = append(data, row...)
	}
	return mat.NewDense(r, c, data)
}

func equalMatrix(a, b *mat.Dense) bool {
	ar, ac := a.Dims()
	br, bc := b.Dims()
	if ar != br || ac != bc {
		return false
	}
	for i := range ar {
		for j := range ac {
			if a.At(i, j) != b.At(i, j) {
				return false
			}
		}
	}
	return true
}

// closeMatrix compares two matrices within a tolerance, for results that come
// out of an iterative solver.
func closeMatrix(a, b *mat.Dense, tol float64) bool {
	ar, ac := a.Dims()
	br, bc := b.Dims()
	if ar != br || ac != bc {
		return false
	}
	for i := range ar {
		for j := range ac {
			d := a.At(i, j) - b.At(i, j)
			if d < 0 {
				d = -d
			}
			if d > tol {
				return false
			}
		}
	}
	return true
}

func TestNumTensorToMatrix(t *testing.T) {
	// A rank 2 tensor maps directly.
	mx := testDense(t, [][]float64{{1, 2, 3}, {4, 5, 6}})
	obj := &Tensor{T: newTensorForTest(t, []float32{1, 2, 3, 4, 5, 6}, []int{2, 3})}
	got, errObj := numMatrixArg("test", 1, obj)
	if errObj != nil {
		t.Fatalf("numMatrixArg: %v", errObj.Inspect())
	}
	if !equalMatrix(got, mx) {
		t.Fatalf("tensor to matrix mismatch: got %v", mat.Formatted(got))
	}
}

func TestNumTensorRank1IsRow(t *testing.T) {
	obj := &Tensor{T: newTensorForTest(t, []float32{1, 2, 3}, []int{3})}
	got, errObj := numMatrixArg("test", 1, obj)
	if errObj != nil {
		t.Fatalf("numMatrixArg: %v", errObj.Inspect())
	}
	r, c := got.Dims()
	if r != 1 || c != 3 {
		t.Fatalf("rank 1 tensor became %dx%d, want 1x3", r, c)
	}
}

func TestNumTensorRank3Flattens(t *testing.T) {
	obj := &Tensor{T: newTensorForTest(t, []float32{1, 2, 3, 4, 5, 6, 7, 8}, []int{2, 2, 2})}
	got, errObj := numMatrixArg("test", 1, obj)
	if errObj != nil {
		t.Fatalf("numMatrixArg: %v", errObj.Inspect())
	}
	r, c := got.Dims()
	if r != 2 || c != 4 {
		t.Fatalf("rank 3 tensor became %dx%d, want 2x4", r, c)
	}
}

func TestNumListToMatrix(t *testing.T) {
	l := &List{Elements: []Object{
		&List{Elements: []Object{&Float{Value: 1}, &Float{Value: 2}}},
		&List{Elements: []Object{&Float{Value: 3}, &Float{Value: 4}}},
	}}
	got, errObj := numMatrixArg("test", 1, l)
	if errObj != nil {
		t.Fatalf("numMatrixArg: %v", errObj.Inspect())
	}
	if !equalMatrix(got, testDense(t, [][]float64{{1, 2}, {3, 4}})) {
		t.Fatalf("list to matrix mismatch: %v", mat.Formatted(got))
	}
}

func TestNumMatrixRoundTrip(t *testing.T) {
	// A matrix to tensor to matrix round trip is exact, because the f64 path
	// does not go through float32.
	mx := testDense(t, [][]float64{{1.25, 2.5}, {3.75, 4.125}})
	t64, err := ml.NewFloat64Tensor(matrixToFloat64Data(mx), []int{2, 2}, ml.CPU)
	if err != nil {
		t.Fatalf("NewFloat64Tensor: %v", err)
	}
	back, err := tensorToMatrix(t64)
	if err != nil {
		t.Fatalf("tensorToMatrix: %v", err)
	}
	if !equalMatrix(back, mx) {
		t.Fatalf("round trip lost data: %v", mat.Formatted(back))
	}
}

func TestNumSolve(t *testing.T) {
	a := testDense(t, [][]float64{{2, 1}, {1, 3}})
	b := testDense(t, [][]float64{{3}, {5}})
	x, err := denseSolve(a, b)
	if err != nil {
		t.Fatalf("denseSolve: %v", err)
	}
	if !closeMatrix(x, testDense(t, [][]float64{{0.8}, {1.4}}), 1e-9) {
		t.Fatalf("solve = %v", mat.Formatted(x))
	}
}

func TestNumDet(t *testing.T) {
	got, err := denseDet(testDense(t, [][]float64{{1, 2}, {3, 4}}))
	if err != nil {
		t.Fatalf("denseDet: %v", err)
	}
	if got != -2 {
		t.Fatalf("det = %v, want -2", got)
	}
}

func TestNumRank(t *testing.T) {
	// A rank deficient matrix: the second row is twice the first.
	got, ok := denseRank(testDense(t, [][]float64{{1, 2}, {2, 4}}))
	if !ok {
		t.Fatal("denseRank did not converge")
	}
	if got != 1 {
		t.Fatalf("rank = %d, want 1", got)
	}
}

func TestNumCholesky(t *testing.T) {
	// The Cholesky factor of [[4,2],[2,3]] is [[2,0],[1,sqrt(2)]].
	got, err := denseCholesky(testDense(t, [][]float64{{4, 2}, {2, 3}}))
	if err != nil {
		t.Fatalf("denseCholesky: %v", err)
	}
	if got.At(0, 0) != 2 || got.At(1, 0) != 1 || got.At(0, 1) != 0 {
		t.Fatalf("cholesky = %v", mat.Formatted(got))
	}
}

func TestNumCholeskyRejectsNonPD(t *testing.T) {
	if _, err := denseCholesky(testDense(t, [][]float64{{1, 2}, {2, 1}})); err == nil {
		t.Fatal("cholesky should reject a non positive definite matrix")
	}
}

func TestNumSVDValues(t *testing.T) {
	obj := denseSVDObject(testDense(t, [][]float64{{3, 0}, {0, 2}}))
	m, ok := obj.(*Map)
	if !ok {
		t.Fatalf("svd did not return a map: %s", obj.Inspect())
	}
	hk := HashKey{Type: STRING_OBJ, Value: HashObject(&Stringo{Value: "values"})}
	pair, found := m.Pairs.Get(hk)
	if !found {
		t.Fatal("svd map has no values key")
	}
	l, ok := pair.Value.(*List)
	if !ok || len(l.Elements) != 2 {
		t.Fatalf("svd values is not a 2 element list")
	}
	first, _ := l.Elements[0].(*Float)
	if first == nil || first.Value != 3 {
		t.Fatalf("largest singular value = %v, want 3", pair.Value.Inspect())
	}
}

// newTensorForTest builds a float32 tensor, failing the test on error.
func newTensorForTest(t *testing.T, data []float32, shape []int) *ml.Tensor {
	t.Helper()
	tensor, err := ml.NewTensor(data, shape, ml.Float32, ml.CPU)
	if err != nil {
		t.Fatalf("ml.NewTensor: %v", err)
	}
	return tensor
}

func closeTo(a, b, tol float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= tol
}

func TestNumNormalKnownValues(t *testing.T) {
	d, err := makeNormal(0, 1, nil)
	if err != nil {
		t.Fatalf("makeNormal: %v", err)
	}
	// The true value is 0.97500210..., not exactly 0.975.
	if !closeTo(d.CDF(1.96), 0.9750021048517795, 1e-12) {
		t.Fatalf("cdf(1.96) = %v, want 0.9750021048517795", d.CDF(1.96))
	}
	if !closeTo(d.Quantile(0.5), 0, 1e-9) {
		t.Fatalf("median = %v, want 0", d.Quantile(0.5))
	}
	if !closeTo(d.Mean(), 0, 1e-9) {
		t.Fatalf("mean = %v, want 0", d.Mean())
	}
	if !closeTo(d.Variance(), 1, 1e-9) {
		t.Fatalf("variance = %v, want 1", d.Variance())
	}
}

func TestNumNormalRejectsBadSigma(t *testing.T) {
	if _, err := makeNormal(0, 0, nil); err == nil {
		t.Fatal("makeNormal should reject a non-positive sigma")
	}
}

func TestNumDistributionMeans(t *testing.T) {
	u, err := makeUniform(0, 2, nil)
	if err != nil {
		t.Fatalf("makeUniform: %v", err)
	}
	if !closeTo(u.Mean(), 1, 1e-9) {
		t.Fatalf("uniform mean = %v, want 1", u.Mean())
	}

	e, err := makeExponential(2, nil)
	if err != nil {
		t.Fatalf("makeExponential: %v", err)
	}
	if !closeTo(e.Mean(), 0.5, 1e-9) {
		t.Fatalf("exponential mean = %v, want 0.5", e.Mean())
	}

	g, err := makeGamma(2, 1, nil)
	if err != nil {
		t.Fatalf("makeGamma: %v", err)
	}
	if !closeTo(g.Mean(), 2, 1e-9) {
		t.Fatalf("gamma mean = %v, want 2", g.Mean())
	}

	b, err := makeBeta(2, 2, nil)
	if err != nil {
		t.Fatalf("makeBeta: %v", err)
	}
	if !closeTo(b.Mean(), 0.5, 1e-9) {
		t.Fatalf("beta mean = %v, want 0.5", b.Mean())
	}

	p, err := makePoisson(3, nil)
	if err != nil {
		t.Fatalf("makePoisson: %v", err)
	}
	if !closeTo(p.Mean(), 3, 1e-9) {
		t.Fatalf("poisson mean = %v, want 3", p.Mean())
	}

	bi, err := makeBinomial(10, 0.5, nil)
	if err != nil {
		t.Fatalf("makeBinomial: %v", err)
	}
	if !closeTo(bi.Mean(), 5, 1e-9) {
		t.Fatalf("binomial mean = %v, want 5", bi.Mean())
	}
}

// TestNumDiscreteDistributionsHaveNoQuantile documents the gonum limitation:
// Poisson and Binomial expose no Quantile, so the handle stores nil and the
// builtin reports it rather than returning a silently wrong number.
func TestNumDiscreteDistributionsHaveNoQuantile(t *testing.T) {
	p, err := makePoisson(3, nil)
	if err != nil {
		t.Fatalf("makePoisson: %v", err)
	}
	if p.Quantile != nil {
		t.Fatal("poisson should not claim a quantile")
	}
	bi, err := makeBinomial(10, 0.5, nil)
	if err != nil {
		t.Fatalf("makeBinomial: %v", err)
	}
	if bi.Quantile != nil {
		t.Fatal("binomial should not claim a quantile")
	}
}

func TestNumSeededSamplingIsReproducible(t *testing.T) {
	a, err := makeNormal(0, 1, numSourceForTest(42))
	if err != nil {
		t.Fatalf("makeNormal: %v", err)
	}
	b, err := makeNormal(0, 1, numSourceForTest(42))
	if err != nil {
		t.Fatalf("makeNormal: %v", err)
	}
	xa := a.Sample(5)
	xb := b.Sample(5)
	for i := range xa {
		if xa[i] != xb[i] {
			t.Fatalf("seeded samples differ at %d: %v vs %v", i, xa[i], xb[i])
		}
	}
}

func TestNumHistogram(t *testing.T) {
	counts, edges, err := histogramOf([]float64{1, 2, 3, 4}, 2)
	if err != nil {
		t.Fatalf("histogramOf: %v", err)
	}
	if len(counts) != 2 || counts[0] != 2 || counts[1] != 2 {
		t.Fatalf("counts = %v, want [2 2]", counts)
	}
	if len(edges) != 3 || edges[0] != 1 || edges[2] != 4 {
		t.Fatalf("edges = %v, want [1 2.5 4]", edges)
	}
}

func TestNumHistogramConstantSeries(t *testing.T) {
	// A constant series must not divide by zero.
	counts, _, err := histogramOf([]float64{5, 5, 5}, 3)
	if err != nil {
		t.Fatalf("histogramOf: %v", err)
	}
	total := 0.0
	for _, c := range counts {
		total += c
	}
	if total != 3 {
		t.Fatalf("constant series lost samples: %v", counts)
	}
}

// TestNumTTestKnownValue checks the one sample t-test against the hand-computed
// statistic for [1, 2, 3, 4]: mean 2.5, variance 5/3, se = sqrt(5/12).
func TestNumTTestKnownValue(t *testing.T) {
	tStat, p, err := tTest([]float64{1, 2, 3, 4}, 0)
	if err != nil {
		t.Fatalf("tTest: %v", err)
	}
	wantT := 2.5 / math.Sqrt((5.0/3.0)/4.0)
	if !closeTo(tStat, wantT, 1e-9) {
		t.Fatalf("t = %v, want %v", tStat, wantT)
	}
	if p <= 0 || p >= 1 {
		t.Fatalf("p = %v, want a probability", p)
	}
}

func TestNumTTestRejectsSmallSamples(t *testing.T) {
	if _, _, err := tTest([]float64{1}, 0); err == nil {
		t.Fatal("tTest should need at least 2 samples")
	}
}

// TestNumChiSquareKnownValue checks the goodness of fit statistic by hand:
// (10-15)^2/15 + (20-15)^2/15 = 10/3.
func TestNumChiSquareKnownValue(t *testing.T) {
	chiSq, p, err := chiSquareTest([]float64{10, 20}, []float64{15, 15})
	if err != nil {
		t.Fatalf("chiSquareTest: %v", err)
	}
	if !closeTo(chiSq, 10.0/3.0, 1e-9) {
		t.Fatalf("chi-square = %v, want %v", chiSq, 10.0/3.0)
	}
	if p <= 0 || p >= 1 {
		t.Fatalf("p = %v, want a probability", p)
	}
}

func TestNumChiSquareRejectsMismatch(t *testing.T) {
	if _, _, err := chiSquareTest([]float64{1, 2}, []float64{1}); err == nil {
		t.Fatal("chiSquareTest should reject a bin count mismatch")
	}
}

// numSourceForTest builds a seeded rand/v2 source.
func numSourceForTest(seed uint64) rand.Source {
	return rand.NewPCG(seed, 0)
}

func TestNumTrapezoidal(t *testing.T) {
	// The trapezoidal rule on y = x^2 sampled at 0, 1, 2 gives 3.
	v, err := integrateTrapezoidal([]float64{0, 1, 2}, []float64{0, 1, 4})
	if err != nil {
		t.Fatalf("integrateTrapezoidal: %v", err)
	}
	if !closeTo(v, 3, 1e-12) {
		t.Fatalf("trapezoidal = %v, want 3", v)
	}
}

func TestNumSimpsons(t *testing.T) {
	// Simpson's rule on the same data gives 8/3, which is exact for a quadratic.
	v, err := integrateSimpsons([]float64{0, 1, 2}, []float64{0, 1, 4})
	if err != nil {
		t.Fatalf("integrateSimpsons: %v", err)
	}
	if !closeTo(v, 8.0/3.0, 1e-12) {
		t.Fatalf("simpsons = %v, want 2.6667", v)
	}
}

func TestNumIntegrationRejectsMismatch(t *testing.T) {
	if _, err := integrateTrapezoidal([]float64{0, 1}, []float64{0}); err == nil {
		t.Fatal("trapezoidal should reject a length mismatch")
	}
	if _, err := integrateSimpsons([]float64{0}, []float64{0, 1}); err == nil {
		t.Fatal("simpsons should reject a length mismatch")
	}
}

func TestNumFFTRoundTrip(t *testing.T) {
	seq := []float64{1, 2, 3, 4}
	re, im, err := fftTransform(seq)
	if err != nil {
		t.Fatalf("fftTransform: %v", err)
	}
	// The half spectrum of a length 4 input has 3 bins.
	if len(re) != 3 || len(im) != 3 {
		t.Fatalf("half spectrum has %d/%d bins, want 3/3", len(re), len(im))
	}
	if !closeTo(re[0], 10, 1e-12) {
		t.Fatalf("DC bin = %v, want 10", re[0])
	}
	back, err := ifftInverse(re, im)
	if err != nil {
		t.Fatalf("ifftInverse: %v", err)
	}
	if len(back) != 4 {
		t.Fatalf("inverse has %d values, want 4", len(back))
	}
	// gonum's transform pair is unnormalized, so the round trip is n times the
	// input. That is documented behavior, not a bug.
	for i := range seq {
		want := seq[i] * 4
		if !closeTo(back[i], want, 1e-9) {
			t.Fatalf("inverse[%d] = %v, want %v", i, back[i], want)
		}
	}
}

func TestNumFFTRejectsEmpty(t *testing.T) {
	if _, _, err := fftTransform(nil); err == nil {
		t.Fatal("fft should reject an empty sequence")
	}
}

func TestNumFFTFreqsMatchesSpectrum(t *testing.T) {
	// The frequency list must line up with the half spectrum fft returns.
	re, _, err := fftTransform(make([]float64, 8))
	if err != nil {
		t.Fatalf("fftTransform: %v", err)
	}
	freqs, err := fftFreqs(8)
	if err != nil {
		t.Fatalf("fftFreqs: %v", err)
	}
	if len(freqs) != len(re) {
		t.Fatalf("freqs has %d entries, spectrum has %d", len(freqs), len(re))
	}
	if !closeTo(freqs[1], 0.125, 1e-12) {
		t.Fatalf("freqs[1] = %v, want 0.125", freqs[1])
	}
}

func TestNumDCT(t *testing.T) {
	v, err := dctTransform([]float64{1, 2, 3, 4})
	if err != nil {
		t.Fatalf("dctTransform: %v", err)
	}
	if len(v) != 4 {
		t.Fatalf("dct has %d values, want 4", len(v))
	}
	if !closeTo(v[0], 15, 1e-9) {
		t.Fatalf("dct[0] = %v, want 15", v[0])
	}
}

func TestNumInterpLinear(t *testing.T) {
	f, err := interpLinear([]float64{0, 1, 2}, []float64{0, 1, 4})
	if err != nil {
		t.Fatalf("interpLinear: %v", err)
	}
	if !closeTo(f.Predict(0.5), 0.5, 1e-9) {
		t.Fatalf("linear at 0.5 = %v, want 0.5", f.Predict(0.5))
	}
	if !closeTo(f.Predict(1.5), 2.5, 1e-9) {
		t.Fatalf("linear at 1.5 = %v, want 2.5", f.Predict(1.5))
	}
	// The knots are recovered exactly.
	for i, x := range []float64{0, 1, 2} {
		want := []float64{0, 1, 4}[i]
		if !closeTo(f.Predict(x), want, 1e-9) {
			t.Fatalf("linear at knot %v = %v, want %v", x, f.Predict(x), want)
		}
	}
}

func TestNumInterpCubicPassesThroughKnots(t *testing.T) {
	f, err := interpCubic([]float64{0, 1, 2}, []float64{0, 1, 4})
	if err != nil {
		t.Fatalf("interpCubic: %v", err)
	}
	for i, x := range []float64{0, 1, 2} {
		want := []float64{0, 1, 4}[i]
		if !closeTo(f.Predict(x), want, 1e-9) {
			t.Fatalf("cubic at knot %v = %v, want %v", x, f.Predict(x), want)
		}
	}
}

func TestNumInterpRejectsTooFewPoints(t *testing.T) {
	if _, err := interpLinear([]float64{0}, []float64{0}); err == nil {
		t.Fatal("interpLinear should need at least 2 points")
	}
	if _, err := interpCubic([]float64{0}, []float64{0}); err == nil {
		t.Fatal("interpCubic should need at least 2 points")
	}
}

func TestNumRomberg(t *testing.T) {
	v, err := integrateRomberg([]float64{0, 1, 4}, 1)
	if err != nil {
		t.Fatalf("integrateRomberg: %v", err)
	}
	if !closeTo(v, 8.0/3.0, 1e-9) {
		t.Fatalf("romberg = %v, want 2.6667", v)
	}
	if _, err := integrateRomberg([]float64{0, 1}, 0); err == nil {
		t.Fatal("romberg should reject dx = 0")
	}
}

func TestNumFFTParsevalSanity(t *testing.T) {
	// A pure DC signal has all its energy in bin 0.
	re, im, err := fftTransform([]float64{2, 2, 2, 2})
	if err != nil {
		t.Fatalf("fftTransform: %v", err)
	}
	if !closeTo(re[0], 8, 1e-9) {
		t.Fatalf("DC bin = %v, want 8", re[0])
	}
	for i := 1; i < len(re); i++ {
		if math.Abs(re[i]) > 1e-9 || math.Abs(im[i]) > 1e-9 {
			t.Fatalf("bin %d of a DC signal is nonzero: %v %v", i, re[i], im[i])
		}
	}
}
