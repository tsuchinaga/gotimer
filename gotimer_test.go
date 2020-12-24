package gotimer

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
)

func Test_NewTime(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		hour, minute, second int
		want1                *Time
		want2                error
	}{
		{name: "00:00:00は可能", hour: 0, minute: 0, second: 0, want1: &Time{hour: 0, minute: 0, second: 0}},
		{name: "23:59:59は可能", hour: 23, minute: 59, second: 59, want1: &Time{hour: 23, minute: 59, second: 59}},
		{name: "-01:00:00はエラー", hour: -1, minute: 0, second: 0, want2: TimeHourError},
		{name: "00:-01:00はエラー", hour: 0, minute: -1, second: 0, want2: TimeMinuteError},
		{name: "00:00:-01はエラー", hour: 0, minute: 0, second: -1, want2: TimeSecondError},
		{name: "24:00:00はエラー", hour: 24, minute: 0, second: 0, want2: TimeHourError},
		{name: "00:60:00はエラー", hour: 0, minute: 60, second: 0, want2: TimeMinuteError},
		{name: "00:00:60はエラー", hour: 0, minute: 0, second: 60, want2: TimeSecondError},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1, got2 := NewTime(test.hour, test.minute, test.second)
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
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
		{name: "開始時刻がなければ現在日時 + intervalが返される",
			timer: &Timer{interval: 15 * time.Second},
			now:   time.Date(2020, 12, 21, 10, 05, 30, 123456789, time.Local),
			want:  time.Date(2020, 12, 21, 10, 05, 45, 0, time.Local)},
		{name: "当日に未来の開始時刻があればその開始日時が返される",
			timer: &Timer{interval: 15 * time.Second, startTimes: []*Time{{10, 0, 0}, {10, 10, 0}, {10, 20, 0}}},
			now:   time.Date(2020, 12, 21, 10, 05, 30, 123456789, time.Local),
			want:  time.Date(2020, 12, 21, 10, 10, 0, 0, time.Local)},
		{name: "開始時刻がすべて過去なら翌日の開始日時が返される",
			timer: &Timer{interval: 15 * time.Second, startTimes: []*Time{{1, 0, 0}, {6, 0, 0}, {10, 0, 0}}},
			now:   time.Date(2020, 12, 21, 10, 05, 30, 123456789, time.Local),
			want:  time.Date(2020, 12, 22, 1, 0, 0, 0, time.Local)},
		{name: "現在日時と重なった開始時刻は返されない",
			timer: &Timer{interval: 15 * time.Second, startTimes: []*Time{{10, 0, 0}, {10, 5, 30}, {10, 6, 0}}},
			now:   time.Date(2020, 12, 21, 10, 05, 30, 123456789, time.Local),
			want:  time.Date(2020, 12, 21, 10, 6, 0, 0, time.Local)},
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

func Test_Timer_nextStop(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		timer *Timer
		now   time.Time
		want  time.Time
	}{
		{name: "停止時刻がなければ現在日時 + 1日が返される",
			timer: &Timer{interval: 15 * time.Second},
			now:   time.Date(2020, 12, 21, 10, 05, 30, 123456789, time.Local),
			want:  time.Date(2020, 12, 22, 10, 05, 30, 0, time.Local)},
		{name: "当日に未来の停止時刻があればその停止日時が返される",
			timer: &Timer{interval: 15 * time.Second, stopTimes: []*Time{{10, 0, 0}, {10, 10, 0}, {10, 20, 0}}},
			now:   time.Date(2020, 12, 21, 10, 05, 30, 123456789, time.Local),
			want:  time.Date(2020, 12, 21, 10, 10, 0, 0, time.Local)},
		{name: "停止時刻がすべて過去なら翌日の停止日時が返される",
			timer: &Timer{interval: 15 * time.Second, stopTimes: []*Time{{1, 0, 0}, {6, 0, 0}, {10, 0, 0}}},
			now:   time.Date(2020, 12, 21, 10, 05, 30, 123456789, time.Local),
			want:  time.Date(2020, 12, 22, 1, 0, 0, 0, time.Local)},
		{name: "現在日時と重なった停止時刻は返されない",
			timer: &Timer{interval: 15 * time.Second, stopTimes: []*Time{{10, 0, 0}, {10, 5, 30}, {10, 6, 0}}},
			now:   time.Date(2020, 12, 21, 10, 05, 30, 123456789, time.Local),
			want:  time.Date(2020, 12, 21, 10, 6, 0, 0, time.Local)},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.timer.nextStop(test.now)
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
			timer: &Timer{startTimes: []*Time{{12, 0, 0}}},
			now:   time.Date(2020, 12, 21, 11, 0, 0, 0, time.Local),
			want:  time.Date(2020, 12, 21, 12, 0, 0, 0, time.Local)},
		{name: "前回の実行日時からinterval後の日時を返す",
			timer: &Timer{interval: 15 * time.Second, startTimes: []*Time{{12, 0, 0}}, next: time.Date(2020, 12, 21, 11, 0, 0, 0, time.Local)},
			now:   time.Date(2020, 12, 21, 11, 0, 0, 0, time.Local),
			want:  time.Date(2020, 12, 21, 11, 0, 15, 0, time.Local)},
		{name: "前回の実行日時からinterval後の日時の間に停止日時があれば次の開始時刻を返す",
			timer: &Timer{
				interval:   15 * time.Second,
				startTimes: []*Time{{12, 0, 0}},
				stopTimes:  []*Time{{11, 0, 10}},
				next:       time.Date(2020, 12, 21, 11, 0, 0, 0, time.Local)},
			now:  time.Date(2020, 12, 21, 11, 0, 0, 0, time.Local),
			want: time.Date(2020, 12, 21, 12, 0, 0, 0, time.Local)},
		{name: "前回の実行日時からinterval後の日時の間に停止日時があればそれが停止日時前であっても次の開始時刻を返す",
			timer: &Timer{
				interval:   15 * time.Second,
				startTimes: []*Time{{11, 0, 8}},
				stopTimes:  []*Time{{11, 0, 10}},
				next:       time.Date(2020, 12, 21, 11, 0, 0, 0, time.Local)},
			now:  time.Date(2020, 12, 21, 11, 0, 0, 0, time.Local),
			want: time.Date(2020, 12, 21, 11, 0, 8, 0, time.Local)},
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

func Test_Timer_AddStartTime(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		startTimes []*Time
		time       *Time
		want       []*Time
	}{
		{name: "空っぽのところに追加できる",
			startTimes: nil,
			time:       &Time{12, 0, 0},
			want:       []*Time{{12, 0, 0}}},
		{name: "すでに入っているところに追加した場合、自動で並び変えられる",
			startTimes: []*Time{{0, 0, 0}, {6, 0, 0}, {18, 0, 0}},
			time:       &Time{12, 0, 0},
			want:       []*Time{{0, 0, 0}, {6, 0, 0}, {12, 0, 0}, {18, 0, 0}}},
		{name: "同じ時刻は重複してはいらない",
			startTimes: []*Time{{0, 0, 0}, {6, 0, 0}, {12, 0, 0}, {18, 0, 0}},
			time:       &Time{12, 0, 0},
			want:       []*Time{{0, 0, 0}, {6, 0, 0}, {12, 0, 0}, {18, 0, 0}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			timer := new(Timer)
			timer.startTimes = test.startTimes
			timer.AddStartTime(test.time)
			if !reflect.DeepEqual(test.want, timer.startTimes) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, timer.startTimes)
			}
		})
	}
}

