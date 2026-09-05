# ---------- build stage ----------
FROM golang:1.27-alpine AS build

WORKDIR /src

ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/aichat ./cmd/server

# ---------- runtime stage ----------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=build /out/aichat /app/aichat
COPY prompts /app/prompts

WORKDIR /app

EXPOSE 3001

ENTRYPOINT ["/app/aichat"]
