/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package fileprocessor

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/schnauzersoft/ami-util/internal/aws"
)

const (
	FilePerm = 0o600
)

type Processor struct {
	verbose bool
}

func NewProcessor(verbose bool) *Processor {
	return &Processor{
		verbose: verbose,
	}
}

func (p *Processor) ProcessFile(filePath string, replacements []aws.AMIReplacement) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	originalContent := string(content)
	newContent, replaceCount := aws.ReplaceAMIsInContent(originalContent, replacements)

	if replaceCount == 0 {
		if p.verbose {
			log.Printf("No AMI replacements needed in %s", filePath)
		}

		return nil
	}

	backupPath := filePath + ".backup"

	err = os.WriteFile(backupPath, content, FilePerm)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}

	err = os.WriteFile(filePath, []byte(newContent), FilePerm)
	if err != nil {
		return fmt.Errorf("failed to write updated file: %w", err)
	}

	log.Printf("Updated %d AMI references in %s (backup created at %s)", replaceCount, filePath, backupPath)

	return nil
}

func (p *Processor) ProcessDirectory(dirPath string, replacements []aws.AMIReplacement) error {
	files, err := p.collectFiles(dirPath)
	if err != nil {
		return err
	}

	totalReplacements := p.processFiles(files, replacements)

	log.Printf("Total AMI replacements made: %d across %d files", totalReplacements, len(files))

	return nil
}

func (p *Processor) FindAMIsInFile(filePath string) ([]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return aws.ExtractAMIPatterns(string(content)), nil
}

func (p *Processor) collectFiles(dirPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || strings.HasSuffix(path, ".backup") {
			return nil
		}

		if p.isTextFile(path) {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", dirPath, err)
	}

	return files, nil
}

func (p *Processor) processFiles(files []string, replacements []aws.AMIReplacement) int {
	totalReplacements := 0

	for _, file := range files {
		replaceCount, err := p.processSingleFile(file, replacements)
		if err != nil {
			log.Printf("Warning: failed to process file %s: %v", file, err)

			continue
		}

		totalReplacements += replaceCount
	}

	return totalReplacements
}

func (p *Processor) processSingleFile(file string, replacements []aws.AMIReplacement) (int, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}

	originalContent := string(content)
	newContent, replaceCount := aws.ReplaceAMIsInContent(originalContent, replacements)

	if replaceCount > 0 {
		err := p.updateFileWithBackup(file, content, newContent)
		if err != nil {
			return 0, err
		}

		log.Printf("Updated %d AMI references in %s (backup created at %s)", replaceCount, file, file+".backup")
	} else if p.verbose {
		log.Printf("No AMI replacements needed in %s", file)
	}

	return replaceCount, nil
}

func (p *Processor) updateFileWithBackup(file string, originalContent []byte, newContent string) error {
	backupPath := file + ".backup"

	err := os.WriteFile(backupPath, originalContent, FilePerm)
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	err = os.WriteFile(file, []byte(newContent), FilePerm)
	if err != nil {
		return fmt.Errorf("failed to write updated file: %w", err)
	}

	return nil
}

func (p *Processor) isTextFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	textExtensions := map[string]bool{
		".yaml": true,
		".yml":  true,
		".json": true,
		".txt":  true,
		".tf":   true,
		".hcl":  true,
		".conf": true,
		".cfg":  true,
		".ini":  true,
		".env":  true,
		".sh":   true,
		".bash": true,
		".zsh":  true,
		".fish": true,
		".ps1":  true,
		".bat":  true,
		".cmd":  true,
		".xml":  true,
		".toml": true,
	}

	if textExtensions[ext] {
		return true
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	amiRegex := regexp.MustCompile(`ami-[a-f0-9]{8,17}`)

	return amiRegex.Match(content)
}
