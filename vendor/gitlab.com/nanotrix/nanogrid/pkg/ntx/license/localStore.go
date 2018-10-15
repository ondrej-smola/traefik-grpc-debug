package license

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"

	"github.com/gogo/protobuf/proto"

	"time"

	"sync"

	"github.com/ondrej-smola/mesos-go-http/lib/backoff"
	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/cipher"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/timeutil"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/zk"
)

var ErrNoLicenseFound = errors.New("no license found")
var ErrNotInitialized = errors.New("not initialized")
var ErrAlreadyInitialized = errors.New("already initialized")

type Provider interface {
	License() (*License, error)
}

type Store interface {
	Activate(lic *LicenseResponse) error
	Cluster() (*Cluster, error)
	Challenge() (*ChallengeRequest, error)
	Provider
}

const (
	ZkLicenseBasePath                      = "/ntx/license"
	ZkActiveLicensePath                    = ZkLicenseBasePath + "/active"
	ZkClusterIdPath                        = ZkLicenseBasePath + "/cluster-id"
	ZkChallengePath                        = ZkLicenseBasePath + "/challenge"
	ZkCreditPath                           = "/ntx/credit"
	MinClusterIdChars                      = 32
	ActiveLicenseCreatedMaxTimeSkewSeconds = 60
)

type ZkLocalStore struct {
	zkCl zk.Client

	lock *sync.Mutex
	cph  cipher.Cipher
}

var _ = Store(&ZkLocalStore{})

func NewZkStore(cl zk.Client) *ZkLocalStore {
	return &ZkLocalStore{zkCl: cl}
}

func (z *ZkLocalStore) Init(clusterId, hostname string) error {
	if len(clusterId) < MinClusterIdChars {
		return errors.Errorf("clusterId: must be at least %v characters long", MinClusterIdChars)
	}

	prev, err := z.zkCl.Get(ZkClusterIdPath)
	if prev != nil {
		return ErrAlreadyInitialized
	} else if err == zk.ErrKeyNotFound { // not initialized
		if err := z.zkCl.AtomicPut(ZkClusterIdPath, util.MustProto(&Cluster{Id: clusterId, Hostname: hostname}), nil); err != nil {
			return errors.Wrapf(err, "zk: put: %v", ZkClusterIdPath)
		}

		cl, err := z.clusterWithoutValidation()
		if err != nil {
			return errors.Wrap(err, "get cluster info")
		}

		cph := NewCipherForCluster(cl)
		encChallenge, err := cph.Encrypt(ClusterToChallenge(cl))
		if err != nil {
			return errors.Wrap(err, "encrypt challenge")
		}
		if err := z.zkCl.AtomicPut(ZkChallengePath, encChallenge, nil); err != nil {
			return errors.Wrap(err, "save challenge")
		}

		encCreditSnapshot, err := cph.Encrypt(util.MustProto(&CreditSnapshot{ClusterId: clusterId, Created: timeutil.NowUTCUnixSeconds()}))
		if err != nil {
			return errors.Wrap(err, "encrypt credit snapshot")
		}

		if err := z.zkCl.AtomicPut(ZkCreditPath, encCreditSnapshot, nil); err != nil {
			return errors.Wrap(err, "save challenge")
		}
		return nil
	} else {
		return errors.Wrapf(err, "zk: exists: %v", ZkClusterIdPath)
	}
}

func (z *ZkLocalStore) Cluster() (*Cluster, error) {
	cl, err := z.clusterWithoutValidation()
	if err != nil {
		return nil, err
	}

	if err := z.checkNotModified(cl); err != nil {
		return nil, err
	} else {
		return cl, nil
	}
}

func (z *ZkLocalStore) clusterWithoutValidation() (*Cluster, error) {
	kv, err := z.zkCl.Get(ZkClusterIdPath)
	if err != nil {
		if err == zk.ErrKeyNotFound {
			return nil, ErrNotInitialized
		} else {
			return nil, errors.Wrap(err, "zk: cluster-id")
		}
	}

	buff := make([]byte, 16)
	binary.BigEndian.PutUint64(buff, uint64(kv.Stat.Mtime))
	binary.BigEndian.PutUint64(buff[8:], uint64(kv.Stat.Mzxid))

	hash := sha512.New()
	hash.Write(buff)

	cl := &Cluster{}
	if err := proto.Unmarshal(kv.Value, cl); err != nil {
		return nil, errors.Wrap(err, "cluster-id")
	}

	cl.Created = uint64(kv.Stat.Ctime / 1000)
	cl.Key = base64.URLEncoding.EncodeToString(hash.Sum(nil))

	return cl, errors.Wrap(cl.Valid(), "cluster")
}

