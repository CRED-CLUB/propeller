# Stage 1 - build stage
######################################
FROM golang:1.23.3 AS builder

RUN mkdir -p /src
WORKDIR /src

ADD . /src
RUN make proto-generate
RUN make build


# Stage 2 - Binary stage
######################################
FROM alpine:latest

# Install dependencies to run the binary
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    libc6-compat \
    gcompat

RUN mkdir -p /src/config

ENV BINPATH=/bin
WORKDIR $BINPATH

COPY --from=builder /src/bin/propeller $BINPATH
COPY --from=builder /src/config/propeller.toml /src/config/propeller.toml

CMD ["./propeller"]
