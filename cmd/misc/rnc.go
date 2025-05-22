package misc

import (
	"archive/zip"
	"bufio"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/mattn/go-sqlite3" // Blank import for side effects (driver registration)
	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	// "gorm.io/driver/sqlite"
	// "gorm.io/gorm"
)

// rncCmd represents the rnc command
var RncCmd = &cobra.Command{
	Use:   "rnc",
	Short: "Busqueda de Empresas por RNC",
	Long:  `Busqueda de Empresas por RNC`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("rnc called")
	// },
}

var RncDownloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Descargar Base de Datos",
	Long:  `Descargar Base de Datos de RNC de la DGII`,
	Run: func(cmd *cobra.Command, args []string) {
		extractedFilePath, err := dgiiDownload()
		if err != nil {
			fmt.Println("Error al descargar la base de datos:", err)
			return
		}
		if extractedFilePath == "" { // Download was skipped by user
			fmt.Println("Descarga cancelada por el usuario.")
			return
		}
		err = creacionBd(extractedFilePath)
		if err != nil {
			fmt.Println("Error al crear la base de datos procesada:", err)
		} else {
			fmt.Println("Base de datos procesada creada exitosamente.")
		}
	},
}

var RncSearchCmd = &cobra.Command{
	Use:   "find",
	Short: "find company in database",
	Long:  `find company in database by name or RNC, Razon Social, etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(inputs.TextInput("Ingrese términos de búsqueda:", "Ejemplo: Banco Popular"))
		m, err := p.Run()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		textInputModel, ok := m.(inputs.TextInputModel)
		if !ok {
			fmt.Println("Error: no se pudo obtener el modelo de entrada de texto")
			return
		}

		searchQuery := textInputModel.TextInput.Value()
		if searchQuery == "" {
			fmt.Println("Búsqueda cancelada.")
			return
		}

		results, err := searchInDatabase(searchQuery)
		if err != nil {
			fmt.Println("Error al buscar en la base de datos:", err)
			return
		}

		if len(results) == 0 {
			fmt.Println("No se encontraron resultados para su búsqueda.")
			return
		}

		tableProgram := tea.NewProgram(initialTableModel(results))
		if _, err := tableProgram.Run(); err != nil {
			fmt.Println("Error al mostrar la tabla:", err)
		}

	},
}

func init() {
	MiscCmd.AddCommand(RncCmd)
	RncCmd.AddCommand(RncDownloadCmd)
	RncCmd.AddCommand(RncSearchCmd)
}

// Result structure to hold database query results
type Result struct {
	RNC                  string
	RazonSocial          string
	ActividadEconomica   string
	FechaInicioActividad string
	Estado               string
	RegimenPago          string
}

// Search in the database for records matching any of the search terms
func searchInDatabase(searchQuery string) ([]Result, error) {
	dbPath := filepath.Join(viper.GetString("config_path"), "bd", "dgii.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Split the search query into individual terms
	terms := strings.Fields(searchQuery)
	if len(terms) == 0 {
		return nil, fmt.Errorf("no search terms provided")
	}

	// Build the SQL query to match any of the terms
	query := `SELECT RNC, RazonSocial, ActividadEconomica, FechaInicioActividad, Estado, RegimenPago 
              FROM tabla WHERE `

	var conditions []string
	args := []interface{}{}

	for _, term := range terms {
		termWithWildcards := "%" + term + "%"
		conditions = append(conditions, "(RNC LIKE ? OR RazonSocial LIKE ?)")
		args = append(args, termWithWildcards, termWithWildcards)
	}

	query += strings.Join(conditions, " AND ")
	query += " LIMIT 100" // Limit the results to avoid overwhelming the UI

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	var results []Result
	for rows.Next() {
		var result Result
		if err := rows.Scan(
			&result.RNC,
			&result.RazonSocial,
			&result.ActividadEconomica,
			&result.FechaInicioActividad,
			&result.Estado,
			&result.RegimenPago,
		); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}
// tableModel represents the state of the table UI
type tableModel struct {
	table   table.Model
	results []Result
}

// Initialize the table model with search results
func initialTableModel(results []Result) tableModel {
	columns := []table.Column{
		{Title: "RNC", Width: 15},
		{Title: "Razón Social", Width: 40},
		{Title: "Actividad Económica", Width: 30},
		{Title: "Fecha Inicio", Width: 15},
		{Title: "Estado", Width: 12},
		{Title: "Régimen Pago", Width: 15},
	}

	rows := []table.Row{}
	for _, result := range results {
		rows = append(rows, table.Row{
			result.RNC,
			result.RazonSocial,
			result.ActividadEconomica,
			result.FechaInicioActividad,
			result.Estado,
			result.RegimenPago,
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("12")).
		Bold(false)
	t.SetStyles(s)

	return tableModel{
		table:   t,
		results: results,
	}
}

// Init initializes the table model.
func (m tableModel) Init() tea.Cmd {
	return nil
}

// procesarDatos procesa el RNC y la razón social seleccionados
func (m tableModel) procesarDatos(rnc, razonSocial string) tea.Cmd {
	return tea.Sequence(
		tea.Printf("Procesando datos del contribuyente:"),
		tea.Printf("RNC: %s", rnc),
		tea.Printf("Razón Social: %s", razonSocial),
		// Aquí puedes agregar más lógica de procesamiento
	)
}



// Update handles messages and updates the model.
func (m tableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			selectedRow := m.table.SelectedRow()
			if len(selectedRow) >= 2 {
				return m, m.procesarDatos(selectedRow[0], selectedRow[1])
			}
			return m, tea.Quit
		case "tab":
			selectedRow := m.table.SelectedRow()
			if len(selectedRow) >= 2 {
				return m, tea.Printf("%s - %s", selectedRow[0], selectedRow[1])
			}
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the table.
func (m tableModel) View() string {
	return lipgloss.NewStyle().Margin(1, 1).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			"Resultados de búsqueda - Presione 'q' o 'esc' para salir \n'enter' para procesar el registro seleccionado\n",
			m.table.View(),
			fmt.Sprintf("Total de resultados: %d", len(m.results)),
		),
	)
}

func dgiiDownload() (string, error) {
	url := viper.GetString("url.dgii")
	if url == "" {
		return "", fmt.Errorf("URL not found in config")
	}

	carpeta := viper.GetString("config_path") + "/bd"
	if err := os.MkdirAll(carpeta, 0755); err != nil {
		return "", fmt.Errorf("failed to create bd directory: %v", err)
	}

	// Tentative name for the extracted file (assuming it's a .txt or .csv, DGII often uses .txt)
	// We will confirm the actual name after extraction.
	// For the check, we'll look for any .txt or .csv files in the bd/TMP directory.
	tmpDir := filepath.Join(carpeta, "TMP")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create TMP directory: %v", err)
	}

	files, _ := filepath.Glob(filepath.Join(tmpDir, "*.txt"))
	csvFiles, _ := filepath.Glob(filepath.Join(tmpDir, "*.csv"))
	files = append(files, csvFiles...)

	if len(files) > 0 {
		fmt.Printf("File exists (%s). Do you want to continue and overwrite it? (y/n): ", filepath.Base(files[0]))
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input != "s" && input != "si" && input != "y" && input != "yes" {
			return csvFiles[0], nil // Return empty path and no error to indicate user cancellation
		}
	}

	// Create a channel to signal when download is complete
	done := make(chan struct{})

	// Start spinner in goroutine
	go func() {
		p := tea.NewProgram(inputs.InitialSpinnerModel("Descargando Base de Datos...DGII"))
		go func() {
			<-done // Wait for download to complete
			p.Quit()
		}()
		if _, err := p.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	// Download file
	resp, err := http.Get(url)
	if err != nil {
		close(done) // Signal spinner on early error
		return "", fmt.Errorf("failed to initiate download: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		close(done) // Signal spinner if status is not OK
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create zip file
	zipFilePath := filepath.Join(carpeta, "DGII_RNC.zip")
	out, err := os.Create(zipFilePath)
	if err != nil {
		close(done) // Signal spinner on early error
		// fmt.Println(url) // Already handled by spinner error message
		return "", fmt.Errorf("failed to create zip file: %v", err)
	}
	defer out.Close()

	// Write the body to file in chunks
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		close(done) // Signal spinner on early error
		return "", fmt.Errorf("failed to write zip file: %v", err)
	}

	// Extract zip file
	zipReader, err := zip.OpenReader(zipFilePath)
	if err != nil {
		close(done) // Signal spinner on early error
		return "", fmt.Errorf("failed to open zip file: %v", err)
	}
	defer zipReader.Close()

	var extractedFilePath string // Use a single variable to store the final path

	for _, file := range zipReader.File {
		// Open the file from zip
		zippedFile, err := file.Open()
		if err != nil {
			close(done)
			return "", fmt.Errorf("failed to open zipped file from archive: %v", err)
		}
		defer zippedFile.Close()

		// Create TMP directory inside bd for extracted files
		// Ensures that we don't overwrite existing bd/activo.csv or bd/suspendido.csv prematurely
		targetPath := filepath.Join(tmpDir, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(targetPath, file.Mode())
		} else {
			// Ensure the directory for the file exists
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				close(done)
				return "", fmt.Errorf("failed to create directory for extracted file: %v", err)
			}

			outputFile, err := os.OpenFile(
				targetPath,
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				file.Mode(),
			)
			if err != nil {
				close(done) // Signal spinner on early error
				return "", fmt.Errorf("failed to create extracted file: %v", err)
			}
			// defer outputFile.Close() // Close inside the loop

			_, err = io.Copy(outputFile, zippedFile)
			outputFile.Close() // Close immediately after copy
			if err != nil {
				close(done) // Signal spinner on early error
				return "", fmt.Errorf("failed to write extracted file: %v", err)
			}
			// Assuming the first non-directory file is the one we want
			if extractedFilePath == "" && !file.FileInfo().IsDir() {
				extractedFilePath = targetPath
			}
		}
	}

	close(done) // NOW signal the spinner to stop

	if extractedFilePath == "" {
		return "", fmt.Errorf("no file was extracted from the zip archive")
	}
	return extractedFilePath, nil
}

func creacionBd(sourceCsvPath string) error {
	carpetaBd := filepath.Join(viper.GetString("config_path"), "bd")
	dbPath := filepath.Join(carpetaBd, "dgii.db")

	// Remove existing database file to start fresh
	_ = os.Remove(dbPath) // Ignore error if file doesn't exist

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to connect to database %s: %v", dbPath, err)
	}
	defer db.Close()

	// Tune SQLite for performance
	pragmas := []string{
		"PRAGMA synchronous = OFF;",
		"PRAGMA journal_mode = MEMORY;",
		"PRAGMA temp_store = MEMORY;",
	}
	for _, p := range pragmas {
		_, err = db.Exec(p)
		if err != nil {
			return fmt.Errorf("failed to execute pragma '%s': %v", p, err)
		}
	}

	// Create table
	createTableSQL := `CREATE TABLE IF NOT EXISTS tabla (
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		RNC TEXT NOT NULL,
		RazonSocial TEXT NOT NULL,
		ActividadEconomica TEXT,
		FechaInicioActividad TEXT,
		Estado TEXT,
		RegimenPago TEXT
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	// Create index on RazonSocial after table creation
	createIndexSQL := `CREATE INDEX IF NOT EXISTS idx_razonsocial ON tabla (RazonSocial);`
	_, err = db.Exec(createIndexSQL)
	if err != nil {
		return fmt.Errorf("failed to create index on RazonSocial: %v", err)
	}

	inputFile, err := os.Open(sourceCsvPath)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo CSV fuente (%s): %v", sourceCsvPath, err)
	}
	defer inputFile.Close()

	csvReader := csv.NewReader(inputFile)
	csvReader.Comma = ','          // Assuming the DGII file is comma-separated
	csvReader.LazyQuotes = true    // Handles non-standard quotes
	csvReader.FieldsPerRecord = -1 // Allow variable number of fields

	lineNumber := 0
	isFirstLine := true
	batchSize := 10000 // Commit every 10,000 records
	rowCountInBatch := 0

	done := make(chan struct{})
	spinnerActive := true // To control spinner message update

	p := tea.NewProgram(inputs.InitialSpinnerModel(fmt.Sprintf("Procesando %s...", filepath.Base(sourceCsvPath))))
	go func() {
		if _, err := p.Run(); err != nil {
			fmt.Printf("Spinner error: %v\n", err)
		}
	}()

	defer func() {
		if spinnerActive {
			close(done) // Ensure spinner stops if function exits early due to error
			p.Quit()    // Quit the bubbletea program
		}
	}()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	stmt, err := tx.Prepare(`INSERT INTO tabla (RNC, RazonSocial, ActividadEconomica, FechaInicioActividad, Estado, RegimenPago) VALUES (?, ?, ?, ?, ?, ?)`) // Ensure table name matches CREATE TABLE
	if err != nil {
		tx.Rollback() // Rollback on prepare error
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	for {
		lineNumber++
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			if perr, ok := err.(*csv.ParseError); ok && (perr.Err == csv.ErrFieldCount || perr.Err == csv.ErrBareQuote || perr.Err == csv.ErrQuote) {
				fmt.Printf("Advertencia (Línea %d): Error de parseo CSV: %v. Contenido: %v. Saltando línea.\n", lineNumber, perr, row)
				continue
			}
			// For other read errors, it might be better to stop and report
			tx.Rollback()
			return fmt.Errorf("error al leer el archivo CSV (Línea %d): %v", lineNumber, err)
		}

		if isFirstLine {
			isFirstLine = false
			if len(row) > 0 && (strings.ToUpper(strings.TrimSpace(row[0])) == "RNC" || strings.ToUpper(strings.TrimSpace(row[0])) == "CEDULA") {
				fmt.Printf("Cabecera detectada y omitida (Línea %d): %v\n", lineNumber, row)
				continue
			}
		}

		// Ensure row has enough columns before trying to access them
		if len(row) < 6 {
			fmt.Printf("Advertencia (Línea %d): Fila con menos de 6 columnas. Contenido: %v. Saltando línea.\n", lineNumber, row)
			continue
		}

		rncStr := strings.TrimSpace(row[0])
		razonSocial := strings.TrimSpace(row[1])
		actividadEconomica := strings.TrimSpace(row[2])
		fechaInicioActividad := strings.TrimSpace(row[3])
		estado := strings.TrimSpace(row[4])
		regimenPago := strings.TrimSpace(row[5])

		_, err = stmt.Exec(rncStr, razonSocial, actividadEconomica, fechaInicioActividad, estado, regimenPago)
		if err != nil {
			// Log detailed error, including the data that failed
			fmt.Printf("Error (Línea %d, RNC %s): Al guardar registro: %v. Datos: [%s, %s, ...]. Saltando.\n", lineNumber, rncStr, err, rncStr, razonSocial)
			// Decide if you want to rollback the whole transaction or just skip this record
			// For now, we skip and let the batch commit later.
			continue
		}

		rowCountInBatch++
		if rowCountInBatch >= batchSize {
			err = tx.Commit()
			if err != nil {
				tx.Rollback() // Attempt rollback on commit error
				return fmt.Errorf("failed to commit transaction after %d rows: %v", rowCountInBatch, err)
			}
			fmt.Printf("Lote de %d registros procesado (Total líneas: %d).\n", rowCountInBatch, lineNumber)
			rowCountInBatch = 0
			tx, err = db.Begin() // Start a new transaction
			if err != nil {
				return fmt.Errorf("failed to begin new transaction: %v", err)
			}
			// Re-prepare statement for the new transaction
			stmt, err = tx.Prepare(`INSERT INTO tabla (RNC, RazonSocial, ActividadEconomica, FechaInicioActividad, Estado, RegimenPago) VALUES (?, ?, ?, ?, ?, ?)`) // Ensure table name matches
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to re-prepare statement: %v", err)
			}
		}
	}

	// Commit any remaining records in the last batch
	if rowCountInBatch > 0 {
		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to commit final transaction: %v", err)
		}
		fmt.Printf("Lote final de %d registros procesado (Total líneas: %d).", rowCountInBatch, lineNumber)
	}

	spinnerActive = false
	close(done)
	p.Quit() // Ensure bubbletea program is quit

	fmt.Printf("%s\n", inputs.InfoStyle.Render(fmt.Sprintf("Procesamiento de %s completado. Total líneas leídas: %d.", filepath.Base(sourceCsvPath), lineNumber)))
	return nil
}
