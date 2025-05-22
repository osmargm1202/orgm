package adm

import (
	"log"
	"github.com/spf13/viper"
)

func InitializePostgrest() (string, map[string]string) {

	// Get PostgREST URL from config
	postgrestURL := viper.GetString("url.postgrest")
	if postgrestURL == "" {
		log.Fatal("Error: url.postgrest is not defined in config file")
		return "", nil
	}

	// Initialize headers
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Prefer"] = "return=representation"
	headers["accept"] = "application/json"
	headers["CF-Access-Client-Id"] = viper.GetString("cloudflare.CF_ACCESS_CLIENT_ID")
	headers["CF-Access-Client-Secret"] = viper.GetString("cloudflare.CF_ACCESS_CLIENT_SECRET")

	return postgrestURL, headers
}

func InitializeApi() (string, map[string]string) {

	// Get API URL from config
	apiURL := viper.GetString("url.apis")
	if apiURL == "" {
		log.Fatal("Error: url.apis is not defined in config file")
		return "", nil
	}

	// Initialize headers
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["accept"] = "application/json"
	headers["CF-Access-Client-Id"] = viper.GetString("cloudflare.CF_ACCESS_CLIENT_ID")
	headers["CF-Access-Client-Secret"] = viper.GetString("cloudflare.CF_ACCESS_CLIENT_SECRET")

	return apiURL, headers
}