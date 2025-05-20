package cmd

import (
	"fmt"
	"net/http"
	"time"

	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check the application",
	Long:  `Check the urls of the application and configuration files`,
	Run: func(cmd *cobra.Command, args []string) {
		VerifyUrls()
	},
}

type url struct {
	busqueda string
	headers  map[string]string
}

func convertUrls(busqueda string, headers map[string]string) url {
	return url{
		busqueda: busqueda,
		headers:  headers,
	}
}

func defineList() []url {
		
	var listUrls = []url{

		func() url {
			url, headers := InitializeApi()
			return convertUrls(url, headers)
		}(),
		func() url {
			url, headers := InitializePostgrest()
			return convertUrls(url, headers)
		}(),
		func() url {
			headers := make(map[string]string)
			headers["Content-Type"] = "text/html"
			headers["accept"] = "text/html"
			return convertUrls("https://cloud.orgmapp.com", headers)
		}(),
	}

	return listUrls
}

func checkUrl(pagina url) string {

	req, err := http.NewRequest("GET", pagina.busqueda, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return ""
	}

	for key, value := range pagina.headers {
		req.Header.Add(key, value)
	}

	// Make request with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return ""
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		fmt.Println("API returned status code:", resp.StatusCode)
		return ""
	}

	return "OK"
}

func VerifyUrls() {
	listUrls := defineList()
	for _, url := range listUrls {
		fmt.Println(inputs.InfoStyle.Render(url.busqueda))
		res := checkUrl(url)
		if res == "OK" {
			fmt.Println(inputs.SuccessStyle.Render(res))
		} else {
			fmt.Println(inputs.ErrorStyle.Render(res))
		}
	}
}

func init() {
	RootCmd.AddCommand(checkCmd)
}
