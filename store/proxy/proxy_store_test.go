package proxy

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetDeletionOptions(t *testing.T) {
	req, err := http.NewRequest("DELETE", "https://test.url/api", nil)
	assert.Empty(t, err)
	prop := metav1.DeletePropagationBackground
	expected := &metav1.DeleteOptions{
		PropagationPolicy: &prop,
	}
	options, err := getDeleteOption(req)
	assert.Empty(t, err)
	assert.Equal(t, options, expected, "unexpected deletion options for empty query")

	req.URL.RawQuery = "gracePeriodSeconds=0"
	period := int64(0)
	expected = &metav1.DeleteOptions{
		PropagationPolicy:  &prop,
		GracePeriodSeconds: &period,
	}
	options, err = getDeleteOption(req)
	assert.Empty(t, err)
	assert.Equal(t, options, expected, "unexpected deletion options for query 'gracePeriodSeconds=0'")
}
