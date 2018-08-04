package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"

	hue "github.com/collinux/gohue"
	"gopkg.in/yaml.v2"
)

func generateConfig(outputFile *string) {
	bridges, err := hue.FindBridges()
	if err != nil {
		panic(fmt.Sprintf("Error while searching for Hue bridges on the local network: %v\n", err))
	}
	if len(bridges) == 0 {
		panic("No Hue bridges found on the local network.\n")
	} else if len(bridges) == 1 {
		fmt.Printf("Found 1 Hue bridge on the local network: %s (%v).\n", bridges[0].Info.Device.FriendlyName, bridges[0].IPAddress)
	} else {
		panic("Found multiple Hue bridges on the local network.\n")
	}

	bridge := bridges[0]
	var apiKey string

	for {
		apiKey, err = bridge.CreateUser("hue_exporter")
		if err != nil {
			print("Creating API key failed. Have you pushed the link button on your hub?\nPress Enter to continue.")
			bufio.NewReader(os.Stdin).ReadBytes('\n')
		} else {
			break
		}
	}
	println("Successfully created API user.")

	config, err := yaml.Marshal(Config{
		IPAddr: bridge.IPAddress,
		APIKey: apiKey,
	})

	if err != nil {
		panic(fmt.Sprintf("Error while generating configuration file content: %v.\n", err))
	}

	err = ioutil.WriteFile(*outputFile, config, 0660)
	if err != nil {
		panic(fmt.Sprintf("Error while writing configuration file content to %s: %v.\n", *outputFile, err))
	}
	fmt.Printf("Configuration written to %s.", *outputFile)
}
