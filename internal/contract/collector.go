package contract

/**
 * Collector
 * @author ctc
 * @description 实现本契约都可以作为采集器
 */
type Collector interface {
	Listen()
	Close()
	Put(interface{}) error
	Common
}
