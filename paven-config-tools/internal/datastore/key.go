package datastore

import "strings"

type KeyType int

const (
	EntityKey KeyType = iota
	EntityProperty
	EntityParent
)

type Key interface {
	Extract(entiy GenericEntiy) string
	GetType() KeyType
}

type KeyKey struct {
}

func (k *KeyKey) Extract(entity GenericEntiy) string {
	return entity.K.Name
}

func (k *KeyKey) GetType() KeyType {
	return EntityKey
}

type PropertyKey struct {
	Property string
}

func NewPropertyKey(configId string) PropertyKey {
	return PropertyKey{
		Property: strings.TrimPrefix(configId, "property:"),
	}
}

func (k *PropertyKey) Extract(entity GenericEntiy) string {
	return entity.Properties[k.Property].(string)
}

func (k *PropertyKey) GetType() KeyType {
	return EntityProperty
}

type ParentKey struct {
	Parent string
}

func NewParentKey(configId string) ParentKey {
	return ParentKey{
		Parent: strings.TrimPrefix(configId, "parent:"),
	}
}

func (k *ParentKey) Extract(entity GenericEntiy) string {
	key := entity.K
	for {
		key = key.Parent
		if key.Kind == k.Parent {
			return key.Name
		}
	}
}

func (k *ParentKey) GetType() KeyType {
	return EntityParent
}

func NewKey(configId string) Key {
	if strings.HasPrefix(configId, "key") {
		return &KeyKey{}
	}
	if strings.HasPrefix(configId, "property") {
		k := NewPropertyKey(configId)
		return &k
	}
	if strings.HasPrefix(configId, "parent") {
		k := NewParentKey(configId)
		return &k
	}
	return nil
}
