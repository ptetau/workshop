package featureflag

import "testing"

// TestFeatureFlag_EnabledForRole_RoleToggle verifies role toggles are applied.
func TestFeatureFlag_EnabledForRole_RoleToggle(t *testing.T) {
	ff := FeatureFlag{
		Key:           "library",
		Description:   "Library",
		EnabledAdmin:  true,
		EnabledCoach:  false,
		EnabledMember: false,
		EnabledTrial:  false,
		BetaOverride:  false,
	}

	if !ff.EnabledForRole("admin", false) {
		t.Fatalf("expected admin enabled")
	}
	if ff.EnabledForRole("coach", false) {
		t.Fatalf("expected coach disabled")
	}
	if ff.EnabledForRole("member", false) {
		t.Fatalf("expected member disabled")
	}
	if ff.EnabledForRole("trial", false) {
		t.Fatalf("expected trial disabled")
	}
}

// TestFeatureFlag_EnabledForRole_BetaOverride verifies beta override enables access.
func TestFeatureFlag_EnabledForRole_BetaOverride(t *testing.T) {
	ff := FeatureFlag{
		Key:           "library",
		Description:   "Library",
		EnabledAdmin:  true,
		EnabledCoach:  true,
		EnabledMember: false,
		EnabledTrial:  false,
		BetaOverride:  true,
	}

	if !ff.EnabledForRole("member", true) {
		t.Fatalf("expected beta member enabled via override")
	}
	if !ff.EnabledForRole("trial", true) {
		t.Fatalf("expected beta trial enabled via override")
	}
	if ff.EnabledForRole("member", false) {
		t.Fatalf("expected non-beta member disabled")
	}
}
