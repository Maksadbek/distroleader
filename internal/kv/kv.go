// Package KV implements Key-Value store backed by Badger
package kv

import "github.com/dgraph-io/badger"

type Store struct {
	kv *badger.KV
}

func Put(key string, value interface{}) error {
	return nil
}

func Get(key string) (interface{}, error) {
	return nil, nil
}

func Delete(key) error {
	return nil
}
