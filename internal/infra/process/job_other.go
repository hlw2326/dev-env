//go:build !windows

package process

type jobObject struct{}

func newJob() *jobObject {
	return &jobObject{}
}

func (j *jobObject) assign(_ any) {}
