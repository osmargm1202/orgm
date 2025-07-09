package adm

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

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
	w.Resize(fyne.NewSize(1000, 800))

	// Crear bindings para los campos de entrada con valores existentes
	monedaBinding := binding.NewString()
	tiempoEntregaBinding := binding.NewString()
	avanceBinding := binding.NewString()
	validezBinding := binding.NewString()
	estadoBinding := binding.NewString()
	idiomaBinding := binding.NewString()
	descripcionBinding := binding.NewString()
	retencionBinding := binding.NewString()

	// Establecer valores existentes
	monedaBinding.Set(cotizacion.Moneda)
	tiempoEntregaBinding.Set(cotizacion.TiempoEntrega)
	avanceBinding.Set(cotizacion.Avance)
	validezBinding.Set(fmt.Sprintf("%d", cotizacion.Validez))
	estadoBinding.Set(cotizacion.Estado)
	idiomaBinding.Set(cotizacion.Idioma)
	descripcionBinding.Set(cotizacion.Descripcion)
	retencionBinding.Set(cotizacion.Retencion)

	// Estructura para manejar múltiples items
	type ItemPresupuesto struct {
		descripcionBinding binding.String
		precioBinding      binding.String
		cantidadBinding    binding.String
		descripcionWidget  *widget.Entry
		precioWidget       *widget.Entry
		cantidadWidget     *widget.Entry
		eliminarWidget     *widget.Button
		separadorWidget    *widget.Separator
	}

	// Array para manejar hasta 10 items
	var items []*ItemPresupuesto
	var mainContainer *fyne.Container

	// Bindings para cálculos
	subtotalBinding := binding.NewFloat()
	descuentoPBinding := binding.NewString()
	descuentoMBinding := binding.NewFloat()
	itbisBinding := binding.NewFloat()
	retencionMBinding := binding.NewFloat()
	totalBinding := binding.NewFloat()

	// Establecer valor por defecto del descuento
	descuentoPBinding.Set(fmt.Sprintf("%.2f", cotizacion.DescuentoP))

	// Función para calcular totales
	calcularTotales := func() {
		var subtotalTotal float64 = 0
		retencionStr, _ := retencionBinding.Get()
		descuentoPStr, _ := descuentoPBinding.Get()

		// Sumar todos los items
		for _, item := range items {
			precioStr, _ := item.precioBinding.Get()
			cantidadStr, _ := item.cantidadBinding.Get()

			precio, err := strconv.ParseFloat(precioStr, 64)
			if err != nil {
				precio = 0
			}

			cantidad, err := strconv.ParseFloat(cantidadStr, 64)
			if err != nil {
				cantidad = 0
			}

			subtotalTotal += precio * cantidad
		}

		subtotal := math.Round(subtotalTotal*100) / 100
		subtotalBinding.Set(subtotal)

		// Calcular descuento
		descuentoP, err := strconv.ParseFloat(descuentoPStr, 64)
		if err != nil {
			descuentoP = 0
		}
		descuentoM := math.Round((subtotal*(descuentoP/100))*100) / 100
		descuentoMBinding.Set(descuentoM)

		// Subtotal después del descuento
		subtotalConDescuento := subtotal - descuentoM

		itbisP := 18.0 // ITBIS fijo al 18%
		itbisM := math.Round((subtotalConDescuento*(itbisP/100))*100) / 100
		itbisBinding.Set(itbisM)

		var retencionP float64
		if retencionStr != "NINGUNA" {
			retencionP = 30.0
		}

		retencionM := math.Round((retencionP*itbisM/100)*100) / 100
		retencionMBinding.Set(retencionM)

		total := math.Round((subtotalConDescuento+itbisM-retencionM)*100) / 100
		totalBinding.Set(total)
	}

	// Función para recrear los items en el contenedor principal (declaración forward)
	var recrearItems func()

	// Función para crear un nuevo item
	crearNuevoItem := func(numero int) *ItemPresupuesto {
		nuevoItem := &ItemPresupuesto{
			descripcionBinding: binding.NewString(),
			precioBinding:      binding.NewString(),
			cantidadBinding:    binding.NewString(),
		}

		// Widgets para el item
		descripcionEntry := widget.NewMultiLineEntry()
		descripcionEntry.Bind(nuevoItem.descripcionBinding)
		descripcionEntry.SetPlaceHolder("Descripción del item...")
		descripcionEntry.Resize(fyne.NewSize(400, 60)) // Altura reducida a 60px

		precioEntry := widget.NewEntryWithData(nuevoItem.precioBinding)
		precioEntry.SetPlaceHolder("0.00")
		precioEntry.OnChanged = func(s string) {
			calcularTotales()
		}

		cantidadEntry := widget.NewEntryWithData(nuevoItem.cantidadBinding)
		cantidadEntry.SetText("1")
		cantidadEntry.OnChanged = func(s string) {
			calcularTotales()
		}

		// Botón para eliminar item
		eliminarBtn := widget.NewButton("X", func() {
			// Encontrar el índice del item a eliminar
			for i, item := range items {
				if item == nuevoItem {
					// Remover del slice
					items = append(items[:i], items[i+1:]...)
					break
				}
			}
			// Recrear los items
			recrearItems()
			calcularTotales()
		})
		eliminarBtn.Importance = widget.DangerImportance

		// Separador para este item
		separador := widget.NewSeparator()

		// Asignar widgets al item
		nuevoItem.descripcionWidget = descripcionEntry
		nuevoItem.precioWidget = precioEntry
		nuevoItem.cantidadWidget = cantidadEntry
		nuevoItem.eliminarWidget = eliminarBtn
		nuevoItem.separadorWidget = separador

		return nuevoItem
	}

	// Implementación de la función para recrear los items
	recrearItems = func() {
		// Remover todos los items del contenedor principal
		// (Necesitaremos encontrar el índice donde empiezan los items y remover desde ahí)
		// Por ahora, esto se implementará después cuando se defina el contenedor principal
	}

	// Obtener datos del presupuesto existente y crear items
	if presupuesto.Presupuesto != nil {
		if presupuestoData, ok := presupuesto.Presupuesto["presupuesto"].([]interface{}); ok && len(presupuestoData) > 0 {
			if parentItem, ok := presupuestoData[0].(map[string]interface{}); ok {
				if children, ok := parentItem["children"].([]interface{}); ok && len(children) > 0 {
					// Crear items basados en los datos existentes
					for i, childData := range children {
						if child, ok := childData.(map[string]interface{}); ok {
							nuevoItem := crearNuevoItem(i + 1)

							// Extraer datos del presupuesto existente
							if desc, ok := child["descripcion"].(string); ok {
								nuevoItem.descripcionBinding.Set(desc)
							}
							if precio, ok := child["precio"].(float64); ok {
								nuevoItem.precioBinding.Set(fmt.Sprintf("%.2f", precio))
							}
							if cantidad, ok := child["cantidad"].(float64); ok {
								nuevoItem.cantidadBinding.Set(fmt.Sprintf("%.0f", cantidad))
							}

							items = append(items, nuevoItem)
						}
					}
				}
			}
		}
	}

	// Si no hay items existentes, crear al menos uno
	if len(items) == 0 {
		primerItem := crearNuevoItem(1)
		items = append(items, primerItem)
	}

	// Botón para agregar más items
	agregarItemBtn := widget.NewButton("+ Agregar Item", func() {
		if len(items) < 10 {
			nuevoItem := crearNuevoItem(len(items) + 1)
			items = append(items, nuevoItem)
			recrearItems()
			calcularTotales()
		}
	})
	agregarItemBtn.Importance = widget.SuccessImportance

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

	descuentoEntry := widget.NewEntryWithData(descuentoPBinding)
	descuentoEntry.SetPlaceHolder("0.00")
	descuentoEntry.OnChanged = func(s string) {
		calcularTotales()
	}

	// Widgets para mostrar totales
	subtotalLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(subtotalBinding, "Subtotal: %.2f"))
	descuentoLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(descuentoMBinding, "Descuento: %.2f"))
	itbisLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(itbisBinding, "ITBIS (18%%): %.2f"))
	retencionLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(retencionMBinding, "Retención: %.2f"))
	totalLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(totalBinding, "Total: %.2f"))
	totalLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Botón de guardar
	guardarBtn := widget.NewButton("Actualizar Cotización", func() {
		// Validar que haya al menos un item con datos
		tieneItems := false
		for _, item := range items {
			desc, _ := item.descripcionBinding.Get()
			precio, _ := item.precioBinding.Get()
			if strings.TrimSpace(desc) != "" && strings.TrimSpace(precio) != "" {
				tieneItems = true
				break
			}
		}

		if !tieneItems {
			dialog := widget.NewLabel("Error: Debe agregar al menos un item con descripción y precio")
			dialog.Alignment = fyne.TextAlignCenter
			container := container.NewVBox(dialog, widget.NewButton("Cerrar", func() {
				w.Close()
			}))
			w.SetContent(container)
			return
		}

		// Obtener valores de los bindings
		moneda, _ := monedaBinding.Get()
		tiempoEntrega, _ := tiempoEntregaBinding.Get()
		avance, _ := avanceBinding.Get()
		validezStr, _ := validezBinding.Get()
		estado, _ := estadoBinding.Get()
		idioma, _ := idiomaBinding.Get()
		descripcion, _ := descripcionBinding.Get()
		retencion, _ := retencionBinding.Get()

		validez, _ := strconv.Atoi(validezStr)
		descuentoPStr, _ := descuentoPBinding.Get()
		descuentoP, _ := strconv.ParseFloat(descuentoPStr, 64)

		subtotal, _ := subtotalBinding.Get()
		descuentoM, _ := descuentoMBinding.Get()
		itbisM, _ := itbisBinding.Get()
		retencionM, _ := retencionMBinding.Get()
		total, _ := totalBinding.Get()

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
			DescuentoP:    descuentoP,
			DescuentoM:    descuentoM,
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

		// Crear estructura del presupuesto actualizado con múltiples items
		var children []item
		for i, itemData := range items {
			desc, _ := itemData.descripcionBinding.Get()
			precioStr, _ := itemData.precioBinding.Get()
			cantidadStr, _ := itemData.cantidadBinding.Get()

			// Solo agregar items que tengan descripción y precio
			if strings.TrimSpace(desc) != "" && strings.TrimSpace(precioStr) != "" {
				precio, _ := strconv.ParseFloat(precioStr, 64)
				cantidad, _ := strconv.ParseFloat(cantidadStr, 64)
				if cantidad <= 0 {
					cantidad = 1
				}

				childItem := item{
					ID:          uuid.New().String()[:6],
					Item:        fmt.Sprintf("P-%d", i+1),
					Total:       math.Round((precio*cantidad)*100) / 100,
					Moneda:      moneda,
					Precio:      precio,
					Unidad:      "Ud.",
					Cantidad:    cantidad,
					Descripcion: desc,
				}
				children = append(children, childItem)
			}
		}

		parentItemObj := parentItem{
			ID:          uuid.New().String()[:6],
			Item:        "I-1",
			Total:       subtotal,
			Moneda:      "",
			Precio:      "",
			Unidad:      "Ud.",
			Cantidad:    1.0,
			Children:    children,
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
		widget.NewLabel("Descuento (%):"), descuentoEntry,
		widget.NewLabel("Retención:"), retencionSelect,
	)

	descripcionContainer := container.NewBorder(
		widget.NewLabel("Descripción:"), nil, nil, nil,
		descripcionEntry,
	)

	// Contenedor principal donde se agregarán todos los elementos
	mainContainer = container.NewVBox(
		headerContainer,
		container.NewPadded(formContainer),
		container.NewPadded(descripcionContainer),
		widget.NewLabel("Items del presupuesto:"),
		container.NewHBox(widget.NewSeparator(), agregarItemBtn),
		widget.NewSeparator(),
	)

	// Agregar items iniciales al contenedor principal
	for i, item := range items {
		mainContainer.Add(widget.NewLabel(fmt.Sprintf("Item %d:", i+1)))

		// Crear contenedor horizontal con proporciones flexibles usando splits
		descripcionContainer := container.NewBorder(widget.NewLabel("Descripción:"), nil, nil, nil, item.descripcionWidget)
		precioContainer := container.NewBorder(widget.NewLabel("Precio:"), nil, nil, nil, item.precioWidget)
		cantidadContainer := container.NewBorder(widget.NewLabel("Cantidad:"), nil, nil, nil, item.cantidadWidget)

		// Usar Split containers para crear proporciones 60%-20%-20%
		rightSplit := container.NewHSplit(precioContainer, cantidadContainer)
		rightSplit.SetOffset(0.5) // 50-50 split between precio and cantidad

		mainSplit := container.NewHSplit(descripcionContainer, rightSplit)
		mainSplit.SetOffset(0.6) // 60% para descripción, 40% para el resto

		itemRow := container.NewBorder(nil, nil, nil, item.eliminarWidget, mainSplit)
		mainContainer.Add(itemRow)

		if i < len(items)-1 {
			mainContainer.Add(item.separadorWidget)
		}
	}

	// Totales
	totalesContainer := container.NewVBox(
		widget.NewSeparator(),
		subtotalLabel,
		descuentoLabel,
		itbisLabel,
		retencionLabel,
		totalLabel,
	)

	mainContainer.Add(container.NewPadded(totalesContainer))
	mainContainer.Add(container.NewCenter(guardarBtn))

	// Implementar la función recrearItems
	recrearItems = func() {
		// Remover todos los objetos desde el primer separador después del botón agregar hasta antes de totales
		newObjects := make([]fyne.CanvasObject, 0)

		// Agregar elementos hasta (e incluyendo) el separador después del botón agregar
		foundSeparator := false
		for i, obj := range mainContainer.Objects {
			newObjects = append(newObjects, obj)
			if !foundSeparator {
				if _, ok := obj.(*widget.Separator); ok {
					// Verificar si el objeto anterior contiene el botón agregar
					if i > 0 {
						if hbox, ok := mainContainer.Objects[i-1].(*fyne.Container); ok {
							// Si es un HBox que probablemente contiene el botón agregar
							if len(hbox.Objects) > 1 {
								foundSeparator = true
								break
							}
						}
					}
				}
			}
		}

		// Agregar los items actuales
		for i, item := range items {
			newObjects = append(newObjects, widget.NewLabel(fmt.Sprintf("Item %d:", i+1)))

			// Crear contenedor horizontal con proporciones flexibles usando splits
			descripcionContainer := container.NewBorder(widget.NewLabel("Descripción:"), nil, nil, nil, item.descripcionWidget)
			precioContainer := container.NewBorder(widget.NewLabel("Precio:"), nil, nil, nil, item.precioWidget)
			cantidadContainer := container.NewBorder(widget.NewLabel("Cantidad:"), nil, nil, nil, item.cantidadWidget)

			// Usar Split containers para crear proporciones 60%-20%-20%
			rightSplit := container.NewHSplit(precioContainer, cantidadContainer)
			rightSplit.SetOffset(0.5) // 50-50 split between precio and cantidad

			mainSplit := container.NewHSplit(descripcionContainer, rightSplit)
			mainSplit.SetOffset(0.6) // 60% para descripción, 40% para el resto

			itemRow := container.NewBorder(nil, nil, nil, item.eliminarWidget, mainSplit)

			newObjects = append(newObjects, itemRow)

			if i < len(items)-1 {
				newObjects = append(newObjects, item.separadorWidget)
			}
		}

		// Agregar los totales y botón guardar (últimos 2 elementos del contenedor original)
		if len(mainContainer.Objects) >= 2 {
			newObjects = append(newObjects, mainContainer.Objects[len(mainContainer.Objects)-2:]...)
		}

		// Reemplazar todos los objetos
		mainContainer.Objects = newObjects
		mainContainer.Refresh()
	}

	// Poner el contenido en un scroll para toda la aplicación
	scrollContainer := container.NewScroll(mainContainer)
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
