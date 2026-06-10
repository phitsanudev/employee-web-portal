FROM golang:1.23-bookworm AS builder

WORKDIR /src/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/employee-api ./cmd/api

FROM debian:bookworm-slim

WORKDIR /app/backend

COPY --from=builder /bin/employee-api /usr/local/bin/employee-api
COPY backend/config ./config
COPY backend/docs ./docs
COPY backend/uploads/.gitkeep ./uploads/.gitkeep

ENV APP_ENV=production
ENV PORT=8080

EXPOSE 8080

CMD ["employee-api"]
