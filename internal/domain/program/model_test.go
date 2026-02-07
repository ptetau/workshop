package program_test

import (
	"testing"

	"workshop/internal/domain/program"
)

// TestProgram_Validate tests validation of Program.
func TestProgram_Validate(t *testing.T) {
	tests := []struct {
		name    string
		prog    program.Program
		wantErr bool
	}{
		{
			name:    "valid adults program",
			prog:    program.Program{ID: "1", Name: "Adults BJJ", Type: program.TypeAdults},
			wantErr: false,
		},
		{
			name:    "valid kids program",
			prog:    program.Program{ID: "2", Name: "Kids BJJ", Type: program.TypeKids},
			wantErr: false,
		},
		{
			name:    "empty name",
			prog:    program.Program{ID: "3", Name: "", Type: program.TypeAdults},
			wantErr: true,
		},
		{
			name:    "whitespace name",
			prog:    program.Program{ID: "4", Name: "   ", Type: program.TypeAdults},
			wantErr: true,
		},
		{
			name:    "invalid type",
			prog:    program.Program{ID: "5", Name: "Senior", Type: "senior"},
			wantErr: true,
		},
		{
			name:    "empty type",
			prog:    program.Program{ID: "6", Name: "Test", Type: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.prog.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Program.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
