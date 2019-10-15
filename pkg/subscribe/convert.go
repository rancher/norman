package subscribe

import (
	"io"

	"github.com/rancher/norman/pkg/api/writer"
	"github.com/rancher/norman/pkg/types"
)

type Converter struct {
	writer.EncodingResponseWriter
	apiOp *types.APIRequest
	obj   interface{}
}

func MarshallObject(apiOp *types.APIRequest, event types.APIEvent) types.APIEvent {
	if event.Error != nil {
		return event
	}

	if event.Object.IsNil() {
		return event
	}

	data, err := newConverter(apiOp).ToAPIObject(event.Object)
	if err != nil {
		event.Error = err
		return event
	}

	event.Data = data.Raw()
	return event
}

func newConverter(apiOp *types.APIRequest) *Converter {
	c := &Converter{
		apiOp: apiOp,
	}
	c.EncodingResponseWriter = writer.EncodingResponseWriter{
		ContentType: "application/json",
		Encoder:     c.Encoder,
	}
	return c
}

func (c *Converter) ToAPIObject(data interface{}) (types.APIObject, error) {
	c.obj = nil
	if err := c.VersionBody(c.apiOp, nil, data); err != nil {
		return types.APIObject{}, err
	}
	return types.ToAPI(c.obj), nil
}

func (c *Converter) Encoder(w io.Writer, obj interface{}) error {
	c.obj = obj
	return nil
}
