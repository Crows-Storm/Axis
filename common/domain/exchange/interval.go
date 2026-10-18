package exchange

type Interval string

const (
	OneSecond      Interval = "1s"
	FiveMinutes    Interval = "5m"
	FifteenMinutes Interval = "15m"
	ThirtyMinutes  Interval = "30m"
	OneHour        Interval = "1h"
	OneDay         Interval = "1d"
	OneWeek        Interval = "1w"
	OneMonth       Interval = "1M"
	OneYear        Interval = "1Y"
)

func (i Interval) Valid() bool {
	switch i {
	case OneSecond, FiveMinutes, FifteenMinutes, ThirtyMinutes,
		OneHour, OneDay, OneWeek, OneMonth, OneYear:
		return true
	default:
		return false
	}
}
