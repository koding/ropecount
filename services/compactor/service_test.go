package compactor

import (
	"errors"
	"flag"
	"math/rand"
	"strconv"
	"testing"
	"time"

	redigo "github.com/garyburd/redigo/redis"
	"github.com/koding/redis"
	"github.com/koding/ropecount/pkg"
)

func withApp(fn func(app *pkg.App)) {
	name := "compator_test"
	conf := flag.CommandLine // tests add more stuff

	pkg.AddHTTPConf(conf)
	pkg.AddRedisConf(conf)

	app := pkg.NewApp(name, conf)
	fn(app)
}

func Test_compactorService_incrementMapValues(t *testing.T) {
	withApp(func(app *pkg.App) {
		var redisConn *redis.RedisSession
		{
			redisConn = app.MustGetRedis()
			rand.Seed(time.Now().UnixNano())
			prefix := strconv.Itoa(rand.Int())
			redisConn.SetPrefix(prefix)
		}

		type fields struct {
			app *pkg.App
		}
		type args struct {
			redisConn *redis.RedisSession
			target    string
			fns       map[string]int64
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			wantErr bool
		}{
			{
				name: "bare call",
				fields: fields{
					app: app,
				},
				args: args{
					redisConn: redisConn,
					target:    "my_test_hash_bare_call",
					fns: map[string]int64{
						"key1": 1,
						"key2": 2,
						"key3": 3,
					},
				},
				wantErr: false,
			},
			{
				name: "empty call",
				fields: fields{
					app: app,
				},
				args: args{
					redisConn: redisConn,
					target:    "my_test_hash_empty_call",
					fns:       map[string]int64{},
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c := &compactorService{
					app: tt.fields.app,
				}
				if err := c.incrementMapValues(tt.args.redisConn, tt.args.target, tt.args.fns); (err != nil) != tt.wantErr {
					t.Errorf("compactorService.incrementMapValues() error = %v, wantErr %v", err, tt.wantErr)
				}

				fns, err := redigo.Int64Map(redisConn.HashGetAll(tt.args.target))
				if err != nil {
					t.Errorf("redigo.Int64Map(redisConn.HashGetAll(tt.args.target)) error = %v", err)
				}

				if len(fns) != len(tt.args.fns) {
					t.Errorf("len(fns) %d != len(tt.args.fns) %d", len(fns), len(tt.args.fns))
				}

				for key, val := range tt.args.fns {
					if fns[key] != val {
						t.Errorf(" fns[key] != val | %d != %d", fns[key], val)
					}
				}

				if _, err := redisConn.Del(tt.args.target); err != nil {
					t.Errorf("redisConn.Del(tt.args.target) error = %v", err)
				}
			})
		}
	})
}

func Test_compactorService_merge(t *testing.T) {
	withApp(func(app *pkg.App) {
		var redisConn *redis.RedisSession
		{
			redisConn = app.MustGetRedis()
			rand.Seed(time.Now().UnixNano())
			prefix := strconv.Itoa(rand.Int())
			redisConn.SetPrefix(prefix)
		}

		type fields struct {
			app *pkg.App
		}
		type args struct {
			redisConn  *redis.RedisSession
			source     string
			sourceVals map[string]interface{}
			target     string
			targetVals map[string]interface{}
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			result  map[string]int64
			wantErr bool
		}{
			{
				name: "empty call",
				fields: fields{
					app: app,
				},
				args: args{
					redisConn: redisConn,
					source:    "my_source",
					sourceVals: map[string]interface{}{
						"key1": 1,
						"key2": 2,
					},
					target: "my_target",
					targetVals: map[string]interface{}{
						"key2": 2,
						"key3": 3,
					},
				},
				result: map[string]int64{
					"key1": 1,
					"key2": 4,
					"key3": 3,
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c := &compactorService{
					app: tt.fields.app,
				}

				if err := redisConn.HashMultipleSet(tt.args.source, tt.args.sourceVals); err != nil {
					t.Errorf("redisConn.HashMultipleSet(tt.args.source, tt.args.sourceVals) error = %v, wantErr %v", err, tt.wantErr)
				}

				if err := redisConn.HashMultipleSet(tt.args.target, tt.args.targetVals); err != nil {
					t.Errorf("redisConn.HashMultipleSet(tt.args.source, tt.args.sourceVals) error = %v, wantErr %v", err, tt.wantErr)
				}

				if err := c.merge(tt.args.redisConn, tt.args.source, tt.args.target); (err != nil) != tt.wantErr {
					t.Errorf("compactorService.merge() error = %v, wantErr %v", err, tt.wantErr)
				}

				fns, err := redigo.Int64Map(redisConn.HashGetAll(tt.args.target))
				if err != nil {
					t.Errorf("redigo.Int64Map(redisConn.HashGetAll(tt.args.target)) error = %v", err)
				}

				if len(fns) != len(tt.result) {
					t.Errorf("len(fns) %d != len(tt.result) %d", len(fns), len(tt.result))
				}

				for key, val := range tt.result {
					if fns[key] != val {
						t.Errorf(" fns[key] != val | %d != %d", fns[key], val)
					}
				}

				if _, err := redisConn.Del(tt.args.source, tt.args.target); err != nil {
					t.Errorf("redisConn.Del(tt.args.source, tt.args.target) error = %v", err)
				}
			})
		}
	})
}

