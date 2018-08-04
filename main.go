package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"

	hue "github.com/collinux/gohue"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const namespace = "hue"

var (
	app     = kingpin.New("hue_exporter", "A Prometheus exporter for Philips Hue.")
	version = app.Flag("version", "Print the version and exit.").Short('V').Bool()
	run     = app.Command("run", "Run the exporter.").Default()
	// TODO: update https://github.com/prometheus/prometheus/wiki/Default-port-allocations
	addr     = run.Flag("listen.address", "The address to listen on for HTTP requests.").Short('l').Default(":9366").TCP()
	config   = run.Flag("config.file", "The config file to use.").Short('c').Default("hue_exporter.yml").ExistingFile()
	generate = app.Command("generate", "Generate configuration for Hue exporter.")
	output   = generate.Flag("output.file", "The output file to use.").Short('o').Default("hue_exporter.yml").String()
)

type Config struct {
	IPAddr       string `yaml:"ip_address"`
	APIKey       string `yaml:"api_key"`
	SensorConfig struct {
		IgnoreTypes []string `yaml:"ignore_types"`
	} `yaml:"sensors"`
}

func runServer() {

	raw, err := ioutil.ReadFile(*config)
	if err != nil {
		log.Fatalf("Error reading config file: %v\n", err)
	}

	var cfg Config
	err = yaml.Unmarshal(raw, &cfg)
	if err != nil {
		log.Fatalf("Error parsing config file: %v\n", err)
	}

	bridge, err := hue.NewBridge(cfg.IPAddr)
	if err != nil {
		log.Fatalf("Error connecting to Hue bridge at %v: %v\n", cfg.IPAddr, err)
	}

	err = bridge.Login(cfg.APIKey)
	if err != nil {
		log.Fatalf("Error authenticating with Hue bridge at %v: %v\n", cfg.IPAddr, err)
	}

	prometheus.MustRegister(NewGroupCollector(namespace, bridge))
	prometheus.MustRegister(NewLightCollector(namespace, bridge))
	prometheus.MustRegister(NewSensorCollector(namespace, bridge, cfg.SensorConfig.IgnoreTypes))

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe((*addr).String(), nil))
}

func main() {
	command := kingpin.MustParse(app.Parse(os.Args[1:]))
	if *version {
		println(VERSION)
	} else {
		switch command {
		case run.FullCommand():
			runServer()
		case generate.FullCommand():
			generateConfig(output)
		}
	}
}
