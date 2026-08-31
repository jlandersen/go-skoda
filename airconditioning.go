package skoda

import (
	"context"

	"github.com/jlandersen/go-skoda/internal/api"
)

const (
	AirConditioningStateCooling          = "COOLING"
	AirConditioningStateHeating          = "HEATING"
	AirConditioningStateHeatingAuxiliary = "HEATING_AUXILIARY"
	AirConditioningStateOff              = "OFF"
	AirConditioningStateVentilation      = "VENTILATION"
	AirConditioningStateCompleted        = "COMPLETED"
	AirConditioningStateUnknown          = "UNKNOWN"
	AirConditioningStateUnsupported      = "UNSUPPORTED"
	// AuxiliaryHeatingStatePreheating and ActiveVentilationStatePreheating are
	// reported by the auxiliary heater and active ventilation only.
	AuxiliaryHeatingStatePreheating      = "PREHEATING"
	ActiveVentilationStatePreheating     = "PREHEATING"
	AuxiliaryHeatingStartModeHeating     = "HEATING"
	AuxiliaryHeatingStartModeVentilation = "VENTILATION"
)

// TargetTemperature contains a cabin temperature and its unit.
type TargetTemperature = api.TargetTemperature

// WindowHeating contains the electric window-heating state.
type WindowHeating = api.WindowHeating

// AirConditioning contains the air-conditioning state and settings.
type AirConditioning = api.AirConditioning

// AuxiliaryHeating contains the auxiliary heater state and settings.
type AuxiliaryHeating = api.AuxiliaryHeating

// ActiveVentilation contains the active ventilation state and duration.
type ActiveVentilation = api.ActiveVentilation

// AirConditioning retrieves only the air-conditioning section for a vehicle.
func (c *Client) AirConditioning(ctx context.Context, vin string) (*AirConditioning, error) {
	response, err := c.Vehicle(ctx, vin, IncludeAirConditioning)
	if err != nil {
		return nil, err
	}
	if response.Vehicle.AirConditioning == nil {
		return nil, &VehicleDataError{Include: IncludeAirConditioning, Errors: vehicleErrors(response)}
	}
	return response.Vehicle.AirConditioning, nil
}
