package cmd

import (
	"log"
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
					DB:                 getDatabase(),
					RedisClient:        getRedisClient(),
					AuthService:        getAuthService(),
					AuthHandler:        getAuthHandler(),
					DrugHandler:        getDrugHandler(stockOpnameStartDate),
					SaleHandler:        getSaleHandler(),
					ProcurementHandler: getProcurementHandler(),
					StockOpnameHandler: getStockOpnameHandler(),
					ShiftHandler:       getShiftHandler(),
				},
			)
		},
	},
	init: func(cmd *cobra.Command) {
		cmd.Flags().String("stock-opname-start-date", time.Now().AddDate(0, 0, -14).Format(time.DateOnly), "Stock opname start date")

		viper.BindPFlag("stock_opname_start_date", cmd.Flags().Lookup("stock-opname-start-date"))
	},
}

func init() {
	if serveCmd.init != nil {
		serveCmd.init(serveCmd.command)
	}

	initAppCommand(serveCmd.command)
}
