package proxy

import (
	"context"
	"sync"

	"github.com/rancher/norman/pkg/types"
	convert2 "github.com/rancher/norman/pkg/types/convert"
	merge2 "github.com/rancher/norman/pkg/types/convert/merge"
	values2 "github.com/rancher/norman/pkg/types/values"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

type ClientGetter interface {
	Client(ctx *types.APIOperation, schema *types.Schema) (dynamic.ResourceInterface, error)
}

type Store struct {
	clientGetter ClientGetter
}

func NewProxyStore(clientGetter ClientGetter) types.Store {
	return &errorStore{
		Store: &Store{
			clientGetter: clientGetter,
		},
	}
}

func (s *Store) ByID(apiOp *types.APIOperation, schema *types.Schema, id string) (map[string]interface{}, error) {
	_, result, err := s.byID(apiOp, schema, id)
	return result, err
}

func (s *Store) byID(apiOp *types.APIOperation, schema *types.Schema, id string) (string, map[string]interface{}, error) {
	k8sClient, err := s.clientGetter.Client(apiOp, schema)
	if err != nil {
		return "", nil, err
	}

	resp, err := k8sClient.Get(id, metav1.GetOptions{})
	if err != nil {
		return "", nil, err
	}
	return s.singleResult(apiOp, schema, resp)
}

func (s *Store) List(apiOp *types.APIOperation, schema *types.Schema, opt *types.QueryOptions) ([]map[string]interface{}, error) {
	resultList := &unstructured.UnstructuredList{}

	var (
		errGroup errgroup.Group
		mux      sync.Mutex
	)

	if len(apiOp.Namespaces) <= 1 {
		k8sClient, err := s.clientGetter.Client(apiOp, schema)
		if err != nil {
			return nil, err
		}

		resultList, err = k8sClient.List(metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
	} else {
		allNS := apiOp.Namespaces
		for _, ns := range allNS {
			nsCopy := ns
			errGroup.Go(func() error {
				list, err := s.listNamespace(nsCopy, *apiOp, schema)
				if err != nil {
					return err
				}

				mux.Lock()
				resultList.Items = append(resultList.Items, list.Items...)
				mux.Unlock()

				return nil
			})
		}
		if err := errGroup.Wait(); err != nil {
			return nil, err
		}
	}

	var result []map[string]interface{}
	for _, obj := range resultList.Items {
		result = append(result, s.fromInternal(apiOp, schema, obj.Object))
	}

	return apiOp.AccessControl.FilterList(apiOp, schema, result), nil
}

func (s *Store) listNamespace(namespace string, apiOp types.APIOperation, schema *types.Schema) (*unstructured.UnstructuredList, error) {
	apiOp.Namespaces = []string{namespace}
	k8sClient, err := s.clientGetter.Client(&apiOp, schema)
	if err != nil {
		return nil, err
	}

	return k8sClient.List(metav1.ListOptions{})
}

func (s *Store) Watch(apiOp *types.APIOperation, schema *types.Schema, opt *types.QueryOptions) (chan map[string]interface{}, error) {
	c, err := s.realWatch(apiOp, schema, &types.QueryOptions{})
	//c, err := s.shareWatch(apiOp, schema, opt)
	if err != nil {
		return nil, err
	}

	return convert2.Chan(c, func(data map[string]interface{}) map[string]interface{} {
		return apiOp.AccessControl.Filter(apiOp, schema, data)
	}), nil
}

func (s *Store) realWatch(apiOp *types.APIOperation, schema *types.Schema, opt *types.QueryOptions) (chan map[string]interface{}, error) {
	k8sClient, err := s.clientGetter.Client(apiOp, schema)
	if err != nil {
		return nil, err
	}

	timeout := int64(60 * 30)
	watcher, err := k8sClient.Watch(metav1.ListOptions{
		Watch:           true,
		TimeoutSeconds:  &timeout,
		ResourceVersion: "0",
	})
	if err != nil {
		return nil, err
	}

	watchingContext, cancelWatchingContext := context.WithCancel(apiOp.Request.Context())
	go func() {
		<-watchingContext.Done()
		logrus.Debugf("stopping watcher for %s", schema.ID)
		watcher.Stop()
	}()

	result := make(chan map[string]interface{})
	go func() {
		for event := range watcher.ResultChan() {
			data := event.Object.(*unstructured.Unstructured)
			s.fromInternal(apiOp, schema, data.Object)
			if event.Type == watch.Deleted && data.Object != nil {
				data.Object[".removed"] = true
			}
			result <- data.Object
		}
		logrus.Debugf("closing watcher for %s", schema.ID)
		close(result)
		cancelWatchingContext()
	}()

	return result, nil
}

func (s *Store) Create(apiOp *types.APIOperation, schema *types.Schema, data map[string]interface{}) (map[string]interface{}, error) {
	if err := s.toInternal(schema.Mapper, data); err != nil {
		return nil, err
	}

	values2.PutValue(data, apiOp.GetUser(), "metadata", "annotations", "field.cattle.io/creatorId")
	values2.PutValue(data, "norman", "metadata", "labels", "cattle.io/creator")

	name, _ := values2.GetValueN(data, "metadata", "name").(string)
	if name == "" {
		generated, _ := values2.GetValueN(data, "metadata", "generateName").(string)
		if generated == "" {
			values2.PutValue(data, types.GenerateName(schema.ID), "metadata", "name")
		}
	}

	k8sClient, err := s.clientGetter.Client(apiOp, schema)
	if err != nil {
		return nil, err
	}

	resp, err := k8sClient.Create(&unstructured.Unstructured{Object: data}, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	_, result, err := s.singleResult(apiOp, schema, resp)
	return result, err
}

func (s *Store) toInternal(mapper types.Mapper, data map[string]interface{}) error {
	if mapper != nil {
		if err := mapper.ToInternal(data); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Update(apiOp *types.APIOperation, schema *types.Schema, data map[string]interface{}, id string) (map[string]interface{}, error) {
	var (
		result map[string]interface{}
		err    error
	)

	k8sClient, err := s.clientGetter.Client(apiOp, schema)
	if err != nil {
		return nil, err
	}

	if err := s.toInternal(schema.Mapper, data); err != nil {
		return nil, err
	}

	for i := 0; i < 5; i++ {
		resp, err := k8sClient.Get(id, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		resourceVersion, existing, rawErr := s.singleResult(apiOp, schema, resp)
		if rawErr != nil {
			return nil, rawErr
		}

		existing = merge2.APIUpdateMerge(schema.InternalSchema, apiOp.Schemas, existing, data, apiOp.Option("replace") == "true")

		values2.PutValue(existing, resourceVersion, "metadata", "resourceVersion")
		if len(apiOp.Namespaces) > 0 {
			values2.PutValue(existing, apiOp.Namespaces[0], "metadata", "namespace")
		}
		values2.PutValue(existing, id, "metadata", "name")

		resp, err = k8sClient.Update(resp, metav1.UpdateOptions{})
		if errors.IsConflict(err) {
			continue
		} else if err != nil {
			return nil, err
		}
		_, result, err = s.singleResult(apiOp, schema, resp)
		return result, err
	}

	return result, err
}

func (s *Store) Delete(apiOp *types.APIOperation, schema *types.Schema, id string) (map[string]interface{}, error) {
	k8sClient, err := s.clientGetter.Client(apiOp, schema)
	if err != nil {
		return nil, err
	}

	if err := k8sClient.Delete(id, nil); err != nil {
		return nil, err
	}

	_, obj, err := s.byID(apiOp, schema, id)
	if err != nil {
		return nil, nil
	}
	return obj, nil
}

func (s *Store) singleResult(apiOp *types.APIOperation, schema *types.Schema, result *unstructured.Unstructured) (string, map[string]interface{}, error) {
	version, data := result.GetResourceVersion(), result.Object
	s.fromInternal(apiOp, schema, data)
	return version, data, nil
}

func (s *Store) fromInternal(apiOp *types.APIOperation, schema *types.Schema, data map[string]interface{}) map[string]interface{} {
	if apiOp.Option("export") == "true" {
		delete(data, "status")
	}
	if schema.Mapper != nil {
		schema.Mapper.FromInternal(data)
	}

	return data
}
