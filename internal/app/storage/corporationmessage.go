package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/ErikKalkoken/evebuddy/internal/app"
	"github.com/ErikKalkoken/evebuddy/internal/app/storage/queries"
	"github.com/ErikKalkoken/evebuddy/internal/optional"
)

type UpdateCorporationMessageParams struct {
	CorporationID int32
	Message       string
	SourceURL     optional.Optional[string]
	UpdatedBy     optional.Optional[int32]
}

func (st *Storage) GetCorporationMessage(ctx context.Context, corporationID int32) (*app.CorporationMessage, error) {
	wrapErr := func(err error) error {
		return fmt.Errorf("GetCorporationMessage %d: %w", corporationID, err)
	}
	if corporationID == 0 {
		return nil, wrapErr(app.ErrInvalid)
	}
	r, err := st.qRO.GetCorporationMessage(ctx, int64(corporationID))
	if err != nil {
		return nil, wrapErr(convertGetError(err))
	}
	o := corporationMessageFromDBModel(r)
	return o, nil
}

func (st *Storage) UpdateCorporationMessage(ctx context.Context, arg UpdateCorporationMessageParams) (*app.CorporationMessage, error) {
	wrapErr := func(err error) error {
		return fmt.Errorf("UpdateCorporationMessage %+v: %w", arg, err)
	}
	if arg.CorporationID == 0 || arg.Message == "" {
		return nil, wrapErr(app.ErrInvalid)
	}
	arg2 := queries.UpdateCorporationMessageParams{
		CorporationID: int64(arg.CorporationID),
		Message:       arg.Message,
		SourceUrl:     optional.ToNullString(arg.SourceURL),
		UpdatedAt:     time.Now().UTC(),
		UpdatedBy:     optional.ToNullInt64(arg.UpdatedBy),
	}
	if _, err := st.qRW.UpdateCorporationMessage(ctx, arg2); err != nil {
		return nil, wrapErr(err)
	}
	return st.GetCorporationMessage(ctx, arg.CorporationID)
}

func corporationMessageFromDBModel(r queries.GetCorporationMessageRow) *app.CorporationMessage {
	var updatedBy *app.EveEntity
	if r.UpdatedBy.Valid {
		name := r.UpdatedByName.String
		updatedBy = &app.EveEntity{ID: int32(r.UpdatedBy.Int64), Name: name, Category: app.EveEntityCharacter}
	}
	return &app.CorporationMessage{
		CorporationID: int32(r.CorporationID),
		Message:       r.Message,
		SourceURL:     optional.FromNullString(r.SourceUrl),
		UpdatedAt:     r.UpdatedAt,
		UpdatedBy:     updatedBy,
	}
}
