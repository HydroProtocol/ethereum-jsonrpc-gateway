FROM golang:1.13 as builder
WORKDIR /app
COPY . /app
RUN go build -v -installsuffix cgo -ldflags '-s -w' -o /app/ethereum-jsonrpc-gateway main.go

FROM alpine
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/ethereum-jsonrpc-gateway /app/ethereum-jsonrpc-gateway
# COPY --from=builder /app/config.json /app/config.json
WORKDIR /app
RUN addgroup -S appuser && adduser -S -G appuser appuser
USER appuser
ENTRYPOINT [ "/app/ethereum-jsonrpc-gateway" ]
CMD [ "start" ]
