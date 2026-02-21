#include "video_platform_client.h"
#include <iostream>

int main() {
    std::cout << "=== 视频平台C++客户端应用 ===" << std::endl;
    std::cout << "使用方法: " << std::endl;
    std::cout << "  1. 运行视频平台服务 (默认端口: 50051)" << std::endl;
    std::cout << "  2. 确保服务正常运行" << std::endl;
    std::cout << "  3. 运行此客户端应用" << std::endl;
    std::cout << "\n按Enter键继续..." << std::endl;
    std::cin.get();

    try {
        // 创建视频平台客户端
        VideoPlatformClient client("localhost:50051");

        std::cout << "\n连接视频平台服务成功!" << std::endl;

        // 示例操作
        std::cout << "\n执行示例操作..." << std::endl;

        // 1. 发现设备
        std::cout << "\n1. 发现设备..." << std::endl;
        auto devices = client.DiscoverDevices(3);
        std::cout << "发现设备数量: " << devices.size() << std::endl;

        // 2. 获取设备列表
        std::cout << "\n2. 获取设备列表..." << std::endl;
        devices = client.GetDevices();
        std::cout << "设备列表: " << devices.size() << " 个设备" << std::endl;
        for (size_t i = 0; i < devices.size(); i++) {
            const auto& device = devices[i];
            std::cout << "  " << i+1 << ". " << device.name 
                      << " (IP: " << device.ip 
                      << ", 状态: " << (device.status == DeviceStatus::ONLINE ? "在线" : "离线") 
                      << ")" << std::endl;
        }

        // 3. 如果有设备，启动第一个设备的流
        if (!devices.empty()) {
            const auto& device = devices[0];
            std::cout << "\n3. 启动设备流..." << std::endl;
            std::cout << "选择设备: " << device.name << " (ID: " << device.id << ")" << std::endl;

            auto stream = client.StartStream(device.id, "1", "RTSP");
            std::cout << "启动流成功: " << stream.name << " (ID: " << stream.id << ")" << std::endl;
            std::cout << "流地址: " << stream.url << std::endl;

            // 4. 获取流列表
            std::cout << "\n4. 获取流列表..." << std::endl;
            auto streams = client.GetStreams();
            std::cout << "流列表: " << streams.size() << " 个流" << std::endl;
            for (const auto& s : streams) {
                std::cout << "  - " << s.name << " (状态: " 
                          << (s.status == StreamStatus::PLAYING ? "播放中" : "停止") 
                          << ")" << std::endl;
            }

            // 5. 创建会话
            std::cout << "\n5. 创建会话..." << std::endl;
            auto session = client.CreateSession(stream.id, "test-client");
            std::cout << "创建会话成功: " << session.id << std::endl;
            std::cout << "会话状态: " << (session.status == SessionStatus::ACTIVE ? "活跃" : "关闭") << std::endl;

            // 6. 关闭会话
            std::cout << "\n6. 关闭会话..." << std::endl;
            bool closed = client.CloseSession(session.id);
            std::cout << "关闭会话: " << (closed ? "成功" : "失败") << std::endl;

            // 7. 停止流
            std::cout << "\n7. 停止流..." << std::endl;
            bool stopped = client.StopStream(stream.id);
            std::cout << "停止流: " << (stopped ? "成功" : "失败") << std::endl;
        }

        std::cout << "\n操作完成!" << std::endl;

    } catch (const std::exception& e) {
        std::cerr << "错误: " << e.what() << std::endl;
        return 1;
    }

    std::cout << "\n按Enter键退出..." << std::endl;
    std::cin.get();

    return 0;
}
