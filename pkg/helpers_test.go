package pkg

import (
	"reflect"
	"testing"
	"time"
)

func TestGenerateKeyNames(t *testing.T) {
	type args struct {
		tr time.Time
	}
	tests := []struct {
		name string
		args args
		want *AllKeys
	}{
		{
			name: "default test",
			args: args{
				tr: time.Date(2017, time.March, 7, 06, 30, 0, 0, time.UTC),
			},
			want: &AllKeys{
				Src: KeyNames{
					CurrentCounterSet:  "set:counter:src:1488868200",
					CurrentCounterHSet: "hset:counter:src:1488868200",
					HourlyCounterSet:   "set:counter:src:1488866400",
					HourlyCounterHSet:  "hset:counter:src:1488866400",
				},
				Dst: KeyNames{
					CurrentCounterSet:  "set:counter:dst:1488868200",
					CurrentCounterHSet: "hset:counter:dst:1488868200",
					HourlyCounterSet:   "set:counter:dst:1488866400",
					HourlyCounterHSet:  "hset:counter:dst:1488866400",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateKeyNames(tt.args.tr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateKeyNames() = %v, want %v", got, tt.want)
			}
		})
	}
}
