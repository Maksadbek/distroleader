// Package KV implements Key-Value store backed by Badger
package kv

import (
	"github.com/dgraph-io/badger"
)

type Store struct {
	kv *badger.KV
}

func New(dir, file string) (*Store, error) {
	opt := badger.DefaultOptions

	opt.Dir = dir
	opt.ValueDir = file

	kv, err := badger.NewKV(&opt)
	if err != nil {
		return nil, err
	}

	return &Store{
		kv: kv,
	}, nil
}

func (s *Store) Put(key string, value string) error {
	return s.kv.Set([]byte(key), []byte(value), 0x00)
}

func (s *Store) Get(key string) (string, error) {
	var item badger.KVItem

	if err := s.kv.Get([]byte(key), &item); err != nil {
		return "", err
	}

	return string(item.Value()), nil
}

func (s *Store) Delete(key string) error {
	return s.kv.Delete([]byte(key))
}
