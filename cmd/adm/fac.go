package adm

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
)

// Structs para los comprobantes fiscales
type NCF struct {
	ID     int    `json:"id,omitempty"`
	Numero string `json:"numero"`
	Fecha  string `json:"fecha"`
}

type NCFC struct {
	ID     int    `json:"id,omitempty"`
	Numero string `json:"numero"`
	Fecha  string `json:"fecha"`
}

type NCG struct {
	ID     int    `json:"id,omitempty"`
	Numero string `json:"numero"`
	Fecha  string `json:"fecha"`
}

type NCRE struct {
	ID     int    `json:"id,omitempty"`
	Numero string `json:"numero"`
	Fecha  string `json:"fecha"`
}

// Struct para la factura
type Factura struct {
	ID                int     `json:"id,omitempty"`
	IDCotizacion      int     `json:"id_cotizacion"`
	IDCliente         int     `json:"id_cliente"`
	IDProyecto        int     `json:"id_proyecto"`
	Moneda            string  `json:"moneda"`
	TipoFactura       string  `json:"tipo_factura"`
	Fecha             string  `json:"fecha"`
	TasaMoneda        float64 `json:"tasa_moneda"`
	Original          string  `json:"original"`
	Estado            string  `json:"estado"`
	Idioma            string  `json:"idioma"`
	Comprobante       string  `json:"comprobante"`
	ComprobanteValido string  `json:"comprobante_valido"`
	Subtotal          float64 `json:"subtotal"`
	Indirectos        float64 `json:"indirectos"`
	DescuentoP        float64 `json:"descuentop"`
	DescuentoM        float64 `json:"descuentom"`
	RetencionP        float64 `json:"retencionp"`
	RetencionM        float64 `json:"retencionm"`
	ItbisP            float64 `json:"itbisp"`
	ItbisM            float64 `json:"itbism"`
	TotalSinItbis     float64 `json:"total_sin_itbis"`
	Total             float64 `json:"total"`
}

// Struct para el presupuesto facturado
type PresupuestoFacturado struct {
	ID          int                    `json:"id,omitempty"`
	IDFactura   int                    `json:"id_factura"`
	Presupuesto map[string]interface{} `json:"presupuesto"`
}

var facCmd = &cobra.Command{
	Use:   "invoice",
	Short: "Crear una nueva factura desde cotización",
	Long:  `Crear una nueva factura basada en una cotización existente`,
	Run: func(cmd *cobra.Command, args []string) {
		crearFacturaDesdeCorización()
	},
}

func init() {
	NewCmd.AddCommand(facCmd)
}

