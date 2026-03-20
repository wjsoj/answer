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

ARG REGISTRY=docker.xuanyuan.run
FROM ${REGISTRY}/golang:1.24-alpine AS golang-builder
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

# Install system tools (cached as long as packages don't change)
RUN --mount=type=cache,target=/var/cache/apk \
    sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add build-base git bash nodejs npm \
    && npm config set registry https://registry.npmmirror.com \
    && npm install -g pnpm@9.7.0

WORKDIR ${BUILD_DIR}

# Copy only files needed for pnpm install — changing Go source won't bust this layer
COPY ui/package.json ui/pnpm-lock.yaml ui/
COPY ui/scripts/ ui/scripts/
COPY ui/src/plugins/ ui/src/plugins/

# Install npm deps (cached when package.json / lockfile unchanged)
RUN --mount=type=cache,target=/root/.pnpm-store \
    pnpm config set registry https://registry.npmmirror.com \
    && pnpm config set store-dir /root/.pnpm-store \
    && cd ui \
    && node ./scripts/importPlugins.js \
    && pnpm install --no-frozen-lockfile --registry=https://registry.npmmirror.com

# Copy full source (node_modules persists from above layer)
COPY . ${BUILD_DIR}

# Build UI and Go binary with persistent caches
RUN --mount=type=cache,target=/root/.pnpm-store \
    --mount=type=cache,target=/root/.cache/go-build \
    pnpm config set store-dir /root/.pnpm-store \
    && cd ui \
    && node ./scripts/importPlugins.js \
    && node ./scripts/env.js \
    && pnpm run build \
    && cd .. \
    && make clean build

RUN chmod 755 answer
RUN find ui/src/plugins -name "node_modules" -type d -exec rm -rf {} + 2>/dev/null || true
RUN --mount=type=cache,target=/root/.cache/go-build \
    /bin/bash script/build_plugin.sh
RUN cp answer /usr/bin/answer

RUN mkdir -p /data/uploads && chmod 777 /data/uploads \
    && mkdir -p /data/i18n && cp -r i18n/*.yaml /data/i18n

ARG REGISTRY=docker.xuanyuan.run
FROM ${REGISTRY}/alpine
LABEL maintainer="linkinstar@apache.org"

ARG TIMEZONE
ENV TIMEZONE=${TIMEZONE:-"Asia/Shanghai"}

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk update \
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
