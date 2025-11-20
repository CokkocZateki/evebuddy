package ui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	kxwidget "github.com/ErikKalkoken/fyne-kx/widget"

	"github.com/ErikKalkoken/evebuddy/internal/app"
	"github.com/ErikKalkoken/evebuddy/internal/app/icons"
	"github.com/ErikKalkoken/evebuddy/internal/optional"
	iwidget "github.com/ErikKalkoken/evebuddy/internal/widget"
)

type corporationMotd struct {
	widget.BaseWidget

	corporation *app.Corporation
	director    *app.Character
	messageData *app.CorporationMessage

	editButton *widget.Button
	info       *widget.Label
	logo       *kxwidget.TappableImage
	message    *widget.Label
	source     *widget.Hyperlink
	title      *widget.Label
	updated    *widget.Label
	u          *baseUI
}

func newCorporationMotd(u *baseUI) *corporationMotd {
	message := widget.NewLabel("")
	message.Wrapping = fyne.TextWrapWord
	logo := kxwidget.NewTappableImage(icons.BlankSvg, nil)
	logo.SetMinSize(fyne.NewSquareSize(128))
	logo.SetFillMode(canvas.ImageFillContain)
	a := &corporationMotd{
		editButton: widget.NewButtonWithIcon("Add message", theme.ContentAddIcon(), nil),
		info:       widget.NewLabel("Directors can share a message of the day for their members."),
		logo:       logo,
		message:    message,
		source:     widget.NewHyperlink("", nil),
		title:      widget.NewLabel("Message of the day"),
		updated:    widget.NewLabel(""),
		u:          u,
	}
	a.updated.TextStyle = fyne.TextStyle{Italic: true}
	a.ExtendBaseWidget(a)
	a.editButton.OnTapped = a.openEditor
	a.u.currentCorporationExchanged.AddListener(func(_ context.Context, c *app.Corporation) {
		a.corporation = c
	})
	return a
}

func (a *corporationMotd) CreateRenderer() fyne.WidgetRenderer {
	header := container.NewBorder(nil, a.info, nil, nil, container.NewHBox(
		container.NewPadded(a.logo),
		container.NewVBox(a.title, a.updated, a.source),
	))
	content := container.NewVScroll(a.message)
	content.SetMinSize(fyne.NewSize(0, 300))
	c := container.NewBorder(
		header,
		container.NewHBox(layout.NewSpacer(), a.editButton),
		nil,
		nil,
		content,
	)
	return widget.NewSimpleRenderer(c)
}

func (a *corporationMotd) update() {
	corp := a.u.currentCorporation()
	if corp == nil {
		fyne.Do(func() {
			a.corporation = nil
			a.messageData = nil
			a.director = nil
			a.applyView()
		})
		return
	}
	ctx := context.Background()
	msg, msgErr := a.u.rs.GetCorporationMessage(ctx, corp.ID)
	if msgErr != nil && !errors.Is(msgErr, app.ErrNotFound) {
		slog.Error("Failed to load corporation message", "corporationID", corp.ID, "err", msgErr)
	}
	director, directorErr := a.u.rs.FindDirectorCharacter(ctx, corp.ID)
	if directorErr != nil && !errors.Is(directorErr, app.ErrNotFound) {
		slog.Error("Failed to look up director", "corporationID", corp.ID, "err", directorErr)
	}
	fyne.Do(func() {
		a.corporation = corp
		a.messageData = msg
		if directorErr == nil {
			a.director = director
		} else if director != nil {
			a.director = director
		} else {
			a.director = nil
		}
		a.applyView()
	})
}

