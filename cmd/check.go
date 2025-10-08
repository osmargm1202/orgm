package cmd

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check the application",
	Long:  `Check the connectivity to the application servers and endpoints using TCP ping`,
	Run: func(cmd *cobra.Command, args []string) {
		verifyAllUrls()
	},
}

func extractHostAndPort(urlStr string) (string, string, error) {
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "https://" + urlStr
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", "", err
	}

	hostname := parsedURL.Hostname()
	port := parsedURL.Port()

	// Set default ports if not specified
	if port == "" {
		if parsedURL.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	return hostname, port, nil
}

func pingEndpoint(fullUrl string) string {
	hostname, port, err := extractHostAndPort(fullUrl)
	if err != nil {
		return "Invalid URL"
	}

	address := net.JoinHostPort(hostname, port)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return "Connection failed"
	}
	defer conn.Close()

	return "OK"
}

func verifyAllUrls() {
	// Obtener todas las claves que empiezan con "url."
	allKeys := viper.AllKeys()
	urlKeys := make(map[string]string)
	
	for _, key := range allKeys {
		if strings.HasPrefix(key, "url.") {
			value := viper.GetString(key)
			if value != "" {
				urlKeys[key] = value
			}
		}
	}
	
	if len(urlKeys) == 0 {
		fmt.Printf("%s\n", inputs.WarningStyle.Render("No URL configurations found"))
		return
	}
	
	// Imprimir cada URL y hacer ping
	for key, value := range urlKeys {
		fmt.Println(inputs.InfoStyle.Render(fmt.Sprintf("%s: %s", key, value)))
		start := time.Now()
		res := pingEndpoint(value)
		elapsed := time.Since(start)
		if res == "OK" {
			fmt.Printf("%s (%dms)\n", inputs.SuccessStyle.Render(res), elapsed.Milliseconds())
		} else {
			fmt.Printf("%s (%dms)\n", inputs.ErrorStyle.Render(res), elapsed.Milliseconds())
		}
		fmt.Println() // Línea en blanco para separar
	}
}


func init() {
	RootCmd.AddCommand(checkCmd)
}
