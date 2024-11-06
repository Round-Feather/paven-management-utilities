# Paven Config Tools

## Description

CLI utility for comparing the database config between two environments

## Build

Build the executable binary file with the standard go build command

```shell
go build 
```

## Run

Run the binary with two environment arguments that match the env names in the `source.yml` config.
Currently, the `source.yml` is hardcoded to be in the same directory as the binary. 

```shell
./paven-purge-events <source env> <target env>
```
## Potential Enhancements

* Parameterize the sources yaml location
* Simplify the sources yaml