# Config Diff

## Description

CLI utility for comparing the application properties for between environments

## Build

Build the executable binary file with the standard go build command

```shell
go build 
```

## Run

Run the binary with a path to the config properties repo and two environment arguments.

```shell
./config-diff <path to paven-config-properties repo> <source env> <target env>
```