package subscribe

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rancher/norman/v2/pkg/types"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout:  60 * time.Second,
	EnableCompression: true,
}

type Subscribe struct {
	Stop            bool   `json:"stop,omitempty"`
	ResourceType    string `json:"resourceType,omitempty"`
	ResourceVersion string `json:"resourceVersion,omitempty"`
}

func Handler(apiOp *types.APIRequest) (types.APIObject, error) {
	err := handler(apiOp)
	if err != nil {
		logrus.Errorf("Error during subscribe %v", err)
	}
	return types.APIObject{}, err
}

func handler(apiOp *types.APIRequest) error {
	c, err := upgrader.Upgrade(apiOp.Response, apiOp.Request, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	watches := NewWatchSession(apiOp)
	defer watches.Close()

	events := watches.Watch(c)
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return nil
			}
			if err := writeData(apiOp, c, event); err != nil {
				return err
			}
		case <-t.C:
			if err := writeData(apiOp, c, types.APIEvent{Name: "ping"}); err != nil {
				return err
			}
		}
	}
}

func writeData(apiOp *types.APIRequest, c *websocket.Conn, event types.APIEvent) error {
	event = MarshallObject(apiOp, event)
	if event.Error != nil {
		event.Name = "resource.error"
		event.Data = map[string]interface{}{
			"error": event.Error.Error(),
		}
	}

	messageWriter, err := c.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	defer messageWriter.Close()

	return json.NewEncoder(messageWriter).Encode(event)
}
