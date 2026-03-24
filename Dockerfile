FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o psl .

FROM alpine:3.20
COPY --from=builder /app/psl /usr/local/bin/psl
ENTRYPOINT ["psl"]
