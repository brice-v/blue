# MNIST CNN Classification (LeNet-5 Style)

Demonstrates **convolutional neural networks** using Born ML Framework's Conv2D and MaxPool2D operations.

## Architecture

**MNISTNetCNN** - LeNet-5 inspired architecture:

```
Input: [batch, 1, 28, 28]
  ↓
Conv2D(1→6, 5x5) → [batch, 6, 24, 24]
ReLU
MaxPool2D(2x2, stride=2) → [batch, 6, 12, 12]
  ↓
Conv2D(6→16, 5x5) → [batch, 16, 8, 8]
ReLU
MaxPool2D(2x2, stride=2) → [batch, 16, 4, 4]
  ↓
Flatten → [batch, 256]
  ↓
Linear(256→120)
ReLU
Linear(120→84)
ReLU
Linear(84→10)
  ↓
Output: [batch, 10] (logits)
```

**Parameters**: 44,426 (vs MLP: 101,770)

## Features Demonstrated

### Convolutional Operations
- **Conv2D**: 2D convolution with:
  - Multi-channel input/output
  - Configurable kernel size, stride, padding
  - Xavier initialization
  - Bias support
  - Full autodiff integration

### Pooling Operations
- **MaxPool2D**: Max pooling with:
  - Configurable kernel size and stride
  - Max index tracking for gradients
  - Gradient routing to max positions only
  - No learnable parameters

### End-to-End CNN Training
- Convolutional feature extraction
- Spatial downsampling via pooling
- Flatten transition to fully connected
- Multi-layer classification head
- CrossEntropyLoss with softmax
- Adam optimizer with bias correction

## Quick Start

```bash
# Synthetic data (for testing)
go run . -synthetic -epochs 5

# Real MNIST data
go run . -data ./data -epochs 10 -batch 64 -lr 0.001
```

## Implementation Highlights

### Conv2D Backend (CPU)
```go
// Im2col algorithm for efficient convolution
func conv2dFloat32(output, input, kernel *tensor.RawTensor, ...) {
    // Step 1: Transform patches to columns
    colBuf := make([]float32, colHeight*colWidth)
    im2colFloat32(colBuf, inputData, ...)

    // Step 2: Matrix multiplication (reuse optimized matmul)
    // Step 3: Rearrange output
}
```

### MaxPool2D Gradient Routing
```go
// Forward: Track max indices
maxIndices := computeMaxIndices(input, output, kernelSize, stride)

// Backward: Route gradients only to max positions
for outIdx, maxPos := range maxIndices {
    inputGradData[maxPos] += outputGradData[outIdx]
}
```

### Bias Broadcasting with ReshapeOp
```go
// Reshape bias for broadcasting: [out_channels] → [1, out_channels, 1, 1]
biasReshaped := bias.Reshape(1, outChannels, 1, 1)
output = output.Add(biasReshaped)  // ReshapeOp records for gradient flow!
```

## Performance

Typical results on MNIST (5 epochs):
- **Training Accuracy**: ~98%
- **Validation Accuracy**: ~97-98%
- **Training Time**: ~2-3 minutes (CPU, single-threaded)

CNN benefits:
- ✓ Spatial invariance via weight sharing
- ✓ Local feature detection
- ✓ Fewer parameters than MLP
- ✓ Better generalization on image data

## Comparison with MLP

| Metric | CNN (LeNet-5) | MLP (2-layer) |
|--------|---------------|---------------|
| Parameters | 44,426 | 101,770 |
| Validation Acc | ~97-98% | ~97% |
| Architecture | Convolutional | Fully Connected |
| Inductive Bias | Spatial locality | None |

## References

- **LeNet-5**: LeCun et al., 1998 - "Gradient-Based Learning Applied to Document Recognition"
- **Im2col**: Chellapilla et al., 2006 - "High Performance Convolutional Neural Networks for Document Processing"

## Files

- `model.go` - MNISTNetCNN architecture
- `main.go` - Training loop
- `data.go` - MNIST data loading
- `idx_reader.go` - IDX file format parser

## Next Steps

- Add BatchNorm2D for improved training stability
- Implement data augmentation (rotation, translation)
- Try deeper architectures (ResNet-style)
- Benchmark GPU backend performance
