# Stage 1 - build stage
######################################
FROM golang:1.23.3 as builder

RUN mkdir -p /src
WORKDIR /src

ADD . /src
RUN make proto-generate
RUN make build


# Stage 2 - Binary stage
######################################
FROM golang:1.21.13

ENV BINPATH /bin
WORKDIR $BINPATH/

COPY --from=builder /src/bin/propeller $BINPATH
COPY --from=builder /src/config/propeller.toml /etc/propeller.toml

RUN rm -rf /var/lib/apt/lists/* && apt-get clean && apt-get update

CMD ["./propeller"]
