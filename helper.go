package main

import (
	"fmt"
	"os"
)

// ANSI  color codes
var colors = map[string]string{
	"black":   "\033[30m",
	"red":     "\033[31m",
	"green":   "\033[32m",
	"yellow":  "\033[33m",
	"blue":    "\033[34m",
	"magenta": "\033[35m",
	"cyan":    "\033[36m",
	"white":   "\033[37m",
	"reset":   "\033[0m",
}

// ReadAndPrintFile reads a txt file and prints it in color
func ReadAndPrintFile(path string, color string) error {
	// Read file contents
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Get color code
	colorCode, exists := colors[color]
	if !exists {
		colorCode = colors["reset"]
	}

	// Print colored text
	fmt.Print(colorCode)
	fmt.Print(string(data))
	fmt.Print(colors["reset"])

	return nil
}
