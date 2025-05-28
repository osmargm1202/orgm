package adm

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
)

var nuevoProyectoCmd = &cobra.Command{
	Use:   "project",
	Short: "create a new project",
	RunE: func(cmd *cobra.Command, args []string) error {
		proyecto := CrearNuevoProyecto()
		fmt.Println("Proyecto creado:", proyecto)
		return nil
	},
}

func init() {
	NewCmd.AddCommand(nuevoProyectoCmd)
}

type Proyecto struct {
	ID          int    `json:"id"`
	Ubicacion   string `json:"ubicacion"`
	Nombre      string `json:"nombre_proyecto"`
	Descripcion string `json:"descripcion"`
}

func BuscarProyecto() Proyecto {
	var proyecto Proyecto

	// Usar el componente de selección para elegir el método de búsqueda
	items := []inputs.Item{
		{ID: "search", Name: "Buscar por términos", Desc: "Buscar por nombre o ubicación del proyecto"},
		{ID: "id", Name: "Buscar por ID", Desc: "Buscar directamente por el ID del proyecto"},
		{ID: "new", Name: "Crear nuevo proyecto", Desc: "Crear un nuevo proyecto en el sistema"},
	}

	selectedMethod := inputs.SelectList("Seleccione método de búsqueda de proyectos", items)

	if selectedMethod.ID == "id" {
		// Buscar por ID
		idStr := inputs.GetInput("Ingrese el ID del proyecto")
		_, err := strconv.Atoi(idStr) // Validar que sea un número
		if err != nil {
			fmt.Println("ID inválido")
			return proyecto
		}

		postgrestURL, headers := InitializePostgrest()
		requestURL := postgrestURL + "/proyecto?id=eq." + idStr

		resp, err := MakeRequest("GET", requestURL, headers, nil)
		if err != nil {
			fmt.Println("Error al buscar proyecto:", err)
			return proyecto
		}

		var proyectos []Proyecto
		if err := json.Unmarshal(resp, &proyectos); err != nil {
			fmt.Println("Error al procesar respuesta:", err)
			return proyecto
		}

		if len(proyectos) > 0 {
			return proyectos[0]
		}

		fmt.Println("No se encontró proyecto con ID:", idStr)
		return proyecto
	} else if selectedMethod.ID == "new" {
		// Crear nuevo proyecto
		return CrearNuevoProyecto()
	} else {
		// Buscar por términos
		query := inputs.GetInput("Ingrese nombre o ubicación del proyecto: ")

		if query == "" {
			fmt.Println("Búsqueda vacía")
			return proyecto
		}

		postgrestURL, headers := InitializePostgrest()

		// Escapar el query parameter correctamente para URLs
		escapedQuery := url.QueryEscape(query)

		// Search by nombre, ubicacion or descripcion using ilike
		requestURL := fmt.Sprintf("%s/proyecto?select=*&or=(nombre_proyecto.ilike.*%s*,ubicacion.ilike.*%s*)",
			postgrestURL, escapedQuery, escapedQuery)

		resp, err := MakeRequest("GET", requestURL, headers, nil)
		if err != nil {
			fmt.Println("Error al buscar proyectos:", err)
			return proyecto
		}

		var proyectos []Proyecto
		if err := json.Unmarshal(resp, &proyectos); err != nil {
			fmt.Println("Error al procesar proyectos:", err)
			return proyecto
		}

		if len(proyectos) == 0 {
			fmt.Println("No se encontraron proyectos")
			return proyecto
		}

		items := make([]inputs.Item, len(proyectos))
		for i, p := range proyectos {
			items[i] = inputs.Item{
				ID:   strconv.Itoa(p.ID),
				Name: p.Nombre,
				Desc: fmt.Sprintf("ID: %d | %s", p.ID, p.Ubicacion),
			}
		}

		selectedItem := inputs.SelectList("Seleccione un proyecto", items)

		id, _ := strconv.Atoi(selectedItem.ID)
		for _, p := range proyectos {
			if p.ID == id {
				proyecto = p
				break
			}
		}
	}

	return proyecto
}

