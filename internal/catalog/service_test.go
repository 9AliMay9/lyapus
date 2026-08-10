package catalog

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestTeamServiceCreateTeamNormalizeAndDelegates(t *testing.T) {
	repository := &recordingTeamRepository{
		team: Team{
			ID:   1,
			Slug: "platform",
			Name: "Platform",
		},
	}
	service := NewTeamService(repository)

	got, err := service.CreateTeam(context.Background(), CreateTeamInput{
		Slug: "platform",
		Name: "	Platform	",
	})
	if err != nil {
		t.Fatalf("CreateTeam() error = %v", err)
	}

	if got != repository.team {
		t.Fatalf("CreateTeam() = %#v, want %#v", got, repository.team)
	}
	if repository.createInput == nil {
		t.Fatal("CreateTeam() did not call repository")
	}

	want := CreateTeamInput{
		Slug: "platform",
		Name: "Platform",
	}
	if *repository.createInput != want {
		t.Fatalf("CreateTeam() repository input = %#v, want %#v", *repository.createInput, want)
	}
}

func TestTeamServiceUpdateTeamNormalizesPartialInput(t *testing.T) {
	repository := &recordingTeamRepository{}
	service := NewTeamService(repository)

	name := "	Platform Engineering	"
	_, err := service.UpdateTeam(context.Background(), 1, UpdateTeamInput{
		Name: &name,
	})
	if err != nil {
		t.Fatalf("UpdateTeam() error = %v", err)
	}

	if repository.updateInput == nil {
		t.Fatal("UpdateTeam() did not call repository")
	}
	if repository.updateID != 1 {
		t.Fatalf("UpdateTeam() repository ID = %d, want 1", repository.updateID)
	}
	if repository.updateInput.Slug != nil {
		t.Fatalf("UpdateTeam() repository slug = %q, want nil", *repository.updateInput.Slug)
	}
	if repository.updateInput.Name == nil {
		t.Fatal("UpdateTeam() repository name = nil, want value")
	}
	if *repository.updateInput.Name != "Platform Engineering" {
		t.Fatalf("UpdateTeam() repository name = %q, want %q", *repository.updateInput.Name, "Platform Engineering")
	}
}

func TestTeamServiceListTeamsDefaultsAndNormalizesCursor(t *testing.T) {
	repository := &recordingTeamRepository{}
	service := NewTeamService(repository)

	cursor := TeamCursor{
		CreatedAt: time.Date(2026, time.August, 6, 12, 0, 0, 0, time.FixedZone("CST", 8*60*60)),
		ID:        42,
	}

	_, err := service.ListTeams(context.Background(), ListTeamsInput{
		After: &cursor,
	})
	if err != nil {
		t.Fatalf("ListTeams() error = %v", err)
	}

	if repository.listInput == nil {
		t.Fatal("ListTeams() did not call repository")
	}
	if repository.listInput.Limit != defaultTeamListLimit {
		t.Fatalf("ListTeams() repository limit = %d, want %d", repository.listInput.Limit, defaultTeamListLimit)
	}
	if repository.listInput.After == nil {
		t.Fatal("ListTeams() repository cursor = nil, want value")
	}
	if repository.listInput.After.ID != cursor.ID {
		t.Fatalf("ListTeams() repository cursor ID = %d, want %d", repository.listInput.After.ID, cursor.ID)
	}
	if repository.listInput.After.CreatedAt.Location() != time.UTC {
		t.Fatalf("ListTeams() repository cursor location = %s, want UTC", repository.listInput.After.CreatedAt.Location())
	}
	if !repository.listInput.After.CreatedAt.Equal(cursor.CreatedAt) {
		t.Fatalf("ListTeams() repository cursor time = %s, want %s", repository.listInput.After.CreatedAt, cursor.CreatedAt)
	}
}

func TestTeamServiceGetAndDeleteDelegate(t *testing.T) {
	repository := &recordingTeamRepository{
		team: Team{ID: 7},
	}
	service := NewTeamService(repository)

	got, err := service.GetTeamByID(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetTeamByID() error = %v", err)
	}
	if got != repository.team {
		t.Fatalf("GetTeamByID() = %#v, want %#v", got, repository.team)
	}
	if repository.getID != 7 {
		t.Fatalf("GetTeamByID() repository ID = %d, want 7", repository.getID)
	}

	if err := service.DeleteTeam(context.Background(), 7); err != nil {
		t.Fatalf("DeleteTeam() error = %v", err)
	}
	if repository.deleteID != 7 {
		t.Fatalf("DeleteTeam() repository ID = %d, want 7", repository.deleteID)
	}
}

