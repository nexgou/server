FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /out/nexgou-benchmark ./benchmark/nexgou/cmd/server

FROM alpine:3.20
WORKDIR /app
RUN mkdir -p /app/data
COPY --from=build /out/nexgou-benchmark /app/nexgou-benchmark
ENV PORT=3001 DB_PATH=/app/data/db.sqlite SERVICE_NAME=nexgou SERVICE_VERSION=2.0.0
EXPOSE 3001
CMD ["/app/nexgou-benchmark"]