//go:build decoupled || unit
// +build decoupled unit

package run_completion

import "context"

type MockKfpApi struct {
	runConfiguration string
	err              error
}

func (mka *MockKfpApi) GetRunConfiguration(_ context.Context, _ string) (string, error) {
	return mka.runConfiguration, mka.err
}

func (mka *MockKfpApi) reset() {
	mka.runConfiguration = ""
	mka.err = nil
}

func (mka *MockKfpApi) returnRunConfigurationForRun() string {
	mka.runConfiguration = randomString()
	mka.err = nil

	return mka.runConfiguration
}

func (mka *MockKfpApi) error(err error) {
	mka.runConfiguration = ""
	mka.err = err
}
