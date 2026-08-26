package util

import (
	"math"
	"math/bits"
)

func CheckOverflow(leftVal, rightVal int64) bool {
	sum := leftVal + rightVal
	return (leftVal > 0 && rightVal > 0 && sum < 0) ||
		(leftVal < 0 && rightVal < 0 && sum >= 0)
}

func CheckUnderflow(leftVal, rightVal int64) bool {
	diff := leftVal - rightVal
	return (rightVal > 0 && diff > leftVal) ||
		(rightVal < 0 && diff < leftVal)
}

func CheckOverflowMul(leftVal, rightVal int64) bool {
	if leftVal == 0 || rightVal == 0 || leftVal == 1 || rightVal == 1 {
		return false
	}
	if leftVal == math.MinInt64 || rightVal == math.MinInt64 {
		return true
	}
	result := leftVal * rightVal
	return result/rightVal != leftVal
}

func CheckOverflowPow(leftVal, rightVal int64) bool {
	switch {
	case leftVal == 0 || rightVal == 0 || leftVal == 1 || rightVal == 1:
		return false
	case leftVal == math.MinInt64 || rightVal == math.MinInt64:
		return true
	case rightVal < 0:
		return false
	}
	negBase := leftVal < 0
	base := uint64(leftVal)
	if negBase {
		base = -base
	}
	const limit = uint64(1) << 63
	result := uint64(1)
	for i := uint64(0); i < uint64(rightVal); i++ {
		hi, lo := bits.Mul64(result, base)
		if hi != 0 {
			return true
		}
		result = lo
		if result >= limit && i < uint64(rightVal)-1 {
			return true
		}
	}
	if result > math.MaxInt64 {
		return !negBase || rightVal%2 != 1 || result != limit
	}
	return false
}
