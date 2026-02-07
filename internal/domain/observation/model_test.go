package observation_test

import (
	"testing"
	"time"

	"workshop/internal/domain/observation"
)

// TestObservation_Validate tests validation of Observation.
func TestObservation_Validate(t *testing.T) {
	tests := []struct {
		name    string
		obs     observation.Observation
		wantErr bool
	}{
		{
			name:    "valid observation",
			obs:     observation.Observation{ID: "1", MemberID: "m1", AuthorID: "coach1", Content: "Good guard retention", CreatedAt: time.Now()},
			wantErr: false,
		},
		{
			name:    "empty member ID",
			obs:     observation.Observation{ID: "2", AuthorID: "coach1", Content: "note", CreatedAt: time.Now()},
			wantErr: true,
		},
		{
			name:    "empty author ID",
			obs:     observation.Observation{ID: "3", MemberID: "m1", Content: "note", CreatedAt: time.Now()},
			wantErr: true,
		},
		{
			name:    "empty content",
			obs:     observation.Observation{ID: "4", MemberID: "m1", AuthorID: "coach1", CreatedAt: time.Now()},
			wantErr: true,
		},
		{
			name:    "zero created_at",
			obs:     observation.Observation{ID: "5", MemberID: "m1", AuthorID: "coach1", Content: "note"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.obs.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
