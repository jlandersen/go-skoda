package skoda

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

const testVIN = "TMBJB9NY6RF123456"

func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		handler(w, r)
	}))
	client, err := NewClient("test-api-key", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	if err != nil {
		server.Close()
		t.Fatalf("NewClient() error = %v", err)
	}
	return client, server
}

func TestVehicle(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/v1/vehicles/"+testVIN {
			t.Errorf("path = %s, want /api/v1/vehicles/%s", r.URL.Path, testVIN)
		}
		if got := r.URL.Query().Get("include"); got != "info,status,fuelStatus,odometer,parkingPosition,airConditioning,auxiliaryHeating,activeVentilation,charging,chargingProfiles" {
			t.Errorf("include = %q", got)
		}
		if got := r.Header.Get("X-API-Key"); got != "test-api-key" {
			t.Errorf("X-API-Key = %q", got)
		}
		if got := r.Header.Get("Authorization"); got != "" {
			t.Errorf("Authorization = %q, want empty", got)
		}

		w.Header().Set("X-API-Key-Expires-At", "2026-09-30T00:00:00Z")
		w.Header().Set("RateLimit-Limit", "20")
		w.Header().Set("RateLimit-Remaining", "19")
		w.Header().Set("RateLimit-Reset", "3600")
		_, err := w.Write([]byte(`{
			"vehicle": {
				"vin": "TMBJB9NY6RF123456",
				"name": "My Enyaq",
				"licensePlate": "AB 12345",
				"status": {
					"overall": {"doorsLocked":"YES","locked":"YES","doors":"CLOSED","windows":"CLOSED","lights":"OFF","reliableLockStatus":"LOCKED"},
					"detail": {"sunroof":"CLOSED","trunk":"CLOSED","bonnet":"CLOSED"},
					"carCapturedTimestamp": "2026-08-31T12:00:00Z"
				},
				"charging": {
					"isVehicleInSavedLocation": true,
					"status": {
						"state": "CHARGING",
						"chargeType": "AC",
						"chargePowerInKw": 11,
						"remainingTimeToFullyChargedInMinutes": 90,
						"fullyChargedAt": "2026-08-31T13:30:00Z",
						"battery": {"stateOfChargeInPercent":75,"remainingCruisingRangeInMeters":310000}
					},
					"settings": {"targetStateOfChargeInPercent":80,"maxChargeCurrentAc":"MAXIMUM","maxChargeCurrentAcAmpere":32}
				},
				"airConditioning": {
					"state": "OFF",
					"targetTemperature": {"value":21.5,"unit":"CELSIUS"},
					"airConditioningWithoutExternalPower": true,
					"windowHeating": {"enabled":false,"front":"OFF","rear":"OFF"}
				},
				"fuelStatus": {
					"carType":"HYBRID",
					"totalRangeInKm":520,
					"primaryEngineRange":{"engineType":"GASOLINE","currentFuelLevelInPercent":87,"remainingRangeInKm":350}
				},
				"odometer":{"mileageInKm":12753},
				"parkingPosition":{"state":"PARKED","gpsCoordinates":{"latitude":37.4224428,"longitude":-122.0842467},"formattedAddress":"Prazska 4A"},
				"auxiliaryHeating":{"state":"OFF","startMode":"HEATING","durationInSeconds":600},
				"activeVentilation":{"state":"OFF","durationInSeconds":300},
				"chargingProfiles": {
					"profiles":[{
						"id":123456,
						"name":"Home",
						"settings":{"maxChargingCurrent":"MAXIMUM","minBatteryStateOfCharge":{"enabled":true,"minimumBatteryStateOfChargeInPercent":20}},
						"preferredChargingTimes":[{"id":1,"enabled":true,"startTime":"22:00","endTime":"06:00"}],
						"timers":[{"id":2,"enabled":true,"time":"07:30","type":"RECURRING","recurringOn":["MONDAY","FRIDAY"]}]
					}],
					"currentVehiclePositionProfile":{"id":123456,"name":"Home","targetStateOfChargeInPercent":80,"nextChargingTime":"22:00"}
				}
			},
			"errors": [{"type":"RENDER_UNAVAILABLE","description":"The render could not be retrieved."}]
		}`))
		if err != nil {
			t.Errorf("writing response: %v", err)
		}
	})
	defer server.Close()

	response, err := client.Vehicle(
		context.Background(),
		testVIN,
		IncludeInfo,
		IncludeStatus,
		IncludeFuelStatus,
		IncludeOdometer,
		IncludeParkingPosition,
		IncludeAirConditioning,
		IncludeAuxiliaryHeating,
		IncludeActiveVentilation,
		IncludeCharging,
		IncludeChargingProfiles,
	)
	if err != nil {
		t.Fatalf("Vehicle() error = %v", err)
	}
	if response.Vehicle.Name == nil || *response.Vehicle.Name != "My Enyaq" {
		t.Errorf("name = %v", response.Vehicle.Name)
	}
	if response.Vehicle.Status == nil || response.Vehicle.Status.Overall.DoorsLocked != "YES" {
		t.Fatal("vehicle status was not decoded")
	}
	if response.Vehicle.Charging == nil || response.Vehicle.Charging.Status == nil {
		t.Fatal("charging status was not decoded")
	}
	if response.Vehicle.Charging.Status.State == nil || *response.Vehicle.Charging.Status.State != ChargingStateCharging {
		t.Errorf("charging state = %v", response.Vehicle.Charging.Status.State)
	}
	if response.Vehicle.Charging.Status.Battery == nil || *response.Vehicle.Charging.Status.Battery.StateOfChargeInPercent != 75 {
		t.Fatal("battery state of charge was not decoded")
	}
	if response.Vehicle.AirConditioning == nil || response.Vehicle.AirConditioning.TargetTemperature == nil {
		t.Fatal("air conditioning was not decoded")
	}
	if got := response.Vehicle.AirConditioning.TargetTemperature.Value; got != 21.5 {
		t.Errorf("target temperature = %v", got)
	}
	if response.Vehicle.FuelStatus == nil || response.Vehicle.FuelStatus.PrimaryEngineRange == nil {
		t.Fatal("fuel status was not decoded")
	}
	if got := *response.Vehicle.FuelStatus.PrimaryEngineRange.CurrentFuelLevelInPercent; got != 87 {
		t.Errorf("fuel level = %v", got)
	}
	if response.Vehicle.Odometer == nil || response.Vehicle.Odometer.MileageInKm != 12753 {
		t.Errorf("odometer = %#v", response.Vehicle.Odometer)
	}
	if response.Vehicle.ParkingPosition == nil || response.Vehicle.ParkingPosition.GpsCoordinates == nil {
		t.Fatal("parking position was not decoded")
	}
	if got := response.Vehicle.ParkingPosition.GpsCoordinates.Latitude; got != 37.4224428 {
		t.Errorf("latitude = %v", got)
	}
	if response.Vehicle.AuxiliaryHeating == nil || response.Vehicle.AuxiliaryHeating.DurationInSeconds == nil {
		t.Fatal("auxiliary heating was not decoded")
	}
	if response.Vehicle.ActiveVentilation == nil || response.Vehicle.ActiveVentilation.DurationInSeconds == nil {
		t.Fatal("active ventilation was not decoded")
	}
	if response.Vehicle.ChargingProfiles == nil || len(response.Vehicle.ChargingProfiles.Profiles) != 1 {
		t.Fatal("charging profiles were not decoded")
	}
	profile := response.Vehicle.ChargingProfiles.Profiles[0]
	if len(profile.PreferredChargingTimes) != 1 || len(profile.Timers) != 1 {
		t.Errorf("charging profile = %#v", profile)
	}
	if response.Errors == nil || len(*response.Errors) != 1 || (*response.Errors)[0].Type != VehicleErrorRenderUnavailable {
		t.Errorf("errors = %#v", response.Errors)
	}
	if response.Metadata.APIKeyExpiresAt != "2026-09-30T00:00:00Z" {
		t.Errorf("API key expiry = %q", response.Metadata.APIKeyExpiresAt)
	}
	if response.Metadata.RateLimitRemaining != "19" {
		t.Errorf("rate limit remaining = %q", response.Metadata.RateLimitRemaining)
	}
}

