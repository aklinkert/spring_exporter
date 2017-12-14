// Copyright © 2017 Alexander Pinnecke <alexander.pinnecke@googlemail.com>
//

package cmd

import (
	"net/http"

	"os"

	"github.com/KalypsoCloud/spring_exporter/spring"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"github.com/spf13/cobra"
)

var (
	insecure          bool
	basicAuthUser     string
	basicAuthPassword string
	scrapeListen      string
	scrapeEndpoint    string
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export <spring-endpoint>",
	Short: "Export metrics",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Usage()
			os.Exit(1)
		}

		endpoint := args[0]

		logger := log.Base()
		logger.SetLevel("warn")
		exp := spring.NewExporter(logger, spring.Namespace, insecure, endpoint, basicAuthUser, basicAuthPassword)

		prometheus.MustRegister(exp)
		prometheus.MustRegister(version.NewCollector("spring_exporter"))

		log.Infof("Exporting spring endpoint: %v", endpoint)
		log.Info("Starting spring_exporter", version.Info())
		log.Info("Build context", version.BuildContext())
		log.Infof("Starting Server: %s", scrapeListen)

		http.Handle(scrapeEndpoint, promhttp.Handler())
		log.Fatal(http.ListenAndServe(scrapeListen, nil))
	},
}

func init() {
	RootCmd.AddCommand(exportCmd)

	exportCmd.Flags().BoolVarP(&insecure, "insecure", "i", false, "Insecure https mode")
	exportCmd.Flags().StringVar(&basicAuthUser, "basic-auth-user", "", "HTTP Basic auth user")
	exportCmd.Flags().StringVar(&basicAuthPassword, "basic-auth-password", "", "HTTP Basic auth password")
	exportCmd.Flags().StringVarP(&scrapeListen, "listen", "l", ":9321", "Host/Port to listen on")
	exportCmd.Flags().StringVarP(&scrapeEndpoint, "endpoint", "e", "/metrics", "Path to listen on")
}
