//go:build !windows

package process

type jobObject struct{}

func newJobObject() *jobObject {
	return &jobObject{}
}

func (j *jobObject) assign(_ any) {}
