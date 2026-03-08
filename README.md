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

	skoda "github.com/jlandersen/go-skoda"
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

## Disclaimer

This project is not an official API client for the Skoda API. It is not affiliated with, not is it endorsed by, or associated with Skoda Auto or any of its subsidiaries.

Use this project at your own risk. Skoda Auto may update or modify its API without notice, which could render this client inoperative or non-compliant. The maintainers of this project are not responsible for any misuse, legal implications, or damages arising from its use.

Ensure compliance with Skoda Auto's terms of service and any applicable laws when using this software.

## License

[MIT](LICENSE)
