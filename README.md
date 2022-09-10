# Hue exporter

[![Travis CI build](https://travis-ci.org/mitchellrj/hue_exporter.svg?branch=master)](https://travis-ci.org/mitchellrj/hue_exporter) [![Go report](https://goreportcard.com/badge/github.com/mitchellrj/hue_exporter)](https://goreportcard.com/report/github.com/mitchellrj/hue_exporter)

This is a Prometheus exporter for the Philips Hue system. It's written in Go.

It exposes metrics about lights, groups and sensors. I hope you like it.

There are other Hue exporter projects. They didn't quite do what I wanted and to make them do what I wanted would have meant some pretty major changes. They might work for you though:

* [huepro](https://github.com/stevenbower/huepro)
* [hue_exporter](https://github.com/nilsding/hue_exporter)

Here's a screenshot of an [example Grafana dashboard](docs/grafana.json):
![Grafana](https://raw.githubusercontent.com/mitchellrj/hue_exporter/master/docs/grafana.png)

## Light metrics

Each light metric is labelled with the friendly name, the model, the type, the product name, the manufacturer name, and the unique ID.

* `hue_light_brightness`
* `hue_light_hue`
* `hue_light_saturation`
* `hue_light_on`: `0` means off, `1` means on
* `hue_light_reachable`: `0` or `1` representing false or true

## Group metrics

Each group metric is labelled with the name, and the type.

* `hue_group_brightness`
* `hue_group_hue`
* `hue_group_saturation`
* `hue_group_on`: `0` means off, `1` means some lights within the group are on, `2` means all lights within the group are on

## Sensor metrics

Each sensor metric is labelled with the friendly name, the model, the type, the product name, the manufacturer name, the unique ID and the device ID (for motion sensor components). The device ID is a truncated version of the unique ID, that may be used to group the individual sensors that make up a single physical device.

* `hue_sensor_value`: value varies depending on the `type` of the sensor. For switches, it's the value of the last button pressed; for daylight and presence sensors it's a `0` or `1` representing false or true values; for the temperature sensor it's hundredths of a degree celsius; for the light level sensor it's Lux.
* `hue_sensor_battery`: battery level percentage (0 for sensors that have no battery)
* `hue_sensor_last_updated`: last updated timestamp (Unix epoch)
* `hue_sensor_on`: `0` or `1` representing false or true
* `hue_sensor_reachable`: `0` or `1` representing false or true

Some sensor type values you might find useful:

* `Daylight`: the Hue Hub's built-in "daylight" sensor, based on sunset / sunrise in your configured location
* `ZGPSwitch`: the Hue tap switch
* `ZLLSwitch`: the Hue dimmer switch
* `ClipGenericStatus`: a generic sensor, usually a pseudo-sensor created through the API for automation purposes
* `ZLLTemperature`: the temperature sensor in the Hue motion sensor
* `ZLLPresence`: the presence sensor in the Hue motion sensor
* `ZLLLightLevel`: the light level sensor in the Hue motion sensor

## General metrics

* `hue_group_scrapes_failed`, `hue_light_scrapes_failed`, `hue_sensor_scrapes_failed`: count of failures when trying to scrape from the Hue API.
* `hue_bridge_restarts`: count of times the bridge has restarted (*estimated based on sensor data*).

## Metric structure

> Hey, why didn't you combine the metrics for brightness and hue and saturation and on and reachable?

I tried that, but because of the different label combinations, it didn't really work and you ended up with a load of empty labels.

## Configuration

There's an example configuration file `hue_exporter.example.yml` in this repository, but you can also generate one! Run `hue_exporter generate` to have the app discover to your Hue bridge and create an API user for itself, then write the necessary configuration.

You can also replace some configuration value with OS environment variables.

```yaml
ip_address: ${IP_ADDRESS}
api_key: ${API_KEY}
```

## Running

```
hue_exporter --config.file hue_exporter.yml --listen.address :9366
```

Those flag values are the defaults, so you could just run `hue_exporter` on its own if you're happy with those.

### Docker

There are a few docker images built, including ones for ARM7 (Raspberry Pi). You can find these on [Docker Hub](https://hub.docker.com/r/mitchellrj/hue_exporter). They expose `/etc/hue_exporter` as a volume for you to generate or pass in your own configuration.

```
docker run -p 9366:9366 -v my_config.yml:/etc/hue_exporter/config.yml mitchellrj/hue_exporter:latest
```

## License

MIT / X11 Consortium license. I'd prefer to use Apache 2.0, but the excellent Hue library that this app uses is GPL 2.0 and that isn't compatible with Apache.
