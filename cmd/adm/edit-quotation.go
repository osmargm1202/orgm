package adm

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var editQuotationCmd = &cobra.Command{
	Use:   "quotation [id]",
	Short: "Modificar una cotización existente",
	Long: `Interfaz gráfica para modificar una cotización existente.

Ejemplos:
  orgm adm edit quotation          - Abrir búsqueda interactiva para seleccionar cotización
  orgm adm edit quotation 123      - Editar directamente la cotización con ID 123

El comando abrirá una interfaz gráfica con Fyne donde podrás:
- Modificar datos básicos (moneda, tiempo de entrega, validez, estado, etc.)
- Editar descripción y detalles del presupuesto
- Actualizar precios y cantidades
- Ver totales calculados automáticamente (subtotal, ITBIS, retención)
- Guardar cambios en la base de datos`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			// Si se proporciona un ID como argumento
			id, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Printf("Error: El ID '%s' no es un número válido\n", args[0])
				return
			}
			editarCotizacionPorID(id)
		} else {
			// Si no se proporciona ID, usar búsqueda interactiva
			editarCotizacion()
		}
	},
}

func init() {
	EditCmd.AddCommand(editQuotationCmd)
}

func editarCotizacion() {
	// Buscar cotización existente
	cotizacionOriginal := BuscarCotizacion()
	if cotizacionOriginal.ID == 0 {
		fmt.Println("No se seleccionó ninguna cotización")
		return
	}

	editarCotizacionPorID(cotizacionOriginal.ID)
}

