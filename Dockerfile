# 1st stage build go binary
FROM golang:1.26.4

WORKDIR /usr/app-be

RUN apt-get update && \
    apt-get install -y build-essential \
        gcc-mingw-w64-ucrt64    \
        g++-mingw-w64-ucrt64    \
        gcc-mingw-w64           \
        gcc-aarch64-linux-gnu   \ 
        g++-aarch64-linux-gnu   \
        gcc-arm-linux-gnueabihf \
        g++-arm-linux-gnueabihf \
        gcc-i686-linux-gnu      \ 
        libx11-dev              \
        libxcursor-dev          \
        libxrandr-dev           \
        libxinerama-dev         \
        libxi-dev               \
        libglx-dev              \
        libgl1-mesa-dev         \
        libxxf86vm-dev &&       \
    rm -rf /var/lib/apt/lists/*

# check windows cross compiler
RUN i686-w64-mingw32-gcc --version
RUN x86_64-w64-mingw32ucrt-gcc --version
RUN x86_64-w64-mingw32ucrt-g++ --version

# check arm cross compiler
RUN arm-linux-gnueabihf-gcc --version
RUN arm-linux-gnueabihf-g++ --version
RUN aarch64-linux-gnu-gcc --version
RUN aarch64-linux-gnu-g++ --version

# check linux x86 compiler
RUN i686-linux-gnu-gcc --version

COPY go.mod go.sum .
RUN go mod download
COPY . .

RUN go clean -cache

# debian x64
RUN GOOS=linux      \
    GOARCH=amd64    \
    go build -ldflags "-s -w" -o ./app-linux64-latest

# windows x64
RUN GOOS=windows    \
    GOARCH=amd64    \ 
    CGO_ENABLED=1   \
    CC=x86_64-w64-mingw32ucrt-gcc   \
    CXX=x86_64-w64-mingw32ucrt-g++  \
    go build -ldflags "-s -w -H=windowsgui -extldflags=-static" -o ./app-win64-latest.exe

RUN mkdir /usr/app
RUN mv ./app-* /usr/app/

RUN mkdir /usr/app/logs
RUN mkdir /usr/app/uploads
RUN mkdir /usr/app/tmp

CMD ["sh", "-c", "./app-linux64-latest"]