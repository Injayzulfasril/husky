// SPDX-License-Identifier: ice License 1.0

package analytics

import (
	"context"
	_ "embed"
	"io"

	"github.com/ice-blockchain/wintr/analytics/tracking"
	messagebroker "github.com/ice-blockchain/wintr/connectors/message_broker"
	storage "github.com/ice-blockchain/wintr/connectors/storage/v2"
	"github.com/ice-blockchain/wintr/time"
)

// Public API.

type (
	TrackActionCommand struct {
		Action *tracking.Action `json:"action,omitempty"`
		ID     string           `json:"id,omitempty" example:"edfd8c02-75e0-4687-9ac2-1ce4723865c4"`
		UserID string           `json:"userId,omitempty" example:"edfd8c02-75e0-4687-9ac2-1ce4723865c4"`
	}
	SetUserAttributesCommand struct {
		Attributes map[string]any `json:"attributes,omitempty"`
		UserID     string         `json:"userId,omitempty" example:"edfd8c02-75e0-4687-9ac2-1ce4723865c4"`
	}
	Processor interface {
		io.Closer
		CheckHealth(context.Context) error
	}
)

// Private API.

const (
	applicationYamlKey = "analytics"
)

// .
var (
	//go:embed DDL.sql
	ddl string
)

type (
	trackedAction struct {
		SentAt *time.Time `json:"sentAt,omitempty"`
		ID     string     `json:"id,omitempty"`
	}
	setUserAttributesSource struct {
		*processor
	}
	trackActionSource struct {
		*processor
	}
	processor struct {
		*repository
	}
	repository struct {
		cfg            *config
		shutdown       func() error
		db             *storage.DB
		trackingClient tracking.Client
	}
	config struct {
		messagebroker.Config `mapstructure:",squash"` //nolint:tagliatelle // Nope.
	}
)