func (a *corporationMotd) applyView() {
	if a.corporation == nil {
		a.logo.SetResource(icons.BlankSvg)
		a.title.SetText("Message of the day")
		a.message.SetText("Select a corporation to view its message of the day.")
		a.updated.SetText("")
		a.source.Hide()
		a.editButton.Hide()
		a.info.SetText("Directors can share a message of the day for their members.")
		return
	}
	a.title.SetText(fmt.Sprintf("%s message of the day", a.corporation.EveCorporation.Name))
	a.logo.OnTapped = func() {
		a.u.ShowInfoWindow(app.EveEntityCorporation, a.corporation.ID)
	}
	iwidget.RefreshTappableImageAsync(a.logo, func() (fyne.Resource, error) {
		return a.u.eis.CorporationLogo(a.corporation.ID, 256)
	})
	if a.messageData == nil {
		a.message.SetText("No message has been posted for this corporation yet.")
		a.updated.SetText("Directors can add a message to keep members informed.")
		a.source.Hide()
		a.editButton.SetIcon(theme.ContentAddIcon())
		a.editButton.SetText("Add message")
	} else {
		a.message.SetText(a.messageData.Message)
		updatedText := fmt.Sprintf("Updated %s", a.messageData.UpdatedAt.Format(app.DateTimeFormat))
		if a.messageData.UpdatedBy != nil {
			updatedText = fmt.Sprintf("%s by %s", updatedText, a.messageData.UpdatedBy.Name)
		}
		a.updated.SetText(updatedText)
		if a.messageData.SourceURL.IsEmpty() {
			a.source.Hide()
		} else {
			link := a.messageData.SourceURL.ValueOrZero()
			parsed, err := url.Parse(link)
			if err == nil {
				a.source.SetURL(parsed)
				a.source.SetText(link)
				a.source.Show()
			} else {
				a.source.Hide()
			}
		}
		a.editButton.SetIcon(theme.DocumentCreateIcon())
		a.editButton.SetText("Edit message")
	}
	if a.director != nil {
		a.editButton.Enable()
		a.editButton.Show()
		directorName := fmt.Sprintf("%d", a.director.ID)
		if a.director.EveCharacter != nil {
			directorName = a.director.EveCharacter.Name
		}
		a.info.SetText(fmt.Sprintf("Editing as director %s", directorName))
	} else {
		a.editButton.Disable()
		a.editButton.Show()
		a.info.SetText("Add a director character from this corporation to edit the message.")
	}
}

func (a *corporationMotd) openEditor() {
	if a.corporation == nil {
		return
	}
	if a.director == nil {
		a.u.ShowInformationDialog("Director required", "Only director characters from this corporation can edit the message of the day.", a.u.MainWindow())
		return
	}
	messageText := ""
	sourceText := ""
	if a.messageData != nil {
		messageText = a.messageData.Message
		sourceText = a.messageData.SourceURL.StringFunc("", func(v string) string { return v })
	}
	messageEntry := widget.NewMultiLineEntry()
	messageEntry.SetText(messageText)
	messageEntry.SetPlaceHolder("Enter the corporation's message of the day")
	messageEntry.SetMinRowsVisible(8)
	sourceEntry := widget.NewEntry()
	sourceEntry.SetPlaceHolder("Optional secret gist or other URL")
	sourceEntry.SetText(sourceText)
	loadButton := widget.NewButtonWithIcon("Load from URL", theme.DownloadIcon(), func() {
		urlText := strings.TrimSpace(sourceEntry.Text)
		if urlText == "" {
			a.u.ShowInformationDialog("Missing URL", "Enter a URL to import the message text from.", a.u.MainWindow())
			return
		}
		go func() {
			content, err := a.u.rs.FetchCorporationMessageFromURL(context.Background(), urlText)
			if err != nil {
				fyne.Do(func() {
					a.u.NewErrorDialog("Failed to fetch message", err, a.u.MainWindow()).Show()
				})
				return
			}
			fyne.Do(func() {
				messageEntry.SetText(content)
			})
		}()
	})
	sourceRow := container.NewBorder(nil, nil, nil, loadButton, sourceEntry)
	form := dialog.NewForm(
		"Update corporation message",
		"Save",
		"Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Message", container.NewVScroll(messageEntry)),
			widget.NewFormItem("Source URL", sourceRow),
		},
		func(ok bool) {
			if !ok {
				return
			}
			content := strings.TrimSpace(messageEntry.Text)
			if content == "" {
				a.u.ShowInformationDialog("Missing message", "Please enter the message of the day.", a.u.MainWindow())
				return
			}
			source := strings.TrimSpace(sourceEntry.Text)
			var sourceURL optional.Optional[string]
			if source != "" {
				sourceURL = optional.New(source)
			}
			go func() {
				msg, err := a.u.rs.UpdateCorporationMessage(context.Background(), app.UpdateCorporationMessageParams{
					CorporationID: a.corporation.ID,
					CharacterID:   a.director.ID,
					Message:       content,
					SourceURL:     sourceURL,
				})
				if err != nil {
					fyne.Do(func() {
						a.u.NewErrorDialog("Failed to save message", err, a.u.MainWindow()).Show()
					})
					return
				}
				fyne.Do(func() {
					a.messageData = msg
					a.applyView()
					a.u.snackbar.Show("Message of the day updated")
				})
			}()
		},
		a.u.MainWindow(),
	)
	form.Resize(fyne.NewSize(520, 520))
	a.u.ModifyShortcutsForDialog(form, a.u.MainWindow())
	form.Show()
}