func Test_compactorService_withLock(t *testing.T) {
	withApp(func(app *pkg.App) {
		var redisConn *redis.RedisSession
		{
			redisConn = app.MustGetRedis()
			rand.Seed(time.Now().UnixNano())
			prefix := strconv.Itoa(rand.Int())
			redisConn.SetPrefix(prefix)
		}

		type fields struct {
			app *pkg.App
		}
		type args struct {
			redisConn *redis.RedisSession
			queueName string
			fn        func(srcMember string) error
		}

		queueName := "my_queue"
		checkQueueLength := func(queueName string, length int) {
			if members, err := redisConn.GetSetMembers(queueName); err != nil {
				t.Errorf("redisConn.GetSetMembers(%q) error = %v", queueName, err)
			} else if len(members) != length {
				t.Errorf("len(redisConn.GetSetMembers(%q) (%d) != %d", queueName, len(members), length)
			}
		}

		tests := []struct {
			name     string
			fields   fields
			args     args
			beforeOp func()
			afterOp  func()
			wantErr  bool
		}{
			{
				name: "non existing queue should not call the callback",
				fields: fields{
					app: app,
				},
				args: args{
					redisConn: redisConn,
					queueName: queueName,
					fn: func(member string) error {
						t.FailNow()
						return nil
					},
				},
				wantErr: true,
			},
			{
				name: "empty queue should not call the callback",
				fields: fields{
					app: app,
				},
				args: args{
					redisConn: redisConn,
					queueName: queueName,
					fn: func(member string) error {
						t.FailNow()
						return nil
					},
				},
				beforeOp: func() {
					// make sure we have the set with no members.
					if _, err := redisConn.AddSetMembers(queueName, "val"); err != nil {
						t.Errorf("redisConn.AddSetMembers(my_queue, val) error = %v", err)
					}
					if _, err := redisConn.RemoveSetMembers(queueName, "val"); err != nil {
						t.Errorf("redisConn.RemoveSetMembers(queueName, val) error = %v", err)
					}
				},
				afterOp: func() {
					queueName := queueName
					checkQueueLength(queueName, 0)
					queueName += "_processing"
					checkQueueLength(queueName, 0)
				},
				wantErr: true,
			},
			{
				name: "when callback fails, item should be moved back to origin queue",
				fields: fields{
					app: app,
				},
				args: args{
					redisConn: redisConn,
					queueName: queueName,
					fn: func(member string) error {
						return errors.New("text string")
					},
				},
				beforeOp: func() {
					// make sure we have the set with no members.
					if _, err := redisConn.AddSetMembers(queueName, "val"); err != nil {
						t.Errorf("redisConn.AddSetMembers(queueName, val) error = %v", err)
					}
				},
				afterOp: func() {
					queueName := queueName
					checkQueueLength(queueName, 1)
					queueName += "_processing"
					checkQueueLength(queueName, 0)

				},
				wantErr: true,
			},
			{
				name: "on successful op, both queues should be empty",
				fields: fields{
					app: app,
				},
				args: args{
					redisConn: redisConn,
					queueName: queueName,
					fn: func(member string) error {
						return nil
					},
				},
				beforeOp: func() {
					// make sure we have the set with no members.
					if _, err := redisConn.AddSetMembers(queueName, "val"); err != nil {
						t.Errorf("redisConn.AddSetMembers(queueName, val) error = %v", err)
					}
				},
				afterOp: func() {
					queueName := queueName
					checkQueueLength(queueName, 0)
					queueName += "_processing"
					checkQueueLength(queueName, 0)

				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c := &compactorService{
					app: tt.fields.app,
				}
				if tt.beforeOp != nil {
					tt.beforeOp()
				}
				if err := c.withLock(tt.args.redisConn, tt.args.queueName, tt.args.fn); (err != nil) != tt.wantErr {
					t.Errorf("compactorService.withLock() error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.afterOp != nil {
					tt.afterOp()
				}
				if _, err := redisConn.Del(queueName); err != nil {
					t.Errorf("redisConn.Del(%q) error = %v", queueName, err)
				}

				pQueueName := queueName + "_processing"
				if _, err := redisConn.Del(pQueueName); err != nil {
					t.Errorf("redisConn.Del(%q) error = %v", pQueueName, err)
				}
			})
		}
	})
}

func Test_generateHashKeys(t *testing.T) {
	type args struct {
		currentSuffix string
		hourlySuffix  string
		srcMember     string
	}
	tests := []struct {
		name       string
		args       args
		wantSource string
		wantTarget string
	}{
		{
			name: "empty keys",
			args: args{
				currentSuffix: "",
				hourlySuffix:  "",
				srcMember:     "",
			},
			wantSource: "hset:counter::",
			wantTarget: "hset:counter::",
		},
		{
			name: "with valid keys",
			args: args{
				currentSuffix: "currentSuffix",
				hourlySuffix:  "hourlySuffix",
				srcMember:     "member",
			},
			wantSource: "hset:counter:currentSuffix:member",
			wantTarget: "hset:counter:hourlySuffix:member",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSource, gotTarget := generateHashKeys(tt.args.currentSuffix, tt.args.hourlySuffix, tt.args.srcMember)
			if gotSource != tt.wantSource {
				t.Errorf("generateHashKeys() gotSource = %v, want %v", gotSource, tt.wantSource)
			}
			if gotTarget != tt.wantTarget {
				t.Errorf("generateHashKeys() gotTarget = %v, want %v", gotTarget, tt.wantTarget)
			}
		})
	}
}

