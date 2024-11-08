# Paven Config Tools

## Description

CLI utility for updating datastore confiurations

## Build

Build the executable binary file with the standard go build command

```shell
go build 
```

## Run

Run the binary with two arguments `-config=config.yml` which provides the correct configuration. additionally the `-outputDir=./output` provides the output location for downloaded files


```shell
./paven-datastore-management -config=config.yaml -outputDir=./output
```

## Potential Enhancements

* Configure enviroments and configurable tables
* Add ability to detect delitions and delete the data from that element
