package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"ncobase/plugin/resource/structs"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation/validator"
)

// ImageProcessorInterface defines the interface for image processing operations
type ImageProcessorInterface interface {
	CreateThumbnail(ctx context.Context, reader io.Reader, filename string, maxWidth, maxHeight int) ([]byte, error)
	ResizeImage(ctx context.Context, reader io.Reader, filename string, maxWidth, maxHeight int) ([]byte, error)
	ProcessImage(ctx context.Context, reader io.Reader, filename string, options *structs.ProcessingOptions) ([]byte, types.JSON, error)
	GetImageDimensions(ctx context.Context, reader io.Reader, filename string) (int, int, error)
}

// imageProcessor provides image processing capabilities
type imageProcessor struct{}

// NewImageProcessor creates a new image processor service
func NewImageProcessor() ImageProcessorInterface {
	return &imageProcessor{}
}

// CreateThumbnail creates a thumbnail of an image with the specified dimensions
func (p *imageProcessor) CreateThumbnail(ctx context.Context, reader io.Reader, filename string, maxWidth, maxHeight int) ([]byte, error) {
	// Decode the image
	src, format, err := image.Decode(reader)
	if err != nil {
		logger.Errorf(ctx, "Error decoding image for thumbnail: %v", err)
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize the image to create thumbnail
	thumbnail := imaging.Fit(src, maxWidth, maxHeight, imaging.Lanczos)

	// Encode the thumbnail
	var buf bytes.Buffer
	var encodeErr error

	// Use the same format as the original
	switch format {
	case "jpeg":
		encodeErr = jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: 85})
	case "png":
		encodeErr = png.Encode(&buf, thumbnail)
	default:
		// Default to JPEG if format not explicitly supported
		encodeErr = jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: 85})
	}

	if encodeErr != nil {
		logger.Errorf(ctx, "Error encoding thumbnail: %v", encodeErr)
		return nil, fmt.Errorf("failed to encode thumbnail: %w", encodeErr)
	}

	return buf.Bytes(), nil
}

// ResizeImage resizes an image to fit within the specified dimensions while preserving aspect ratio
func (p *imageProcessor) ResizeImage(ctx context.Context, reader io.Reader, filename string, maxWidth, maxHeight int) ([]byte, error) {
	// Decode the image
	src, format, err := image.Decode(reader)
	if err != nil {
		logger.Errorf(ctx, "Error decoding image for resize: %v", err)
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize the image
	resized := imaging.Fit(src, maxWidth, maxHeight, imaging.Lanczos)

	// Encode the resized image
	var buf bytes.Buffer
	var encodeErr error

	// Use the same format as the original
	switch format {
	case "jpeg":
		encodeErr = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 85})
	case "png":
		encodeErr = png.Encode(&buf, resized)
	default:
		// Default to JPEG if format not explicitly supported
		encodeErr = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 85})
	}

	if encodeErr != nil {
		logger.Errorf(ctx, "Error encoding resized image: %v", encodeErr)
		return nil, fmt.Errorf("failed to encode resized image: %w", encodeErr)
	}

	return buf.Bytes(), nil
}

// ProcessImage processes an image according to the provided options
func (p *imageProcessor) ProcessImage(ctx context.Context, reader io.Reader, filename string, options *structs.ProcessingOptions) ([]byte, types.JSON, error) {
	// Read all bytes from the reader
	fileBytes, err := io.ReadAll(reader)
	if err != nil {
		logger.Errorf(ctx, "Error reading image bytes: %v", err)
		return nil, nil, err
	}

	// Create a reader from the bytes
	r := bytes.NewReader(fileBytes)

	// Get image dimensions before processing
	src, format, err := image.Decode(r)
	if err != nil {
		logger.Errorf(ctx, "Error decoding image: %v", err)
		return nil, nil, err
	}

	// Reset reader position
	r.Seek(0, io.SeekStart)

	// Get original dimensions
	bounds := src.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	// Create metadata to return
	metadata := types.JSON{
		"width":         origWidth,
		"height":        origHeight,
		"original_size": len(fileBytes),
		"format":        format,
	}

	// Apply processing according to options
	var processedImage = src

	// Resize if needed
	if options.ResizeImage && (options.MaxWidth > 0 || options.MaxHeight > 0) {
		maxWidth := options.MaxWidth
		if maxWidth <= 0 {
			maxWidth = origWidth
		}

		maxHeight := options.MaxHeight
		if maxHeight <= 0 {
			maxHeight = origHeight
		}

		processedImage = imaging.Fit(processedImage, maxWidth, maxHeight, imaging.Lanczos)

		// Update metadata
		newBounds := processedImage.Bounds()
		metadata["width"] = newBounds.Dx()
		metadata["height"] = newBounds.Dy()
		metadata["resized"] = true
	}

	// Encode the processed image
	var buf bytes.Buffer
	var encodeErr error

	// Convert format if specified
	outputFormat := format
	if options.ConvertFormat != "" {
		outputFormat = strings.ToLower(strings.TrimPrefix(options.ConvertFormat, "."))
		metadata["converted_format"] = outputFormat
	}

	// Set quality for compression
	quality := 85 // Default quality
	if options.CompressImage && options.CompressionQuality > 0 && options.CompressionQuality <= 100 {
		quality = options.CompressionQuality
		metadata["compressed"] = true
		metadata["compression_quality"] = quality
	}

	// Encode based on format
	switch outputFormat {
	case "jpeg", "jpg":
		encodeErr = jpeg.Encode(&buf, processedImage, &jpeg.Options{Quality: quality})
	case "png":
		encodeErr = png.Encode(&buf, processedImage)
	default:
		// Default to JPEG
		encodeErr = jpeg.Encode(&buf, processedImage, &jpeg.Options{Quality: quality})
		outputFormat = "jpeg"
	}

	if encodeErr != nil {
		logger.Errorf(ctx, "Error encoding processed image: %v", encodeErr)
		return nil, metadata, fmt.Errorf("failed to encode processed image: %w", encodeErr)
	}

	// Update final metadata
	metadata["processed_size"] = buf.Len()
	metadata["compression_ratio"] = float64(buf.Len()) / float64(len(fileBytes))

	return buf.Bytes(), metadata, nil
}

// GetImageDimensions extracts dimensions from an image
func (p *imageProcessor) GetImageDimensions(ctx context.Context, reader io.Reader, filename string) (int, int, error) {
	if !validator.IsImageFile(filename) {
		return 0, 0, errors.New("not an image file")
	}

	// Decode the image header to get dimensions
	img, _, err := image.DecodeConfig(reader)
	if err != nil {
		logger.Errorf(ctx, "Error decoding image dimensions: %v", err)
		return 0, 0, err
	}

	return img.Width, img.Height, nil
}
