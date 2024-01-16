package configs

import (
	"flag"
	"os"
)

type Config struct {
	RunAddr        string
	DSN            string
	AccrualBaseURL string
}

func Parse() Config {
	config := Config{}
	config.RunAddr = os.Getenv("RUN_ADDRESS")
	config.DSN = os.Getenv("DATABASE_URI")
	config.AccrualBaseURL = os.Getenv("ACCRUAL_SYSTEM_ADDRESS")

	var flagRunAddr, flagDSN, flagAccrualBaseURL string
	flag.StringVar(&flagRunAddr, "a", "", "server's address")
	flag.StringVar(&flagDSN, "d", "", "database URI")
	flag.StringVar(&flagAccrualBaseURL, "r", "", "accrual service address")
	flag.Parse()

	if flagRunAddr != "" {
		config.RunAddr = flagRunAddr
	}
	if flagDSN != "" {
		config.DSN = flagDSN
	}
	if flagAccrualBaseURL != "" {
		config.AccrualBaseURL = flagAccrualBaseURL
	}

	return config
}
