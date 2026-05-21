FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Allow go to download and update modules (no go.sum required on first build)
ENV GOFLAGS="-mod=mod"
ENV GONOSUMDB="*"

COPY go.mod ./
RUN go mod download 2>/dev/null || true

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /server .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /server .

EXPOSE 8080
CMD ["/app/server"]
