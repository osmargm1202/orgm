package adm

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
)

var nuevoClienteCmd = &cobra.Command{
	Use:   "client",
	Short: "create a new client",
	RunE: func(cmd *cobra.Command, args []string) error {
		cliente := CrearNuevoCliente()
		fmt.Println("Cliente creado:", cliente)
		return nil
	},
}

func init() {
	NewCmd.AddCommand(nuevoClienteCmd)
}

type Cliente struct {
	ID                     int    `json:"id"`
	Nombre                 string `json:"nombre"`
	NombreComercial        string `json:"nombre_comercial"`
	Numero                 string `json:"numero"`
	Correo                 string `json:"correo"`
	Direccion              string `json:"direccion"`
	Ciudad                 string `json:"ciudad"`
	Provincia              string `json:"provincia"`
	Telefono               string `json:"telefono"`
	Representante          string `json:"representante"`
	TelefonoRepresentante  string `json:"telefono_representante"`
	ExtensionRepresentante string `json:"extension_representante"`
	CelularRepresentante   string `json:"celular_representante"`
	CorreoRepresentante    string `json:"correo_representante"`
	TipoFactura            string `json:"tipo_factura"`
	FechaActualizacion     string `json:"fecha_actualizacion"`
}

func BuscarCliente() Cliente {
	var cliente Cliente

	// Usar el componente de selección para elegir el método de búsqueda
	items := []inputs.Item{
		{ID: "search", Name: "Buscar por términos", Desc: "Buscar por nombre, número o nombre comercial"},
		{ID: "id", Name: "Buscar por ID", Desc: "Buscar directamente por el ID del cliente"},
		{ID: "new", Name: "Crear nuevo cliente", Desc: "Crear un nuevo cliente en el sistema"},
	}

	selectedMethod := inputs.SelectList("Seleccione método de búsqueda de clientes", items)

	if selectedMethod.ID == "id" {
		// Buscar por ID
		idStr := inputs.GetInput("Ingrese el ID del cliente")
		_, err := strconv.Atoi(idStr) // Validar que sea un número
		if err != nil {
			fmt.Println("ID inválido")
			return cliente
		}

		postgrestURL, headers := InitializePostgrest()
		url := postgrestURL + "/cliente?id=eq." + idStr

		resp, err := MakeRequest("GET", url, headers, nil)
		if err != nil {
			fmt.Println("Error al buscar cliente:", err)
			return cliente
		}

		var clientes []Cliente
		if err := json.Unmarshal(resp, &clientes); err != nil {
			fmt.Println("Error al procesar respuesta:", err)
			return cliente
		}

		if len(clientes) > 0 {
			return clientes[0]
		}

		fmt.Println("No se encontró cliente con ID:", idStr)
		return cliente
	} else if selectedMethod.ID == "new" {
		// Crear nuevo cliente
		return CrearNuevoCliente()
	} else {
		// Buscar por términos
		p := tea.NewProgram(initialClienteModel())

		model, err := p.Run()
		if err != nil {
			fmt.Println("Error al buscar cliente:", err)
			return cliente
		}

		clientModel, ok := model.(clienteModel)
		if ok && clientModel.selectedCliente.ID > 0 {
			return clientModel.selectedCliente
		}
	}

	return cliente
}

