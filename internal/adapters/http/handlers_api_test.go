package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"workshop/internal/adapters/http/middleware"
	accountStore "workshop/internal/adapters/storage/account"
	attendanceStore "workshop/internal/adapters/storage/attendance"
	classTypeStore "workshop/internal/adapters/storage/classtype"
	gradingStore "workshop/internal/adapters/storage/grading"
	holidayStore "workshop/internal/adapters/storage/holiday"
	injuryStore "workshop/internal/adapters/storage/injury"
	memberStore "workshop/internal/adapters/storage/member"
	messageStore "workshop/internal/adapters/storage/message"
	milestoneStore "workshop/internal/adapters/storage/milestone"
	noticeStore "workshop/internal/adapters/storage/notice"
	observationStore "workshop/internal/adapters/storage/observation"
	programStore "workshop/internal/adapters/storage/program"
	scheduleStore "workshop/internal/adapters/storage/schedule"
	termStore "workshop/internal/adapters/storage/term"
	trainingGoalStore "workshop/internal/adapters/storage/traininggoal"
	waiverStore "workshop/internal/adapters/storage/waiver"

	accountDomain "workshop/internal/domain/account"
	attendanceDomain "workshop/internal/domain/attendance"
	classTypeDomain "workshop/internal/domain/classtype"
	gradingDomain "workshop/internal/domain/grading"
	holidayDomain "workshop/internal/domain/holiday"
	injuryDomain "workshop/internal/domain/injury"
	memberDomain "workshop/internal/domain/member"
	messageDomain "workshop/internal/domain/message"
	milestoneDomain "workshop/internal/domain/milestone"
	noticeDomain "workshop/internal/domain/notice"
	observationDomain "workshop/internal/domain/observation"
	programDomain "workshop/internal/domain/program"
	scheduleDomain "workshop/internal/domain/schedule"
	termDomain "workshop/internal/domain/term"
	trainingGoalDomain "workshop/internal/domain/traininggoal"
	waiverDomain "workshop/internal/domain/waiver"
)

// --- Mock stores for Layer 1b+ ---

type mockAccountStore struct {
	accounts map[string]accountDomain.Account
}

// GetByID implements the mock AccountStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockAccountStore) GetByID(ctx context.Context, id string) (accountDomain.Account, error) {
	if a, ok := m.accounts[id]; ok {
		return a, nil
	}
	return accountDomain.Account{}, sql.ErrNoRows
}

// GetByEmail implements the mock AccountStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockAccountStore) GetByEmail(ctx context.Context, email string) (accountDomain.Account, error) {
	for _, a := range m.accounts {
		if a.Email == email {
			return a, nil
		}
	}
	return accountDomain.Account{}, sql.ErrNoRows
}

// Save implements the mock AccountStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockAccountStore) Save(ctx context.Context, a accountDomain.Account) error {
	if m.accounts == nil {
		m.accounts = make(map[string]accountDomain.Account)
	}
	m.accounts[a.ID] = a
	return nil
}

// Delete implements the mock AccountStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockAccountStore) Delete(ctx context.Context, id string) error {
	delete(m.accounts, id)
	return nil
}

// List implements the mock AccountStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockAccountStore) List(ctx context.Context, filter accountStore.ListFilter) ([]accountDomain.Account, error) {
	var list []accountDomain.Account
	for _, a := range m.accounts {
		list = append(list, a)
	}
	return list, nil
}

// Count implements the mock AccountStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockAccountStore) Count(ctx context.Context) (int, error) {
	return len(m.accounts), nil
}

// SaveActivationToken implements the mock AccountStore for testing.
// PRE: valid parameters
// POST: returns nil
func (m *mockAccountStore) SaveActivationToken(ctx context.Context, token accountDomain.ActivationToken) error {
	return nil
}

// GetActivationTokenByToken implements the mock AccountStore for testing.
// PRE: valid parameters
// POST: returns error (stub)
func (m *mockAccountStore) GetActivationTokenByToken(ctx context.Context, token string) (accountDomain.ActivationToken, error) {
	return accountDomain.ActivationToken{}, errors.New("not found")
}

// InvalidateTokensForAccount implements the mock AccountStore for testing.
// PRE: valid parameters
// POST: returns nil
func (m *mockAccountStore) InvalidateTokensForAccount(ctx context.Context, accountID string) error {
	return nil
}

type mockNoticeStore struct {
	notices map[string]noticeDomain.Notice
}

// GetByID implements the mock NoticeStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockNoticeStore) GetByID(ctx context.Context, id string) (noticeDomain.Notice, error) {
	if n, ok := m.notices[id]; ok {
		return n, nil
	}
	return noticeDomain.Notice{}, sql.ErrNoRows
}

// Save implements the mock NoticeStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockNoticeStore) Save(ctx context.Context, n noticeDomain.Notice) error {
	if m.notices == nil {
		m.notices = make(map[string]noticeDomain.Notice)
	}
	m.notices[n.ID] = n
	return nil
}

// Delete implements the mock NoticeStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockNoticeStore) Delete(ctx context.Context, id string) error {
	delete(m.notices, id)
	return nil
}

// List implements the mock NoticeStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockNoticeStore) List(ctx context.Context, filter noticeStore.ListFilter) ([]noticeDomain.Notice, error) {
	var list []noticeDomain.Notice
	for _, n := range m.notices {
		list = append(list, n)
	}
	return list, nil
}

// ListPublished implements the mock NoticeStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockNoticeStore) ListPublished(ctx context.Context, noticeType string, now time.Time) ([]noticeDomain.Notice, error) {
	var list []noticeDomain.Notice
	for _, n := range m.notices {
		if n.Status == noticeDomain.StatusPublished && (noticeType == "" || n.Type == noticeType) {
			list = append(list, n)
		}
	}
	return list, nil
}

type mockMessageStore struct {
	messages map[string]messageDomain.Message
}

// GetByID implements the mock MessageStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockMessageStore) GetByID(ctx context.Context, id string) (messageDomain.Message, error) {
	if msg, ok := m.messages[id]; ok {
		return msg, nil
	}
	return messageDomain.Message{}, sql.ErrNoRows
}

// Save implements the mock MessageStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockMessageStore) Save(ctx context.Context, msg messageDomain.Message) error {
	if m.messages == nil {
		m.messages = make(map[string]messageDomain.Message)
	}
	m.messages[msg.ID] = msg
	return nil
}

// Delete implements the mock MessageStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockMessageStore) Delete(ctx context.Context, id string) error {
	delete(m.messages, id)
	return nil
}

// ListByReceiverID implements the mock MessageStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockMessageStore) ListByReceiverID(ctx context.Context, receiverID string) ([]messageDomain.Message, error) {
	var list []messageDomain.Message
	for _, msg := range m.messages {
		if msg.ReceiverID == receiverID {
			list = append(list, msg)
		}
	}
	return list, nil
}

