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
RUN make build

FROM golang:1.14 as brbuilder
ENV GO111MODULE=on
RUN mkdir -p /go/src/github.com/pingcap/br
WORKDIR /go/src/github.com/pingcap/br
RUN git clone https://github.com/pingcap/br.git .
RUN git checkout release-4.0
RUN make build

FROM perconalab/sysbench
COPY --from=builder /src/bin/* /bin/
COPY --from=tpcbuilder /go/src/github.com/pingcap/go-tpc/bin/* /bin/
COPY --from=ycsbbuilder /go/src/github.com/pingcap/go-ycsb/bin/* /bin/
COPY --from=brbuilder /go/src/github.com/pingcap/br/bin/* /bin/
ENV PATH="$PATH:/bin"
