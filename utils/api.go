package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/osmargm1202/orgm/cmd"
)

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

func DivisaUSD(desde string, a string, cantidad float64) (float64, error) {
	if desde == "" {
		desde = "USD"
	}
	if a == "" {
		a = "DOP"
	}
	if cantidad <= 0 {
		cantidad = 1
	}

	url, headers := cmd.InitializeApi()
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

