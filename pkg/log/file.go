package log

import (
	"io/ioutil"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newFileCore(bin string) zapcore.Core {

	err := os.MkdirAll("/tmp/gamedb/", os.ModePerm)
	if err != nil {
		panic(err)
	}

	f, err := os.OpenFile("/tmp/gamedb/"+bin+".log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	return fileCore{
		file:         f,
		levelEnabler: zap.NewAtomicLevelAt(zapcore.DebugLevel),
		encoder:      zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		output:       zapcore.AddSync(ioutil.Discard),
	}
}

type fileCore struct {
	file         *os.File
	levelEnabler zapcore.LevelEnabler
	encoder      zapcore.Encoder
	output       zapcore.WriteSyncer
}

func (g *fileCore) clone() *fileCore {

	return &fileCore{
		file:         g.file,
		levelEnabler: g.levelEnabler,
		encoder:      g.encoder.Clone(),
		output:       g.output,
	}
}

func (g fileCore) Enabled(level zapcore.Level) bool {
	return level.Enabled(level)
}

func (g fileCore) With(fields []zapcore.Field) zapcore.Core {

	clone := g.clone()
	for k := range fields {
		fields[k].AddTo(clone.encoder)
	}
	return clone
}

func (g fileCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {

	if g.Enabled(entry.Level) {
		return checkedEntry.AddCore(entry, g)
	}
	return checkedEntry
}

func (g fileCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {

	buf, err := g.encoder.EncodeEntry(entry, fields)
	if err != nil {
		return err
	}

	_, err = g.file.WriteString(buf.String())
	return err
}

func (g fileCore) Sync() error {
	return g.output.Sync()
}
