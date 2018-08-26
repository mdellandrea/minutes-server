# Minutes Server

REST API for minutes calculations, written in Go.

Valid time strings are zero-padded strings in the form "HH:MM ${Meridiem}".

## Requirements

* Docker
* HTTP client such as `curl` or `wget`

## Getting Started

Start Server:

```
$ docker-compose up
```

Stop Server:

`^c` or

```
$ docker-compose down
```

Basic HTTP request to running server:

```
$ curl -X POST http://localhost:8080/time
```