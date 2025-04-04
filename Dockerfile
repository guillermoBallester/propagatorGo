FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN ls -la cmd/

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o propagatorGo ./cmd/propagator

RUN ls -la /app

FROM alpine:3.21

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/propagatorGo /app/propagatorGo

RUN ls -la /app

# Run the application
CMD ["/app/propagatorGo"]