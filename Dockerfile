FROM golang:1.10.3 AS buildcontainer

ENV REPO_PATH /go/src/github.com/mdellandrea/minutes-server

WORKDIR $REPO_PATH
RUN curl https://glide.sh/get | sh

COPY . .
RUN glide install
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o minutes-server

##########################################################

FROM alpine:3.8

ENV REPO_PATH /go/src/github.com/mdellandrea/minutes-server

WORKDIR /usr/local/bin
COPY --from=buildcontainer $REPO_PATH/minutes-server .
CMD minutes-server

EXPOSE 8080
