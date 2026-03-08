package skoda

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)

	client := NewClient()
	client.apiURL = server.URL
	client.baseURL = server.URL
	client.tokens = &IDKSession{
		AccessToken:  "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJleHAiOjk5OTk5OTk5OTl9.",
		RefreshToken: "fake-refresh-token",
		IDToken:      "fake-id-token",
	}

	return client, server
}

func TestGarage(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/v2/garage") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("missing Authorization header")
		}

		json.NewEncoder(w).Encode(garageResponse{
			Vehicles: []GarageEntry{
				{VIN: "TMBJB9NY6RF123456", Name: "My Enyaq", State: "ACTIVATED", Title: "ENYAQ iV 80"},
			},
		})
	})
	defer server.Close()

	vehicles, err := client.Garage(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vehicles) != 1 {
		t.Fatalf("expected 1 vehicle, got %d", len(vehicles))
	}
	if vehicles[0].VIN != "TMBJB9NY6RF123456" {
		t.Errorf("expected VIN TMBJB9NY6RF123456, got %s", vehicles[0].VIN)
	}
}

func TestVehicleInfo(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/v2/garage/vehicles/") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		json.NewEncoder(w).Encode(VehicleInfo{
			VIN:   "TMBJB9NY6RF123456",
			Name:  "My Enyaq",
			State: "ACTIVATED",
			Specification: Specification{
				Title:     "ENYAQ iV 80",
				Model:     "ENYAQ",
				ModelYear: "2024",
				Body:      "SUV",
				Engine:    Engine{Type: "electric", PowerInKW: 150},
				Battery:   &Battery{CapacityInKWh: 77},
			},
			Capabilities: Capabilities{
				Capabilities: []Capability{
					{ID: CapabilityCharging, Statuses: []string{}},
					{ID: CapabilityAirConditioning, Statuses: []string{}},
					{ID: CapabilityVehicleHealthWarningsWakeUp, Statuses: []string{}},
				},
			},
		})
	})
	defer server.Close()

	info, err := client.VehicleInfo(context.Background(), "TMBJB9NY6RF123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.VIN != "TMBJB9NY6RF123456" {
		t.Errorf("expected VIN TMBJB9NY6RF123456, got %s", info.VIN)
	}
	if !info.HasCapability(CapabilityCharging) {
		t.Error("expected vehicle to have CHARGING capability")
	}
	if !info.IsCapabilityAvailable(CapabilityAirConditioning) {
		t.Error("expected AIR_CONDITIONING to be available")
	}
	if info.HasCapability(CapabilityHonkAndFlash) {
		t.Error("expected vehicle to NOT have HONK_AND_FLASH capability")
	}
	if info.Specification.Battery == nil {
		t.Fatal("expected battery to be present")
	}
	if info.Specification.Battery.CapacityInKWh != 77 {
		t.Errorf("expected battery capacity 77, got %d", info.Specification.Battery.CapacityInKWh)
	}
}

func TestCharging(t *testing.T) {
	soc := 75
	rangeMeters := int64(310000)
	power := 11.0
	rate := 38.5
	remaining := int64(120)

	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Charging{
			IsVehicleInSavedLocation: true,
			Status: &ChargingStatus{
				Battery: ChargingBattery{
					StateOfChargeInPercent:         &soc,
					RemainingCruisingRangeInMeters: &rangeMeters,
				},
				State:                                ChargingStateCharging,
				ChargePowerInKW:                      &power,
				ChargingRateInKilometersPerHour:      &rate,
				ChargeType:                           ChargeTypeAC,
				RemainingTimeToFullyChargedInMinutes: &remaining,
			},
		})
	})
	defer server.Close()

	charging, err := client.Charging(context.Background(), "TMBJB9NY6RF123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if charging.Status == nil {
		t.Fatal("expected charging status to be present")
	}
	if charging.Status.State != ChargingStateCharging {
		t.Errorf("expected state CHARGING, got %s", charging.Status.State)
	}
	if *charging.Status.Battery.StateOfChargeInPercent != 75 {
		t.Errorf("expected SoC 75, got %d", *charging.Status.Battery.StateOfChargeInPercent)
	}
	if *charging.Status.ChargePowerInKW != 11.0 {
		t.Errorf("expected charge power 11.0, got %f", *charging.Status.ChargePowerInKW)
	}
}

