# Builder Image
FROM golang:1.10.3-alpine AS BUILDER

WORKDIR /go/src/app
COPY . .

RUN apk update && \
    apk upgrade && \
    apk add git && \
    wget -O - https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

RUN dep ensure -vendor-only
RUN CGO_ENABLED=0 go build

# Docker Image
FROM python:3-alpine

WORKDIR /app
ENV AWS_REGION eu-west-2

RUN apk add --no-cache ca-certificates 
RUN pip install requests 
COPY --from=BUILDER /go/src/app/k8ecr /app/k8ecr
COPY autodeploy.py /app/

ENTRYPOINT /app/autodeploy.py
