package collector

const DefaultCollectorName = "Default"

type DefaultCollector struct {
}

func (c DefaultCollector) Name() string          { return DefaultCollectorName }
func (c DefaultCollector) Listen()               {}
func (c DefaultCollector) Close()                {}
func (c DefaultCollector) Put(interface{}) error { return nil }