// CountUnread implements the mock MessageStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockMessageStore) CountUnread(ctx context.Context, receiverID string) (int, error) {
	count := 0
	for _, msg := range m.messages {
		if msg.ReceiverID == receiverID && msg.ReadAt.IsZero() {
			count++
		}
	}
	return count, nil
}

type mockObservationStore struct {
	observations map[string]observationDomain.Observation
}

// GetByID implements the mock ObservationStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockObservationStore) GetByID(ctx context.Context, id string) (observationDomain.Observation, error) {
	if o, ok := m.observations[id]; ok {
		return o, nil
	}
	return observationDomain.Observation{}, sql.ErrNoRows
}

// Save implements the mock ObservationStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockObservationStore) Save(ctx context.Context, o observationDomain.Observation) error {
	if m.observations == nil {
		m.observations = make(map[string]observationDomain.Observation)
	}
	m.observations[o.ID] = o
	return nil
}

// Delete implements the mock ObservationStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockObservationStore) Delete(ctx context.Context, id string) error {
	delete(m.observations, id)
	return nil
}

// ListByMemberID implements the mock ObservationStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockObservationStore) ListByMemberID(ctx context.Context, memberID string) ([]observationDomain.Observation, error) {
	var list []observationDomain.Observation
	for _, o := range m.observations {
		if o.MemberID == memberID {
			list = append(list, o)
		}
	}
	return list, nil
}

type mockGradingRecordStore struct {
	records map[string]gradingDomain.Record
}

// GetByID implements the mock GradingRecordStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockGradingRecordStore) GetByID(ctx context.Context, id string) (gradingDomain.Record, error) {
	if r, ok := m.records[id]; ok {
		return r, nil
	}
	return gradingDomain.Record{}, sql.ErrNoRows
}

// Save implements the mock GradingRecordStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockGradingRecordStore) Save(ctx context.Context, r gradingDomain.Record) error {
	if m.records == nil {
		m.records = make(map[string]gradingDomain.Record)
	}
	m.records[r.ID] = r
	return nil
}

// ListByMemberID implements the mock GradingRecordStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockGradingRecordStore) ListByMemberID(ctx context.Context, memberID string) ([]gradingDomain.Record, error) {
	var list []gradingDomain.Record
	for _, r := range m.records {
		if r.MemberID == memberID {
			list = append(list, r)
		}
	}
	return list, nil
}

type mockGradingConfigStore struct {
	configs map[string]gradingDomain.Config
}

// GetByID implements the mock GradingConfigStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockGradingConfigStore) GetByID(ctx context.Context, id string) (gradingDomain.Config, error) {
	if c, ok := m.configs[id]; ok {
		return c, nil
	}
	return gradingDomain.Config{}, sql.ErrNoRows
}

// Save implements the mock GradingConfigStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockGradingConfigStore) Save(ctx context.Context, c gradingDomain.Config) error {
	if m.configs == nil {
		m.configs = make(map[string]gradingDomain.Config)
	}
	m.configs[c.ID] = c
	return nil
}

// GetByProgramAndBelt implements the mock GradingConfigStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockGradingConfigStore) GetByProgramAndBelt(ctx context.Context, program, belt string) (gradingDomain.Config, error) {
	for _, c := range m.configs {
		if c.Program == program && c.Belt == belt {
			return c, nil
		}
	}
	return gradingDomain.Config{}, sql.ErrNoRows
}

// List implements the mock GradingConfigStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockGradingConfigStore) List(ctx context.Context) ([]gradingDomain.Config, error) {
	var list []gradingDomain.Config
	for _, c := range m.configs {
		list = append(list, c)
	}
	return list, nil
}

type mockGradingMemberConfigStore struct {
	configs map[string]gradingDomain.MemberConfig
}

// Save implements the mock GradingMemberConfigStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockGradingMemberConfigStore) Save(ctx context.Context, mc gradingDomain.MemberConfig) error {
	if m.configs == nil {
		m.configs = make(map[string]gradingDomain.MemberConfig)
	}
	m.configs[mc.MemberID+"|"+mc.Belt] = mc
	return nil
}

// GetByMemberAndBelt implements the mock GradingMemberConfigStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockGradingMemberConfigStore) GetByMemberAndBelt(ctx context.Context, memberID, belt string) (gradingDomain.MemberConfig, error) {
	if mc, ok := m.configs[memberID+"|"+belt]; ok {
		return mc, nil
	}
	return gradingDomain.MemberConfig{}, sql.ErrNoRows
}

// ListByMemberID implements the mock GradingMemberConfigStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockGradingMemberConfigStore) ListByMemberID(ctx context.Context, memberID string) ([]gradingDomain.MemberConfig, error) {
	var list []gradingDomain.MemberConfig
	for _, mc := range m.configs {
		if mc.MemberID == memberID {
			list = append(list, mc)
		}
	}
	return list, nil
}

type mockGradingProposalStore struct {
	proposals map[string]gradingDomain.Proposal
}

// GetByID implements the mock GradingProposalStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockGradingProposalStore) GetByID(ctx context.Context, id string) (gradingDomain.Proposal, error) {
	if p, ok := m.proposals[id]; ok {
		return p, nil
	}
	return gradingDomain.Proposal{}, sql.ErrNoRows
}

// Save implements the mock GradingProposalStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockGradingProposalStore) Save(ctx context.Context, p gradingDomain.Proposal) error {
	if m.proposals == nil {
		m.proposals = make(map[string]gradingDomain.Proposal)
	}
	m.proposals[p.ID] = p
	return nil
}

// ListPending implements the mock GradingProposalStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockGradingProposalStore) ListPending(ctx context.Context) ([]gradingDomain.Proposal, error) {
	var list []gradingDomain.Proposal
	for _, p := range m.proposals {
		if p.Status == gradingDomain.ProposalPending {
			list = append(list, p)
		}
	}
	return list, nil
}

// ListByMemberID implements the mock GradingProposalStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockGradingProposalStore) ListByMemberID(ctx context.Context, memberID string) ([]gradingDomain.Proposal, error) {
	var list []gradingDomain.Proposal
	for _, p := range m.proposals {
		if p.MemberID == memberID {
			list = append(list, p)
		}
	}
	return list, nil
}

type mockMilestoneStore struct {
	milestones map[string]milestoneDomain.Milestone
}

// GetByID implements the mock MilestoneStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockMilestoneStore) GetByID(ctx context.Context, id string) (milestoneDomain.Milestone, error) {
	if ms, ok := m.milestones[id]; ok {
		return ms, nil
	}
	return milestoneDomain.Milestone{}, sql.ErrNoRows
}

// Save implements the mock MilestoneStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockMilestoneStore) Save(ctx context.Context, ms milestoneDomain.Milestone) error {
	if m.milestones == nil {
		m.milestones = make(map[string]milestoneDomain.Milestone)
	}
	m.milestones[ms.ID] = ms
	return nil
}

