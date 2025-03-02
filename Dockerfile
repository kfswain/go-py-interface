# Dockerfile has specific requirement to put this ARG at the beginning:
# https://docs.docker.com/engine/reference/builder/#understand-how-arg-and-from-interact
ARG BUILDER_IMAGE=alpine:edge

## Multistage build
FROM ${BUILDER_IMAGE} as builder
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

# Dependencies
RUN apk add --no-cache --update go gcc g++ python3
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN go build -o /main

FROM alpine:edge
WORKDIR /
COPY --from=builder /main /main


ENTRYPOINT ["/main"]