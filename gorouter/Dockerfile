FROM golang:1.12


EXPOSE 8081 8082

COPY gorouter.yml gorouter.yml


RUN mkdir -p $GOPATH/src/code.cloudfoundry.org/gorouter

RUN git clone https://github.com/cloudfoundry/routing-release
RUN cd routing-release && ./scripts/update && cp -rf src/* $GOPATH/src/
RUN cd $GOPATH/src/code.cloudfoundry.org/gorouter && go install code.cloudfoundry.org/gorouter


ENTRYPOINT ["/go/bin/gorouter", "-c", "gorouter.yml"]
