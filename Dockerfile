# Use an official Golang image as the base image
FROM golang:1.22 as builder

# Install protobuf-compiler
# Install protobuf-compiler and protoc-gen-go
RUN apt-get update && apt-get install -y protobuf-compiler && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Set the working directory inside the container
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . .

# Run the make command to build the project
RUN make tao

# Use a minimal image to run the application
FROM debian:bullseye-slim

# Set the working directory inside the container
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/bin/tao/server /bin/tao/server

# Expose port 7051
EXPOSE 7051

# Command to run the executable
CMD ["/bin/tao/server"]
