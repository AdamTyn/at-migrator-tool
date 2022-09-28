package log

import "testing"

func TestDebug(t *testing.T) {
	Debug("string", 1)
}

func TestDebugF(t *testing.T) {
	DebugF("string=%s", "xxx")
}

func TestInfo(t *testing.T) {
	Info("string", 1)
}

func TestInfoF(t *testing.T) {
	InfoF("string=%s", "xxx")
}

func TestNotice(t *testing.T) {
	Notice("string", 1)
}

func TestNoticeF(t *testing.T) {
	NoticeF("string=%s", "xxx")
}

func TestWarn(t *testing.T) {
	Warn("string", 1)
}

func TestWarnF(t *testing.T) {
	WarnF("string=%s", "xxx")
}

func TestException(t *testing.T) {
	Exception("string", 1)
}

func TestExceptionF(t *testing.T) {
	ExceptionF("string=%s", "xxx")
}

func TestError(t *testing.T) {
	Error("string", 1)
}

func TestErrorF(t *testing.T) {
	ErrorF("string=%s", "xxx")
}
