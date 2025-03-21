# TM-to-ABRP

Teslamate MQTT to A Better Route Planner (ABRP) Bridge. A simple user interface allows you to enable updating your Tesla's live data from Teslamate to A Better Route Planner for a specified time period.

## Usage

Run this container in isolation, or as part of your Teslamate's `docker-compose.yml`. Having MQTT set up for Teslamate is a requirement.

```yml
version: "3"

services:
  app:
    image: ghcr.io/tma/tm-to-abrp:latest
    environment:
      - PATH_PREFIX=/abrp # optional, default is /
      - TZ=Europe/Berlin # set timezone
      - MQTT=tcp://localhost:1883 # mqtt URL
      - MQTT_USERNAME=username # optional, MQTT username
      - MQTT_PASSWORD=password # optional, MQTT password
      - MQTT_TLS=false # optional, set to true to enable TLS
      - MQTT_TLS_SKIP_VERIFY=false # optional, set to true to skip TLS certificate verification
      - TM_CAR_NUMBER=1 # teslamate car number
      - ABRP_CAR_MODEL=xyz # check values via https://api.iternio.com/1/tlm/get_carmodels_list
      - ABRP_TOKEN=xyz # car token
      - ABRP_API_KEY=xyz
```

## Development

Use Codespaces on GitHub. To build and run the application using Docker, do this:

```sh
DOCKER_BUILDKIT=1 docker build . -t app --target test && docker run -it --rm --network host app
```

## Credits

Heavily influenced by the Python-based implementation created by @letienne: [letienne/teslamate-abrp](https://github.com/letienne/teslamate-abrp)
