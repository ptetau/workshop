package classtype_test

import (
	"testing"

	"workshop/internal/domain/classtype"
)

// TestClassType_Validate tests validation of ClassType.
func TestClassType_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ct      classtype.ClassType
		wantErr bool
	}{
		{
			name:    "valid class type",
			ct:      classtype.ClassType{ID: "1", ProgramID: "prog-1", Name: "Fundamentals"},
			wantErr: false,
		},
		{
			name:    "empty name",
			ct:      classtype.ClassType{ID: "2", ProgramID: "prog-1", Name: ""},
			wantErr: true,
		},
		{
			name:    "whitespace name",
			ct:      classtype.ClassType{ID: "3", ProgramID: "prog-1", Name: "   "},
			wantErr: true,
		},
		{
			name:    "empty program ID",
			ct:      classtype.ClassType{ID: "4", ProgramID: "", Name: "No-Gi"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ct.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ClassType.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
