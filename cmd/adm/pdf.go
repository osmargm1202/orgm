package adm

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	pdfFlag bool
)

var PdfCmd = &cobra.Command{
	Use:   "pdf",
	Short: "Print a PDF",
	Long:  `Print a PDF from the API`,
}

var PrintFacturaCmd = &cobra.Command{
	Use:   "invoice [id]",
	Short: "Print an Invoice",
	Long:  `Print an Invoice from the API`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Error: Invoice ID is required")
			return
		}
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Error: Invoice ID must be an integer")
			return
		}
		pdf, _ := cmd.Flags().GetBool("pdf")

		filePath, err := GetFactura(id, pdf)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if pdf {
			fmt.Printf("Invoice %d saved as %s\n", id, filePath)
		} else {
			fmt.Printf("Invoice %d saved as %s\n", id, filePath)
		}
	},
}


var PrintCotizacionCmd = &cobra.Command{
	Use:   "quotation [id]",
	Short: "Print a Quotation",
	Long:  `Print a Quotation from the API`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Error: Quotation ID is required")
			return
		}
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Error: Quotation ID must be an integer")
			return
		}
		pdf, _ := cmd.Flags().GetBool("pdf")

		filePath, err := GetCotizacion(id, pdf)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if pdf {
			fmt.Printf("Quotation %d saved as %s\n", id, filePath)
		} else {
			fmt.Printf("Quotation %d saved as %s\n", id, filePath)
		}
	},
}

func init() {
	AdmCmd.AddCommand(PdfCmd)
	PrintFacturaCmd.Flags().BoolVarP(&pdfFlag, "pdf", "p", false, "Convert output to PDF (LibreOffice is required)")
	PdfCmd.AddCommand(PrintFacturaCmd)
	PrintCotizacionCmd.Flags().BoolVarP(&pdfFlag, "pdf", "p", false, "Convert output to PDF (LibreOffice is required)")
	PdfCmd.AddCommand(PrintCotizacionCmd)
}

func moveFileToDesktop(sourcePath string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting user home directory: %w", err)
	}
	desktopPath := filepath.Join(homeDir, "Desktop")
	fileName := filepath.Base(sourcePath)
	destPath := filepath.Join(desktopPath, fileName)

	// Create Desktop directory if it doesn't exist
	if _, err := os.Stat(desktopPath); os.IsNotExist(err) {
		err = os.MkdirAll(desktopPath, 0755)
		if err != nil {
			return "", fmt.Errorf("error creating desktop directory %s: %w", desktopPath, err)
		}
	}

	err = os.Rename(sourcePath, destPath)
	if err != nil {
		// If rename fails (e.g. cross-device link), try copy and delete
		in, err := os.Open(sourcePath)
		if err != nil {
			return "", fmt.Errorf("error opening source file for copy %s: %w", sourcePath, err)
		}
		defer in.Close()

		out, err := os.Create(destPath)
		if err != nil {
			return "", fmt.Errorf("error creating destination file for copy %s: %w", destPath, err)
		}
		defer out.Close()

		_, err = io.Copy(out, in)
		if err != nil {
			return "", fmt.Errorf("error copying file from %s to %s: %w", sourcePath, destPath, err)
		}
		out.Close() // Ensure file is closed before removing source

		err = os.Remove(sourcePath)
		if err != nil {
			// Log error but don't fail the whole operation if copy was successful
			fmt.Printf("Warning: could not delete original file after copying %s: %v\n", sourcePath, err)
		}
	}
	return destPath, nil
}

func convertToPDF(docxPath string) (string, error) {
	if _, err := exec.LookPath("libreoffice"); err != nil {
		return "", fmt.Errorf("libreoffice not found in PATH. Please install LibreOffice to use the PDF conversion feature: %w", err)
	}

	// Convert in the current directory first, then move
	outDir := filepath.Dir(docxPath)
	cmd := exec.Command("libreoffice", "--headless", "--convert-to", "pdf", docxPath, "--outdir", outDir)
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error converting to PDF: %w", err)
	}

	pdfFileName := filepath.Base(docxPath[0:len(filepath.Base(docxPath))-len(filepath.Ext(docxPath))] + ".pdf")
	generatedPdfPath := filepath.Join(outDir, pdfFileName)

	// Verify PDF creation
	if _, err := os.Stat(generatedPdfPath); os.IsNotExist(err) {
		return "", fmt.Errorf("PDF file not found after conversion: %s", generatedPdfPath)
	}

	// Delete the original DOCX file
	err = os.Remove(docxPath)
	if err != nil {
		// Log error but don't fail the whole operation if PDF was created
		fmt.Printf("Warning: could not delete original DOCX file %s: %v\n", docxPath, err)
	}
	return generatedPdfPath, nil // Return path of PDF in current dir, to be moved later
}

func GetCotizacion(id int, convertToPdfFlag bool) (string, error) {
	apiURL, headers := InitializeApi()
	if apiURL == "" {
		return "", fmt.Errorf("error initializing API URL")
	}

	url := fmt.Sprintf("%s/cot/%d", apiURL, id)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error response status: %d", resp.StatusCode)
	}

	// Create output file in current directory first
	localFilename := fmt.Sprintf("cotizacion_%d.docx", id)
	out, err := os.Create(localFilename)
	if err != nil {
		return "", fmt.Errorf("error creating file: %v", err)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		out.Close()              // Close before attempting to remove on error
		os.Remove(localFilename) // Clean up partially written file
		return "", fmt.Errorf("error writing file: %v", err)
	}
	out.Close() // Close the file before conversion or moving

	finalFilePath := localFilename

	if convertToPdfFlag {
		pdfPath, err := convertToPDF(localFilename)
		if err != nil {
			// Don't remove localFilename here, as convertToPDF might have already removed it or user might want the docx
			return "", fmt.Errorf("error converting quotation to PDF: %w", err)
		}
		finalFilePath = pdfPath
	}

	desktopPath, err := moveFileToDesktop(finalFilePath)
	if err != nil {
		// If move fails, return the path in the current directory as a fallback
		fmt.Printf("Warning: could not move file to desktop: %v. File saved at %s\n", err, finalFilePath)
		return finalFilePath, nil
	}

	return desktopPath, nil
}

func GetFactura(id int, convertToPdfFlag bool) (string, error) {
	apiURL, headers := InitializeApi()
	if apiURL == "" {
		return "", fmt.Errorf("error initializing API URL")
	}

	url := fmt.Sprintf("%s/fac/%d", apiURL, id)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error response status: %d", resp.StatusCode)
	}

	// Create output file in current directory first
	localFilename := fmt.Sprintf("factura_%d.docx", id)
	out, err := os.Create(localFilename)
	if err != nil {
		return "", fmt.Errorf("error creating file: %v", err)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		out.Close()              // Close before attempting to remove on error
		os.Remove(localFilename) // Clean up partially written file
		return "", fmt.Errorf("error writing file: %v", err)
	}
	out.Close() // Close the file before conversion or moving

	finalFilePath := localFilename

	if convertToPdfFlag {
		pdfPath, err := convertToPDF(localFilename)
		if err != nil {
			// Don't remove localFilename here, as convertToPDF might have already removed it or user might want the docx
			return "", fmt.Errorf("error converting invoice to PDF: %w", err)
		}
		finalFilePath = pdfPath
	}

	desktopPath, err := moveFileToDesktop(finalFilePath)
	if err != nil {
		// If move fails, return the path in the current directory as a fallback
		fmt.Printf("Warning: could not move file to desktop: %v. File saved at %s\n", err, finalFilePath)
		return finalFilePath, nil
	}

	return desktopPath, nil
}
