package metrics

import (
	"fmt"
	"time"

	"github.com/cen-ngc5139/BeePF/loader/lib/src/meta"
	"github.com/cen-ngc5139/BeePF/loader/lib/src/observability/topology"
	"go.uber.org/zap"
)

type NodeMetricsCollector struct {
	*topology.NodeMetricsCollector
}

// 默认5秒采集一次指标
const defaultInterval = 5 * time.Second

// NewNodeMetricsCollector 创建一个新的节点指标采集器，使用给定的时间间隔和日志记录器
func NewNodeMetricsCollector(interval time.Duration, logger *zap.Logger) (*NodeMetricsCollector, error) {
	collector, err := topology.NewNodeMetricsCollector(interval, logger)
	if err != nil {
		return nil, fmt.Errorf("fail to create node metrics collector: %v", err)
	}

	if err := collector.Start(); err != nil {
		return nil, fmt.Errorf("fail to start node metrics collector: %v", err)
	}

	return &NodeMetricsCollector{
		NodeMetricsCollector: collector,
	}, nil
}

// NewDefaultNodeMetricsCollector 创建一个使用默认参数的节点指标采集器
func NewDefaultNodeMetricsCollector() *NodeMetricsCollector {
	logger, _ := zap.NewProduction()
	collector, err := NewNodeMetricsCollector(defaultInterval, logger)
	if err != nil {
		logger.Error("failed to create node metrics collector with default params", zap.Error(err))
		return &NodeMetricsCollector{}
	}

	return collector
}

func (c *NodeMetricsCollector) GetMetrics() (map[uint32]*meta.ProgMetricsStats, error) {
	// 如果collector为nil，返回空map
	if c.NodeMetricsCollector == nil {
		return map[uint32]*meta.ProgMetricsStats{}, nil
	}

	curr, err := c.NodeMetricsCollector.GetMetrics()
	if err != nil {
		return nil, fmt.Errorf("fail to get node metrics: %v", err)
	}

	for id, s := range curr {
		if s.Stats.CPUTimePercent == 0 {
			delete(curr, id)
		}
	}

	return curr, nil
}
