package gotimer

import (
	"reflect"
	"testing"
	"time"
)

func Test_Term_runnable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		term Term
		now  time.Time
		want bool
	}{
		{name: "start == stopで、now < startならfalse",
			term: Term{Time(9 * 60 * 60), Time(9 * 60 * 60)},
			now:  time.Date(2020, 12, 25, 8, 59, 59, 0, time.Local),
			want: false},
		{name: "start == stopで、now == startならtrue",
			term: Term{Time(9 * 60 * 60), Time(9 * 60 * 60)},
			now:  time.Date(2020, 12, 25, 9, 0, 0, 0, time.Local),
			want: true},
		{name: "start == stopで、stop < nowならfalse",
			term: Term{Time(9 * 60 * 60), Time(9 * 60 * 60)},
			now:  time.Date(2020, 12, 25, 9, 0, 1, 0, time.Local),
			want: false},
		{name: "start < stopで、now < startならfalse",
			term: Term{Time(9 * 60 * 60), Time(11*60*60 + 30*60)},
			now:  time.Date(2020, 12, 25, 8, 59, 59, 0, time.Local),
			want: false},
		{name: "start < stopで、now == startならtrue",
			term: Term{Time(9 * 60 * 60), Time(11*60*60 + 30*60)},
			now:  time.Date(2020, 12, 25, 9, 0, 0, 0, time.Local),
			want: true},
		{name: "start < stopで、start < now < stopならtrue",
			term: Term{Time(9 * 60 * 60), Time(11*60*60 + 30*60)},
			now:  time.Date(2020, 12, 25, 10, 15, 0, 0, time.Local),
			want: true},
		{name: "start < stopで、now == stopならtrue",
			term: Term{Time(9 * 60 * 60), Time(11*60*60 + 30*60)},
			now:  time.Date(2020, 12, 25, 11, 30, 0, 0, time.Local),
			want: true},
		{name: "start < stopで、stop < nowならfalse",
			term: Term{Time(9 * 60 * 60), Time(11*60*60 + 30*60)},
			now:  time.Date(2020, 12, 25, 11, 30, 1, 0, time.Local),
			want: false},
		{name: "stop < startで、stop < now < startならfalse",
			term: Term{Time(16*60*60 + 30*60), Time(5*60*60 + 24*60 + 59)},
			now:  time.Date(2020, 12, 25, 16, 29, 59, 0, time.Local),
			want: false},
		{name: "stop < startで、now == startならtrue",
			term: Term{Time(16*60*60 + 30*60), Time(5*60*60 + 24*60 + 59)},
			now:  time.Date(2020, 12, 25, 16, 30, 0, 0, time.Local),
			want: true},
		{name: "stop < startで、start < now < 24:00:00ならtrue",
			term: Term{Time(16*60*60 + 30*60), Time(5*60*60 + 24*60 + 59)},
			now:  time.Date(2020, 12, 25, 23, 59, 59, 0, time.Local),
			want: true},
		{name: "stop < startで、00:00:00 < now < stopならtrue",
			term: Term{Time(16*60*60 + 30*60), Time(5*60*60 + 24*60 + 59)},
			now:  time.Date(2020, 12, 26, 0, 0, 1, 0, time.Local),
			want: true},
		{name: "stop < startで、now == stopならtrue",
			term: Term{Time(16*60*60 + 30*60), Time(5*60*60 + 24*60 + 59)},
			now:  time.Date(2020, 12, 26, 5, 24, 59, 0, time.Local),
			want: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.term.runnable(test.now)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_Term_Equal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a, b Term
		want bool
	}{
		{name: "startとstopが一致しているならtrue", want: true,
			a: Term{Time(9 * 60 * 60), Time(15 * 60 * 60)},
			b: Term{Time(9 * 60 * 60), Time(15 * 60 * 60)}},
		{name: "startだけが一致しているならfalse", want: false,
			a: Term{Time(9 * 60 * 60), Time(15 * 60 * 60)},
			b: Term{Time(9 * 60 * 60), Time(11*60*60 + 30*60)}},
		{name: "stopだけが一致しているならfalse", want: false,
			a: Term{Time(9 * 60 * 60), Time(15 * 60 * 60)},
			b: Term{Time(12*60*60 + 30*60), Time(15 * 60 * 60)}},
		{name: "両方一致していないならfalse", want: false,
			a: Term{Time(9 * 60 * 60), Time(11*60*60 + 30*60)},
			b: Term{Time(12*60*60 + 30*60), Time(15 * 60 * 60)}},
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

func Test_Term_In(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		term Term
		time Time
		want bool
	}{
		{name: "start == stopでtime < startならfalse", want: false,
			term: Term{Time(8 * 60 * 60), Time(8 * 60 * 60)},
			time: Time(7*60*60 + 59*60 + 59)},
		{name: "start == stopでstart == time == stopならtrue", want: true,
			term: Term{Time(8 * 60 * 60), Time(8 * 60 * 60)},
			time: Time(8 * 60 * 60)},
		{name: "start == stopでstop < timeならfalse", want: false,
			term: Term{Time(8 * 60 * 60), Time(8 * 60 * 60)},
			time: Time(8*60*60 + 1)},
		{name: "start < stopでtime < stopならfalse", want: false,
			term: Term{Time(8*60*60 + 45*60), Time(15*60*60 + 15*60)},
			time: Time(8*60*60 + 44*60 + 59)},
		{name: "start < stopでstart == timeならtrue", want: true,
			term: Term{Time(8*60*60 + 45*60), Time(15*60*60 + 15*60)},
			time: Time(8*60*60 + 45*60)},
		{name: "start < stopでstart < time < stopならtrue", want: true,
			term: Term{Time(8*60*60 + 45*60), Time(15*60*60 + 15*60)},
			time: Time(9 * 60 * 60)},
		{name: "start < stopでstop == timeならtrue", want: true,
			term: Term{Time(8*60*60 + 45*60), Time(15*60*60 + 15*60)},
			time: Time(15*60*60 + 15*60)},
		{name: "start < stopでstop < timeならfalse", want: false,
			term: Term{Time(8*60*60 + 45*60), Time(15*60*60 + 15*60)},
			time: Time(15*60*60 + 15*60 + 1)},
		{name: "stop < startでstop < time < startならfalse", want: false,
			term: Term{Time(16*60*60 + 30*60), Time(5*60*60 + 30*60)},
			time: Time(16*60*60 + 29*60 + 59)},
		{name: "stop < startでstart == timeならtrue", want: true,
			term: Term{Time(16*60*60 + 30*60), Time(5*60*60 + 30*60)},
			time: Time(16*60*60 + 30*60)},
		{name: "stop < startでstart < timeならtrue", want: true,
			term: Term{Time(16*60*60 + 30*60), Time(5*60*60 + 30*60)},
			time: Time(23*60*60 + 59*60 + 59)},
		{name: "stop < startでtime < stopならtrue", want: true,
			term: Term{Time(16*60*60 + 30*60), Time(5*60*60 + 30*60)},
			time: Time(0)},
		{name: "stop < startでstop == timeならtrue", want: true,
			term: Term{Time(16*60*60 + 30*60), Time(5*60*60 + 30*60)},
			time: Time(5*60*60 + 30*60)},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.term.In(test.time)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_NewTerm(t *testing.T) {
	t.Parallel()
	start := NewTime(12, 30, 0)
	stop := NewTime(15, 0, 0)
	want := Term{start: start, stop: stop}
	got := NewTerm(start, stop)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want, got)
	}
}

func Test_Term_runnableSecond(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		term Term
		want int
	}{
		{name: "00:00:00, 00:00:00は1", term: NewTerm(NewTime(0, 0, 0), NewTime(0, 0, 0)), want: 1},
		{name: "00:00:00, 00:00:01は2", term: NewTerm(NewTime(0, 0, 0), NewTime(0, 0, 1)), want: 2},
		{name: "00:00:00, 00:00:59は60", term: NewTerm(NewTime(0, 0, 0), NewTime(0, 0, 59)), want: 60},
		{name: "00:00:00, 00:01:00は61", term: NewTerm(NewTime(0, 0, 0), NewTime(0, 1, 0)), want: 61},
		{name: "00:00:00, 00:59:59は3600", term: NewTerm(NewTime(0, 0, 0), NewTime(0, 59, 59)), want: 3600},
		{name: "00:00:00, 01:00:00は3601", term: NewTerm(NewTime(0, 0, 0), NewTime(1, 0, 0)), want: 3601},
		{name: "00:00:00, 23:59:59は86400", term: NewTerm(NewTime(0, 0, 0), NewTime(23, 59, 59)), want: 24 * 60 * 60},
		{name: "00:00:01, 00:00:00は86400", term: NewTerm(NewTime(0, 0, 1), NewTime(0, 0, 0)), want: 24 * 60 * 60},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.term.runnableSecond()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}
