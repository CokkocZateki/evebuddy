package app

import (
	"time"

	"github.com/ErikKalkoken/evebuddy/internal/optional"
)

type CorporationMessage struct {
	CorporationID int32
	Message       string
	SourceURL     optional.Optional[string]
	UpdatedAt     time.Time
	UpdatedBy     *EveEntity
}

type UpdateCorporationMessageParams struct {
	CorporationID int32
	CharacterID   int32
	Message       string
	SourceURL     optional.Optional[string]
}