func crearFacturaDesdeCorización() {
	fmt.Println("\n=== CREAR FACTURA DESDE COTIZACIÓN ===")

	// Paso 1: Buscar cotización
	cotizacion := BuscarCotizacion()
	if cotizacion.ID == 0 {
		fmt.Println("No se seleccionó ninguna cotización")
		return
	}

	// Mostrar información de la cotización seleccionada
	fmt.Printf("\n--- Cotización Seleccionada ---\n")
	fmt.Printf("ID: %d\n", cotizacion.ID)
	fmt.Printf("Fecha: %s\n", cotizacion.Fecha)
	fmt.Printf("Total: %.2f %s\n", cotizacion.Total, cotizacion.Moneda)
	fmt.Printf("Estado: %s\n", cotizacion.Estado)

	// Obtener presupuesto de la cotización
	presupuesto, err := obtenerPresupuestoCotizacion(cotizacion.ID)
	if err != nil {
		fmt.Printf("Error al obtener presupuesto: %v\n", err)
		return
	}

	// Paso 2: Solicitar total de la factura
	totalFacturaStr := inputs.GetInputWithDefault(
		fmt.Sprintf("Ingrese el total de la factura (por defecto: %.2f)", cotizacion.Total),
		fmt.Sprintf("%.2f", cotizacion.Total),
	)

	totalFactura, err := strconv.ParseFloat(totalFacturaStr, 64)
	if err != nil {
		fmt.Println("Total inválido")
		return
	}

	// Paso 3: Crear la factura
	factura, err := crearFactura(cotizacion, totalFactura)
	if err != nil {
		fmt.Printf("Error al crear factura: %v\n", err)
		return
	}

	// Paso 4: Guardar el presupuesto facturado
	err = guardarPresupuestoFacturado(factura.ID, cotizacion.ID, presupuesto.Presupuesto)
	if err != nil {
		fmt.Printf("Error al guardar presupuesto facturado: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Factura creada exitosamente con ID: %d\n", factura.ID)
	fmt.Printf("Total facturado: %.2f %s\n", factura.Total, factura.Moneda)
}

func obtenerPresupuestoCotizacion(idCotizacion int) (*Presupuesto, error) {
	postgrestURL, headers := InitializePostgrest()
	url := fmt.Sprintf("%s/presupuesto?id_cotizacion=eq.%d", postgrestURL, idCotizacion)

	resp, err := MakeRequest("GET", url, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("error al buscar presupuesto: %w", err)
	}

	var presupuestos []Presupuesto
	if err := json.Unmarshal(resp, &presupuestos); err != nil {
		return nil, fmt.Errorf("error al procesar presupuesto: %w", err)
	}

	if len(presupuestos) == 0 {
		return &Presupuesto{
			IDCotizacion: idCotizacion,
			Presupuesto:  make(map[string]interface{}),
		}, nil
	}

	return &presupuestos[0], nil
}

func crearFactura(cotizacion Cotizacion, totalFactura float64) (*Factura, error) {
	// Obtener datos del cliente para determinar tipo de factura
	cliente, err := obtenerCliente(cotizacion.IDCliente)
	if err != nil {
		return nil, fmt.Errorf("error al obtener cliente: %w", err)
	}

	// Calcular proporción para ajustar todos los valores
	proporcion := totalFactura / cotizacion.Total

	// Obtener comprobante disponible
	comprobante, comprobanteValido, err := obtenerComprobanteDisponible(cliente.TipoFactura)
	if err != nil {
		fmt.Printf("Advertencia: No se pudo obtener comprobante: %v\n", err)
		comprobante = ""
		comprobanteValido = ""
	}

	// Crear la factura con valores proporcionales
	factura := &Factura{
		IDCotizacion:      cotizacion.ID,
		IDCliente:         cotizacion.IDCliente,
		IDProyecto:        cotizacion.IDProyecto,
		Moneda:            cotizacion.Moneda,
		TipoFactura:       cliente.TipoFactura,
		Fecha:             time.Now().Format("02/01/2006"), // Sí, está en formato día/mes/año
		TasaMoneda:        cotizacion.TasaMoneda,
		Original:          "VENDEDOR",
		Estado:            "GENERADA",
		Idioma:            cotizacion.Idioma,
		Comprobante:       comprobante,
		ComprobanteValido: comprobanteValido,
		Subtotal:          cotizacion.Subtotal * proporcion,
		Indirectos:        cotizacion.Indirectos * proporcion,
		DescuentoP:        cotizacion.DescuentoP,
		DescuentoM:        cotizacion.DescuentoM * proporcion,
		RetencionP:        cotizacion.RetencionP,
		RetencionM:        cotizacion.RetencionM * proporcion,
		ItbisP:            cotizacion.ItbisP,
		ItbisM:            cotizacion.ItbisM * proporcion,
		TotalSinItbis:     (cotizacion.Total - cotizacion.ItbisM) * proporcion,
		Total:             totalFactura,
	}

	// Guardar la factura
	return guardarFactura(factura)
}

func obtenerCliente(idCliente int) (*Cliente, error) {
	postgrestURL, headers := InitializePostgrest()
	url := fmt.Sprintf("%s/cliente?id=eq.%d", postgrestURL, idCliente)

	resp, err := MakeRequest("GET", url, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("error al buscar cliente: %w", err)
	}

	var clientes []Cliente
	if err := json.Unmarshal(resp, &clientes); err != nil {
		return nil, fmt.Errorf("error al procesar cliente: %w", err)
	}

	if len(clientes) == 0 {
		return nil, fmt.Errorf("no se encontró cliente con ID %d", idCliente)
	}

	return &clientes[0], nil
}

func obtenerComprobanteDisponible(tipoFactura string) (string, string, error) {
	// Mapear el tipo de factura a la tabla correspondiente
	var tabla string
	switch tipoFactura {
	case "NCFC":
		tabla = "ncfc"
	case "NCF":
		tabla = "ncf"
	case "NCG":
		tabla = "ncg"
	case "NCRE":
		tabla = "ncre"
	default:
		return "", "", fmt.Errorf("tipo de factura no válido: %s", tipoFactura)
	}

	postgrestURL, headers := InitializePostgrest()

	// Buscar comprobantes no usados y válidos
	url := fmt.Sprintf("%s/%s?select=numero,fecha&order=numero", postgrestURL, tabla)

	resp, err := MakeRequest("GET", url, headers, nil)
	if err != nil {
		return "", "", fmt.Errorf("error al buscar comprobantes: %w", err)
	}

	var comprobantes []map[string]interface{}
	if err := json.Unmarshal(resp, &comprobantes); err != nil {
		return "", "", fmt.Errorf("error al procesar comprobantes: %w", err)
	}

	fechaActual := time.Now()

	for _, comp := range comprobantes {
		numero, ok1 := comp["numero"].(string)
		fechaStr, ok2 := comp["fecha"].(string)

		if !ok1 || !ok2 {
			continue
		}

		// Verificar si el comprobante no ha sido usado en otra factura
		if !esComprobanteUsado(numero) {
			// Intentar parsear la fecha con múltiples formatos
			fechaValida, err := parsearFecha(fechaStr)
			if err != nil {
				fmt.Printf("Advertencia: No se pudo parsear fecha '%s' para comprobante %s: %v\n", fechaStr, numero, err)
				continue
			}

			if fechaActual.Before(fechaValida) || fechaActual.Equal(fechaValida) {
				return numero, fechaStr, nil
			}
		}
	}

	return "", "", fmt.Errorf("no se encontraron comprobantes disponibles para tipo %s", tipoFactura)
}

// Función auxiliar para parsear fechas en múltiples formatos
func parsearFecha(fechaStr string) (time.Time, error) {
	// Formato de fecha esperado: mes/dia/año
	formato := "01/02/2006" // mm/dd/yyyy

	fecha, err := time.Parse(formato, fechaStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("la fecha '%s' no está en el formato correcto (mes/dia/año)", fechaStr)
	}

	return fecha, nil
}

func esComprobanteUsado(numero string) bool {
	postgrestURL, headers := InitializePostgrest()
	url := fmt.Sprintf("%s/factura?comprobante=eq.%s", postgrestURL, numero)

	resp, err := MakeRequest("GET", url, headers, nil)
	if err != nil {
		return true // En caso de error, asumir que está usado
	}

	var facturas []Factura
	if err := json.Unmarshal(resp, &facturas); err != nil {
		return true // En caso de error, asumir que está usado
	}

	return len(facturas) > 0
}

func guardarFactura(factura *Factura) (*Factura, error) {
	// Obtener próximo ID
	nextID, err := obtenerProximoIDFactura()
	if err != nil {
		return nil, fmt.Errorf("error al obtener próximo ID: %w", err)
	}

	factura.ID = nextID

	postgrestURL, headers := InitializePostgrest()
	url := postgrestURL + "/factura"

	jsonData, err := json.Marshal(factura)
	if err != nil {
		return nil, fmt.Errorf("error al serializar factura: %w", err)
	}

	resp, err := MakeRequest("POST", url, headers, jsonData)
	if err != nil {
		return nil, fmt.Errorf("error al guardar factura: %w", err)
	}

	var facturasResp []Factura
	if err := json.Unmarshal(resp, &facturasResp); err != nil {
		return nil, fmt.Errorf("error al procesar respuesta: %w", err)
	}

	if len(facturasResp) == 0 {
		return nil, fmt.Errorf("no se recibió respuesta de la factura guardada")
	}

	return &facturasResp[0], nil
}

func obtenerProximoIDFactura() (int, error) {
	postgrestURL, headers := InitializePostgrest()
	url := postgrestURL + "/factura?select=id&order=id.desc&limit=1"

	resp, err := MakeRequest("GET", url, headers, nil)
	if err != nil {
		return 0, fmt.Errorf("error al buscar último ID: %w", err)
	}

	var facturas []Factura
	if err := json.Unmarshal(resp, &facturas); err != nil {
		return 0, fmt.Errorf("error al procesar respuesta: %w", err)
	}

	if len(facturas) == 0 {
		return 1, nil
	}

	return facturas[0].ID + 1, nil
}

func guardarPresupuestoFacturado(idFactura int, idCotizacion int, presupuesto map[string]interface{}) error {
	presupuestoFacturado := &PresupuestoFacturado{
		ID:          idCotizacion,
		IDFactura:   idFactura,
		Presupuesto: presupuesto,
	}

	postgrestURL, headers := InitializePostgrest()
	url := postgrestURL + "/presupuestofacturado"

	jsonData, err := json.Marshal(presupuestoFacturado)
	if err != nil {
		return fmt.Errorf("error al serializar presupuesto facturado: %w", err)
	}

	_, err = MakeRequest("POST", url, headers, jsonData)
	if err != nil {
		return fmt.Errorf("error al guardar presupuesto facturado: %w", err)
	}

	return nil
}
