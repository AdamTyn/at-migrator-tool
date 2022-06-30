package contract

/**
 * Process
 * @author ctc
 * @description 实现本契约都可以在app容器中运行
 */
type Process interface {
	Start()
	Shutdown()
	Common
}
