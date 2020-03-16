package provider

import (
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models"
)

// Factory is responsible to create a new driver based on the metadata.
type Factory func(meta *models.Metadata) (Driver, error)

// DragonflyFactory creates dragonfly driver
func DragonflyFactory(meta *models.Metadata) (Driver, error) {
	return &DragonflyDriver{meta}, nil
}

// KrakenFactory creates kraken driver
func KrakenFactory(meta *models.Metadata) (Driver, error) {
	return &KrakenDriver{meta}, nil
}
