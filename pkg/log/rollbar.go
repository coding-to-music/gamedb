package log

import (
	"io"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/ldflags"
	"github.com/rollbar/rollbar-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newRollbarCore(sync bool) *rollbarCore {

	rollbar.SetToken(config.C.RollbarSecret)

	core := &rollbarCore{
		encoder: zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()),
		output:  zapcore.AddSync(io.Discard),
	}

	if sync {
		core.client = rollbar.NewSync(config.C.RollbarSecret, config.C.Environment, ldflags.CommitCount, "", "")
	} else {
		core.client = rollbar.NewAsync(config.C.RollbarSecret, config.C.Environment, ldflags.CommitCount, "", "")
	}

	return core
}

type rollbarCore struct {
	client *rollbar.Client

	encoder zapcore.Encoder
	output  zapcore.WriteSyncer
}

func (c *rollbarCore) clone() *rollbarCore {

	return &rollbarCore{
		client: c.client,

		encoder: c.encoder.Clone(),
		output:  zapcore.AddSync(io.Discard),
	}
}

func (c *rollbarCore) Enabled(level zapcore.Level) bool {
	return level.Enabled(level)
}

func (c *rollbarCore) With(fields []zapcore.Field) zapcore.Core {

	clone := c.clone()
	for k := range fields {
		fields[k].AddTo(clone.encoder)
	}
	return clone
}

func (c *rollbarCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {

	if c.Enabled(entry.Level) {
		return checkedEntry.AddCore(entry, c)
	}

	return checkedEntry
}

func (c *rollbarCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {

	buf, err := c.encoder.EncodeEntry(entry, fields)
	if err != nil {
		return err
	}

	switch entry.Level {
	case zapcore.DebugLevel:
		c.client.Message(rollbar.DEBUG, buf.String())
	case zapcore.InfoLevel:
		c.client.Message(rollbar.INFO, buf.String())
	case zapcore.WarnLevel:
		c.client.Message(rollbar.WARN, buf.String())
	case zapcore.ErrorLevel:
		c.client.Message(rollbar.ERR, buf.String())
	case zapcore.DPanicLevel:
		c.client.Message(rollbar.CRIT, buf.String())
	case zapcore.PanicLevel:
		c.client.Message(rollbar.CRIT, buf.String())
	case zapcore.FatalLevel:
		c.client.Message(rollbar.CRIT, buf.String())
	}

	return nil
}

func (c *rollbarCore) Sync() error {

	c.client.Wait()

	return c.output.Sync()
}
