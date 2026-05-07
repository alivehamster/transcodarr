FROM golang:1.26.2-alpine AS builder
WORKDIR /app
RUN apk add --no-cache gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
COPY libs/ ./libs/
RUN CGO_ENABLED=1 go build -o server .

FROM node:24-alpine AS node-builder
WORKDIR /app
COPY ./frontend .
RUN npm install
RUN npm run build

FROM jlesage/handbrake:v26.03.3

WORKDIR /app

RUN apk add --no-cache ffmpeg

COPY --from=builder /app/server .
COPY --from=node-builder /app/dist ./frontend/dist

CMD ["/app/server"]
