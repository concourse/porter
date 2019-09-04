FROM concourse/golang-builder as builder

RUN mkdir /src
WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download
RUN go install github.com/onsi/ginkgo/ginkgo
COPY . .
RUN ./scripts/test
RUN cd ./watch && go build .
RUN cd ./outputs && go build .
RUN cd ./inputs && go build .
RUN cd ./logstream && go build .

FROM ubuntu:bionic AS porter

RUN apt-get update && apt-get -y install \
      iproute2 \
      ca-certificates \
      file \
      dumb-init

COPY --from=builder /src/outputs /opt/porter/out
COPY --from=builder /src/inputs /opt/porter/in
COPY --from=builder /src/logstream /opt/porter/logstream
RUN chmod +x /opt/porter/*

FROM porter