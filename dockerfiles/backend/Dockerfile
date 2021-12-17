FROM golang:1.17 AS builder

WORKDIR /go/src/github.com/dtan4/grpc-chat

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o server github.com/dtan4/grpc-chat/backend/cmd/server

FROM gcr.io/distroless/static:nonroot

COPY --from=builder /go/src/github.com/dtan4/grpc-chat/server /server

EXPOSE 50051

CMD ["/server"]
