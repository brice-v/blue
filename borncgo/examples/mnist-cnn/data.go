package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"blue/borncgo/internal/tensor"
)

// MNISTData holds a batch of MNIST images and labels.
type MNISTData struct {
	Images [][]float32 // [num_samples, 784]
	Labels []int32     // [num_samples]
}

// LoadMNISTCSV loads MNIST data from CSV file.
//
// CSV Format (Kaggle-style):
//
//	label,pixel0,pixel1,...,pixel783
//	5,0,0,12,...,0
//	0,0,0,0,...,0
//
// Parameters:
//   - filename: Path to CSV file
//   - maxSamples: Maximum number of samples to load (0 = load all)
//
// Returns:
//   - MNISTData with images normalized to [0, 1] range
//
// Note: For production use, consider github.com/petar/GoMNIST library
// which handles the official IDX binary format.
func LoadMNISTCSV(filename string, maxSamples int) (*MNISTData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file is empty or missing header")
	}

	// Skip header row
	records = records[1:]

	// Limit samples if requested
	numSamples := len(records)
	if maxSamples > 0 && numSamples > maxSamples {
		numSamples = maxSamples
		records = records[:numSamples]
	}

	images := make([][]float32, numSamples)
	labels := make([]int32, numSamples)

	for i, record := range records {
		if len(record) != 785 { // 1 label + 784 pixels
			return nil, fmt.Errorf("invalid record length at row %d: got %d, want 785", i+1, len(record))
		}

		// Parse label
		label, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, fmt.Errorf("invalid label at row %d: %w", i+1, err)
		}
		if label < 0 || label > 9 {
			return nil, fmt.Errorf("label out of range [0, 9] at row %d: %d", i+1, label)
		}
		labels[i] = int32(label)

		// Parse pixels and normalize to [0, 1]
		images[i] = make([]float32, 784)
		for j := 0; j < 784; j++ {
			pixel, err := strconv.Atoi(record[j+1])
			if err != nil {
				return nil, fmt.Errorf("invalid pixel at row %d, column %d: %w", i+1, j+1, err)
			}
			// Normalize: 0-255 → 0.0-1.0
			images[i][j] = float32(pixel) / 255.0
		}
	}

	return &MNISTData{
		Images: images,
		Labels: labels,
	}, nil
}

// LoadMNIST loads MNIST data from official IDX binary files.
//
// This function loads the official MNIST dataset in IDX format using a
// custom IDX reader (see idx_reader.go).
//
// Parameters:
//   - dataDir: Directory containing MNIST files (train-images-idx3-ubyte, etc.)
//   - train: If true, load training set (60,000 samples), else test set (10,000 samples)
//   - maxSamples: Maximum number of samples to load (0 = load all)
//
// Returns:
//   - MNISTData with images normalized to [0, 1] range
//
// Expected files in dataDir:
//   - train-images-idx3-ubyte (or t10k-images-idx3-ubyte for test)
//   - train-labels-idx1-ubyte (or t10k-labels-idx1-ubyte for test)
//
// Download MNIST from: http://yann.lecun.com/exdb/mnist/
func LoadMNIST(dataDir string, train bool, maxSamples int) (*MNISTData, error) {
	var imageFile, labelFile string

	if train {
		imageFile = filepath.Join(dataDir, "train-images-idx3-ubyte")
		labelFile = filepath.Join(dataDir, "train-labels-idx1-ubyte")
	} else {
		imageFile = filepath.Join(dataDir, "t10k-images-idx3-ubyte")
		labelFile = filepath.Join(dataDir, "t10k-labels-idx1-ubyte")
	}

	// Load images using custom IDX reader
	imagesRaw, err := readIDXImages(imageFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load images: %w", err)
	}

	// Load labels using custom IDX reader
	labelsRaw, err := readIDXLabels(labelFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load labels: %w", err)
	}

	if len(imagesRaw) != len(labelsRaw) {
		return nil, fmt.Errorf("image count (%d) != label count (%d)", len(imagesRaw), len(labelsRaw))
	}

	numSamples := len(imagesRaw)
	if maxSamples > 0 && numSamples > maxSamples {
		numSamples = maxSamples
	}

	images := make([][]float32, numSamples)
	labels := make([]int32, numSamples)

	for i := 0; i < numSamples; i++ {
		// Convert image bytes to float32 and normalize to [0, 1]
		images[i] = make([]float32, 784)
		for j := 0; j < 784; j++ {
			// Each pixel is 0-255, normalize to [0, 1]
			images[i][j] = float32(imagesRaw[i][j]) / 255.0
		}

		labels[i] = int32(labelsRaw[i])
	}

	return &MNISTData{
		Images: images,
		Labels: labels,
	}, nil
}

