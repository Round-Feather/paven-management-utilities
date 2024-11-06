package datastore

import (
	"cloud.google.com/go/datastore"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

type GenericEntiy struct {
	K                   *datastore.Key         `datastore:"__key__" json:"key"`
	Properties          map[string]interface{} `datastore:"-" json:"properties"`
	DatastoreProperties map[string]interface{} `datastore:"-" json:"datastore-properties"`
}

func (ge *GenericEntiy) Load(ps []datastore.Property) error {
	for _, p := range ps {
		v, err := parseValue(p.Value)
		if err != nil {
			fmt.Println(err)
		}
		dv, err := parseDatastoreValue(p.Value)
		if err != nil {
			fmt.Println(err)
		}

		if ge.Properties == nil {
			ge.Properties = make(map[string]interface{})
		}
		if ge.DatastoreProperties == nil {
			ge.DatastoreProperties = make(map[string]interface{})
		}

		ge.Properties[p.Name] = v
		ge.DatastoreProperties[p.Name] = dv

	}
	return nil
}

func (ge *GenericEntiy) Save() ([]datastore.Property, error) {
	return []datastore.Property{}, nil
}

func parseValue(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	switch v.(type) {
	case int64, bool, string, float64:
		return v, nil
	default:
		switch reflect.TypeOf(v).Kind() {
		case reflect.Slice:
			s := make([]interface{}, 0)
			for _, sv := range v.([]interface{}) {
				psv, err := parseValue(sv)
				if err != nil {
					return nil, err
				}
				s = append(s, psv)
			}
			return s, nil
		case reflect.Pointer:
			e, ok := v.(*datastore.Entity)
			if ok {
				m := make(map[string]interface{})
				for _, p := range e.Properties {
					mv, err := parseValue(p.Value)
					if err != nil {
						fmt.Println(err)
					}
					m[p.Name] = mv
				}
				return m, nil
			}
		default:
			return nil, errors.New("unexpected type")
		}
	}
	return nil, nil
}

func parseDatastoreValue(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	switch v.(type) {
	case int64:
		return map[string]interface{}{"integerValue": strconv.Itoa(int(v.(int64)))}, nil
	case bool:
		return map[string]interface{}{"booleanValue": v}, nil
	case string:
		return map[string]interface{}{"stringValue": v}, nil
	case float64:
		return map[string]interface{}{"doubleValue": v}, nil
	default:
		switch reflect.TypeOf(v).Kind() {
		case reflect.Slice:
			s := make([]interface{}, 0)
			for _, sv := range v.([]interface{}) {
				psv, err := parseDatastoreValue(sv)
				if err != nil {
					return nil, err
				}
				s = append(s, psv)
			}
			return map[string]interface{}{
				"arrayValue": map[string]interface{}{
					"values": s,
				},
			}, nil
		case reflect.Pointer:
			e, ok := v.(*datastore.Entity)
			if ok {
				m := make(map[string]interface{})
				for _, p := range e.Properties {
					mv, err := parseDatastoreValue(p.Value)
					if err != nil {
						fmt.Println(err)
					}
					m[p.Name] = mv
				}
				return map[string]interface{}{
					"entityValue": map[string]interface{}{
						"properties": m,
					},
				}, nil
			}
		default:
			return nil, errors.New("unexpected type")
		}
	}
	return nil, nil
}
