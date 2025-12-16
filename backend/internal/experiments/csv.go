package experiments

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
)

type CSVWriter struct {
	file   *os.File
	writer *csv.Writer
}

// NewCSVWriter creates a CSV writer. If appendFile is false, it truncates and writes header.
// If appendFile is true, it creates the file if needed and writes header only when the file is empty.
func NewCSVWriter(path string, header []string, appendFile bool) (*CSVWriter, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("mkdir: %w", err)
	}

	flags := os.O_CREATE | os.O_WRONLY
	if appendFile {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	f, err := os.OpenFile(path, flags, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open csv: %w", err)
	}

	w := csv.NewWriter(f)

	needHeader := !appendFile
	if appendFile {
		if st, statErr := f.Stat(); statErr == nil {
			needHeader = st.Size() == 0
		}
	}
	if needHeader && len(header) > 0 {
		if err := w.Write(header); err != nil {
			f.Close()
			return nil, fmt.Errorf("write header: %w", err)
		}
		w.Flush()
		if err := w.Error(); err != nil {
			f.Close()
			return nil, fmt.Errorf("flush header: %w", err)
		}
	}

	return &CSVWriter{file: f, writer: w}, nil
}

func (w *CSVWriter) Write(record []string) error {
	if err := w.writer.Write(record); err != nil {
		return err
	}
	w.writer.Flush()
	return w.writer.Error()
}

func (w *CSVWriter) Close() error {
	w.writer.Flush()
	if err := w.writer.Error(); err != nil {
		w.file.Close()
		return err
	}
	return w.file.Close()
}
