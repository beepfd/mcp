package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cen-ngc5139/beepf-mcp/pkg/tools/metrics"
	"github.com/cen-ngc5139/beepf-mcp/pkg/tools/observability"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

func main() {
	// 创建 MCP 服务器
	mcpServer := server.NewMCPServer(
		"BeePF MCP Server",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
		server.WithRecovery(),
	)

	// 添加观测性相关接口工具
	// v1.GET("/observability/topo", topoService.Topo())
	topoTool := mcp.NewTool("topo",
		mcp.WithDescription("获取系统拓扑信息，展示 eBPF 程序和网络拓扑"),
	)
	mcpServer.AddTool(topoTool, topoHandler)

	// v1.GET("/observability/topo/prog", topoService.Prog())
	progTool := mcp.NewTool("prog",
		mcp.WithDescription("获取所有 eBPF 程序列表"),
	)
	mcpServer.AddTool(progTool, progHandler)

	// v1.GET("/observability/topo/prog/:progId", topoService.ProgDetail())
	progDetailTool := mcp.NewTool("prog_detail",
		mcp.WithDescription("获取特定 eBPF 程序的详细信息"),
		mcp.WithString("progId",
			mcp.Required(),
			mcp.Description("eBPF 程序ID"),
		),
	)
	mcpServer.AddTool(progDetailTool, progDetailHandler)

	// v1.GET("/observability/topo/prog/:progId/dump", topoService.ProgDump())
	progDumpTool := mcp.NewTool("prog_dump",
		mcp.WithDescription("获取特定 eBPF 程序的字节码转储"),
		mcp.WithString("progId",
			mcp.Required(),
			mcp.Description("eBPF 程序ID"),
		),
		mcp.WithString("format",
			mcp.Description("转储格式，可选 'xlated' 或 'jited'"),
			mcp.Enum("xlated", "jited"),
		),
	)
	mcpServer.AddTool(progDumpTool, progDumpHandler)

	// v1.GET("/observability/node/metrics", nodeMetricsService.GetMetrics())
	nodeMetricsTool := mcp.NewTool("node_metrics",
		mcp.WithDescription("获取节点性能指标数据"),
	)
	mcpServer.AddTool(nodeMetricsTool, nodeMetricsHandler)

	// 处理 Ctrl+C 信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// 监听 Ctrl+C 信号
	go func() {
		<-sigChan
		fmt.Println("\n收到终止信号，正在关闭服务...")
		os.Exit(0)
	}()

	// 定义SSE服务的端口和路径
	port := 8080
	endpoint := "/sse"

	// 启动服务器（SSE方式）
	fmt.Printf("启动 BeePF MCP 服务器 (SSE模式)...\n服务地址: http://0.0.0.0:%d%s\n", port, endpoint)
	sseServer := server.NewSSEServer(mcpServer, server.WithBaseURL("http://192.168.200.200:8080"))
	log.Printf("SSE server listening on :8080")
	if err := sseServer.Start(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
		os.Exit(1)
	}

	// 保留原有的Stdio服务能力，作为备选方案
	// 如果想使用Stdio模式，取消下面的注释，并注释掉上面的ServeSSE部分
	/*
		fmt.Println("启动 BeePF MCP 服务器 (Stdio模式)...")
		if err := server.ServeStdio(mcpServer); err != nil {
			fmt.Printf("服务器错误: %v\n", err)
		}
	*/
}

// 处理拓扑信息请求
func topoHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	topoService := observability.NewTopo()

	topology, err := topoService.GetTopo()
	if err != nil {
		return nil, fmt.Errorf("获取拓扑信息失败: %v", err)
	}

	// 将拓扑数据转为 JSON
	jsonData, err := json.MarshalIndent(topology, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("序列化拓扑数据失败: %v", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// 处理节点程序列表请求
func progHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	topoService := observability.NewTopo()

	progs, err := topoService.ListProgs()
	if err != nil {
		return nil, fmt.Errorf("获取程序列表失败: %v", err)
	}

	// 将程序列表转为 JSON
	jsonData, err := json.MarshalIndent(progs, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("序列化程序列表数据失败: %v", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// 处理节点程序详情请求
func progDetailHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	progID, ok := request.Params.Arguments["progId"].(string)
	if !ok {
		return nil, fmt.Errorf("progId 必须是字符串类型")
	}

	topoService := observability.NewTopo()

	detail, err := topoService.GetProgDetail(progID)
	if err != nil {
		return nil, fmt.Errorf("获取程序详情失败: %v", err)
	}

	// 将程序详情转为 JSON
	jsonData, err := json.MarshalIndent(detail, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("序列化程序详情数据失败: %v", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// 处理节点程序数据转储请求
func progDumpHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	progID, ok := request.Params.Arguments["progId"].(string)
	if !ok {
		return nil, fmt.Errorf("progId 必须是字符串类型")
	}

	format, _ := request.Params.Arguments["format"].(string)
	if format == "" {
		format = "xlated" // 默认使用 xlated 格式
	}

	topoService := observability.NewTopo()

	var dump []byte
	var err error

	if format == "xlated" {
		dump, err = topoService.DumpXlated(progID)
	} else if format == "jited" {
		dump, err = topoService.DumpJited(progID)
	} else {
		return nil, fmt.Errorf("不支持的转储格式: %s", format)
	}

	if err != nil {
		return nil, fmt.Errorf("获取程序转储失败: %v", err)
	}

	// dump 是字节数组，转换为可读字符串
	dumpStr := fmt.Sprintf("eBPF 程序 %s 的 %s 转储:\n%s", progID, format, string(dump))

	return mcp.NewToolResultText(dumpStr), nil
}

// 处理节点指标数据请求
func nodeMetricsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 使用metrics包中的NodeMetricsCollector获取节点指标数据
	nodeCollector, err := metrics.NewNodeMetricsCollector(5*time.Second, zap.NewNop())

	// 获取节点指标数据
	nodeMetrics, err := nodeCollector.GetMetrics()
	if err != nil {
		return nil, fmt.Errorf("获取节点指标数据失败: %v", err)
	}

	// 将节点指标数据转为JSON
	jsonData, err := json.MarshalIndent(nodeMetrics, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("序列化节点指标数据失败: %v", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}
