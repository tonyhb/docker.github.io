package changelist

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/Sirupsen/logrus"
)

// FileChangelist stores all the changes as files
type FileChangelist struct {
	dir string
}

// NewFileChangelist is a convenience method for returning FileChangeLists
func NewFileChangelist(dir string) (*FileChangelist, error) {
	logrus.Debug("Making dir path: ", dir)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return nil, err
	}
	return &FileChangelist{dir: dir}, nil
}

// List returns a list of sorted changes
func (cl FileChangelist) List() []Change {
	var changes []Change
	dir, err := os.Open(cl.dir)
	if err != nil {
		return changes
	}
	defer dir.Close()
	fileInfos, err := dir.Readdir(0)
	if err != nil {
		return changes
	}
	sort.Sort(fileChanges(fileInfos))
	for _, f := range fileInfos {
		if f.IsDir() {
			continue
		}
		raw, err := ioutil.ReadFile(path.Join(cl.dir, f.Name()))
		if err != nil {
			// TODO(david): How should we handle this?
			logrus.Warn(err.Error())
			continue
		}
		c := &TufChange{}
		err = json.Unmarshal(raw, c)
		if err != nil {
			// TODO(david): How should we handle this?
			logrus.Warn(err.Error())
			continue
		}
		changes = append(changes, c)
	}
	return changes
}

// Add adds a change to the file change list
func (cl FileChangelist) Add(c Change) error {
	cJSON, err := json.Marshal(c)
	if err != nil {
		return err
	}
	filename := fmt.Sprintf("%020d_%s.change", time.Now().UnixNano(), uuid.New())
	return ioutil.WriteFile(path.Join(cl.dir, filename), cJSON, 0644)
}

// Clear clears the change list
func (cl FileChangelist) Clear(archive string) error {
	dir, err := os.Open(cl.dir)
	if err != nil {
		return err
	}
	defer dir.Close()
	files, err := dir.Readdir(0)
	if err != nil {
		return err
	}
	for _, f := range files {
		os.Remove(path.Join(cl.dir, f.Name()))
	}
	return nil
}

// Close is a no-op
func (cl FileChangelist) Close() error {
	// Nothing to do here
	return nil
}

type fileChanges []os.FileInfo

// Len returns the length of a file change list
func (cs fileChanges) Len() int {
	return len(cs)
}

// Less compares the names of two different file changes
func (cs fileChanges) Less(i, j int) bool {
	return cs[i].Name() < cs[j].Name()
}

// Swap swaps the position of two file changes
func (cs fileChanges) Swap(i, j int) {
	tmp := cs[i]
	cs[i] = cs[j]
	cs[j] = tmp
}
