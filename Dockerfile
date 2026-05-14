FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG SERVICE
RUN CGO_ENABLED=0 go build -o /app/service ./cmd/${SERVICE}

FROM alpine:3.19
RUN apk add --no-cache tzdata
COPY --from=builder /app/service /app/service
EXPOSE 8080
CMD ["/app/service"]
