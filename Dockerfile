FROM golang:1.24-alpine AS build
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
ARG SERVICE
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/service ./cmd/${SERVICE}

FROM alpine:3.20
WORKDIR /app

COPY --from=build /out/service /app/service
COPY web /app/web
COPY api /app/api

EXPOSE 8080
CMD ["/app/service"]
