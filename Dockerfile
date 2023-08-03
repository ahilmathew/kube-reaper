FROM golang:1.18.1-buster as builder

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY *.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /opt/janitor

FROM alpine:latest as final
WORKDIR /app
COPY --from=builder /opt/janitor .
ENTRYPOINT ["/app/janitor"]
