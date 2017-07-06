FROM golang:1.8

ENV KUBE_LATEST_VERSION="v1.6.6"

RUN apt update && apt install -y --no-install-recommends ca-certificates curl \
   && curl -L https://storage.googleapis.com/kubernetes-release/release/${KUBE_LATEST_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl \
   && chmod +x /usr/local/bin/kubectl \
   && curl -L https://storage.googleapis.com/kubernetes-helm/helm-v2.4.2-linux-amd64.tar.gz  | tar xzf - linux-amd64/helm \
   && chmod +x ./linux-amd64/helm \
   && mv ./linux-amd64/helm /bin/helm \
   && rm -rf ./linux-amd64 \
   && mkdir -p /root/.helm/plugins/ \
   && curl -L https://github.com/app-registry/helm-plugin/releases/download/v0.3.7/registry-helm-plugin-v0.3.7-dev-linux-x64.tar.gz | tar xzf - registry \
   && mv ./registry /root/.helm/plugins/ \
   && apt-get remove -y curl \
   && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /go/src/github.com/fgimenez/lookout

ADD main.go /go/src/github.com/fgimenez/lookout

WORKDIR /go/src/github.com/fgimenez/lookout

RUN go install github.com/fgimenez/lookout

ENTRYPOINT ["/go/bin/lookout"]
