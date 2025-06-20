package config

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

func Warning(message string) {
	yellow := color.New(color.FgYellow).SprintFunc()
	fmt.Println(yellow("[WARN] %s", message))
}

func LoadEnv() {
	env := os.Getenv("APP_ENV")

	if env == "development" {
		if err := godotenv.Load(); err != nil {
			Warning("Failed to load .env file. Continuing without.")
		}
	}
}

func GetEnvVariable(variableToCheck string) (string, bool) {
	envVar := os.Getenv(variableToCheck)

	if envVar != "" {
		return envVar, true
	}

	return envVar, false

}
