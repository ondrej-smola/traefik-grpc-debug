package auth

import (
	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/timeutil"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
)

type Users []*User
type Groups []*Group

func (v *UserStoreSnapshot) Valid() error {
	if v == nil {
		return errors.New("is nil")
	}

	if err := Users(v.Users).Valid(); err != nil {
		return errors.Wrap(err, "users")
	}

	if err := Groups(v.Groups).Valid(); err != nil {
		return errors.Wrap(err, "groups")
	}

	return nil
}

func (v *UserStoreSnapshot) Clone() *UserStoreSnapshot {
	return &UserStoreSnapshot{
		Groups: Groups(v.Groups).Clone(),
		Users:  Users(v.Users).Clone(),
	}
}

func (u Users) Valid() error {
	for i, iu := range u {
		if err := iu.Valid(); err != nil {
			return errors.Wrapf(err, "user[%v]", i)
		}
	}

	return nil
}

func (u Users) Clone() Users {
	cpy := make(Users, len(u))

	for i, iu := range u {
		cpy[i] = iu.Clone()
	}

	return cpy
}

func (u *User) Clone() *User {
	return &User{
		Id:          u.Id,
		Secret:      u.Secret,
		Email:       u.Email,
		Groups:      util.CloneStringSlice(u.Groups),
		Permissions: util.CloneStringSlice(u.Permissions),
		Expire:      u.Expire,
		Created:     u.Created,
		Blocked:     u.Blocked,
	}
}

func (u *User) Valid() error {
	if u == nil {
		return errors.Errorf("is nil")
	}

	if u.Id == "" {
		return errors.Errorf("id: is blank")
	}

	if u.Secret == "" {
		return errors.Errorf("secret: is blank")
	}

	if u.Email == "" {
		return errors.Errorf("email: is blank")
	}

	if u.Created == 0 {
		return errors.Errorf("created: is zero")
	}

	return nil
}

func (u *User) Expired() bool {
	return u.Expire > 0 && u.Expire < timeutil.NowUTCUnixSeconds()
}

func (g Groups) Valid() error {
	for i, ig := range g {
		if err := ig.Valid(); err != nil {
			return errors.Wrapf(err, "group[%v]", i)
		}
	}

	return nil
}

func (g Groups) Clone() Groups {
	cpy := make(Groups, len(g))

	for i, ig := range g {
		cpy[i] = ig.Clone()
	}

	return cpy
}

func (g *Group) Clone() *Group {
	return &Group{
		Id:          g.Id,
		Permissions: util.CloneStringSlice(g.Permissions),
	}
}

func (g *Group) Valid() error {
	if g == nil {
		return errors.Errorf("is nil")
	}

	if g.Id == "" {
		return errors.Errorf("id: is blank")
	}

	return nil
}
