#include "video_platform_client.h"
#include "video_platform.grpc.pb.h"
#include "video_platform.pb.h"

#include <grpcpp/grpcpp.h>
#include <iostream>

// 构造函数
VideoPlatformClient::VideoPlatformClient(const std::string& server_address) {
    // 创建gRPC通道
    auto channel = grpc::CreateChannel(server_address, grpc::InsecureChannelCredentials());
    // 创建客户端存根
    stub_ = videoplatform::VideoPlatform::NewStub(channel);
}

// 析构函数
VideoPlatformClient::~VideoPlatformClient() {
}

// 设备管理
std::vector<DeviceInfo> VideoPlatformClient::DiscoverDevices(int timeout) {
    grpc::ClientContext context;
    videoplatform::DiscoverDevicesRequest request;
    videoplatform::DiscoverDevicesResponse response;

    request.set_timeout(timeout);

    grpc::Status status = stub_->DiscoverDevices(&context, request, &response);
    if (!status.ok()) {
        std::cerr << "DiscoverDevices failed: " << status.error_message() << std::endl;
        return {};
    }

    std::vector<DeviceInfo> devices;
    for (const auto& device : response.devices()) {
        DeviceInfo info;
        info.id = device.id();
        info.name = device.name();
        info.type = static_cast<DeviceType>(device.type());
        info.ip = device.ip();
        info.port = device.port();
        info.username = device.username();
        info.password = device.password();
        info.status = static_cast<DeviceStatus>(device.status());
        info.manufacturer = device.manufacturer();
        info.model = device.model();
        info.firmware = device.firmware();
        info.serial_number = device.serial_number();
        devices.push_back(info);
    }

    return devices;
}

DeviceInfo VideoPlatformClient::AddDevice(const std::string& name, DeviceType type, const std::string& ip, int port, 
                                         const std::string& username, const std::string& password) {
    grpc::ClientContext context;
    videoplatform::AddDeviceRequest request;
    videoplatform::AddDeviceResponse response;

    request.set_name(name);
    request.set_type(static_cast<videoplatform::DeviceType>(type));
    request.set_ip(ip);
    request.set_port(port);
    request.set_username(username);
    request.set_password(password);

    grpc::Status status = stub_->AddDevice(&context, request, &response);
    if (!status.ok()) {
        std::cerr << "AddDevice failed: " << status.error_message() << std::endl;
        return DeviceInfo{};
    }

    const auto& device = response.device();
    DeviceInfo info;
    info.id = device.id();
    info.name = device.name();
    info.type = static_cast<DeviceType>(device.type());
    info.ip = device.ip();
    info.port = device.port();
    info.username = device.username();
    info.password = device.password();
    info.status = static_cast<DeviceStatus>(device.status());
    info.manufacturer = device.manufacturer();
    info.model = device.model();
    info.firmware = device.firmware();
    info.serial_number = device.serial_number();

    return info;
}

DeviceInfo VideoPlatformClient::GetDevice(const std::string& device_id) {
    grpc::ClientContext context;
    videoplatform::GetDeviceRequest request;
    videoplatform::GetDeviceResponse response;

    request.set_device_id(device_id);

    grpc::Status status = stub_->GetDevice(&context, request, &response);
    if (!status.ok()) {
        std::cerr << "GetDevice failed: " << status.error_message() << std::endl;
        return DeviceInfo{};
    }

    const auto& device = response.device();
    DeviceInfo info;
    info.id = device.id();
    info.name = device.name();
    info.type = static_cast<DeviceType>(device.type());
    info.ip = device.ip();
    info.port = device.port();
    info.username = device.username();
    info.password = device.password();
    info.status = static_cast<DeviceStatus>(device.status());
    info.manufacturer = device.manufacturer();
    info.model = device.model();
    info.firmware = device.firmware();
    info.serial_number = device.serial_number();

    return info;
}

