FROM golang:latest

# Set working directory
WORKDIR /app

# Clone your repository from GitHub
RUN git clone https://github.com/20RitikSingh/Ecommerce-Flower-Shop .

# Build your Go application
RUN go build -o server server.go

# Command to run the server
CMD ["./server"]
