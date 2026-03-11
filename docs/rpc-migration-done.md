# RPC 迁移完成文档

本文档记录编译 `api/proto/` 中所有 Protobuf 文件的脚本指令。

## Proto 文件列表

当前项目包含以下 Protobuf 文件：

```
api/proto/
├── common/
│   └── types.proto
└── ai/
    ├── asr.proto
    ├── nlu.proto
    └── tts.proto
```

## 编译指令

### 1. 安装依赖

```bash
# 安装 protoc 编译器
# Ubuntu/Debian
sudo apt-get install -y protobuf-compiler

# macOS
brew install protobuf

# 验证安装
protoc --version
```

### 2. 安装 Go 插件

```bash
# 安装 protoc-gen-go 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# 安装 protoc-gen-go-grpc 插件（如需 gRPC）
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 确保 $GOPATH/bin 在 PATH 中
export PATH=$PATH:$(go env GOPATH)/bin
```

### 3. 编译所有 Proto 文件

```bash
# 进入项目根目录
cd /home/pp/go_repo/Yunyez

# 创建输出目录
mkdir -p pkg/proto

# 编译所有 proto 文件
protoc --go_out=pkg/proto --go_opt=paths=source_relative \
       --go-grpc_out=pkg/proto --go-grpc_opt=paths=source_relative \
       api/proto/common/types.proto \
       api/proto/ai/asr.proto \
       api/proto/ai/nlu.proto \
       api/proto/ai/tts.proto
```

### 4. 批量编译脚本

```bash
#!/bin/bash
# scripts/compile-proto.sh

set -e

PROTO_DIR="api/proto"
OUT_DIR="pkg/proto"

echo "🔨 开始编译 Protobuf 文件..."

# 创建输出目录
mkdir -p "$OUT_DIR"

# 查找所有 proto 文件并编译
find "$PROTO_DIR" -name "*.proto" | while read -r file; do
    echo "  编译：$file"
    protoc --go_out="$OUT_DIR" --go_opt=paths=source_relative \
           --go-grpc_out="$OUT_DIR" --go-grpc_opt=paths=source_relative \
           "$file"
done

echo "✅ Protobuf 编译完成！"
echo "📁 输出目录：$OUT_DIR"
```

### 5. 一键执行

```bash
chmod +x scripts/compile-proto.sh
./scripts/compile-proto.sh
```

## 编译产物

编译后生成的文件结构：

```
pkg/proto/
├── common/
│   └── types.pb.go
└── ai/
    ├── asr.pb.go
    ├── asr_grpc.pb.go
    ├── nlu.pb.go
    ├── nlu_grpc.pb.go
    ├── tts.pb.go
    └── tts_grpc.pb.go
```

## 验证编译结果

```bash
# 检查生成的文件
ls -la pkg/proto/common/
ls -la pkg/proto/ai/

# 验证 Go 代码可编译
go build ./pkg/proto/...
```

## 注意事项

1. **版本号**: 确保 `protoc` 与 `protoc-gen-go` 版本兼容
2. **路径配置**: 确保 `$GOPATH/bin` 在 `PATH` 环境变量中
3. **import 路径**: proto 文件中的 `import` 语句需使用相对路径
4. **package 命名**: Go package 名应与目录结构对应

## 迁移完成确认

- [x] 所有 proto 文件已识别
- [x] 编译脚本已编写
- [ ] 编译脚本已执行
- [ ] 生成的 Go 代码已验证
- [ ] 单元测试已通过

---

**文档创建日期**: 2026 年 3 月 12 日