std::vector<DeviceInfo> VideoPlatformClient::GetDevices() {
    grpc::ClientContext context;
    videoplatform::GetDevicesRequest request;
    videoplatform::GetDevicesResponse response;

    grpc::Status status = stub_->GetDevices(&context, request, &response);
    if (!status.ok()) {
        std::cerr << "GetDevices failed: " << status.error_message() << std::endl;
        return {};
    }

    std::vector<DeviceInfo> devices;
    for (const auto& device : response.devices()) {
        DeviceInfo info;
        info.id = device.id();
        info.name = device.name();
        info.type = static_cast<DeviceType>(device.type());
        info.ip = device.ip();
        info.port = device.port();
        info.username = device.username();
        info.password = device.password();
        info.status = static_cast<DeviceStatus>(device.status());
        info.manufacturer = device.manufacturer();
        info.model = device.model();
        info.firmware = device.firmware();
        info.serial_number = device.serial_number();
        devices.push_back(info);
    }

    return devices;
}

bool VideoPlatformClient::DeleteDevice(const std::string& device_id) {
    grpc::ClientContext context;
    videoplatform::DeleteDeviceRequest request;
    videoplatform::DeleteDeviceResponse response;

    request.set_device_id(device_id);

    grpc::Status status = stub_->DeleteDevice(&context, request, &response);
    if (!status.ok()) {
        std::cerr << "DeleteDevice failed: " << status.error_message() << std::endl;
        return false;
    }

    return response.success();
}

DeviceStatus VideoPlatformClient::GetDeviceStatus(const std::string& device_id) {
    grpc::ClientContext context;
    videoplatform::GetDeviceStatusRequest request;
    videoplatform::GetDeviceStatusResponse response;

    request.set_device_id(device_id);

    grpc::Status status = stub_->GetDeviceStatus(&context, request, &response);
    if (!status.ok()) {
        std::cerr << "GetDeviceStatus failed: " << status.error_message() << std::endl;
        return DeviceStatus::UNKNOWN;
    }

    return static_cast<DeviceStatus>(response.status());
}

// 流媒体管理
StreamInfo VideoPlatformClient::StartStream(const std::string& device_id, const std::string& channel_id, 
                                           const std::string& protocol) {
    grpc::ClientContext context;
    videoplatform::StartStreamRequest request;
    videoplatform::StartStreamResponse response;

    request.set_device_id(device_id);
    request.set_channel_id(channel_id);
    request.set_protocol(protocol);

    grpc::Status status = stub_->StartStream(&context, request, &response);
    if (!status.ok()) {
        std::cerr << "StartStream failed: " << status.error_message() << std::endl;
        return StreamInfo{};
    }

    const auto& stream = response.stream();
    StreamInfo info;
    info.id = stream.id();
    info.device_id = stream.device_id();
    info.channel_id = stream.channel_id();
    info.name = stream.name();
    info.url = stream.url();
    info.status = static_cast<StreamStatus>(stream.status());
    info.protocol = stream.protocol();
    info.codec = stream.codec();
    info.resolution = stream.resolution();
    info.bitrate = stream.bitrate();

    return info;
}

bool VideoPlatformClient::StopStream(const std::string& stream_id) {
    grpc::ClientContext context;
    videoplatform::StopStreamRequest request;
    videoplatform::StopStreamResponse response;

    request.set_stream_id(stream_id);

    grpc::Status status = stub_->StopStream(&context, request, &response);
    if (!status.ok()) {
        std::cerr << "StopStream failed: " << status.error_message() << std::endl;
        return false;
    }

    return response.success();
}

StreamInfo VideoPlatformClient::GetStream(const std::string& stream_id) {
    grpc::ClientContext context;
    videoplatform::GetStreamRequest request;
    videoplatform::GetStreamResponse response;

    request.set_stream_id(stream_id);

    grpc::Status status = stub_->GetStream(&context, request, &response);
    if (!status.ok()) {
        std::cerr << "GetStream failed: " << status.error_message() << std::endl;
        return StreamInfo{};
    }

    const auto& stream = response.stream();
    StreamInfo info;
    info.id = stream.id();
    info.device_id = stream.device_id();
    info.channel_id = stream.channel_id();
    info.name = stream.name();
    info.url = stream.url();
    info.status = static_cast<StreamStatus>(stream.status());
    info.protocol = stream.protocol();
    info.codec = stream.codec();
    info.resolution = stream.resolution();
    info.bitrate = stream.bitrate();

    return info;
}

