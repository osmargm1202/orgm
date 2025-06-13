package misc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var DivisaCmd = &cobra.Command{
	Use:   "divisa",
	Short: "Divisa command",
	Long:  `Divisa command`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			func() (float64, error) {
				a := args[0]
				cantidad := 1.0
				rest, err := DivisaUSD(a, cantidad)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(rest)
				return rest, nil
			}()
		} else if len(args) == 2 {
			func() (float64, error) {
				desde := args[0]
				cantidad, err := strconv.ParseFloat(args[1], 64)
				if err != nil {
					fmt.Println(err)
				}
				rest, err := DivisaUSD(desde, cantidad)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(rest)
				return rest, nil
			}()
		} else {
			resp, err := DivisaUSD("DOP", 1)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(resp)
		}
	},
}

// DivisaRequest represents the request body for currency conversion
type DivisaRequest struct {
	Desde    string  `json:"desde"`
	A        string  `json:"a"`
	Cantidad float64 `json:"cantidad"`
}

// DivisaResponse represents the response from the API
type DivisaResponse struct {
	Resultado float64 `json:"resultado"`
}

func DivisaUSD(a string, cantidad float64) (float64, error) {

	desde := "USD"
	if cantidad < 1 {
		cantidad = 1
	}

	if a == "" {
		a = "DOP"
	}

	url, headers := InitializeApi()
	if url == "" {
		return 0, fmt.Errorf("error getting API URL")
	}

	// Create request body
	reqBody := DivisaRequest{
		Desde:    desde,
		A:        a,
		Cantidad: cantidad,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return 0, fmt.Errorf("error marshaling request: %v", err)
	}

	// Create request
	req, err := http.NewRequest("POST", url+"/divisa", bytes.NewBuffer(jsonBody))
	if err != nil {
		return 0, fmt.Errorf("error creating request: %v", err)
	}

	// Add headers
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	// Make request with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	// Parse response
	var response DivisaResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("error decoding response: %v", err)
	}

	return response.Resultado, nil
}


func init() {
	MiscCmd.AddCommand(DivisaCmd)
}