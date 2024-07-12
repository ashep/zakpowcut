package state

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func Save(ss []string) error {
	fp, err := os.Create("tmp/state.json")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	b, err := json.Marshal(ss)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	if _, err = fp.Write(b); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if err = fp.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	return nil
}

func Get() ([]string, error) {
	fp, err := os.Open("tmp/state.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	b, err := io.ReadAll(fp)
	if err != nil {
		return nil, fmt.Errorf("failed to read from file: %w", err)
	}

	if err = fp.Close(); err != nil {
		return nil, fmt.Errorf("failed to close file: %w", err)
	}

	var res []string
	err = json.Unmarshal(b, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return res, nil
}
