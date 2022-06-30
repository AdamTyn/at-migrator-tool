package collector

import (
	"at-migrator-tool/internal/conf"
	"at-migrator-tool/internal/pkg"
	"log"
)

const FsRobotCollectorName = "FsRobot"

/**
 * FsRobotCollector
 * @author ctc
 * @description 采集器-飞书机器人
 */
type FsRobotCollector struct {
	closed bool
	tunnel chan interface{}
	size   int // 采集器最大缓存数
	logger *log.Logger
	Conf   *conf.Webhook
}

func NewFsRobotCollector(size int, c *conf.Webhook, logger *log.Logger) *FsRobotCollector {
	return &FsRobotCollector{
		tunnel: make(chan interface{}, size+1),
		logger: logger,
		size:   size,
		Conf:   c,
	}
}

func (c FsRobotCollector) Name() string {
	return FsRobotCollectorName
}

func (c *FsRobotCollector) Listen() {
	for {
		select {
		case data := <-c.tunnel:
			reply, err := pkg.JsonPost(c.Conf.FsRobot, data)
			if err != nil {
				c.logger.Printf("[Exception] FsRobotCollector->Listen: %s\n", err.Error())
			} else {
				c.logger.Printf("[Info] FsRobotCollector->Listen,reply=%s\n", string(reply))
			}
		default:
			if c.closed {
				return
			}
		}
	}
}

func (c *FsRobotCollector) Put(in interface{}) error {
	if c.closed {
		return pkg.ErrCollectorClosed
	}
	c.tunnel <- in
	return nil
}

func (c *FsRobotCollector) Close() {
	if !c.closed {
		close(c.tunnel)
	}
	c.closed = true
}
