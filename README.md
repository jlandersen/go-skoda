# go-skoda

A Go package for interacting with the Skoda Connect API. Provides both a Go client library and a CLI tool for querying vehicle data. Zero dependencies.

Inspired by [myskoda](https://github.com/skodaconnect/myskoda) (Python).

## Features

- OAuth2/OIDC authentication with PKCE against VW Group identity provider
- Automatic token refresh
- Read-only endpoints that do not wake the vehicle unless stated otherwise

### Supported Endpoints

| Endpoint | Description |
|---|---|
| Garage | List all vehicles |
| Vehicle Info | Specs, capabilities, software version |
| Charging | SoC, charge power, remaining time, charge type |
| Air Conditioning | AC state, charger connection/lock, target temperature, seat heating |

## Installation

```
go install github.com/jlandersen/go-skoda@latest
```

## Library Usage

```go
package main

import (
	"context"
	"fmt"

	"github.com/jlandersen/go-skoda/internal/skoda"
)

func main() {
	ctx := context.Background()
	client := skoda.NewClient()

	_ = client.Login(ctx, "you@example.com", "secret")

	vehicles, _ := client.Garage(ctx)
	for _, v := range vehicles {
		fmt.Println(v.VIN, v.Name)
	}

	charging, _ := client.Charging(ctx, vehicles[0].VIN)
	fmt.Printf("SoC: %d%%\n", *charging.Status.Battery.StateOfChargeInPercent)
}
```

## License

[MIT](LICENSE)
