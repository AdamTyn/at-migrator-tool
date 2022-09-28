package collector

import (
	"at-migrator-tool/internal/conf"
	"at-migrator-tool/internal/pkg"
	"at-migrator-tool/internal/pkg/log"
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
	Conf   *conf.Webhook
}

func NewFsRobotCollector(size int, c *conf.Webhook) *FsRobotCollector {
	return &FsRobotCollector{
		tunnel: make(chan interface{}, size+1),
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
				log.ExceptionF("FsRobotCollector->Listen: %s", err.Error())
			} else {
				log.InfoF("FsRobotCollector->Listen,reply=%s", string(reply))
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
	var old bool
	old, c.closed = c.closed, true
	if !old {
		close(c.tunnel)
	}
}
