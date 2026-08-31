package skoda

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/jlandersen/go-skoda/internal/api"
)

// Include identifies a part of the vehicle response.
type Include = api.GetVehicleParamsInclude

const (
	IncludeInfo              = api.GetVehicleParamsIncludeInfo
	IncludeStatus            = api.GetVehicleParamsIncludeStatus
	IncludeFuelStatus        = api.GetVehicleParamsIncludeFuelStatus
	IncludeOdometer          = api.GetVehicleParamsIncludeOdometer
	IncludeParkingPosition   = api.GetVehicleParamsIncludeParkingPosition
	IncludeAirConditioning   = api.GetVehicleParamsIncludeAirConditioning
	IncludeAuxiliaryHeating  = api.GetVehicleParamsIncludeAuxiliaryHeating
	IncludeActiveVentilation = api.GetVehicleParamsIncludeActiveVentilation
	IncludeCharging          = api.GetVehicleParamsIncludeCharging
	IncludeChargingProfiles  = api.GetVehicleParamsIncludeChargingProfiles
)

// Vehicle contains the vehicle and the state sections requested from the API.
type Vehicle = api.Vehicle

// VehicleStatus contains the aggregated and detailed state of doors, windows, and lights.
type VehicleStatus = api.VehicleStatus

// OverallVehicleStatus contains aggregated door, window, lock, and light states.
type OverallVehicleStatus = api.OverallVehicleStatusDto

// VehicleStatusDetail contains individual body-opening states.
type VehicleStatusDetail = api.VehicleStatusDetailDto

// EngineRange contains fuel, charge, and range values for one engine.
type EngineRange = api.EngineRange

// FuelStatus contains fuel and range data for a combustion or hybrid vehicle.
type FuelStatus = api.FuelStatus

// Odometer contains the vehicle mileage.
type Odometer = api.Odometer

// GPSCoordinates contains latitude and longitude in decimal degrees.
type GPSCoordinates = api.ParkingPositionGpsCoordinates

// ParkingPosition contains the vehicle's last known parking position.
type ParkingPosition = api.ParkingPosition

// VehicleError describes a response section that could not be returned.
type VehicleError = api.VehicleError

const (
	VehicleErrorRenderUnavailable            = "RENDER_UNAVAILABLE"
	VehicleErrorStatusUnsupported            = "VEHICLE_STATUS_UNSUPPORTED"
	VehicleErrorStatusDisabled               = "VEHICLE_STATUS_DISABLED"
	VehicleErrorStatusUnavailable            = "VEHICLE_STATUS_UNAVAILABLE"
	VehicleErrorFuelStatusUnsupported        = "FUEL_STATUS_UNSUPPORTED"
	VehicleErrorFuelStatusDisabled           = "FUEL_STATUS_DISABLED"
	VehicleErrorFuelStatusUnavailable        = "FUEL_STATUS_UNAVAILABLE"
	VehicleErrorOdometerUnsupported          = "ODOMETER_UNSUPPORTED"
	VehicleErrorOdometerDisabled             = "ODOMETER_DISABLED"
	VehicleErrorOdometerUnavailable          = "ODOMETER_UNAVAILABLE"
	VehicleErrorParkingPositionUnsupported   = "PARKING_POSITION_UNSUPPORTED"
	VehicleErrorParkingPositionDisabled      = "PARKING_POSITION_DISABLED"
	VehicleErrorParkingPositionUnavailable   = "PARKING_POSITION_UNAVAILABLE"
	VehicleErrorAirConditioningUnsupported   = "AIR_CONDITIONING_UNSUPPORTED"
	VehicleErrorAirConditioningDisabled      = "AIR_CONDITIONING_DISABLED"
	VehicleErrorAirConditioningUnavailable   = "AIR_CONDITIONING_UNAVAILABLE"
	VehicleErrorAuxiliaryHeatingUnsupported  = "AUXILIARY_HEATING_UNSUPPORTED"
	VehicleErrorAuxiliaryHeatingDisabled     = "AUXILIARY_HEATING_DISABLED"
	VehicleErrorAuxiliaryHeatingUnavailable  = "AUXILIARY_HEATING_UNAVAILABLE"
	VehicleErrorActiveVentilationUnsupported = "ACTIVE_VENTILATION_UNSUPPORTED"
	VehicleErrorActiveVentilationDisabled    = "ACTIVE_VENTILATION_DISABLED"
	VehicleErrorActiveVentilationUnavailable = "ACTIVE_VENTILATION_UNAVAILABLE"
	VehicleErrorChargingUnsupported          = "CHARGING_UNSUPPORTED"
	VehicleErrorChargingDisabled             = "CHARGING_DISABLED"
	VehicleErrorChargingUnavailable          = "CHARGING_UNAVAILABLE"
	VehicleErrorChargingProfilesUnsupported  = "CHARGING_PROFILES_UNSUPPORTED"
	VehicleErrorChargingProfilesDisabled     = "CHARGING_PROFILES_DISABLED"
	VehicleErrorChargingProfilesUnavailable  = "CHARGING_PROFILES_UNAVAILABLE"
)

// VehicleData contains the vehicle and the partial-data errors returned with it.
type VehicleData = api.VehicleResponse

// VehicleResponse contains generated vehicle data, partial-data errors, and response metadata.
type VehicleResponse struct {
	VehicleData
	Metadata ResponseMetadata `json:"-"`
}

// VehicleDataError indicates that a requested response section was omitted.
type VehicleDataError struct {
	Include Include
	Errors  []VehicleError
}

func (e *VehicleDataError) Error() string {
	if len(e.Errors) == 0 {
		return fmt.Sprintf("skoda public API: %s data was omitted", e.Include)
	}
	return fmt.Sprintf("skoda public API: %s data was omitted: %s", e.Include, e.Errors[0].Type)
}

// Vehicle returns vehicle information and selected state sections for a VIN.
// Omitting includes requests every section supported by the vehicle.
func (c *Client) Vehicle(ctx context.Context, vin string, includes ...Include) (*VehicleResponse, error) {
	if len(vin) != 17 {
		return nil, fmt.Errorf("VIN must contain exactly 17 characters")
	}
	for _, include := range includes {
		if !include.Valid() {
			return nil, fmt.Errorf("invalid vehicle include %q", include)
		}
	}

	requestURL := fmt.Sprintf("%s/api/v1/vehicles/%s", c.baseURL, url.PathEscape(vin))
	if len(includes) > 0 {
		includeValues := make([]string, len(includes))
		for i, include := range includes {
			includeValues[i] = string(include)
		}
		query := url.Values{"include": {strings.Join(includeValues, ",")}}
		requestURL += "?" + query.Encode()
	}

	var body api.VehicleResponse
	metadata, err := c.doGet(ctx, requestURL, &body)
	if err != nil {
		return nil, fmt.Errorf("getting vehicle: %w", err)
	}
	return &VehicleResponse{VehicleData: body, Metadata: metadata}, nil
}

func vehicleErrors(response *VehicleResponse) []VehicleError {
	if response.Errors == nil {
		return nil
	}
	return *response.Errors
}
