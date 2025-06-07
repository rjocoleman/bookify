package services

import (
	"archive/zip"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pgaskin/kepubify/v4/kepub"
)

type ProcessorService struct{}

func NewProcessorService() *ProcessorService {
	return &ProcessorService{}
}

func (p *ProcessorService) ProcessEPUB(inputPath, outputPath string, progressCallback func(int)) error {
	if progressCallback != nil {
		progressCallback(10)
	}

	// Check if input file exists and get its info
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("input file does not exist: %w", err)
	}
	log.Printf("Processing EPUB: %s (size: %d bytes)", inputPath, info.Size())

	// Open EPUB as ZIP archive (since EPUB is a ZIP file)
	zipReader, err := zip.OpenReader(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open EPUB as ZIP: %w", err)
	}
	defer func() {
		if closeErr := zipReader.Close(); closeErr != nil {
			log.Printf("Warning: Failed to close ZIP reader: %v", closeErr)
		}
	}()

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		if closeErr := outputFile.Close(); closeErr != nil {
			log.Printf("Warning: Failed to close output file: %v", closeErr)
		}
	}()

	if progressCallback != nil {
		progressCallback(50)
	}

	converter := kepub.NewConverter()
	err = converter.Convert(context.Background(), outputFile, &zipReader.Reader)
	if err != nil {
		return fmt.Errorf("kepubify conversion failed: %w", err)
	}

	if progressCallback != nil {
		progressCallback(100)
	}

	return nil
}

func (p *ProcessorService) CleanFilename(filename string) string {
	clean := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`).ReplaceAllString(filename, "_")

	lowerClean := strings.ToLower(clean)
	if strings.HasSuffix(lowerClean, ".epub") {
		// Find the last occurrence of .epub (case insensitive) and replace it
		lastIndex := strings.LastIndex(lowerClean, ".epub")
		if lastIndex != -1 {
			return clean[:lastIndex] + ".kepub.epub"
		}
	}
	return clean + ".kepub.epub"
}

func (p *ProcessorService) PrepareOutputPath(tempDir, originalFilename string) (string, error) {
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	cleanName := p.CleanFilename(originalFilename)
	return filepath.Join(tempDir, cleanName), nil
}