func TestAirConditioning(t *testing.T) {
	windowHeating := true

	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(AirConditioning{
			State:                  AirConditioningStateOff,
			ChargerConnectionState: ConnectionStateConnected,
			ChargerLockState:       ChargerLockedStateLocked,
			WindowHeatingEnabled:   &windowHeating,
			TargetTemperature: &TargetTemperature{
				TemperatureValue: 21.5,
				UnitInCar:        "CELSIUS",
			},
		})
	})
	defer server.Close()

	ac, err := client.AirConditioning(context.Background(), "TMBJB9NY6RF123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ac.State != AirConditioningStateOff {
		t.Errorf("expected state OFF, got %s", ac.State)
	}
	if ac.ChargerConnectionState != ConnectionStateConnected {
		t.Errorf("expected charger CONNECTED, got %s", ac.ChargerConnectionState)
	}
	if ac.ChargerLockState != ChargerLockedStateLocked {
		t.Errorf("expected charger LOCKED, got %s", ac.ChargerLockState)
	}
	if ac.TargetTemperature == nil || ac.TargetTemperature.TemperatureValue != 21.5 {
		t.Error("expected target temperature 21.5")
	}
}

func TestAPIError(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	_, err := client.Garage(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "403") {
		t.Errorf("expected error to contain 403, got: %s", errStr)
	}
}

func TestIsTokenExpired(t *testing.T) {
	validToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJleHAiOjk5OTk5OTk5OTl9."
	if isTokenExpired(validToken) {
		t.Error("expected token to NOT be expired")
	}

	expiredToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJleHAiOjB9."
	if !isTokenExpired(expiredToken) {
		t.Error("expected token to be expired")
	}

	if !isTokenExpired("not-a-jwt") {
		t.Error("expected invalid token to be treated as expired")
	}
}

func TestTokenRefresh(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/authentication/refresh-token") {
			json.NewEncoder(w).Encode(IDKSession{
				AccessToken:  "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJleHAiOjk5OTk5OTk5OTl9.",
				RefreshToken: "new-refresh-token",
				IDToken:      "new-id-token",
			})
			return
		}
		if strings.Contains(r.URL.Path, "/v2/garage") {
			json.NewEncoder(w).Encode(garageResponse{Vehicles: []GarageEntry{}})
			return
		}
		w.WriteHeader(404)
	}))
	defer server.Close()

	client := NewClient()
	client.apiURL = server.URL
	client.baseURL = server.URL
	client.tokens = &IDKSession{
		AccessToken:  "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJleHAiOjB9.",
		RefreshToken: "old-refresh-token",
		IDToken:      "old-id-token",
	}

	_, err := client.Garage(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.tokens.RefreshToken != "new-refresh-token" {
		t.Errorf("expected refresh token to be updated, got %s", client.tokens.RefreshToken)
	}
}

func TestCSRFParsing(t *testing.T) {
	html := `<html><head><script>
window._IDK = {
    csrf_token: 'csrf-token-123',
    templateModel: {"hmac":"hmac-789","relayState":"relay-state-456"},
    csrf_parameterName: '_csrf',
}
</script></head></html>`

	csrf, err := parseCSRF(html)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if csrf.csrf != "csrf-token-123" {
		t.Errorf("expected csrf 'csrf-token-123', got '%s'", csrf.csrf)
	}
	if csrf.relayState != "relay-state-456" {
		t.Errorf("expected relayState 'relay-state-456', got '%s'", csrf.relayState)
	}
	if csrf.hmac != "hmac-789" {
		t.Errorf("expected hmac 'hmac-789', got '%s'", csrf.hmac)
	}
}

func TestPKCEChallenge(t *testing.T) {
	verifier := "test-verifier-string"
	c1 := pkceChallenge(verifier)
	c2 := pkceChallenge(verifier)
	if c1 != c2 {
		t.Error("expected same challenge for same verifier")
	}
	if c1 == "" {
		t.Error("expected non-empty challenge")
	}
}
