# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

FROM golang:1.24-alpine AS golang-builder
LABEL maintainer="linkinstar@apache.org"

ARG GOPROXY
ENV GOPROXY=${GOPROXY:-https://goproxy.cn,direct}

ENV GOPATH=/go
ENV GOROOT=/usr/local/go
ENV PACKAGE=github.com/apache/answer
ENV BUILD_DIR=${GOPATH}/src/${PACKAGE}
ENV ANSWER_MODULE=${BUILD_DIR}

ARG TAGS="sqlite sqlite_unlock_notify"
ENV TAGS="bindata timetzdata $TAGS"
ARG CGO_EXTRA_CFLAGS

COPY . ${BUILD_DIR}
WORKDIR ${BUILD_DIR}
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk --no-cache add build-base git bash nodejs npm \
    && npm config set registry https://registry.npmmirror.com \
    && npm install -g pnpm@9.7.0 \
    && pnpm config set registry https://registry.npmmirror.com \
    && pnpm config set store-dir /root/.pnpm-store \
    && cd ui \
    && rm -rf node_modules \
    && node ./scripts/importPlugins.js \
    && pnpm install --no-frozen-lockfile --registry=https://registry.npmmirror.com \
    && node ./scripts/env.js \
    && pnpm run build \
    && cd .. \
    && make clean build

RUN chmod 755 answer
RUN ["/bin/bash","-c","script/build_plugin.sh"]
RUN cp answer /usr/bin/answer

RUN mkdir -p /data/uploads && chmod 777 /data/uploads \
    && mkdir -p /data/i18n && cp -r i18n/*.yaml /data/i18n

FROM alpine
LABEL maintainer="linkinstar@apache.org"

ARG TIMEZONE
ENV TIMEZONE=${TIMEZONE:-"Asia/Shanghai"}

RUN apk update \
    && apk --no-cache add \
        bash \
        ca-certificates \
        curl \
        dumb-init \
        gettext \
        openssh \
        sqlite \
        gnupg \
        tzdata \
    && ln -sf /usr/share/zoneinfo/${TIMEZONE} /etc/localtime \
    && echo "${TIMEZONE}" > /etc/timezone

COPY --from=golang-builder /usr/bin/answer /usr/bin/answer
COPY --from=golang-builder /data /data
COPY /script/entrypoint.sh /entrypoint.sh
RUN chmod 755 /entrypoint.sh

VOLUME /data
EXPOSE 80
ENTRYPOINT ["/entrypoint.sh"]
