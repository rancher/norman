package proxy

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	"github.com/pkg/errors"
	"github.com/rancher/norman/v2/pkg/types"
	"github.com/rancher/norman/v2/pkg/types/convert"
	"github.com/rancher/norman/v2/pkg/types/values"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apitypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

type ClientGetter interface {
	Client(ctx *types.APIRequest, schema *types.Schema) (dynamic.ResourceInterface, error)
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

func (s *Store) ByID(apiOp *types.APIRequest, schema *types.Schema, id string) (types.APIObject, error) {
	_, result, err := s.byID(apiOp, schema, id)
	return types.ToAPI(result), err
}

func (s *Store) byID(apiOp *types.APIRequest, schema *types.Schema, id string) (string, map[string]interface{}, error) {
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

func max(old int, newInt string) int {
	v, err := strconv.Atoi(newInt)
	if err != nil {
		return old
	}
	if v > old {
		return v
	}
	return old
}

func (s *Store) List(apiOp *types.APIRequest, schema *types.Schema, opt *types.QueryOptions) (types.APIObject, error) {
	resultList := &unstructured.UnstructuredList{}

	var (
		errGroup errgroup.Group
		mux      sync.Mutex
		revision int
	)

	if len(apiOp.Namespaces) <= 1 {
		k8sClient, err := s.clientGetter.Client(apiOp, schema)
		if err != nil {
			return types.APIObject{}, err
		}

		resultList, err = k8sClient.List(listOpts(apiOp))
		if err != nil {
			return types.APIObject{}, err
		}
		revision = max(revision, resultList.GetResourceVersion())
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
				revision = max(revision, list.GetResourceVersion())
				mux.Unlock()

				return nil
			})
		}
		if err := errGroup.Wait(); err != nil {
			return types.APIObject{}, err
		}
	}

	var result []map[string]interface{}
	for _, obj := range resultList.Items {
		result = append(result, s.fromInternal(apiOp, schema, obj.Object))
	}

	apiObject := types.ToAPI(result)
	if revision > 0 {
		apiObject.ListRevision = strconv.Itoa(revision)
	}
	return apiObject, nil
}

func listOpts(apiOp *types.APIRequest) metav1.ListOptions {
	opts := metav1.ListOptions{
		LabelSelector:   apiOp.Request.URL.Query().Get("labelSelector"),
		ResourceVersion: apiOp.Request.URL.Query().Get("resourceVersion"),
		Continue:        apiOp.Request.URL.Query().Get("continue"),
	}
	if apiOp.Pagination.Limit != nil {
		opts.Limit = *apiOp.Pagination.Limit
	}
	return opts
}

func (s *Store) listNamespace(namespace string, apiOp types.APIRequest, schema *types.Schema) (*unstructured.UnstructuredList, error) {
	apiOp.Namespaces = []string{namespace}
	k8sClient, err := s.clientGetter.Client(&apiOp, schema)
	if err != nil {
		return nil, err
	}

	return k8sClient.List(listOpts(&apiOp))
}

func returnErr(err error, c chan types.APIEvent) {
	c <- types.APIEvent{
		Name:  "resource.error",
		Error: err,
	}
}

func (s *Store) listAndWatch(apiOp *types.APIRequest, k8sClient dynamic.ResourceInterface, schema *types.Schema, w types.WatchRequest, result chan types.APIEvent) {
	rev := w.Revision
	if rev == "" {
		list, err := k8sClient.List(metav1.ListOptions{
			Limit: 1,
		})
		if err != nil {
			returnErr(errors.Wrapf(err, "failed to list %s", schema.ID), result)
			return
		}
		rev = list.GetResourceVersion()
	} else if rev == "-1" {
		rev = ""
	}

	timeout := int64(60 * 30)
	watcher, err := k8sClient.Watch(metav1.ListOptions{
		Watch:           true,
		TimeoutSeconds:  &timeout,
		ResourceVersion: rev,
	})
	if err != nil {
		returnErr(errors.Wrapf(err, "stopping watch for %s: %v", schema.ID, err), result)
		return
	}
	defer watcher.Stop()
	logrus.Debugf("opening watcher for %s", schema.ID)

	go func() {
		<-apiOp.Request.Context().Done()
		watcher.Stop()
	}()

	for event := range watcher.ResultChan() {
		data := event.Object.(*unstructured.Unstructured)
		result <- s.toAPIEvent(apiOp, schema, event.Type, data)
	}
}

func (s *Store) Watch(apiOp *types.APIRequest, schema *types.Schema, w types.WatchRequest) (chan types.APIEvent, error) {
	k8sClient, err := s.clientGetter.Client(apiOp, schema)
	if err != nil {
		return nil, err
	}

	result := make(chan types.APIEvent)
	go func() {
		s.listAndWatch(apiOp, k8sClient, schema, w, result)
		logrus.Debugf("closing watcher for %s", schema.ID)
		close(result)
	}()
	return result, nil
}

