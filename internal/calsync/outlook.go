package calsync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"github.com/ivanlee1999/sesh/internal/config"
	"github.com/ivanlee1999/sesh/internal/db"
)

const outlookProvider = "outlook"

// OutlookProvider syncs sessions to Microsoft Outlook/365 Calendar.
type OutlookProvider struct {
	cfg config.OutlookCalendarConfig
}

// NewOutlook creates a new Outlook Calendar provider.
func NewOutlook(cfg config.OutlookCalendarConfig) *OutlookProvider {
	return &OutlookProvider{cfg: cfg}
}

func (o *OutlookProvider) Name() string { return "Outlook Calendar" }

func (o *OutlookProvider) oauthConfig() *oauth2.Config {
	tenant := o.cfg.TenantID
	if tenant == "" {
		tenant = "common"
	}
	return &oauth2.Config{
		ClientID: o.cfg.ClientID,
		Scopes:   []string{"Calendars.ReadWrite", "offline_access"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize", tenant),
			TokenURL: fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenant),
		},
	}
}

// Authenticate runs the OAuth2 flow and saves the token.
func (o *OutlookProvider) Authenticate() error {
	token, err := RunAuthFlow(o.oauthConfig())
	if err != nil {
		return err
	}
	return SaveToken(outlookProvider, token)
}

// AuthenticateQuiet runs the OAuth2 flow without printing to stdout (for TUI use).
func (o *OutlookProvider) AuthenticateQuiet() error {
	token, err := RunAuthFlowQuiet(o.oauthConfig())
	if err != nil {
		return err
	}
	return SaveToken(outlookProvider, token)
}

// outlookEvent represents a Microsoft Graph calendar event.
type outlookEvent struct {
	Subject string            `json:"subject"`
	Body    outlookBody       `json:"body"`
	Start   outlookDateTime   `json:"start"`
	End     outlookDateTime   `json:"end"`
}

type outlookBody struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

type outlookDateTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

// CreateEvent creates a calendar event for the given session.
func (o *OutlookProvider) CreateEvent(session *db.SessionRecord) error {
	token, err := LoadToken(outlookProvider)
	if err != nil {
		return fmt.Errorf("load outlook token: %w (run 'sesh calendar auth outlook' first)", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	oauthCfg := o.oauthConfig()
	tokenSource := oauthCfg.TokenSource(ctx, token)

	// Re-save token if it was refreshed
	newToken, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("refresh outlook token: %w", err)
	}
	if newToken.AccessToken != token.AccessToken {
		_ = SaveToken(outlookProvider, newToken)
	}

	client := oauth2.NewClient(ctx, tokenSource)

	event := outlookEvent{
		Subject: SessionSummary(session),
		Body: outlookBody{
			ContentType: "text",
			Content:     SessionDescription(session),
		},
		Start: outlookDateTime{
			DateTime: session.StartedAt,
			TimeZone: "UTC",
		},
		End: outlookDateTime{
			DateTime: session.EndedAt,
			TimeZone: "UTC",
		},
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal outlook event: %w", err)
	}

	url := "https://graph.microsoft.com/v1.0/me/events"
	if o.cfg.CalendarID != "" {
		url = fmt.Sprintf("https://graph.microsoft.com/v1.0/me/calendars/%s/events", o.cfg.CalendarID)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create outlook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("outlook calendar API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("outlook calendar API error %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// IsAuthenticated checks if a valid token exists.
func (o *OutlookProvider) IsAuthenticated() bool {
	return HasToken(outlookProvider)
}
