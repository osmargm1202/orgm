package misc

import (
	"archive/zip"
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

func init() {
	MiscCmd.AddCommand(RncCmd)
	RncCmd.AddCommand(RncDownloadCmd)
}

type Tabla struct {
	ID                   uint   `gorm:"primaryKey;autoIncrement"`
	RNC                  string `gorm:"not null"`
	RazonSocial          string `gorm:"not null;index"`
	ActividadEconomica   string
	FechaInicioActividad string
	Estado               string
	RegimenPago          string
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

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database %s: %v", dbPath, err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&Tabla{})
	if err != nil {
		return fmt.Errorf("failed to auto-migrate database schema: %v", err)
	}

	inputFile, err := os.Open(sourceCsvPath)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo CSV fuente (%s): %v", sourceCsvPath, err)
	}
	defer inputFile.Close()

	csvReader := csv.NewReader(inputFile)
	csvReader.Comma = ','
	csvReader.LazyQuotes = true
	csvReader.FieldsPerRecord = -1



	lineNumber := 0

	isFirstLine := true

	done := make(chan struct{})

	go func() {
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
			return fmt.Errorf("error al leer el archivo CSV (Línea %d): %v", lineNumber, err)
		}

		if isFirstLine {
			isFirstLine = false
			// Trim spaces from the first cell for robust header check
			if len(row) > 0 && (strings.ToUpper(strings.TrimSpace(row[0])) == "RNC" || strings.ToUpper(strings.TrimSpace(row[0])) == "CEDULA") {
				fmt.Printf("Cabecera detectada y omitida (Línea %d): %v\n", lineNumber, row)
				continue
			}
			// If it wasn't a header, this line should be processed, so fall through.
		}


		rncStr := strings.TrimSpace(row[0])
		razonSocial := strings.TrimSpace(row[1])
		actividadEconomica := strings.TrimSpace(row[2])
		fechaInicioActividad := strings.TrimSpace(row[3])
		estado := strings.TrimSpace(row[4])
		regimenPago := strings.TrimSpace(row[5])


		activo := Tabla{
			RNC:                  rncStr,
			RazonSocial:          razonSocial,
			ActividadEconomica:   actividadEconomica,
			FechaInicioActividad: fechaInicioActividad,
			Estado:               estado,
			RegimenPago:          regimenPago,
		}
		result := db.Create(&activo)
		if result.Error != nil {
			// Log detailed error, including the data that failed
			fmt.Printf("Error (Línea %d, RNC %s): Al guardar registro activo: %v. Datos: %+v. Saltando.\n", lineNumber, rncStr, result.Error, activo)
			continue
		}
		}
	}()

	<-done

	fmt.Printf("%s\n", inputs.InfoStyle.Render(fmt.Sprintf("Procesamiento de %s completado.", filepath.Base(sourceCsvPath))))



	return nil
}
