
package adm

import (
	"fmt"
	"strconv"
	"encoding/json"
	"github.com/osmargm1202/orgm/inputs"
)

type Servicio struct {
	ID        int    `json:"id"`
	Nombre    string `json:"nombre"`
	Descripcion string `json:"descripcion"`
}




func BuscarServicio() Servicio {
	var servicio Servicio
	postgrestURL, headers := InitializePostgrest()
	url := postgrestURL + "/servicio?select=*"

	resp, err := MakeRequest("GET", url, headers, nil)
	if err != nil {
		fmt.Println("Error al buscar servicios:", err)
		return servicio
	}

	var servicios []Servicio
	if err := json.Unmarshal(resp, &servicios); err != nil {
		fmt.Println("Error al procesar servicios:", err)
		return servicio
	}

	if len(servicios) == 0 {
		fmt.Println("No se encontraron servicios")
		return servicio
	}

	items := make([]inputs.Item, len(servicios))
	for i, s := range servicios {
		items[i] = inputs.Item{
			ID:   strconv.Itoa(s.ID),
			Name: s.Nombre,
			Desc: s.Descripcion,
		}
	}

	selectedItem := inputs.SelectList("Seleccione un servicio", items)

	id, _ := strconv.Atoi(selectedItem.ID)
	for _, s := range servicios {
		if s.ID == id {
			servicio = s
			break
		}
	}

	return servicio
}