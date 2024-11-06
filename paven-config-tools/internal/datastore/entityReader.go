package datastore

import (
	"cloud.google.com/go/datastore"
	"context"
	"fmt"
	"google.golang.org/api/iterator"
)

type Reader struct {
	client    *datastore.Client
	Namespace string
	Kind      string
	Entities  []GenericEntiy
}

func ReadConfig(ctx context.Context, client *datastore.Client, namespace string, kind string) Reader {
	q := datastore.NewQuery(kind)
	q = q.Namespace(namespace)

	it := client.Run(ctx, q)

	entities := make([]GenericEntiy, 0)

	for {
		var e GenericEntiy
		k, err := it.Next(&e)
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Println(err)
		}
		e.K = k
		entities = append(entities, e)
	}

	return Reader{
		client:    client,
		Namespace: namespace,
		Kind:      kind,
		Entities:  entities,
	}
}
