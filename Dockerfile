# Step 1: Build the Go binary
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
# Adjust the path to your main.go if it's inside a cmd/ folder
RUN CGO_ENABLED=0 GOOS=linux go build -o lms-backend main.go

# Step 2: Run the binary in a tiny container
FROM alpine:latest  

WORKDIR /root/

# Copy the pre-built binary from the builder stage
COPY --from=builder /app/lms-backend .

# Hugging Face Spaces route traffic to port 7860 by default
EXPOSE 7860

CMD ["./lms-backend"]