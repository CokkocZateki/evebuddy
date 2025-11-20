package corporationservice

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ErikKalkoken/evebuddy/internal/app"
	"github.com/ErikKalkoken/evebuddy/internal/app/storage"
	"github.com/ErikKalkoken/evebuddy/internal/optional"
)

func (s *CorporationService) FetchCorporationMessageFromURL(ctx context.Context, rawURL string) (string, error) {
	wrapErr := func(err error) error {
		return fmt.Errorf("fetch corporation message: %w", err)
	}
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", wrapErr(app.ErrInvalid)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", wrapErr(err)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", wrapErr(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return "", wrapErr(fmt.Errorf("unexpected status code %d", resp.StatusCode))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", wrapErr(err)
	}
	message := strings.TrimSpace(string(body))
	if message == "" {
		return "", wrapErr(app.ErrInvalid)
	}
	return message, nil
}

func (s *CorporationService) FindDirectorCharacter(ctx context.Context, corporationID int32) (*app.Character, error) {
	cc, err := s.cs.ListCharacters(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range cc {
		if c.EveCharacter == nil || c.EveCharacter.Corporation.ID != corporationID {
			continue
		}
		roles, err := s.cs.ListRoles(ctx, c.ID)
		if err != nil {
			return nil, err
		}
		if hasDirectorRole(roles) {
			return c, nil
		}
	}
	return nil, app.ErrNotFound
}

func (s *CorporationService) GetCorporationMessage(ctx context.Context, corporationID int32) (*app.CorporationMessage, error) {
	return s.st.GetCorporationMessage(ctx, corporationID)
}

func (s *CorporationService) UpdateCorporationMessage(ctx context.Context, arg app.UpdateCorporationMessageParams) (*app.CorporationMessage, error) {
	wrapErr := func(err error) error {
		return fmt.Errorf("UpdateCorporationMessage %+v: %w", arg, err)
	}
	arg.Message = strings.TrimSpace(arg.Message)
	if arg.Message == "" || arg.CorporationID == 0 || arg.CharacterID == 0 {
		return nil, wrapErr(app.ErrInvalid)
	}
	character, err := s.cs.GetCharacter(ctx, arg.CharacterID)
	if err != nil {
		return nil, wrapErr(err)
	}
	if character.EveCharacter == nil || character.EveCharacter.Corporation.ID != arg.CorporationID {
		return nil, wrapErr(app.ErrInvalid)
	}
	roles, err := s.cs.ListRoles(ctx, arg.CharacterID)
	if err != nil {
		return nil, wrapErr(err)
	}
	if !hasDirectorRole(roles) {
		return nil, wrapErr(app.ErrInvalid)
	}
	params := storage.UpdateCorporationMessageParams{
		CorporationID: arg.CorporationID,
		Message:       arg.Message,
		SourceURL:     arg.SourceURL,
		UpdatedBy:     optional.New(arg.CharacterID),
	}
	return s.st.UpdateCorporationMessage(ctx, params)
}

func hasDirectorRole(roles []app.CharacterRole) bool {
	for _, r := range roles {
		if r.Role == app.RoleDirector && r.Granted {
			return true
		}
	}
	return false
}
