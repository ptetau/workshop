package featureflag

import (
	"context"
	"database/sql"
	"fmt"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/featureflag"
)

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct {
	db storage.SQLDB
}

// NewSQLiteStore creates a new FeatureFlag store.
func NewSQLiteStore(db storage.SQLDB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// GetByKey retrieves a single FeatureFlag by its stable key.
// PRE: key is non-empty
// POST: Returns the persisted feature flag or an error if not found
// INVARIANT: Store state is not mutated
func (s *SQLiteStore) GetByKey(ctx context.Context, key string) (domain.FeatureFlag, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT key, description, enabled_admin, enabled_coach, enabled_member, enabled_trial, beta_override
		FROM feature_flag
		WHERE key = ?
	`, key)
	return scanFlag(row.Scan)
}

// List returns all persisted feature flags.
// PRE: none
// POST: Returns all persisted flags sorted by key
// INVARIANT: Store state is not mutated
func (s *SQLiteStore) List(ctx context.Context) ([]domain.FeatureFlag, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT key, description, enabled_admin, enabled_coach, enabled_member, enabled_trial, beta_override
		FROM feature_flag
		ORDER BY key
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.FeatureFlag
	for rows.Next() {
		ff, err := scanFlag(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, ff)
	}
	if out == nil {
		out = []domain.FeatureFlag{}
	}
	return out, nil
}

// Save upserts a feature flag.
// PRE: value has a non-empty Key
// POST: Feature flag is persisted (insert or update)
// INVARIANT: No other feature flags are modified
func (s *SQLiteStore) Save(ctx context.Context, value domain.FeatureFlag) error {
	if err := value.Validate(); err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO feature_flag (
			key, description,
			enabled_admin, enabled_coach, enabled_member, enabled_trial,
			beta_override
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			description=excluded.description,
			enabled_admin=excluded.enabled_admin,
			enabled_coach=excluded.enabled_coach,
			enabled_member=excluded.enabled_member,
			enabled_trial=excluded.enabled_trial,
			beta_override=excluded.beta_override
	`,
		value.Key,
		value.Description,
		boolToInt(value.EnabledAdmin),
		boolToInt(value.EnabledCoach),
		boolToInt(value.EnabledMember),
		boolToInt(value.EnabledTrial),
		boolToInt(value.BetaOverride),
	)
	if err != nil {
		return fmt.Errorf("save feature_flag: %w", err)
	}

	return tx.Commit()
}

func scanFlag(scan func(dest ...any) error) (domain.FeatureFlag, error) {
	var ff domain.FeatureFlag
	var enabledAdmin, enabledCoach, enabledMember, enabledTrial, betaOverride int
	if err := scan(
		&ff.Key,
		&ff.Description,
		&enabledAdmin,
		&enabledCoach,
		&enabledMember,
		&enabledTrial,
		&betaOverride,
	); err != nil {
		if err == sql.ErrNoRows {
			return domain.FeatureFlag{}, fmt.Errorf("feature flag not found: %w", err)
		}
		return domain.FeatureFlag{}, err
	}
	ff.EnabledAdmin = enabledAdmin != 0
	ff.EnabledCoach = enabledCoach != 0
	ff.EnabledMember = enabledMember != 0
	ff.EnabledTrial = enabledTrial != 0
	ff.BetaOverride = betaOverride != 0
	return ff, nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
