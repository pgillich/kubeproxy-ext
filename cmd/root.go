package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/pgillich/kubeproxy-ext/configs"
	"github.com/pgillich/kubeproxy-ext/internal/buildinfo"
	"github.com/pgillich/kubeproxy-ext/internal/logger"
	"github.com/pgillich/kubeproxy-ext/internal/proxy"
)

var cfg = configs.Config{} // nolint:gochecknoglobals // the early default is zero value

var rootCmd = &cobra.Command{ // nolint:gochecknoglobals // zero values are good defaults
	Use:   "kubeproxy-ext",
	Short: "Extension to kubeproxy",
	Long:  `Adds extra Pod info`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Build info",
			"AppName", buildinfo.AppName,
			"Version", buildinfo.Version,
			"BuildTime", buildinfo.BuildTime,
			// "GoMod", buildinfo.GoMod,
		)
		if server, err := proxy.New(cfg.Proxy, log.Logger); err != nil {
			log.Error(err, "new proxy")
			os.Exit(1)
		} else {
			server.Serve()
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(err, "Exit")
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

var log *logger.Logger // nolint:gochecknoglobals // OK

func initConfig() {
	log = logger.New()

	configs.SetDefaults()
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Error(err, "Config Unmarshal")
	}
	log.SetLevel(cfg.LogLevel)

	log.Info("Config", logger.MapToKV(viper.AllSettings())...)
}
