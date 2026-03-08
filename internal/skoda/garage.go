package skoda

import (
	"context"
	"fmt"
)

// GarageEntry represents a single vehicle in the user's garage.
type GarageEntry struct {
	VIN              string            `json:"vin"`
	Name             string            `json:"name"`
	State            string            `json:"state"`
	Title            string            `json:"title"`
	DevicePlatform   string            `json:"devicePlatform"`
	SystemModelID    string            `json:"systemModelId"`
	Renders          []Render          `json:"renders"`
	CompositeRenders []CompositeRender `json:"compositeRenders"`
}

type garageResponse struct {
	Vehicles []GarageEntry `json:"vehicles"`
}

// Garage returns the list of vehicles associated with the authenticated user.
func (c *Client) Garage(ctx context.Context) ([]GarageEntry, error) {
	var res garageResponse
	url := fmt.Sprintf("%s/v2/garage?%s", c.apiURL, AllGenerations)
	if err := c.doGet(ctx, url, &res); err != nil {
		return nil, fmt.Errorf("garage: %w", err)
	}
	return res.Vehicles, nil
}