// Delete implements the mock MilestoneStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockMilestoneStore) Delete(ctx context.Context, id string) error {
	delete(m.milestones, id)
	return nil
}

// List implements the mock MilestoneStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockMilestoneStore) List(ctx context.Context) ([]milestoneDomain.Milestone, error) {
	var list []milestoneDomain.Milestone
	for _, ms := range m.milestones {
		list = append(list, ms)
	}
	return list, nil
}

type mockMemberMilestoneStore struct {
	items map[string]milestoneDomain.MemberMilestone
}

// Save implements the mock MemberMilestoneStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockMemberMilestoneStore) Save(ctx context.Context, value milestoneDomain.MemberMilestone) error {
	if m.items == nil {
		m.items = make(map[string]milestoneDomain.MemberMilestone)
	}
	m.items[value.ID] = value
	return nil
}

// ListByMemberID implements the mock MemberMilestoneStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockMemberMilestoneStore) ListByMemberID(ctx context.Context, memberID string) ([]milestoneDomain.MemberMilestone, error) {
	var list []milestoneDomain.MemberMilestone
	for _, item := range m.items {
		if item.MemberID == memberID {
			list = append(list, item)
		}
	}
	return list, nil
}

// MarkNotified implements the mock MemberMilestoneStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockMemberMilestoneStore) MarkNotified(ctx context.Context, id string) error {
	if item, ok := m.items[id]; ok {
		item.Notified = true
		m.items[id] = item
	}
	return nil
}

// ListUnnotifiedByMemberID implements the mock MemberMilestoneStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockMemberMilestoneStore) ListUnnotifiedByMemberID(ctx context.Context, memberID string) ([]milestoneDomain.MemberMilestone, error) {
	var list []milestoneDomain.MemberMilestone
	for _, item := range m.items {
		if item.MemberID == memberID && !item.Notified {
			list = append(list, item)
		}
	}
	return list, nil
}

type mockTrainingGoalStore struct {
	goals map[string]trainingGoalDomain.TrainingGoal
}

// GetByID implements the mock TrainingGoalStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockTrainingGoalStore) GetByID(ctx context.Context, id string) (trainingGoalDomain.TrainingGoal, error) {
	if g, ok := m.goals[id]; ok {
		return g, nil
	}
	return trainingGoalDomain.TrainingGoal{}, sql.ErrNoRows
}

// Save implements the mock TrainingGoalStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockTrainingGoalStore) Save(ctx context.Context, g trainingGoalDomain.TrainingGoal) error {
	if m.goals == nil {
		m.goals = make(map[string]trainingGoalDomain.TrainingGoal)
	}
	m.goals[g.ID] = g
	return nil
}

// Delete implements the mock TrainingGoalStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockTrainingGoalStore) Delete(ctx context.Context, id string) error {
	delete(m.goals, id)
	return nil
}

// GetActiveByMemberID implements the mock TrainingGoalStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockTrainingGoalStore) GetActiveByMemberID(ctx context.Context, memberID string) (trainingGoalDomain.TrainingGoal, error) {
	for _, g := range m.goals {
		if g.MemberID == memberID && g.Active {
			return g, nil
		}
	}
	return trainingGoalDomain.TrainingGoal{}, sql.ErrNoRows
}

// ListByMemberID implements the mock TrainingGoalStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockTrainingGoalStore) ListByMemberID(ctx context.Context, memberID string) ([]trainingGoalDomain.TrainingGoal, error) {
	var list []trainingGoalDomain.TrainingGoal
	for _, g := range m.goals {
		if g.MemberID == memberID {
			list = append(list, g)
		}
	}
	return list, nil
}

type mockScheduleStore struct {
	schedules map[string]scheduleDomain.Schedule
}

// GetByID implements the mock ScheduleStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockScheduleStore) GetByID(ctx context.Context, id string) (scheduleDomain.Schedule, error) {
	if s, ok := m.schedules[id]; ok {
		return s, nil
	}
	return scheduleDomain.Schedule{}, sql.ErrNoRows
}

// Save implements the mock ScheduleStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockScheduleStore) Save(ctx context.Context, s scheduleDomain.Schedule) error {
	if m.schedules == nil {
		m.schedules = make(map[string]scheduleDomain.Schedule)
	}
	m.schedules[s.ID] = s
	return nil
}

// Delete implements the mock ScheduleStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockScheduleStore) Delete(ctx context.Context, id string) error {
	delete(m.schedules, id)
	return nil
}

// List implements the mock ScheduleStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockScheduleStore) List(ctx context.Context) ([]scheduleDomain.Schedule, error) {
	var list []scheduleDomain.Schedule
	for _, s := range m.schedules {
		list = append(list, s)
	}
	return list, nil
}

// ListByDay implements the mock ScheduleStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockScheduleStore) ListByDay(ctx context.Context, day string) ([]scheduleDomain.Schedule, error) {
	var list []scheduleDomain.Schedule
	for _, s := range m.schedules {
		if s.Day == day {
			list = append(list, s)
		}
	}
	return list, nil
}

// ListByClassTypeID implements the mock ScheduleStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockScheduleStore) ListByClassTypeID(ctx context.Context, classTypeID string) ([]scheduleDomain.Schedule, error) {
	var list []scheduleDomain.Schedule
	for _, s := range m.schedules {
		if s.ClassTypeID == classTypeID {
			list = append(list, s)
		}
	}
	return list, nil
}

type mockTermStore struct {
	terms map[string]termDomain.Term
}

// GetByID implements the mock TermStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockTermStore) GetByID(ctx context.Context, id string) (termDomain.Term, error) {
	if t, ok := m.terms[id]; ok {
		return t, nil
	}
	return termDomain.Term{}, sql.ErrNoRows
}

// Save implements the mock TermStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockTermStore) Save(ctx context.Context, t termDomain.Term) error {
	if m.terms == nil {
		m.terms = make(map[string]termDomain.Term)
	}
	m.terms[t.ID] = t
	return nil
}

// Delete implements the mock TermStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockTermStore) Delete(ctx context.Context, id string) error {
	delete(m.terms, id)
	return nil
}

// List implements the mock TermStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockTermStore) List(ctx context.Context) ([]termDomain.Term, error) {
	var list []termDomain.Term
	for _, t := range m.terms {
		list = append(list, t)
	}
	return list, nil
}

type mockHolidayStore struct {
	holidays map[string]holidayDomain.Holiday
}

// GetByID implements the mock HolidayStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockHolidayStore) GetByID(ctx context.Context, id string) (holidayDomain.Holiday, error) {
	if h, ok := m.holidays[id]; ok {
		return h, nil
	}
	return holidayDomain.Holiday{}, sql.ErrNoRows
}

