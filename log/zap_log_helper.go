package log

import (
	"bytes"
	"context"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Option func(c *ZapHelperBuilder)

func AddHook(hook Hook) Option {
	return func(c *ZapHelperBuilder) {
		c.hooks.Add(hook)
	}
}

type ZapHelperBuilder struct {
	hooks LevelHooks
}

func NewZapHelper(logger *zap.Logger, opt ...Option) Helper {
	var cfg = &ZapHelperBuilder{hooks: make(LevelHooks, 6)}
	for _, v := range opt {
		v(cfg)
	}

	c := zapLoggerHelper{
		logger: logger,
		pool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
		Hooks: cfg.hooks,
	}
	c.initLogLevel()
	return &c
}

type zapLoggerHelper struct {
	logger *zap.Logger
	Hooks  LevelHooks
	level  zapcore.Level
	pool   *sync.Pool
}

func (c *zapLoggerHelper) WithContext(ctx context.Context) FieldLogger {
	return &zapFieldLogger{
		ctx:    ctx,
		helper: c,
		entry:  newZapLogEntry(ctx),
	}
}

// pollBuff 返回一个可用的 buff 和一个释放 bytes.Buffer 的函数
func (c *zapLoggerHelper) pollBuff() (buff *bytes.Buffer, release func()) {
	buff, _ = (c.pool.Get()).(*bytes.Buffer)
	buff.Reset()
	return buff, func() {
		buff.Reset()
		c.pool.Put(buff)
	}
}

func (c *zapLoggerHelper) levelEnabled(level zapcore.Level) bool {
	return c.level <= level
}

// 初始化 level 值
func (c *zapLoggerHelper) initLogLevel() {
	var levels = []zapcore.Level{
		zapcore.DebugLevel,
		zapcore.InfoLevel,
		zapcore.WarnLevel,
		zapcore.ErrorLevel,
		zapcore.FatalLevel,
	}

	for _, v := range levels {
		if c.logger.Core().Enabled(v) {
			c.level = v
			break
		}
	}
}

func newZapLogEntry(ctx context.Context) Entry {
	return &zapLoggerEntry{ctx: ctx, fields: make([]zap.Field, 0, 3)}
}
