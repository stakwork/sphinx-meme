# build stage
FROM golang as builder

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build

# final stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates openssl

COPY --from=builder /app/sphinx-meme /app/
COPY --from=builder /app/.env /app/

EXPOSE 5000
ENTRYPOINT ["/app/sphinx-meme"]