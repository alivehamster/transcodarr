FROM alpine:latest

# Install dependencies
RUN apk add --no-cache curl autoconf automake busybox cmake g++ jansson-dev lame-dev libass-dev libjpeg-turbo-dev libtheora-dev libtool libvorbis-dev libvpx-dev libxml2-dev m4 make meson nasm ninja numactl-dev opus-dev patch pkgconf python3 speex-dev tar x264-dev

# Intel QSV dependencies
RUN apk add --no-cache libva-dev libdrm-dev

WORKDIR /app

RUN mkdir -p handbrake && \
    curl -L https://github.com/HandBrake/HandBrake/releases/download/1.11.1/HandBrake-1.11.1-source.tar.bz2 -o handbrake-source.tar.bz2 && \
    tar -xjf handbrake-source.tar.bz2 --strip-components=1 -C handbrake && \
    rm handbrake-source.tar.bz2

WORKDIR /app/handbrake

RUN ./configure --disable-gtk --enable-qsv --launch-jobs=$(nproc) --launch

RUN mv build/HandBrakeCLI /usr/local/bin/handbrake-cli

WORKDIR /app

RUN rm -rf handbrake

