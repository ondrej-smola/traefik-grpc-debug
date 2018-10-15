package license

import (
	"sync"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/cipher"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/timeutil"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/zk"
)

type RequestAcceptor interface {
	Accept() bool
}

type CreditStore interface {
	Use(count uint64)
	RequestAcceptor
	Snapshot() (*CreditSnapshot, error)
}

type CreditIndexFn func() uint64

func MonthCreditIndexFn() uint64 {
	now := timeutil.NowUTC()
	return uint64(now.Year()*12 + int(now.Month()-1))
}

type ZkCreditStore struct {
	cl zk.Client

	creditMaxTimeSkew        time.Duration // max difference between credit created and zk mtime
	creditPersistMaxAttempts int
	indexFn                  CreditIndexFn
	limitsW                  LimitsWatcher
	cph                      cipher.Cipher

	lock        *sync.Mutex
	outOfCredit bool
	usedCredit  uint64
}

type ZkCreditStoreOpt func(z *ZkCreditStore)

func WithZkCreditStorePersistMaxAttempts(max int) ZkCreditStoreOpt {
	return func(z *ZkCreditStore) {
		z.creditPersistMaxAttempts = max
	}
}

func WithZkCreditStoreCreditIndexFn(fn CreditIndexFn) ZkCreditStoreOpt {
	return func(z *ZkCreditStore) {
		z.indexFn = fn
	}
}

func WithZkCreditStoreMaxTimeSkew(max time.Duration) ZkCreditStoreOpt {
	return func(z *ZkCreditStore) {
		z.creditMaxTimeSkew = max
	}
}

func NewZkCreditStore(cl zk.Client, limitsW LimitsWatcher, cph cipher.Cipher, opts ...ZkCreditStoreOpt) (*ZkCreditStore, error) {
	z := &ZkCreditStore{
		cl:                       cl,
		creditMaxTimeSkew:        1 * time.Minute,
		creditPersistMaxAttempts: 50,
		indexFn:                  MonthCreditIndexFn,
		limitsW:                  limitsW,
		cph:                      cph,

		lock:        &sync.Mutex{},
		outOfCredit: true,
	}

	for _, o := range opts {
		o(z)
	}

	return z, z.init()
}

func (z *ZkCreditStore) init() error {
	monthIndex := z.indexFn()
	c, _, err := z.getCreditSnapshot()
	if err != nil {
		return err
	}

	z.outOfCredit = c.Credits[monthIndex] > z.limitsW.Limits().MonthlyCredit
	return nil
}

func (z *ZkCreditStore) Accept() bool {
	z.lock.Lock()
	defer z.lock.Unlock()
	return !z.outOfCredit
}

func (z *ZkCreditStore) Use(count uint64) {
	z.lock.Lock()
	z.usedCredit += count
	z.lock.Unlock()
}

func (z *ZkCreditStore) Snapshot() (*CreditSnapshot, error) {
	c, _, err := z.getCreditSnapshot()
	return c, err
}

// blocks till synchronization error occurs
func (z *ZkCreditStore) Synchronize() error {
	index := z.indexFn()

	z.lock.Lock()
	creditToStore := z.usedCredit
	z.usedCredit = 0
	z.lock.Unlock()

	for i := 0; i < z.creditPersistMaxAttempts; i++ {
		c, kv, err := z.getCreditSnapshot()
		if err != nil {
			return err
		}

		c.Credits[index] += creditToStore
		c.Created = timeutil.NowUTCUnixSeconds()
		z.lock.Lock()
		z.outOfCredit = c.Credits[index] > z.limitsW.Limits().MonthlyCredit
		z.lock.Unlock()

		encCreditBytes, err := z.encryptCreditSnapshot(c)
		if err != nil {
			return errors.Wrap(err, "encrypt credit")
		}

		if err := z.cl.AtomicPut(ZkCreditPath, encCreditBytes, kv); err != nil {
			if err == zk.ErrKeyModified {
				continue
			} else {
				return errors.Wrap(err, "zk update credit")
			}
		}
		break
	}

	return nil
}

func (z *ZkCreditStore) getCreditSnapshot() (*CreditSnapshot, *zk.KVPair, error) {
	kv, err := z.cl.Get(ZkCreditPath)
	if err != nil {
		return nil, nil, errors.Wrap(err, "credit snapshot")
	}

	creditBytes, err := z.cph.Decrypt(kv.Value)
	if err != nil {
		return nil, nil, errors.Wrap(err, "decrypt credit")
	}

	c := &CreditSnapshot{}

	if err := proto.Unmarshal(creditBytes, c); err != nil {
		return nil, nil, err
	}

	if err := c.Valid(); err != nil {
		return nil, nil, errors.Wrap(err, "credit")
	}

	persistedTime := time.Unix(kv.Stat.Mtime/1000, 0)
	created := time.Unix(int64(c.Created), 0)

	if persistedTime.Before(created.Add(-z.creditMaxTimeSkew)) || persistedTime.After(created.Add(z.creditMaxTimeSkew)) {
		return nil, nil, errors.Errorf("credit time skew: %v", persistedTime.Sub(created))
	}

	if c.Credits == nil {
		c.Credits = make(map[uint64]uint64)
	}

	return c, kv, nil
}

func (z *ZkCreditStore) encryptCreditSnapshot(c *CreditSnapshot) ([]byte, error) {
	return z.cph.Encrypt(util.MustProto(c))
}
