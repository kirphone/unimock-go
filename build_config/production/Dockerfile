FROM golang:latest as dev

WORKDIR /go/src/app

COPY ../../go.mod .
COPY ../../go.sum .

RUN go get -d -v ./...

COPY ../.. .

RUN go install -v ./...

CMD ["unimock"]