func TestVehicleWithoutIncludes(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("query = %q, want empty", r.URL.RawQuery)
		}
		_, err := w.Write([]byte(`{"vehicle":{"vin":"TMBJB9NY6RF123456"},"errors":[]}`))
		if err != nil {
			t.Errorf("writing response: %v", err)
		}
	})
	defer server.Close()

	if _, err := client.Vehicle(context.Background(), testVIN); err != nil {
		t.Fatalf("Vehicle() error = %v", err)
	}
}

func TestVehicleValidatesRequest(t *testing.T) {
	var requests atomic.Int32
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer server.Close()

	tests := []struct {
		name     string
		vin      string
		includes []Include
		want     string
	}{
		{name: "short VIN", vin: "123", want: "17 characters"},
		{name: "invalid include", vin: testVIN, includes: []Include{"unknown"}, want: "invalid vehicle include"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.Vehicle(context.Background(), tt.vin, tt.includes...)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Vehicle() error = %v, want containing %q", err, tt.want)
			}
		})
	}
	if got := requests.Load(); got != 0 {
		t.Errorf("requests = %d, want 0", got)
	}
}

func TestGuardRedirectsStripsAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		from     string
		to       string
		wantKept bool
	}{
		{name: "same origin", from: "https://api.example.com/a", to: "https://api.example.com/b", wantKept: true},
		{name: "other host", from: "https://api.example.com/a", to: "https://evil.example.net/a", wantKept: false},
		{name: "scheme downgrade", from: "https://api.example.com/a", to: "http://api.example.com/a", wantKept: false},
		{name: "other port", from: "https://api.example.com/a", to: "https://api.example.com:8443/a", wantKept: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guarded := guardRedirects(&http.Client{})
			original, err := http.NewRequest(http.MethodGet, tt.from, nil)
			if err != nil {
				t.Fatalf("NewRequest() error = %v", err)
			}
			redirect, err := http.NewRequest(http.MethodGet, tt.to, nil)
			if err != nil {
				t.Fatalf("NewRequest() error = %v", err)
			}
			redirect.Header.Set("X-API-Key", "test-api-key")

			if err := guarded.CheckRedirect(redirect, []*http.Request{original}); err != http.ErrUseLastResponse {
				t.Errorf("CheckRedirect() = %v, want ErrUseLastResponse", err)
			}
			if kept := redirect.Header.Get("X-API-Key") != ""; kept != tt.wantKept {
				t.Errorf("API key kept = %v, want %v", kept, tt.wantKept)
			}
		})
	}
}

