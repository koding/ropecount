package compactor

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/koding/redis"
	"github.com/ropelive/count/pkg"
	"github.com/ropelive/count/pkg/mongodb"
)

func withApp(fn func(app *pkg.App)) {
	name := "compator_test"
	app := pkg.NewApp(name, pkg.ConfigureHTTP(), pkg.ConfigureRedis(), pkg.ConfigureMongo())
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
			source    string
			fns       map[string]int64
		}
		source := "hset:counter:src:1488868203:cihangir"
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
					source:    source,
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
					source:    source,
					fns:       map[string]int64{},
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c := &compactorService{
					app: tt.fields.app,
				}
				if err := c.incrementMapValues(tt.args.source, tt.args.fns); (err != nil) != tt.wantErr {
					t.Errorf("compactorService.incrementMapValues() error = %v, wantErr %v", err, tt.wantErr)
				}

				if tt.wantErr {
					return
				}

				parsedKeys := pkg.ParseKeyName(tt.args.source)
				fns, err := mongodb.GetCompaction(app.MustGetMongo(), parsedKeys.Name, parsedKeys.Direction, parsedKeys.Segment)
				if err != nil {
					t.Errorf("mongodb.GetCompaction() error = %v", err)
				}

				if len(fns) != len(tt.args.fns) {
					t.Errorf("len(fns) %d != len(tt.args.fns) %d", len(fns), len(tt.args.fns))
				}

				for key, val := range tt.args.fns {
					if fns[key] != val {
						t.Errorf(" fns[key] != val | %d != %d", fns[key], val)
					}
				}

				if _, err := redisConn.Del(tt.args.source); err != nil {
					t.Errorf("redisConn.Del(tt.args.source) error = %v", err)
				}

				if err := mongodb.DeleteCompaction(app.MustGetMongo(), parsedKeys.Name, parsedKeys.Direction, parsedKeys.Segment); err != nil {
					t.Errorf("mongodb.DeleteCompaction() error = %v", err)
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
		}
		source := "hset:counter:src:1488868201:cihangir"
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
					source:    source,
					sourceVals: map[string]interface{}{
						"key1": 1,
						"key2": 2,
					},
				},
				result: map[string]int64{
					"key1": 1,
					"key2": 2,
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

				if err := c.merge(tt.args.redisConn, tt.args.source); (err != nil) != tt.wantErr {
					t.Errorf("compactorService.merge() error = %v, wantErr %v", err, tt.wantErr)
				}

				parsedKeys := pkg.ParseKeyName(tt.args.source)
				fns, err := mongodb.GetCompaction(app.MustGetMongo(), parsedKeys.Name, parsedKeys.Direction, parsedKeys.Segment)
				if err != nil {
					t.Errorf("mongodb.GetCompaction() error = %v", err)
				}

				if len(fns) != len(tt.result) {
					t.Errorf("len(fns) %d != len(tt.result) %d", len(fns), len(tt.result))
				}

				for key, val := range tt.result {
					if fns[key] != val {
						t.Errorf(" fns[key] != val | %d != %d", fns[key], val)
					}
				}

				if _, err := redisConn.Del(tt.args.source); err != nil {
					t.Errorf("redisConn.Del(tt.args.source) error = %v", err)
				}

				if err := mongodb.DeleteCompaction(app.MustGetMongo(), parsedKeys.Name, parsedKeys.Direction, parsedKeys.Segment); err != nil {
					t.Errorf("mongodb.DeleteCompaction() error = %v", err)
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
						t.Errorf("redisConn.AddSetMembers(queueName, val) error = %v", err)
					}
					if _, err := redisConn.RemoveSetMembers(queueName, "val"); err != nil {
						t.Errorf("redisConn.RemoveSetMembers(queueName, val) error = %v", err)
					}
				},
				afterOp: func() {
					queueName := queueName
					checkQueueLength(t, redisConn, queueName, 0)
					queueName += "_processing"
					checkQueueLength(t, redisConn, queueName, 0)
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
					checkQueueLength(t, redisConn, queueName, 1)
					queueName += "_processing"
					checkQueueLength(t, redisConn, queueName, 0)

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
					checkQueueLength(t, redisConn, queueName, 0)
					queueName += "_processing"
					checkQueueLength(t, redisConn, queueName, 0)
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

func Test_compactorService_process(t *testing.T) {
	withApp(func(app *pkg.App) {
		var redisConn *redis.RedisSession
		{
			redisConn = app.MustGetRedis()
			rand.Seed(time.Now().UnixNano())
			prefix := strconv.Itoa(rand.Int())
			redisConn.SetPrefix(prefix)
		}

		tr := time.Date(2017, time.March, 7, 06, 30, 0, 0, time.UTC)
		keyNames := pkg.GenerateKeyNames(tr)
		type fields struct {
			app *pkg.App
		}
		type args struct {
			redisConn *redis.RedisSession
			keyNames  pkg.KeyNames
			tr        time.Time
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
					redisConn: redisConn,
					keyNames:  keyNames.Src,
					tr:        tr, // not important
				},
				wantErr: true, // should get no item err
			},
			{
				name: "on multiple items, should only process one of them",
				fields: fields{
					app: app,
				},
				args: args{
					redisConn: redisConn,
					keyNames:  keyNames.Src,
					tr:        tr,
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
				if err := c.process(tt.args.redisConn, tt.args.keyNames, tt.args.tr); (err != nil) != tt.wantErr {
					t.Errorf("compactorService.process() error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.afterOp != nil {
					tt.afterOp()
				}
			})
		}
	})
}

