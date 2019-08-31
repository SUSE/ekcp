FROM opensuse/leap

ENV GOPATH=/go/
ADD . $GOPATH/src/github.com/mudler/ekcp

ENV HOST=0.0.0.0
ENV PORT=8830
ENV DOMAIN=nip.io
ENV DOMAIN_REGISTER=true
ENV KIND_VERSION=0.2.1
EXPOSE 8830
RUN zypper in -y go make git docker wget
RUN go install github.com/mudler/ekcp
RUN wget https://github.com/kubernetes-sigs/kind/releases/download/${KIND_VERSION}/kind-linux-amd64 -O /usr/local/bin/kind && chmod +x /usr/local/bin/kind
ENTRYPOINT ["/go/bin/ekcp"]
