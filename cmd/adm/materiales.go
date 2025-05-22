package adm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"


	tea "github.com/charmbracelet/bubbletea"
	"github.com/olekukonko/tablewriter"
	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
)

// ResourceType representa el tipo de recurso a buscar
type ResourceType string

const (
	MaterialType    ResourceType = "material"
	ServicioType    ResourceType = "servicios"
	ManoDeObraType  ResourceType = "manodeobra"
	EquipoType      ResourceType = "equipos"
	HerramientaType ResourceType = "herramientas"
	IndirectoType   ResourceType = "indirectos"
)

// Resource representa cualquier tipo de recurso que se puede agregar a un presupuesto
type Resource struct {
	ID          int                    `json:"id"`
	Nombre      string                 `json:"nombre"`
	Descripcion map[string]interface{} `json:"descripcion"`
	Unidad      string                 `json:"unidad"`
	Fecha       string                 `json:"fecha"`
}

var buscarCmd = &cobra.Command{
	Use:   "mat",
	Short: "Buscar Recursos",
	Long:  `Buscar Recursos en la base de datos`,
	Run: func(cmd *cobra.Command, args []string) {
		recursos, err := BuscarRecursos(strings.Join(args, " "))
		if err != nil {
			fmt.Println(err)
		}

		printTable(recursos)

	},
}

func init() {
	AdmCmd.AddCommand(buscarCmd)
}


func printTable(recursos []Resource) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"ID", "Nombre", "moneda", "Precio", "Unidad"})
	for _, recurso := range recursos {
		monedas := recurso.Descripcion["monedas"].(map[string]interface{})
		fmt.Println(monedas)
		table.Append([]string{fmt.Sprintf("%d", recurso.ID), recurso.Nombre, "RD$", fmt.Sprintf("%v", monedas["RD"]), recurso.Unidad})
	}
	table.Render()
}

// BuscarRecursos busca recursos del tipo especificado que coincidan con el nombre
func BuscarRecursos(nombreBusqueda string) ([]Resource, error) {
	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		return nil, fmt.Errorf("URL de PostgREST no disponible")
	}

	// Crear lista de tipos de recursos
	tiposList := []string{
		string(MaterialType),
		string(ServicioType),
		string(ManoDeObraType),
		string(EquipoType),
		string(HerramientaType),
		string(IndirectoType),
	}

	// Inicializar modelo de selección
	model := inputs.SelectionModel(
		tiposList,
		"Seleccione el tipo de recurso",
		"Use las flechas para navegar y Enter para seleccionar",
	)

	// Ejecutar programa de selección
	p := tea.NewProgram(model)
	m, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar selector: %v", err)
	}

	// Obtener selección
	finalModel := m.(inputs.SelectionModelS)
	if finalModel.Quitting {
		return nil, fmt.Errorf("operación cancelada por el usuario")
	}

	// Convertir selección a ResourceType
	tipo := ResourceType(finalModel.Choices[finalModel.Cursor].Title)

	// Construir URL con la consulta
	queryURL := fmt.Sprintf("%s/%s?nombre=ilike.*%s*&limit=20",
		postgrestURL,
		url.PathEscape(string(tipo)),
		url.QueryEscape(nombreBusqueda))

	// Crear la solicitud
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error al crear solicitud: %v", err)
	}

	// Agregar encabezados
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	// Enviar la solicitud
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error al enviar solicitud: %v", err)
	}
	defer resp.Body.Close()

	// Leer la respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error al leer respuesta: %v", err)
	}

	// Verificar código de estado
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error del servidor: %s", string(body))
	}

	// Parsear los resultados
	var recursos []Resource
	if err := json.Unmarshal(body, &recursos); err != nil {
		return nil, fmt.Errorf("error al parsear respuesta: %v", err)
	}

	return recursos, nil
}



