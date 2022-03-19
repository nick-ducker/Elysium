ARG GO_VERSION=1.17
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

COPY ./cmd/elysium-main .
COPY go.* ./

RUN go get
RUN go build -o ./elysium

FROM alpine:3.15

RUN apk update && apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/elysium .

CMD [ "./elysium" ]