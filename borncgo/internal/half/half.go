// Package half converts the 16-bit floating-point storage encodings used by
// model files — IEEE 754 half-precision (binary16) and bfloat16 — to float32.
// Born has no native half-precision dtype, so format readers widen on load.
package half

import "math"

// Float16ToFloat32 converts an IEEE 754 half-precision (binary16) value to
// float32, handling subnormals, infinities, and NaNs.
func Float16ToFloat32(h uint16) float32 {
	// Extract sign, exponent, and mantissa.
	sign := (h >> 15) & 0x1
	exp := (h >> 10) & 0x1F
	mant := h & 0x3FF

	var result uint32

	switch exp {
	case 0:
		if mant == 0 {
			// Signed zero.
			result = uint32(sign) << 31
		} else {
			// Subnormal: value = (-1)^sign * 2^(-14) * (mant / 1024). Compute in
			// float64 to avoid uint16 underflow while shifting into normal range.
			f := float64(mant) / 1024.0 * math.Pow(2, -14)
			if sign == 1 {
				f = -f
			}
			return float32(f)
		}
	case 0x1F:
		// Infinity or NaN.
		result = (uint32(sign) << 31) | 0x7F800000 | (uint32(mant) << 13)
	default:
		// Normal: re-bias the exponent from 15 to 127 and widen the mantissa.
		result = (uint32(sign) << 31) | (uint32(exp+127-15) << 23) | (uint32(mant) << 13)
	}

	return math.Float32frombits(result)
}

// BFloat16ToFloat32 converts a bfloat16 value to float32. bfloat16 is the top
// 16 bits of the IEEE 754 float32 representation, so widening is a left shift.
func BFloat16ToFloat32(b uint16) float32 {
	return math.Float32frombits(uint32(b) << 16)
}
