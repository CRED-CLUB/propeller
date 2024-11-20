# Stage 1 - build stage
######################################
FROM golang:1.21.13 as builder

RUN mkdir -p /src
WORKDIR /src

ADD . /src
RUN make proto-generate
RUN make build


# Stage 2 - Binary stage
######################################
FROM golang:1.21.13

ENV BINPATH /bin
ENV WORKDIR $BINPATH
WORKDIR $BINPATH/

COPY --from=builder /src/bin/propeller $BINPATH
COPY config/ $BINPATH/config/

RUN rm -rf /var/lib/apt/lists/* && apt-get clean && apt-get update

CMD ["./propeller"]
