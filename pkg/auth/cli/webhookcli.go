package cli

import (
	"os"
	"time"

	"github.com/rancher/norman/pkg/auth"
	"github.com/urfave/cli"
)

type WebhookConfig struct {
	WebhookAuthentication bool
	WebhookKubeconfig     string
	WebhookURL            string
	CacheTTLSeconds       int
}

func (w *WebhookConfig) MustWebhookMiddleware() auth.Middleware {
	m, err := w.WebhookMiddleware()
	if err != nil {
		panic("failed to create webhook middleware: " + err.Error())
	}
	return m
}

func (w *WebhookConfig) WebhookMiddleware() (auth.Middleware, error) {
	if !w.WebhookAuthentication {
		return nil, nil
	}

	config := w.WebhookKubeconfig
	if config == "" && w.WebhookURL != "" {
		tempFile, err := auth.WebhookConfigForURL(w.WebhookURL)
		if err != nil {
			return nil, err
		}
		defer os.Remove(tempFile)
		config = tempFile
	}

	return auth.NewWebhookMiddleware(time.Duration(w.CacheTTLSeconds)*time.Second, config)
}

func Flags(config *WebhookConfig) []cli.Flag {
	return []cli.Flag{
		cli.BoolTFlag{
			Name:        "webhook-auth",
			EnvVar:      "WEBHOOK_AUTH",
			Destination: &config.WebhookAuthentication,
		},
		cli.StringFlag{
			Name:        "webhook-kubeconfig",
			EnvVar:      "WEBHOOK_KUBECONFIG",
			Destination: &config.WebhookKubeconfig,
		},
		cli.StringFlag{
			Name:        "webhook-url",
			EnvVar:      "WEBHOOK_URL",
			Destination: &config.WebhookURL,
		},
		cli.IntFlag{
			Name:        "webhook-cache-ttl",
			EnvVar:      "WEBHOOK_CACHE_TTL",
			Destination: &config.CacheTTLSeconds,
		},
	}
}
