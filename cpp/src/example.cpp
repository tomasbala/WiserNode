#include "video_platform_client.h"
#include <iostream>

int main() {
    // 创建视频平台客户端
    VideoPlatformClient client("localhost:50051");

    std::cout << "=== 视频平台C++客户端示例 ===" << std::endl;

    // 1. 发现设备
    std::cout << "\n1. 发现设备..." << std::endl;
    auto devices = client.DiscoverDevices(5);
    std::cout << "发现设备数量: " << devices.size() << std::endl;
    for (const auto& device : devices) {
        std::cout << "  - " << device.name << " (" << device.ip << ")" << std::endl;
    }

    // 2. 添加设备
    std::cout << "\n2. 添加设备..." << std::endl;
    auto newDevice = client.AddDevice(
        "测试摄像头",
        DeviceType::ONVIF,
        "192.168.1.100",
        80,
        "admin",
        "admin123"
    );
    std::cout << "添加设备成功: " << newDevice.name << " (ID: " << newDevice.id << ")" << std::endl;

    // 3. 获取设备列表
    std::cout << "\n3. 获取设备列表..." << std::endl;
    devices = client.GetDevices();
    std::cout << "设备列表: " << devices.size() << " 个设备" << std::endl;
    for (const auto& device : devices) {
        std::cout << "  - " << device.name << " (ID: " << device.id << ")" << std::endl;
    }

    // 4. 启动流
    std::cout << "\n4. 启动流..." << std::endl;
    auto stream = client.StartStream(newDevice.id, "1", "RTSP");
    std::cout << "启动流成功: " << stream.name << " (ID: " << stream.id << ")" << std::endl;
    std::cout << "流地址: " << stream.url << std::endl;

    // 5. 获取流列表
    std::cout << "\n5. 获取流列表..." << std::endl;
    auto streams = client.GetStreams();
    std::cout << "流列表: " << streams.size() << " 个流" << std::endl;
    for (const auto& s : streams) {
        std::cout << "  - " << s.name << " (ID: " << s.id << ")" << std::endl;
    }

    // 6. 创建会话
    std::cout << "\n6. 创建会话..." << std::endl;
    auto session = client.CreateSession(stream.id, "client-123");
    std::cout << "创建会话成功: " << session.id << std::endl;
    std::cout << "会话状态: " << (session.status == SessionStatus::ACTIVE ? "活跃" : "关闭") << std::endl;

    // 7. 获取会话列表
    std::cout << "\n7. 获取会话列表..." << std::endl;
    auto sessions = client.GetSessions();
    std::cout << "会话列表: " << sessions.size() << " 个会话" << std::endl;
    for (const auto& s : sessions) {
        std::cout << "  - " << s.id << " (状态: " << (s.status == SessionStatus::ACTIVE ? "活跃" : "关闭") << ")" << std::endl;
    }

    // 8. 关闭会话
    std::cout << "\n8. 关闭会话..." << std::endl;
    bool closed = client.CloseSession(session.id);
    std::cout << "关闭会话: " << (closed ? "成功" : "失败") << std::endl;

    // 9. 停止流
    std::cout << "\n9. 停止流..." << std::endl;
    bool stopped = client.StopStream(stream.id);
    std::cout << "停止流: " << (stopped ? "成功" : "失败") << std::endl;

    // 10. 删除设备
    std::cout << "\n10. 删除设备..." << std::endl;
    bool deleted = client.DeleteDevice(newDevice.id);
    std::cout << "删除设备: " << (deleted ? "成功" : "失败") << std::endl;

    std::cout << "\n=== 示例完成 ===" << std::endl;

    return 0;
}
