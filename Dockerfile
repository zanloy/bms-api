FROM golang:1.16.0-alpine3.13 AS builder

# Setup VA certs for downloading on VPN
RUN su root -c "apk --no-cache add ca-certificates"
COPY certs/* /usr/local/share/ca-certificates/
RUN su root -c "update-ca-certificates"

# Setup environment
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Move to the appropriate working directory
WORKDIR /build

# Copy and download dependency go modules
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build our application
RUN go build -o bms-api .

# Move to /dist for isolating the binary
WORKDIR /dist

# Copy our binary to new directory
RUN cp /build/bms-api .

# Build our final (smaller) image
FROM zanloy/alpine-va:3.13.3

# Environment for our running app
ENV GIN_MODE=release \
    PORT=8080

# Expose the port we will be using
EXPOSE 8080

# Setup VA certs for downloading on VPN
RUN su root -c "apk --no-cache add ca-certificates"
COPY certs/* /usr/local/share/ca-certificates/
RUN su root -c "update-ca-certificates"

# Copy in configuration file
COPY bms-api.yaml /

# Copy in bms-api binary from builder image
COPY --from=builder /dist/bms-api /

# Command to use as entrypoint
ENTRYPOINT ["/bms-api"]
