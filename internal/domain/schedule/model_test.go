package schedule_test

import (
	"testing"

	"workshop/internal/domain/schedule"
)

// TestSchedule_Validate tests validation of Schedule.
func TestSchedule_Validate(t *testing.T) {
	tests := []struct {
		name    string
		sched   schedule.Schedule
		wantErr bool
	}{
		{
			name:    "valid schedule",
			sched:   schedule.Schedule{ID: "1", ClassTypeID: "ct-1", Day: schedule.Monday, StartTime: "18:00", EndTime: "19:30"},
			wantErr: false,
		},
		{
			name:    "valid saturday",
			sched:   schedule.Schedule{ID: "2", ClassTypeID: "ct-1", Day: schedule.Saturday, StartTime: "10:00", EndTime: "11:30"},
			wantErr: false,
		},
		{
			name:    "empty class type ID",
			sched:   schedule.Schedule{ID: "3", ClassTypeID: "", Day: schedule.Monday, StartTime: "18:00", EndTime: "19:30"},
			wantErr: true,
		},
		{
			name:    "invalid day",
			sched:   schedule.Schedule{ID: "4", ClassTypeID: "ct-1", Day: "funday", StartTime: "18:00", EndTime: "19:30"},
			wantErr: true,
		},
		{
			name:    "empty day",
			sched:   schedule.Schedule{ID: "5", ClassTypeID: "ct-1", Day: "", StartTime: "18:00", EndTime: "19:30"},
			wantErr: true,
		},
		{
			name:    "empty start time",
			sched:   schedule.Schedule{ID: "6", ClassTypeID: "ct-1", Day: schedule.Monday, StartTime: "", EndTime: "19:30"},
			wantErr: true,
		},
		{
			name:    "empty end time",
			sched:   schedule.Schedule{ID: "7", ClassTypeID: "ct-1", Day: schedule.Monday, StartTime: "18:00", EndTime: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sched.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Schedule.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
