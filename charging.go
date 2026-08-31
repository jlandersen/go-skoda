package skoda

import (
	"context"

	"github.com/jlandersen/go-skoda/internal/api"
)

const (
	ChargingStateReadyForCharging    = "READY_FOR_CHARGING"
	ChargingStateConnectCable        = "CONNECT_CABLE"
	ChargingStateConserving          = "CONSERVING"
	ChargingStateCharging            = "CHARGING"
	ChargingStateDischarging         = "DISCHARGING"
	ChargingStateChargingInterrupted = "CHARGING_INTERRUPTED"
	ChargeTypeAC                     = "AC"
	ChargeTypeDC                     = "DC"
	ChargeTypeOff                    = "OFF"
	ChargingTimerTypeOneOff          = "ONE_OFF"
	ChargingTimerTypeRecurring       = "RECURRING"
	Monday                           = "MONDAY"
	Tuesday                          = "TUESDAY"
	Wednesday                        = "WEDNESDAY"
	Thursday                         = "THURSDAY"
	Friday                           = "FRIDAY"
	Saturday                         = "SATURDAY"
	Sunday                           = "SUNDAY"
)

// ChargingBattery contains the current high-voltage battery state.
type ChargingBattery = api.BatteryStatus

// ChargingStatus contains the current charging status.
type ChargingStatus = api.ChargingStatus

// ChargingSettings contains the vehicle's charging configuration.
type ChargingSettings = api.ChargingSettings

// Charging contains charging and battery status for a vehicle.
type Charging = api.Charging

// ChargingProfiles contains all saved charging locations for a vehicle.
type ChargingProfiles = api.ChargingProfiles

// ChargingProfile contains a saved charging location and its schedule.
type ChargingProfile = api.ChargingProfile

// ChargingProfileSettings contains settings for a saved charging location.
type ChargingProfileSettings = api.ChargingProfileSettings

// MinBatteryStateOfCharge configures immediate charging below a threshold.
type MinBatteryStateOfCharge = api.MinBatteryStateOfCharge

// ChargingTime describes a preferred charging time window.
type ChargingTime = api.ChargingTime

// ChargingTimer describes a one-off or recurring charging timer.
type ChargingTimer = api.Timer

// ChargingTimerOneOffDay is the weekday of a one-off charging timer.
type ChargingTimerOneOffDay = api.TimerOneOffDay

// ChargingTimerRecurringDay is a weekday a recurring charging timer repeats on.
type ChargingTimerRecurringDay = api.TimerRecurringOn

// CurrentVehiclePositionProfile identifies the saved location containing the vehicle.
type CurrentVehiclePositionProfile = api.CurrentVehiclePositionProfile

// Charging retrieves only the charging section for a vehicle.
func (c *Client) Charging(ctx context.Context, vin string) (*Charging, error) {
	response, err := c.Vehicle(ctx, vin, IncludeCharging)
	if err != nil {
		return nil, err
	}
	if response.Vehicle.Charging == nil {
		return nil, &VehicleDataError{Include: IncludeCharging, Errors: vehicleErrors(response)}
	}
	return response.Vehicle.Charging, nil
}
