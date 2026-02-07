package traininggoal_test

import (
	"testing"
	"time"

	"workshop/internal/domain/traininggoal"
)

// TestTrainingGoal_Validate tests validation of TrainingGoal.
func TestTrainingGoal_Validate(t *testing.T) {
	tests := []struct {
		name    string
		goal    traininggoal.TrainingGoal
		wantErr bool
	}{
		{
			name:    "valid weekly goal",
			goal:    traininggoal.TrainingGoal{ID: "1", MemberID: "m1", Target: 3, Period: traininggoal.PeriodWeekly, CreatedAt: time.Now(), Active: true},
			wantErr: false,
		},
		{
			name:    "valid monthly goal",
			goal:    traininggoal.TrainingGoal{ID: "2", MemberID: "m1", Target: 12, Period: traininggoal.PeriodMonthly, CreatedAt: time.Now(), Active: true},
			wantErr: false,
		},
		{
			name:    "empty member ID",
			goal:    traininggoal.TrainingGoal{ID: "3", Target: 3, Period: traininggoal.PeriodWeekly},
			wantErr: true,
		},
		{
			name:    "zero target",
			goal:    traininggoal.TrainingGoal{ID: "4", MemberID: "m1", Target: 0, Period: traininggoal.PeriodWeekly},
			wantErr: true,
		},
		{
			name:    "invalid period",
			goal:    traininggoal.TrainingGoal{ID: "5", MemberID: "m1", Target: 3, Period: "yearly"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.goal.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