// Save implements the mock HolidayStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockHolidayStore) Save(ctx context.Context, h holidayDomain.Holiday) error {
	if m.holidays == nil {
		m.holidays = make(map[string]holidayDomain.Holiday)
	}
	m.holidays[h.ID] = h
	return nil
}

// Delete implements the mock HolidayStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockHolidayStore) Delete(ctx context.Context, id string) error {
	delete(m.holidays, id)
	return nil
}

// List implements the mock HolidayStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockHolidayStore) List(ctx context.Context) ([]holidayDomain.Holiday, error) {
	var list []holidayDomain.Holiday
	for _, h := range m.holidays {
		list = append(list, h)
	}
	return list, nil
}

type mockClassTypeStore struct {
	classTypes map[string]classTypeDomain.ClassType
}

// GetByID implements the mock ClassTypeStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockClassTypeStore) GetByID(ctx context.Context, id string) (classTypeDomain.ClassType, error) {
	if ct, ok := m.classTypes[id]; ok {
		return ct, nil
	}
	return classTypeDomain.ClassType{}, sql.ErrNoRows
}

// Save implements the mock ClassTypeStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockClassTypeStore) Save(ctx context.Context, ct classTypeDomain.ClassType) error {
	if m.classTypes == nil {
		m.classTypes = make(map[string]classTypeDomain.ClassType)
	}
	m.classTypes[ct.ID] = ct
	return nil
}

// Delete implements the mock ClassTypeStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockClassTypeStore) Delete(ctx context.Context, id string) error {
	delete(m.classTypes, id)
	return nil
}

// List implements the mock ClassTypeStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockClassTypeStore) List(ctx context.Context) ([]classTypeDomain.ClassType, error) {
	var list []classTypeDomain.ClassType
	for _, ct := range m.classTypes {
		list = append(list, ct)
	}
	return list, nil
}

// ListByProgramID implements the mock ClassTypeStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockClassTypeStore) ListByProgramID(ctx context.Context, programID string) ([]classTypeDomain.ClassType, error) {
	var list []classTypeDomain.ClassType
	for _, ct := range m.classTypes {
		if ct.ProgramID == programID {
			list = append(list, ct)
		}
	}
	return list, nil
}

type mockProgramStore struct {
	programs map[string]programDomain.Program
}

// GetByID implements the mock ProgramStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockProgramStore) GetByID(ctx context.Context, id string) (programDomain.Program, error) {
	if p, ok := m.programs[id]; ok {
		return p, nil
	}
	return programDomain.Program{}, sql.ErrNoRows
}

// Save implements the mock ProgramStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockProgramStore) Save(ctx context.Context, p programDomain.Program) error {
	if m.programs == nil {
		m.programs = make(map[string]programDomain.Program)
	}
	m.programs[p.ID] = p
	return nil
}

// Delete implements the mock ProgramStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockProgramStore) Delete(ctx context.Context, id string) error {
	delete(m.programs, id)
	return nil
}

// List implements the mock ProgramStore for testing.
// PRE: valid parameters
// POST: returns expected result
func (m *mockProgramStore) List(ctx context.Context) ([]programDomain.Program, error) {
	var list []programDomain.Program
	for _, p := range m.programs {
		list = append(list, p)
	}
	return list, nil
}

// --- Test helpers ---

// newFullStores returns a Stores with all mock stores initialized.
func newFullStores() *Stores {
	return &Stores{
		AccountStore:             &mockAccountStore{accounts: make(map[string]accountDomain.Account)},
		MemberStore:              &mockMemberStore{members: make(map[string]memberDomain.Member)},
		WaiverStore:              &mockWaiverStore{waivers: make(map[string]waiverDomain.Waiver)},
		InjuryStore:              &mockInjuryStore{injuries: make(map[string]injuryDomain.Injury)},
		AttendanceStore:          &mockAttendanceStore{attendances: make(map[string]attendanceDomain.Attendance)},
		ProgramStore:             &mockProgramStore{programs: make(map[string]programDomain.Program)},
		ClassTypeStore:           &mockClassTypeStore{classTypes: make(map[string]classTypeDomain.ClassType)},
		ScheduleStore:            &mockScheduleStore{schedules: make(map[string]scheduleDomain.Schedule)},
		TermStore:                &mockTermStore{terms: make(map[string]termDomain.Term)},
		HolidayStore:             &mockHolidayStore{holidays: make(map[string]holidayDomain.Holiday)},
		NoticeStore:              &mockNoticeStore{notices: make(map[string]noticeDomain.Notice)},
		GradingRecordStore:       &mockGradingRecordStore{records: make(map[string]gradingDomain.Record)},
		GradingConfigStore:       &mockGradingConfigStore{configs: make(map[string]gradingDomain.Config)},
		GradingProposalStore:     &mockGradingProposalStore{proposals: make(map[string]gradingDomain.Proposal)},
		GradingMemberConfigStore: &mockGradingMemberConfigStore{configs: make(map[string]gradingDomain.MemberConfig)},
		MessageStore:             &mockMessageStore{messages: make(map[string]messageDomain.Message)},
		ObservationStore:         &mockObservationStore{observations: make(map[string]observationDomain.Observation)},
		MilestoneStore:           &mockMilestoneStore{milestones: make(map[string]milestoneDomain.Milestone)},
		MemberMilestoneStore:     &mockMemberMilestoneStore{items: make(map[string]milestoneDomain.MemberMilestone)},
		TrainingGoalStore:        &mockTrainingGoalStore{goals: make(map[string]trainingGoalDomain.TrainingGoal)},
	}
}

// authRequest returns a request with the given session injected into context.
func authRequest(method, url string, body string, sess middleware.Session) *http.Request {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, url, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, url, nil)
	}
	ctx := middleware.ContextWithSession(req.Context(), sess)
	return req.WithContext(ctx)
}

var adminSession = middleware.Session{
	AccountID: "admin-001",
	Email:     "admin@test.com",
	Role:      "admin",
	CreatedAt: time.Now(),
}

var coachSession = middleware.Session{
	AccountID: "coach-001",
	Email:     "coach@test.com",
	Role:      "coach",
	CreatedAt: time.Now(),
}

var memberSession = middleware.Session{
	AccountID: "member-001",
	Email:     "marcus@test.com",
	Role:      "member",
	CreatedAt: time.Now(),
}

// --- Tests: /api/notices ---

