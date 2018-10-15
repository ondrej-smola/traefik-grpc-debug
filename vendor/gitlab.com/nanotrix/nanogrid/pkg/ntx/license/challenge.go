package license

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"

	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
)

func ClusterToChallenge(c *Cluster) []byte {
	return util.MustJsonPb(c)
}

func ClusterFromChallenge(ch []byte) (*Cluster, error) {
	c := &Cluster{}

	if err := jsonpb.Unmarshal(bytes.NewReader(ch), c); err != nil {
		return nil, errors.Wrap(err, "unmarshal")
	} else if err := c.Valid(); err != nil {
		return nil, errors.Wrap(err, "challenge")
	} else {
		return c, nil
	}
}

func DecodeCreditSnapshot(cl *Cluster, body []byte) (*CreditSnapshot, error) {
	cph := NewCipherForCluster(cl)
	creditSnapBytes, err := cph.Decrypt(body)
	if err != nil {
		return nil, errors.Wrap(err, "decrypt credit")
	}

	snap := &CreditSnapshot{}
	if err := proto.Unmarshal(creditSnapBytes, snap); err != nil {
		return nil, errors.Wrap(err, "unmarshal credit snapshot")
	} else if err := snap.Valid(); err != nil {
		return nil, errors.Wrap(err, "credit snapshot")
	} else if snap.ClusterId != cl.Id {
		return nil, errors.Errorf("credit snapshot: invalid cluster id: %v", snap.ClusterId)
	}

	return snap, nil
}

func EncodeChallenge(ch *Challenge) (string, error) {
	cph := NewCipherForCluster(ch.Cluster)
	credBytes, err := cph.Encrypt(util.MustProto(ch.Credit))

	if err != nil {
		return "", errors.Wrap(err, "encrypt credit")
	}

	clBytes := ClusterToChallenge(ch.Cluster)

	return fmt.Sprintf(
		"%v.%v",
		base64.StdEncoding.EncodeToString(clBytes),
		base64.StdEncoding.EncodeToString(credBytes)), nil
}

func DecodeChallenge(challenge string) (*Challenge, error) {
	challengeParts := strings.Split(challenge, ".")
	if len(challengeParts) != 2 {
		return nil, errors.Errorf("challenge string: unexpected number of parts: %v", len(challengeParts))
	}

	clusterPart := challengeParts[0]
	creditPart := challengeParts[1]

	clusterBytes, err := base64.StdEncoding.DecodeString(clusterPart)
	if err != nil {
		return nil, errors.Wrap(err, "cluster part")
	}

	cluster, err := ClusterFromChallenge(clusterBytes)
	if err != nil {
		return nil, errors.Wrap(err, "decode cluster")
	}

	creditBytes, err := base64.StdEncoding.DecodeString(creditPart)
	if err != nil {
		return nil, errors.Wrap(err, "cluster part")
	}

	snap, err := DecodeCreditSnapshot(cluster, creditBytes)
	if err != nil {
		return nil, errors.Wrap(err, "credit snapshot")
	}

	ch := &Challenge{Cluster: cluster, Credit: snap}

	return ch, errors.Wrap(ch.Valid(), "challenge")
}
