package gotimer

import (
	"reflect"
	"testing"
)

func Test_Time_hour(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		time Time
		want int
	}{
		{name: "00:00:00なら0", time: NewTime(0, 0, 0), want: 0},
		{name: "23:59:59なら23", time: NewTime(23, 59, 59), want: 23},
		{name: "12:24:48は12", time: NewTime(12, 24, 48), want: 12},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.time.hour()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_Time_minute(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		time Time
		want int
	}{
		{name: "00:00:00は0", time: NewTime(0, 0, 0), want: 0},
		{name: "23:59:59は59", time: NewTime(23, 59, 59), want: 59},
		{name: "12:24:48は24", time: NewTime(12, 24, 48), want: 24},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.time.minute()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_Time_second(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		time Time
		want int
	}{
		{name: "00:00:00は0", time: NewTime(0, 0, 0), want: 0},
		{name: "23:59:59は59", time: NewTime(23, 59, 59), want: 59},
		{name: "12:24:48は48", time: NewTime(12, 24, 48), want: 48},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.time.second()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_NewTime(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		hour, minute, second int
		want                 Time
	}{
		{name: "00:00:00 => 0, 0, 0", hour: 0, minute: 0, second: 0, want: Time(0)},
		{name: "23:59:59 => 23, 59, 59", hour: 23, minute: 59, second: 59, want: Time(23*60*60 + 59*60 + 59)},
		{name: "-24:00:00 => 0, 0, 0", hour: -24, minute: 0, second: 0, want: Time(0)},
		{name: "-01:00:00 => 23, 0, 0", hour: -1, minute: 0, second: 0, want: Time(23 * 60 * 60)},
		{name: "00:-01:00 => 23, 59, 0", hour: 0, minute: -1, second: 0, want: Time(23*60*60 + 59*60)},
		{name: "00:00:-01 => 23, 59, 59", hour: 0, minute: 0, second: -1, want: Time(23*60*60 + 59*60 + 59)},
		{name: "00:00:-61 => 23, 58, 59", hour: 0, minute: 0, second: -61, want: Time(23*60*60 + 58*60 + 59)},
		{name: "00:00:-86400 => 0, 0, 0", hour: 0, minute: 0, second: -86400, want: Time(0)},
		{name: "00:00:-86401 => 23, 59, 59", hour: 0, minute: 0, second: -86401, want: Time(23*60*60 + 59*60 + 59)},
		{name: "24:59:-86401 => 23, 59, 59", hour: 24, minute: 59, second: -86401, want: Time(58*60 + 59)},
		{name: "24:60:60 => 1, 1, 0", hour: 24, minute: 60, second: 60, want: Time(1*60*60 + 1*60)},
		{name: "24:00:00 => 0, 0, 0", hour: 24, minute: 0, second: 0, want: Time(0)},
		{name: "00:60:00 => 1, 0, 0", hour: 0, minute: 60, second: 0, want: Time(1 * 60 * 60)},
		{name: "00:00:60 => 0, 1, 0", hour: 0, minute: 0, second: 60, want: Time(1 * 60)},
		{name: "23:59:60 => 0, 0, 0", hour: 23, minute: 59, second: 60, want: Time(0)},
		{name: "24:00:-01 => 23, 59, 59", hour: 24, minute: 0, second: -1, want: Time(23*60*60 + 59*60 + 59)},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := NewTime(test.hour, test.minute, test.second)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}
