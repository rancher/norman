package proxy

import (
	"bytes"
	"encoding/json"
	"github.com/rancher/norman/authorization"
	"github.com/rancher/norman/types"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"
	"net/http"
	"sync"
	"testing"
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

func TestList(t *testing.T) {

	var data = v1.ConfigMapList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMapList",
			APIVersion: "v1",
		},
		ListMeta: metav1.ListMeta{
			ResourceVersion:    "v1",
			RemainingItemCount: new(int64),
		},
		Items: []v1.ConfigMap{
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "default",
				},
				Immutable: new(bool),
				Data: map[string]string{
					"a": "av",
					"b": "bv",
					"c": "cv",
				},
			},
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test2",
					Namespace: "default",
				},
				Immutable: new(bool),
				Data: map[string]string{
					"a2": "av",
					"b2": "bv",
					"c2": "cv",
				},
			},
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test3",
					Namespace: "default",
				},
				Immutable: new(bool),
				Data: map[string]string{
					"a3": "av",
					"b3": "bv",
					"c3": "cv",
				},
			},
		},
	}

	clientGetter := mockClientGetter{
		&fake.RESTClient{
			NegotiatedSerializer: serializer.NewCodecFactory(runtime.NewScheme()),
		},
	}

	typer := runtime.NewScheme()

	var sut = &Store{
		Mutex:          sync.Mutex{},
		clientGetter:   &clientGetter,
		group:          "",
		version:        "v1",
		kind:           "ConfigMap",
		resourcePlural: "configmaps",
		typer:          typer,
	}

	schema := types.Schema{
		Mapper: types.Mappers{},
	}

	req, _ := http.NewRequest(http.MethodGet, "", nil)
	apiContext := types.APIContext{
		Request:       req,
		AccessControl: &authorization.AllAccess{},
	}

	// no results
	{
		body := data
		body.Items = nil
		var fakeResponse bytes.Buffer
		_ = json.NewEncoder(&fakeResponse).Encode(body)
		clientGetter.RESTClient.Resp = &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(&fakeResponse),
		}

		res, err := sut.List(&apiContext, &schema, &types.QueryOptions{})

		assert.NoError(t, err)
		assert.IsType(t, []map[string]interface{}{}, res)
		assert.Len(t, res, 0)
	}

	// generic type
	{
		body := data
		var fakeResponse bytes.Buffer
		_ = json.NewEncoder(&fakeResponse).Encode(body)
		clientGetter.RESTClient.Resp = &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(&fakeResponse),
		}

		res, err := sut.List(&apiContext, &schema, &types.QueryOptions{})

		assert.NoError(t, err)
		assert.IsType(t, []map[string]interface{}{}, res)
		assert.Len(t, res, 3)
	}

	_ = v1.SchemeBuilder.AddToScheme(typer)

	// specific type
	{
		body := data
		var fakeResponse bytes.Buffer
		_ = json.NewEncoder(&fakeResponse).Encode(body)
		clientGetter.RESTClient.Resp = &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(&fakeResponse),
		}

		res, err := sut.List(&apiContext, &schema, &types.QueryOptions{})

		assert.NoError(t, err)
		assert.IsType(t, []map[string]interface{}{}, res)
		assert.Len(t, res, 3)
	}
}

type mockClientGetter struct {
	*fake.RESTClient
}

func (m mockClientGetter) UnversionedClient(_ *types.APIContext, _ types.StorageContext) (rest.Interface, error) {
	return m.RESTClient, nil
}

func (m mockClientGetter) APIExtClient(_ *types.APIContext, _ types.StorageContext) (clientset.Interface, error) {
	return nil, nil
}
