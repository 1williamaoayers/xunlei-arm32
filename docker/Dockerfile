# 迅雷 ARM32 原生客户端 Docker 镜像
# 无需 QEMU，直接运行

FROM alpine:3.19

LABEL org.opencontainers.image.source=https://github.com/1williamaoayers/xunlei-arm32
LABEL org.opencontainers.image.description="迅雷云盘 ARM32 原生客户端"

ARG TARGETARCH

# 安装基础依赖
RUN apk add --no-cache ca-certificates tzdata \
  && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
  && echo "Asia/Shanghai" > /etc/timezone

# 复制对应架构的二进制
COPY bin/xunlei-native-amd64 /usr/bin/xunlei-native-amd64
COPY bin/xunlei-native-arm64 /usr/bin/xunlei-native-arm64
COPY bin/xunlei-native-arm /usr/bin/xunlei-native-arm

# 根据架构创建符号链接
RUN case "${TARGETARCH}" in \
  arm) ln -sf /usr/bin/xunlei-native-arm /usr/bin/xunlei ;; \
  arm64) ln -sf /usr/bin/xunlei-native-arm64 /usr/bin/xunlei ;; \
  *) ln -sf /usr/bin/xunlei-native-amd64 /usr/bin/xunlei ;; \
  esac \
  && chmod +x /usr/bin/xunlei-native-*

# 创建数据目录
RUN mkdir -p /xunlei/data /xunlei/downloads

# 环境变量
ENV XL_PORT=2345 \
  XL_DATA_DIR=/xunlei/data

EXPOSE 2345

VOLUME ["/xunlei/data", "/xunlei/downloads"]

CMD ["/usr/bin/xunlei"]
