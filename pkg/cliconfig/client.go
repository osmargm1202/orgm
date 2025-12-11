package cliconfig

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	// WorkerURL is the base URL for the CLI configuration worker
	WorkerURL = "https://cli-config.or-gm.com"
)

// ConfigCache represents the cached configuration
type ConfigCache struct {
	APIURLs map[string]string `json:"api_urls,omitempty"`
}

// getCachePath returns the path to the cache file
func getCachePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error obteniendo directorio home: %v", err)
	}
	configDir := filepath.Join(homeDir, ".config", "orgm")
	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("error creando directorio de configuración: %v", err)
	}
	return filepath.Join(configDir, "config.json"), nil
}

// loadCache loads the cache from config.json
func loadCache() (*ConfigCache, error) {
	cachePath, err := getCachePath()
	if err != nil {
		return nil, err
	}

	cache := &ConfigCache{
		APIURLs: make(map[string]string),
	}

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return cache, nil
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("error leyendo config.json: %v", err)
	}

	if len(data) == 0 {
		return cache, nil
	}

	// Try to parse as full config (with other fields)
	var fullConfig map[string]interface{}
	if err := json.Unmarshal(data, &fullConfig); err == nil {
		if apiURLs, ok := fullConfig["api_urls"].(map[string]interface{}); ok {
			for k, v := range apiURLs {
				if str, ok := v.(string); ok {
					cache.APIURLs[k] = str
				}
			}
		}
	}

	return cache, nil
}

// saveCache saves the cache to config.json (merging with existing config)
func saveCache(cache *ConfigCache) error {
	cachePath, err := getCachePath()
	if err != nil {
		return err
	}

	// Load existing config to merge
	var existingConfig map[string]interface{}
	if _, err := os.Stat(cachePath); err == nil {
		data, err := os.ReadFile(cachePath)
		if err == nil && len(data) > 0 {
			json.Unmarshal(data, &existingConfig)
		}
	}

	if existingConfig == nil {
		existingConfig = make(map[string]interface{})
	}

	// Update API URLs
	existingConfig["api_urls"] = cache.APIURLs

	// Save merged config
	data, err := json.MarshalIndent(existingConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("error codificando config.json: %v", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("error escribiendo config.json: %v", err)
	}

	return nil
}

// ConfigResponse represents the response from the worker API
type ConfigResponse struct {
	Success bool   `json:"success"`
	Key     string `json:"key,omitempty"`
	Value   string `json:"value,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// GetConfig retrieves a configuration value from the worker API
func GetConfig(key string) (string, error) {
	// timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	// log.Printf("[DEBUG %s] Consultando API worker para llave: %s", timestamp, key)

	url := fmt.Sprintf("%s/?key=%s", WorkerURL, key)
	// log.Printf("[DEBUG %s] URL de consulta: %s", timestamp, url)

	resp, err := http.Get(url)
	if err != nil {
		// log.Printf("[DEBUG %s] Error al realizar petición HTTP: %v", timestamp, err)
		return "", fmt.Errorf("error al consultar API worker: %w", err)
	}
	defer resp.Body.Close()

	// log.Printf("[DEBUG %s] Respuesta recibida con status code: %d", timestamp, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// log.Printf("[DEBUG %s] Error al leer cuerpo de respuesta: %v", timestamp, err)
		return "", fmt.Errorf("error al leer respuesta: %w", err)
	}

	// log.Printf("[DEBUG %s] Cuerpo de respuesta: %s", timestamp, string(body))

	var configResp ConfigResponse
	if err := json.Unmarshal(body, &configResp); err != nil {
		// log.Printf("[DEBUG %s] Error al decodificar JSON: %v", timestamp, err)
		return "", fmt.Errorf("error al decodificar respuesta JSON: %w", err)
	}

	if !configResp.Success {
		// log.Printf("[DEBUG %s] Respuesta indica error: %s - %s", timestamp, configResp.Error, configResp.Message)
		return "", fmt.Errorf("error del API worker: %s - %s", configResp.Error, configResp.Message)
	}

	// log.Printf("[DEBUG %s] Valor obtenido exitosamente para llave %s", timestamp, key)
	return configResp.Value, nil
}

// GetCachedConfig gets a configuration value with caching
// It first checks the cache, then fetches from worker if not found or if test fails
// key: the worker key to fetch
// testFunc: optional function to test if the cached value is still valid (returns error if invalid)
func GetCachedConfig(key string, testFunc func(string) error) (string, error) {
	// timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	// log.Printf("[DEBUG %s] Obteniendo configuración con caché para llave: %s", timestamp, key)

	// Load cache
	cache, err := loadCache()
	if err != nil {
		// log.Printf("[DEBUG %s] Error cargando caché: %v, continuando sin caché", timestamp, err)
		cache = &ConfigCache{APIURLs: make(map[string]string)}
	}

	// Try cached value first
	if cachedValue, exists := cache.APIURLs[key]; exists && cachedValue != "" {
		// log.Printf("[DEBUG %s] Valor encontrado en caché: %s", timestamp, cachedValue)

	// Test the cached value if test function provided
	if testFunc != nil {
		if err := testFunc(cachedValue); err != nil {
			// log.Printf("[DEBUG %s] Valor en caché falló la prueba: %v, obteniendo nuevo desde worker", timestamp, err)
			// Value failed, remove from cache and fetch new one
			delete(cache.APIURLs, key)
			saveCache(cache) // Save updated cache
		} else {
			// log.Printf("[DEBUG %s] Valor en caché válido, usando: %s", timestamp, cachedValue)
			return cachedValue, nil
		}
	} else {
		// No test function, use cached value directly
		// log.Printf("[DEBUG %s] Usando valor en caché (sin prueba): %s", timestamp, cachedValue)
		return cachedValue, nil
	}
	}

	// log.Printf("[DEBUG %s] No hay valor en caché o falló, obteniendo desde worker", timestamp)

	// Fetch from worker
	value, err := GetConfig(key)
	if err != nil {
		return "", err
	}

	// Test the value if test function provided
	if testFunc != nil {
		if err := testFunc(value); err != nil {
			// log.Printf("[DEBUG %s] Valor del worker falló la prueba: %v", timestamp, err)
			return "", fmt.Errorf("valor obtenido del worker no es válido: %v", err)
		}
	}

	// Save to cache
	cache.APIURLs[key] = value
	if err := saveCache(cache); err != nil {
		// log.Printf("[DEBUG %s] Error guardando valor en caché: %v", timestamp, err)
	} else {
		// log.Printf("[DEBUG %s] Valor guardado en caché exitosamente", timestamp)
	}

	return value, nil
}

