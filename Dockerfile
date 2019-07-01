FROM golang:latest

RUN mkdir /src
WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN cd ./outputs && go build .
RUN cd ./inputs && go build .
