FROM golang:1.10.3 AS buildcontainer

WORKDIR "/go/src/github.com/mdellandrea/minutes-server"
COPY . .

RUN curl https://glide.sh/get | sh
RUN glide install
RUN go build

##########################################################

FROM alpine:3.8

COPY --from=buildcontainer /go/src/github.com/mdellandrea/minutes-server /go/bin
CMD ["minutes-server"]

EXPOSE 8080
