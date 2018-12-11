package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/jmhodges/clock"
	"github.com/zimosworld/pebble/ca"
	"github.com/zimosworld/pebble/cmd"
	"github.com/zimosworld/pebble/db"
	"github.com/zimosworld/pebble/va"
	"github.com/zimosworld/pebble/wfe"
)

const (
    tlsDisabled = "PEBBLE_TLS_DISABLED"
)

type config struct {
	Pebble struct {
		ListenAddress string
		HTTPPort      int
		TLSPort       int
		Certificate   string
		PrivateKey    string
	}
}

func main() {
	configFile := flag.String(
		"config",
		"test/config/pebble-config.json",
		"File path to the Pebble configuration file")
	strictMode := flag.Bool(
		"strict",
		false,
		"Enable strict mode to test upcoming API breaking changes")
	flag.Parse()
	if *configFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Log to stdout
	logger := log.New(os.Stdout, "Pebble ", log.LstdFlags)

	var c config
	err := cmd.ReadConfigFile(*configFile, &c)
	cmd.FailOnError(err, "Reading JSON config file into config structure")

	clk := clock.Default()
	db := db.NewMemoryStore()
	ca := ca.New(logger, db)
	va := va.New(logger, clk, c.Pebble.HTTPPort, c.Pebble.TLSPort)

	wfe := wfe.New(logger, clk, db, va, ca, *strictMode)
	muxHandler := wfe.Handler()

	tlsDisabled := os.Getenv(tlsDisabled)

    logger.Printf("Pebble running, listening on: %s\n", c.Pebble.ListenAddress)

    switch tlsDisabled {
	case "1", "true", "True", "TRUE":
	    err = http.ListenAndServe(
                c.Pebble.ListenAddress,
                muxHandler)
	default:
	    err = http.ListenAndServeTLS(
        		c.Pebble.ListenAddress,
        		c.Pebble.Certificate,
        		c.Pebble.PrivateKey,
        		muxHandler)
	}

	cmd.FailOnError(err, "Calling ListenAndServeTLS()")
}
