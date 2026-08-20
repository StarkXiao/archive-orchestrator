# Keep the Go toolchain in the evaluation image so code can be changed, built, and tested inside it.
FROM golang:1.26

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build ./...

CMD ["bash"]
