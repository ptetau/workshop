package featureflag

// DefaultFlags returns the known feature flags and their default settings.
//
// These are intended to represent broad, user-visible areas of the product.
// As new major features are added, append to this list.
func DefaultFlags() []FeatureFlag {
	return []FeatureFlag{
		{
			Key:           "member_mgmt",
			Description:   "Member management (member list, profiles)",
			EnabledAdmin:  true,
			EnabledCoach:  true,
			EnabledMember: false,
			EnabledTrial:  false,
		},
		{
			Key:           "attendance",
			Description:   "Attendance (check-ins, attendance views)",
			EnabledAdmin:  true,
			EnabledCoach:  true,
			EnabledMember: false,
			EnabledTrial:  false,
		},
		{
			Key:           "calendar",
			Description:   "Calendar (club events, competitions)",
			EnabledAdmin:  true,
			EnabledCoach:  true,
			EnabledMember: true,
			EnabledTrial:  true,
		},
		{
			Key:           "curriculum",
			Description:   "Curriculum (rotors, themes/topics)",
			EnabledAdmin:  true,
			EnabledCoach:  true,
			EnabledMember: true,
			EnabledTrial:  false,
		},
		{
			Key:           "library",
			Description:   "Technical library (themes + clips)",
			EnabledAdmin:  true,
			EnabledCoach:  true,
			EnabledMember: true,
			EnabledTrial:  false,
		},
		{
			Key:           "messages",
			Description:   "Messages (member messaging)",
			EnabledAdmin:  true,
			EnabledCoach:  true,
			EnabledMember: true,
			EnabledTrial:  true,
		},
		{
			Key:           "training_log",
			Description:   "Training log",
			EnabledAdmin:  true,
			EnabledCoach:  true,
			EnabledMember: true,
			EnabledTrial:  true,
		},
		{
			Key:           "grading",
			Description:   "Grading (readiness, proposals, admin grading)",
			EnabledAdmin:  true,
			EnabledCoach:  true,
			EnabledMember: false,
			EnabledTrial:  false,
		},
		{
			Key:           "emails",
			Description:   "Email system",
			EnabledAdmin:  true,
			EnabledCoach:  false,
			EnabledMember: false,
			EnabledTrial:  false,
		},
		{
			Key:           "kiosk",
			Description:   "Kiosk",
			EnabledAdmin:  true,
			EnabledCoach:  true,
			EnabledMember: false,
			EnabledTrial:  false,
		},
	}
}
