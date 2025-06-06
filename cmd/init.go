package cmd

import (
	"log"

	"github.com/spf13/viper"
	"github.com/studio-b12/gowebdav"
)

func InitializePostgrest() (string, map[string]string) {

	// Get PostgREST URL from configs
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

func InitializeNextcloud() *gowebdav.Client {

	// Get Nextcloud URL from config
	nextcloudURL := viper.GetString("nextcloud.url")
	if nextcloudURL == "" {
		log.Fatal("Error: nextcloud.url is not defined in config file")
		return nil
	}
	username := viper.GetString("nextcloud.username")
	password := viper.GetString("nextcloud.password")

	nextcloudURL = nextcloudURL + "/remote.php/dav/files/" + username

	client := gowebdav.NewClient(nextcloudURL, username, password)

	// if err := client.Connect(); err != nil {
	// 	log.Fatal("Error connecting to Nextcloud:", err)
	// 	return nil
	// }

	return client
}
