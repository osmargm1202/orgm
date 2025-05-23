package adm

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/osmargm1202/orgm/inputs"
)

type Ubicacion struct {
	ID                int    `json:"id"`
	Provincia         string `json:"provincia"`
	Distrito          string `json:"distrito"`
	DistritoMunicipal string `json:"distritomunicipal"`
}

// GetFromDatabase obtiene una ubicación específica por ID desde la base de datos
func (u *Ubicacion) GetFromDatabase() bool {
	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		fmt.Println("Error: URL de PostgREST no configurada")
		return false
	}

	url := fmt.Sprintf("%s/ubicacion?id=eq.%d", postgrestURL, u.ID)

	resp, err := MakeRequest("GET", url, headers, nil)
	if err != nil {
		fmt.Printf("Error al conectar con la base de datos: %v\n", err)
		return false
	}

	var ubicaciones []Ubicacion
	if err := json.Unmarshal(resp, &ubicaciones); err != nil {
		fmt.Printf("Error al procesar respuesta: %v\n", err)
		return false
	}

	if len(ubicaciones) > 0 {
		locationData := ubicaciones[0]
		u.Provincia = locationData.Provincia
		u.Distrito = locationData.Distrito
		u.DistritoMunicipal = locationData.DistritoMunicipal
		return true
	}

	return false
}

// GetCombinedString devuelve una representación combinada de la ubicación
func (u *Ubicacion) GetCombinedString() string {
	var parts []string

	if u.DistritoMunicipal != "" {
		parts = append(parts, u.DistritoMunicipal)
	}
	if u.Distrito != "" {
		parts = append(parts, u.Distrito)
	}
	if u.Provincia != "" {
		parts = append(parts, u.Provincia)
	}

	return strings.Join(parts, ", ")
}

// BuscarUbicacion permite buscar y seleccionar una ubicación de la base de datos
func BuscarUbicacion() Ubicacion {
	var ubicacion Ubicacion

	// Usar el componente de selección para elegir el método de búsqueda
	items := []inputs.Item{
		{ID: "search", Name: "Buscar por términos", Desc: "Buscar por provincia, distrito o distrito municipal"},
		{ID: "id", Name: "Buscar por ID", Desc: "Buscar directamente por el ID de la ubicación"},
	}

	selectedMethod := inputs.SelectList("Seleccione método de búsqueda de ubicación", items)

	if selectedMethod.ID == "id" {
		// Buscar por ID
		idStr := inputs.GetInput("Ingrese el ID de la ubicación")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			fmt.Println("ID inválido")
			return ubicacion
		}

		ubicacion.ID = id
		if ubicacion.GetFromDatabase() {
			return ubicacion
		} else {
			fmt.Println("No se encontró ubicación con ID:", idStr)
			return Ubicacion{}
		}
	} else {
		// Buscar por términos
		query := inputs.GetInput("Ingrese término de búsqueda (provincia, distrito, distrito municipal): ")

		if query == "" {
			fmt.Println("Búsqueda vacía")
			return ubicacion
		}

		ubicaciones := SearchUbicaciones(query)
		if len(ubicaciones) == 0 {
			fmt.Println("No se encontraron ubicaciones")
			return ubicacion
		}

		// Crear items para la selección
		items := make([]inputs.Item, len(ubicaciones))
		for i, ub := range ubicaciones {
			items[i] = inputs.Item{
				ID:   strconv.Itoa(ub.ID),
				Name: ub.GetCombinedString(),
				Desc: fmt.Sprintf("Provincia: %s", ub.Provincia),
			}
		}

		selectedItem := inputs.SelectList("Seleccione una ubicación", items)

		id, _ := strconv.Atoi(selectedItem.ID)
		for _, ub := range ubicaciones {
			if ub.ID == id {
				ubicacion = ub
				break
			}
		}
	}

	return ubicacion
}

// SearchUbicaciones busca ubicaciones por términos en provincia, distrito o distrito municipal
func SearchUbicaciones(query string) []Ubicacion {
	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		fmt.Println("Error: URL de PostgREST no configurada")
		return nil
	}

	// Escapar el query parameter correctamente para URLs
	escapedQuery := url.QueryEscape(query)

	// Buscar en provincia, distrito o distritomunicipal usando ilike
	requestURL := fmt.Sprintf("%s/ubicacion?select=*&or=(provincia.ilike.*%s*,distrito.ilike.*%s*,distritomunicipal.ilike.*%s*)",
		postgrestURL, escapedQuery, escapedQuery, escapedQuery)

	resp, err := MakeRequest("GET", requestURL, headers, nil)
	if err != nil {
		fmt.Printf("Error al buscar ubicaciones: %v\n", err)
		return nil
	}

	var ubicaciones []Ubicacion
	if err := json.Unmarshal(resp, &ubicaciones); err != nil {
		fmt.Printf("Error al procesar ubicaciones: %v\n", err)
		return nil
	}

	return ubicaciones
}

// GetAllUbicaciones obtiene todas las ubicaciones de la base de datos
func GetAllUbicaciones() []Ubicacion {
	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		fmt.Println("Error: URL de PostgREST no configurada")
		return nil
	}

	url := postgrestURL + "/ubicacion?select=*&order=provincia,distrito,distritomunicipal"

	resp, err := MakeRequest("GET", url, headers, nil)
	if err != nil {
		fmt.Printf("Error al obtener ubicaciones: %v\n", err)
		return nil
	}

	var ubicaciones []Ubicacion
	if err := json.Unmarshal(resp, &ubicaciones); err != nil {
		fmt.Printf("Error al procesar ubicaciones: %v\n", err)
		return nil
	}

	return ubicaciones
}
