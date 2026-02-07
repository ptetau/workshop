package milestone_test

import (
	"testing"

	"workshop/internal/domain/milestone"
)

// TestMilestone_Validate tests validation of Milestone.
func TestMilestone_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ms      milestone.Milestone
		wantErr bool
	}{
		{
			name:    "valid classes milestone",
			ms:      milestone.Milestone{ID: "1", Name: "Century Club", Metric: milestone.MetricClasses, Threshold: 100},
			wantErr: false,
		},
		{
			name:    "valid mat hours milestone",
			ms:      milestone.Milestone{ID: "2", Name: "200 Mat Hours", Metric: milestone.MetricMatHours, Threshold: 200},
			wantErr: false,
		},
		{
			name:    "valid streak milestone",
			ms:      milestone.Milestone{ID: "3", Name: "Iron Streak", Metric: milestone.MetricStreakWeeks, Threshold: 52, BadgeIcon: "🔥"},
			wantErr: false,
		},
		{
			name:    "empty name",
			ms:      milestone.Milestone{ID: "4", Metric: milestone.MetricClasses, Threshold: 10},
			wantErr: true,
		},
		{
			name:    "invalid metric",
			ms:      milestone.Milestone{ID: "5", Name: "Bad", Metric: "invalid", Threshold: 10},
			wantErr: true,
		},
		{
			name:    "zero threshold",
			ms:      milestone.Milestone{ID: "6", Name: "Zero", Metric: milestone.MetricClasses, Threshold: 0},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ms.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