func CrearNuevoCliente() Cliente {
	var cliente Cliente

	fmt.Println("\n=== CREAR NUEVO CLIENTE ===")
	nextID, err := GetLastID("cliente")
	if err != nil {
		fmt.Println("Error al obtener el último ID:", err)
		return Cliente{}
	}

	cliente.ID = nextID

	// Datos básicos del cliente
	cliente.Nombre = inputs.GetInput("Nombre del cliente")
	if cliente.Nombre == "" {
		fmt.Println("El nombre del cliente es obligatorio")
		return Cliente{}
	}

	cliente.NombreComercial = inputs.GetInputWithDefault("Nombre comercial", "")
	cliente.Numero = inputs.GetInputWithDefault("Número/RNC", "")
	cliente.Correo = inputs.GetInputWithDefault("Correo electrónico", "")
	cliente.Telefono = inputs.GetInputWithDefault("Teléfono", "")

	// Dirección
	cliente.Direccion = inputs.GetInputWithDefault("Dirección", "")
	cliente.Ciudad = inputs.GetInputWithDefault("Ciudad", "")
	cliente.Provincia = inputs.GetInputWithDefault("Provincia", "")

	// Datos del representante
	cliente.Representante = inputs.GetInputWithDefault("Nombre del representante", "")
	cliente.TelefonoRepresentante = inputs.GetInputWithDefault("Teléfono del representante", "")
	cliente.ExtensionRepresentante = inputs.GetInputWithDefault("Extensión del representante", "")
	cliente.CelularRepresentante = inputs.GetInputWithDefault("Celular del representante", "")
	cliente.CorreoRepresentante = inputs.GetInputWithDefault("Correo del representante", "")

	// Tipo de factura
	tipoFacturaItems := []inputs.Item{
		{ID: "NCFC", Name: "NCFC", Desc: "Comprobante Fiscal de Crédito"},
		{ID: "NCF", Name: "NCF", Desc: "Comprobante Fiscal"},
		{ID: "NCG", Name: "NCG", Desc: "Comprobante Gubernamental"},
		{ID: "NCRE", Name: "NCRE", Desc: "Régimen Especial"},
	}

	tipoFacturaSelected := inputs.SelectList("Seleccione el tipo de factura", tipoFacturaItems)
	cliente.TipoFactura = tipoFacturaSelected.ID

	// Fecha de actualización (actual)
	cliente.FechaActualizacion = time.Now().Format("02/01/2006")

	// Confirmar antes de guardar
	fmt.Println("\n=== RESUMEN DEL CLIENTE ===")
	fmt.Printf("Nombre: %s\n", cliente.Nombre)
	fmt.Printf("Nombre Comercial: %s\n", cliente.NombreComercial)
	fmt.Printf("Número/RNC: %s\n", cliente.Numero)
	fmt.Printf("Correo: %s\n", cliente.Correo)
	fmt.Printf("Teléfono: %s\n", cliente.Telefono)
	fmt.Printf("Dirección: %s\n", cliente.Direccion)
	fmt.Printf("Ciudad: %s\n", cliente.Ciudad)
	fmt.Printf("Provincia: %s\n", cliente.Provincia)
	fmt.Printf("Representante: %s\n", cliente.Representante)
	fmt.Printf("Tipo de Factura: %s\n", cliente.TipoFactura)

	confirmarItems := []inputs.Item{
		{ID: "si", Name: "Sí", Desc: "Guardar el cliente"},
		{ID: "no", Name: "No", Desc: "Cancelar y no guardar"},
	}

	confirmar := inputs.SelectList("¿Confirma que desea guardar este cliente?", confirmarItems)

	if confirmar.ID == "si" {
		// Guardar cliente
		clienteGuardado := GuardarCliente(cliente)
		if clienteGuardado.ID > 0 {
			fmt.Printf("Cliente guardado exitosamente con ID: %d\n", clienteGuardado.ID)
			return clienteGuardado
		} else {
			fmt.Println("Error al guardar el cliente")
			return Cliente{}
		}
	} else {
		fmt.Println("Creación de cliente cancelada")
		return Cliente{}
	}
}

func GetLastID(table string) (int, error) {
	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		return 0, fmt.Errorf("URL de PostgREST no configurada")
	}

	url := postgrestURL + "/" + table + "?select=id"

	resp, err := MakeRequest("GET", url, headers, nil)
	if err != nil {
		return 0, fmt.Errorf("error en request: %w", err)
	}

	var clientes []Cliente

	if err := json.Unmarshal(resp, &clientes); err != nil {
		return 0, fmt.Errorf("error al deserializar respuesta: %w", err)
	}

	if len(clientes) == 0 {
		return 1, nil
	}
	var maxID int = 0
	for _, c := range clientes {
		if c.ID > maxID {
			maxID = c.ID
		}
	}
	maxID += 1
	// existingIDs := make(map[int]bool)
	// for _, c := range clientes {
	// 	existingIDs[c.ID] = true
	// }

	// // Buscar el primer número faltante del 1 al 40
	// for i := 1; i <= 144; i++ {
	// 	if !existingIDs[i] {
	// 		fmt.Printf("El número %d no existe en la secuencia\n", i)
	// 	}
	// }
	// fmt.Println("existingIDs:", len(existingIDs))

	fmt.Println("Proximo ID:", maxID)
	return maxID, nil
}

