FROM golang:1.14.0-buster AS builder
WORKDIR /go/src/app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo


FROM scratch
WORKDIR /
COPY --from=builder /go/src/app/vehicle-api .
CMD ["/vehicle-api"]
