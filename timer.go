package gotimer

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

var (
	TimerNotSetContextError  = errors.New("not set ctx")
	TimerNotSetIntervalError = errors.New("not set interval")
	TimerNotSetTaskError     = errors.New("not set task")
	TimerIsRunningError      = errors.New("timer is running now")
)

// Timer - タイマー
type Timer struct {
	interval         time.Duration
	terms            []Term
	currentTerm      int
	ch               chan time.Time
	timerRunning     bool
	taskRunning      int
	parallelRunnable bool
	next             time.Time
	timer            time.Timer
	mtx              sync.Mutex
}

// SetParallelRunnable - タスクの多重実行を許容するか
func (t *Timer) SetParallelRunnable(runnable bool) *Timer {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	if t.timerRunning {
		return t
	}

	t.parallelRunnable = runnable
	return t
}

// AddTerm - 実行期間を追加する
func (t *Timer) AddTerm(term Term) *Timer {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	if t.timerRunning {
		return t
	}

	if t.terms == nil {
		t.terms = []Term{}
	}
	// 重複チェック 同じ期間があれば追加しない
	for _, tt := range t.terms {
		if tt.Equal(term) {
			return t
		}
	}
	t.terms = append(t.terms, term)

	// startが小さいか、startが同じなら実行時間が長いのを前にする
	sort.Slice(t.terms, func(i, j int) bool {
		return t.terms[i].start < t.terms[j].start ||
			(t.terms[i].start == t.terms[j].start && t.terms[i].runnableSecond() > t.terms[j].runnableSecond())
	})
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
	if t.timerRunning {
		t.mtx.Unlock()
		return TimerIsRunningError
	}
	t.timerRunning = true
	defer func() {
		t.mtx.Lock()
		t.timerRunning = false
		t.mtx.Unlock()
	}()
	t.mtx.Unlock()
	t.interval = interval
	if t.terms == nil {
		t.terms = append(t.terms, NewTerm(NewTime(0, 0, 0), NewTime(23, 59, 59)))
	}

	for {
		now := time.Now()

		// 次の実行時刻を決定
		t.next = t.nextTime(now)
		d := t.next.Sub(now)
		tm := time.NewTimer(d)
		select {
		case <-tm.C: // 実行時間が来たら非同期で実行
			go func() {
				if t.incrementTaskRunning() { // タスク実行中でないか、多重起動許容の場合にタスクを実行する
					defer t.decrementTaskRunning()
					task()
				}
			}()
		case <-ctx.Done(): // ctxの終了ならreturn nil
			tm.Stop() // そのまま捨てられるタイマーなので発火済みかなど気にしない
			return nil
		}
	}
}

// incrementTaskRunning - 実行中のタスクのカウントを増やす
//   ただし、多重起動不可なら複数起動はしないので、その場合は実質上限1
func (t *Timer) incrementTaskRunning() bool {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	// タスク実行中かつ、多重起動不可ならロックが取れない
	if t.taskRunning > 0 && !t.parallelRunnable {
		return false
	}

	t.taskRunning++
	return true
}

// decrementTaskRunning - 実行中のタスクのカウントを減らす
func (t *Timer) decrementTaskRunning() {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	t.taskRunning--
}

// nextTime - 次回実行日時を取得する
func (t *Timer) nextTime(now time.Time) time.Time {
	var nt time.Time
	// nextがzeroタイムなら直近の開始日時を設定する、zeroタイムでなければ前回実行日時 + intervalを設定する
	if t.next.IsZero() {
		nt = t.nextStart(now)
	} else {
		nt = t.next.Add(t.interval)

		// 次回実行日時が実行可能でなければ、次の開始時刻を採用する
		if !t.runnable(nt) {
			nt = t.nextStart(now)
		}
	}

	return nt
}

// nextStart - 次の開始日時を取得する
//   期間から次の開始日時が取れなかった場合、翌日の0時を返す
func (t *Timer) nextStart(now time.Time) time.Time {
	for i := 0; i <= 1; i++ {
		for _, term := range t.terms {
			nt := time.Date(now.Year(), now.Month(), now.Day(), term.start.hour(), term.start.minute(), term.start.second(), 0, time.Local)
			nt = nt.AddDate(0, 0, i)
			if nt.After(now) {
				return nt
			}
		}
	}
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.Local)
}

// runnable - Timerの持つtermsをすべて見て、実行可能かを返す
func (t *Timer) runnable(now time.Time) bool {
	for _, term := range t.terms {
		if term.runnable(now) {
			return true
		}
	}

	return false
}
