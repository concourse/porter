FROM concourse/golang-builder as builder

RUN mkdir /src
WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download
RUN go install github.com/onsi/ginkgo/ginkgo
COPY . .
RUN ./scripts/test
RUN cd ./outputs && go build .
RUN cd ./inputs && go build .


FROM ubuntu:bionic AS porter

COPY --from=builder /src/outputs /opt/porter/out
COPY --from=builder /src/inputs /opt/porter/in
RUN chmod +x /opt/porter/*

FROM porter