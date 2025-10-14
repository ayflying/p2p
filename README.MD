# P2P 通信系统

一个基于 Go 语言和 libp2p 构建的对等网络通信系统，提供可靠的分布式节点间通信功能。

## 功能特点

- 基于 libp2p 的去中心化 P2P 网络通信
- GoFrame 框架支持的高效 Web 服务
- Redis 缓存集成
- 模块化架构设计
- 命令行工具支持

## 技术栈

- **Go 语言**: 1.24.0
- **框架**: GoFrame v2.9.3
- **P2P 库**: libp2p v0.43.0
- **存储**: Redis
- **多地址格式**: multiaddr v0.16.0

## 安装要求

- Go 1.24.0 或更高版本
- Redis 服务器
- Git

## 快速开始

### 克隆项目

```bash
git clone https://github.com/ayflying/p2p.git
cd p2p
```

### 安装依赖

```bash
go mod tidy
```

### 构建项目

```bash
go build -o p2p
```

### 运行项目

```bash
./p2p
```

## 目录结构

```
├── api/             # API 定义
│   └── p2p/         # P2P 相关 API
├── internal/        # 内部代码
│   ├── cmd/         # 命令行工具
│   ├── consts/      # 常量定义
│   ├── controller/  # 控制器
│   ├── logic/       # 业务逻辑
│   ├── model/       # 数据模型
│   ├── packed/      # 打包文件
│   └── service/     # 服务层
├── main.go          # 程序入口
├── go.mod           # Go 模块定义
├── go.sum           # 依赖锁定文件
├── Makefile         # 构建脚本
└── README.MD        # 项目文档
```

## 核心模块

### P2P 通信

提供基于 libp2p 的节点发现、连接和通信功能，支持分布式网络中的节点间数据传输。

### Web 服务

基于 GoFrame 框架构建的 Web 服务，提供 RESTful API 接口。

### 缓存系统

集成 Redis 缓存，提高系统性能和数据访问效率。

## 命令行工具

项目提供了命令行工具，支持节点管理、网络状态查询等功能。

```bash
./p2p --help  # 查看帮助信息
```

## 开发指南

### 代码规范

- 遵循 Go 语言标准编码规范
- 保持代码简洁、可读
- 添加适当的注释

### 提交代码

1. Fork 项目仓库
2. 创建功能分支
3. 提交代码变更
4. 创建 Pull Request

## 许可证

[MIT License](https://opensource.org/licenses/MIT)

## 联系方式

如有问题或建议，请通过以下方式联系我们：
- GitHub Issues: https://github.com/ayflying/p2p/issues

---

*最后更新时间: 2024-07-15*