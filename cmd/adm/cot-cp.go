package adm

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
)

// Comando para copiar cotización
var copyCotCmd = &cobra.Command{
	Use:   "quotation-from",
	Short: "Copiar una cotización existente",
	Long:  `Copiar una cotización existente con un cliente y/o proyecto diferente`,
	Run: func(cmd *cobra.Command, args []string) {
		copiarCotizacion()
	},
}

// Estructura para cotización completa con relaciones
type CotizacionCompleta struct {
	Cotizacion  Cotizacion  `json:"cotizacion"`
	Presupuesto Presupuesto `json:"presupuesto"`
	Cliente     Cliente     `json:"cliente"`
	Proyecto    Proyecto    `json:"proyecto"`
	Servicio    Servicio    `json:"servicio"`
}

func init() {
	NewCmd.AddCommand(copyCotCmd)
}

// Buscar cotizaciones por diferentes criterios
func BuscarCotizacion() Cotizacion {
	var cotizacion Cotizacion

	// Método de búsqueda
	items := []inputs.Item{
		{ID: "id", Name: "Buscar por ID", Desc: "Buscar directamente por el ID de la cotización"},
		{ID: "cliente", Name: "Buscar por cliente", Desc: "Buscar cotizaciones de un cliente específico"},
		{ID: "todas", Name: "Ver todas las cotizaciones", Desc: "Mostrar todas las cotizaciones disponibles"},
	}

	selectedMethod := inputs.SelectList("Seleccione método de búsqueda de cotización", items)

	switch selectedMethod.ID {
	case "id":
		return buscarCotizacionPorID()
	case "cliente":
		return buscarCotizacionPorCliente()
	case "todas":
		return buscarTodasLasCotizaciones()
	}

	return cotizacion
}

// Buscar cotización por ID
func buscarCotizacionPorID() Cotizacion {
	var cotizacion Cotizacion

	idStr := inputs.GetInput("Ingrese el ID de la cotización")
	_, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("ID inválido")
		return cotizacion
	}

	postgrestURL, headers := InitializePostgrest()
	url := postgrestURL + "/cotizacion?id=eq." + idStr

	resp, err := MakeRequest("GET", url, headers, nil)
	if err != nil {
		fmt.Println("Error al buscar cotización:", err)
		return cotizacion
	}

	var cotizaciones []Cotizacion
	if err := json.Unmarshal(resp, &cotizaciones); err != nil {
		fmt.Println("Error al procesar respuesta:", err)
		return cotizacion
	}

	if len(cotizaciones) > 0 {
		return cotizaciones[0]
	}

	fmt.Println("No se encontró cotización con ID:", idStr)
	return cotizacion
}

// Buscar cotización por cliente
func buscarCotizacionPorCliente() Cotizacion {
	var cotizacion Cotizacion

	// Buscar cliente
	cliente := BuscarCliente()
	if cliente.ID == 0 {
		fmt.Println("No se seleccionó ningún cliente")
		return cotizacion
	}

	// Buscar cotizaciones del cliente
	postgrestURL, headers := InitializePostgrest()
	url := fmt.Sprintf("%s/cotizacion?select=*&id_cliente=eq.%d&order=id.desc", postgrestURL, cliente.ID)

	resp, err := MakeRequest("GET", url, headers, nil)
	if err != nil {
		fmt.Println("Error al buscar cotizaciones:", err)
		return cotizacion
	}

	var cotizaciones []Cotizacion
	if err := json.Unmarshal(resp, &cotizaciones); err != nil {
		fmt.Println("Error al procesar cotizaciones:", err)
		return cotizacion
	}

	if len(cotizaciones) == 0 {
		fmt.Println("No se encontraron cotizaciones para este cliente")
		return cotizacion
	}

	// Obtener nombres de proyectos para todas las cotizaciones
	proyectos := make(map[int]string)
	for _, c := range cotizaciones {
		if _, exists := proyectos[c.IDProyecto]; !exists {
			proyectoURL := fmt.Sprintf("%s/proyecto?select=nombre_proyecto&id=eq.%d", postgrestURL, c.IDProyecto)
			respProyecto, err := MakeRequest("GET", proyectoURL, headers, nil)
			if err == nil {
				var proyectosResp []map[string]interface{}
				if err := json.Unmarshal(respProyecto, &proyectosResp); err == nil && len(proyectosResp) > 0 {
					if nombre, ok := proyectosResp[0]["nombre_proyecto"].(string); ok {
						proyectos[c.IDProyecto] = nombre
					} else {
						proyectos[c.IDProyecto] = "Proyecto desconocido"
					}
				} else {
					proyectos[c.IDProyecto] = "Proyecto desconocido"
				}
			} else {
				proyectos[c.IDProyecto] = "Proyecto desconocido"
			}
		}
	}

	// Crear items para selección con información del proyecto
	items := make([]inputs.Item, len(cotizaciones))
	for i, c := range cotizaciones {
		nombreProyecto := proyectos[c.IDProyecto]
		items[i] = inputs.Item{
			ID:   strconv.Itoa(c.ID),
			Name: fmt.Sprintf("Cotización #%d - %s", c.ID, c.Fecha),
			Desc: fmt.Sprintf("Proyecto: %s | Estado: %s | Total: %.2f", nombreProyecto, c.Estado, c.Total),
		}
	}

	selectedItem := inputs.SelectList("Seleccione una cotización", items)
	id, _ := strconv.Atoi(selectedItem.ID)

	for _, c := range cotizaciones {
		if c.ID == id {
			return c
		}
	}

	return cotizacion
}

