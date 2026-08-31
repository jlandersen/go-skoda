package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	skoda "github.com/jlandersen/go-skoda"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "vehicle", "info":
		cmdVehicle(args)
	case "charging":
		cmdCharging(args)
	case "ac":
		cmdAirConditioning(args)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprint(os.Stderr, `Usage: go-skoda <command> [arguments]

Commands:
  vehicle -vin=VIN             Show all available vehicle data
  charging -vin=VIN            Show charging status
  ac -vin=VIN                  Show air-conditioning status

Environment variables:
  SKODA_API_KEY                API key created in the MySkoda app
  SKODA_VIN                    Default VIN (used when -vin is not provided)
`)
}

func cmdVehicle(args []string) {
	client, vin := clientAndVIN(args)
	vehicle, err := client.Vehicle(context.Background(), vin)
	if err != nil {
		fatal("vehicle: %v", err)
	}
	printJSON(vehicle)
}

func cmdCharging(args []string) {
	client, vin := clientAndVIN(args)
	charging, err := client.Charging(context.Background(), vin)
	if err != nil {
		fatal("charging: %v", err)
	}
	printJSON(charging)
}

func cmdAirConditioning(args []string) {
	client, vin := clientAndVIN(args)
	airConditioning, err := client.AirConditioning(context.Background(), vin)
	if err != nil {
		fatal("air conditioning: %v", err)
	}
	printJSON(airConditioning)
}

func clientAndVIN(args []string) (*skoda.Client, string) {
	apiKey := strings.TrimSpace(os.Getenv("SKODA_API_KEY"))
	if apiKey == "" {
		fatal("SKODA_API_KEY is required; create a key in the MySkoda app")
	}

	vin := envOrFlag(args, "vin", "SKODA_VIN")
	if vin == "" {
		fatal("VIN required (use -vin or SKODA_VIN)")
	}
	client, err := skoda.NewClient(apiKey)
	if err != nil {
		fatal("client: %v", err)
	}
	return client, vin
}

func envOrFlag(args []string, flagName, envName string) string {
	prefix := "-" + flagName + "="
	for _, arg := range args {
		if strings.HasPrefix(arg, prefix) {
			return strings.TrimPrefix(arg, prefix)
		}
	}
	return os.Getenv(envName)
}

func printJSON(value any) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		fatal("encoding output: %v", err)
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
