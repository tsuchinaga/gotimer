package gotimer

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

var (
	TimeHourError            = errors.New("hours is between 0 to 24")
	TimeMinuteError          = errors.New("minute is between 0 to 60")
	TimeSecondError          = errors.New("second is between 0 to 60")
	TimerNotSetContextError  = errors.New("not set ctx")
	TimerNotSetIntervalError = errors.New("not set interval")
	TimerNotSetTaskError     = errors.New("not set task")
	TimerIsRunningError      = errors.New("timer is running now")
)

// Timer - タイマー
type Timer struct {
	interval   time.Duration
	startTimes []*Time
	stopTimes  []*Time
	ch         chan time.Time
	running    bool
	next       time.Time
	timer      time.Timer
	mtx        sync.Mutex
}

// AddStartTime - 開始時刻を追加する
func (t *Timer) AddStartTime(time *Time) *Timer {
	if t.startTimes == nil {
		t.startTimes = []*Time{}
	}
	if time != nil {
		// 重複チェック 同じ時刻があれば追加しない
		for _, st := range t.startTimes {
			if st.Equal(time) {
				return t
			}
		}

		t.startTimes = append(t.startTimes, time)
		sort.Slice(t.startTimes, t.sortStartTimesFunc)
	}
	return t
}

// AddStopTime - 停止時刻を追加する
func (t *Timer) AddStopTime(time *Time) *Timer {
	if t.stopTimes == nil {
		t.stopTimes = []*Time{}
	}
	if time != nil {
		// 重複チェック 同じ時刻があれば追加しない
		for _, st := range t.stopTimes {
			if st.Equal(time) {
				return t
			}
		}

		t.stopTimes = append(t.stopTimes, time)
		sort.Slice(t.stopTimes, t.sortStopTimesFunc)
	}
	return t
}

// Run - タイマーの開始
func (t *Timer) Run(ctx context.Context, interval time.Duration, task func()) error {
	if ctx == nil {
		return TimerNotSetContextError
	}
	if interval <= 0 {
		return TimerNotSetIntervalError
	}
	if task == nil {
		return TimerNotSetTaskError
	}

	// 開始してフラグを立てるまでは排他ロック
	t.mtx.Lock()
	if t.running {
		t.mtx.Unlock()
		return TimerIsRunningError
	}
	t.running = true
	defer func() {
		t.mtx.Lock()
		t.running = false
		t.mtx.Unlock()
	}()
	t.mtx.Unlock()
	t.interval = interval

	for {
		now := time.Now()

		// 次の実行時刻を決定
		t.next = t.nextTime(now)
		d := t.next.Sub(now)
		tm := time.NewTimer(d)
		select {
		case <-tm.C: // 実行時間が来たら実行
			task()
		case <-ctx.Done(): // ctxの終了ならreturn nil
			tm.Stop() // そのまま捨てられるタイマーなので発火済みかなど気にしない
			return nil
		}
	}
}

// sortStartTimesFunc - 開始時刻をソートする関数
func (t *Timer) sortStartTimesFunc(i, j int) bool {
	return t.startTimes[i].hour < t.startTimes[j].hour ||
		(t.startTimes[i].hour == t.startTimes[j].hour && t.startTimes[i].minute < t.startTimes[j].minute) ||
		(t.startTimes[i].hour == t.startTimes[j].hour && t.startTimes[i].minute == t.startTimes[j].minute && t.startTimes[i].second == t.startTimes[j].second)
}

// sortStopTimesFunc - 停止時刻をソートする関数
func (t *Timer) sortStopTimesFunc(i, j int) bool {
	return t.stopTimes[i].hour < t.stopTimes[j].hour ||
		(t.stopTimes[i].hour == t.stopTimes[j].hour && t.stopTimes[i].minute < t.stopTimes[j].minute) ||
		(t.stopTimes[i].hour == t.stopTimes[j].hour && t.stopTimes[i].minute == t.stopTimes[j].minute && t.stopTimes[i].second == t.stopTimes[j].second)
}

// nextTime - 次回実行日時を取得する
func (t *Timer) nextTime(now time.Time) time.Time {
	var nt time.Time
	// nextがzeroタイムなら直近の開始日時を設定する、zeroタイムでなければ前回実行日時 + intervalを設定する
	if t.next.IsZero() {
		nt = t.nextStart(now)
	} else {
		nt = t.next.Add(t.interval)

		// 前回実行日時から次回実行日時までに停止日時があればその次の実行時刻を取る
		nst := t.nextStop(now)
		if t.next.Before(nst) && !nt.Before(nst) { // 前回実行日時 < 停止日時 && 停止日時 <= 次回実行日時
			nt = t.nextStart(now)
		}
	}

	return nt
}

// nextStart - 次の開始日時を取得する
// startTimesが空で次の開始時刻が取れない場合、現在時刻 + インターバルを返す
func (t *Timer) nextStart(now time.Time) time.Time {
	for i := 0; i <= 1; i++ {
		for _, st := range t.startTimes {
			nt := time.Date(now.Year(), now.Month(), now.Day(), st.hour, st.minute, st.second, 0, time.Local)
			nt = nt.AddDate(0, 0, i)
			if nt.After(now) {
				return nt
			}
		}
	}
	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, time.Local).Add(t.interval)
}

// nextStop - 次の停止日時を取得する
// stopTimesが空で次の開始時刻が取れない場合、1日後を返す
func (t *Timer) nextStop(now time.Time) time.Time {
	for i := 0; i <= 1; i++ {
		for _, st := range t.stopTimes {
			nt := time.Date(now.Year(), now.Month(), now.Day(), st.hour, st.minute, st.second, 0, time.Local)
			nt = nt.AddDate(0, 0, i)
			if nt.After(now) {
				return nt
			}
		}
	}
	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, time.Local).AddDate(0, 0, 1)
}

// NewTime - 新しいgotimer.Timeを生成する
func NewTime(hour, minute, second int) (*Time, error) {
	if hour < 0 || 24 <= hour {
		return nil, TimeHourError
	}

	if minute < 0 || 60 <= minute {
		return nil, TimeMinuteError
	}

	if second < 0 || 60 <= second {
		return nil, TimeSecondError
	}

	return &Time{hour: hour, minute: minute, second: second}, nil
}

// Time - 時分秒だけを持った構造体
type Time struct {
	hour   int
	minute int
	second int
}

func (t *Time) Equal(time *Time) bool {
	return t.hour == time.hour && t.minute == time.minute && t.second == time.second
}
