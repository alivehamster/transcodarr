FROM golang:1.26.2-trixie AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
COPY libs/ ./libs/
RUN CGO_ENABLED=1 go build -o server .

FROM node:24-trixie AS node-builder
WORKDIR /app
COPY ./frontend .
RUN npm install
RUN npm run build

FROM zocker160/handbrake-nvenc:110x

WORKDIR /app

RUN apt-get update && apt-get install -y ffmpeg

COPY --from=builder /app/server .
COPY --from=node-builder /app/dist ./frontend/dist

CMD ["/app/server"]
