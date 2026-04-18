# GB28181 SIP Server

基于 Go + ZLMediaKit 的 GB28181 国标视频平台服务器

参考 wvp-pro 核心实现，提供精简的 Go 版本。

## 项目结构

```
.
├── cmd/server/main.go              # 主程序入口
├── configs/config.yaml             # 配置文件
├── internal/
│   ├── api/http/                   # HTTP API 服务
│   │   ├── handler/                # 请求处理器
│   │   │   ├── device.go           # 设备管理
│   │   │   ├── stream.go           # 流媒体控制
│   │   │   ├── ptz.go              # 云台控制
│   │   │   └── cascade.go          # 级联管理
│   │   └── server.go               # HTTP 路由
│   ├── config/config.go            # 配置加载
│   ├── models/
│   │   ├── models.go               # 数据模型
│   │   ├── record.go               # 录像模型
│   │   ├── cascade.go              # 级联模型
│   │   └── sip.go                  # SIP 消息模型
│   └── services/
│       ├── device/service.go       # 设备管理服务
│       ├── sip/
│       │   ├── server.go           # SIP 服务端
│       │   ├── message.go          # SIP 消息构建
│       │   └── auth.go             # Digest MD5 鉴权
│       ├── stream/
│       │   ├── service.go          # 流媒体服务
│       │   └── zlm.go              # ZLMediaKit API 封装
│       ├── record/service.go       # 录像回放服务
│       └── cascade/service.go      # 级联对接服务
├── go.mod
└── README.md
```

## 技术栈

- **Go 1.21+**: GB28181 SIP 信令服务、设备管理、HTTP API
- **ZLMediaKit**: C++ 流媒体服务器，负责 RTP 收流、PS 解封装、转 FLV/WebRTC/RTSP
- **Gin**: HTTP Web 框架
- **Viper**: 配置管理

## 核心功能

### 设备管理
- SIP UDP 5060 端口监听
- 设备注册、Digest MD5 鉴权
- 设备心跳保活、在线状态管理
- Catalog 目录获取、通道管理

### 实时视频
- 实时视频预览（INVITE）
- 支持 FLV/RTSP/HLS/WebRTC 多协议输出

### 录像回放
- 历史录像查询（RecordInfo）
- 录像回放（INVITE Playback）
- 回放控制

### 云台控制
- PTZ 控制（上下左右、变倍变焦）

### 级联对接
- 作为上级平台：推送目录、转发视频
- 作为下级平台：注册到上级、接收指令

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 配置 ZLMediaKit

确保 ZLMediaKit 已启动，并修改 `configs/config.yaml` 中的配置：

```yaml
zlmediakit:
  host: "127.0.0.1"
  http_port: 80
  secret: "your-secret-key"
```

### 3. 启动服务

```bash
go run cmd/server/main.go -config configs/config.yaml
```

## 配置说明

```yaml
server:
  id: "34020000002000000001"   # 服务器 SIP ID
  domain: "3402000000"          # SIP 域
  host: "0.0.0.0"               # 监听地址
  sip_port: 5060                # SIP 端口
  http_port: 8080               # HTTP API 端口
  password: "12345678"          # 设备注册密码

zlmediakit:
  host: "127.0.0.1"             # ZLMediaKit 地址
  http_port: 80                 # ZLMediaKit HTTP API 端口
  secret: "your-secret"         # ZLMediaKit API Secret

device:
  heartbeat_timeout: 60         # 心跳超时时间（秒）
  register_expire: 3600         # 注册有效期（秒）
```

## HTTP API

### 设备管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/devices | 获取设备列表 |
| GET | /api/v1/devices/:device_id | 获取设备详情 |
| GET | /api/v1/devices/:device_id/channels | 获取设备通道 |
| POST | /api/v1/devices/:device_id/catalog | 查询设备目录 |

### 实时视频

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/stream/play | 开始实时预览 |
| DELETE | /api/v1/stream/play/:device_id/:channel_id | 停止预览 |
| GET | /api/v1/streams | 获取流列表 |
| GET | /api/v1/streams/:stream_id | 获取流详情 |

### 录像回放

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/records/query | 查询录像文件 |
| GET | /api/v1/records/:channel_id | 获取录像列表 |
| POST | /api/v1/stream/playback | 开始录像回放 |
| DELETE | /api/v1/stream/playback/:device_id/:channel_id | 停止回放 |

