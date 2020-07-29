FROM golang:alpine

# Set necessary environmet variables needed for our image
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Copy and download dependency using go mod
# COPY go.mod .
# COPY go.sum .
# RUN go mod download

# Copy the code into the container
COPY . .

# Build main application
RUN go build -o main main.go

# Build Terminal application
RUN go build -o term term.go

# Export necessary port
EXPOSE 9999 10101


# Command to run when starting the container
CMD ["./main"]
