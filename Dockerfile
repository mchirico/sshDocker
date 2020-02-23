FROM golang:1.13.7-alpine3.11 AS build

RUN apk add --no-cache git

WORKDIR $GOPATH/src/github.com/mchirico/docker-action

# Copy the entire project and build it

COPY . $GOPATH/src/github.com/mchirico/docker-action

RUN go get -v -t -d .
RUN go build -o /bin/project

# Special files
RUN mkdir -p /credentials
COPY USER /credentials/USER
COPY SERVER /credentials/SERVER
COPY id_rsa /credentials/id_rsa


# This results in a single layer image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=build /bin/project /bin/project
RUN mkdir -p /credentials
COPY --from=build /credentials/USER /credentials/USER
COPY --from=build /credentials/SERVER /credentials/SERVER
COPY --from=build /credentials/id_rsa /credentials/id_rsa

ENTRYPOINT ["/bin/project"]
# Args to project
CMD []

