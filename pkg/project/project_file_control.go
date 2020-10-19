package project

import (
	"fmt"
	"io/ioutil"
	"path"
)

// FileControl contains wakatime project file parameters.
type FileControl struct {
	Path    string
	Project string
}

func (fc FileControl) Write() error {
	err := ioutil.WriteFile(path.Join(fc.Path, defaultProjectFile), []byte(fc.Project+"\n"), 0600)
	if err != nil {
		return fmt.Errorf("failed to save wakatime project file: %s", err)
	}

	return nil
}
