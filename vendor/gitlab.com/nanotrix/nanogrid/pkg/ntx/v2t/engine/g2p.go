package engine

import (
	"github.com/pkg/errors"
)

func (r *G2PRequest) Valid() error {
	if r == nil {
		return errors.New("is nil")
	}

	if r.Lang == "" {
		return errors.New("language: is blank")
	}

	if r.Text == "" {
		return errors.New("language: text blank")
	}

	return nil
}
