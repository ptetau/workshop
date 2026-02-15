package featureflag

import "errors"

// FeatureFlag holds server-enforced availability controls for a feature.
//
// Key is stable and referenced by code (routes/templates).
//
// NOTE: We store booleans per role explicitly rather than using maps to keep
// storage and JSON payloads simple.
type FeatureFlag struct {
	Key         string
	Description string

	EnabledAdmin  bool
	EnabledCoach  bool
	EnabledMember bool
	EnabledTrial  bool

	// BetaOverride enables the feature for beta testers even if it is disabled
	// for their role.
	BetaOverride bool
}

var (
	ErrMissingKey = errors.New("feature flag key is required")
)

// Validate checks required fields for a FeatureFlag.
// PRE: FeatureFlag struct is initialized
// POST: Returns error if validation fails, nil otherwise
func (f *FeatureFlag) Validate() error {
	if f.Key == "" {
		return ErrMissingKey
	}
	return nil
}

// EnabledForRole returns true if the feature is enabled for the given role,
// considering the beta override.
//
// PRE: role is a valid session role string
// INVARIANT: f is not mutated
func (f FeatureFlag) EnabledForRole(role string, isBetaTester bool) bool {
	if isBetaTester && f.BetaOverride {
		return true
	}
	switch role {
	case "admin":
		return f.EnabledAdmin
	case "coach":
		return f.EnabledCoach
	case "member":
		return f.EnabledMember
	case "trial":
		return f.EnabledTrial
	default:
		return false
	}
}
