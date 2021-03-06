// © 2016-2017 Helix OpCo LLC. All rights reserved.
// Initial Author: Chris Williams

/*
This adds a field to log reports showing where the logger was called from
*/
package reported_at

import (
	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers"

	"context"
	"errors"
	"fmt"
	"regexp"
	"runtime"
)

type provider struct {
	config       *Config
	nextProvider providers.LogProvider
}

type Config struct {
	// Ignore these stack frames for purposes of reportedAt
	IgnoreStackFrames *regexp.Regexp
}

var RecommendedConfig = Config{
	IgnoreStackFrames: regexp.MustCompile("myhelix/contextlogger"),
}

var alwaysIgnore = regexp.MustCompile("<autogenerated>")

func LogProvider(nextProvider providers.LogProvider, config Config) (providers.LogProvider, error) {
	if nextProvider == nil {
		return nil, errors.New("ReportedAt log provider requires a base provider")
	}
	return provider{&config, nextProvider}, nil
}

func (p provider) reportedAt(ctx context.Context) context.Context {
	pc := make([]uintptr, 50)
	runtime.Callers(1, pc)
	frameData := runtime.FuncForPC(pc[0])
	thisFile, _ := frameData.FileLine(pc[0])
	for _, frame := range pc {
		frameData := runtime.FuncForPC(frame)
		file, line := frameData.FileLine(frame)
		if file != thisFile && !alwaysIgnore.MatchString(file) &&
			(p.config.IgnoreStackFrames == nil || !p.config.IgnoreStackFrames.MatchString(file)) {
			return log.ContextWithFields(ctx, log.Fields{
				"reportedAt": fmt.Sprintf("%s:%d", file, line),
			})
		}
	}
	// Just in case we don't find anything
	return ctx
}

// We always extract merry Values from an error, but only for Error level do we print a traceback
func (p provider) Error(ctx context.Context, report bool, args ...interface{}) {
	p.nextProvider.Error(p.reportedAt(ctx), report, args...)
}

func (p provider) Warn(ctx context.Context, report bool, args ...interface{}) {
	p.nextProvider.Warn(p.reportedAt(ctx), report, args...)
}

func (p provider) Info(ctx context.Context, report bool, args ...interface{}) {
	p.nextProvider.Info(p.reportedAt(ctx), report, args...)
}

func (p provider) Debug(ctx context.Context, report bool, args ...interface{}) {
	p.nextProvider.Debug(p.reportedAt(ctx), report, args...)
}

func (p provider) Record(ctx context.Context, metrics map[string]interface{}) {
	p.nextProvider.Record(ctx, metrics)
}

func (p provider) RecordEvent(ctx context.Context, eventName string, metrics map[string]interface{}) {
	p.nextProvider.RecordEvent(ctx, eventName, metrics)
}

func (p provider) Wait() {
	p.nextProvider.Wait()
}
