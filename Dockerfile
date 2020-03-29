FROM golang:1.13 as build

WORKDIR /homemaker

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY *.go .
COPY cmd cmd
COPY internal internal

RUN go test -v ./...
RUN go build

COPY test test
RUN cd test && go test -c

FROM ubuntu

ENV HOMEMAKER_DOCKER_TEST_ENV defined

COPY --from=build /homemaker/homemaker /bin/
COPY --from=build /homemaker/test/test.test /bin/

WORKDIR /home/ubuntu/.config/homemaker
COPY test .
CMD ["test.test", "-ginkgo.v"] 