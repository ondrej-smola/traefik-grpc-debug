package zk

import (
	"bytes"

	"github.com/samuel/go-zookeeper/zk"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
)

type KVPair struct {
	Key   string
	Value []byte
	Stat  *zk.Stat
}

type KVPairs []*KVPair

func NewKVPair(k string, v ...byte) *KVPair {
	return &KVPair{Key: k, Value: v, Stat: &zk.Stat{}}
}

func (k *KVPair) Equal(o *KVPair) bool {
	if k.Key != o.Key {
		return false
	}

	if !bytes.Equal(k.Value, o.Value) {
		return false
	}

	return k.Stat.Version == o.Stat.Version
}

func (k *KVPair) Clone() *KVPair {
	cpy := *k.Stat

	return &KVPair{
		Key:   k.Key,
		Value: util.CloneBytes(k.Value),
		Stat:  &cpy,
	}
}

func (k KVPairs) Clone() KVPairs {
	cpy := make(KVPairs, len(k))

	for i, ki := range k {

		cpy[i] = ki.Clone()
	}

	return cpy
}
