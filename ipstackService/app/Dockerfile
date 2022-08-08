FROM golang:alpine

COPY go.mod .
COPY go.sum .
ENV GOPATH=/
RUN go mod download

#build appliction
COPY . .
RUN go build -o ipstack ./cmd/main/app.go

CMD ["./ipstack"]