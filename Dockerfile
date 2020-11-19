FROM golang:1.14-alpine as builder
MAINTAINER 5kbpers
RUN apk add --no-cache \
    make \
    git \
    bash \
    curl \
    gcc \
    g++

ENV GO111MODULE=on
WORKDIR /src
COPY . .
RUN make build

FROM golang:1.14-alpine as tpcbuilder
RUN apk add --no-cache \
    make \
    git \
    bash \
    curl \
    gcc \
    g++
ENV GO111MODULE=on
RUN mkdir -p /go/src/github.com/pingcap/go-tpc
WORKDIR /go/src/github.com/pingcap/go-tpc
RUN git clone https://github.com/pingcap/go-tpc.git .
RUN make build

FROM golang:1.14-alpine as ycsbbuilder
RUN apk add --no-cache \
    make \
    git \
    bash \
    curl \
    gcc \
    g++
ENV GO111MODULE=on
RUN mkdir -p /go/src/github.com/pingcap/go-ycsb
WORKDIR /go/src/github.com/pingcap/go-ycsb
RUN git clone https://github.com/pingcap/go-ycsb.git .
RUN make build


FROM severalnines/sysbench:latest
COPY --from=builder /src/bin/* /bin/
COPY --from=tpcbuilder /go/src/github.com/pingcap/go-tpc/bin/* /bin/
COPY --from=ycsbbuilder /go/src/github.com/pingcap/go-ycsb/bin/* /bin/
ENV PATH="$PATH:/bin"

CMD ["/bin/stability_bench"]