func TestVehicleDoesNotLeakAPIKeyOnRedirect(t *testing.T) {
	var leakedKey atomic.Value
	leakedKey.Store("")
	elsewhere := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		leakedKey.Store(r.Header.Get("X-API-Key"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"vehicle":{"vin":"TMBJB9NY6RF123456"}}`))
	}))
	defer elsewhere.Close()

	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, elsewhere.URL+r.URL.Path, http.StatusFound)
	})
	defer server.Close()

	_, err := client.Vehicle(context.Background(), testVIN)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("Vehicle() error = %v, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusFound {
		t.Errorf("status = %d, want %d (redirect must not be followed)", apiErr.StatusCode, http.StatusFound)
	}
	if got := leakedKey.Load().(string); got != "" {
		t.Errorf("API key leaked to redirect target: %q", got)
	}
}

func TestVehicleStripsAPIKeyWhenCallerFollowsRedirects(t *testing.T) {
	var receivedKey atomic.Value
	receivedKey.Store("unset")
	elsewhere := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedKey.Store(r.Header.Get("X-API-Key"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"vehicle":{"vin":"TMBJB9NY6RF123456"}}`))
	}))
	defer elsewhere.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, elsewhere.URL+r.URL.Path, http.StatusFound)
	}))
	defer server.Close()

	followRedirects := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return nil },
	}
	client, err := NewClient("test-api-key", WithBaseURL(server.URL), WithHTTPClient(followRedirects))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if _, err := client.Vehicle(context.Background(), testVIN); err != nil {
		t.Fatalf("Vehicle() error = %v", err)
	}
	if got := receivedKey.Load().(string); got != "" {
		t.Errorf("API key leaked to redirect target: %q", got)
	}
	if followRedirects.CheckRedirect == nil {
		t.Error("caller-supplied HTTP client was mutated")
	}
}

