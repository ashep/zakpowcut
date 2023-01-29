package state

import (
	"fmt"
	"io"
	"os"
	"time"
)

func Save(t time.Time) error {
	fp, err := os.Create("tmp/state.json")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	b, err := t.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	if _, err = fp.Write(b); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	if err = fp.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	return nil
}

func Get() (time.Time, error) {
	var r time.Time

	fp, err := os.Open("tmp/state.json")
	if err != nil {
		return r, fmt.Errorf("failed to open file: %w", err)
	}

	b, err := io.ReadAll(fp)
	if err != nil {
		return r, fmt.Errorf("failed to read from file: %w", err)
	}

	if err = r.UnmarshalJSON(b); err != nil {
		return r, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	if err = fp.Close(); err != nil {
		return r, fmt.Errorf("failed to close file: %w", err)
	}

	return r, nil
}
