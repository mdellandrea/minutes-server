# Minutes Server

REST API for minutes calculations, written in Go.

## Requirements

* [Docker](https://www.docker.com/products/docker-desktop)
* HTTP client such as `curl` or `wget`

## Getting Started

Start Server:

```
$ docker-compose up
```

Stop Server:

```
$ docker-compose down
```

# Details

Valid time strings are zero-padded strings in the form "HH:MM ${Meridiem}".

For Example:
```
01:15 PM
07:05 AM
12:01 PM
```

Setup a new timeId:
```
$ curl -X POST http://localhost:8080/time
{"timeId":"fe2eaa26-babd-48f0-b4e0-e32c61ed7543","currentTime":"12:00 PM"}
```

Get current time for timeId:
```
$ curl http://localhost:8080/time/fe2eaa26-babd-48f0-b4e0-e32c61ed7543
{"currentTime":"12:00 PM"}
```

Add minutes integer to current time for timeId:
```
$ curl -X PUT http://localhost:8080/time/fe2eaa26-babd-48f0-b4e0-e32c61ed7543 -d '{"addMinutes":61}'
{"currentTime":"01:01 PM"}
```

Delete a timeId:
```
$ curl -X DELETE http://localhost:8080/time/fe2eaa26-babd-48f0-b4e0-e32c61ed7543
```

---

This REST API is based on twelve-factor app design and includes many elements of modern productionized microservices such as:

* Vendored dependencies
* Configuration via environment variables
* Multistage Docker build
* Heartbeat endpoint to establish reachability
* Leveled logging
* Stateless process
* Abstracted data store access using a [Data Access Object](https://www.oracle.com/technetwork/java/dataaccessobject-138824.html)
* Graceful server shutdown
* Idiomatic Go testing patterns such as table tests
