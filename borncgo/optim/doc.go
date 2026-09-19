// Copyright 2025 Born ML Framework. All rights reserved.
// Use of this source code is governed by an Apache 2.0
// license that can be found in the LICENSE file.

// Package optim provides optimization algorithms for training neural networks.
//
// # Overview
//
// This package contains:
//   - SGD: Stochastic Gradient Descent with momentum
//   - Adam: Adaptive Moment Estimation with bias correction
//   - Optimizer interface for custom optimizers
//
// # Basic Usage
//
//	import (
//	    "blue/borncgo/optim"
//	    "blue/borncgo/nn"
//	    "blue/borncgo/backend/cpu"
//	)
//
//	func main() {
//	    backend := cpu.New()
//	    model := nn.NewLinear(784, 10, backend)
//
//	    // Create optimizer
//	    optimizer := optim.NewAdam(
//	        model.Parameters(),
//	        optim.AdamConfig{
//	            LR:    0.001,
//	            Betas: [2]float32{0.9, 0.999},
//	        },
//	        backend,
//	    )
//
//	    // Training loop
//	    for epoch := range 10 {
//	        // Forward pass
//	        loss := criterion.Forward(model.Forward(x), y)
//
//	        // Backward pass
//	        optimizer.ZeroGrad()
//	        grads := backend.Backward(loss.Raw())
//	        optimizer.Step(grads)
//	    }
//	}
//
// # Optimizers
//
// SGD (Stochastic Gradient Descent):
//
//	optimizer := optim.NewSGD(
//	    model.Parameters(),
//	    optim.SGDConfig{
//	        LR:       0.01,
//	        Momentum: 0.9,
//	    },
//	    backend,
//	)
//
// Adam (Adaptive Moment Estimation):
//
//	optimizer := optim.NewAdam(
//	    model.Parameters(),
//	    optim.AdamConfig{
//	        LR:      0.001,
//	        Betas:   [2]float32{0.9, 0.999},
//	        Epsilon: 1e-8,
//	    },
//	    backend,
//	)
//
// # Training Loop Pattern
//
//	for epoch := range numEpochs {
//	    for batch := range dataLoader {
//	        // 1. Zero gradients
//	        optimizer.ZeroGrad()
//
//	        // 2. Forward pass
//	        output := model.Forward(batch.Input)
//	        loss := criterion.Forward(output, batch.Target)
//
//	        // 3. Backward pass
//	        grads := backend.Backward(loss.Raw())
//
//	        // 4. Update parameters
//	        optimizer.Step(grads)
//	    }
//	}
package optim
