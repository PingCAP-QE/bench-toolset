FROM golang:1.14 as builder
MAINTAINER 5kbpers
ENV GO111MODULE=on
WORKDIR /src
COPY . .
RUN make build

FROM golang:1.14 as tpcbuilder
ENV GO111MODULE=on
RUN mkdir -p /go/src/github.com/pingcap/go-tpc
WORKDIR /go/src/github.com/pingcap/go-tpc
RUN git clone https://github.com/pingcap/go-tpc.git .
RUN make build

FROM golang:1.14 as ycsbbuilder
ENV GO111MODULE=on
RUN mkdir -p /go/src/github.com/pingcap/go-ycsb
WORKDIR /go/src/github.com/pingcap/go-ycsb
RUN git clone https://github.com/pingcap/go-ycsb.git .
RUN GO111MODULE=on go build -o bin/go-ycsb ./cmd/*

FROM golang:1.14 as brbuilder
ENV GO111MODULE=on
RUN mkdir -p /go/src/github.com/pingcap/br
WORKDIR /go/src/github.com/pingcap/br
RUN git clone https://github.com/pingcap/br.git .
RUN make build

FROM golang:1.14
RUN curl -s https://packagecloud.io/install/repositories/akopytov/sysbench/script.deb.sh | bash
RUN apt -y install sysbench default-mysql-client
RUN mkdir -p /ycsb/workloads
COPY --from=builder /src/bin/* /bin/
COPY --from=tpcbuilder /go/src/github.com/pingcap/go-tpc/bin/* /bin/
COPY --from=ycsbbuilder /go/src/github.com/pingcap/go-ycsb/bin/* /bin/
COPY --from=ycsbbuilder /go/src/github.com/pingcap/go-ycsb/workloads/* /ycsb/workloads/
COPY --from=brbuilder /go/src/github.com/pingcap/br/bin/* /bin/
ENV PATH="$PATH:/bin"