func CrearNuevoProyecto() Proyecto {
	var proyecto Proyecto

	fmt.Println("\n=== CREAR NUEVO PROYECTO ===")

	nextID, err := GetLastIDProyecto()
	if err != nil {
		fmt.Println("Error al obtener el último ID:", err)
		return Proyecto{}
	}

	proyecto.ID = nextID

	// Datos del proyecto
	proyecto.Nombre = inputs.GetInput("Nombre del proyecto")
	if proyecto.Nombre == "" {
		fmt.Println("El nombre del proyecto es obligatorio")
		return Proyecto{}
	}

	// Buscar ubicación usando el sistema de ubicaciones
	fmt.Println("\n--- Seleccionar ubicación del proyecto ---")
	ubicacion := BuscarUbicacion()
	if ubicacion.ID == 0 {
		fmt.Println("No se seleccionó ninguna ubicación")
		return Proyecto{}
	}

	// Guardar la ubicación como string combinado
	proyecto.Ubicacion = ubicacion.GetCombinedString()

	proyecto.Descripcion = inputs.GetInputWithDefault("Descripción del proyecto", "")

	// Confirmar antes de guardar
	fmt.Println("\n=== RESUMEN DEL PROYECTO ===")
	fmt.Printf("ID: %d\n", proyecto.ID)
	fmt.Printf("Nombre: %s\n", proyecto.Nombre)
	fmt.Printf("Ubicación: %s\n", proyecto.Ubicacion)
	fmt.Printf("Descripción: %s\n", proyecto.Descripcion)

	confirmarItems := []inputs.Item{
		{ID: "si", Name: "Sí", Desc: "Guardar el proyecto"},
		{ID: "no", Name: "No", Desc: "Cancelar y no guardar"},
	}

	confirmar := inputs.SelectList("¿Confirma que desea guardar este proyecto?", confirmarItems)

	if confirmar.ID == "si" {
		// Guardar proyecto
		proyectoGuardado := GuardarProyecto(proyecto)
		if proyectoGuardado.ID > 0 {
			fmt.Printf("Proyecto guardado exitosamente con ID: %d\n", proyectoGuardado.ID)
			return proyectoGuardado
		} else {
			fmt.Println("Error al guardar el proyecto")
			return Proyecto{}
		}
	} else {
		fmt.Println("Creación de proyecto cancelada")
		return Proyecto{}
	}
}

func GetLastIDProyecto() (int, error) {
	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		return 0, fmt.Errorf("URL de PostgREST no configurada")
	}

	url := postgrestURL + "/proyecto?select=id"

	resp, err := MakeRequest("GET", url, headers, nil)
	if err != nil {
		return 0, fmt.Errorf("error en request: %w", err)
	}

	var proyectos []Proyecto

	if err := json.Unmarshal(resp, &proyectos); err != nil {
		return 0, fmt.Errorf("error al deserializar respuesta: %w", err)
	}

	if len(proyectos) == 0 {
		return 1, nil
	}

	var maxID int = 0
	for _, p := range proyectos {
		if p.ID > maxID {
			maxID = p.ID
		}
	}
	maxID += 1

	fmt.Println("Próximo ID de proyecto:", maxID)
	return maxID, nil
}

func GuardarProyecto(proyecto Proyecto) Proyecto {
	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		fmt.Println("Error: URL de PostgREST no configurada")
		return Proyecto{}
	}

	url := postgrestURL + "/proyecto"

	proyectoParaGuardar := Proyecto{
		ID:          proyecto.ID,
		Nombre:      proyecto.Nombre,
		Ubicacion:   proyecto.Ubicacion,
		Descripcion: proyecto.Descripcion,
	}

	proyectoBytes, err := json.Marshal(proyectoParaGuardar)
	if err != nil {
		fmt.Println("Error al serializar proyecto:", err)
		return Proyecto{}
	}

	// Agregar header para indicar que queremos la respuesta con el registro creado
	headers["Prefer"] = "return=representation"

	resp, err := MakeRequest("POST", url, headers, proyectoBytes)
	if err != nil {
		fmt.Println("Error al guardar proyecto:", err)
		return Proyecto{}
	}

	// Intentar deserializar la respuesta
	var proyectosResp []Proyecto
	if err := json.Unmarshal(resp, &proyectosResp); err != nil {
		// Intentar deserializar como objeto único
		var proyectoResp Proyecto
		if err := json.Unmarshal(resp, &proyectoResp); err != nil {
			fmt.Println("Error al procesar respuesta:", err)
			return Proyecto{}
		}
		return proyectoResp
	}

	if len(proyectosResp) > 0 {
		return proyectosResp[0]
	}

	return Proyecto{}
}
