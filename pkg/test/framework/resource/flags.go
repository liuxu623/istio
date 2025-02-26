// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resource

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"istio.io/istio/pkg/test/framework/components/echo/echotypes"
	"istio.io/istio/pkg/test/framework/config"
	"istio.io/istio/pkg/test/framework/label"
)

var settingsFromCommandLine = DefaultSettings()

// SettingsFromCommandLine returns settings obtained from command-line flags. config.Parse must be called before
// calling this function.
func SettingsFromCommandLine(testID string) (*Settings, error) {
	if !config.Parsed() {
		panic("config.Parse must be called before this function")
	}

	s := settingsFromCommandLine.Clone()
	s.TestID = testID

	f, err := label.ParseSelector(s.SelectorString)
	if err != nil {
		return nil, err
	}
	s.Selector = f

	s.SkipMatcher, err = NewMatcher(s.SkipString)
	if err != nil {
		return nil, err
	}

	for _, wl := range s.skipWorkloadClasses {
		s.SkipWorkloadClasses.Insert(strings.Split(wl, ",")...)
	}
	if s.skipVM {
		s.SkipWorkloadClasses.Insert(echotypes.VM)
	}
	if s.skipTProxy {
		s.SkipWorkloadClasses.Insert(echotypes.TProxy)
	}
	if s.skipDelta {
		// TODO we may also want to trigger this if we have an old verion
		s.SkipWorkloadClasses.Insert(echotypes.Delta)
	}

	if err = validate(s); err != nil {
		return nil, err
	}

	return s, nil
}

// validate checks that user has not passed invalid flag combinations to test framework.
func validate(s *Settings) error {
	if s.FailOnDeprecation && s.NoCleanup {
		return fmt.Errorf("checking for deprecation occurs at cleanup level, thus flags -istio.test.nocleanup and" +
			" -istio.test.deprecation_failure must not be used at the same time")
	}

	if s.Revision != "" {
		if s.Revisions != nil {
			return fmt.Errorf("cannot use --istio.test.revision and --istio.test.revisions at the same time," +
				" --istio.test.revisions will take precedence and --istio.test.revision will be ignored")
		}
		// use Revision as the sole revision in RevVerMap
		s.Revisions = RevVerMap{
			s.Revision: "",
		}
	} else if s.Revisions != nil {
		// TODO(Monkeyanator) remove once existing jobs are migrated to use compatibility flag.
		s.Compatibility = true
	}

	if s.Revisions == nil && s.Compatibility {
		return fmt.Errorf("cannot use --istio.test.compatibility without setting --istio.test.revisions")
	}

	return nil
}

// init registers the command-line flags that we can exposed for "go test".
func init() {
	flag.StringVar(&settingsFromCommandLine.BaseDir, "istio.test.work_dir", os.TempDir(),
		"Local working directory for creating logs/temp files. If left empty, os.TempDir() is used.")

	var env string
	flag.StringVar(&env, "istio.test.env", "", "Deprecated. This flag does nothing")

	flag.BoolVar(&settingsFromCommandLine.NoCleanup, "istio.test.nocleanup", settingsFromCommandLine.NoCleanup,
		"Do not cleanup resources after test completion")

	flag.BoolVar(&settingsFromCommandLine.CIMode, "istio.test.ci", settingsFromCommandLine.CIMode,
		"Enable CI Mode. Additional logging and state dumping will be enabled.")

	flag.StringVar(&settingsFromCommandLine.SelectorString, "istio.test.select", settingsFromCommandLine.SelectorString,
		"Comma separated list of labels for selecting tests to run (e.g. 'foo,+bar-baz').")

	flag.Var(&settingsFromCommandLine.SkipString, "istio.test.skip",
		"Skip tests matching the regular expression. This follows the semantics of -test.run.")

	flag.Var(&settingsFromCommandLine.skipWorkloadClasses, "istio.test.skipWorkloads",
		"Skips deploying and using workloads of the given comma-separated classes (e.g. vm, proxyless, etc.)")

	flag.IntVar(&settingsFromCommandLine.Retries, "istio.test.retries", settingsFromCommandLine.Retries,
		"Number of times to retry tests")

	flag.BoolVar(&settingsFromCommandLine.StableNamespaces, "istio.test.stableNamespaces", settingsFromCommandLine.StableNamespaces,
		"If set, will use consistent namespace rather than randomly generated. Useful with nocleanup to develop tests.")

	flag.BoolVar(&settingsFromCommandLine.FailOnDeprecation, "istio.test.deprecation_failure", settingsFromCommandLine.FailOnDeprecation,
		"Make tests fail if any usage of deprecated stuff (e.g. Envoy flags) is detected.")

	flag.StringVar(&settingsFromCommandLine.Revision, "istio.test.revision", settingsFromCommandLine.Revision,
		"If set to XXX, overwrite the default namespace label (istio-injection=enabled) with istio.io/rev=XXX.")

	flag.BoolVar(&settingsFromCommandLine.skipVM, "istio.test.skipVM", settingsFromCommandLine.skipVM,
		"Skip VM related parts in all tests.")

	flag.BoolVar(&settingsFromCommandLine.skipDelta, "istio.test.skipDelta", settingsFromCommandLine.skipDelta,
		"Skip Delta XDS related parts in all tests.")

	flag.BoolVar(&settingsFromCommandLine.skipTProxy, "istio.test.skipTProxy", settingsFromCommandLine.skipTProxy,
		"Skip TProxy related parts in all tests.")

	flag.BoolVar(&settingsFromCommandLine.Compatibility, "istio.test.compatibility", settingsFromCommandLine.Compatibility,
		"Transparently deploy echo instances pointing to each revision set in `Revisions`")

	flag.Var(&settingsFromCommandLine.Revisions, "istio.test.revisions", "Istio CP revisions available to the test framework and their corresponding versions.")
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return fmt.Sprint([]string(*i))
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
