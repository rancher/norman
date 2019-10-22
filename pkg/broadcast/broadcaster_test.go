package broadcast

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInject(t *testing.T) {
	assert := assert.New(t)
	watchChan := make(chan map[string]interface{})
	testBroadcaster := &Broadcaster{
		inject: watchChan,
	}
	data := map[string]interface{}{
		"test": "test",
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		test := <-testBroadcaster.inject
		assert.Equal("test", test["test"])
		wg.Done()
	}()

	err := testBroadcaster.Inject(data)
	assert.Nil(err)
	wg.Wait()
	testBroadcaster.inject = nil
	err = testBroadcaster.Inject(data)
	assert.Error(err)
	assert.Equal("broadcaster cannot be injected", err.Error())
}
