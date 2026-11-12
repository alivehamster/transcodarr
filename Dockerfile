# Transcodarr backend builder
FROM golang:1.26.3-trixie AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o ./server ./cmd/

# frontend builder
FROM node:24-trixie AS node-builder
WORKDIR /app
COPY ./frontend .
RUN npm install
RUN npm run build

# HandBrake builder
FROM debian:trixie AS handbrake-builder

ENV HANDBRAKE_VERSION_TAG=1.11.2

ENV HANDBRAKE_URL_GIT=https://github.com/HandBrake/HandBrake.git

ENV DEBIAN_FRONTEND=noninteractive

WORKDIR /HB

## Prepare
RUN apt-get update

## Build dependencies
RUN apt-get install -y autoconf automake build-essential cmake git libass-dev libbz2-dev libfontconfig-dev libfreetype-dev libfribidi-dev libharfbuzz-dev libjansson-dev liblzma-dev libmp3lame-dev libnuma-dev libogg-dev libopus-dev libsamplerate0-dev libspeex-dev libssl-dev libtheora-dev libtool libtool-bin libturbojpeg0-dev libvorbis-dev libvpx-dev libx11-dev libx264-dev libxml2-dev m4 make meson nasm ninja-build patch pkg-config zlib1g-dev

## Intel QSV dependencies
RUN apt-get install -y libva-dev libdrm-dev

## Nvidia NVENC/NVDEC dependencies
RUN apt-get install -y clang llvm

## Download HandBrake sources
RUN echo "Downloading HandBrake sources..."
RUN git clone $HANDBRAKE_URL_GIT --branch $HANDBRAKE_VERSION_TAG

## Compile HandBrake
WORKDIR /HB/HandBrake

RUN ./configure --prefix=/usr/local \
                --debug=none \
                --enable-fdk-aac \
                --enable-x265 \
                --enable-numa \
                --enable-qsv \
                --enable-nvenc \
                --enable-nvdec \
                --disable-gtk \
                --launch-jobs=$(nproc) \
                --launch

RUN make --directory=build install

# Transcodarr
FROM debian:trixie

WORKDIR /app

ENV NVIDIA_VISIBLE_DEVICES=all
ENV NVIDIA_DRIVER_CAPABILITIES=all
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update

# Required by Transcodarr
RUN apt-get install -y mediainfo

# Required by HandBrake
RUN apt-get install -y --no-install-recommends \
    libnuma1 \
    libass9 \
    libvpx9 \
    libspeex1 \
    libmp3lame0 \
    libopus0 \
    libva-drm2 \
    libdrm2 \
    libva2 \
    libtheora0 \
    libvorbis0a \
    libvorbisenc2 \
    libx264-164 \
    libjansson4 \
    libturbojpeg0 \
    libmfx-gen1.2 \
    intel-media-va-driver \
    && rm -rf /var/lib/apt/lists/*
    
COPY --from=handbrake-builder /usr/local /usr/local
COPY --from=node-builder /app/dist ./frontend/dist
COPY --from=builder /app/server .

CMD ["/app/server"]