func Test_generateSegmentSuffixes(t *testing.T) {
	type args struct {
		directionSuffix string
		tr              time.Time
	}
	tests := []struct {
		name        string
		args        args
		wantCurrent string
		wantHourly  string
	}{
		{
			name: "hour in the middle",
			args: args{
				directionSuffix: "src",
				tr:              time.Date(2017, time.March, 7, 06, 30, 0, 0, time.UTC),
			},
			wantCurrent: "src:1488868200",
			wantHourly:  "src:1488866400",
		},

		{
			name: "hour on the greater part",
			args: args{
				directionSuffix: "src",
				tr:              time.Date(2017, time.March, 7, 06, 45, 0, 0, time.UTC),
			},
			wantCurrent: "src:1488869100",
			wantHourly:  "src:1488866400",
		},
		{
			name: "5 min on the smaller part - should not be rounded to downward since this function accepts the values as is",
			args: args{
				directionSuffix: "src",
				tr:              time.Date(2017, time.March, 7, 06, 41, 0, 0, time.UTC),
			},
			wantCurrent: "src:1488868860",
			wantHourly:  "src:1488866400",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCurrent, gotHourly := generateSegmentSuffixes(tt.args.directionSuffix, tt.args.tr)
			if gotCurrent != tt.wantCurrent {
				t.Errorf("generateSegmentSuffixes() gotCurrent = %v, want %v", gotCurrent, tt.wantCurrent)
			}
			if gotHourly != tt.wantHourly {
				t.Errorf("generateSegmentSuffixes() gotHourly = %v, want %v", gotHourly, tt.wantHourly)
			}
		})
	}
}

func Test_compactorService_process(t *testing.T) {
	withApp(func(app *pkg.App) {
		var redisConn *redis.RedisSession
		{
			redisConn = app.MustGetRedis()
			rand.Seed(time.Now().UnixNano())
			prefix := strconv.Itoa(rand.Int())
			redisConn.SetPrefix(prefix)
		}

		type fields struct {
			app *pkg.App
		}
		type args struct {
			redisConn       *redis.RedisSession
			directionSuffix string
			tr              time.Time
		}
		tests := []struct {
			name     string
			fields   fields
			args     args
			wantErr  bool
			beforeOp func()
			afterOp  func()
		}{
			{
				name: "empty call",
				fields: fields{
					app: app,
				},
				args: args{
					redisConn:       redisConn,
					directionSuffix: "src",
					tr:              time.Now(), // not important
				},
				wantErr: true, // should get no item err
			},
			{
				name: "on multiple items, should only process one of them",
				fields: fields{
					app: app,
				},
				args: args{
					redisConn:       redisConn,
					directionSuffix: "src",
					tr:              time.Date(2017, time.March, 7, 06, 30, 0, 0, time.UTC),
				},
				wantErr: false,
				beforeOp: func() {
					queueName := "set:counter:src:1488868200"
					if _, err := redisConn.AddSetMembers(queueName, "val1", "val2"); err != nil {
						t.Errorf("redisConn.AddSetMembers(queueName, val) error = %v", err)
					}
				},
				afterOp: func() {
					queueName := "set:counter:src:1488868200"
					checkQueueLength(t, redisConn, queueName, 1)
					queueName += "_processing"
					checkQueueLength(t, redisConn, queueName, 0)
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c := &compactorService{
					app: tt.fields.app,
				}
				if tt.beforeOp != nil {
					tt.beforeOp()
				}
				if err := c.process(tt.args.redisConn, tt.args.directionSuffix, tt.args.tr); (err != nil) != tt.wantErr {
					t.Errorf("compactorService.process() error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.afterOp != nil {
					tt.afterOp()
				}
			})
		}
	})
}
