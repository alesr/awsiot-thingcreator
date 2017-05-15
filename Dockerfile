FROM golang:1.7-wheezy

RUN apt-get update && \
    apt-get install -y git-core && \
    apt-get clean

# AWS credentials
COPY .aws /root/.aws

WORKDIR $GOPATH/src/github.com/nesto/awsiot-thingcreator

COPY things/ things/
COPY main.go .

# Go get package dependencies
RUN go get -d ./...

RUN mkdir certificates

RUN go build

ENTRYPOINT ["./awsiot-thingcreator"]
