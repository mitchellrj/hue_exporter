FROM        armhf/alpine:latest
MAINTAINER  Richard Mitchell <hue-exporter@mitch.org.uk>

COPY ./.build/linux-armv7/hue_exporter /bin/hue_exporter
COPY hue_exporter.example.yml   /etc/hue_exporter/config.yml

EXPOSE      9366
ENTRYPOINT  [ "/bin/hue_exporter" ]
CMD         [ "--config.file=/etc/hue_exporter/config.yml" ]
