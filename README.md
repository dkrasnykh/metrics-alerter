# metrics-alerter

Сервис сбора метрик и алертинга

## Description

Kлиент-серверное приложение, где агент собирает и отправляет значения системных метрик серверу с заданной периодичностью.
Сервер обрабатывает, сохраняет, отдаёт по запросу значения метрик.

## Build Agent

```shell
cd $GOPATH/src/github.com/dkrasnykh/metrics-alerter/cmd/agent
go1.21.3 build -o agent *.go
```

## Build Server

```shell
cd $GOPATH/src/github.com/dkrasnykh/metrics-alerter/cmd/server
go1.21.3 build -o server *.go