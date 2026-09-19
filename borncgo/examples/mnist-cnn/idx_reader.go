package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// readIDXImages reads MNIST image file in IDX format.
//
// IDX file format for images:
//
//	magic number: 0x00000803 (2051)
//	number of images: 4 bytes
//	number of rows: 4 bytes (28)
//	number of cols: 4 bytes (28)
//	pixel data: unsigned bytes (0-255)
func readIDXImages(filename string) ([][]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read magic number
	var magic uint32
	if err := binary.Read(file, binary.BigEndian, &magic); err != nil {
		return nil, fmt.Errorf("failed to read magic: %w", err)
	}
	if magic != 2051 {
		return nil, fmt.Errorf("invalid magic number: got %d, want 2051", magic)
	}

	// Read dimensions
	var numImages, numRows, numCols uint32
	if err := binary.Read(file, binary.BigEndian, &numImages); err != nil {
		return nil, err
	}
	if err := binary.Read(file, binary.BigEndian, &numRows); err != nil {
		return nil, err
	}
	if err := binary.Read(file, binary.BigEndian, &numCols); err != nil {
		return nil, err
	}

	imageSize := int(numRows * numCols)
	images := make([][]byte, numImages)

	// Read all images
	for i := range images {
		images[i] = make([]byte, imageSize)
		if _, err := io.ReadFull(file, images[i]); err != nil {
			return nil, fmt.Errorf("failed to read image %d: %w", i, err)
		}
	}

	return images, nil
}

// readIDXLabels reads MNIST label file in IDX format.
//
// IDX file format for labels:
//
//	magic number: 0x00000801 (2049)
//	number of labels: 4 bytes
//	label data: unsigned bytes (0-9)
func readIDXLabels(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read magic number
	var magic uint32
	if err := binary.Read(file, binary.BigEndian, &magic); err != nil {
		return nil, fmt.Errorf("failed to read magic: %w", err)
	}
	if magic != 2049 {
		return nil, fmt.Errorf("invalid magic number: got %d, want 2049", magic)
	}

	// Read number of labels
	var numLabels uint32
	if err := binary.Read(file, binary.BigEndian, &numLabels); err != nil {
		return nil, err
	}

	// Read all labels
	labels := make([]byte, numLabels)
	if _, err := io.ReadFull(file, labels); err != nil {
		return nil, fmt.Errorf("failed to read labels: %w", err)
	}

	return labels, nil
}
