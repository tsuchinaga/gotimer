package gotimer

import "time"

// NewTerm - 新しい期間を返す
func NewTerm(start, stop Time) Term {
	return Term{start: start, stop: stop}
}

// Term - 期間の設定
type Term struct {
	start Time
	stop  Time
}

// runnable - startとstopの間にnowがあれば実行可能
func (t *Term) runnable(now time.Time) bool {
	n := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, now.Location())
	start := time.Date(now.Year(), now.Month(), now.Day(), t.start.hour(), t.start.minute(), t.start.second(), 0, now.Location())
	stop := time.Date(now.Year(), now.Month(), now.Day(), t.stop.hour(), t.stop.minute(), t.stop.second(), 0, now.Location())
	if t.stop < t.start {
		nt := NewTime(now.Hour(), now.Minute(), now.Second())
		if nt <= t.stop { // now <= stop なら、startを前日にする
			start = start.AddDate(0, 0, -1)
		} else {
			stop = stop.AddDate(0, 0, 1)
		}
	}
	return !n.Before(start) && !n.After(stop)
}

// runnableSecond - 実行可能期間を秒で返す
func (t *Term) runnableSecond() int {
	sec := int(t.stop) - int(t.start) + 1
	if sec <= 0 {
		sec += 24 * 60 * 60
	}
	return sec
}

func (t *Term) Equal(term Term) bool {
	return t.start == term.start && t.stop == term.stop
}

func (t *Term) In(time Time) bool {
	if t.stop < t.start { // stop < start
		return time <= t.stop || t.start <= time // start <= time || time <= stop
	} else { // start <= stop
		return t.start <= time && time <= t.stop // start <= time <= stop
	}
}
