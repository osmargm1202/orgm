package adm

import (
	"fmt"
	"log"
	"time"

	"github.com/johnfercher/maroto/v2"

	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"

	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CotizacionData represents the quotation data
type CotizacionData struct {
	// Company info
	CompanyName    string
	CompanyAddress string
	CompanyPhone   string
	CompanyEmail   string
	RNC            string

	// Client info
	ClientName    string
	ClientContact string
	ClientRNC     string
	ClientAddress string

	// Quotation info
	QuotationNumber string
	ClientCode      string
	Category        string
	Date            string
	Project         string
	Location        string
	Service         string
	Description     string

	// Items
	Items []CotizacionItem

	// Financial
	Subtotal  float64
	ITBIS     float64
	Retencion float64
	Total     float64

	// Logo
	ClientID string
}

type CotizacionItem struct {
	Item        string
	Description string
	Quantity    int
	Unit        string
	UnitPrice   float64
	TotalPrice  float64
}

// Sample data - this should come from parameters in a real implementation
func getSampleCotizacionData() CotizacionData {
	return CotizacionData{
		CompanyName:    "ORGM",
		CompanyAddress: "Av. 27 de Febrero #596, Sector Renacimiento, Distrito Nacional",
		CompanyPhone:   "809-405-9420 / 829-988-3375",
		CompanyEmail:   "adm@orgm.com / info@orgm.com",
		RNC:            "1-31-91523-1",

		ClientName:    "DAPEC",
		ClientContact: "OSMAR GARCIA",
		ClientRNC:     "131351247",
		ClientAddress: "OSMAR GARCIA",

		QuotationNumber: "0534",
		ClientCode:      "0003",
		Category:        "DSEAT",
		Date:            "22-05-2025",
		Project:         "SEAT SANTA CLARA 138KV PARA PFV",
		Location:        "SANTO DOMINGO, DISTRITO NACIONAL",
		Service:         "DISEÑO DE SUBESTACIÓN ELÉCTRICA DE ALTA TENSIÓN",
		Description:     "DISEÑO DE SUBESTACION DE ALTA TENSION INCLUYENDO OBRA CIVIL Y ESTUDIO DE SUELO",

		Items: []CotizacionItem{
			{
				Item:        "P-1",
				Description: "DISEÑO DE SUBESTACION DE ALTA TENSION INCLUYENDO OBRA CIVIL Y ESTUDIO DE SUELO",
				Quantity:    1,
				Unit:        "Ud",
				UnitPrice:   854351.69,
				TotalPrice:  854351.69,
			},
		},

		Subtotal:  854351.69,
		ITBIS:     153783.30,
		Retencion: 46134.99,
		Total:     962000.00,

		ClientID: "0003",
	}
}

// GenerateInvoiceCmd represents the command to generate an invoice PDF
var generateFacturaCmd = &cobra.Command{
	Use:   "pdf",
	Short: "Generate a quotation PDF",
	Long:  `Generate a professional quotation PDF for ORGM with customizable client and project information.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flag values
		clientName, _ := cmd.Flags().GetString("client-name")
		clientContact, _ := cmd.Flags().GetString("client-contact")
		clientRNC, _ := cmd.Flags().GetString("client-rnc")
		clientID, _ := cmd.Flags().GetString("client-id")
		quotationNumber, _ := cmd.Flags().GetString("quotation-number")
		project, _ := cmd.Flags().GetString("project")
		service, _ := cmd.Flags().GetString("service")
		description, _ := cmd.Flags().GetString("description")
		unitPrice, _ := cmd.Flags().GetFloat64("unit-price")
		outputFile, _ := cmd.Flags().GetString("output")

		// Create cotizacion data with flag values or defaults
		data := CotizacionData{
			CompanyName:    "ORGM",
			CompanyAddress: "Av. 27 de Febrero #596, Sector Renacimiento, Distrito Nacional",
			CompanyPhone:   "809-405-9420 / 829-988-3375",
			CompanyEmail:   "adm@orgm.com / info@orgm.com",
			RNC:            "1-31-91523-1",

			ClientName:    getStringOrDefault(clientName, "DAPEC"),
			ClientContact: getStringOrDefault(clientContact, "OSMAR GARCIA"),
			ClientRNC:     getStringOrDefault(clientRNC, "131351247"),
			ClientAddress: getStringOrDefault(clientContact, "OSMAR GARCIA"),

			QuotationNumber: getStringOrDefault(quotationNumber, "0534"),
			ClientCode:      getStringOrDefault(clientID, "0003"),
			Category:        "DSEAT",
			Date:            time.Now().Format("02-01-2006"),
			Project:         getStringOrDefault(project, "SEAT SANTA CLARA 138KV PARA PFV"),
			Location:        "SANTO DOMINGO, DISTRITO NACIONAL",
			Service:         getStringOrDefault(service, "DISEÑO DE SUBESTACIÓN ELÉCTRICA DE ALTA TENSIÓN"),
			Description:     getStringOrDefault(description, "DISEÑO DE SUBESTACION DE ALTA TENSION INCLUYENDO OBRA CIVIL Y ESTUDIO DE SUELO"),

			Items: []CotizacionItem{
				{
					Item:        "P-1",
					Description: getStringOrDefault(description, "DISEÑO DE SUBESTACION DE ALTA TENSION INCLUYENDO OBRA CIVIL Y ESTUDIO DE SUELO"),
					Quantity:    1,
					Unit:        "Ud",
					UnitPrice:   getFloat64OrDefault(unitPrice, 854351.69),
					TotalPrice:  getFloat64OrDefault(unitPrice, 854351.69),
				},
			},

			ClientID: getStringOrDefault(clientID, "0003"),
		}

		// Calculate financial values
		data.Subtotal = data.Items[0].TotalPrice
		data.ITBIS = data.Subtotal * 0.18
		data.Retencion = data.Subtotal * 0.054
		data.Total = data.Subtotal + data.ITBIS - data.Retencion

		m := GetMarotoWithData(data)
		document, err := m.Generate()
		if err != nil {
			log.Fatal(err.Error())
		}

		fileName := getStringOrDefault(outputFile, "cotizacion.pdf")
		err = document.Save(fileName)
		if err != nil {
			log.Fatal(err.Error())
		}

		reportFileName := fileName[:len(fileName)-4] + ".txt"
		err = document.GetReport().Save(reportFileName)
		if err != nil {
			log.Fatal(err.Error())
		}

		log.Printf("PDF generated successfully: %s", fileName)
		return nil
	},
}

func getStringOrDefault(value, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}

func getFloat64OrDefault(value, defaultValue float64) float64 {
	if value != 0 {
		return value
	}
	return defaultValue
}

func init() {
	AdmCmd.AddCommand(generateFacturaCmd)

	// Add flags
	generateFacturaCmd.Flags().StringP("client-name", "c", "", "Client name")
	generateFacturaCmd.Flags().StringP("client-contact", "", "", "Client contact person")
	generateFacturaCmd.Flags().StringP("client-rnc", "", "", "Client RNC")
	generateFacturaCmd.Flags().StringP("client-id", "", "", "Client ID for logo")
	generateFacturaCmd.Flags().StringP("quotation-number", "q", "", "Quotation number")
	generateFacturaCmd.Flags().StringP("project", "p", "", "Project name")
	generateFacturaCmd.Flags().StringP("service", "s", "", "Service description")
	generateFacturaCmd.Flags().StringP("description", "d", "", "Detailed description")
	generateFacturaCmd.Flags().Float64P("unit-price", "", 0, "Unit price for the service")
	generateFacturaCmd.Flags().StringP("output", "o", "", "Output PDF filename")
}

func GetMaroto() core.Maroto {
	data := getSampleCotizacionData()
	return GetMarotoWithData(data)
}

func GetMarotoWithData(data CotizacionData) core.Maroto {
	cfg := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(10).
		WithTopMargin(10).
		WithRightMargin(10).
		WithBottomMargin(10).
		Build()

	mrt := maroto.New(cfg)
	m := maroto.NewMetricsDecorator(mrt)

	// Header with logo and company info
	m.AddRows(getCompanyHeader(data))

	// Client and quotation info section
	m.AddRows(getClientSection(data))

	// Service description
	m.AddRows(getServiceSection(data))

	// Items table header
	m.AddRows(getItemsHeader())

	// Items
	m.AddRows(getItemsRows(data.Items)...)

	// Notes section
	m.AddRows(getNotesSection())

	// Financial summary
	m.AddRows(getFinancialSummary(data)...)

	// Footer
	m.AddRows(getFooter(data))

	return m
}

func getCompanyHeader(data CotizacionData) core.Row {
	logoURL := fmt.Sprintf("%s/images/logos/%s", viper.GetString("url.img"), data.ClientID)

	return row.New(25).Add(
		image.NewFromFileCol(3, logoURL, props.Rect{
			Center:  true,
			Percent: 80,
		}),
		text.NewCol(6, "ORGM\nDESARROLLO DE INGENIERÍA Y PROYECTOS", props.Text{
			Size:  12,
			Style: fontstyle.Bold,
			Align: align.Center,
			Top:   8,
		}),
		col.New(3).Add(
			text.New("RNC: 1-31-91523-1", props.Text{
				Size:  8,
				Align: align.Right,
				Top:   2,
			}),
			text.New("ORGM E.I.R.L.", props.Text{
				Size:  8,
				Align: align.Right,
				Top:   5,
			}),
			text.New(data.CompanyAddress, props.Text{
				Size:  7,
				Align: align.Right,
				Top:   8,
			}),
		),
	)
}

func getClientSection(data CotizacionData) core.Row {
	return row.New(30).Add(
		col.New(6).Add(
			text.New("DATOS GENERALES:", props.Text{
				Size:  9,
				Style: fontstyle.Bold,
				Top:   0,
			}),
			text.New(fmt.Sprintf("CLIENTE: %s", data.ClientName), props.Text{
				Size: 8,
				Top:  3,
			}),
			text.New(fmt.Sprintf("BR: %s", data.ClientContact), props.Text{
				Size: 8,
				Top:  6,
			}),
			text.New(fmt.Sprintf("RNC: %s", data.ClientRNC), props.Text{
				Size: 8,
				Top:  9,
			}),
			text.New(fmt.Sprintf("CONTACTO: %s", data.ClientAddress), props.Text{
				Size: 8,
				Top:  12,
			}),
			text.New(fmt.Sprintf("FECHA: %s", data.Date), props.Text{
				Size: 8,
				Top:  15,
			}),
			text.New(fmt.Sprintf("PROYECTO: %s", data.Project), props.Text{
				Size: 8,
				Top:  18,
			}),
			text.New(fmt.Sprintf("UBICACIÓN: %s", data.Location), props.Text{
				Size: 8,
				Top:  21,
			}),
		),
		col.New(6).Add(
			text.New("COTIZACIÓN", props.Text{
				Size:  14,
				Style: fontstyle.Bold,
				Align: align.Center,
				Top:   0,
			}),
			text.New(fmt.Sprintf("NÚMERO: %s", data.QuotationNumber), props.Text{
				Size:  8,
				Align: align.Right,
				Top:   4,
			}),
			text.New(fmt.Sprintf("CÓDIGO CLIENTE: %s", data.ClientCode), props.Text{
				Size:  8,
				Align: align.Right,
				Top:   7,
			}),
			text.New(fmt.Sprintf("CATEGORÍA: %s", data.Category), props.Text{
				Size:  8,
				Align: align.Right,
				Top:   10,
			}),
			text.New(fmt.Sprintf("SERVICIO: %s", data.Service), props.Text{
				Size:  8,
				Align: align.Center,
				Top:   15,
				Color: &props.WhiteColor,
			}),
		).WithStyle(&props.Cell{
			BackgroundColor: getDarkGrayColor(),
			BorderType:      border.Full,
			BorderColor:     &props.BlackColor,
		}),
	)
}

func getServiceSection(data CotizacionData) core.Row {
	return row.New(15).Add(
		col.New(2).Add(
			text.New("DESCRIPCIÓN GENERAL:", props.Text{
				Size:  9,
				Style: fontstyle.Bold,
				Top:   2,
			}),
		),
		col.New(10).Add(
			text.New(data.Description, props.Text{
				Size: 8,
				Top:  2,
			}),
		),
	).WithStyle(&props.Cell{
		BorderType:  border.Full,
		BorderColor: &props.BlackColor,
	})
}

func getItemsHeader() core.Row {
	darkGray := getDarkGrayColor()
	return row.New(8).Add(
		text.NewCol(1, "ITEM", props.Text{
			Size:  8,
			Style: fontstyle.Bold,
			Align: align.Center,
			Color: &props.WhiteColor,
			Top:   2,
		}),
		text.NewCol(5, "DESCRIPCIÓN", props.Text{
			Size:  8,
			Style: fontstyle.Bold,
			Align: align.Center,
			Color: &props.WhiteColor,
			Top:   2,
		}),
		text.NewCol(1, "CANTIDAD", props.Text{
			Size:  8,
			Style: fontstyle.Bold,
			Align: align.Center,
			Color: &props.WhiteColor,
			Top:   2,
		}),
		text.NewCol(1, "UNIDAD", props.Text{
			Size:  8,
			Style: fontstyle.Bold,
			Align: align.Center,
			Color: &props.WhiteColor,
			Top:   2,
		}),
		text.NewCol(2, "VALOR/UD.", props.Text{
			Size:  8,
			Style: fontstyle.Bold,
			Align: align.Center,
			Color: &props.WhiteColor,
			Top:   2,
		}),
		text.NewCol(2, "VALOR TOTAL", props.Text{
			Size:  8,
			Style: fontstyle.Bold,
			Align: align.Center,
			Color: &props.WhiteColor,
			Top:   2,
		}),
	).WithStyle(&props.Cell{
		BackgroundColor: darkGray,
		BorderType:      border.Full,
		BorderColor:     &props.BlackColor,
	})
}

func getItemsRows(items []CotizacionItem) []core.Row {
	var rows []core.Row

	for _, item := range items {
		r := row.New(12).Add(
			text.NewCol(1, item.Item, props.Text{
				Size:  8,
				Align: align.Center,
				Top:   3,
			}),
			text.NewCol(5, item.Description, props.Text{
				Size:  8,
				Align: align.Left,
				Top:   3,
			}),
			text.NewCol(1, fmt.Sprintf("%d", item.Quantity), props.Text{
				Size:  8,
				Align: align.Center,
				Top:   3,
			}),
			text.NewCol(1, item.Unit, props.Text{
				Size:  8,
				Align: align.Center,
				Top:   3,
			}),
			text.NewCol(2, fmt.Sprintf("RD$ %.2f", item.UnitPrice), props.Text{
				Size:  8,
				Align: align.Center,
				Top:   3,
			}),
			text.NewCol(2, fmt.Sprintf("RD$ %.2f", item.TotalPrice), props.Text{
				Size:  8,
				Align: align.Center,
				Top:   3,
			}),
		).WithStyle(&props.Cell{
			BorderType:  border.Full,
			BorderColor: &props.BlackColor,
		})

		rows = append(rows, r)
	}

	return rows
}

func getNotesSection() core.Row {
	return row.New(25).Add(
		col.New(8).Add(
			text.New("NOTAS:", props.Text{
				Size:  9,
				Style: fontstyle.Bold,
				Top:   0,
			}),
			text.New("1   TIEMPO DE ENTREGA MÁXIMO DE: 3 SEMANA/S", props.Text{
				Size: 8,
				Top:  3,
			}),
			text.New("2   FORMATO DE PAGO: 60% CON LA ORDEN Y 40% CONTRA ENTREGA", props.Text{
				Size: 8,
				Top:  6,
			}),
			text.New("3   CUENTA EN RD$ BANCO POPULAR DOMINICANO / CUENTA CORRIENTE RD$ / 822826538", props.Text{
				Size: 8,
				Top:  9,
			}),
			text.New("4   COTIZACIÓN VÁLIDA POR 30 DÍAS", props.Text{
				Size: 8,
				Top:  12,
			}),
			text.New("5   TRABAJO EXTRA SERÁ CONSIDERADO COMO UN ADICIONAL CON PREVIA AUTORIZACIÓN.", props.Text{
				Size: 8,
				Top:  15,
			}),
		),
		col.New(4),
	)
}

func getFinancialSummary(data CotizacionData) []core.Row {
	return []core.Row{
		row.New(4).Add(
			col.New(8),
			text.NewCol(2, "SUBTOTAL:", props.Text{
				Size:  8,
				Style: fontstyle.Bold,
				Align: align.Right,
				Top:   1,
			}),
			text.NewCol(2, fmt.Sprintf("RD$ %.2f", data.Subtotal), props.Text{
				Size:  8,
				Align: align.Center,
				Top:   1,
			}),
		).WithStyle(&props.Cell{
			BorderType:  border.Full,
			BorderColor: &props.BlackColor,
		}),
		row.New(4).Add(
			col.New(8),
			text.NewCol(2, "ITBIS:", props.Text{
				Size:  8,
				Style: fontstyle.Bold,
				Align: align.Right,
				Top:   1,
			}),
			text.NewCol(2, fmt.Sprintf("RD$ %.2f", data.ITBIS), props.Text{
				Size:  8,
				Align: align.Center,
				Top:   1,
			}),
		).WithStyle(&props.Cell{
			BorderType:  border.Full,
			BorderColor: &props.BlackColor,
		}),
		row.New(4).Add(
			col.New(8),
			text.NewCol(2, "RETENCIÓN:", props.Text{
				Size:  8,
				Style: fontstyle.Bold,
				Align: align.Right,
				Top:   1,
			}),
			text.NewCol(2, fmt.Sprintf("RD$ %.2f", data.Retencion), props.Text{
				Size:  8,
				Align: align.Center,
				Top:   1,
			}),
		).WithStyle(&props.Cell{
			BorderType:  border.Full,
			BorderColor: &props.BlackColor,
		}),
		row.New(4).Add(
			col.New(8),
			text.NewCol(2, "TOTAL A PAGAR:", props.Text{
				Size:  8,
				Style: fontstyle.Bold,
				Align: align.Right,
				Top:   1,
				Color: &props.WhiteColor,
			}),
			text.NewCol(2, fmt.Sprintf("RD$ %.2f", data.Total), props.Text{
				Size:  8,
				Style: fontstyle.Bold,
				Align: align.Center,
				Top:   1,
				Color: &props.WhiteColor,
			}),
		).WithStyle(&props.Cell{
			BackgroundColor: getBlueColor(),
			BorderType:      border.Full,
			BorderColor:     &props.BlackColor,
		}),
	}
}

func getFooter(data CotizacionData) core.Row {
	return row.New(15).Add(
		col.New(12).Add(
			text.New(fmt.Sprintf("Tel: %s, Correos: %s", data.CompanyPhone, data.CompanyEmail), props.Text{
				Size:  8,
				Style: fontstyle.BoldItalic,
				Align: align.Center,
				Top:   5,
				Color: getBlueColor(),
			}),
			text.New("Página 1 de 1", props.Text{
				Size:  8,
				Align: align.Right,
				Top:   10,
			}),
		),
	)
}

func getDarkGrayColor() *props.Color {
	return &props.Color{
		Red:   55,
		Green: 55,
		Blue:  55,
	}
}

func getGrayColor() *props.Color {
	return &props.Color{
		Red:   200,
		Green: 200,
		Blue:  200,
	}
}

func getBlueColor() *props.Color {
	return &props.Color{
		Red:   0,
		Green: 100,
		Blue:  200,
	}
}

func getRedColor() *props.Color {
	return &props.Color{
		Red:   150,
		Green: 10,
		Blue:  10,
	}
}
