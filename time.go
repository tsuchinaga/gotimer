package gotimer

// NewTime - 新しいgotimer.Timeを生成する
func NewTime(hour, minute, second int) Time {
	sec := (hour*60*60 + minute*60 + second) % (24 * 60 * 60)
	if sec < 0 {
		sec = 24*60*60 + sec
	}

	return Time(sec)
}

// Time - 時分秒を秒にした構造体
type Time int

func (t Time) hour() int {
	return int(t) / 60 / 60
}

func (t Time) minute() int {
	return (int(t) / 60) % 60
}

func (t Time) second() int {
	return int(t) % 60
}