func (z *ZkLocalStore) checkNotModified(cl *Cluster) error {
	challKv, err := z.zkCl.Get(ZkChallengePath)
	if err != nil {
		return errors.Wrap(err, "get challenge")
	}

	cph := NewCipherForCluster(cl)

	challengeBytes, err := cph.Decrypt(challKv.Value)
	if err != nil {
		return errors.Wrap(err, "decrypt challenge")
	}

	initCl, err := ClusterFromChallenge(challengeBytes)
	if err != nil {
		return errors.Wrap(err, "unmarshal challenge")
	}

	if !cl.Equal(initCl) {
		return errors.Errorf("cluster signature mismatch")
	}

	return nil
}

func (z *ZkLocalStore) License() (*License, error) {
	cid, err := z.Cluster()
	if err != nil {
		return nil, err
	}

	if err := z.checkNotModified(cid); err != nil {
		return nil, err
	}

	kv, err := z.zkCl.Get(ZkActiveLicensePath)
	if err != nil {
		if err == zk.ErrKeyNotFound {
			return nil, ErrNoLicenseFound
		} else {
			return nil, err
		}
	}

	lic, err := DecodeLicense(string(kv.Value), cid)
	if err != nil {
		return nil, err
	}

	return lic, errors.Wrap(lic.Valid(), "license")
}

func (z *ZkLocalStore) Activate(l *LicenseResponse) error {
	cid, err := z.Cluster()
	if err != nil {
		return err
	}

	lic, err := DecodeLicense(l.License, cid)
	if err != nil {
		return err
	}

	if lic.Expired() {
		return errors.Wrapf(err, "license expired: %v", timeutil.FormatUnixSeconds(lic.Expire))
	}

	if !lic.Cluster.Equal(cid) {
		return errors.Errorf("license: cluster mismatch")
	}

	prev, err := z.zkCl.Get(ZkActiveLicensePath)
	if err != nil && err != zk.ErrKeyNotFound {
		return errors.Wrapf(err, "zk: exists %v", ZkActiveLicensePath)
	}

	return errors.Wrap(z.zkCl.AtomicPut(ZkActiveLicensePath, []byte(l.License), prev), "zk put")
}

func (z *ZkLocalStore) Challenge() (*ChallengeRequest, error) {
	cl, err := z.Cluster()
	if err != nil {
		return nil, err
	}

	kv, err := z.zkCl.Get(ZkCreditPath)
	if err != nil {
		return nil, errors.Wrapf(err, "zk get %v", ZkCreditPath)
	}

	snap, err := DecodeCreditSnapshot(cl, kv.Value)
	if err != nil {
		return nil, errors.Wrap(err, "credit snapshot")
	}

	challenge, err := EncodeChallenge(&Challenge{Cluster: cl, Credit: snap})
	if err != nil {
		return nil, errors.Wrap(err, "encode challenge")
	}

	return &ChallengeRequest{Challenge: challenge}, nil
}

func InfoFromLicense(lic *License) *LicenseInfo {
	licInfo := &LicenseInfo{
		State:        LicenseInfo_ACTIVE,
		Cluster:      lic.Cluster,
		Limits:       lic.Limits,
		TasksVersion: lic.Tasks.Version,
		CustomerId:   lic.CustomerId,
		Created:      lic.Created,
		Expire:       lic.Expire,
	}

	if lic.Expired() {
		licInfo.State = LicenseInfo_EXPIRED
	}

	return licInfo
}

func WaitForActiveLicense(st Provider, max time.Duration) (*License, error) {
	start := time.Now()
	end := start.Add(max)

	retryP := backoff.New(backoff.WithMinWait(10*time.Millisecond), backoff.WithMaxWait(2*time.Second), backoff.Always())
	retry := retryP.New()
	defer retry.Close()

	for range retry.Attempts() {
		lic, err := st.License()
		if err != nil {
			if err != ErrNoLicenseFound && err != ErrNotInitialized {
				return nil, err
			} else if time.Now().After(end) {
				break
			}

		} else if lic.Expired() {
			return nil, errors.Errorf("license expired: %v", timeutil.FormatUnixSeconds(lic.Expire))
		} else {
			return lic, nil
		}
	}

	return nil, errors.Errorf("unable to obtain active license in %v", max)
}

type ValidCheckConfig struct {
	CheckInterval time.Duration
	MaxErrors     int
}

func WatchLicenseIsValid(prov Provider, optCheckInterval ...time.Duration) error {
	checkInterval := time.Minute

	if len(optCheckInterval) > 0 {
		checkInterval = optCheckInterval[0]
	}

	for range time.Tick(checkInterval) {
		lic, err := prov.License()
		if err != nil {
			return errors.Wrap(err, "license watch err")
		}

		if lic.Expired() {
			return errors.Errorf("license expired: %v", timeutil.FormatUnixSeconds(lic.Expire))
		}

		if timeutil.NowUTCUnixSeconds()+ActiveLicenseCreatedMaxTimeSkewSeconds < lic.Created {
			return errors.Errorf("license: time out of sync: created %v, now %v", timeutil.FormatUnixSeconds(lic.Created), timeutil.NowUTCString())
		}
	}

	panic("unreachable")
}
