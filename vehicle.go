package skoda

import (
	"context"
	"fmt"
)

// RenderType describes how the vehicle image was generated.
type RenderType string

const (
	RenderTypeReal RenderType = "REAL"
)

// ViewType identifies the type of composite render.
type ViewType string

const (
	ViewTypeUnmodifiedExteriorSide  ViewType = "UNMODIFIED_EXTERIOR_SIDE"
	ViewTypeUnmodifiedExteriorFront ViewType = "UNMODIFIED_EXTERIOR_FRONT"
	ViewTypeHome                    ViewType = "HOME"
	ViewTypeChargingLight           ViewType = "CHARGING_LIGHT"
	ViewTypeChargingDark            ViewType = "CHARGING_DARK"
	ViewTypePluggedInDark           ViewType = "PLUGGED_IN_DARK"
	ViewTypePluggedInLight          ViewType = "PLUGGED_IN_LIGHT"
)

// Render is a single vehicle image with a URL and viewpoint.
type Render struct {
	URL       string     `json:"url"`
	Type      RenderType `json:"type"`
	Order     int        `json:"order"`
	ViewPoint string     `json:"viewPoint"`
}

// CompositeRender is a set of layered renders for a specific view type.
type CompositeRender struct {
	ViewType ViewType `json:"viewType"`
	Layers   []Render `json:"layers"`
}

// CapabilityID identifies a vehicle capability.
type CapabilityID string

const (
	CapabilityAccess                      CapabilityID = "ACCESS"
	CapabilityAirConditioning             CapabilityID = "AIR_CONDITIONING"
	CapabilityAuxiliaryHeating            CapabilityID = "AUXILIARY_HEATING"
	CapabilityCharging                    CapabilityID = "CHARGING"
	CapabilityChargingMEB                 CapabilityID = "CHARGING_MEB"
	CapabilityChargingMQB                 CapabilityID = "CHARGING_MQB"
	CapabilityChargingProfiles            CapabilityID = "CHARGING_PROFILES"
	CapabilityDepartureTimers             CapabilityID = "DEPARTURE_TIMERS"
	CapabilityFuelStatus                  CapabilityID = "FUEL_STATUS"
	CapabilityHonkAndFlash                CapabilityID = "HONK_AND_FLASH"
	CapabilityParkingPosition             CapabilityID = "PARKING_POSITION"
	CapabilityState                       CapabilityID = "STATE"
	CapabilityTripStatistics              CapabilityID = "TRIP_STATISTICS"
	CapabilityVehicleHealthInspection     CapabilityID = "VEHICLE_HEALTH_INSPECTION"
	CapabilityVehicleHealthWarnings       CapabilityID = "VEHICLE_HEALTH_WARNINGS"
	CapabilityVehicleHealthWarningsWakeUp CapabilityID = "VEHICLE_HEALTH_WARNINGS_WITH_WAKE_UP"
	CapabilityVehicleWakeUp               CapabilityID = "VEHICLE_WAKE_UP"
	CapabilityVehicleWakeUpTrigger        CapabilityID = "VEHICLE_WAKE_UP_TRIGGER"
	CapabilityPredictiveWakeUp            CapabilityID = "PREDICTIVE_WAKE_UP"
	CapabilityWindowHeating               CapabilityID = "WINDOW_HEATING"
)

// Capability represents a single vehicle capability and its status.
type Capability struct {
	ID       CapabilityID `json:"id"`
	Statuses []string     `json:"statuses"`
}

// IsAvailable returns true if the capability has no error statuses,
// meaning it can currently be used.
func (cap Capability) IsAvailable() bool {
	return len(cap.Statuses) == 0
}

// Engine describes the vehicle's engine.
type Engine struct {
	Type             string  `json:"type"`
	PowerInKW        int     `json:"powerInKW"`
	CapacityInLiters float64 `json:"capacityInLiters,omitempty"`
}

// Battery describes the vehicle's traction battery.
type Battery struct {
	CapacityInKWh int `json:"capacityInKWh"`
}

// Gearbox describes the vehicle's gearbox.
type Gearbox struct {
	Type string `json:"type"`
}

// Specification describes the physical features of the vehicle.
type Specification struct {
	Title                string   `json:"title"`
	Model                string   `json:"model"`
	ModelYear            string   `json:"modelYear"`
	Body                 string   `json:"body"`
	SystemCode           string   `json:"systemCode"`
	SystemModelID        string   `json:"systemModelId"`
	Engine               Engine   `json:"engine"`
	Battery              *Battery `json:"battery,omitempty"`
	Gearbox              Gearbox  `json:"gearbox"`
	TrimLevel            string   `json:"trimLevel,omitempty"`
	ManufacturingDate    string   `json:"manufacturingDate"`
	MaxChargingPowerInKW int      `json:"maxChargingPowerInKW,omitempty"`
}

// Capabilities holds the list of capabilities for a vehicle.
type Capabilities struct {
	Capabilities []Capability `json:"capabilities"`
}

// VehicleInfo contains detailed information about a specific vehicle.
type VehicleInfo struct {
	VIN                 string            `json:"vin"`
	Name                string            `json:"name"`
	State               string            `json:"state"`
	Specification       Specification     `json:"specification"`
	Capabilities        Capabilities      `json:"capabilities"`
	Renders             []Render          `json:"renders"`
	CompositeRenders    []CompositeRender `json:"compositeRenders"`
	DevicePlatform      string            `json:"devicePlatform"`
	WorkshopModeEnabled bool              `json:"workshopModeEnabled"`
	SoftwareVersion     string            `json:"softwareVersion,omitempty"`
	LicensePlate        string            `json:"licensePlate,omitempty"`
}

// HasCapability checks whether the vehicle has a given capability,
// regardless of whether it's currently available.
func (v *VehicleInfo) HasCapability(id CapabilityID) bool {
	for _, cap := range v.Capabilities.Capabilities {
		if cap.ID == id {
			return true
		}
	}
	return false
}

// IsCapabilityAvailable checks whether the vehicle has the capability
// and it is currently available (no error statuses).
func (v *VehicleInfo) IsCapabilityAvailable(id CapabilityID) bool {
	for _, cap := range v.Capabilities.Capabilities {
		if cap.ID == id {
			return cap.IsAvailable()
		}
	}
	return false
}

// RenderByViewPoint returns the first render matching the given viewpoint, or nil.
func (v *VehicleInfo) RenderByViewPoint(viewPoint string) *Render {
	for _, r := range v.Renders {
		if r.ViewPoint == viewPoint {
			return &r
		}
	}
	return nil
}

// CompositeRenderByViewType returns the first composite render matching the given view type, or nil.
func (v *VehicleInfo) CompositeRenderByViewType(viewType ViewType) *CompositeRender {
	for _, cr := range v.CompositeRenders {
		if cr.ViewType == viewType {
			return &cr
		}
	}
	return nil
}

// VehicleInfo retrieves detailed information for a vehicle by VIN,
// including capabilities and specification.
func (c *Client) VehicleInfo(ctx context.Context, vin string) (*VehicleInfo, error) {
	var res VehicleInfo
	url := fmt.Sprintf("%s/v2/garage/vehicles/%s?%s", c.apiURL, vin, AllGenerations)
	if err := c.doGet(ctx, url, &res); err != nil {
		return nil, fmt.Errorf("vehicle info: %w", err)
	}
	return &res, nil
}
