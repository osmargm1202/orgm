package adm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var PdfCmd = &cobra.Command{
	Use:   "pdf",
	Short: "Print documents",
	Long:  `Print documents from the API`,
}

var PrintFacturaCmd = &cobra.Command{
	Use:   "invoice [id...]",
	Short: "Print one or more Invoices",
	Long:  `Print one or more Invoices from the API`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Error: At least one Invoice ID is required")
			return
		}

		for _, arg := range args {
			id, err := strconv.Atoi(arg)
			if err != nil {
				fmt.Printf("Error: Invoice ID '%s' must be an integer\n", arg)
				continue
			}

			filePath, err := GetFactura(id)
			if err != nil {
				fmt.Printf("Error processing invoice %d: %v\n", id, err)
				continue
			}

			fmt.Printf("Invoice %d saved as %s\n", id, filePath)
		}
	},
}

var PrintCotizacionCmd = &cobra.Command{
	Use:   "quotation [id...]",
	Short: "Print one or more Quotations",
	Long:  `Print one or more Quotations from the API`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Error: At least one Quotation ID is required")
			return
		}

		for _, arg := range args {
			id, err := strconv.Atoi(arg)
			if err != nil {
				fmt.Printf("Error: Quotation ID '%s' must be an integer\n", arg)
				continue
			}

			filePath, err := GetCotizacion(id)
			if err != nil {
				fmt.Printf("Error processing quotation %d: %v\n", id, err)
				continue
			}

			fmt.Printf("Quotation %d saved as %s\n", id, filePath)
		}
	},
}

func init() {
	AdmCmd.AddCommand(PdfCmd)
	PdfCmd.AddCommand(PrintFacturaCmd)
	PdfCmd.AddCommand(PrintCotizacionCmd)
}

// Estructura para el schema.json
type Schema struct {
	Tipos map[string]struct {
		Carpetas []string `json:"carpetas"`
	} `json:"tipos"`
}

// Función para crear estructura de carpetas basada en fecha
func crearEstructuraCarpetas(tipoDocumento string, fecha time.Time) (string, error) {
	// Obtener la ruta base de configuración
	configPath := viper.GetString("config_path")
	if configPath == "" {
		return "", fmt.Errorf("la variable 'config_path' no está configurada en viper")
	}

	// Obtener la ruta base de administración
	rutaAdmin := viper.GetString("carpetas.administracion")
	if rutaAdmin == "" {
		return "", fmt.Errorf("la variable 'carpetas.administracion' no está configurada en viper")
	}

	// Expandir ~ a homedir si está presente
	if len(rutaAdmin) > 0 && rutaAdmin[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("error obteniendo directorio home: %w", err)
		}
		rutaAdmin = filepath.Join(home, rutaAdmin[1:])
	}

	// Crear estructura: [rutaBase]/[año]/[mes]/
	ano := fecha.Format("2006")
	mes := fecha.Format("01")
	rutaBaseTemporal := filepath.Join(rutaAdmin, ano, mes)

	// Leer schema.json para obtener las carpetas de administración
	schemaPath := filepath.Join(configPath, "folder", "schema.json")
	schemaFile, err := os.ReadFile(schemaPath)
	if err != nil {
		return "", fmt.Errorf("error leyendo schema.json desde %s: %w", schemaPath, err)
	}

	var schema Schema
	if err := json.Unmarshal(schemaFile, &schema); err != nil {
		return "", fmt.Errorf("error parseando schema.json: %w", err)
	}

	// Obtener las carpetas de administración
	adminTipo, exists := schema.Tipos["Administracion"]
	if !exists {
		return "", fmt.Errorf("no se encontró el tipo 'Administracion' en schema.json")
	}

	// Crear todas las carpetas de administración
	for _, carpeta := range adminTipo.Carpetas {
		rutaCompleta := filepath.Join(rutaBaseTemporal, carpeta)
		err := os.MkdirAll(rutaCompleta, 0755)
		if err != nil {
			return "", fmt.Errorf("error creando carpeta %s: %w", rutaCompleta, err)
		}
	}

	// Determinar la carpeta específica donde guardar el documento
	var carpetaDestino string
	if tipoDocumento == "factura" {
		carpetaDestino = "Ventas"
	} else if tipoDocumento == "cotizacion" {
		carpetaDestino = "Cotizaciones"
	} else {
		return "", fmt.Errorf("tipo de documento no válido: %s", tipoDocumento)
	}

	rutaFinal := filepath.Join(rutaBaseTemporal, carpetaDestino)
	return rutaFinal, nil
}

