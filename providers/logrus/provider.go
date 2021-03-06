// © 2016-2017 Helix OpCo LLC. All rights reserved.
// Initial Author: Chris Williams

/*
This package provides a concrete implementation of LogProvider using Logrus
*/
package logrus

import (
	"github.com/Sirupsen/logrus"

	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers"

	"context"
	"io"
	"time"
)

type provider struct {
	*logrus.Entry
	nextProvider providers.LogProvider
}

type Config struct {
	Output    io.Writer
	Level     string
	Formatter *logrus.TextFormatter
}

var RecommendedFormatter = &logrus.TextFormatter{
	FullTimestamp:   true,
	DisableColors:   true,
	TimestampFormat: time.RFC3339Nano,
}

func LogProvider(nextProvider providers.LogProvider, config Config) (l providers.LogProvider, err error) {
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		return
	}

	l = provider{logrus.NewEntry(&logrus.Logger{
		Out:       config.Output,
		Formatter: config.Formatter,
		Hooks:     make(logrus.LevelHooks),
		Level:     level,
	}), nextProvider}
	return
}

func (p provider) entryFor(ctx context.Context) *logrus.Entry {
	return p.Entry.WithFields(logrus.Fields(log.FieldsFromContext(ctx)))
}

func (p provider) Error(ctx context.Context, report bool, args ...interface{}) {
	p.entryFor(ctx).Error(args...)
	if p.nextProvider != nil {
		p.nextProvider.Error(ctx, report, args...)
	}
}

func (p provider) Warn(ctx context.Context, report bool, args ...interface{}) {
	p.entryFor(ctx).Warn(args...)
	if p.nextProvider != nil {
		p.nextProvider.Warn(ctx, report, args...)
	}
}

func (p provider) Info(ctx context.Context, report bool, args ...interface{}) {
	p.entryFor(ctx).Info(args...)
	if p.nextProvider != nil {
		p.nextProvider.Info(ctx, report, args...)
	}
}

func (p provider) Debug(ctx context.Context, report bool, args ...interface{}) {
	p.entryFor(ctx).Debug(args...)
	if p.nextProvider != nil {
		p.nextProvider.Debug(ctx, report, args...)
	}
}

func (p provider) Record(ctx context.Context, metrics map[string]interface{}) {
	p.WithFields(metrics).Info("Reporting metrics")
	if p.nextProvider != nil {
		p.nextProvider.Record(ctx, metrics)
	}
}

func (p provider) RecordEvent(ctx context.Context, eventName string, metrics map[string]interface{}) {
	p.WithField("eventName", eventName).WithFields(metrics).Info("Reporting metrics")
	if p.nextProvider != nil {
		p.nextProvider.RecordEvent(ctx, eventName, metrics)
	}
}

func (p provider) Wait() {
	if p.nextProvider != nil {
		p.nextProvider.Wait()
	}
}
