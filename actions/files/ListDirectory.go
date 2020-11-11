// Package files is generated by actiongenerator tooling
// Make sure to insert real Description here
package files

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/perbol/workflow/payload"
	"github.com/perbol/workflow/property"
	"github.com/perbol/workflow/register"
)

// ListDirectory is used to list all FILES in a given path
type ListDirectory struct {
	// Cfg is values needed to properly run the Handle func
	Cfg        *property.Configuration `json:"configs" yaml:"configs"`
	Name       string                  `json:"action" yaml:"action_name"`
	path       string
	buffertime int64
	found      map[string]int64
	sync.Mutex `json:"-" yaml:"-"`

	subscriptionless bool
}

var (
	// DefaultBufferTime is how long in seconds a file should be rememberd
	DefaultBufferTime int64 = 3600
)

func init() {
	register.Register("ListDirectory", NewListDirectoryAction())
}

// NewListDirectoryAction generates a new ListDirectory action
func NewListDirectoryAction() *ListDirectory {
	act := &ListDirectory{
		Cfg: &property.Configuration{
			Properties: make([]*property.Property, 0),
		},
		Name:             "ListDirectory",
		found:            make(map[string]int64),
		subscriptionless: true,
	}

	act.Cfg.AddProperty("path", "the path to search for", true)
	act.Cfg.AddProperty("buffertime", "the time in seconds for how long a found file should be rememberd and not relisted", false)

	return act
}

// GetActionName is used to retrun a unqiue string name
func (a *ListDirectory) GetActionName() string {
	return a.Name
}

// Handle is used to list all files in a direcory
func (a *ListDirectory) Handle(p payload.Payload) ([]payload.Payload, error) {
	files, err := ioutil.ReadDir(a.path)
	if err != nil {
		return nil, err
	}
	a.Lock()
	for k, v := range a.found {
		if time.Now().Unix()-v > a.buffertime {
			delete(a.found, k) // If the item is older than given time setting, delete it from buffer
		}
	}
	a.Unlock()
	foundfiles := make([]payload.Payload, 0)

	for _, f := range files {
		if f.IsDir() == false {
			file := filepath.Base(f.Name())
			var filepath string
			if strings.HasSuffix(a.path, "/") {
				filepath = fmt.Sprintf("%s%s", a.path, file)
			} else {
				filepath = fmt.Sprintf("%s/%s", a.path, file)
			}
			if _, ok := a.found[filepath]; !ok {
				foundfiles = append(foundfiles, payload.BasePayload{
					Payload: []byte(filepath),
					Source:  "ListDirectory",
				})
				a.found[filepath] = time.Now().Unix()
			}
		}
	}
	/* Should a Buffer of found files be kept? */
	return foundfiles, nil
}

// ValidateConfiguration is used to see that all needed configurations are assigned before starting
func (a *ListDirectory) ValidateConfiguration() (bool, []string) {
	// Check if Cfgs are there as needed
	// Needs a Directory to monitor
	pathProp := a.Cfg.GetProperty("path")
	missing := make([]string, 0)
	if pathProp == nil {
		missing = append(missing, "path")
		return false, missing
	}
	bufferProp := a.Cfg.GetProperty("buffertime")
	if bufferProp.Value == nil {
		a.buffertime = DefaultBufferTime
	} else {
		value, err := bufferProp.Int64()
		if err != nil {
			missing = append(missing, "buffertime")
			return false, missing
		}
		a.buffertime = value
	}

	a.path = pathProp.String()
	return true, nil
}

// GetConfiguration will return the CFG for the action
func (a *ListDirectory) GetConfiguration() *property.Configuration {
	return a.Cfg
}

// Subscriptionless is used to send out true
func (a *ListDirectory) Subscriptionless() bool {
	return a.subscriptionless
}