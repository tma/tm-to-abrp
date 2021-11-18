# MQTT-to-ABRP

## Usage

```yml
version: "3"

services:
  app:
    build: /path/to/mqtt-to-abrp-repo
    environment:
      - PATH_PREFIX=/abrp # optional, default is /
      - MQTT=tcp://localhost:1883 # mqtt URL
      - TM_CAR_NUMBER=1 # teslamate car number
      - ABRP_CAR_MODEL=xyz # check values via https://api.iternio.com/1/tlm/get_carmodels_list
      - ABRP_TOKEN=xyz # car token
      - ABRP_API_KEY=xyz
```

## Credits

Heavily influenced by Python-based implementation created by @letienne: https://github.com/letienne/teslamate-abrp
