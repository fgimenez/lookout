FROM golang:1.8-alpine

ENV KUBE_LATEST_VERSION="v1.6.6"

RUN apk add --update ca-certificates \
   && apk add --update -t deps curl \
   && curl -L https://storage.googleapis.com/kubernetes-release/release/${KUBE_LATEST_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl \
   && chmod +x /usr/local/bin/kubectl \
   && curl -L https://storage.googleapis.com/kubernetes-helm/helm-v2.4.2-linux-amd64.tar.gz  | tar xzf - linux-amd64/helm \
   && chmod +x ./linux-amd64/helm \
   && mv ./linux-amd64/helm /bin/helm \
   && rm -rf ./linux-amd64 \
   && mkdir -p ~/.helm/plugins/ \
   && curl -L https://github.com/app-registry/helm-plugin/releases/download/v0.3.7/registry-helm-plugin-v0.3.7-dev-linux-x64.tar.gz | tar xzf - registry \
   && mv ./registry ~/.helm/plugins/ \
   && apk del --purge deps \
   && rm /var/cache/apk/*

RUN mkdir -p /go/src/github.com/fgimenez/lookout

ADD main.go /go/src/github.com/fgimenez/lookout

WORKDIR /go/src/github.com/fgimenez/lookout

RUN go install github.com/fgimenez/lookout

ENTRYPOINT ["/go/bin/lookout"]
