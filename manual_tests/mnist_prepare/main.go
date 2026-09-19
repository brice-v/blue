// Command mnist_prepare converts the raw MNIST IDX files into the little
// tensor format that ml.save/ml.load use, so the blue manual test can train
// without a data loader in the language.
//
// The IDX files live in borncgo/examples/mnist/data. This writes
// manual_tests/mnist_data/{train,test}_{images,labels}.bin.
//
// Usage (from the repo root):
//
//	go run ./manual_tests/mnist_prepare
//	go run ./manual_tests/mnist_prepare -train 60000 -test 10000
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const magic = "BLUETNSR1"

func main() {
	dataDir := flag.String("data", "borncgo/examples/mnist/data", "directory with the IDX files")
	outDir := flag.String("out", "manual_tests/mnist_data", "directory to write .bin tensors")
	nTrain := flag.Int("train", 6000, "number of training examples")
	nTest := flag.Int("test", 1000, "number of test examples")
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fatal(err)
	}

	trX, trY := loadSplit(*dataDir, "train", *nTrain)
	teX, teY := loadSplit(*dataDir, "t10k", *nTest)

	writeTensor(filepath.Join(*outDir, "train_images.bin"), []int{len(trY), 784}, trX)
	writeTensor(filepath.Join(*outDir, "train_labels.bin"), []int{len(trY)}, trY)
	writeTensor(filepath.Join(*outDir, "test_images.bin"), []int{len(teY), 784}, teX)
	writeTensor(filepath.Join(*outDir, "test_labels.bin"), []int{len(teY)}, teY)

	fmt.Printf("wrote %d train and %d test examples to %s\n", len(trY), len(teY), *outDir)
}

// loadSplit reads the image and label IDX files for a split and returns
// images (row-major, scaled to [0,1]) and labels as float32.
func loadSplit(dir, split string, limit int) ([]float32, []float32) {
	images := readIDX(filepath.Join(dir, split+"-images-idx3-ubyte"))
	labels := readIDX(filepath.Join(dir, split+"-labels-idx1-ubyte"))

	n := len(labels)
	if limit > 0 && limit < n {
		n = limit
	}
	imagesPer := len(images) / len(labels)

	outImages := make([]float32, n*imagesPer)
	for i := 0; i < n*imagesPer; i++ {
		outImages[i] = float32(images[i]) / 255.0
	}
	outLabels := make([]float32, n)
	for i := 0; i < n; i++ {
		outLabels[i] = float32(labels[i])
	}
	return outImages, outLabels
}

// readIDX parses an IDX file: 4 magic bytes, a big-endian uint32 dimension per
// remaining header word, then the payload bytes.
func readIDX(path string) []byte {
	raw, err := os.ReadFile(path)
	if err != nil {
		fatal(err)
	}
	if len(raw) < 4 {
		fatal(fmt.Errorf("%s: too short", path))
	}
	// The low byte of the magic is the number of following dimension words;
	// IDX data is always uint8, and the dimensions are big-endian uint32.
	ndim := int(raw[3])
	offset := 4 + ndim*4
	if offset > len(raw) {
		fatal(fmt.Errorf("%s: malformed header", path))
	}
	return raw[offset:]
}

// writeTensor writes the same format as ml.Save.
func writeTensor(path string, shape []int, data []float32) {
	f, err := os.Create(path)
	if err != nil {
		fatal(err)
	}
	defer func() { _ = f.Close() }()

	must(binary.Write(f, binary.LittleEndian, []byte(magic)))
	must(binary.Write(f, binary.LittleEndian, uint8(0))) // dtype 0 = Float32
	must(binary.Write(f, binary.LittleEndian, uint32(len(shape))))
	for _, d := range shape {
		must(binary.Write(f, binary.LittleEndian, int64(d)))
	}
	must(binary.Write(f, binary.LittleEndian, data))
}

func must(err error) {
	if err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "mnist_prepare:", err)
	os.Exit(1)
}
