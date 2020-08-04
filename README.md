# tcmdtool

Import and/or export [Async API](https://www.asyncapi.com/) to/from [TIBCO Cloud™ Metadata (TCMD)](https://www.tibco.com/products/tibco-cloud-metadata). Generate and build [Flogo](https://www.flogo.io/) App from the exported Async API definition, and demonstrate a sample Flogo App that subscribes and processes [MQTT](https://mosquitto.org/) messages.

## Installation

Download and install [Go](https://golang.org/dl/).

Clone and install this tool:

```bash
git clone https://github.com/yxuco/tcmdtool.git
cd tcmdtool
go install
```

Create a TCMD technical user, and then create a config file `.tcmdtool` similar to [sample.tcmdtool](./sample.tcmdtool), and set the TCMD server `url`, `ebxuser` and `password` in the config.

## Import and export AsyncAPI

In an empty working folder, import sample AsyncAPI definition, [streetlights.yml](./test-data/streetlights.yml), into TCMD.

```bash
tcmdtool import --config /path/to/.tcmdtool -i /path/to/tcmdtool/test-data/streetlights.yml
```

In TCMD, verify that a new TCMD asset `streetlights` is created together with all its related assets and data types.

In the working folder, export the `streetlights` defintion from TCMD using `yaml` data format.

```bash
tcmdtool export --config /path/to/.tcmdtool -r streetlights -f yaml
```

Verify that the generated file `streetlights.yaml` in the working folder contains the same definitions as that in the original sample, [streetlights.yml](./test-data/streetlights.yml).

Optionally, cleanup the test data from TCMD if they are no longer used:

```bash
tcmdtool clean --config /path/to/.tcmdtool -i /path/to/tcmdtool/test-data/streetlights.yml
```

## Generate and build Flogo App

Following instructions are based on the open-source Flogo project, [asyncapi](https://github.com/project-flogo/asyncapi).

Install Flogo tools:

```bash
go get -u github.com/project-flogo/cli/...
go get -u github.com/project-flogo/asyncapi/...
```

In the previous working folder that contains the exported AsyncAPI file `streetlights.yaml`, generate and build a Flogo App:

```bash
asyncapi -input streetlights.yaml -type flogodescriptor
flogo create --cv v0.9.3-0.20190610180641-336db421a17a -f flogo.json streetlights
mv support.go streetlights/src/
cd streetlights/src
go mod edit -replace github.com/project-flogo/core@v0.9.4-hf.1=github.com/project-flogo/core@v0.9.3-0.20190610180641-336db421a17a
cd ..
flogo build
```

The above commands generated a Flogo App that is built as an executable, `streetlights/bin/streetlights`. The Flogo App implements the specified AsyncAPIs, and it subscribes and logs [MQTT](https://mosquitto.org/) messages. The generated Flogo model `flogo.json` in the working folder can be edited and recompiled to include more advanced Flogo activities for event processing.

## Testing

Use [Docker](https://www.docker.com/) to start [MQTT](https://mosquitto.org/) broker for testing:

```bash
docker run -it -p 1883:1883 -p 9001:9001 eclipse-mosquitto
```

In another terminal, start the `streetlighs` Flogo App from the working folder:

```bash
bin/streetlights
```

In another terminal, find the docker container ID of the MQTT broker, and send a test message from the docker container:

```bash
docker ps
docker exec -it <MOSQUITTO CONTAINER ID> /bin/sh
mosquitto_pub -m 'on' -t smartylighting/streetlights/1/0/action/1/turn/on
```

Verify that a log message is printed in the `streetlights` Flogo App terminal.
