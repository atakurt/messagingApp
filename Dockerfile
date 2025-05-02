FROM golang:1.24-alpine AS builder

RUN adduser -D -u 10001 appuser

RUN apk add --no-cache git ca-certificates

# Install swag CLI
RUN go install github.com/swaggo/swag/cmd/swag@latest

ENV PATH=$PATH:/go/bin
ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

# Generate Swagger docs
RUN swag init --generalInfo cmd/api/main.go --output docs

RUN go build -ldflags="-s -w" -o messagingApp ./cmd/api

FROM gcr.io/distroless/static-debian11

WORKDIR /app
COPY --from=builder /app/messagingApp .

COPY --from=builder /app/docs ./docs

USER 10001:10001

ENV TZ=UTC
ENV ENV=production

ENTRYPOINT ["./messagingApp"]