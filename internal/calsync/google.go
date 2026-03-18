package calsync

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gcal "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/ivanlee1999/sesh/internal/config"
	"github.com/ivanlee1999/sesh/internal/db"
)

const googleProvider = "google"

// GoogleProvider syncs sessions to Google Calendar.
type GoogleProvider struct {
	cfg config.GoogleCalendarConfig
}

// NewGoogle creates a new Google Calendar provider.
func NewGoogle(cfg config.GoogleCalendarConfig) *GoogleProvider {
	return &GoogleProvider{cfg: cfg}
}

func (g *GoogleProvider) Name() string { return "Google Calendar" }

func (g *GoogleProvider) oauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     g.cfg.ClientID,
		ClientSecret: g.cfg.ClientSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/calendar.events"},
		Endpoint:     google.Endpoint,
	}
}

// Authenticate runs the OAuth2 flow and saves the token.
func (g *GoogleProvider) Authenticate() error {
	token, err := RunAuthFlow(g.oauthConfig())
	if err != nil {
		return err
	}
	return SaveToken(googleProvider, token)
}

// AuthenticateQuiet runs the OAuth2 flow without printing to stdout (for TUI use).
func (g *GoogleProvider) AuthenticateQuiet() error {
	token, err := RunAuthFlowQuiet(g.oauthConfig())
	if err != nil {
		return err
	}
	return SaveToken(googleProvider, token)
}

// CreateEvent creates a calendar event for the given session.
func (g *GoogleProvider) CreateEvent(session *db.SessionRecord) error {
	token, err := LoadToken(googleProvider)
	if err != nil {
		return fmt.Errorf("load google token: %w (run 'sesh calendar auth google' first)", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	oauthCfg := g.oauthConfig()
	tokenSource := oauthCfg.TokenSource(ctx, token)

	// Re-save token if it was refreshed
	newToken, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("refresh google token: %w", err)
	}
	if newToken.AccessToken != token.AccessToken {
		_ = SaveToken(googleProvider, newToken)
	}

	svc, err := gcal.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return fmt.Errorf("create google calendar service: %w", err)
	}

	startTime, err := time.Parse("2006-01-02T15:04:05", session.StartedAt)
	if err != nil {
		return fmt.Errorf("parse start time: %w", err)
	}
	endTime, err := time.Parse("2006-01-02T15:04:05", session.EndedAt)
	if err != nil {
		return fmt.Errorf("parse end time: %w", err)
	}

	event := &gcal.Event{
		Summary:     SessionSummary(session),
		Description: SessionDescription(session),
		Start: &gcal.EventDateTime{
			DateTime: startTime.Format(time.RFC3339),
		},
		End: &gcal.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
		},
	}

	calID := g.cfg.CalendarID
	if calID == "" {
		calID = "primary"
	}

	_, err = svc.Events.Insert(calID, event).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("insert google calendar event: %w", err)
	}

	return nil
}

// IsAuthenticated checks if a valid token exists.
func (g *GoogleProvider) IsAuthenticated() bool {
	return HasToken(googleProvider)
}
