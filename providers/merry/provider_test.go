// © 2016-2017 Helix OpCo LLC. All rights reserved.
// Initial Author: Chris Williams

package merry

import (
	"bytes"
	"context"
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/ansel1/merry"
	"github.com/myhelix/contextlogger/providers"
	cl_logrus "github.com/myhelix/contextlogger/providers/logrus"
	. "github.com/onsi/gomega"
	"testing"
)

var output *bytes.Buffer
var testProvider providers.LogProvider

func setup(t *testing.T) {
	RegisterTestingT(t)

	output = new(bytes.Buffer)
	outputProvider, err := cl_logrus.LogProvider(nil, cl_logrus.Config{
		output,
		"debug",
		&logrus.TextFormatter{
			DisableColors:   true,
			TimestampFormat: "sometime", // Omit timestamp to make output predictable
		},
	})
	Expect(err).To(BeNil())
	testProvider, err = LogProvider(outputProvider)
	Expect(err).To(BeNil())
}

func TestValueExtraction(t *testing.T) {
	setup(t)

	testProvider.Info(context.Background(), false, merry.New("it broke").WithValue("how", "badly"))
	Expect(output.String()).To(MatchRegexp(`time=sometime level=info msg="it broke" how=badly`))
}

func TestUserMessage(t *testing.T) {
	setup(t)

	testProvider.Info(context.Background(), false, merry.New("it broke").WithUserMessage("all good"))
	Expect(output.String()).To(MatchRegexp(`time=sometime level=info msg="it broke" userMessage="all good"`))
}

func TestMerryTraceback(t *testing.T) {
	setup(t)

	err := merry.New("it broke").WithValue("how", "badly")

	testProvider.Error(context.Background(), false, err)
	Expect(output.String()).To(MatchRegexp(`time=sometime level=error msg="it broke" how=badly ~stackTrace=".*myhelix/contextlogger/providers/merry/provider_test.go:.*
.*TestMerryTraceback: err := merry.New\("it broke"\).WithValue\("how", "badly"\).*
.*
.*
.*
.*
.*"`))
}

func TestErrorTraceback(t *testing.T) {
	setup(t)

	err := errors.New("it broke")

	testProvider.Error(context.Background(), false, err)
	Expect(output.String()).To(MatchRegexp(`time=sometime level=error msg="it broke" ~stackTrace=".*myhelix/contextlogger/providers/merry/provider.go:.*
.*provider.extractContext: wrapped := merry.Wrap\(err\).*
.*
.*
.*
.*myhelix/contextlogger/providers/merry/provider_test.go:.*
.*TestErrorTraceback: testProvider.Error\(context.Background\(\), false, err\).*`))
}

/* If you mix an error with other crap, you lose the good metadata */
func TestErrorMisc(t *testing.T) {
	setup(t)

	err := merry.New("it broke").WithValue("how", "badly")

	testProvider.Error(context.Background(), false, err, "foo", errors.New("bar"))
	Expect(output.String()).To(MatchRegexp(`time=sometime level=error msg="it brokefoobar"`))
}
