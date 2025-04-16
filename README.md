# BeePF MCP Server

基于 Model Context Protocol (MCP) 实现的 BeePF 可观测性接口集成工具。

## 简介

本项目使用 [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) 框架将 BeePF 项目中的可观测性相关接口转换为 MCP 工具，使大语言模型能够通过 MCP 协议调用这些接口获取系统数据。

## 功能特性

本工具提供了以下 MCP 工具：

- **topo**: 获取系统拓扑信息
- **prog**: 获取节点程序列表
- **prog_detail**: 获取特定节点程序详情（需要 progId 参数）
- **prog_dump**: 获取特定节点程序数据转储（需要 progId 参数）
- **node_metrics**: 获取节点指标数据

## 安装

```bash
# 克隆项目
git clone https://github.com/cen-ngc5139/beepf-mcp.git
cd beepf-mcp

# 安装依赖
go mod tidy

# 构建项目
go build -o beepf-mcp .
```

## 使用方法

### 直接启动

```bash
./beepf-mcp
```

服务器将以标准输入/输出方式启动，可以与支持 MCP 的客户端（例如 [MCPHost](https://github.com/mark3labs/mcphost)）进行交互。

### 与 MCPHost 一起使用

1. 安装 MCPHost

```bash
go install github.com/mark3labs/mcphost@latest
```

2. 配置 MCPHost 

在 ~/.mcp.json 中添加以下配置：

```json
{
  "mcpServers": {
    "beepf": {
      "command": "/path/to/beepf-mcp",
      "args": []
    }
  }
}
```

3. 启动 MCPHost

```bash
mcphost
```

4. 在 MCPHost 中使用命令

```
/tools
```

列出可用工具，然后可以通过交互方式使用这些工具。

## 示例交互

```
> 请获取系统拓扑信息

使用 topo 工具获取系统拓扑信息...
系统拓扑信息：节点数 5，连接数 8

> 列出所有节点程序

使用 prog 工具获取节点程序列表...
节点程序列表：[prog_001, prog_002, prog_003]

> 获取 prog_001 的详细信息

使用 prog_detail 工具获取节点 prog_001 的详情...
节点程序 prog_001 详情：类型=XDP，运行时间=120s
```

## 联系与贡献

欢迎提交 Issue 和 Pull Request 来改进此项目。
