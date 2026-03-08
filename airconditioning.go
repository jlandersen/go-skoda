package skoda

import (
	"context"
	"fmt"
	"time"
)

// AirConditioningState represents the state of the air conditioning.
type AirConditioningState string

const (
	AirConditioningStateCooling          AirConditioningState = "COOLING"
	AirConditioningStateHeating          AirConditioningState = "HEATING"
	AirConditioningStateHeatingAuxiliary AirConditioningState = "HEATING_AUXILIARY"
	AirConditioningStateOff              AirConditioningState = "OFF"
	AirConditioningStateOn               AirConditioningState = "ON"
	AirConditioningStateVentilation      AirConditioningState = "VENTILATION"
)

// ConnectionState represents whether the charger is connected.
type ConnectionState string

const (
	ConnectionStateConnected    ConnectionState = "CONNECTED"
	ConnectionStateDisconnected ConnectionState = "DISCONNECTED"
)

// ChargerLockedState represents the lock state of the charger.
type ChargerLockedState string

const (
	ChargerLockedStateLocked   ChargerLockedState = "LOCKED"
	ChargerLockedStateUnlocked ChargerLockedState = "UNLOCKED"
	ChargerLockedStateInvalid  ChargerLockedState = "INVALID"
)

// WindowHeatingState describes the state of front/rear window heating.
type WindowHeatingState struct {
	Front string `json:"front"`
	Rear  string `json:"rear"`
}

// TargetTemperature describes the target temperature setting.
type TargetTemperature struct {
	TemperatureValue float64 `json:"temperatureValue"`
	UnitInCar        string  `json:"unitInCar"`
}

// SeatHeating describes seat heating state.
type SeatHeating struct {
	FrontLeft  *bool `json:"frontLeft,omitempty"`
	FrontRight *bool `json:"frontRight,omitempty"`
}

// AirConditioning contains the air conditioning state for a vehicle.
type AirConditioning struct {
	State                  AirConditioningState `json:"state"`
	ChargerConnectionState ConnectionState      `json:"chargerConnectionState,omitempty"`
	ChargerLockState       ChargerLockedState   `json:"chargerLockState,omitempty"`
	WindowHeatingState     *WindowHeatingState  `json:"windowHeatingState,omitempty"`
	TargetTemperature      *TargetTemperature   `json:"targetTemperature,omitempty"`
	SeatHeatingActivated   *SeatHeating         `json:"seatHeatingActivated,omitempty"`
	HeaterSource           string               `json:"heaterSource,omitempty"`
	WindowHeatingEnabled   *bool                `json:"windowHeatingEnabled,omitempty"`
	CarCapturedTimestamp   *time.Time           `json:"carCapturedTimestamp,omitempty"`
	SteeringWheelPosition  string               `json:"steeringWheelPosition,omitempty"`
}

// AirConditioning retrieves the current air conditioning state for a vehicle.
// This is a safe read-only endpoint that does not wake the car.
func (c *Client) AirConditioning(ctx context.Context, vin string) (*AirConditioning, error) {
	var res AirConditioning
	url := fmt.Sprintf("%s/v2/air-conditioning/%s", c.apiURL, vin)
	if err := c.doGet(ctx, url, &res); err != nil {
		return nil, fmt.Errorf("air conditioning: %w", err)
	}
	return &res, nil
}