// TestHandleNotices_GET_Unauthenticated tests the corresponding handler.
func TestHandleNotices_GET_Unauthenticated(t *testing.T) {
	stores = newFullStores()
	req := httptest.NewRequest("GET", "/api/notices", nil)
	rec := httptest.NewRecorder()
	handleNotices(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

// TestHandleNotices_GET_AllNotices tests the corresponding handler.
func TestHandleNotices_GET_AllNotices(t *testing.T) {
	stores = newFullStores()
	stores.NoticeStore.Save(context.Background(), noticeDomain.Notice{
		ID: "n1", Type: noticeDomain.TypeSchoolWide, Status: noticeDomain.StatusPublished, Title: "Test",
	})

	req := authRequest("GET", "/api/notices", "", adminSession)
	rec := httptest.NewRecorder()
	handleNotices(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusOK)
	}
	var notices []noticeDomain.Notice
	json.NewDecoder(rec.Body).Decode(&notices)
	if len(notices) != 1 {
		t.Errorf("got %d notices, want 1", len(notices))
	}
}

// TestHandleNotices_GET_WithTypeFilter tests the corresponding handler.
func TestHandleNotices_GET_WithTypeFilter(t *testing.T) {
	stores = newFullStores()
	stores.NoticeStore.Save(context.Background(), noticeDomain.Notice{
		ID: "n1", Type: noticeDomain.TypeSchoolWide, Status: noticeDomain.StatusPublished,
	})
	stores.NoticeStore.Save(context.Background(), noticeDomain.Notice{
		ID: "n2", Type: noticeDomain.TypeClassSpecific, Status: noticeDomain.StatusPublished,
	})

	req := authRequest("GET", "/api/notices?type="+noticeDomain.TypeSchoolWide, "", adminSession)
	rec := httptest.NewRecorder()
	handleNotices(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusOK)
	}
	var notices []noticeDomain.Notice
	json.NewDecoder(rec.Body).Decode(&notices)
	if len(notices) != 1 {
		t.Errorf("got %d notices, want 1 (school_wide only)", len(notices))
	}
}

