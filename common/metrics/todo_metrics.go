package metrics

type TodoMetrics struct {
}

func (t TodoMetrics) Inc(_ string, _ int) {

}

func (t TodoMetrics) IncCounter(name string, tags ...string) {
	//TODO implement me
	panic("implement me")
}

func (t TodoMetrics) RecordTimer(name string, duration float64, tags ...string) {
	//TODO implement me
	panic("implement me")
}

func (t TodoMetrics) RecordHistogram(name string, value float64, tags ...string) {
	//TODO implement me
	panic("implement me")
}