// Buscar en todas las cotizaciones
func buscarTodasLasCotizaciones() Cotizacion {
	var cotizacion Cotizacion

	postgrestURL, headers := InitializePostgrest()
	url := postgrestURL + "/cotizacion?select=*&order=id.desc&limit=50"

	resp, err := MakeRequest("GET", url, headers, nil)
	if err != nil {
		fmt.Println("Error al buscar cotizaciones:", err)
		return cotizacion
	}

	var cotizaciones []Cotizacion
	if err := json.Unmarshal(resp, &cotizaciones); err != nil {
		fmt.Println("Error al procesar cotizaciones:", err)
		return cotizacion
	}

	if len(cotizaciones) == 0 {
		fmt.Println("No se encontraron cotizaciones")
		return cotizacion
	}

	// Obtener nombres de proyectos para todas las cotizaciones
	proyectos := make(map[int]string)
	for _, c := range cotizaciones {
		if _, exists := proyectos[c.IDProyecto]; !exists {
			proyectoURL := fmt.Sprintf("%s/proyecto?select=nombre_proyecto&id=eq.%d", postgrestURL, c.IDProyecto)
			respProyecto, err := MakeRequest("GET", proyectoURL, headers, nil)
			if err == nil {
				var proyectosResp []map[string]interface{}
				if err := json.Unmarshal(respProyecto, &proyectosResp); err == nil && len(proyectosResp) > 0 {
					if nombre, ok := proyectosResp[0]["nombre_proyecto"].(string); ok {
						proyectos[c.IDProyecto] = nombre
					} else {
						proyectos[c.IDProyecto] = "Proyecto desconocido"
					}
				} else {
					proyectos[c.IDProyecto] = "Proyecto desconocido"
				}
			} else {
				proyectos[c.IDProyecto] = "Proyecto desconocido"
			}
		}
	}

	// Crear items para selección con información del proyecto
	items := make([]inputs.Item, len(cotizaciones))
	for i, c := range cotizaciones {
		nombreProyecto := proyectos[c.IDProyecto]
		items[i] = inputs.Item{
			ID:   strconv.Itoa(c.ID),
			Name: fmt.Sprintf("Cotización #%d - %s", c.ID, c.Fecha),
			Desc: fmt.Sprintf("Proyecto: %s | Estado: %s | Total: %.2f", nombreProyecto, c.Estado, c.Total),
		}
	}

	selectedItem := inputs.SelectList("Seleccione una cotización", items)
	id, _ := strconv.Atoi(selectedItem.ID)

	for _, c := range cotizaciones {
		if c.ID == id {
			return c
		}
	}

	return cotizacion
}

