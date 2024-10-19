# Use the official Golang image as the base image
FROM golang:1.22

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies

COPY ./go.mod ./
COPY ./go.sum ./
RUN go mod download

# Copy the server code to the container
COPY ./internal ./internal
COPY ./tech_ui ./tech_ui
COPY ./config ./config


ENV CONFIG_PATH="/app/config/docker.yaml"
# Build the Go application
RUN cd /app/internal/cmd && go build -o main


# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["/app/internal/cmd/main"]