func (s *Store) toAPIEvent(apiOp *types.APIRequest, schema *types.Schema, et watch.EventType, obj *unstructured.Unstructured) types.APIEvent {
	name := types.ChangeAPIEvent
	switch et {
	case watch.Deleted:
		name = types.RemoveAPIEvent
	case watch.Added:
		name = types.CreateAPIEvent
	}

	s.fromInternal(apiOp, schema, obj.Object)

	return types.APIEvent{
		Name:     name,
		Revision: obj.GetResourceVersion(),
		Object:   types.ToAPI(obj.Object),
	}
}

func (s *Store) Create(apiOp *types.APIRequest, schema *types.Schema, params types.APIObject) (types.APIObject, error) {
	data := params.Map()
	if err := s.toInternal(schema.Mapper, data); err != nil {
		return types.APIObject{}, err
	}

	name, _ := values.GetValueN(data, "metadata", "name").(string)
	if name == "" {
		generated, _ := values.GetValueN(data, "metadata", "generateName").(string)
		if generated == "" {
			values.PutValue(data, types.GenerateName(schema.ID), "metadata", "name")
		}
	}

	k8sClient, err := s.clientGetter.Client(apiOp, schema)
	if err != nil {
		return types.APIObject{}, err
	}

	resp, err := k8sClient.Create(&unstructured.Unstructured{Object: data}, metav1.CreateOptions{})
	if err != nil {
		return types.APIObject{}, err
	}
	_, result, err := s.singleResult(apiOp, schema, resp)
	return types.ToAPI(result), err
}

func (s *Store) toInternal(mapper types.Mapper, data map[string]interface{}) error {
	if mapper != nil {
		if err := mapper.ToInternal(data); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Update(apiOp *types.APIRequest, schema *types.Schema, params types.APIObject, id string) (types.APIObject, error) {
	var (
		result map[string]interface{}
		err    error
		data   = params.Map()
	)

	k8sClient, err := s.clientGetter.Client(apiOp, schema)
	if err != nil {
		return types.APIObject{}, err
	}

	if apiOp.Method == http.MethodPatch {
		bytes, err := ioutil.ReadAll(io.LimitReader(apiOp.Request.Body, 2<<20))
		if err != nil {
			return types.APIObject{}, err
		}

		pType := apitypes.StrategicMergePatchType
		if apiOp.Request.Header.Get("content-type") == string(apitypes.JSONPatchType) {
			pType = apitypes.JSONPatchType
		}

		resp, err := k8sClient.Patch(id, pType, bytes, metav1.PatchOptions{})
		if err != nil {
			return types.APIObject{}, err
		}

		_, result, err = s.singleResult(apiOp, schema, resp)
		return types.ToAPI(result), err
	}

	if err := s.toInternal(schema.Mapper, data); err != nil {
		return types.APIObject{}, err
	}

	resourceVersion := convert.ToString(values.GetValueN(data, "metadata", "resourceVersion"))
	if resourceVersion == "" {
		return types.APIObject{}, fmt.Errorf("metadata.resourceVersion is required for update")
	}

	resp, err := k8sClient.Update(&unstructured.Unstructured{Object: data}, metav1.UpdateOptions{})
	if err != nil {
		return types.APIObject{}, err
	}

	_, result, err = s.singleResult(apiOp, schema, resp)
	return types.ToAPI(result), err
}

func (s *Store) Delete(apiOp *types.APIRequest, schema *types.Schema, id string) (types.APIObject, error) {
	k8sClient, err := s.clientGetter.Client(apiOp, schema)
	if err != nil {
		return types.APIObject{}, err
	}

	if err := k8sClient.Delete(id, nil); err != nil {
		return types.APIObject{}, err
	}

	_, obj, err := s.byID(apiOp, schema, id)
	if err != nil {
		return types.APIObject{}, nil
	}
	return types.ToAPI(obj), nil
}

func (s *Store) singleResult(apiOp *types.APIRequest, schema *types.Schema, result *unstructured.Unstructured) (string, map[string]interface{}, error) {
	version, data := result.GetResourceVersion(), result.Object
	s.fromInternal(apiOp, schema, data)
	return version, data, nil
}

func (s *Store) fromInternal(apiOp *types.APIRequest, schema *types.Schema, data map[string]interface{}) map[string]interface{} {
	if apiOp.Option("export") == "true" {
		delete(data, "status")
	}
	if schema.Mapper != nil {
		schema.Mapper.FromInternal(data)
	}

	data["type"] = schema.ID
	return data
}