std::vector<StreamInfo> VideoPlatformClient::GetStreams() {
    grpc::ClientContext context;
    videoplatform::GetStreamsRequest request;
    videoplatform::GetStreamsResponse response;

    grpc::Status status = stub_->GetStreams(&context, request, &response);
    if (!status.ok()) {
        std::cerr << "GetStreams failed: " << status.error_message() << std::endl;
        return {};
    }

    std::vector<StreamInfo> streams;
    for (const auto& stream : response.streams()) {
        StreamInfo info;
        info.id = stream.id();
        info.device_id = stream.device_id();
        info.channel_id = stream.channel_id();
        info.name = stream.name();
        info.url = stream.url();
        info.status = static_cast<StreamStatus>(stream.status());
        info.protocol = stream.protocol();
        info.codec = stream.codec();
        info.resolution = stream.resolution();
        info.bitrate = stream.bitrate();
        streams.push_back(info);
    }

    return streams;
}

std::string VideoPlatformClient::GetStreamUrl(const std::string& stream_id) {
    grpc::ClientContext context;
    videoplatform::GetStreamUrlRequest request;
    videoplatform::GetStreamUrlResponse response;

    request.set_stream_id(stream_id);

    grpc::Status status = stub_->GetStreamUrl(&context, request, &response);
    if (!status.ok()) {
        std::cerr << "GetStreamUrl failed: " << status.error_message() << std::endl;
        return "";
    }

    return response.url();
}

// 会话管理
SessionInfo VideoPlatformClient::CreateSession(const std::string& stream_id, const std::string& client_id) {
    grpc::ClientContext context;
    videoplatform::CreateSessionRequest request;
    videoplatform::CreateSessionResponse response;

    request.set_stream_id(stream_id);
    request.set_client_id(client_id);

    grpc::Status status = stub_->CreateSession(&context, request, &response);
    if (!status.ok()) {
        std::cerr << "CreateSession failed: " << status.error_message() << std::endl;
        return SessionInfo{};
    }

    const auto& session = response.session();
    SessionInfo info;
    info.id = session.id();
    info.stream_id = session.stream_id();
    info.client_id = session.client_id();
    info.status = static_cast<SessionStatus>(session.status());
    info.start_time = session.start_time();
    info.end_time = session.end_time();

    return info;
}

bool VideoPlatformClient::CloseSession(const std::string& session_id) {
    grpc::ClientContext context;
    videoplatform::CloseSessionRequest request;
    videoplatform::CloseSessionResponse response;

    request.set_session_id(session_id);

    grpc::Status status = stub_->CloseSession(&context, request, &response);
    if (!status.ok()) {
        std::cerr << "CloseSession failed: " << status.error_message() << std::endl;
        return false;
    }

    return response.success();
}

SessionInfo VideoPlatformClient::GetSession(const std::string& session_id) {
    grpc::ClientContext context;
    videoplatform::GetSessionRequest request;
    videoplatform::GetSessionResponse response;

    request.set_session_id(session_id);

    grpc::Status status = stub_->GetSession(&context, request, &response);
    if (!status.ok()) {
        std::cerr << "GetSession failed: " << status.error_message() << std::endl;
        return SessionInfo{};
    }

    const auto& session = response.session();
    SessionInfo info;
    info.id = session.id();
    info.stream_id = session.stream_id();
    info.client_id = session.client_id();
    info.status = static_cast<SessionStatus>(session.status());
    info.start_time = session.start_time();
    info.end_time = session.end_time();

    return info;
}

std::vector<SessionInfo> VideoPlatformClient::GetSessions() {
    grpc::ClientContext context;
    videoplatform::GetSessionsRequest request;
    videoplatform::GetSessionsResponse response;

    grpc::Status status = stub_->GetSessions(&context, request, &response);
    if (!status.ok()) {
        std::cerr << "GetSessions failed: " << status.error_message() << std::endl;
        return {};
    }

    std::vector<SessionInfo> sessions;
    for (const auto& session : response.sessions()) {
        SessionInfo info;
        info.id = session.id();
        info.stream_id = session.stream_id();
        info.client_id = session.client_id();
        info.status = static_cast<SessionStatus>(session.status());
        info.start_time = session.start_time();
        info.end_time = session.end_time();
        sessions.push_back(info);
    }

    return sessions;
}
