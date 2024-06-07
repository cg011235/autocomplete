# Use the official Golang image
FROM golang:1.22.4-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o main ./cmd/server

# Expose port 8080
EXPOSE 8080

# Command to run the executable
CMD ["./main"]

