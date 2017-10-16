package client

const (
	DIAL_TYPE = "dial"
)

type Dial struct {
	Resource `yaml:"-"`

	Address string `json:"address,omitempty" yaml:"address,omitempty"`

	ClusterId string `json:"clusterId,omitempty" yaml:"cluster_id,omitempty"`

	Protocol string `json:"protocol,omitempty" yaml:"protocol,omitempty"`
}

type DialCollection struct {
	Collection
	Data   []Dial `json:"data,omitempty"`
	client *DialClient
}

type DialClient struct {
	rancherClient *RancherClient
}

type DialOperations interface {
	List(opts *ListOpts) (*DialCollection, error)
	Create(opts *Dial) (*Dial, error)
	Update(existing *Dial, updates interface{}) (*Dial, error)
	ById(id string) (*Dial, error)
	Delete(container *Dial) error
}

func newDialClient(rancherClient *RancherClient) *DialClient {
	return &DialClient{
		rancherClient: rancherClient,
	}
}

func (c *DialClient) Create(container *Dial) (*Dial, error) {
	resp := &Dial{}
	err := c.rancherClient.doCreate(DIAL_TYPE, container, resp)
	return resp, err
}

func (c *DialClient) Update(existing *Dial, updates interface{}) (*Dial, error) {
	resp := &Dial{}
	err := c.rancherClient.doUpdate(DIAL_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *DialClient) List(opts *ListOpts) (*DialCollection, error) {
	resp := &DialCollection{}
	err := c.rancherClient.doList(DIAL_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *DialCollection) Next() (*DialCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &DialCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *DialClient) ById(id string) (*Dial, error) {
	resp := &Dial{}
	err := c.rancherClient.doById(DIAL_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *DialClient) Delete(container *Dial) error {
	return c.rancherClient.doResourceDelete(DIAL_TYPE, &container.Resource)
}
