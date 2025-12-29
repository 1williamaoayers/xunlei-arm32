FROM --platform=$BUILDPLATFORM ubuntu:focal AS temp
ARG TARGETARCH

ENV LANG=C.UTF-8 LANGUAGE=zh_CN.UTF-8 LC_ALL=C

RUN sed -i 's/archive.ubuntu.com/mirrors.bfsu.edu.cn/g' /etc/apt/sources.list \
  && sed -i 's/security.ubuntu.com/mirrors.bfsu.edu.cn/g' /etc/apt/sources.list \
  && sed -i 's/ports.ubuntu.com/mirrors.bfsu.edu.cn/g' /etc/apt/sources.list \
  && DEBIAN_FRONTEND=noninteractive apt-get update \
  && apt-get -y --no-install-recommends install tzdata ca-certificates \
  && rm -rf /var/lib/apt/lists/* \
  && echo "Asia/Shanghai" >/etc/timezone \
  && rm -f /etc/localtime && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

RUN mkdir -p /rootfs/etc/ssl/certs /rootfs/usr/bin \
  && cp --parents /etc/ssl/certs/ca-certificates.crt /rootfs/ \
  && cp --parents /etc/timezone /rootfs/ \
  && cp --parents /etc/localtime /rootfs/

COPY bin/xlp-amd64 /rootfs/usr/bin/xlp-amd64
COPY bin/xlp-arm64 /rootfs/usr/bin/xlp-arm64
COPY bin/xlp-arm /rootfs/usr/bin/xlp-arm

RUN case "${TARGETARCH}" in \
  arm) ln -sf /usr/bin/xlp-arm /rootfs/usr/bin/xlp ;; \
  arm64) ln -sf /usr/bin/xlp-arm64 /rootfs/usr/bin/xlp ;; \
  *) ln -sf /usr/bin/xlp-amd64 /rootfs/usr/bin/xlp ;; \
  esac \
  && chmod +x /rootfs/usr/bin/xlp-*

FROM ubuntu:focal

LABEL org.opencontainers.image.source=https://github.com/1williamaoayers/xunlei-arm32

COPY --from=temp /rootfs/ /

ENV LANG=C.UTF-8 LANGUAGE=zh_CN.UTF-8 LC_ALL=C \
  XL_DEBUG=false \
  XL_DASHBOARD_PORT=2345 \
  XL_DIR_DOWNLOAD=/xunlei/downloads \
  XL_DIR_DATA=/xunlei/data \
  XL_CHROOT=/xunlei \
  XL_PREVENT_UPDATE=true

CMD [ "/usr/bin/xlp" ]
