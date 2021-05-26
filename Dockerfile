FROM golang:1.15

ENV CONF_FILE_PATH=conf/docker/app.ini GO111MODULE=on GOPROXY=https://goproxy.cn GOSUMDB=off

RUN mkdir -p $GOPATH/src/github.com/Walker-PI/gateway-admin

COPY . $GOPATH/src/github.com/Walker-PI/gateway-admin

WORKDIR $GOPATH/src/github.com/Walker-PI/gateway-admin

RUN sh build.sh

EXPOSE 9922

ENTRYPOINT ["./output/bootstrap.sh"]