func TestVehicleRequiresAPIKey(t *testing.T) {
	client, err := NewClient("")
	if client != nil {
		t.Errorf("NewClient() client = %#v, want nil", client)
	}
	if err == nil || !strings.Contains(err.Error(), "API key is required") {
		t.Fatalf("NewClient() error = %v", err)
	}
}

func TestAPIError(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("RateLimit-Limit", "20")
		w.Header().Set("RateLimit-Remaining", "0")
		w.Header().Set("RateLimit-Reset", "1800")
		w.Header().Set("Retry-After", "1800")
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, err := w.Write([]byte(`{
			"type":"https://public.api.connect.skoda-auto.cz/problems/rate-limit-exceeded",
			"title":"Too Many Requests",
			"status":429,
			"detail":"The API rate limit has been exceeded.",
			"instance":"/api/v1/vehicles/TMBJB9NY6RF123456"
		}`))
		if err != nil {
			t.Errorf("writing response: %v", err)
		}
	})
	defer server.Close()

	_, err := client.Vehicle(context.Background(), testVIN)
	if err == nil {
		t.Fatal("Vehicle() error = nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusTooManyRequests {
		t.Errorf("status = %d", apiErr.StatusCode)
	}
	if apiErr.Problem == nil || apiErr.Problem.Type == nil || !strings.HasSuffix(*apiErr.Problem.Type, "/rate-limit-exceeded") {
		t.Errorf("problem = %#v", apiErr.Problem)
	}
	if apiErr.Metadata.RetryAfter != "1800" {
		t.Errorf("retry after = %q", apiErr.Metadata.RetryAfter)
	}
	if !strings.Contains(err.Error(), "API rate limit has been exceeded") {
		t.Errorf("error = %q", err)
	}
}

func TestConvenienceReads(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("include") {
		case "charging":
			_, err := w.Write([]byte(`{"vehicle":{"vin":"TMBJB9NY6RF123456","charging":{"isVehicleInSavedLocation":false,"status":{"state":"READY_FOR_CHARGING"}}}}`))
			if err != nil {
				t.Errorf("writing charging response: %v", err)
			}
		case "airConditioning":
			_, err := w.Write([]byte(`{"vehicle":{"vin":"TMBJB9NY6RF123456","airConditioning":{"state":"OFF"}}}`))
			if err != nil {
				t.Errorf("writing air-conditioning response: %v", err)
			}
		default:
			t.Errorf("unexpected include %q", r.URL.Query().Get("include"))
		}
	})
	defer server.Close()

	charging, err := client.Charging(context.Background(), testVIN)
	if err != nil {
		t.Fatalf("Charging() error = %v", err)
	}
	if charging.Status == nil || charging.Status.State == nil || *charging.Status.State != ChargingStateReadyForCharging {
		t.Errorf("charging = %#v", charging)
	}

	airConditioning, err := client.AirConditioning(context.Background(), testVIN)
	if err != nil {
		t.Fatalf("AirConditioning() error = %v", err)
	}
	if airConditioning.State != AirConditioningStateOff {
		t.Errorf("air conditioning state = %q", airConditioning.State)
	}
}

func TestConvenienceReadReportsPartialDataError(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`{
			"vehicle":{"vin":"TMBJB9NY6RF123456"},
			"errors":[{"type":"CHARGING_UNSUPPORTED","description":"Charging is not supported."}]
		}`))
		if err != nil {
			t.Errorf("writing response: %v", err)
		}
	})
	defer server.Close()

	_, err := client.Charging(context.Background(), testVIN)
	var dataErr *VehicleDataError
	if !errors.As(err, &dataErr) {
		t.Fatalf("Charging() error = %v, want *VehicleDataError", err)
	}
	if dataErr.Include != IncludeCharging || len(dataErr.Errors) != 1 {
		t.Errorf("data error = %#v", dataErr)
	}
}
