FROM golang:1.12
ENV GOPROXY=https://athens.i.ddex.io
WORKDIR /app
COPY . /app
RUN go build -o main -v -installsuffix cgo -ldflags '-s -w' .

FROM alpine
RUN apk --no-cache add ca-certificates
COPY --from=0 /app/main /bin/app
CMD ["app"]
