#ifndef VIDEO_PLATFORM_CLIENT_H
#define VIDEO_PLATFORM_CLIENT_H

#include <string>
#include <vector>
#include <memory>

// 前向声明
namespace videoplatform {
class VideoPlatform;
class DiscoverDevicesRequest;
class DiscoverDevicesResponse;
class AddDeviceRequest;
class AddDeviceResponse;
class GetDeviceRequest;
class GetDeviceResponse;
class GetDevicesRequest;
class GetDevicesResponse;
class DeleteDeviceRequest;
class DeleteDeviceResponse;
class GetDeviceStatusRequest;
class GetDeviceStatusResponse;
class StartStreamRequest;
class StartStreamResponse;
class StopStreamRequest;
class StopStreamResponse;
class GetStreamRequest;
class GetStreamResponse;
class GetStreamsRequest;
class GetStreamsResponse;
class GetStreamUrlRequest;
class GetStreamUrlResponse;
class CreateSessionRequest;
class CreateSessionResponse;
class CloseSessionRequest;
class CloseSessionResponse;
class GetSessionRequest;
class GetSessionResponse;
class GetSessionsRequest;
class GetSessionsResponse;
}

// 设备类型
enum class DeviceType {
    UNKNOWN,
    ONVIF,
    GB28181,
    RTSP
};

// 设备状态
enum class DeviceStatus {
    UNKNOWN,
    ONLINE,
    OFFLINE,
    ERROR
};

// 流状态
enum class StreamStatus {
    UNKNOWN,
    PLAYING,
    STOPPED,
    ERROR
};

// 会话状态
enum class SessionStatus {
    UNKNOWN,
    ACTIVE,
    CLOSED,
    ERROR
};

// 设备信息结构
typedef struct DeviceInfo {
    std::string id;
    std::string name;
    DeviceType type;
    std::string ip;
    int port;
    std::string username;
    std::string password;
    DeviceStatus status;
    std::string manufacturer;
    std::string model;
    std::string firmware;
    std::string serial_number;
} DeviceInfo;

// 流信息结构
typedef struct StreamInfo {
    std::string id;
    std::string device_id;
    std::string channel_id;
    std::string name;
    std::string url;
    StreamStatus status;
    std::string protocol;
    std::string codec;
    std::string resolution;
    int bitrate;
} StreamInfo;

// 会话信息结构
typedef struct SessionInfo {
    std::string id;
    std::string stream_id;
    std::string client_id;
    SessionStatus status;
    long long start_time;
    long long end_time;
} SessionInfo;

// 视频平台客户端类
class VideoPlatformClient {
public:
    // 构造函数
    VideoPlatformClient(const std::string& server_address);
    
    // 析构函数
    ~VideoPlatformClient();
    
    // 设备管理
    std::vector<DeviceInfo> DiscoverDevices(int timeout);
    DeviceInfo AddDevice(const std::string& name, DeviceType type, const std::string& ip, int port, 
                        const std::string& username = "", const std::string& password = "");
    DeviceInfo GetDevice(const std::string& device_id);
    std::vector<DeviceInfo> GetDevices();
    bool DeleteDevice(const std::string& device_id);
    DeviceStatus GetDeviceStatus(const std::string& device_id);
    
    // 流媒体管理
    StreamInfo StartStream(const std::string& device_id, const std::string& channel_id = "", 
                          const std::string& protocol = "");
    bool StopStream(const std::string& stream_id);
    StreamInfo GetStream(const std::string& stream_id);
    std::vector<StreamInfo> GetStreams();
    std::string GetStreamUrl(const std::string& stream_id);
    
    // 会话管理
    SessionInfo CreateSession(const std::string& stream_id, const std::string& client_id);
    bool CloseSession(const std::string& session_id);
    SessionInfo GetSession(const std::string& session_id);
    std::vector<SessionInfo> GetSessions();
    
private:
    // gRPC客户端
    std::unique_ptr<videoplatform::VideoPlatform::Stub> stub_;
};

#endif // VIDEO_PLATFORM_CLIENT_H