func Test_Timer_AddStopTime(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		stopTimes []*Time
		time      *Time
		want      []*Time
	}{
		{name: "空っぽのところに追加できる",
			stopTimes: nil,
			time:      &Time{12, 0, 0},
			want:      []*Time{{12, 0, 0}}},
		{name: "すでに入っているところに追加した場合、自動で並び変えられる",
			stopTimes: []*Time{{0, 0, 0}, {6, 0, 0}, {18, 0, 0}},
			time:      &Time{12, 0, 0},
			want:      []*Time{{0, 0, 0}, {6, 0, 0}, {12, 0, 0}, {18, 0, 0}}},
		{name: "同じ時刻は重複してはいらない",
			stopTimes: []*Time{{0, 0, 0}, {6, 0, 0}, {12, 0, 0}, {18, 0, 0}},
			time:      &Time{12, 0, 0},
			want:      []*Time{{0, 0, 0}, {6, 0, 0}, {12, 0, 0}, {18, 0, 0}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			timer := new(Timer)
			timer.stopTimes = test.stopTimes
			timer.AddStopTime(test.time)
			if !reflect.DeepEqual(test.want, timer.stopTimes) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, timer.stopTimes)
			}
		})
	}
}

func Test_Time_Equal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a, b *Time
		want bool
	}{
		{name: "一致していればtrue",
			a:    &Time{1, 2, 3},
			b:    &Time{1, 2, 3},
			want: true},
		{name: "一致していなければfalse",
			a:    &Time{1, 2, 3},
			b:    &Time{1, 2, 0},
			want: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.a.Equal(test.b)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

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
	var count int
	timer := new(Timer)
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
