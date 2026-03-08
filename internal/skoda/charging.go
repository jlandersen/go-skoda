package skoda

import (
	"context"
	"fmt"
	"time"
)

// ChargingState represents the current state of charging.
type ChargingState string

const (
	ChargingStateReadyForCharging    ChargingState = "READY_FOR_CHARGING"
	ChargingStateConnectCable        ChargingState = "CONNECT_CABLE"
	ChargingStateConserving          ChargingState = "CONSERVING"
	ChargingStateCharging            ChargingState = "CHARGING"
	ChargingStateChargingInterrupted ChargingState = "CHARGING_INTERRUPTED"
	ChargingStateError               ChargingState = "ERROR"
)

// ChargeType represents the type of charging connection.
type ChargeType string

const (
	ChargeTypeAC  ChargeType = "AC"
	ChargeTypeDC  ChargeType = "DC"
	ChargeTypeOff ChargeType = "OFF"
)

// MaxChargeCurrent represents the maximum charge current setting.
type MaxChargeCurrent string

const (
	MaxChargeCurrentMaximum MaxChargeCurrent = "MAXIMUM"
	MaxChargeCurrentReduced MaxChargeCurrent = "REDUCED"
)

// ChargingBattery contains the current battery state.
type ChargingBattery struct {
	StateOfChargeInPercent         *int   `json:"stateOfChargeInPercent"`
	RemainingCruisingRangeInMeters *int64 `json:"remainingCruisingRangeInMeters"`
}

// ChargingStatus contains the current charging status.
type ChargingStatus struct {
	Battery                              ChargingBattery `json:"battery"`
	State                                ChargingState   `json:"state,omitempty"`
	ChargePowerInKW                      *float64        `json:"chargePowerInKw,omitempty"`
	ChargingRateInKilometersPerHour      *float64        `json:"chargingRateInKilometersPerHour,omitempty"`
	ChargeType                           ChargeType      `json:"chargeType,omitempty"`
	RemainingTimeToFullyChargedInMinutes *int64          `json:"remainingTimeToFullyChargedInMinutes,omitempty"`
}

// ChargingSettings contains the charging configuration.
type ChargingSettings struct {
	AutoUnlockPlugWhenCharged    string `json:"autoUnlockPlugWhenCharged,omitempty"`
	MaxChargeCurrentAC           string `json:"maxChargeCurrentAc,omitempty"`
	TargetStateOfChargeInPercent *int   `json:"targetStateOfChargeInPercent,omitempty"`
	PreferredChargeMode          string `json:"preferredChargeMode,omitempty"`
	ChargingCareMode             string `json:"chargingCareMode,omitempty"`
	BatterySupport               string `json:"batterySupport,omitempty"`
}

// Charging contains the full charging information for a vehicle.
type Charging struct {
	IsVehicleInSavedLocation bool             `json:"isVehicleInSavedLocation"`
	Status                   *ChargingStatus  `json:"status,omitempty"`
	Settings                 ChargingSettings `json:"settings"`
	CarCapturedTimestamp     *time.Time       `json:"carCapturedTimestamp,omitempty"`
}

// Charging retrieves the current charging state for a vehicle.
// This is a safe read-only endpoint that does not wake the car.
func (c *Client) Charging(ctx context.Context, vin string) (*Charging, error) {
	var res Charging
	url := fmt.Sprintf("%s/v1/charging/%s", c.apiURL, vin)
	if err := c.doGet(ctx, url, &res); err != nil {
		return nil, fmt.Errorf("charging: %w", err)
	}
	return &res, nil
}
