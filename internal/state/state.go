package state

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func Save(st map[string][]string) error {
	fp, err := os.Create("tmp/state.json")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	b, err := json.Marshal(st)
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

func Get() (map[string][]string, error) {
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

	res := make(map[string][]string)
	if err = json.Unmarshal(b, &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return res, nil
}
