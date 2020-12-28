package gotimer

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"
)

func Test_Timer_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		timer    *Timer
		ctx      context.Context
		interval time.Duration
		task     func()
		want     error
	}{
		{name: "ctxが未設定ならerror", timer: &Timer{}, want: TimerNotSetContextError},
		{name: "intervalが1未満ならerror", timer: &Timer{}, ctx: context.Background(), want: TimerNotSetIntervalError},
		{name: "taskがnilならerror", timer: &Timer{}, ctx: context.Background(), interval: 15 * time.Second, want: TimerNotSetTaskError},
		{name: "isRunningがtrueならerror", timer: &Timer{timerRunning: true}, ctx: context.Background(), interval: 15 * time.Second, task: func() {}, want: TimerIsRunningError},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.timer.Run(test.ctx, test.interval, test.task)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_Timer_Run_Cancel(t *testing.T) {
	t.Parallel()
	var count int
	timer := new(Timer)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	interval := 3 * time.Second
	task := func() { count++ }
	got := timer.Run(ctx, interval, task)
	if count != 0 || got != nil {
		t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), 0, nil, count, got)
	}
	cancel()
}

func Test_Timer_Run_Task(t *testing.T) {
	t.Parallel()
	now := time.Now()
	var count int
	timer := new(Timer)
	timer.AddTerm(NewTerm(NewTime(now.Hour(), now.Minute(), now.Second()+1), NewTime(now.Hour(), now.Minute(), now.Second()+10)))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	interval := 1 * time.Second
	task := func() { count++ }
	got := timer.Run(ctx, interval, task)
	if count != 3 || got != nil {
		t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), 3, nil, count, got)
	}
	cancel()
}

func Test_Timer_Run_Parallel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		parallel bool
		timeout  time.Duration
		interval time.Duration
		want     int
	}{
		{name: "Parallelが許可されている場合、複数回実行される", parallel: true, timeout: 1 * time.Second, interval: 100 * time.Millisecond, want: 10},
		{name: "Parallelが許可されていない場合、1回しか実行されない", parallel: false, timeout: 1 * time.Second, interval: 100 * time.Millisecond, want: 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var count int
			var mtx sync.Mutex
			task := func() {
				mtx.Lock()
				count++
				mtx.Unlock()
				time.Sleep(test.timeout)
			}
			timer := new(Timer)
			ctx, cancel := context.WithTimeout(context.Background(), test.timeout)
			defer cancel()
			err := timer.SetParallelRunnable(test.parallel).Run(ctx, test.interval, task)
			if test.want != count && err != nil {
				t.Errorf("%s error\nwant: %+v\ngot: %+v, %+v\n", t.Name(), test.want, count, err)
			}
		})
	}
}