// Obtener cotización completa con todos sus datos relacionados
func ObtenerCotizacionCompleta(idCotizacion int) (*CotizacionCompleta, error) {
	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		return nil, fmt.Errorf("URL de PostgREST no configurada")
	}

	// Obtener cotización
	cotizacionURL := fmt.Sprintf("%s/cotizacion?id=eq.%d", postgrestURL, idCotizacion)
	resp, err := MakeRequest("GET", cotizacionURL, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("error al obtener cotización: %w", err)
	}

	var cotizaciones []Cotizacion
	if err := json.Unmarshal(resp, &cotizaciones); err != nil {
		return nil, fmt.Errorf("error al procesar cotización: %w", err)
	}

	if len(cotizaciones) == 0 {
		return nil, fmt.Errorf("no se encontró cotización con ID %d", idCotizacion)
	}

	cotizacion := cotizaciones[0]

	// Obtener presupuesto
	presupuestoURL := fmt.Sprintf("%s/presupuesto?id_cotizacion=eq.%d", postgrestURL, idCotizacion)
	resp, err = MakeRequest("GET", presupuestoURL, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("error al obtener presupuesto: %w", err)
	}

	var presupuestos []Presupuesto
	if err := json.Unmarshal(resp, &presupuestos); err != nil {
		return nil, fmt.Errorf("error al procesar presupuesto: %w", err)
	}

	var presupuesto Presupuesto
	if len(presupuestos) > 0 {
		presupuesto = presupuestos[0]
	}

	// Obtener cliente
	clienteURL := fmt.Sprintf("%s/cliente?id=eq.%d", postgrestURL, cotizacion.IDCliente)
	resp, err = MakeRequest("GET", clienteURL, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("error al obtener cliente: %w", err)
	}

	var clientes []Cliente
	if err := json.Unmarshal(resp, &clientes); err != nil {
		return nil, fmt.Errorf("error al procesar cliente: %w", err)
	}

	var cliente Cliente
	if len(clientes) > 0 {
		cliente = clientes[0]
	}

	// Obtener proyecto
	proyectoURL := fmt.Sprintf("%s/proyecto?id=eq.%d", postgrestURL, cotizacion.IDProyecto)
	resp, err = MakeRequest("GET", proyectoURL, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("error al obtener proyecto: %w", err)
	}

	var proyectos []Proyecto
	if err := json.Unmarshal(resp, &proyectos); err != nil {
		return nil, fmt.Errorf("error al procesar proyecto: %w", err)
	}

	var proyecto Proyecto
	if len(proyectos) > 0 {
		proyecto = proyectos[0]
	}

	// Obtener servicio
	servicioURL := fmt.Sprintf("%s/servicio?id=eq.%d", postgrestURL, cotizacion.IDServicio)
	resp, err = MakeRequest("GET", servicioURL, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("error al obtener servicio: %w", err)
	}

	var servicios []Servicio
	if err := json.Unmarshal(resp, &servicios); err != nil {
		return nil, fmt.Errorf("error al procesar servicio: %w", err)
	}

	var servicio Servicio
	if len(servicios) > 0 {
		servicio = servicios[0]
	}

	return &CotizacionCompleta{
		Cotizacion:  cotizacion,
		Presupuesto: presupuesto,
		Cliente:     cliente,
		Proyecto:    proyecto,
		Servicio:    servicio,
	}, nil
}

