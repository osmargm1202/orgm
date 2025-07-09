
package adm

import (
	"fmt"
	"encoding/json"
)


// ObtenerProximoIDPresupuesto consulta el ID máximo existente de presupuesto y devuelve el siguiente
func ObtenerProximoIDPresupuesto() (int, error) {
	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		return 0, fmt.Errorf("URL de PostgREST no configurada")
	}

	// Consultar el presupuesto con el ID más alto
	url := postgrestURL + "/presupuesto?select=id&order=id.desc&limit=1"

	resp, err := MakeRequest("GET", url, headers, nil)
	if err != nil {
		return 0, fmt.Errorf("error al consultar ID máximo de presupuesto: %w", err)
	}

	// Imprimir respuesta para depuración
	fmt.Println("Respuesta al buscar ID máximo de presupuesto:", string(resp))

	// Si no hay datos, comenzar desde 1
	if len(resp) == 0 || string(resp) == "[]" || string(resp) == "null" {
		fmt.Println("No se encontraron presupuestos existentes, iniciando desde ID 1")
		return 1, nil
	}

	// Intentar deserializar como array
	var presupuestos []map[string]interface{}
	err = json.Unmarshal(resp, &presupuestos)

	if err != nil {
		// Si falla como array, intentar como objeto único
		var presupuesto map[string]interface{}
		err = json.Unmarshal(resp, &presupuesto)
		if err != nil {
			return 0, fmt.Errorf("error al procesar respuesta JSON: %w", err)
		}

		// Extraer ID del objeto único, manejando diferentes tipos
		return extraerID(presupuesto["id"])
	}

	if len(presupuestos) == 0 {
		return 1, nil
	}

	// Extraer ID del primer elemento del array
	proximoID, err := extraerID(presupuestos[0]["id"])
	if err != nil {
		return 0, err
	}

	fmt.Println("Próximo ID de presupuesto:", proximoID)
	return proximoID, nil
}
