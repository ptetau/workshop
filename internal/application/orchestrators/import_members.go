package orchestrators

import (
	"context"
	"encoding/csv"
	"io"
	"log/slog"
	"net/mail"
	"strconv"
	"strings"

	memberStore "workshop/internal/adapters/storage/member"
	domain "workshop/internal/domain/member"
)

// ImportMembersInput carries the parsed CSV reader and import options.
// PRE: Reader is a valid CSV stream with a header row; AdminAccountID is non-empty.
// POST: Returns aggregate counts and per-row errors; writes are skipped when DryRun=true.
// INVARIANT: Existing members are never deleted; IDs are preserved on update.
type ImportMembersInput struct {
	Reader         io.Reader
	AdminAccountID string
	DryRun         bool
	UpdateMode     bool
}

// ImportMembersResult holds aggregate counts and per-row errors from an import run.
type ImportMembersResult struct {
	Total   int
	Created int
	Updated int
	Skipped int
	Errors  []ImportMembersRowError
	DryRun  bool
	Unknown []string
}

// ImportMembersRowError describes a validation or processing error for a single CSV row.
type ImportMembersRowError struct {
	Row     int
	Message string
}

// ImportMembersDeps holds external dependencies for the import orchestrator.
type ImportMembersDeps struct {
	MemberStore memberStore.Store
	GenerateID  func() string
}

// ExecuteImportMembers parses a CSV stream and creates or updates member records.
// PRE: Input.Reader contains a valid CSV with at least NAME and EMAIL columns.
// POST: Members are created/updated/skipped according to DryRun and UpdateMode flags;
//
//	aggregate counts and per-row errors are returned; audit log is emitted.
//
// INVARIANT: When DryRun=true no writes occur; existing member IDs are always preserved on update.
func ExecuteImportMembers(ctx context.Context, input ImportMembersInput, deps ImportMembersDeps) (ImportMembersResult, error) {
	cr := csv.NewReader(input.Reader)
	cr.TrimLeadingSpace = true

	header, err := cr.Read()
	if err != nil {
		return ImportMembersResult{}, err
	}

	colIdx := make(map[string]int, len(header))
	for i, h := range header {
		colIdx[strings.ToUpper(strings.TrimSpace(h))] = i
	}

	if _, ok := colIdx["NAME"]; !ok {
		return ImportMembersResult{}, &ImportMembersValidationError{Message: "CSV missing required column: NAME"}
	}
	if _, ok := colIdx["EMAIL"]; !ok {
		return ImportMembersResult{}, &ImportMembersValidationError{Message: "CSV missing required column: EMAIL"}
	}

	known := map[string]bool{
		"ID": true, "ACCOUNTID": true, "NAME": true, "EMAIL": true,
		"PROGRAM": true, "STATUS": true, "FEE": true, "FREQUENCY": true, "GRADINGMETRIC": true,
	}
	var unknownCols []string
	for _, h := range header {
		if !known[strings.ToUpper(strings.TrimSpace(h))] {
			unknownCols = append(unknownCols, h)
		}
	}

	getCol := func(row []string, col string) string {
		i, ok := colIdx[col]
		if !ok || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}

	result := ImportMembersResult{DryRun: input.DryRun, Unknown: unknownCols}
	rowNum := 1

	for {
		row, err := cr.Read()
		if err != nil {
			break
		}
		rowNum++
		result.Total++

		name := getCol(row, "NAME")
		rawEmail := getCol(row, "EMAIL")

		if strings.TrimSpace(name) == "" {
			result.Errors = append(result.Errors, ImportMembersRowError{Row: rowNum, Message: "name is required"})
			continue
		}

		addr, parseErr := mail.ParseAddress(rawEmail)
		if parseErr != nil {
			result.Errors = append(result.Errors, ImportMembersRowError{Row: rowNum, Message: "invalid email: " + rawEmail})
			continue
		}
		email := strings.ToLower(addr.Address)

		program := strings.ToLower(getCol(row, "PROGRAM"))
		if program != domain.ProgramAdults && program != domain.ProgramKids {
			program = domain.ProgramAdults
		}
		status := strings.ToLower(getCol(row, "STATUS"))
		if status != domain.StatusActive && status != domain.StatusInactive && status != domain.StatusArchived {
			status = domain.StatusActive
		}
		fee, _ := strconv.Atoi(getCol(row, "FEE"))
		frequency := getCol(row, "FREQUENCY")
		gradingMetric := getCol(row, "GRADINGMETRIC")
		if gradingMetric != domain.MetricSessions && gradingMetric != domain.MetricHours {
			gradingMetric = domain.MetricSessions
		}

		existing, lookupErr := deps.MemberStore.GetByEmail(ctx, email)
		exists := lookupErr == nil

		if exists && !input.UpdateMode {
			result.Skipped++
			continue
		}

		if input.DryRun {
			if exists {
				result.Updated++
			} else {
				result.Created++
			}
			continue
		}

		if exists {
			existing.Name = name
			existing.Program = program
			existing.Status = status
			if fee > 0 {
				existing.Fee = fee
			}
			if frequency != "" {
				existing.Frequency = frequency
			}
			if gradingMetric != "" {
				existing.GradingMetric = gradingMetric
			}
			if err := deps.MemberStore.Save(ctx, existing); err != nil {
				slog.Error("members_import_save_failed", "row", rowNum, "email", email, "err", err)
				result.Errors = append(result.Errors, ImportMembersRowError{Row: rowNum, Message: "save failed (see server log)"})
				continue
			}
			result.Updated++
		} else {
			m := domain.Member{
				ID:            deps.GenerateID(),
				Name:          name,
				Email:         email,
				Program:       program,
				Status:        status,
				Fee:           fee,
				Frequency:     frequency,
				GradingMetric: gradingMetric,
			}
			if err := deps.MemberStore.Save(ctx, m); err != nil {
				slog.Error("members_import_save_failed", "row", rowNum, "email", email, "err", err)
				result.Errors = append(result.Errors, ImportMembersRowError{Row: rowNum, Message: "save failed (see server log)"})
				continue
			}
			result.Created++
		}
	}

	slog.Info("members_import",
		"admin", input.AdminAccountID,
		"dry_run", input.DryRun,
		"update_mode", input.UpdateMode,
		"total", result.Total,
		"created", result.Created,
		"updated", result.Updated,
		"skipped", result.Skipped,
		"errors", len(result.Errors),
	)

	return result, nil
}

// ImportMembersValidationError is returned when the CSV structure is invalid (e.g. missing required columns).
type ImportMembersValidationError struct {
	Message string
}

// Error implements the error interface.
// PRE: e.Message is set.
// POST: returns the validation error message string.
// INVARIANT: message is never empty for a valid ImportMembersValidationError.
func (e *ImportMembersValidationError) Error() string {
	return e.Message
}
