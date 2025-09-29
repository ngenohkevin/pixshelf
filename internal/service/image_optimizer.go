package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

type ImageOptimizer struct {
	cachePath string
}

func NewImageOptimizer(cachePath string) *ImageOptimizer {
	// Ensure cache directory exists
	os.MkdirAll(cachePath, 0755)
	return &ImageOptimizer{
		cachePath: cachePath,
	}
}

func (o *ImageOptimizer) GetOrCreateVariant(originalPath string, width int) (string, error) {
	// Check if variant exists
	variantPath := o.getVariantPath(originalPath, width)
	if _, err := os.Stat(variantPath); err == nil {
		return variantPath, nil
	}

	// Ensure variant directory exists
	variantDir := filepath.Dir(variantPath)
	if err := os.MkdirAll(variantDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create variant directory: %w", err)
	}

	// Create on-demand (only common sizes)
	src, err := imaging.Open(originalPath)
	if err != nil {
		return "", fmt.Errorf("failed to open image: %w", err)
	}

	// Resize maintaining aspect ratio
	resized := imaging.Resize(src, width, 0, imaging.Lanczos)

	// Save with 85% quality
	err = imaging.Save(resized, variantPath, imaging.JPEGQuality(85))
	if err != nil {
		return "", fmt.Errorf("failed to save variant: %w", err)
	}

	return variantPath, nil
}

func (o *ImageOptimizer) getVariantPath(originalPath string, width int) string {
	// Extract filename without extension
	base := filepath.Base(originalPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	// Create variant filename
	variantName := fmt.Sprintf("%s_%dw%s", name, width, ext)

	// Create subdirectory structure in cache
	relDir := filepath.Dir(strings.TrimPrefix(originalPath, "/"))
	if strings.HasPrefix(relDir, "static/images") {
		relDir = strings.TrimPrefix(relDir, "static/images")
	}

	return filepath.Join(o.cachePath, relDir, variantName)
}
