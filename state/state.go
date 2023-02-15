package state

import (
	"fmt"
	"io"
	"os"
)

func Save(s string) error {
	fp, err := os.Create("tmp/state.json")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	if _, err = fp.Write([]byte(s)); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	if err = fp.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	return nil
}

func Get() (string, error) {
	fp, err := os.Open("tmp/state.json")
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}

	b, err := io.ReadAll(fp)
	if err != nil {
		return "", fmt.Errorf("failed to read from file: %w", err)
	}

	if err = fp.Close(); err != nil {
		return "", fmt.Errorf("failed to close file: %w", err)
	}

	return string(b), nil
}
