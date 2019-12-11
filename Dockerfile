FROM node:11.1.0-alpine as builder
RUN mkdir -p /usr/src/app
RUN apk update && apk add git make

RUN git clone https://github.com/marmelab/react-admin.git /usr/src/app/react-admin
RUN cd /usr/src/app/react-admin && make install
#RUN npm link react-admin/packages/react-admin
#RUN yarn install
RUN rm -rf /usr/src/app/react-admin/examples/simple
ADD ./ui /usr/src/app/react-admin/examples/simple
WORKDIR /usr/src/app/react-admin/examples/simple
RUN cd /usr/src/app/react-admin && make install
RUN cd /usr/src/app/react-admin/examples/simple && yarn build


FROM opensuse/leap

ENV GOPATH=/go/

ENV HOST=0.0.0.0
ENV PORT=8830
ENV DOMAIN=nip.io
ENV DOMAIN_REGISTER=true
ARG KIND_VERSION=v0.6.1
EXPOSE 8830
RUN zypper in -y go make git docker wget
RUN wget https://github.com/kubernetes-sigs/kind/releases/download/${KIND_VERSION}/kind-linux-amd64 -O /usr/local/bin/kind && chmod +x /usr/local/bin/kind

ADD . $GOPATH/src/github.com/mudler/ekcp
RUN go install github.com/mudler/ekcp
WORKDIR /go/bin
RUN mkdir -p public
COPY --from=builder /usr/src/app/react-admin/examples/simple/dist ./public/ui
ENTRYPOINT ["/go/bin/ekcp"]
