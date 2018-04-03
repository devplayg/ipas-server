package classifier

import (
	"github.com/devplayg/ipas-server"
	log "github.com/sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
	"path/filepath"
)

type Classifier struct {
	engine *ipasserver.Engine
	watcher *fsnotify.Watcher
}


func NewClassifier(engine *ipasserver.Engine) *Classifier {
	return &Classifier{
		engine: engine,
	}
}

func (c *Classifier) Stop() error {
	if c.watcher!= nil {
		if err := c.watcher.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Classifier) Start() error {
	done := make(chan bool, 2)
	var err error
	c.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case event := <-c.watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Debugf("event: %s", event)
				}
			case err := <-c.watcher.Errors:
				log.Error(err)
			}
		}
	}()

	err = c.watcher.Add(filepath.Join(c.engine.ProcessDir, "data"))
	if err != nil {
		return err
	}

	return nil
}


func abc() {

}