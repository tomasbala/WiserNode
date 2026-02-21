# 视频平台项目

支持ONVIF、GB28181、RTSP协议的视频平台，提供C++ API接口。

## 环境要求

### 必要依赖
1. **Go语言**：1.18+  
   - 下载地址：https://golang.org/dl/
   - 安装后需配置环境变量

2. **GStreamer**：1.18+  
   - 下载地址：https://gstreamer.freedesktop.org/download/
   - 安装时需选择"Complete"安装

3. **Protocol Buffers**：3.19+  
   - 下载地址：https://github.com/protocolbuffers/protobuf/releases

4. **gRPC**：
   - Go gRPC：`go get google.golang.org/grpc`
   - C++ gRPC：需单独安装

### 可选依赖
1. **PostgreSQL**：13+  
   - 用于设备信息存储

2. **Redis**：6+  
   - 用于缓存和会话管理

## 项目结构

```
video-platform/
├── cmd/
│   └── server/
│       └── main.go           # 服务启动入口
├── internal/
│   ├── api/                  # API服务层
│   │   ├── grpc/             # gRPC服务
│   │   └── http/             # RESTful API
│   ├── protocols/            # 协议支持层
│   │   ├── onvif/            # ONVIF协议
│   │   ├── gb28181/          # GB28181协议
│   │   └── rtsp/             # RTSP协议
│   ├── services/             # 核心服务层
│   │   ├── device/           # 设备管理服务
│   │   ├── stream/           # 流媒体服务
│   │   └── session/          # 会话管理服务
│   └── models/               # 数据模型
├── proto/                    # gRPC协议定义
│   └── video_platform.proto  # 服务接口定义
├── cpp/                      # C++接口
│   ├── include/              # 头文件
│   ├── src/                  # 源码
│   └── build/                # 构建目录
├── go.mod                    # Go依赖管理
├── go.sum                    # 依赖校验
└── README.md                 # 项目说明
```

## 快速开始

### 1. 安装依赖

```bash
# 安装Go依赖
go mod tidy

# 安装GStreamer
# 参考：https://gstreamer.freedesktop.org/documentation/installing/index.html

# 安装Protocol Buffers
# 参考：https://github.com/protocolbuffers/protobuf#installation
```

### 2. 构建项目

```bash
# 生成gRPC代码
protoc --go_out=. --go-grpc_out=. proto/video_platform.proto

# 构建服务
go build -o bin/server cmd/server/main.go

# 构建C++接口（需要CMake）
cd cpp/build
cmake ..
cmake --build .
```

### 3. 运行服务

```bash
# 启动服务
./bin/server

# 默认端口：
# - gRPC: 50051
# - HTTP: 8080
```

### 4. C++客户端调用

```cpp
// 参考 cpp/src/example.cpp
```

## API文档

### gRPC服务

- **设备管理**：发现设备、添加设备、获取设备状态
- **流媒体管理**：启动流、停止流、获取流地址
- **会话管理**：创建会话、销毁会话、获取会话状态

### RESTful API

- `GET /api/devices` - 获取设备列表
- `POST /api/devices` - 添加设备
- `GET /api/devices/{id}` - 获取设备详情
- `DELETE /api/devices/{id}` - 删除设备
- `POST /api/streams` - 启动流
- `DELETE /api/streams/{id}` - 停止流
- `GET /api/streams/{id}/url` - 获取流地址

## 技术方案

详见技术方案文档。

## 注意事项

1. **设备认证**：不同协议的设备认证方式不同，请参考对应协议文档
2. **网络配置**：确保设备和服务在同一网络或可访问
3. **性能优化**：对于大量设备，建议配置Redis缓存和负载均衡
4. **安全配置**：生产环境请配置TLS和访问控制

## 故障排查

### 常见问题

1. **设备无法发现**：检查网络连接和设备网络设置
2. **流无法播放**：检查设备编码格式和GStreamer插件
3. **API调用失败**：检查服务是否运行和权限设置

### 日志查看

服务日志默认输出到标准输出，可通过配置文件修改。

## 联系方式

如有问题，请联系技术支持。