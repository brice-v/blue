package tensor

import internal "blue/borncgo/internal/tensor"

// SetSeed seeds the shared random generator so Randn/Rand are reproducible.
func SetSeed(seed int64) {
	internal.SetSeed(seed)
}
