package orchestrators

import (
	"context"
	"log/slog"

	"workshop/internal/domain/classtype"
	"workshop/internal/domain/program"

	"github.com/google/uuid"
)

// ProgramStoreForSeed defines the store interface needed by SeedPrograms.
type ProgramStoreForSeed interface {
	Save(ctx context.Context, p program.Program) error
	List(ctx context.Context) ([]program.Program, error)
}

// ClassTypeStoreForSeed defines the store interface needed by SeedPrograms.
type ClassTypeStoreForSeed interface {
	Save(ctx context.Context, ct classtype.ClassType) error
	List(ctx context.Context) ([]classtype.ClassType, error)
}

// SeedProgramsDeps holds dependencies for SeedPrograms.
type SeedProgramsDeps struct {
	ProgramStore   ProgramStoreForSeed
	ClassTypeStore ClassTypeStoreForSeed
}

// ExecuteSeedPrograms creates default Programs and ClassTypes if none exist.
func ExecuteSeedPrograms(ctx context.Context, deps SeedProgramsDeps) error {
	existing, err := deps.ProgramStore.List(ctx)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil // Already seeded
	}

	// Create programs
	adultsID := uuid.New().String()
	kidsID := uuid.New().String()

	programs := []program.Program{
		{ID: adultsID, Name: "Adults", Type: program.TypeAdults},
		{ID: kidsID, Name: "Kids", Type: program.TypeKids},
	}

	for _, p := range programs {
		if err := deps.ProgramStore.Save(ctx, p); err != nil {
			return err
		}
	}

	// Create class types
	classTypes := []classtype.ClassType{
		{ID: uuid.New().String(), ProgramID: adultsID, Name: "Fundamentals"},
		{ID: uuid.New().String(), ProgramID: adultsID, Name: "No-Gi"},
		{ID: uuid.New().String(), ProgramID: adultsID, Name: "Competition"},
		{ID: uuid.New().String(), ProgramID: adultsID, Name: "Open Mat"},
		{ID: uuid.New().String(), ProgramID: kidsID, Name: "Kids Fundamentals"},
		{ID: uuid.New().String(), ProgramID: kidsID, Name: "Kids Advanced"},
	}

	for _, ct := range classTypes {
		if err := deps.ClassTypeStore.Save(ctx, ct); err != nil {
			return err
		}
	}

	slog.Info("seed_event", "event", "programs_seeded", "programs", len(programs), "class_types", len(classTypes))
	return nil
}
