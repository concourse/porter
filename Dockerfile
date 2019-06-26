FROM golang:latest

RUN mkdir /src
WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build outputpusher
RUN go build inputpuller
