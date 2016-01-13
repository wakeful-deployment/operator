FROM buildpack-deps:jessie-curl

MAINTAINER Nathan Herald <nathan.herald@microsoft.com>

RUN apt-get update \
 && mkdir /opt/app \
 && mkdir /opt/app/bin \
 && mkdir /opt/src

RUN apt-get install -y --no-install-recommends g++ gcc libc6-dev make

ADD ./vendor/docker/docker-latest /usr/bin/docker
RUN chmod +x /usr/bin/docker

ENV GOLANG_VERSION 1.5.2
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_SHA1 cae87ed095e8d94a81871281d35da7829bd1234e

RUN curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz \
 && echo "$GOLANG_DOWNLOAD_SHA1  golang.tar.gz" | sha1sum -c - \
 && tar -C /usr/local -xzf golang.tar.gz \
 && rm golang.tar.gz

ENV GOPATH /opt
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

ADD . /opt/src/github.com/wakeful-deployment/operator

RUN cd /opt/src/github.com/wakeful-deployment/operator \
 && go build \
 && mv ./operator /opt/app/

ARG sha
ARG start

RUN echo $start > /opt/start \
 && chmod +x /opt/start

RUN echo $sha > /opt/app/sha

CMD /opt/start

