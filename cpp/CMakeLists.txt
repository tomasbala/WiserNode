cmake_minimum_required(VERSION 3.14)
project(video_platform_client CXX)

set(CMAKE_CXX_STANDARD 11)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 查找依赖
find_package(Protobuf REQUIRED)
find_package(gRPC REQUIRED)

# 设置源文件
set(SOURCE_FILES
    src/client.cpp
    src/video_platform.pb.cc
    src/video_platform.grpc.pb.cc
)

# 设置头文件目录
include_directories(
    include
    ${Protobuf_INCLUDE_DIRS}
    ${gRPC_INCLUDE_DIRS}
)

# 添加动态库
add_library(video_platform_client SHARED ${SOURCE_FILES})

# 链接库
target_link_libraries(video_platform_client
    ${Protobuf_LIBRARIES}
    ${gRPC_LIBRARIES}
    gRPC::grpc++
)

# 添加可执行文件
add_executable(video_platform_client_app src/app.cpp)
target_link_libraries(video_platform_client_app
    video_platform_client
)

# 添加示例
add_executable(example src/example.cpp)
target_link_libraries(example
    video_platform_client
)

# 安装配置
install(TARGETS video_platform_client video_platform_client_app example
    RUNTIME DESTINATION bin
    LIBRARY DESTINATION lib
    ARCHIVE DESTINATION lib
)

# 安装头文件
install(DIRECTORY include/
    DESTINATION include/video_platform
)