func GuardarCliente(cliente Cliente) Cliente {
	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		fmt.Println("Error: URL de PostgREST no configurada")
		return Cliente{}
	}

	url := postgrestURL + "/cliente"

	// No incluir ID en el JSON ya que será generado automáticamente
	clienteParaGuardar := Cliente{
		ID:                     cliente.ID,
		Nombre:                 cliente.Nombre,
		NombreComercial:        cliente.NombreComercial,
		Numero:                 cliente.Numero,
		Correo:                 cliente.Correo,
		Direccion:              cliente.Direccion,
		Ciudad:                 cliente.Ciudad,
		Provincia:              cliente.Provincia,
		Telefono:               cliente.Telefono,
		Representante:          cliente.Representante,
		TelefonoRepresentante:  cliente.TelefonoRepresentante,
		ExtensionRepresentante: cliente.ExtensionRepresentante,
		CelularRepresentante:   cliente.CelularRepresentante,
		CorreoRepresentante:    cliente.CorreoRepresentante,
		TipoFactura:            cliente.TipoFactura,
		FechaActualizacion:     cliente.FechaActualizacion,
	}

	clienteBytes, err := json.Marshal(clienteParaGuardar)
	if err != nil {
		fmt.Println("Error al serializar cliente:", err)
		return Cliente{}
	}

	// Agregar header para indicar que queremos la respuesta con el registro creado
	headers["Prefer"] = "return=representation"

	resp, err := MakeRequest("POST", url, headers, clienteBytes)
	if err != nil {
		fmt.Println("Error al guardar cliente:", err)
		return Cliente{}
	}

	// Intentar deserializar la respuesta
	var clientesResp []Cliente
	if err := json.Unmarshal(resp, &clientesResp); err != nil {
		// Intentar deserializar como objeto único
		var clienteResp Cliente
		if err := json.Unmarshal(resp, &clienteResp); err != nil {
			fmt.Println("Error al procesar respuesta:", err)
			return Cliente{}
		}
		return clienteResp
	}

	if len(clientesResp) > 0 {
		return clientesResp[0]
	}

	return Cliente{}
}

// Command to search for clientes
func SearchClientes(query string) tea.Cmd {
	return func() tea.Msg {
		postgrestURL, headers := InitializePostgrest()
		if postgrestURL == "" {
			return clientesSearchResult{err: fmt.Errorf("URL de PostgREST no configurada")}
		}

		// Buscar por número, nombre o nombre_comercial
		url := postgrestURL + "/cliente?select=id,nombre,telefono,correo,direccion,nombre_comercial,numero&or=(nombre.ilike.*" + query + "*,nombre_comercial.ilike.*" + query + "*,numero.ilike.*" + query + "*)"

		resp, err := MakeRequest("GET", url, headers, nil)
		if err != nil {
			return clientesSearchResult{err: fmt.Errorf("error en request: %w", err)}
		}

		// Imprimir la respuesta para depuración
		fmt.Println("Respuesta JSON:", string(resp))

		// Intentar deserializar como array
		var clientes []Cliente
		err = json.Unmarshal(resp, &clientes)

		if err != nil {
			fmt.Println("Error deserializando como array:", err)

			// Si falla, intentar deserializar como objeto único
			var cliente Cliente
			err = json.Unmarshal(resp, &cliente)
			if err != nil {
				return clientesSearchResult{err: fmt.Errorf("error deserializando JSON: %w", err)}
			}
			// Si es un objeto único, crear un array con ese elemento
			clientes = []Cliente{cliente}
		}

		if len(clientes) == 0 {
			return clientesSearchResult{err: fmt.Errorf("no se encontraron clientes")}
		}

		return clientesSearchResult{clientes: clientes}
	}
}

// Message for cliente search results
type clientesSearchResult struct {
	clientes []Cliente
	err      error
}
