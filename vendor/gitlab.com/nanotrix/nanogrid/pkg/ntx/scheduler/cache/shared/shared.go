package shared

import "github.com/pkg/errors"

func (r *TaskRequest) Valid() error {
	if r == nil {
		return errors.New("is nil")
	}

	if err := r.Config.Valid(); err != nil {
		return errors.Wrap(err, "config")
	}

	return nil
}

func (r *TaskResponse) Valid() error {
	if r == nil {
		return errors.New("is nil")
	}

	if r.Found {

		if r.Id == "" {
			return errors.New("id: blank")
		}

		if r.Endpoint == "" {
			return errors.New("endpoint: blank")
		}
	}
	return nil
}

func (r *Tasks) Valid() error {
	if r == nil {
		return errors.New("is nil")
	}

	for i, id := range r.Ids {
		if id == "" {
			return errors.Errorf("ids[%v]: blank", i)
		}
	}

	return nil
}
