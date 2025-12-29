#!/bin/sh
# ============================================================
# 玩客云 OpenWrt 迅雷一键安装脚本
# 适用于: 玩客云一代 (S805 ARM32) + OpenWrt
# ============================================================

set -e

# 配置
INSTALL_DIR="/opt/xunlei"
DATA_DIR="/opt/xunlei/data"
DOWNLOAD_DIR="/mnt/sda1/downloads"
RELEASE_URL="https://github.com/1williamaoayers/xunlei-arm32/releases/latest/download"

echo "============================================"
echo "  玩客云 OpenWrt 迅雷安装脚本"
echo "============================================"

# 检查架构
ARCH=$(uname -m)
if [ "$ARCH" != "armv7l" ] && [ "$ARCH" != "armv7" ]; then
    echo "警告: 当前架构为 $ARCH，此脚本针对 ARM32 优化"
fi

# 安装依赖
echo "[1/5] 安装依赖..."
opkg update
opkg install qemu-aarch64 ca-certificates wget

# 创建目录
echo "[2/5] 创建目录..."
mkdir -p "$INSTALL_DIR" "$DATA_DIR" "$DOWNLOAD_DIR"

# 下载二进制
echo "[3/5] 下载迅雷程序..."
wget -O "$INSTALL_DIR/xlp" "$RELEASE_URL/xlp-arm"
chmod +x "$INSTALL_DIR/xlp"

# 创建启动脚本
echo "[4/5] 创建系统服务..."
cat > /etc/init.d/xunlei << 'INITEOF'
#!/bin/sh /etc/rc.common

START=99
STOP=10
USE_PROCD=1

INSTALL_DIR="/opt/xunlei"
DATA_DIR="/opt/xunlei/data"
DOWNLOAD_DIR="/mnt/sda1/downloads"

start_service() {
    procd_open_instance
    procd_set_param command "$INSTALL_DIR/xlp"
    procd_append_param command --dir_data="$DATA_DIR"
    procd_append_param command --dir_download="$DOWNLOAD_DIR"
    procd_append_param command --chroot=/
    procd_set_param respawn
    procd_set_param stdout 1
    procd_set_param stderr 1
    procd_close_instance
}

stop_service() {
    killall xlp 2>/dev/null || true
}
INITEOF

chmod +x /etc/init.d/xunlei

# 启用服务
echo "[5/5] 启用服务..."
/etc/init.d/xunlei enable
/etc/init.d/xunlei start

# 获取 IP
IP=$(ip addr show br-lan 2>/dev/null | grep 'inet ' | awk '{print $2}' | cut -d/ -f1 | head -1)
if [ -z "$IP" ]; then
    IP=$(ip addr show eth0 2>/dev/null | grep 'inet ' | awk '{print $2}' | cut -d/ -f1 | head -1)
fi

echo ""
echo "============================================"
echo "  安装完成!"
echo "============================================"
echo ""
echo "  访问地址: http://${IP:-<你的IP>}:2345"
echo "  数据目录: $DATA_DIR"
echo "  下载目录: $DOWNLOAD_DIR"
echo ""
echo "  管理命令:"
echo "    启动: /etc/init.d/xunlei start"
echo "    停止: /etc/init.d/xunlei stop"
echo "    重启: /etc/init.d/xunlei restart"
echo "    状态: /etc/init.d/xunlei status"
echo ""
