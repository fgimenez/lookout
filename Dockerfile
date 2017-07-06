FROM golang:1.8-alpine

ENV KUBE_LATEST_VERSION="v1.6.6"

RUN apk add --update ca-certificates \
   && apk add --update -t deps curl \
   && curl -L https://storage.googleapis.com/kubernetes-release/release/${KUBE_LATEST_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl \
   && chmod +x /usr/local/bin/kubectl \
   && apk del --purge deps \
   && rm /var/cache/apk/*

RUN mkdir -p /go/src/github.com/fgimenez/lookout

ADD main.go /go/src/github.com/fgimenez/lookout

WORKDIR /go/src/github.com/fgimenez/lookout

RUN go install github.com/fgimenez/lookout

ENTRYPOINT ["/go/bin/lookout"]
