# First stage: Build the application
FROM golang:1.20-alpine3.18 AS builder

# Install necessary build tools and ca-certificates
RUN apk --no-cache add ca-certificates gcc g++ make

WORKDIR /app

# Copy and download dependencies using go mod
COPY go.* ./
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -a -installsuffix cgo -o gateh8 ./cmd/main.go

# Second stage: Create the final image from scratch
FROM scratch AS final

# Copy SSL root certificates from the builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the compiled application from the builder stage
COPY --from=builder /app/gateh8 /gateh8

# Set the application as the container's entry point
ENTRYPOINT ["/gateh8"]
