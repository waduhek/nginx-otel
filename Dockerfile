FROM golang:1.24.2
WORKDIR /app
COPY main.go tracing.go go.mod go.sum .
RUN go mod tidy
CMD ["go", "run", "."]
