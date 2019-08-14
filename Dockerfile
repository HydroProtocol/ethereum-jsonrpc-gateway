FROM golang:1.12
ENV GOPROXY=https://athens.i.ddex.io
WORKDIR /app
COPY . /app
RUN go build -o main -v -installsuffix cgo -ldflags '-s -w' .

FROM alpine
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
RUN apk --no-cache add ca-certificates
COPY --from=0 /app/main /bin/app
CMD ["app"]
