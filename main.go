package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

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
		MatchNames  bool     `yaml:"match_names"`
	} `yaml:"sensors"`
}

// Bridge is an interface for the bridge struct from Collinux/gohue to allow stubbing in tests
type Bridge interface {
	Login(string) error
	GetAllSensors() ([]hue.Sensor, error)
	GetAllLights() ([]hue.Light, error)
	GetAllGroups() ([]hue.Group, error)
}

func readConfig(raw []byte, cfg *Config) {
	err := yaml.Unmarshal(raw, cfg)
	if err != nil {
		log.Fatalf("Error parsing config file: %v\n", err)
	}
}

func newBridge(ipAddr string) Bridge {
	bridge, err := hue.NewBridge(ipAddr)
	if err != nil {
		log.Fatalf("Error connecting to Hue bridge at %v: %v\n", ipAddr, err)
	}
	return bridge
}

func setupPrometheus(bridge Bridge, cfg *Config) {
	err := bridge.Login((*cfg).APIKey)
	if err != nil {
		log.Fatalf("Error authenticating with Hue bridge at %v: %v\n", (*cfg).IPAddr, err)
	}

	prometheus.MustRegister(NewGroupCollector(namespace, bridge))
	prometheus.MustRegister(NewLightCollector(namespace, bridge))
	prometheus.MustRegister(NewSensorCollector(namespace, bridge, (*cfg).SensorConfig.IgnoreTypes, (*cfg).SensorConfig.MatchNames))
}

func listen() {
	http.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{
		Addr:         (*addr).String(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func runServer() {
	var cfg Config

	raw, err := ioutil.ReadFile(*config)
	if err != nil {
		log.Fatalf("Error reading config file: %v\n", err)
	}
	readConfig(raw, &cfg)
	bridge := newBridge(cfg.IPAddr)
	setupPrometheus(bridge, &cfg)
	listen()
}

func main() {
	command := kingpin.MustParse(app.Parse(os.Args[1:]))
	if *version {
		fmt.Printf("%s\n", VERSION)
	} else {
		switch command {
		case run.FullCommand():
			runServer()
		case generate.FullCommand():
			generateConfig(output)
		}
	}
}
