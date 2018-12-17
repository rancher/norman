package objectset

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"sync"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/rancher/norman/objectclient"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes/scheme"
)

const (
	LabelApplied = "objectset.rio.cattle.io/applied"
	LabelInputID = "objectset.rio.cattle.io/inputid"
)

var (
	patchCache     = map[schema.GroupVersionKind]patchCacheEntry{}
	patchCacheLock = sync.Mutex{}
)

type patchCacheEntry struct {
	patchType types.PatchType
	lookup    strategicpatch.LookupPatchMeta
}

func prepareObjectForCreate(inputID string, obj runtime.Object) error {
	serialized, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	meta, err := meta.Accessor(obj)
	if err != nil {
		return err
	}
	annotations := meta.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations[LabelInputID] = inputID
	annotations[LabelApplied] = appliedToAnnotation(serialized)
	meta.SetAnnotations(annotations)

	return nil
}

func (o *DesiredSet) compareObjects(client objectclient.GenericClient, debugID, inputID string, gvk schema.GroupVersionKind, oldObject, newObject runtime.Object, force bool) error {
	oldMetadata, err := meta.Accessor(oldObject)
	if err != nil {
		return err
	}

	if !force && (o.owner != nil || len(o.objs.inputs) > 0) && oldMetadata.GetAnnotations()[LabelInputID] == inputID {
		return nil
	}

	logrus.Infof("DesiredSet - Inspecting %s %s/%s for %s", gvk, oldMetadata.GetNamespace(), oldMetadata.GetName(), debugID)

	original, err := getOriginal(inputID, oldMetadata)
	if err != nil {
		return err
	}

	if err := prepareObjectForCreate(inputID, newObject); err != nil {
		return err
	}

	modified, err := json.Marshal(newObject)
	if err != nil {
		return err
	}

	current, err := json.Marshal(oldObject)
	if err != nil {
		return err
	}

	patchType, patch, err := doPatch(gvk, original, modified, current)
	if err != nil {
		return errors.Wrap(err, "patch generation")
	}

	if string(patch) == "{}" || len(patch) < 2 {
		logrus.Infof("DesiredSet - No change %s %s/%s for %s", gvk, oldMetadata.GetNamespace(), oldMetadata.GetName(), debugID)
		return nil
	}

	logrus.Infof("DesiredSet - Updated %s %s/%s for %s -- patch -- %s", gvk, oldMetadata.GetNamespace(), oldMetadata.GetName(), debugID, patch)
	_, err = client.Patch(oldMetadata.GetName(), oldObject, patchType, patch)
	return err
}

func getOriginal(inputID string, obj v1.Object) ([]byte, error) {
	original := appliedFromAnnotation(obj.GetAnnotations()[LabelApplied])
	if len(original) == 0 {
		return []byte("{}"), nil
	}

	mapObj := &unstructured.Unstructured{}
	err := json.Unmarshal(original, mapObj)
	if err != nil {
		return nil, err
	}

	if err := prepareObjectForCreate(inputID, mapObj); err != nil {
		return nil, err
	}

	return json.Marshal(mapObj)
}

func appliedFromAnnotation(str string) []byte {
	if len(str) == 0 || str[0] == '{' {
		return []byte(str)
	}

	b, err := base64.RawStdEncoding.DecodeString(str)
	if err != nil {
		return nil
	}

	r, err := gzip.NewReader(bytes.NewBuffer(b))
	if err != nil {
		return nil
	}

	b, err = ioutil.ReadAll(r)
	if err != nil {
		return nil
	}

	return b
}

func appliedToAnnotation(b []byte) string {
	if len(b) < 1024 {
		return string(b)
	}
	buf := &bytes.Buffer{}
	w := gzip.NewWriter(buf)
	if _, err := w.Write(b); err != nil {
		return string(b)
	}
	if err := w.Close(); err != nil {
		return string(b)
	}
	return base64.RawStdEncoding.EncodeToString(buf.Bytes())
}

// doPatch is adapted from "kubectl apply"
func doPatch(gvk schema.GroupVersionKind, original, modified, current []byte) (types.PatchType, []byte, error) {
	var patchType types.PatchType
	var patch []byte
	var lookupPatchMeta strategicpatch.LookupPatchMeta

	patchType, lookupPatchMeta, err := getPatchStyle(gvk)
	if err != nil {
		return patchType, nil, err
	}

	if patchType == types.StrategicMergePatchType {
		patch, err = strategicpatch.CreateThreeWayMergePatch(original, modified, current, lookupPatchMeta, true)
	} else {
		patch, err = jsonmergepatch.CreateThreeWayJSONMergePatch(original, modified, current)
	}

	if err != nil {
		logrus.Errorf("Failed to calcuated patch: %v", err)
	}

	return patchType, patch, err
}

func getPatchStyle(gvk schema.GroupVersionKind) (types.PatchType, strategicpatch.LookupPatchMeta, error) {
	var (
		patchType       types.PatchType
		lookupPatchMeta strategicpatch.LookupPatchMeta
	)

	patchCacheLock.Lock()
	entry, ok := patchCache[gvk]
	patchCacheLock.Unlock()

	if ok {
		return entry.patchType, entry.lookup, nil
	}

	versionedObject, err := scheme.Scheme.New(gvk)

	if runtime.IsNotRegisteredError(err) {
		patchType = types.MergePatchType
	} else if err != nil {
		return patchType, nil, err
	} else {
		patchType = types.StrategicMergePatchType
		lookupPatchMeta, err = strategicpatch.NewPatchMetaFromStruct(versionedObject)
		if err != nil {
			return patchType, nil, err
		}
	}

	patchCacheLock.Lock()
	patchCache[gvk] = patchCacheEntry{
		patchType: patchType,
		lookup:    lookupPatchMeta,
	}
	patchCacheLock.Unlock()

	return patchType, lookupPatchMeta, nil
}