func Test_Timer_SetParallelRunnable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		timerRunning bool
		want         bool
	}{
		{name: "timerRunningでなければ設定が反映される", timerRunning: false, want: true},
		{name: "timerRunningであれば設定が反映されない", timerRunning: true, want: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			timer := &Timer{timerRunning: test.timerRunning}
			timer.SetParallelRunnable(true)
			got := timer.parallelRunnable
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_Timer_AddTerm(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		timerRunning bool
		terms        []Term
		term         Term
		want         []Term
	}{
		{name: "timerRunningであれば変更が反映されない",
			timerRunning: true,
			terms:        []Term{NewTerm(NewTime(8, 45, 0), NewTime(15, 15, 0))},
			term:         NewTerm(NewTime(16, 30, 0), NewTime(5, 30, 0)),
			want:         []Term{NewTerm(NewTime(8, 45, 0), NewTime(15, 15, 0))}},
		{name: "timerRunningでなければ期間が追加される",
			timerRunning: false,
			terms:        []Term{NewTerm(NewTime(8, 45, 0), NewTime(15, 15, 0))},
			term:         NewTerm(NewTime(16, 30, 0), NewTime(5, 30, 0)),
			want:         []Term{NewTerm(NewTime(8, 45, 0), NewTime(15, 15, 0)), NewTerm(NewTime(16, 30, 0), NewTime(5, 30, 0))}},
		{name: "同じ期間があれば追加されない",
			timerRunning: false,
			terms:        []Term{NewTerm(NewTime(8, 45, 0), NewTime(15, 15, 0))},
			term:         NewTerm(NewTime(8, 45, 0), NewTime(15, 15, 0)),
			want:         []Term{NewTerm(NewTime(8, 45, 0), NewTime(15, 15, 0))}},
		{name: "追加後、開始時刻の昇順、停止時刻の降順で並び変えられる",
			timerRunning: false,
			terms:        []Term{NewTerm(NewTime(8, 45, 0), NewTime(15, 15, 0))},
			term:         NewTerm(NewTime(8, 45, 0), NewTime(5, 30, 0)),
			want:         []Term{NewTerm(NewTime(8, 45, 0), NewTime(5, 30, 0)), NewTerm(NewTime(8, 45, 0), NewTime(15, 15, 0))}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			timer := Timer{timerRunning: test.timerRunning, terms: test.terms}
			timer.AddTerm(test.term)
			got := timer.terms
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_Timer_nextTime(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		timer *Timer
		now   time.Time
		want  time.Time
	}{
		{name: "前回の実行日時が分からなければ次の開始日時を返す",
			timer: &Timer{terms: []Term{NewTerm(NewTime(12, 0, 0), NewTime(13, 0, 0))}},
			now:   time.Date(2020, 12, 21, 11, 0, 0, 0, time.Local),
			want:  time.Date(2020, 12, 21, 12, 0, 0, 0, time.Local)},
		{name: "前回の実行日時からinterval後の日時を返す",
			timer: &Timer{
				interval: 15 * time.Second,
				terms:    []Term{NewTerm(NewTime(11, 0, 0), NewTime(12, 0, 0))},
				next:     time.Date(2020, 12, 21, 11, 0, 0, 0, time.Local)},
			now:  time.Date(2020, 12, 21, 11, 0, 0, 0, time.Local),
			want: time.Date(2020, 12, 21, 11, 0, 15, 0, time.Local)},
		{name: "前回の実行日時からinterval後の日時の間に停止日時があれば次の開始時刻を返す",
			timer: &Timer{
				interval: 15 * time.Second,
				terms:    []Term{NewTerm(NewTime(11, 0, 0), NewTime(11, 0, 10))},
				next:     time.Date(2020, 12, 21, 11, 0, 0, 0, time.Local)},
			now:  time.Date(2020, 12, 21, 11, 0, 0, 0, time.Local),
			want: time.Date(2020, 12, 22, 11, 0, 0, 0, time.Local)},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.timer.nextTime(test.now)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_Timer_runnable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		terms []Term
		now   time.Time
		want  bool
	}{
		{name: "termsのいずれかがtrueならtrueを返す",
			terms: []Term{
				NewTerm(NewTime(0, 0, 0), NewTime(1, 0, 0)),
				NewTerm(NewTime(2, 0, 0), NewTime(3, 0, 0)),
				NewTerm(NewTime(4, 0, 0), NewTime(5, 0, 0)),
				NewTerm(NewTime(6, 0, 0), NewTime(7, 0, 0)),
			},
			now:  time.Date(2020, 12, 28, 6, 30, 0, 0, time.Local),
			want: true},
		{name: "termsのいずれもfalseならfalseを返す",
			terms: []Term{
				NewTerm(NewTime(0, 0, 0), NewTime(1, 0, 0)),
				NewTerm(NewTime(2, 0, 0), NewTime(3, 0, 0)),
				NewTerm(NewTime(4, 0, 0), NewTime(5, 0, 0)),
				NewTerm(NewTime(6, 0, 0), NewTime(7, 0, 0)),
			},
			now:  time.Date(2020, 12, 28, 5, 30, 0, 0, time.Local),
			want: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := (&Timer{terms: test.terms}).runnable(test.now)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_Timer_nextStart(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		timer *Timer
		now   time.Time
		want  time.Time
	}{
		{name: "開始時刻がなければ翌日の00:00:00が返される",
			timer: &Timer{interval: 15 * time.Second, terms: []Term{}},
			now:   time.Date(2020, 12, 21, 10, 05, 30, 123456789, time.Local),
			want:  time.Date(2020, 12, 22, 0, 0, 0, 0, time.Local)},
		{name: "当日に未来の開始時刻があればその開始日時が返される",
			timer: &Timer{interval: 15 * time.Second, terms: []Term{NewTerm(NewTime(10, 10, 0), NewTime(10, 20, 0))}},
			now:   time.Date(2020, 12, 21, 10, 05, 30, 123456789, time.Local),
			want:  time.Date(2020, 12, 21, 10, 10, 0, 0, time.Local)},
		{name: "開始時刻がすべて過去なら翌日の開始日時が返される",
			timer: &Timer{interval: 15 * time.Second, terms: []Term{
				NewTerm(NewTime(1, 0, 0), NewTime(23, 59, 59)),
				NewTerm(NewTime(2, 0, 0), NewTime(23, 59, 59)),
				NewTerm(NewTime(10, 0, 0), NewTime(23, 59, 59)),
			}},
			now:  time.Date(2020, 12, 21, 10, 05, 30, 123456789, time.Local),
			want: time.Date(2020, 12, 22, 1, 0, 0, 0, time.Local)},
		{name: "現在日時と重なった開始時刻は返されない",
			timer: &Timer{interval: 15 * time.Second, terms: []Term{
				NewTerm(NewTime(10, 0, 0), NewTime(23, 59, 59)),
				NewTerm(NewTime(10, 5, 30), NewTime(23, 59, 59)),
				NewTerm(NewTime(10, 6, 0), NewTime(23, 59, 59)),
			}},
			now:  time.Date(2020, 12, 21, 10, 05, 30, 123456789, time.Local),
			want: time.Date(2020, 12, 21, 10, 6, 0, 0, time.Local)},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.timer.nextStart(test.now)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_Timer_incrementTaskRunning(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		taskRunning      int
		parallelRunnable bool
		want1            bool
		want2            int
	}{
		{name: "多重起動不可で実行数が1なら実行できないのでfalse",
			taskRunning: 1, parallelRunnable: false, want1: false, want2: 1},
		{name: "多重起動不可で実行数が0なら実行できるのでtrue",
			taskRunning: 0, parallelRunnable: false, want1: true, want2: 1},
		{name: "多重起動可で実行数が1なら実行できるのでtrue",
			taskRunning: 1, parallelRunnable: true, want1: true, want2: 2},
		{name: "多重起動可で実行数が0なら実行できるのでtrue",
			taskRunning: 0, parallelRunnable: true, want1: true, want2: 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			timer := &Timer{taskRunning: test.taskRunning, parallelRunnable: test.parallelRunnable}
			got := timer.incrementTaskRunning()
			if !reflect.DeepEqual(test.want1, got) || !reflect.DeepEqual(test.want2, timer.taskRunning) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got, timer.taskRunning)
			}
		})
	}
}
