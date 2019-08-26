FROM opensuse/leap



ENV HOST=0.0.0.0
ENV PORT=8830
ENV DOMAIN=vcap.me
ENV DOMAIN_REGISTER=true
ENV KIND_VERSION=0.2.1
EXPOSE 8830
RUN zypper in -y go make git docker wget
ENV GOPATH=/go/
RUN go get github.com/mudler/ekcp && go install github.com/mudler/ekcp
RUN wget https://github.com/kubernetes-sigs/kind/releases/download/${KIND_VERSION}/kind-linux-amd64 -O /usr/local/bin/kind && chmod +x /usr/local/bin/kind
ENTRYPOINT ["/go/bin/ekcp"]