// TestHandleNotices_POST_Valid tests the corresponding handler.
func TestHandleNotices_POST_Valid(t *testing.T) {
	stores = newFullStores()
	body := `{"Type":"school_wide","Title":"Grading Day","Content":"Belt grading this Saturday"}`
	req := authRequest("POST", "/api/notices", body, adminSession)
	rec := httptest.NewRecorder()
	handleNotices(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("got %d, want %d. Body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
}

// TestHandleNotices_POST_Unauthenticated tests the corresponding handler.
func TestHandleNotices_POST_Unauthenticated(t *testing.T) {
	stores = newFullStores()
	body := `{"Type":"school_wide","Title":"Test","Content":"test"}`
	req := httptest.NewRequest("POST", "/api/notices", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handleNotices(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

// TestHandleNotices_MethodNotAllowed tests the corresponding handler.
func TestHandleNotices_MethodNotAllowed(t *testing.T) {
	stores = newFullStores()
	req := authRequest("DELETE", "/api/notices", "", adminSession)
	rec := httptest.NewRecorder()
	handleNotices(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("got %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

// TestHandleNotices_POST_WithNewFields tests creating a notice with all new fields.
func TestHandleNotices_POST_WithNewFields(t *testing.T) {
	stores = newFullStores()
	body := `{"Type":"school_wide","Title":"Open Mat","Content":"**Friday** 4-5:30pm","AuthorName":"Coach Pat","ShowAuthor":true,"Color":"blue","VisibleFrom":"2026-03-01T00:00:00Z","VisibleUntil":"2026-03-31T23:59:59Z"}`
	req := authRequest("POST", "/api/notices", body, adminSession)
	rec := httptest.NewRecorder()
	handleNotices(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("got %d, want %d. Body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var n noticeDomain.Notice
	json.NewDecoder(rec.Body).Decode(&n)
	if n.Color != "blue" {
		t.Errorf("expected color=blue, got %s", n.Color)
	}
	if n.AuthorName != "Coach Pat" {
		t.Errorf("expected AuthorName=Coach Pat, got %s", n.AuthorName)
	}
	if !n.ShowAuthor {
		t.Error("expected ShowAuthor=true")
	}
	if n.VisibleFrom.IsZero() {
		t.Error("expected VisibleFrom to be set")
	}
	if n.Status != noticeDomain.StatusDraft {
		t.Errorf("expected status=draft, got %s", n.Status)
	}
}

// TestHandleNotices_POST_InvalidColor tests that invalid color is rejected.
func TestHandleNotices_POST_InvalidColor(t *testing.T) {
	stores = newFullStores()
	body := `{"Type":"school_wide","Title":"Test","Content":"test","Color":"neon_pink"}`
	req := authRequest("POST", "/api/notices", body, adminSession)
	rec := httptest.NewRecorder()
	handleNotices(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// TestHandleNoticeEdit tests editing an existing notice.
func TestHandleNoticeEdit(t *testing.T) {
	stores = newFullStores()
	stores.NoticeStore.Save(context.Background(), noticeDomain.Notice{
		ID: "edit-1", Type: noticeDomain.TypeSchoolWide, Status: noticeDomain.StatusDraft,
		Title: "Original", Content: "Original content", CreatedBy: "admin",
		Color: noticeDomain.ColorOrange,
	})

	body := `{"NoticeID":"edit-1","Title":"Updated Title","Content":"Updated **content**","Color":"red","AuthorName":"Coach Marcus","ShowAuthor":true}`
	req := authRequest("POST", "/api/notices/edit", body, adminSession)
	rec := httptest.NewRecorder()
	handleNoticeEdit(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d. Body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var n noticeDomain.Notice
	json.NewDecoder(rec.Body).Decode(&n)
	if n.Title != "Updated Title" {
		t.Errorf("expected title=Updated Title, got %s", n.Title)
	}
	if n.Color != "red" {
		t.Errorf("expected color=red, got %s", n.Color)
	}
	if n.AuthorName != "Coach Marcus" {
		t.Errorf("expected AuthorName=Coach Marcus, got %s", n.AuthorName)
	}
	if n.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

// TestHandleNoticePin tests pinning and unpinning a notice.
func TestHandleNoticePin(t *testing.T) {
	stores = newFullStores()
	stores.NoticeStore.Save(context.Background(), noticeDomain.Notice{
		ID: "pin-1", Type: noticeDomain.TypeSchoolWide, Status: noticeDomain.StatusPublished,
		Title: "Pinnable", Content: "content", CreatedBy: "admin", Color: noticeDomain.ColorOrange,
	})

	t.Run("pin notice", func(t *testing.T) {
		body := `{"NoticeID":"pin-1","Pinned":true}`
		req := authRequest("POST", "/api/notices/pin", body, adminSession)
		rec := httptest.NewRecorder()
		handleNoticePin(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got %d, want %d. Body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		var n noticeDomain.Notice
		json.NewDecoder(rec.Body).Decode(&n)
		if !n.Pinned {
			t.Error("expected Pinned=true")
		}
	})

	t.Run("unpin notice", func(t *testing.T) {
		body := `{"NoticeID":"pin-1","Pinned":false}`
		req := authRequest("POST", "/api/notices/pin", body, adminSession)
		rec := httptest.NewRecorder()
		handleNoticePin(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got %d, want %d. Body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		var n noticeDomain.Notice
		json.NewDecoder(rec.Body).Decode(&n)
		if n.Pinned {
			t.Error("expected Pinned=false")
		}
	})
}

// TestHandleNotices_POST_NonAdmin tests that non-admin users cannot create notices.
func TestHandleNotices_POST_NonAdmin(t *testing.T) {
	stores = newFullStores()
	body := `{"Type":"school_wide","Title":"Test","Content":"test"}`
	req := authRequest("POST", "/api/notices", body, coachSession)
	rec := httptest.NewRecorder()
	handleNotices(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("got %d, want %d", rec.Code, http.StatusForbidden)
	}
}

// TestHandleNoticeEdit_Unauthenticated tests that unauthenticated users cannot edit notices.
func TestHandleNoticeEdit_Unauthenticated(t *testing.T) {
	stores = newFullStores()
	body := `{"NoticeID":"edit-1","Title":"Hacked"}`
	req := httptest.NewRequest("POST", "/api/notices/edit", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handleNoticeEdit(rec, req)

	if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusForbidden {
		t.Errorf("got %d, want 401 or 403", rec.Code)
	}
}

// TestHandleNoticeEdit_NonAdmin tests that non-admin users cannot edit notices.
func TestHandleNoticeEdit_NonAdmin(t *testing.T) {
	stores = newFullStores()
	body := `{"NoticeID":"edit-1","Title":"Hacked"}`
	req := authRequest("POST", "/api/notices/edit", body, memberSession)
	rec := httptest.NewRecorder()
	handleNoticeEdit(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("got %d, want %d", rec.Code, http.StatusForbidden)
	}
}

// --- Tests: /api/messages ---

// TestHandleMessages_GET_MissingMemberID tests the corresponding handler.
func TestHandleMessages_GET_MissingMemberID(t *testing.T) {
	stores = newFullStores()
	req := authRequest("GET", "/api/messages", "", memberSession)
	rec := httptest.NewRecorder()
	handleMessages(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// TestHandleMessages_GET_WithMemberID tests the corresponding handler.
func TestHandleMessages_GET_WithMemberID(t *testing.T) {
	stores = newFullStores()
	stores.MessageStore.Save(context.Background(), messageDomain.Message{
		ID: "msg1", SenderID: "admin-001", ReceiverID: "member-001",
		Subject: "Hello", Content: "Welcome!", CreatedAt: time.Now(),
	})

	req := authRequest("GET", "/api/messages?member_id=member-001", "", memberSession)
	rec := httptest.NewRecorder()
	handleMessages(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusOK)
	}
	var msgs []messageDomain.Message
	json.NewDecoder(rec.Body).Decode(&msgs)
	if len(msgs) != 1 {
		t.Errorf("got %d messages, want 1", len(msgs))
	}
}

// TestHandleMessages_GET_EmptyResult tests the corresponding handler.
func TestHandleMessages_GET_EmptyResult(t *testing.T) {
	stores = newFullStores()
	req := authRequest("GET", "/api/messages?member_id=nonexistent", "", memberSession)
	rec := httptest.NewRecorder()
	handleMessages(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "[]" {
		t.Errorf("got body %q, want []", rec.Body.String())
	}
}

// TestHandleMessages_POST_Valid tests the corresponding handler.
func TestHandleMessages_POST_Valid(t *testing.T) {
	stores = newFullStores()
	body := `{"ReceiverID":"member-001","Subject":"Test","Content":"Test message"}`
	req := authRequest("POST", "/api/messages", body, adminSession)
	rec := httptest.NewRecorder()
	handleMessages(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("got %d, want %d. Body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
}

// TestHandleMessages_POST_InvalidJSON tests the corresponding handler.
func TestHandleMessages_POST_InvalidJSON(t *testing.T) {
	stores = newFullStores()
	req := authRequest("POST", "/api/messages", "{bad json", adminSession)
	rec := httptest.NewRecorder()
	handleMessages(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// --- Tests: /api/observations ---

// TestHandleObservations_GET_MissingMemberID tests the corresponding handler.
func TestHandleObservations_GET_MissingMemberID(t *testing.T) {
	stores = newFullStores()
	req := authRequest("GET", "/api/observations", "", coachSession)
	rec := httptest.NewRecorder()
	handleObservations(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// TestHandleObservations_GET_WithMemberID tests the corresponding handler.
func TestHandleObservations_GET_WithMemberID(t *testing.T) {
	stores = newFullStores()
	stores.ObservationStore.Save(context.Background(), observationDomain.Observation{
		ID: "obs1", MemberID: "member-001", AuthorID: "coach-001",
		Content: "Good guard retention", CreatedAt: time.Now(),
	})

	req := authRequest("GET", "/api/observations?member_id=member-001", "", coachSession)
	rec := httptest.NewRecorder()
	handleObservations(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusOK)
	}
	var obs []observationDomain.Observation
	json.NewDecoder(rec.Body).Decode(&obs)
	if len(obs) != 1 {
		t.Errorf("got %d observations, want 1", len(obs))
	}
}

// TestHandleObservations_POST_Valid tests the corresponding handler.
func TestHandleObservations_POST_Valid(t *testing.T) {
	stores = newFullStores()
	body := `{"MemberID":"member-001","Content":"Excellent takedowns today"}`
	req := authRequest("POST", "/api/observations", body, coachSession)
	rec := httptest.NewRecorder()
	handleObservations(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("got %d, want %d. Body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
}

// TestHandleObservations_POST_MissingContent tests the corresponding handler.
func TestHandleObservations_POST_MissingContent(t *testing.T) {
	stores = newFullStores()
	body := `{"MemberID":"member-001","Content":""}`
	req := authRequest("POST", "/api/observations", body, coachSession)
	rec := httptest.NewRecorder()
	handleObservations(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// --- Tests: /api/grading/proposals ---

// TestHandleGradingProposals_GET_Empty tests the corresponding handler.
func TestHandleGradingProposals_GET_Empty(t *testing.T) {
	stores = newFullStores()
	req := authRequest("GET", "/api/grading/proposals", "", adminSession)
	rec := httptest.NewRecorder()
	handleGradingProposals(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "[]" {
		t.Errorf("got body %q, want []", rec.Body.String())
	}
}

// TestHandleGradingProposals_GET_WithPending tests the corresponding handler.
func TestHandleGradingProposals_GET_WithPending(t *testing.T) {
	stores = newFullStores()
	stores.GradingProposalStore.Save(context.Background(), gradingDomain.Proposal{
		ID: "p1", MemberID: "member-001", TargetBelt: gradingDomain.BeltBlue,
		ProposedBy: "coach-001", Status: gradingDomain.ProposalPending, CreatedAt: time.Now(),
	})

	req := authRequest("GET", "/api/grading/proposals", "", adminSession)
	rec := httptest.NewRecorder()
	handleGradingProposals(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusOK)
	}
	var proposals []gradingDomain.Proposal
	json.NewDecoder(rec.Body).Decode(&proposals)
	if len(proposals) != 1 {
		t.Errorf("got %d proposals, want 1", len(proposals))
	}
}

// TestHandleGradingProposals_POST_Valid tests the corresponding handler.
func TestHandleGradingProposals_POST_Valid(t *testing.T) {
	stores = newFullStores()
	body := `{"MemberID":"member-001","TargetBelt":"blue","Notes":"Ready for promotion"}`
	req := authRequest("POST", "/api/grading/proposals", body, coachSession)
	rec := httptest.NewRecorder()
	handleGradingProposals(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("got %d, want %d. Body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
}

// TestHandleGradingProposals_POST_MissingFields tests the corresponding handler.
func TestHandleGradingProposals_POST_MissingFields(t *testing.T) {
	stores = newFullStores()
	body := `{"MemberID":"","TargetBelt":"","Notes":""}`
	req := authRequest("POST", "/api/grading/proposals", body, coachSession)
	rec := httptest.NewRecorder()
	handleGradingProposals(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// --- Tests: /api/milestones ---

// TestHandleMilestones_GET_Empty tests the corresponding handler.
func TestHandleMilestones_GET_Empty(t *testing.T) {
	stores = newFullStores()
	req := authRequest("GET", "/api/milestones", "", adminSession)
	rec := httptest.NewRecorder()
	handleMilestones(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusOK)
	}
}

// TestHandleMilestones_GET_WithData tests the corresponding handler.
func TestHandleMilestones_GET_WithData(t *testing.T) {
	stores = newFullStores()
	stores.MilestoneStore.Save(context.Background(), milestoneDomain.Milestone{
		ID: "ms1", Name: "Century Club", Metric: "classes", Threshold: 100, BadgeIcon: "💯",
	})

	req := authRequest("GET", "/api/milestones", "", adminSession)
	rec := httptest.NewRecorder()
	handleMilestones(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusOK)
	}
	var milestones []milestoneDomain.Milestone
	json.NewDecoder(rec.Body).Decode(&milestones)
	if len(milestones) != 1 {
		t.Errorf("got %d milestones, want 1", len(milestones))
	}
}

// TestHandleMilestones_POST_Valid tests the corresponding handler.
func TestHandleMilestones_POST_Valid(t *testing.T) {
	stores = newFullStores()
	body := `{"Name":"First Class","Metric":"classes","Threshold":1,"BadgeIcon":"🥋"}`
	req := authRequest("POST", "/api/milestones", body, adminSession)
	rec := httptest.NewRecorder()
	handleMilestones(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("got %d, want %d. Body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
}

// --- Tests: /api/training-goals ---

// TestHandleTrainingGoals_GET_WithMemberID tests the corresponding handler.
func TestHandleTrainingGoals_GET_WithMemberID(t *testing.T) {
	stores = newFullStores()
	stores.TrainingGoalStore.Save(context.Background(), trainingGoalDomain.TrainingGoal{
		ID: "g1", MemberID: "member-001", Target: 3, Period: "weekly", Active: true,
	})

	req := authRequest("GET", "/api/training-goals?member_id=member-001", "", memberSession)
	rec := httptest.NewRecorder()
	handleTrainingGoals(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusOK)
	}
}

// TestHandleTrainingGoals_POST_Valid tests the corresponding handler.
func TestHandleTrainingGoals_POST_Valid(t *testing.T) {
	stores = newFullStores()
	body := `{"MemberID":"member-001","Target":5,"Period":"weekly"}`
	req := authRequest("POST", "/api/training-goals", body, memberSession)
	rec := httptest.NewRecorder()
	handleTrainingGoals(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("got %d, want %d. Body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
}

// --- Tests: /api/schedules (admin CRUD) ---

// TestHandleSchedules_GET_Unauthenticated tests the corresponding handler.
func TestHandleSchedules_GET_Unauthenticated(t *testing.T) {
	stores = newFullStores()
	req := httptest.NewRequest("GET", "/api/schedules", nil)
	rec := httptest.NewRecorder()
	handleSchedules(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

// TestHandleSchedules_GET_NonAdmin tests the corresponding handler.
func TestHandleSchedules_GET_NonAdmin(t *testing.T) {
	stores = newFullStores()
	req := authRequest("GET", "/api/schedules", "", memberSession)
	rec := httptest.NewRecorder()
	handleSchedules(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("got %d, want %d", rec.Code, http.StatusForbidden)
	}
}

// TestHandleSchedules_GET_Admin tests the corresponding handler.
func TestHandleSchedules_GET_Admin(t *testing.T) {
	stores = newFullStores()
	req := authRequest("GET", "/api/schedules", "", adminSession)
	rec := httptest.NewRecorder()
	handleSchedules(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d, want %d", rec.Code, http.StatusOK)
	}
}

// --- Tests: /api/terms (admin CRUD) ---

// TestHandleTerms_GET_Admin tests the corresponding handler.
func TestHandleTerms_GET_Admin(t *testing.T) {
	stores = newFullStores()
	stores.TermStore.Save(context.Background(), termDomain.Term{
		ID: "t1", Name: "Term 1", StartDate: time.Now(), EndDate: time.Now().AddDate(0, 3, 0),
	})
	req := authRequest("GET", "/api/terms", "", adminSession)
	rec := httptest.NewRecorder()
	handleTerms(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusOK)
	}
	var terms []termDomain.Term
	json.NewDecoder(rec.Body).Decode(&terms)
	if len(terms) != 1 {
		t.Errorf("got %d terms, want 1", len(terms))
	}
}

// --- Tests: /api/holidays (admin CRUD) ---

// TestHandleHolidays_GET_Admin tests the corresponding handler.
func TestHandleHolidays_GET_Admin(t *testing.T) {
	stores = newFullStores()
	stores.HolidayStore.Save(context.Background(), holidayDomain.Holiday{
		ID: "h1", Name: "Waitangi Day", StartDate: time.Now(), EndDate: time.Now(),
	})
	req := authRequest("GET", "/api/holidays", "", adminSession)
	rec := httptest.NewRecorder()
	handleHolidays(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusOK)
	}
	var holidays []holidayDomain.Holiday
	json.NewDecoder(rec.Body).Decode(&holidays)
	if len(holidays) != 1 {
		t.Errorf("got %d holidays, want 1", len(holidays))
	}
}

// --- Tests: /messages and /training-log page handlers ---

// TestHandleMessagesPage_Unauthenticated tests the corresponding handler.
func TestHandleMessagesPage_Unauthenticated(t *testing.T) {
	stores = newFullStores()
	req := httptest.NewRequest("GET", "/messages", nil)
	rec := httptest.NewRecorder()
	handleMessagesPage(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("got %d, want %d (redirect to login)", rec.Code, http.StatusSeeOther)
	}
	if loc := rec.Header().Get("Location"); loc != "/login" {
		t.Errorf("got redirect %q, want /login", loc)
	}
}

// TestHandleMessagesPage_MethodNotAllowed tests the corresponding handler.
func TestHandleMessagesPage_MethodNotAllowed(t *testing.T) {
	stores = newFullStores()
	req := authRequest("POST", "/messages", "", memberSession)
	rec := httptest.NewRecorder()
	handleMessagesPage(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("got %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

// TestHandleTrainingLogPage_Unauthenticated tests the corresponding handler.
func TestHandleTrainingLogPage_Unauthenticated(t *testing.T) {
	stores = newFullStores()
	req := httptest.NewRequest("GET", "/training-log", nil)
	rec := httptest.NewRecorder()
	handleTrainingLogPage(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("got %d, want %d (redirect to login)", rec.Code, http.StatusSeeOther)
	}
}

// TestHandleTrainingLogPage_MethodNotAllowed tests the corresponding handler.
func TestHandleTrainingLogPage_MethodNotAllowed(t *testing.T) {
	stores = newFullStores()
	req := authRequest("POST", "/training-log", "", memberSession)
	rec := httptest.NewRecorder()
	handleTrainingLogPage(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("got %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

// Suppress unused import warnings — these stores are used in newFullStores
var (
	_ accountStore.Store
	_ attendanceStore.Store
	_ classTypeStore.Store
	_ gradingStore.RecordStore
	_ gradingStore.ConfigStore
	_ gradingStore.ProposalStore
	_ holidayStore.Store
	_ injuryStore.Store
	_ memberStore.Store
	_ messageStore.Store
	_ milestoneStore.Store
	_ noticeStore.Store
	_ observationStore.Store
	_ programStore.Store
	_ scheduleStore.Store
	_ termStore.Store
	_ trainingGoalStore.Store
	_ waiverStore.Store
)

// --- Tests: /api/emails/recipients/by-session ---

// TestHandleRecipientsFilterBySession_Success tests filtering recipients by a specific class session.
func TestHandleRecipientsFilterBySession_Success(t *testing.T) {
	stores = newFullStores()
	ctx := context.Background()

	// Seed a member
	stores.MemberStore.Save(ctx, memberDomain.Member{ID: "m1", Name: "Alice", Email: "alice@test.com", Status: "active"})

	// Seed an attendance record
	stores.AttendanceStore.Save(ctx, attendanceDomain.Attendance{
		ID: "a1", MemberID: "m1", ScheduleID: "s1", ClassDate: "2026-02-09",
		CheckInTime: time.Now(),
	})

	req := authRequest("GET", "/api/emails/recipients/by-session?scheduleID=s1&date=2026-02-09", "", adminSession)
	rec := httptest.NewRecorder()
	handleRecipientsFilterBySession(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d. Body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	type memberResult struct {
		ID    string `json:"ID"`
		Name  string `json:"Name"`
		Email string `json:"Email"`
	}
	var results []memberResult
	json.NewDecoder(rec.Body).Decode(&results)
	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
	if len(results) > 0 && results[0].Name != "Alice" {
		t.Errorf("got name %q, want Alice", results[0].Name)
	}
}

// TestHandleRecipientsFilterBySession_MissingParams tests missing query params return 400.
func TestHandleRecipientsFilterBySession_MissingParams(t *testing.T) {
	stores = newFullStores()
	req := authRequest("GET", "/api/emails/recipients/by-session", "", adminSession)
	rec := httptest.NewRecorder()
	handleRecipientsFilterBySession(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// TestHandleRecipientsFilterBySession_NoMatch tests empty result when no attendance matches.
func TestHandleRecipientsFilterBySession_NoMatch(t *testing.T) {
	stores = newFullStores()
	req := authRequest("GET", "/api/emails/recipients/by-session?scheduleID=s1&date=2026-01-01", "", adminSession)
	rec := httptest.NewRecorder()
	handleRecipientsFilterBySession(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "[]" {
		t.Errorf("got %q, want []", rec.Body.String())
	}
}

// --- Tests: /api/emails/recipients/by-class-type ---

// TestHandleRecipientsFilterByClassType_Success tests filtering by class type with lookback.
func TestHandleRecipientsFilterByClassType_Success(t *testing.T) {
	stores = newFullStores()
	ctx := context.Background()

	// Seed class type, schedule, member, and attendance
	stores.ClassTypeStore.Save(ctx, classTypeDomain.ClassType{ID: "ct1", ProgramID: "p1", Name: "Fundamentals"})
	stores.ScheduleStore.Save(ctx, scheduleDomain.Schedule{ID: "s1", ClassTypeID: "ct1", Day: "monday", StartTime: "06:00", EndTime: "07:30"})
	stores.MemberStore.Save(ctx, memberDomain.Member{ID: "m1", Name: "Bob", Email: "bob@test.com", Status: "active"})
	stores.AttendanceStore.Save(ctx, attendanceDomain.Attendance{
		ID: "a1", MemberID: "m1", ScheduleID: "s1", ClassDate: time.Now().Format("2006-01-02"),
		CheckInTime: time.Now(),
	})

	req := authRequest("GET", "/api/emails/recipients/by-class-type?classTypeID=ct1&days=7", "", adminSession)
	rec := httptest.NewRecorder()
	handleRecipientsFilterByClassType(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d. Body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	type memberResult struct {
		ID    string `json:"ID"`
		Name  string `json:"Name"`
		Email string `json:"Email"`
	}
	var results []memberResult
	json.NewDecoder(rec.Body).Decode(&results)
	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
	if len(results) > 0 && results[0].Name != "Bob" {
		t.Errorf("got name %q, want Bob", results[0].Name)
	}
}

// TestHandleRecipientsFilterByClassType_MissingClassTypeID tests missing classTypeID returns 400.
func TestHandleRecipientsFilterByClassType_MissingClassTypeID(t *testing.T) {
	stores = newFullStores()
	req := authRequest("GET", "/api/emails/recipients/by-class-type", "", adminSession)
	rec := httptest.NewRecorder()
	handleRecipientsFilterByClassType(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// TestHandleRecipientsFilterByClassType_NoSchedules tests class type with no schedules returns empty.
func TestHandleRecipientsFilterByClassType_NoSchedules(t *testing.T) {
	stores = newFullStores()
	req := authRequest("GET", "/api/emails/recipients/by-class-type?classTypeID=nonexistent&days=30", "", adminSession)
	rec := httptest.NewRecorder()
	handleRecipientsFilterByClassType(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "[]" {
		t.Errorf("got %q, want []", rec.Body.String())
	}
}

// TestHandleRecipientsFilterByClassType_Unauthenticated tests that unauthenticated requests are rejected.
func TestHandleRecipientsFilterByClassType_Unauthenticated(t *testing.T) {
	stores = newFullStores()
	req := httptest.NewRequest("GET", "/api/emails/recipients/by-class-type?classTypeID=ct1&days=30", nil)
	rec := httptest.NewRecorder()
	handleRecipientsFilterByClassType(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
