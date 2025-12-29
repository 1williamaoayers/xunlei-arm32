//go:build arm && !arm64

package spk

import (
	_ "embed"
)

// ARM32 构建使用 ARM64 的 SPK 包
// 运行时通过 QEMU 用户态模拟执行 ARM64 二进制

//go:embed nasxunlei-armv8.spk
var Bytes []byte

//go:embed nasxunlei-armv8.txt
var Version string