// Función auxiliar para parsear fechas
func parsearFechaDocumento(fechaStr string) (time.Time, error) {
	// Usar el mismo formato que en fac.go: mes/día/año
	formato := "01/02/2006" // mm/dd/yyyy

	fecha, err := time.Parse(formato, fechaStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("la fecha '%s' no está en el formato correcto (mes/dia/año)", fechaStr)
	}

	return fecha, nil
}

func GetCotizacion(id int) (string, error) {
	// Primero obtener los datos de la cotización para obtener la fecha
	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		return "", fmt.Errorf("error initializing PostgREST URL")
	}

	cotizacionURL := fmt.Sprintf("%s/cotizacion?id=eq.%d", postgrestURL, id)
	respData, err := MakeRequest("GET", cotizacionURL, headers, nil)
	if err != nil {
		return "", fmt.Errorf("error getting quotation data: %v", err)
	}

	var cotizaciones []Cotizacion
	if err := json.Unmarshal(respData, &cotizaciones); err != nil {
		return "", fmt.Errorf("error parsing quotation data: %v", err)
	}

	if len(cotizaciones) == 0 {
		return "", fmt.Errorf("quotation with ID %d not found", id)
	}

	// Parsear la fecha de la cotización
	fechaCotizacion, err := parsearFechaDocumento(cotizaciones[0].Fecha)
	if err != nil {
		return "", fmt.Errorf("error parsing quotation date: %w", err)
	}

	// Ahora descargar el archivo DOCX
	apiURL, apiHeaders := InitializeApi()
	if apiURL == "" {
		return "", fmt.Errorf("error initializing API URL")
	}

	url := fmt.Sprintf("%s/cot/%d", apiURL, id)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	for key, value := range apiHeaders {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error response status: %d", resp.StatusCode)
	}

	// Usar la fecha de la cotización para crear la estructura de carpetas
	rutaCarpeta, err := crearEstructuraCarpetas("cotizacion", fechaCotizacion)
	if err != nil {
		return "", fmt.Errorf("error creando estructura de carpetas: %w", err)
	}

	// Crear el archivo en la ruta final
	fileName := fmt.Sprintf("cotizacion_%d.docx", id)
	filePath := filepath.Join(rutaCarpeta, fileName)

	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("error creating file: %v", err)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		out.Close()
		os.Remove(filePath)
		return "", fmt.Errorf("error writing file: %v", err)
	}
	out.Close()

	return filePath, nil
}

func GetFactura(id int) (string, error) {
	// Primero obtener los datos de la factura para obtener la fecha
	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		return "", fmt.Errorf("error initializing PostgREST URL")
	}

	facturaURL := fmt.Sprintf("%s/factura?id=eq.%d", postgrestURL, id)
	respData, err := MakeRequest("GET", facturaURL, headers, nil)
	if err != nil {
		return "", fmt.Errorf("error getting invoice data: %v", err)
	}

	var facturas []Factura
	if err := json.Unmarshal(respData, &facturas); err != nil {
		return "", fmt.Errorf("error parsing invoice data: %v", err)
	}

	if len(facturas) == 0 {
		return "", fmt.Errorf("invoice with ID %d not found", id)
	}

	// Parsear la fecha de la factura
	fechaFactura, err := parsearFechaDocumento(facturas[0].Fecha)
	if err != nil {
		return "", fmt.Errorf("error parsing invoice date: %w", err)
	}

	// Ahora descargar el archivo DOCX
	apiURL, apiHeaders := InitializeApi()
	if apiURL == "" {
		return "", fmt.Errorf("error initializing API URL")
	}

	url := fmt.Sprintf("%s/fac/%d", apiURL, id)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	for key, value := range apiHeaders {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error response status: %d", resp.StatusCode)
	}

	// Usar la fecha de la factura para crear la estructura de carpetas
	rutaCarpeta, err := crearEstructuraCarpetas("factura", fechaFactura)
	if err != nil {
		return "", fmt.Errorf("error creando estructura de carpetas: %w", err)
	}

	// Crear el archivo en la ruta final
	fileName := fmt.Sprintf("factura_%d.docx", id)
	filePath := filepath.Join(rutaCarpeta, fileName)

	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("error creating file: %v", err)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		out.Close()
		os.Remove(filePath)
		return "", fmt.Errorf("error writing file: %v", err)
	}
	out.Close()

	return filePath, nil
}
