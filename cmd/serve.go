package cmd

import (
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/turfaa/vmedis-proxy-api/proxy"
)

// serveCmd represents the serve command
var serveCmd = commandWithInit{
	command: &cobra.Command{
		Use:   "serve",
		Short: "Run the vmedis proxy api server",

		Run: func(cmd *cobra.Command, args []string) {
			stockOpnameStartDateStr := viper.GetString("stock_opname_start_date")
			stockOpnameStartDate, err := time.Parse("2006-01-02", stockOpnameStartDateStr)
			if err != nil {
				log.Fatalf("Error parsing stock opname start date '%s': %s", stockOpnameStartDateStr, err)
			}

			proxy.Run(
				proxy.Config{
					DB:                   getDatabase(),
					RedisClient:          getRedisClient(),
					AuthService:          getAuthService(),
					AuthHandler:          getAuthHandler(),
					DrugHandler:          getDrugHandler(stockOpnameStartDate),
					SaleHandler:          getSaleHandler(),
					ProcurementHandler:   getProcurementHandler(),
					StockOpnameHandler:   getStockOpnameHandler(),
					ShiftHandler:         getShiftHandler(),
					TokenHandler:         getTokenHandler(),
					RejectedDrugHandler:  getRejectedDrugHandler(),
					SupplierHandler:      getSupplierHandler(),
					PurchaseOrderHandler: getPurchaseOrderHandler(),
				},
			)
		},
	},
	init: func(cmd *cobra.Command) {
		cmd.Flags().String("stock-opname-start-date", time.Now().AddDate(0, 0, -14).Format(time.DateOnly), "Stock opname start date")

		viper.BindPFlag("stock_opname_start_date", cmd.Flags().Lookup("stock-opname-start-date"))

		// The pharmacy details printed on purchase orders (surat pesanan).
		// Purchase orders are disabled while all of them are empty.
		for _, f := range []struct{ key, usage string }{
			{"pharmacy_name", "pharmacy name on purchase orders, e.g. Apotek Aulia Farma"},
			{"pharmacy_license_number", "pharmacy license number (nomor izin) on purchase orders"},
			{"pharmacy_address", "pharmacy address on purchase orders"},
			{"pharmacy_email", "pharmacy email on purchase orders"},
			{"pharmacy_phone", "pharmacy phone on purchase orders"},
			{"pharmacy_city", "city purchase orders are written from, printed before their date"},
			{"pharmacist_name", "name of the pharmacist in charge (APJ), without the apt. title"},
			{"pharmacist_license_number", "SIPA of the pharmacist in charge (APJ)"},
		} {
			flag := strings.ReplaceAll(f.key, "_", "-")
			cmd.Flags().String(flag, "", f.usage)
			viper.BindPFlag(f.key, cmd.Flags().Lookup(flag))
		}
	},
}

func init() {
	if serveCmd.init != nil {
		serveCmd.init(serveCmd.command)
	}

	initAppCommand(serveCmd.command)
}
