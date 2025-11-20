package corporationservice_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ErikKalkoken/evebuddy/internal/app"
	"github.com/ErikKalkoken/evebuddy/internal/app/corporationservice"
	"github.com/ErikKalkoken/evebuddy/internal/app/storage/testutil"
	"github.com/ErikKalkoken/evebuddy/internal/optional"
)

func TestUpdateCorporationMessage(t *testing.T) {
	db, st, factory := testutil.NewDBOnDisk(t)
	defer db.Close()
	ctx := context.Background()

	character := factory.CreateCharacter()
	corporation := factory.CreateCorporation(character.EveCharacter.Corporation.ID)
	directorRole := []app.CharacterRole{{CharacterID: character.ID, Role: app.RoleDirector, Granted: true}}
	fakeCS := &corporationservice.CharacterServiceFake{Character: character, Roles: map[int32][]app.CharacterRole{character.ID: directorRole}}
	s := corporationservice.NewFake(st, corporationservice.Params{CharacterService: fakeCS})

	t.Run("stores messages for directors", func(t *testing.T) {
		testutil.TruncateTables(db)
		factory.CreateCorporation(corporation.ID)

		msg, err := s.UpdateCorporationMessage(ctx, app.UpdateCorporationMessageParams{
			CorporationID: corporation.ID,
			CharacterID:   character.ID,
			Message:       "Standby for fleet ops",
			SourceURL:     optional.New("https://example.com/motd"),
		})
		if assert.NoError(t, err) {
			assert.Equal(t, character.ID, msg.UpdatedBy.ID)
			assert.Equal(t, "Standby for fleet ops", msg.Message)
			assert.Equal(t, "https://example.com/motd", msg.SourceURL.MustValue())
		}
	})

	t.Run("rejects non directors", func(t *testing.T) {
		testutil.TruncateTables(db)
		factory.CreateCorporation(corporation.ID)
		fakeCS.Roles = map[int32][]app.CharacterRole{character.ID: []app.CharacterRole{}}

		_, err := s.UpdateCorporationMessage(ctx, app.UpdateCorporationMessageParams{
			CorporationID: corporation.ID,
			CharacterID:   character.ID,
			Message:       "Hello",
		})
		assert.ErrorIs(t, err, app.ErrInvalid)
	})
}

func TestFindDirectorCharacter(t *testing.T) {
	db, st, factory := testutil.NewDBOnDisk(t)
	defer db.Close()
	ctx := context.Background()

	director := factory.CreateCharacter()
	corp := factory.CreateCorporation(director.EveCharacter.Corporation.ID)
	nonDirector := factory.CreateCharacter()
	fakeCS := &corporationservice.CharacterServiceFake{
		Characters: []*app.Character{director, nonDirector},
		Roles: map[int32][]app.CharacterRole{
			director.ID:    []app.CharacterRole{{CharacterID: director.ID, Role: app.RoleDirector, Granted: true}},
			nonDirector.ID: []app.CharacterRole{},
		},
	}
	s := corporationservice.NewFake(st, corporationservice.Params{CharacterService: fakeCS})
	factory.CreateCorporation(nonDirector.EveCharacter.Corporation.ID)

	t.Run("returns director when available", func(t *testing.T) {
		got, err := s.FindDirectorCharacter(ctx, corp.ID)
		if assert.NoError(t, err) {
			assert.Equal(t, director.ID, got.ID)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		_, err := s.FindDirectorCharacter(ctx, 123)
		assert.ErrorIs(t, err, app.ErrNotFound)
	})
}

func TestFetchCorporationMessageFromURL(t *testing.T) {
	db, st, _ := testutil.NewDBOnDisk(t)
	defer db.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world\n"))
	}))
	defer server.Close()

	s := corporationservice.NewFake(st, corporationservice.Params{HTTPClient: server.Client()})
	msg, err := s.FetchCorporationMessageFromURL(context.Background(), server.URL)
	if assert.NoError(t, err) {
		assert.Equal(t, "hello world", msg)
	}
}