func checkQueueLength(t *testing.T, redisConn *redis.RedisSession, queueName string, length int) {
	if members, err := redisConn.GetSetMembers(queueName); err != nil {
		t.Errorf("redisConn.GetSetMembers(%q) error = %v", queueName, err)
	} else if len(members) != length {
		t.Errorf("len(redisConn.GetSetMembers(%q) (%d) != %d", queueName, len(members), length)
	}
}

func Test_compactorService_Process(t *testing.T) {
	withApp(func(app *pkg.App) {
		var redisConn *redis.RedisSession
		{
			redisConn = app.MustGetRedis()
			rand.Seed(time.Now().UnixNano())
			prefix := strconv.Itoa(rand.Int())
			redisConn.SetPrefix(prefix)
		}
		timeoutDur := time.Millisecond * 3
		timeoutCtx, cancel := context.WithTimeout(context.Background(), timeoutDur)
		defer cancel()
		type fields struct {
			app *pkg.App
		}
		type args struct {
			ctx context.Context
			p   ProcessRequest
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
					ctx: context.Background(),
					p: ProcessRequest{
						StartAt: time.Date(2017, time.March, 7, 06, 30, 0, 0, time.UTC),
					},
				},
				wantErr:  false,
				beforeOp: func() {},
				afterOp:  func() {},
			},
			{
				name: "timedout call",
				fields: fields{
					app: app,
				},
				args: args{
					ctx: timeoutCtx,
					p: ProcessRequest{
						StartAt: time.Date(2017, time.March, 7, 06, 30, 0, 0, time.UTC),
					},
				},
				wantErr: true,
				beforeOp: func() {
					time.Sleep(timeoutDur)
				},
				afterOp: func() {},
			},
			{
				name: "invalid redis key type in source direction",
				fields: fields{
					app: app,
				},
				args: args{
					ctx: context.Background(),
					p: ProcessRequest{
						StartAt: time.Date(2017, time.March, 7, 06, 30, 0, 0, time.UTC),
					},
				},
				wantErr: true,
				beforeOp: func() {
					// change the key type in redis
					srcQueueName := "set:counter:src:1488867600"
					if _, err := redisConn.Del(srcQueueName); err != nil {
						t.Errorf("redisConn.Del(%q) error = %v", srcQueueName, err)
					}
					if err := redisConn.Set(srcQueueName, "value"); err != nil {
						t.Errorf("redisConn.Set(%q) error = %v", srcQueueName, err)
					}
				},
				afterOp: func() {
					srcQueueName := "set:counter:src:1488867600"
					if _, err := redisConn.Del(srcQueueName); err != nil {
						t.Errorf("redisConn.Del(%q) error = %v", srcQueueName, err)
					}
				},
			},
			{
				name: "invalid redis key type in target direction",
				fields: fields{
					app: app,
				},
				args: args{
					ctx: context.Background(),
					p: ProcessRequest{
						StartAt: time.Date(2017, time.March, 7, 06, 30, 0, 0, time.UTC),
					},
				},
				wantErr: true,
				beforeOp: func() {
					// change the key type in redis
					dstQueueName := "set:counter:dst:1488867600"
					if _, err := redisConn.Del(dstQueueName); err != nil {
						t.Errorf("redisConn.Del(%q) error = %v", dstQueueName, err)
					}
					if err := redisConn.Set(dstQueueName, "value"); err != nil {
						t.Errorf("redisConn.Set(%q) error = %v", dstQueueName, err)
					}
				},
				afterOp: func() {
					dstQueueName := "set:counter:dst:1488867600"
					if _, err := redisConn.Del(dstQueueName); err != nil {
						t.Errorf("redisConn.Del(%q) error = %v", dstQueueName, err)
					}
				},
			},
			{
				name: "multiple items should be consumed properly",
				fields: fields{
					app: app,
				},
				args: args{
					ctx: context.Background(),
					p: ProcessRequest{
						StartAt: time.Date(2017, time.March, 7, 06, 30, 0, 0, time.UTC),
					},
				},
				wantErr: false,
				beforeOp: func() {
					queueName := "set:counter:dst:1488867600"
					if _, err := redisConn.AddSetMembers(queueName, "val1", "val2"); err != nil {
						t.Errorf("redisConn.AddSetMembers(%q) error = %v", queueName, err)
					}
				},
				afterOp: func() {
					queueName := "set:counter:dst:1488867600"
					if _, err := redisConn.Del(queueName); err != nil {
						t.Errorf("redisConn.Del(%q) error = %v", queueName, err)
					}
					checkQueueLength(t, redisConn, queueName, 0)
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
				if err := c.Process(tt.args.ctx, tt.args.p); (err != nil) != tt.wantErr {
					t.Errorf("compactorService.Process() error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.afterOp != nil {
					tt.afterOp()
				}
			})
		}
	})
}
