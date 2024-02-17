FROM golang:1.22.0-alpine3.19 AS builder
RUN apk add --no-progress --no-cache gcc musl-dev
RUN apk add --no-cache ca-certificates
RUN mkdir -p /apitest
ADD . /apitest
WORKDIR /apitest

RUN go build -ldflags "-s -w" -o . .

# MANAGER API STAGE
FROM alpine as apitest
WORKDIR /app
COPY --from=builder /apitest/restapitest .
ENTRYPOINT ["./restapitest"]
EXPOSE 9090
