package usecase

import (
	"context"
	"errors"
	"testing"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
	announcementdomain "apihorpug/internal/features/announcement/domain"

	"github.com/google/uuid"
)

type fakeRepo struct {
	stored announcementdomain.Announcement
	// getErr makes GetByID fail, as it does for a missing or out-of-scope row.
	getErr error
	// deleted / updated record that the mutation actually reached the repo.
	deleted bool
	updated bool
}

func (r *fakeRepo) Summary(context.Context, uuid.UUID) (announcementdomain.Summary, error) {
	return announcementdomain.Summary{}, nil
}
func (r *fakeRepo) MarkRead(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (r *fakeRepo) Count(context.Context, uuid.UUID, ListFilters) (int64, error) {
	return 0, nil
}
func (r *fakeRepo) List(context.Context, uuid.UUID, ListFilters, int, int) ([]announcementdomain.Announcement, error) {
	return nil, nil
}
func (r *fakeRepo) GetByID(context.Context, uuid.UUID, uuid.UUID) (announcementdomain.Announcement, error) {
	if r.getErr != nil {
		return announcementdomain.Announcement{}, r.getErr
	}
	return r.stored, nil
}
func (r *fakeRepo) Create(_ context.Context, input CreateInput) (announcementdomain.Announcement, error) {
	return announcementdomain.Announcement{
		ID:          uuid.New(),
		DormitoryID: input.DormitoryID,
		Title:       input.Title,
		Category:    input.Category,
		IsPinned:    input.IsPinned,
		IsPublished: *input.IsPublished,
	}, nil
}
func (r *fakeRepo) Update(_ context.Context, _ uuid.UUID, _ uuid.UUID, input UpdateInput) (announcementdomain.Announcement, error) {
	r.updated = true
	after := r.stored
	if input.IsPublished != nil {
		after.IsPublished = *input.IsPublished
	}
	if input.IsPinned != nil {
		after.IsPinned = *input.IsPinned
	}
	if input.Category != nil {
		after.Category = *input.Category
	}
	return after, nil
}
func (r *fakeRepo) Delete(context.Context, uuid.UUID, uuid.UUID) error {
	r.deleted = true
	return nil
}

type fakeLogger struct {
	entries []activitylogusecase.CreateInput
	err     error
}

func (l *fakeLogger) Create(_ context.Context, input activitylogusecase.CreateInput) (activitylogdomain.ActivityLog, error) {
	l.entries = append(l.entries, input)
	return activitylogdomain.ActivityLog{}, l.err
}

func newStored() announcementdomain.Announcement {
	return announcementdomain.Announcement{
		ID:          uuid.New(),
		DormitoryID: uuid.New(),
		Title:       "Water shutdown",
		Category:    announcementdomain.CategoryMaintenance,
		IsPublished: false,
	}
}

func TestCreateRecordsActivity(t *testing.T) {
	logger := &fakeLogger{}
	svc := New(&fakeRepo{}, logger)
	requester := uuid.New()
	dormitory := uuid.New()

	created, err := svc.Create(context.Background(), CreateInput{
		DormitoryID: dormitory,
		Title:       "  Rooftop party  ",
		Category:    announcementdomain.CategoryEvent,
		CreatedBy:   &requester,
	}, "10.0.0.7")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if len(logger.entries) != 1 {
		t.Fatalf("want 1 log entry, got %d", len(logger.entries))
	}
	entry := logger.entries[0]
	if entry.Action != "CREATE" || entry.EntityType != "announcement" {
		t.Errorf("action/entity = %s/%s", entry.Action, entry.EntityType)
	}
	if entry.EntityID == nil || *entry.EntityID != created.ID {
		t.Errorf("entity id = %v, want %v", entry.EntityID, created.ID)
	}
	if entry.DormitoryID == nil || *entry.DormitoryID != dormitory {
		t.Errorf("dormitory id = %v, want %v", entry.DormitoryID, dormitory)
	}
	if entry.UserID == nil || *entry.UserID != requester {
		t.Errorf("user id = %v, want %v", entry.UserID, requester)
	}
	if entry.IPAddress != "10.0.0.7" {
		t.Errorf("ip = %q", entry.IPAddress)
	}
	if entry.Description != "Created announcement: Rooftop party (event, published)" {
		t.Errorf("description = %q", entry.Description)
	}
}

func TestUpdateRecordsWhatChanged(t *testing.T) {
	repo := &fakeRepo{stored: newStored()}
	logger := &fakeLogger{}
	svc := New(repo, logger)
	requester := uuid.New()

	published, pinned := true, true
	if _, err := svc.Update(context.Background(), repo.stored.ID, requester, UpdateInput{
		IsPublished: &published,
		IsPinned:    &pinned,
	}, "10.0.0.7"); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if len(logger.entries) != 1 {
		t.Fatalf("want 1 log entry, got %d", len(logger.entries))
	}
	entry := logger.entries[0]
	want := "Updated announcement: Water shutdown (status draft -> published, pinned false -> true)"
	if entry.Action != "UPDATE" || entry.Description != want {
		t.Errorf("got %s %q, want UPDATE %q", entry.Action, entry.Description, want)
	}
	if entry.UserID == nil || *entry.UserID != requester {
		t.Errorf("user id = %v, want %v", entry.UserID, requester)
	}
}

func TestUpdateWithoutNotableChangeHasNoSuffix(t *testing.T) {
	repo := &fakeRepo{stored: newStored()}
	logger := &fakeLogger{}
	svc := New(repo, logger)

	content := "new wording"
	if _, err := svc.Update(context.Background(), repo.stored.ID, uuid.New(), UpdateInput{Content: &content}, ""); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if got := logger.entries[0].Description; got != "Updated announcement: Water shutdown" {
		t.Errorf("description = %q", got)
	}
}

func TestDeleteRecordsTitleAfterDeleting(t *testing.T) {
	repo := &fakeRepo{stored: newStored()}
	logger := &fakeLogger{}
	svc := New(repo, logger)

	if err := svc.Delete(context.Background(), repo.stored.ID, uuid.New(), "10.0.0.7"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if !repo.deleted {
		t.Fatal("repo.Delete was not called")
	}
	if len(logger.entries) != 1 || logger.entries[0].Description != "Deleted announcement: Water shutdown" {
		t.Errorf("entries = %+v", logger.entries)
	}
	if logger.entries[0].Action != "DELETE" {
		t.Errorf("action = %s", logger.entries[0].Action)
	}
}

// A missing or out-of-scope announcement must fail before anything is changed
// and without leaving a log entry claiming it happened.
func TestUpdateAndDeleteOfMissingAnnouncementLogNothing(t *testing.T) {
	repo := &fakeRepo{getErr: announcementdomain.ErrAnnouncementNotFound}
	logger := &fakeLogger{}
	svc := New(repo, logger)

	title := "x"
	if _, err := svc.Update(context.Background(), uuid.New(), uuid.New(), UpdateInput{Title: &title}, ""); !errors.Is(err, announcementdomain.ErrAnnouncementNotFound) {
		t.Errorf("Update err = %v", err)
	}
	if err := svc.Delete(context.Background(), uuid.New(), uuid.New(), ""); !errors.Is(err, announcementdomain.ErrAnnouncementNotFound) {
		t.Errorf("Delete err = %v", err)
	}

	if repo.updated || repo.deleted {
		t.Error("a mutation reached the repository")
	}
	if len(logger.entries) != 0 {
		t.Errorf("logged %d entries for failed operations", len(logger.entries))
	}
}

// The audit trail is best-effort: a failing logger must not fail the action.
func TestLoggerFailureDoesNotFailMutation(t *testing.T) {
	repo := &fakeRepo{stored: newStored()}
	svc := New(repo, &fakeLogger{err: errors.New("db down")})

	if err := svc.Delete(context.Background(), repo.stored.ID, uuid.New(), ""); err != nil {
		t.Fatalf("Delete failed because logging failed: %v", err)
	}
	if !repo.deleted {
		t.Error("repo.Delete was not called")
	}
}

func TestNilLoggerIsAllowed(t *testing.T) {
	repo := &fakeRepo{stored: newStored()}
	svc := New(repo, nil)

	if err := svc.Delete(context.Background(), repo.stored.ID, uuid.New(), ""); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}