// Función principal para copiar cotización
func copiarCotizacion() {
	fmt.Println("=== COPIAR COTIZACIÓN ===")

	// Paso 1: Seleccionar cotización existente
	cotizacionOriginal := BuscarCotizacion()
	if cotizacionOriginal.ID == 0 {
		fmt.Println("No se seleccionó ninguna cotización")
		return
	}

	// Paso 2: Obtener datos completos de la cotización
	cotizacionCompleta, err := ObtenerCotizacionCompleta(cotizacionOriginal.ID)
	if err != nil {
		fmt.Println("Error al obtener datos completos de la cotización:", err)
		return
	}

	fmt.Printf("Cotización original: #%d - %s\n", cotizacionCompleta.Cotizacion.ID, cotizacionCompleta.Cliente.Nombre)
	fmt.Printf("Proyecto: %s\n", cotizacionCompleta.Proyecto.Nombre)
	fmt.Printf("Servicio: %s\n", cotizacionCompleta.Servicio.Nombre)
	fmt.Printf("Total: %.2f %s\n\n", cotizacionCompleta.Cotizacion.Total, cotizacionCompleta.Cotizacion.Moneda)

	// Paso 3: Preguntar qué cambiar
	items := []inputs.Item{
		{ID: "cliente", Name: "Cambiar solo cliente", Desc: "Mantener el mismo proyecto y servicio"},
		{ID: "proyecto", Name: "Cambiar solo proyecto", Desc: "Mantener el mismo cliente y servicio"},
		{ID: "ambos", Name: "Cambiar cliente y proyecto", Desc: "Mantener solo el servicio"},
		{ID: "todo", Name: "Cambiar cliente, proyecto y servicio", Desc: "Cambiar todos los elementos"},
		{ID: "ninguno", Name: "Mantener todo igual", Desc: "Solo crear una copia exacta"},
	}

	seleccion := inputs.SelectList("¿Qué desea cambiar en la nueva cotización?", items)

	// Paso 4: Obtener nuevos datos según la selección
	nuevoCliente := cotizacionCompleta.Cliente
	nuevoProyecto := cotizacionCompleta.Proyecto
	nuevoServicio := cotizacionCompleta.Servicio

	switch seleccion.ID {
	case "cliente":
		nuevoCliente = BuscarCliente()
		if nuevoCliente.ID == 0 {
			fmt.Println("No se seleccionó ningún cliente")
			return
		}
	case "proyecto":
		nuevoProyecto = BuscarProyecto()
		if nuevoProyecto.ID == 0 {
			fmt.Println("No se seleccionó ningún proyecto")
			return
		}
	case "ambos":
		nuevoCliente = BuscarCliente()
		if nuevoCliente.ID == 0 {
			fmt.Println("No se seleccionó ningún cliente")
			return
		}
		nuevoProyecto = BuscarProyecto()
		if nuevoProyecto.ID == 0 {
			fmt.Println("No se seleccionó ningún proyecto")
			return
		}
	case "todo":
		nuevoCliente = BuscarCliente()
		if nuevoCliente.ID == 0 {
			fmt.Println("No se seleccionó ningún cliente")
			return
		}
		nuevoProyecto = BuscarProyecto()
		if nuevoProyecto.ID == 0 {
			fmt.Println("No se seleccionó ningún proyecto")
			return
		}
		nuevoServicio = BuscarServicio()
		if nuevoServicio.ID == 0 {
			fmt.Println("No se seleccionó ningún servicio")
			return
		}
	}

	// Paso 5: Crear la nueva cotización
	nuevaCotizacion := cotizacionCompleta.Cotizacion
	nuevaCotizacion.ID = 0 // Reset ID para que se genere uno nuevo
	nuevaCotizacion.IDCliente = nuevoCliente.ID
	nuevaCotizacion.IDProyecto = nuevoProyecto.ID
	nuevaCotizacion.IDServicio = nuevoServicio.ID
	nuevaCotizacion.Fecha = time.Now().Format("01/02/2006")
	nuevaCotizacion.Estado = "GENERADA" // Reset estado

	// Paso 6: Mostrar resumen antes de guardar
	fmt.Println("\n=== RESUMEN DE LA NUEVA COTIZACIÓN ===")
	fmt.Printf("Cliente: %s\n", nuevoCliente.Nombre)
	fmt.Printf("Proyecto: %s\n", nuevoProyecto.Nombre)
	fmt.Printf("Servicio: %s\n", nuevoServicio.Nombre)
	fmt.Printf("Total: %.2f %s\n", nuevaCotizacion.Total, nuevaCotizacion.Moneda)

	confirmarItems := []inputs.Item{
		{ID: "si", Name: "Sí", Desc: "Crear la nueva cotización"},
		{ID: "no", Name: "No", Desc: "Cancelar la operación"},
	}

	confirmar := inputs.SelectList("¿Confirma que desea crear esta cotización?", confirmarItems)

	if confirmar.ID == "no" {
		fmt.Println("Operación cancelada")
		return
	}

	// Paso 7: Guardar la nueva cotización
	idNuevaCotizacion := guardarCotizacion(nuevaCotizacion)
	if idNuevaCotizacion == 0 {
		fmt.Println("Error al guardar la nueva cotización")
		return
	}

	// Paso 8: Copiar el presupuesto si existe
	if cotizacionCompleta.Presupuesto.IDCotizacion > 0 {
		nuevoPresupuesto := cotizacionCompleta.Presupuesto
		nuevoPresupuesto.ID = 0 // Reset ID
		nuevoPresupuesto.IDCotizacion = idNuevaCotizacion

		success := GuardarPresupuestoConID(nuevoPresupuesto, idNuevaCotizacion)
		if !success {
			fmt.Printf("Advertencia: La cotización se creó (ID: %d) pero hubo un error al copiar el presupuesto\n", idNuevaCotizacion)
		} else {
			fmt.Printf("✓ Cotización copiada exitosamente con ID: %d\n", idNuevaCotizacion)
			fmt.Printf("✓ Presupuesto copiado correctamente\n")
		}
	} else {
		fmt.Printf("✓ Cotización copiada exitosamente con ID: %d\n", idNuevaCotizacion)
		fmt.Printf("⚠ No había presupuesto en la cotización original\n")
	}

	// Opción para imprimir la nueva cotización
	imprimirItems := []inputs.Item{
		{ID: "si", Name: "Sí", Desc: "Generar el documento de la cotización"},
		{ID: "no", Name: "No", Desc: "Solo guardar sin imprimir"},
	}

	imprimir := inputs.SelectList("¿Desea imprimir la nueva cotización?", imprimirItems)

	if imprimir.ID == "si" {
		filePath, err := GetCotizacion(idNuevaCotizacion)
		if err != nil {
			fmt.Println("Error al generar el documento:", err)
		} else {
			fmt.Printf("Documento generado: %s\n", filePath)
		}
	}
}
