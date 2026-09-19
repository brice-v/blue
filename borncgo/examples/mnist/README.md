# MNIST Classification Example

This example demonstrates Born ML framework's end-to-end training capability using the MNIST handwritten digit classification task.

## Phase 1 Status (Proof-of-Concept)

This is a **Phase 1 proof-of-concept** that demonstrates framework integration:

✅ **Working Components:**
- Tensor API with shape handling and broadcasting
- CPU Backend for tensor operations
- NN Modules (Linear layers, ReLU activation)
- CrossEntropyLoss with numerical stability (log-sum-exp trick)
- Training loop structure
- Accuracy computation

⚠️ **Phase 1 Limitations:**
- Uses synthetic data (not real MNIST)
- Simplified parameter updates (not full gradient descent)
- Manual backward pass (CrossEntropyLoss not integrated with autodiff tape)

## Running the Example

### Quick Demo (Synthetic Data)

```bash
cd examples/mnist
go run .
```

This runs a proof-of-concept with 10 synthetic samples to verify all components work together.

### Expected Output

```
🚀 Born ML Framework - MNIST Classification Example
...
🎓 Starting training...
Epoch  1/20: Loss=2.3815, Train Acc=0.00%, Val Acc=0.00%
...
✅ Training complete!
```

**Note:** 0% accuracy is expected with synthetic data and dummy parameter updates.

## Architecture

### Model: MNISTNet

Simple fully-connected neural network:

```
Input:  784 neurons (28×28 flattened image)
Hidden: 128 neurons (ReLU activation)
Output: 10 neurons (logits for digits 0-9)
```

### Training Configuration

- **Loss Function:** CrossEntropyLoss (LogSoftmax + NLLLoss)
- **Numerical Stability:** log-sum-exp trick for preventing overflow/underflow
- **Learning Rate:** 0.01 (fixed, no schedule)
- **Epochs:** 20

## Code Structure

```
examples/mnist/
├── main.go      # Training loop and integration
├── model.go     # MNISTNet architecture (784→128→10)
├── data.go      # Data loading utilities
└── README.md    # This file
```

## Key Implementation Details

### Numerical Stability

CrossEntropyLoss uses the **log-sum-exp trick** to prevent overflow:

```go
// LogSoftmax(z) = z - (max(z) + log(Σ exp(z - max(z))))
func logSoftmax(z []float32) []float32 {
    maxZ := max(z)
    sumExp := sum(exp(z - maxZ))
    logSumExp := maxZ + log(sumExp)
    return z - logSumExp
}
```

This ensures stable computation even with extreme logit values (z > 88 or z < -88).

### Gradient Formula

CrossEntropyLoss backward pass uses the elegant closed-form solution:

```
∂L/∂logits = softmax(logits) - y_one_hot
```

This avoids computing the full Jacobian matrix.

## Phase 2 Roadmap

For real MNIST training with >90% accuracy, Phase 2 will add:

### 1. Full Autodiff Integration

- Extend autodiff tape to support Softmax/Log operations
- Automatic backward pass through entire network
- Remove manual gradient computation

### 2. Production Optimizers

- Adam optimizer with bias correction (already implemented in TASK-006)
- SGD with momentum
- Learning rate scheduling (cosine annealing, step decay)

### 3. Real Data Loading

**Option 1: CSV Format (Quick Start)**
```bash
# Download Kaggle MNIST CSV
# https://www.kaggle.com/datasets/oddrationale/mnist-in-csv

go run . mnist_train.csv
```

**Option 2: IDX Format (Official)**
```go
import "github.com/petar/GoMNIST"

train, test, _ := GoMNIST.Load("./data")
```

### 4. Training Improvements

- Mini-batch training (batch size 32-64)
- Data shuffling each epoch
- Proper train/val split
- Dropout regularization (0.2-0.5)
- Weight initialization (He/Xavier)

## Expected Performance (Phase 2)

| Architecture | Epochs | Expected Accuracy |
|--------------|--------|-------------------|
| 784-128-10 (simple) | 10 | 97.0-97.5% |
| 784-256-128-10 | 10 | 98.0-98.5% |
| 784-256-128-10 + Dropout | 15 | 98.2-98.7% |

Training time: ~1-2 minutes on modern CPU

## Troubleshooting

### "Loss is NaN"

The log-sum-exp trick should prevent this. If it occurs:
- Check input normalization (should be [0, 1], not [0, 255])
- Verify logits are finite before loss computation
- Ensure weights initialized properly (not zeros)

### "Accuracy stuck at 10%" (random guessing)

This indicates no learning. Check:
- Gradients are not all zeros
- Learning rate is not too small (try 0.001-0.01)
- Forward pass computes correct predictions

### "Accuracy stuck at 30-50%"

Partial learning. Verify:
- Data normalization is correct
- Labels are class indices (0-9), not one-hot
- Loss is decreasing over epochs

## References

### Mathematical Background

- **Cross-Entropy Loss:** [MLDawn Academy](https://www.mldawn.com/back-propagation-with-cross-entropy-and-softmax/)
- **Log-Sum-Exp Trick:** [Gregory Gundersen Blog](https://gregorygundersen.com/blog/2020/02/09/log-sum-exp/)
- **Softmax Gradient:** [Medium Article](https://medium.com/data-science/derivative-of-the-softmax-function-and-the-categorical-cross-entropy-loss-ffceefc081d1)

### Framework Documentation

- **PyTorch CrossEntropyLoss:** [Official Docs](https://pytorch.org/docs/stable/generated/torch.nn.CrossEntropyLoss.html)
- **Burn Training Guide:** Reference in research/crossentropy_mnist_research.md

### Go ML Libraries

- **GoMNIST:** [GitHub](https://github.com/petar/GoMNIST) - Official MNIST IDX format loader
- **Gonum:** [pkg.go.dev](https://pkg.go.dev/gonum.org/v1/gonum) - Numerical computing

## License

Born ML Framework is licensed under [MIT License](../../LICENSE).

## Contributing

This is Phase 1 proof-of-concept. Contributions welcome for Phase 2:

1. Full autodiff integration for CrossEntropyLoss
2. Real MNIST data loader
3. Proper training with Adam optimizer
4. Achieving >90% accuracy

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

---

**Generated with Born ML Framework v0.1-alpha**
*Built with Go 1.25+ and modern ML best practices (2025)*
