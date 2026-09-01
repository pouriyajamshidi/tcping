package printers

import (
	"path/filepath"
	"testing"
)

func TestNewPrinter(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		cfg         PrinterConfig
		wantErr     bool
		expectedErr string
	}{
		{
			name: "JSON Printer Initialization",
			cfg: PrinterConfig{
				OutputJSON: true,
			},
			wantErr: false,
		},
		{
			name: "Database Printer Initialization",
			cfg: PrinterConfig{
				OutputDBPath: ":memory:",
				Target:       "example.com",
				Port:         443,
			},
			wantErr: false,
		},
		{
			name: "CSV Printer Initialization",
			cfg: PrinterConfig{
				OutputCSVPath: filepath.Join(tempDir, "test.csv"),
			},
			wantErr: false,
		},
		{
			name: "Plain Printer Initialization",
			cfg: PrinterConfig{
				NoColor: true,
			},
			wantErr: false,
		},
		{
			name: "Default Color Printer Initialization",
			cfg: PrinterConfig{
				NoColor: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printer, err := NewPrinter(tt.cfg)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewPrinter() expected an error, got nil")
				} else if err.Error() != tt.expectedErr {
					t.Errorf("NewPrinter() expected error %q, got %q", tt.expectedErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("NewPrinter() unexpected error: %v", err)
			}

			if printer == nil {
				t.Errorf("NewPrinter() returned nil printer for valid config")
			}

			if csvPrinter, ok := printer.(*CSVPrinter); ok {
				t.Cleanup(csvPrinter.Done)
			}
		})
	}
}