### 云台控制

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/ptz/control | 云台控制 |

### 级联管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/cascade/platforms | 获取平台列表 |
| POST | /api/v1/cascade/platforms | 添加上级平台 |
| DELETE | /api/v1/cascade/platforms/:platform_id | 删除平台 |
| POST | /api/v1/cascade/platforms/:platform_id/register | 注册到上级平台 |
| POST | /api/v1/cascade/platforms/:platform_id/catalog | 推送目录到上级 |

## API 调用示例

### 开始实时预览

```bash
curl -X POST http://localhost:8080/api/v1/stream/play \
  -H "Content-Type: application/json" \
  -d '{"device_id":"34020000001320000001","channel_id":"34020000001320000001"}'
```

响应：

```json
{
  "code": 0,
  "data": {
    "stream_id": "34020000001320000001_34020000001320000001",
    "ssrc": "1234567890",
    "rtp_port": 10000,
    "rtsp_url": "rtsp://127.0.0.1:554/rtp/...",
    "flv_url": "http://127.0.0.1/rtp/....live.flv",
    "hls_url": "http://127.0.0.1/rtp/.../hls.m3u8",
    "webrtc_url": "http://127.0.0.1/index/api/webrtc?..."
  }
}
```

### 查询录像

```bash
curl -X POST http://localhost:8080/api/v1/records/query \
  -H "Content-Type: application/json" \
  -d '{
    "device_id":"34020000001320000001",
    "channel_id":"34020000001320000001",
    "start_time":"2024-01-01T00:00:00",
    "end_time":"2024-01-02T00:00:00"
  }'
```

### 开始录像回放

```bash
curl -X POST http://localhost:8080/api/v1/stream/playback \
  -H "Content-Type: application/json" \
  -d '{
    "device_id":"34020000001320000001",
    "channel_id":"34020000001320000001",
    "start_time":"2024-01-01T10:00:00",
    "end_time":"2024-01-01T10:30:00"
  }'
```

### 云台控制

```bash
curl -X POST http://localhost:8080/api/v1/ptz/control \
  -H "Content-Type: application/json" \
  -d '{"device_id":"34020000001320000001","channel_id":"34020000001320000001","action":"up","speed":5}'
```

支持的动作：
- `stop` - 停止
- `up` - 上
- `down` - 下
- `left` - 左
- `right` - 右
- `zoom_in` - 放大
- `zoom_out` - 缩小

### 添加上级平台

```bash
curl -X POST http://localhost:8080/api/v1/cascade/platforms \
  -H "Content-Type: application/json" \
  -d '{
    "id":"platform001",
    "name":"上级平台",
    "server_id":"34020000002000000002",
    "server_domain":"3402000000",
    "server_ip":"192.168.1.100",
    "server_port":5060,
    "username":"34020000002000000001",
    "password":"12345678",
    "expires":3600
  }'
```

### 注册到上级平台

```bash
curl -X POST http://localhost:8080/api/v1/cascade/platforms/platform001/register
```

### 推送目录到上级

```bash
curl -X POST http://localhost:8080/api/v1/cascade/platforms/platform001/catalog \
  -H "Content-Type: application/json" \
  -d '{
    "channels":[{"device_id":"34020000001320000001","name":"摄像头1","status":"ONLINE"}]
  }'
```

## SIP 消息格式

参考 wvp-pro 实现，支持完整的 GB28181 消息：

### 注册（REGISTER）
- 支持 Digest MD5 鉴权
- 支持 Expires 过期处理

### 心跳（Keepalive）
- 自动保活机制
- 超时自动离线

### 目录查询（Catalog）
- 支持分页查询
- 自动解析通道信息

### 录像查询（RecordInfo）
- 支持时间范围查询
- 自动解析录像文件信息

### 实时预览（INVITE）
- SDP 协商
- SSRC 管理
- RTP 端口分配

### 录像回放（INVITE Playback）
- 支持 startTime/endTime 参数
- 历史视频流回放

## 开发说明

- SIP 消息解析采用字符串处理，参考 wvp-pro 实现
- Digest MD5 鉴权符合 RFC 2617 规范
- PTZ 控制命令符合 GB28181 协议规范
- 设备心跳超时自动标记为离线
- 级联支持上下级平台对接

## License

MIT
