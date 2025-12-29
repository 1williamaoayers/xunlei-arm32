FROM --platform=$BUILDPLATFORM ubuntu:focal AS temp
ARG TARGETARCH

ENV LANG=C.UTF-8 LANG=zh_CN.UTF-8 LANGUAGE=zh_CN.UTF-8 LC_ALL=C

RUN sed -i 's/archive.ubuntu.com/mirrors.bfsu.edu.cn/g' /etc/apt/sources.list \
  && sed -i 's/security.ubuntu.com/mirrors.bfsu.edu.cn/g' /etc/apt/sources.list \
  && sed -i 's/ports.ubuntu.com/mirrors.bfsu.edu.cn/g' /etc/apt/sources.list \
  && DEBIAN_FRONTEND=noninteractive apt-get update && apt-get -y --no-install-recommends install tzdata ca-certificates xz-utils \
  && rm -rf /var/lib/apt/lists/* \
  && echo "Asia/Shanghai" >/etc/timezone \
  && rm -f /etc/localtime && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

RUN mkdir -p /rootfs/etc/ssl/certs /rootfs/usr/bin \
  && cp --parents /etc/ssl/certs/ca-certificates.crt /rootfs/ \
  && cp --parents /etc/timezone /rootfs/ \
  && cp --parents /etc/localtime /rootfs/ \
  && cp --parents /etc/apt/sources.list /rootfs/

# 复制所有二进制文件
COPY bin/xlp-amd64 /rootfs/usr/bin/xlp-amd64
COPY bin/xlp-arm64 /rootfs/usr/bin/xlp-arm64
COPY bin/xlp-arm /rootfs/usr/bin/xlp-arm

# 根据架构创建 xlp 符号链接
RUN case "${TARGETARCH}" in \
  arm) ln -sf /usr/bin/xlp-arm /rootfs/usr/bin/xlp ;; \
  arm64) ln -sf /usr/bin/xlp-arm64 /rootfs/usr/bin/xlp ;; \
  *) ln -sf /usr/bin/xlp-amd64 /rootfs/usr/bin/xlp ;; \
  esac \
  && chmod +x /rootfs/usr/bin/xlp-*

FROM ubuntu:focal

LABEL org.opencontainers.image.authors=1williamaoayers
LABEL org.opencontainers.image.source=https://github.com/1williamaoayers/xunlei-arm32

ARG TARGETARCH

COPY --from=temp /rootfs/ /

# ARM32 需要安装 QEMU 来运行 ARM64 二进制
RUN if [ "${TARGETARCH}" = "arm" ]; then \
  sed -i 's/archive.ubuntu.com/mirrors.bfsu.edu.cn/g' /etc/apt/sources.list \
  && sed -i 's/ports.ubuntu.com/mirrors.bfsu.edu.cn/g' /etc/apt/sources.list \
  && apt-get update \
  && apt-get install -y --no-install-recommends qemu-user-static \
  && rm -rf /var/lib/apt/lists/*; \
  fi

ENV LANG=C.UTF-8 LANG=zh_CN.UTF-8 LANGUAGE=zh_CN.UTF-8 LC_ALL=C \
  XL_DEBUG=false \
  XL_DASHBOARD_PORT=2345 \
  XL_DASHBOARD_USERNAME= \
  XL_DIR_DOWNLOAD=/xunlei/downloads \
  XL_DIR_DATA=/xunlei/data \
  XL_DASHBOARD_PASSWORD= \
  XL_CHROOT=/xunlei \
  XL_PREVENT_UPDATE=true \
  XL_UID= \
  XL_GID=

CMD [ "/usr/bin/xlp" ]
