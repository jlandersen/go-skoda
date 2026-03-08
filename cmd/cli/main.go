package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	skoda "github.com/jlandersen/go-skoda"
)

const tokenFile = ".go-skoda-token"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "login":
		cmdLogin(args)
	case "garage":
		cmdGarage()
	case "info":
		cmdInfo(args)
	case "charging":
		cmdCharging(args)
	case "ac":
		cmdAC(args)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: go-skoda <command> [arguments]

Commands:
  login  -email=EMAIL -password=PASSWORD   Authenticate and store refresh token
  garage                                   List vehicles
  info   -vin=VIN                          Show vehicle details
  charging -vin=VIN                        Show charging status
  ac     -vin=VIN                          Show air conditioning status

Environment variables:
  SKODA_EMAIL          Email for login
  SKODA_PASSWORD       Password for login
  SKODA_VIN            Default VIN (used when -vin is not provided)
  SKODA_REFRESH_TOKEN  Refresh token (overrides stored token)
`)
}

func cmdLogin(args []string) {
	email := envOrFlag(args, "email", "SKODA_EMAIL")
	password := envOrFlag(args, "password", "SKODA_PASSWORD")

	if email == "" || password == "" {
		fatal("email and password required (use -email/-password flags or SKODA_EMAIL/SKODA_PASSWORD env vars)")
	}

	client := skoda.NewClient()
	ctx := context.Background()

	fmt.Fprintf(os.Stderr, "Logging in as %s...\n", email)
	if err := client.Login(ctx, email, password); err != nil {
		fatal("login failed: %v", err)
	}

	token, err := client.GetRefreshToken()
	if err != nil {
		fatal("getting refresh token: %v", err)
	}

	if err := saveToken(token); err != nil {
		fatal("saving token: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Logged in. Refresh token saved to ~/%s\n", tokenFile)
}

func cmdGarage() {
	client := authedClient()

	vehicles, err := client.Garage(context.Background())
	if err != nil {
		fatal("garage: %v", err)
	}

	printJSON(vehicles)
}

func cmdInfo(args []string) {
	vin := requireVIN(args)
	client := authedClient()

	info, err := client.VehicleInfo(context.Background(), vin)
	if err != nil {
		fatal("info: %v", err)
	}

	printJSON(info)
}

func cmdCharging(args []string) {
	vin := requireVIN(args)
	client := authedClient()

	charging, err := client.Charging(context.Background(), vin)
	if err != nil {
		fatal("charging: %v", err)
	}

	printJSON(charging)
}

func cmdAC(args []string) {
	vin := requireVIN(args)
	client := authedClient()

	ac, err := client.AirConditioning(context.Background(), vin)
	if err != nil {
		fatal("ac: %v", err)
	}

	printJSON(ac)
}

func authedClient() *skoda.Client {
	client := skoda.NewClient()
	ctx := context.Background()

	token := os.Getenv("SKODA_REFRESH_TOKEN")
	if token == "" {
		var err error
		token, err = loadToken()
		if err != nil {
			fatal("not authenticated. Run 'go-skoda login' first or set SKODA_REFRESH_TOKEN")
		}
	}

	if err := client.LoginWithRefreshToken(ctx, token); err != nil {
		fatal("authentication failed: %v", err)
	}

	// Save the new refresh token
	if newToken, err := client.GetRefreshToken(); err == nil {
		saveToken(newToken)
	}

	return client
}

func requireVIN(args []string) string {
	vin := envOrFlag(args, "vin", "SKODA_VIN")
	if vin == "" {
		fatal("VIN required (use -vin flag or SKODA_VIN env var)")
	}
	return vin
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

func tokenPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return tokenFile
	}
	return filepath.Join(home, tokenFile)
}

func saveToken(token string) error {
	return os.WriteFile(tokenPath(), []byte(token), 0600)
}

func loadToken() (string, error) {
	data, err := os.ReadFile(tokenPath())
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func printJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