func editarCotizacionPorID(idCotizacion int) {
	// Verificar que la cotización existe
	_, err := buscarCotizacionPorIDDirecto(idCotizacion)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Obtener datos completos de la cotización
	cotizacionCompleta, err := ObtenerCotizacionCompleta(idCotizacion)
	if err != nil {
		fmt.Printf("Error al obtener datos completos de la cotización ID %d: %v\n", idCotizacion, err)
		return
	}

	cliente := cotizacionCompleta.Cliente
	proyecto := cotizacionCompleta.Proyecto
	servicio := cotizacionCompleta.Servicio
	cotizacion := cotizacionCompleta.Cotizacion
	presupuesto := cotizacionCompleta.Presupuesto

	// Crear la aplicación Fyne
	a := app.New()
	w := a.NewWindow(fmt.Sprintf("Modificar Cotización #%d", cotizacion.ID))
	w.Resize(fyne.NewSize(800, 600))

	// Crear bindings para los campos de entrada con valores existentes
	monedaBinding := binding.NewString()
	tiempoEntregaBinding := binding.NewString()
	avanceBinding := binding.NewString()
	validezBinding := binding.NewString()
	estadoBinding := binding.NewString()
	idiomaBinding := binding.NewString()
	descripcionBinding := binding.NewString()
	retencionBinding := binding.NewString()
	descripcionPresupuestoBinding := binding.NewString()
	precioBinding := binding.NewString()
	cantidadBinding := binding.NewString()

	// Establecer valores existentes
	monedaBinding.Set(cotizacion.Moneda)
	tiempoEntregaBinding.Set(cotizacion.TiempoEntrega)
	avanceBinding.Set(cotizacion.Avance)
	validezBinding.Set(fmt.Sprintf("%d", cotizacion.Validez))
	estadoBinding.Set(cotizacion.Estado)
	idiomaBinding.Set(cotizacion.Idioma)
	descripcionBinding.Set(cotizacion.Descripcion)
	retencionBinding.Set(cotizacion.Retencion)

	// Obtener datos del presupuesto existente
	var existingPrice, existingQuantity float64 = 0, 1
	var existingDescription string = ""

	if presupuesto.Presupuesto != nil {
		if presupuestoData, ok := presupuesto.Presupuesto["presupuesto"].([]interface{}); ok && len(presupuestoData) > 0 {
			if parentItem, ok := presupuestoData[0].(map[string]interface{}); ok {
				if children, ok := parentItem["children"].([]interface{}); ok && len(children) > 0 {
					if child, ok := children[0].(map[string]interface{}); ok {
						if precio, ok := child["precio"].(float64); ok {
							existingPrice = precio
						}
						if cantidad, ok := child["cantidad"].(float64); ok {
							existingQuantity = cantidad
						}
						if desc, ok := child["descripcion"].(string); ok {
							existingDescription = desc
						}
					}
				}
			}
		}
	}

	precioBinding.Set(fmt.Sprintf("%.2f", existingPrice))
	cantidadBinding.Set(fmt.Sprintf("%.0f", existingQuantity))
	descripcionPresupuestoBinding.Set(existingDescription)

	// Bindings para cálculos
	subtotalBinding := binding.NewFloat()
	itbisBinding := binding.NewFloat()
	retencionMBinding := binding.NewFloat()
	totalBinding := binding.NewFloat()

	// Función para calcular totales
	calcularTotales := func() {
		precioStr, _ := precioBinding.Get()
		cantidadStr, _ := cantidadBinding.Get()
		retencionStr, _ := retencionBinding.Get()

		precio, err := strconv.ParseFloat(precioStr, 64)
		if err != nil {
			precio = 0
		}

		cantidad, err := strconv.ParseFloat(cantidadStr, 64)
		if err != nil {
			cantidad = 0
		}

		subtotal := math.Round((precio*cantidad)*100) / 100
		subtotalBinding.Set(subtotal)

		itbisP := 18.0 // ITBIS fijo al 18%
		itbisM := math.Round((subtotal*(itbisP/100))*100) / 100
		itbisBinding.Set(itbisM)

		var retencionP float64
		if retencionStr != "NINGUNA" {
			retencionP = 30.0
		}

		retencionM := math.Round((retencionP*itbisM/100)*100) / 100
		retencionMBinding.Set(retencionM)

		total := math.Round((subtotal+itbisM-retencionM)*100) / 100
		totalBinding.Set(total)
	}

	// Sección de información del cliente
	clienteLabel := widget.NewLabel(fmt.Sprintf("Cliente: %s", cliente.Nombre))
	clienteLabel.TextStyle = fyne.TextStyle{Bold: true}

	proyectoLabel := widget.NewLabel(fmt.Sprintf("Proyecto: %s", proyecto.Nombre))
	servicioLabel := widget.NewLabel(fmt.Sprintf("Servicio: %s", servicio.Nombre))

	cotizacionIDLabel := widget.NewLabel(fmt.Sprintf("Cotización ID: %d", cotizacion.ID))
	cotizacionIDLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Cargar logo del cliente si existe
	var logoContainer *fyne.Container
	imgURL := viper.GetString("uri.img")
	if imgURL != "" {
		logoURL := fmt.Sprintf("%s/images/logos/%d", imgURL, cliente.ID)

		// Intenta cargar el logo
		res, err := http.Get(logoURL)
		if err == nil && res.StatusCode == 200 {
			img := canvas.NewImageFromReader(res.Body, "logo-cliente")
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(200, 100))

			logoContainer = container.NewCenter(img)
		} else {
			// Si no hay logo, mostrar mensaje
			logoContainer = container.NewCenter(widget.NewLabel("Logo no disponible"))
		}
	} else {
		logoContainer = container.NewCenter(widget.NewLabel("URL de imágenes no configurada"))
	}

	// Crear widgets para la entrada de datos
	monedaSelect := widget.NewSelect([]string{"RD$", "USD"}, func(selected string) {
		monedaBinding.Set(selected)
		calcularTotales()
	})
	monedaValue, _ := monedaBinding.Get()
	monedaSelect.SetSelected(monedaValue)

	tiempoEntregaEntry := widget.NewEntryWithData(tiempoEntregaBinding)
	avanceEntry := widget.NewEntryWithData(avanceBinding)
	validezEntry := widget.NewEntryWithData(validezBinding)

	estadoSelect := widget.NewSelect([]string{"GENERADA", "TERMINADA", "APROBADA"}, func(selected string) {
		estadoBinding.Set(selected)
	})
	estadoValue, _ := estadoBinding.Get()
	estadoSelect.SetSelected(estadoValue)

	idiomaEntry := widget.NewEntryWithData(idiomaBinding)
	descripcionEntry := widget.NewMultiLineEntry()
	descripcionEntry.Bind(descripcionBinding)

	retencionSelect := widget.NewSelect([]string{"NINGUNA", "HONORARIOS PROFESIONALES"}, func(selected string) {
		retencionBinding.Set(selected)
		calcularTotales()
	})
	retencionValue, _ := retencionBinding.Get()
	retencionSelect.SetSelected(retencionValue)

	// Widgets para el presupuesto
	descripcionPresupuestoEntry := widget.NewMultiLineEntry()
	descripcionPresupuestoEntry.Bind(descripcionPresupuestoBinding)

	precioEntry := widget.NewEntryWithData(precioBinding)
	precioEntry.OnChanged = func(s string) {
		calcularTotales()
	}

	cantidadEntry := widget.NewEntryWithData(cantidadBinding)
	cantidadEntry.OnChanged = func(s string) {
		calcularTotales()
	}

	// Widgets para mostrar totales
	subtotalLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(subtotalBinding, "Subtotal: %.2f"))
	itbisLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(itbisBinding, "ITBIS (18%%): %.2f"))
	retencionLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(retencionMBinding, "Retención: %.2f"))
	totalLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(totalBinding, "Total: %.2f"))
	totalLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Botón de guardar
	guardarBtn := widget.NewButton("Actualizar Cotización", func() {
		// Obtener valores de los bindings
		moneda, _ := monedaBinding.Get()
		tiempoEntrega, _ := tiempoEntregaBinding.Get()
		avance, _ := avanceBinding.Get()
		validezStr, _ := validezBinding.Get()
		estado, _ := estadoBinding.Get()
		idioma, _ := idiomaBinding.Get()
		descripcion, _ := descripcionBinding.Get()
		retencion, _ := retencionBinding.Get()
		descripcionPresupuesto, _ := descripcionPresupuestoBinding.Get()

		validez, _ := strconv.Atoi(validezStr)

		subtotal, _ := subtotalBinding.Get()
		itbisM, _ := itbisBinding.Get()
		retencionM, _ := retencionMBinding.Get()
		total, _ := totalBinding.Get()

		precioStr, _ := precioBinding.Get()
		cantidadStr, _ := cantidadBinding.Get()
		precio, _ := strconv.ParseFloat(precioStr, 64)
		cantidad, _ := strconv.ParseFloat(cantidadStr, 64)

		// Actualizar cotización existente
		cotizacionActualizada := Cotizacion{
			ID:            cotizacion.ID, // Mantener el ID original
			IDCliente:     cliente.ID,
			IDProyecto:    proyecto.ID,
			IDServicio:    servicio.ID,
			Moneda:        moneda,
			Fecha:         cotizacion.Fecha, // Mantener fecha original
			TiempoEntrega: tiempoEntrega,
			Avance:        avance,
			Validez:       validez,
			Estado:        estado,
			Idioma:        idioma,
			Descripcion:   descripcion,
			Retencion:     retencion,
			Subtotal:      subtotal,
			Indirectos:    0,
			DescuentoP:    0,
			DescuentoM:    0,
			RetencionP:    retencionM * 100 / itbisM, // Calculamos el porcentaje
			RetencionM:    retencionM,
			ItbisP:        18.0,
			ItbisM:        itbisM,
			Total:         total,
		}

		// Actualizar cotización
		success := actualizarCotizacion(cotizacionActualizada)
		if !success {
			dialog := widget.NewLabel("Error al actualizar la cotización")
			dialog.Alignment = fyne.TextAlignCenter
			container := container.NewVBox(dialog, widget.NewButton("Cerrar", func() {
				w.Close()
			}))
			w.SetContent(container)
			return
		}

		// Crear estructura del presupuesto actualizado
		childItem := item{
			ID:          uuid.New().String()[:6],
			Item:        "P-1",
			Total:       subtotal,
			Moneda:      moneda,
			Precio:      precio,
			Unidad:      "Ud.",
			Cantidad:    cantidad,
			Descripcion: descripcionPresupuesto,
		}

		parentItemObj := parentItem{
			ID:          uuid.New().String()[:6],
			Item:        "I-1",
			Total:       subtotal,
			Moneda:      "",
			Precio:      "",
			Unidad:      "Ud.",
			Cantidad:    1.0,
			Children:    []item{childItem},
			Categoria:   "cat1",
			Descripcion: "SERVICIO",
		}

		presupuestoObj := presupuestoStruct{
			Indirectos:  []interface{}{},
			Presupuesto: []parentItem{parentItemObj},
		}

		// Actualizar presupuesto
		presupuestoData := Presupuesto{
			ID:           presupuesto.ID, // Usar ID existente
			IDCotizacion: cotizacion.ID,
			Presupuesto:  map[string]interface{}{"indirectos": presupuestoObj.Indirectos, "presupuesto": presupuestoObj.Presupuesto},
		}

		successPresupuesto := actualizarPresupuesto(presupuestoData)

		if successPresupuesto {
			dialog := widget.NewLabel(fmt.Sprintf("Cotización #%d actualizada con éxito", cotizacion.ID))
			dialog.Alignment = fyne.TextAlignCenter

			printBtn := widget.NewButton("Imprimir Cotización", func() {
				go func() {
					filePath, err := GetCotizacion(cotizacion.ID)
					if err != nil {
						fmt.Println("Error al imprimir cotización:", err)
					} else {
						fmt.Printf("Cotización %d guardada en %s\n", cotizacion.ID, filePath)
					}
				}()
			})

			closeBtn := widget.NewButton("Cerrar", func() {
				w.Close()
			})

			container := container.NewVBox(dialog, printBtn, closeBtn)
			w.SetContent(container)
		} else {
			dialog := widget.NewLabel("Error al actualizar el presupuesto")
			dialog.Alignment = fyne.TextAlignCenter
			container := container.NewVBox(dialog, widget.NewButton("Cerrar", func() {
				w.Close()
			}))
			w.SetContent(container)
		}
	})

	// Organizar la interfaz en contenedores
	headerContainer := container.NewVBox(
		logoContainer,
		cotizacionIDLabel,
		clienteLabel,
		proyectoLabel,
		servicioLabel,
		widget.NewSeparator(),
	)

	formContainer := container.NewGridWithColumns(2,
		widget.NewLabel("Moneda:"), monedaSelect,
		widget.NewLabel("Tiempo de entrega:"), tiempoEntregaEntry,
		widget.NewLabel("Avance (%):"), avanceEntry,
		widget.NewLabel("Validez (días):"), validezEntry,
		widget.NewLabel("Estado:"), estadoSelect,
		widget.NewLabel("Idioma:"), idiomaEntry,
		widget.NewLabel("Retención:"), retencionSelect,
	)

	descripcionContainer := container.NewBorder(
		widget.NewLabel("Descripción:"), nil, nil, nil,
		descripcionEntry,
	)

	presupuestoContainer := container.NewVBox(
		widget.NewLabel("Detalles del presupuesto:"),
		container.NewGridWithColumns(2,
			widget.NewLabel("Descripción:"), descripcionPresupuestoEntry,
			widget.NewLabel("Precio:"), precioEntry,
			widget.NewLabel("Cantidad:"), cantidadEntry,
		),
	)

	totalesContainer := container.NewVBox(
		widget.NewSeparator(),
		subtotalLabel,
		itbisLabel,
		retencionLabel,
		totalLabel,
	)

	// Contenedor principal
	content := container.NewVBox(
		headerContainer,
		container.NewPadded(formContainer),
		container.NewPadded(descripcionContainer),
		container.NewPadded(presupuestoContainer),
		container.NewPadded(totalesContainer),
		container.NewCenter(guardarBtn),
	)

	// Poner el contenido en un scroll en caso de que la ventana sea pequeña
	scrollContainer := container.NewScroll(content)
	w.SetContent(scrollContainer)

	// Inicializar cálculos con valores existentes
	calcularTotales()

	// Mostrar la ventana
	w.ShowAndRun()
}

