FROM golang:1.18 as builder

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY *.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o janitor

FROM golang:1.18 as final

WORKDIR /app
COPY --from=builder /app/janitor /app/janitor
# Run
CMD ["/app/janitor"]