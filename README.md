# go-skoda

A zero-dependency Go client and CLI for the official [MySkoda Public API](https://public.api.connect.skoda-auto.cz/docs).

## Features

- API-key authentication through the public API; the key is never sent across a redirect to another host
- Full vehicle state: doors, windows, lights, fuel, odometer, parking position, climate, charging, and charging profiles
- Selective reads through the public API's `include` parameter
- Partial-data errors and API-key/rate-limit response headers
- Read-only access; this package does not send vehicle commands

## API key

Create an API key in the MySkoda app. Keys are restricted to the vehicles selected when they are created and expire. The public API returns the expiry in `X-API-Key-Expires-At`.

The public API does not provide a garage or vehicle-list endpoint. Callers must retain the VINs selected for the key.

## Installation

```sh
go get github.com/jlandersen/go-skoda
```

Build the CLI from a checkout with `go build -o go-skoda ./cmd/cli`.

## Library usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	skoda "github.com/jlandersen/go-skoda"
)

func main() {
	client, err := skoda.NewClient("your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	response, err := client.Vehicle(
		context.Background(),
		"TMBJB9NY6RF123456",
		skoda.IncludeInfo,
		skoda.IncludeCharging,
	)
	if err != nil {
		log.Fatal(err)
	}

	if response.Vehicle.Name != nil {
		fmt.Println(*response.Vehicle.Name)
	}
	if charging := response.Vehicle.Charging; charging != nil && charging.Status != nil {
		if charging.Status.State != nil {
			fmt.Println(*charging.Status.State)
		}
	}
	fmt.Println("requests remaining:", response.Metadata.RateLimitRemaining)

	if response.Errors != nil {
		for _, partialErr := range *response.Errors {
			fmt.Println(partialErr.Type)
		}
	}
}
```

Omit the include values to request all supported data. The API currently permits 20 requests per hour for each key. Read the `RateLimit-*` values in `VehicleResponse.Metadata` instead of assuming that limit is fixed.

The `Charging` and `AirConditioning` methods are convenience reads. They request only their corresponding section and return `VehicleDataError` if the public API omits it.

## Updating the OpenAPI schema

The official schema is pinned in `api/openapi.json`. The response models in `internal/api/models.gen.go` are generated from it. Only the read-only `getVehicle` operation is included.

Run `make generate` after changing the schema. Generation runs `oapi-codegen` v2.8.0 inside a digest-pinned Go Docker image. The committed models import only the standard library, so library users do not need Docker or generator dependencies. Run `make update-openapi` to download the current official schema and regenerate the models.

## CLI

```sh
export SKODA_API_KEY=your-api-key
export SKODA_VIN=TMBJB9NY6RF123456

go-skoda vehicle
go-skoda charging
go-skoda ac
```

## Migration from the private API

The private mobile API login, refresh-token storage, and embedded mobile-app client ID have been removed. Construct a client with an API key instead:

```go
client, err := skoda.NewClient(os.Getenv("SKODA_API_KEY"))
```

`Garage` and the old specification/capability response are not available in the public API. Use `Vehicle` with a known VIN. Charging and air-conditioning data now follows the public API schema, including `TargetTemperature.Value` and `TargetTemperature.Unit`.

## Disclaimer

This project is not affiliated with, endorsed by, or associated with Skoda Auto or any of its subsidiaries.

Use this project at your own risk. Ensure compliance with Skoda Auto's terms of service and applicable laws.

## License

[MIT](LICENSE)