// Función para actualizar una cotización existente
func actualizarCotizacion(cotizacion Cotizacion) bool {
	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		fmt.Println("URL de PostgREST no configurada")
		return false
	}

	cotizacionJSON, err := json.Marshal(cotizacion)
	if err != nil {
		fmt.Println("Error al convertir cotización a JSON:", err)
		return false
	}

	url := fmt.Sprintf("%s/cotizacion?id=eq.%d", postgrestURL, cotizacion.ID)
	headers["Content-Type"] = "application/json"

	_, err = MakeRequest("PATCH", url, headers, cotizacionJSON)
	if err != nil {
		fmt.Println("Error al actualizar cotización:", err)
		return false
	}

	return true
}

// Función para buscar cotización por ID directamente
func buscarCotizacionPorIDDirecto(id int) (Cotizacion, error) {
	var cotizacion Cotizacion

	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		return cotizacion, fmt.Errorf("URL de PostgREST no configurada")
	}

	url := fmt.Sprintf("%s/cotizacion?id=eq.%d", postgrestURL, id)
	resp, err := MakeRequest("GET", url, headers, nil)
	if err != nil {
		return cotizacion, fmt.Errorf("error al buscar cotización: %w", err)
	}

	var cotizaciones []Cotizacion
	if err := json.Unmarshal(resp, &cotizaciones); err != nil {
		return cotizacion, fmt.Errorf("error al procesar respuesta: %w", err)
	}

	if len(cotizaciones) == 0 {
		return cotizacion, fmt.Errorf("no se encontró cotización con ID %d", id)
	}

	return cotizaciones[0], nil
}

// Función para actualizar un presupuesto existente
func actualizarPresupuesto(presupuesto Presupuesto) bool {
	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		fmt.Println("URL de PostgREST no configurada")
		return false
	}

	presupuestoJSON, err := json.Marshal(presupuesto)
	if err != nil {
		fmt.Println("Error al convertir presupuesto a JSON:", err)
		return false
	}

	url := fmt.Sprintf("%s/presupuesto?id_cotizacion=eq.%d", postgrestURL, presupuesto.IDCotizacion)
	headers["Content-Type"] = "application/json"

	_, err = MakeRequest("PATCH", url, headers, presupuestoJSON)
	if err != nil {
		fmt.Println("Error al actualizar presupuesto:", err)
		return false
	}

	return true
}