// CreateEmbeddedMNIST creates a tiny embedded MNIST dataset for testing.
//
// This generates synthetic data for demonstration when no CSV file is available.
// Contains 10 simple patterns (one for each digit 0-9).
//
// Returns:
//   - MNISTData with 10 synthetic samples
func CreateEmbeddedMNIST() *MNISTData {
	numSamples := 10
	images := make([][]float32, numSamples)
	labels := make([]int32, numSamples)

	// Create simple synthetic patterns for each digit
	for i := 0; i < numSamples; i++ {
		images[i] = make([]float32, 784)
		labels[i] = int32(i) // digit 0-9

		// Create a simple pattern: fill a region based on digit value
		// This is NOT realistic MNIST data, just for testing the pipeline
		startRow := i * 2 // 0, 2, 4, ..., 18
		for row := startRow; row < startRow+8 && row < 28; row++ {
			for col := 5; col < 23; col++ {
				idx := row*28 + col
				if idx < 784 {
					images[i][idx] = 0.8 // bright pixels
				}
			}
		}
	}

	return &MNISTData{
		Images: images,
		Labels: labels,
	}
}

// Batch represents a mini-batch for training.
type Batch[B tensor.Backend] struct {
	ImagesTensor *tensor.Tensor[float32, B]
	LabelsTensor *tensor.Tensor[int32, B]
	Size         int
}

// CreateBatches splits MNIST data into mini-batches.
//
// Parameters:
//   - data: MNIST dataset
//   - batchSize: Size of each mini-batch
//   - shuffle: Whether to shuffle data before batching
//   - backend: Tensor backend to use
//
// Returns:
//   - Slice of batches (last batch may be smaller if data doesn't divide evenly)
func CreateBatches[B tensor.Backend](
	data *MNISTData,
	batchSize int,
	shuffle bool,
	backend B,
) ([]*Batch[B], error) {
	numSamples := len(data.Images)
	if numSamples != len(data.Labels) {
		return nil, fmt.Errorf("images and labels length mismatch")
	}

	// Create indices for shuffling
	indices := make([]int, numSamples)
	for i := range indices {
		indices[i] = i
	}

	// Shuffle if requested
	if shuffle {
		// Simple Fisher-Yates shuffle
		for i := numSamples - 1; i > 0; i-- {
			j := int(float32(i+1) * float32(len(indices)) / float32(len(indices))) // deterministic for now
			if j > i {
				j = i
			}
			indices[i], indices[j] = indices[j], indices[i]
		}
	}

	// Calculate number of batches
	numBatches := (numSamples + batchSize - 1) / batchSize
	batches := make([]*Batch[B], 0, numBatches)

	for i := 0; i < numSamples; i += batchSize {
		end := i + batchSize
		if end > numSamples {
			end = numSamples
		}
		currentBatchSize := end - i

		// Allocate batch tensors
		imagesTensorRaw, err := tensor.NewRaw(
			tensor.Shape{currentBatchSize, 784},
			tensor.Float32,
			backend.Device(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create images tensor: %w", err)
		}

		labelsTensorRaw, err := tensor.NewRaw(
			tensor.Shape{currentBatchSize},
			tensor.Int32,
			backend.Device(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create labels tensor: %w", err)
		}

		// Copy data to tensors
		imagesData := imagesTensorRaw.AsFloat32()
		labelsData := labelsTensorRaw.AsInt32()

		for j := i; j < end; j++ {
			idx := indices[j]
			// Copy image
			copy(imagesData[(j-i)*784:(j-i+1)*784], data.Images[idx])
			// Copy label
			labelsData[j-i] = data.Labels[idx]
		}

		batches = append(batches, &Batch[B]{
			ImagesTensor: tensor.New[float32, B](imagesTensorRaw, backend),
			LabelsTensor: tensor.New[int32, B](labelsTensorRaw, backend),
			Size:         currentBatchSize,
		})
	}

	return batches, nil
}

// NumSamples returns the total number of samples in the dataset.
func (d *MNISTData) NumSamples() int {
	return len(d.Images)
}

// Split splits the dataset into train and validation sets.
//
// Parameters:
//   - validationRatio: Fraction of data to use for validation (e.g., 0.2 for 20%)
//
// Returns:
//   - trainData, validationData
func (d *MNISTData) Split(validationRatio float32) (*MNISTData, *MNISTData) {
	numSamples := d.NumSamples()
	splitIdx := int(float32(numSamples) * (1.0 - validationRatio))

	return &MNISTData{
		Images: d.Images[:splitIdx],
		Labels: d.Labels[:splitIdx],
	}, &MNISTData{
		Images: d.Images[splitIdx:],
		Labels: d.Labels[splitIdx:],
	}
}
