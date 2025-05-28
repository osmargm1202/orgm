package cmd

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check the application",
	Long:  `Check the connectivity to the application servers and endpoints using TCP ping`,
	Run: func(cmd *cobra.Command, args []string) {
		VerifyUrls()
		verifyCloudUrl()
		verifyPostgrestUrl()
		verifyImgUrl()
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

var ApiEndPoints = []string{
	"/cot",
	"/fac",
	"/firmapdf",
}

func VerifyUrls() {
	apiUrl, _ := InitializeApi()

	for _, endpoint := range ApiEndPoints {
		fullEndpoint := apiUrl + endpoint
		fmt.Println(inputs.InfoStyle.Render(fmt.Sprintf("Ping %s", fullEndpoint)))
		start := time.Now()
		res := pingEndpoint(fullEndpoint)
		elapsed := time.Since(start)
		if res == "OK" {
			fmt.Printf("%s (%dms)\n", inputs.SuccessStyle.Render(res), elapsed.Milliseconds())
		} else {
			fmt.Printf("%s (%dms)\n", inputs.ErrorStyle.Render(res), elapsed.Milliseconds())
		}
	}
}

func verifyCloudUrl() {
	url := "https://cloud.orgmapp.com"
	fmt.Println(inputs.InfoStyle.Render(fmt.Sprintf("Ping %s", url)))
	start := time.Now()
	res := pingEndpoint(url)
	elapsed := time.Since(start)
	if res == "OK" {
		fmt.Printf("%s (%dms)\n", inputs.SuccessStyle.Render(res), elapsed.Milliseconds())
	} else {
		fmt.Printf("%s (%dms)\n", inputs.ErrorStyle.Render(res), elapsed.Milliseconds())
	}
}

func verifyPostgrestUrl() {
	postgrestUrl, _ := InitializePostgrest()

	fmt.Println(inputs.InfoStyle.Render(fmt.Sprintf("Ping %s", postgrestUrl)))
	start := time.Now()
	res := pingEndpoint(postgrestUrl)
	elapsed := time.Since(start)
	if res == "OK" {
		fmt.Printf("%s (%dms)\n", inputs.SuccessStyle.Render(res), elapsed.Milliseconds())
	} else {
		fmt.Printf("%s (%dms)\n", inputs.ErrorStyle.Render(res), elapsed.Milliseconds())
	}
}

func verifyImgUrl() {
	url := "https://img.orgmapp.com/list/images/test"
	fmt.Println(inputs.InfoStyle.Render(fmt.Sprintf("Ping %s", url)))
	start := time.Now()
	res := pingEndpoint(url)
	elapsed := time.Since(start)
	if res == "OK" {
		fmt.Printf("%s (%dms)\n", inputs.SuccessStyle.Render(res), elapsed.Milliseconds())
	} else {
		fmt.Printf("%s (%dms)\n", inputs.ErrorStyle.Render(res), elapsed.Milliseconds())
	}
}

func init() {
	RootCmd.AddCommand(checkCmd)
}
