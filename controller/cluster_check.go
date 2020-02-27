package controller

import (
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
)

type ObjectClusterName interface {
	ClusterName() string
}

func ObjectInCluster(cluster string, obj interface{}) bool {
	if o, ok := obj.(ObjectClusterName); ok {
		return o.ClusterName() == cluster
	}

	logrus.Warnf("ObjectClusterName not implemented by type %T", obj)

	var clusterName string

	if v, ok := obj.(map[string]interface{}); ok {
		if name, ok := v["ClusterName"]; ok {
			clusterName = name.(string)
		}

		if clusterName == "" {
			if c, ok := v["ProjectName"]; ok {
				if parts := strings.SplitN(c.(string), ":", 2); len(parts) == 2 {
					clusterName = parts[0]
				}
			}
		}

		if clusterName == "" {
			if s, ok := v["Spec"].(map[string]interface{}); ok {
				if name, ok := s["ClusterName"]; ok {
					clusterName = name.(string)
				}
			}
		}

		if clusterName == "" {
			if s, ok := v["Spec"].(map[string]interface{}); ok {
				if name, ok := s["ProjectName"]; ok {
					if parts := strings.SplitN(name.(string), ":", 2); len(parts) == 2 {
						clusterName = parts[0]
					}
				}
			}
		}

		if clusterName == "" {
			if annos, ok := v["Annotations"].(map[string]interface{}); ok {
				if projectID, ok := annos["field.cattle.io/projectId"]; ok {
					if parts := strings.SplitN(projectID.(string), ":", 2); len(parts) == 2 {
						clusterName = parts[0]
					}
				}
			}
		}

		if clusterName == "" {
			if namespace, ok := v["Namespace"]; ok {
				clusterName = namespace.(string)
			}
		}

	}

	return clusterName == cluster
}

func getValue(obj interface{}, name ...string) reflect.Value {
	v := reflect.ValueOf(obj)
	t := v.Type()
	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = v.Type()
	}

	field := v.FieldByName(name[0])
	if !field.IsValid() || len(name) == 1 {
		return field
	}

	return getFieldValue(field, name[1:]...)
}

func getFieldValue(v reflect.Value, name ...string) reflect.Value {
	field := v.FieldByName(name[0])
	if len(name) == 1 {
		return field
	}
	return getFieldValue(field, name[1:]...)
}
