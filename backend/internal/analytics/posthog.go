package analytics

import (
	"os"

	"github.com/posthog/posthog-go"
	"margin.at/internal/logger"
)

type Client struct {
	ph      posthog.Client
	enabled bool
}

func New() *Client {
	token := os.Getenv("POSTHOG_PROJECT_TOKEN")
	if token == "" {
		logger.Info("PostHog analytics disabled (POSTHOG_PROJECT_TOKEN not set)")
		return &Client{}
	}

	host := os.Getenv("POSTHOG_HOST")
	if host == "" {
		host = "https://us.i.posthog.com"
	}

	ph, err := posthog.NewWithConfig(token, posthog.Config{
		Endpoint: host,
	})
	if err != nil {
		logger.Error("Failed to initialise PostHog client: %v", err)
		return &Client{}
	}

	logger.Info("PostHog analytics enabled (host: %s)", host)
	return &Client{ph: ph, enabled: true}
}

func (c *Client) Capture(distinctID, event string, properties map[string]interface{}) {
	if !c.enabled || c.ph == nil {
		return
	}
	props := posthog.NewProperties()
	for k, v := range properties {
		props.Set(k, v)
	}

	_ = c.ph.Enqueue(posthog.Capture{
		DistinctId: distinctID,
		Event:      event,
		Properties: props,
	})
}

func (c *Client) Identify(distinctID string, properties map[string]interface{}) {
	if !c.enabled || c.ph == nil {
		return
	}
	traits := posthog.NewProperties()
	for k, v := range properties {
		traits.Set(k, v)
	}
	_ = c.ph.Enqueue(posthog.Identify{
		DistinctId: distinctID,
		Properties: traits,
	})
}

func (c *Client) Close() {
	if c.enabled && c.ph != nil {
		c.ph.Close()
	}
}
