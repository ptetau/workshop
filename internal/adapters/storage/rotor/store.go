package rotor

import (
	"context"

	domain "workshop/internal/domain/rotor"
)

// Store persists Rotor, RotorTheme, Topic, TopicSchedule, and Vote state.
type Store interface {
	// Rotor CRUD
	SaveRotor(ctx context.Context, r domain.Rotor) error
	GetRotor(ctx context.Context, id string) (domain.Rotor, error)
	ListRotorsByClassType(ctx context.Context, classTypeID string) ([]domain.Rotor, error)
	GetActiveRotor(ctx context.Context, classTypeID string) (domain.Rotor, error)
	DeleteRotor(ctx context.Context, id string) error

	// RotorTheme CRUD
	SaveRotorTheme(ctx context.Context, t domain.RotorTheme) error
	ListThemesByRotor(ctx context.Context, rotorID string) ([]domain.RotorTheme, error)
	DeleteRotorTheme(ctx context.Context, id string) error

	// Topic CRUD
	SaveTopic(ctx context.Context, t domain.Topic) error
	GetTopic(ctx context.Context, id string) (domain.Topic, error)
	ListTopicsByTheme(ctx context.Context, rotorThemeID string) ([]domain.Topic, error)
	DeleteTopic(ctx context.Context, id string) error
	ReorderTopics(ctx context.Context, rotorThemeID string, topicIDs []string) error

	// TopicSchedule
	SaveTopicSchedule(ctx context.Context, s domain.TopicSchedule) error
	GetActiveScheduleForTheme(ctx context.Context, rotorThemeID string) (domain.TopicSchedule, error)
	ListSchedulesByTheme(ctx context.Context, rotorThemeID string) ([]domain.TopicSchedule, error)

	// Votes
	SaveVote(ctx context.Context, v domain.Vote) error
	CountVotesForTopic(ctx context.Context, topicID string) (int, error)
	HasVoted(ctx context.Context, topicID, accountID string) (bool, error)
	DeleteVotesForTopic(ctx context.Context, topicID string) error
}