func TestTeamServiceRejectsInvalidInputBeforeRepository(t *testing.T) {
	tests := []struct {
		name string
		call func(*TeamService) error
	}{
		{
			name: "invalid create slug",
			call: func(service *TeamService) error {
				_, err := service.CreateTeam(context.Background(), CreateTeamInput{
					Slug: "Platform",
					Name: "Platform",
				})
				return err
			},
		},
		{
			name: "zero get ID",
			call: func(service *TeamService) error {
				_, err := service.GetTeamByID(context.Background(), 0)
				return err
			},
		},
		{
			name: "too large list limit",
			call: func(service *TeamService) error {
				_, err := service.ListTeams(context.Background(), ListTeamsInput{
					Limit: maxTeamListLimit + 1,
				})
				return err
			},
		},
		{
			name: "empty update",
			call: func(service *TeamService) error {
				_, err := service.UpdateTeam(context.Background(), 1, UpdateTeamInput{})
				return err
			},
		},
		{
			name: "zero delete ID",
			call: func(service *TeamService) error {
				return service.DeleteTeam(context.Background(), 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &recordingTeamRepository{}
			service := NewTeamService(repository)

			err := tt.call(service)

			if !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("error = %v, want ErrInvalidArgument", err)
			}
			if repository.callCount() != 0 {
				t.Fatalf("repository call count = %d, want 0", repository.callCount())
			}
		})
	}
}

func TestNormalizeTeamSlug(t *testing.T) {
	tests := []struct {
		name    string
		slug    string
		wantErr bool
	}{
		{
			name: "valid",
			slug: "platform-engineering",
		},
		{
			name:    "uppercase",
			slug:    "Platform",
			wantErr: true,
		},
		{
			name:    "leading hyphen",
			slug:    "-platform",
			wantErr: true,
		},
		{
			name:    "too long",
			slug:    "a" + strings.Repeat("b", 63),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeTeamSlug(tt.slug)

			if tt.wantErr {
				if !errors.Is(err, ErrInvalidArgument) {
					t.Fatalf("normalizeTeamSlug() error = %v, want ErrInvalidArgument", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeTeamSlug() error = %v", err)
			}
			if got != tt.slug {
				t.Fatalf("normalizeTeamSlug() = %q, want %q", got, tt.slug)
			}
		})
	}
}

func TestNormalizeTeamName(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{
			name:  "trim whitespace",
			value: "	Platform Engineering	",
			want:  "Platform Engineering",
		},
		{
			name:    "empty after trim",
			value:   " \t ",
			wantErr: true,
		},
		{
			name:  "one hundred unicode characters",
			value: strings.Repeat("界", 100),
			want:  strings.Repeat("界", 100),
		},
		{
			name:    "one hundred and one unicode characters",
			value:   strings.Repeat("界", 101),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeTeamName(tt.value)

			if tt.wantErr {
				if !errors.Is(err, ErrInvalidArgument) {
					t.Fatalf("normalizeTeamName() error = %v, want ErrInvalidArgument", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeTeamName() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("normalizeTeamName() = %q, want %q", got, tt.want)
			}
		})
	}
}

type recordingTeamRepository struct {
	createInput *CreateTeamInput
	getID       int64
	listInput   *ListTeamsInput
	updateID    int64
	updateInput *UpdateTeamInput
	deleteID    int64
	team        Team
	page        TeamPage
	err         error
	calls       int
}

func (r *recordingTeamRepository) CreateTeam(_ context.Context, input CreateTeamInput) (Team, error) {
	r.calls++
	r.createInput = &input
	return r.team, r.err
}

func (r *recordingTeamRepository) GetTeamByID(_ context.Context, id int64) (Team, error) {
	r.calls++
	r.getID = id
	return r.team, r.err
}

func (r *recordingTeamRepository) ListTeams(_ context.Context, input ListTeamsInput) (TeamPage, error) {
	r.calls++
	r.listInput = &input
	return r.page, r.err
}

func (r *recordingTeamRepository) UpdateTeam(_ context.Context, id int64, input UpdateTeamInput) (Team, error) {
	r.calls++
	r.updateID = id
	r.updateInput = &input
	return r.team, r.err
}

func (r *recordingTeamRepository) DeleteTeam(_ context.Context, id int64) error {
	r.calls++
	r.deleteID = id
	return r.err
}

func (r *recordingTeamRepository) callCount() int {
	return r.calls
}
