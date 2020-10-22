## This is a multi-stage Dockerfile for efficiency
## and size reduction of the final Docker image.
## We specify the base image we need for our
## go application
FROM golang:latest AS build
## Copy source
WORKDIR /app
COPY . .
## Download required modules 
RUN go mod download
## Build a statically-linked Go binary for Linux
RUN CGO_ENABLED=0 GOOS=linux go build -a -o server .
## New build phase -- create binary-only image
FROM alpine:latest
## Add support for HTTPS
RUN apk update && \
    apk upgrade && \
    apk add ca-certificates
## Work Directory
WORKDIR /
## Copy files from previous build container
COPY --from=build /app/server ./
## Check results
RUN env && pwd && find .
## Start the application
CMD ["./server"]
