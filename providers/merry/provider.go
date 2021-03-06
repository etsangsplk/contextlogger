// © 2016-2017 Helix OpCo LLC. All rights reserved.
// Initial Author: Chris Williams

/*
This extracts merry Values into logger Fields, then passes along to the base logger
*/
package merry

import (
	"github.com/ansel1/merry"

	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers"

	"context"
)

type provider struct {
	nextProvider providers.LogProvider
}

func LogProvider(nextProvider providers.LogProvider) (providers.LogProvider, merry.Error) {
	if nextProvider == nil {
		return nil, merry.New("Merry log provider requires a base provider")
	}
	return provider{nextProvider}, nil
}

// Extract fields from merry error values if input was exactly one error
func (p provider) extractContext(ctx context.Context, args []interface{}, includeTrace bool) context.Context {
	if len(args) == 1 {
		if err, ok := args[0].(error); ok {
			fields := make(log.Fields)
			for key, val := range merry.Values(err) {
				if key, ok := key.(string); ok {
					switch key {
					case "stack", "message":
					// Merry built-ins; ignore
					case "user message":
						fields["userMessage"] = val
					default:
						fields[key] = val
					}
				}
			}
			// Call merry.Wrap to generate trace for non-merry errors; that trace will be to
			// here, not to where the error was generated, but better than nothing.
			wrapped := merry.Wrap(err)
			// Put stack into context, for providers that might need it (e.g. Rollbar)
			ctx = log.ContextWithStack(ctx, merry.Stack(wrapped))
			if includeTrace {
				// Use tilde to sort stacktrace last, which at least for logrus is more readable
				fields["~stackTrace"] = merry.Stacktrace(wrapped)
			}
			return log.ContextWithFields(ctx, fields)
		}
	}
	// No error found
	return ctx
}

// We always extract merry Values from an error, but only for Error level do we print a traceback
func (p provider) Error(ctx context.Context, report bool, args ...interface{}) {
	p.nextProvider.Error(p.extractContext(ctx, args, true), report, args...)
}

func (p provider) Warn(ctx context.Context, report bool, args ...interface{}) {
	p.nextProvider.Warn(p.extractContext(ctx, args, false), report, args...)
}

func (p provider) Info(ctx context.Context, report bool, args ...interface{}) {
	p.nextProvider.Info(p.extractContext(ctx, args, false), report, args...)
}

func (p provider) Debug(ctx context.Context, report bool, args ...interface{}) {
	p.nextProvider.Debug(p.extractContext(ctx, args, false), report, args...)
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
