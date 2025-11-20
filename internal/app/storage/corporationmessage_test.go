package storage_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ErikKalkoken/evebuddy/internal/app"
	"github.com/ErikKalkoken/evebuddy/internal/app/storage"
	"github.com/ErikKalkoken/evebuddy/internal/app/storage/testutil"
	"github.com/ErikKalkoken/evebuddy/internal/optional"
)

func TestCorporationMessage(t *testing.T) {
	db, st, factory := testutil.NewDBInMemory()
	defer db.Close()
	ctx := context.Background()

	t.Run("can store and retrieve", func(t *testing.T) {
		testutil.TruncateTables(db)
		character := factory.CreateCharacter()
		corporation := factory.CreateCorporation(character.EveCharacter.Corporation.ID)

		msg, err := st.UpdateCorporationMessage(ctx, storage.UpdateCorporationMessageParams{
			CorporationID: corporation.ID,
			Message:       "Hello pilots!",
			UpdatedBy:     optional.New(character.ID),
		})
		if assert.NoError(t, err) {
			assert.Equal(t, corporation.ID, msg.CorporationID)
			assert.Equal(t, "Hello pilots!", msg.Message)
			if assert.NotNil(t, msg.UpdatedBy) {
				assert.Equal(t, character.ID, msg.UpdatedBy.ID)
			}
		}

		stored, err := st.GetCorporationMessage(ctx, corporation.ID)
		if assert.NoError(t, err) {
			assert.Equal(t, msg.Message, stored.Message)
			if assert.NotNil(t, stored.UpdatedBy) {
				assert.Equal(t, msg.UpdatedBy.ID, stored.UpdatedBy.ID)
			}
		}
	})

	t.Run("stores source URL optionally", func(t *testing.T) {
		testutil.TruncateTables(db)
		character := factory.CreateCharacter()
		corporation := factory.CreateCorporation(character.EveCharacter.Corporation.ID)
		sourceURL := optional.New("https://example.com/motd.txt")

		msg, err := st.UpdateCorporationMessage(ctx, storage.UpdateCorporationMessageParams{
			CorporationID: corporation.ID,
			Message:       "Welcome",
			SourceURL:     sourceURL,
			UpdatedBy:     optional.New(character.ID),
		})
		if assert.NoError(t, err) {
			assert.Equal(t, sourceURL, msg.SourceURL)
		}
	})

	t.Run("returns error for invalid params", func(t *testing.T) {
		testutil.TruncateTables(db)
		_, err := st.UpdateCorporationMessage(ctx, storage.UpdateCorporationMessageParams{})
		assert.ErrorIs(t, err, app.ErrInvalid)
	})
}
