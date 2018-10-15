package auth

import (
	"time"

	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/auth/perm"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/scheduler"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/timeutil"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
)

func TestNtxToken() *NtxToken {
	tkn := &NtxToken{
		Iss:  "test-client-issuer",
		Iat:  timeutil.NowUTC().Add(-time.Hour).Unix(),
		Exp:  timeutil.NowUTC().Add(time.Hour).Unix(),
		Aud:  []string{"test-client-audience"},
		Sub:  "test-client",
		Task: TestNtxTokenTask(),
		Permissions: []string{
			perm.UseRole("default"),
			perm.RunTask("test-task"),
		},
	}

	if err := tkn.Valid(); err != nil {
		panic(errors.Wrap(err, "ntx-tkn"))
	}

	return tkn
}

func TestNtxTokenTask() *NtxToken_Task {
	tkn := &NtxToken_Task{
		Id:    "test-task",
		Cfg:   scheduler.TestTaskConfiguration(),
		Role:  "default",
		Label: "test",
	}

	if err := tkn.Valid(); err != nil {
		panic(errors.Wrap(err, "ntx-tkn-task"))
	}

	return tkn
}

func TestUser() *User {
	u := &User{
		Id:          util.NewUUID(),
		Secret:      util.RandString(8),
		Email:       util.RandString(16) + "@email.com",
		Groups:      []string{"test-group"},
		Permissions: []string{perm.RunAnyTask()},
		Created:     timeutil.NowUTCUnixSeconds(),
		Expire:      timeutil.NowUTCUnixSeconds() + 3600,
	}
	if err := u.Valid(); err != nil {
		panic(errors.Wrap(err, "user"))
	}

	return u
}

func TestGroup() *Group {
	g := &Group{
		Id:          util.RandString(8),
		Permissions: []string{perm.RunAnyTask(), perm.SystemAdmin()},
	}
	if err := g.Valid(); err != nil {
		panic(errors.Wrap(err, "user"))
	}

	return g
}
