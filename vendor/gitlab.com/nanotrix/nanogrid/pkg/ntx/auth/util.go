package auth

import (
	"time"

	"github.com/golang/protobuf/proto"

	"math"

	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/auth/perm"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/timeutil"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
)

type ProviderConfigs []*AuthProviderConfig

func (t *OAuthToken) Valid() error {
	if t == nil {
		return errors.New("is nil")
	}

	if t.AccessToken == "" {
		return errors.New("access_token: is blank")
	}

	if t.TokenType == "" {
		return errors.New("token_type: is blank")
	}

	return nil
}

func (t *OAuthToken) Clone() *OAuthToken {
	return &OAuthToken{
		AccessToken: t.AccessToken,
		TokenType:   t.TokenType,
	}
}

func (c *OAuthClientCredentials) Clone() *OAuthClientCredentials {
	return proto.Clone(c).(*OAuthClientCredentials)
}

func (t *NtxToken) Clone() *NtxToken {
	return &NtxToken{
		Iss:         t.Iss,
		Iat:         t.Iat,
		Nbf:         t.Nbf,
		Exp:         t.Exp,
		Aud:         util.CloneStringSlice(t.Aud),
		Sub:         t.Sub,
		Jti:         t.Jti,
		Permissions: util.CloneStringSlice(t.Permissions),
		Email:       t.Email,
		Task:        t.Task.Clone(),
	}
}

func (t *AccessToken) Clone() *AccessToken {
	return &AccessToken{
		AccessToken: t.AccessToken,
		ExpiresAt:   t.ExpiresAt,
	}
}

func (c *NtxToken) HasPermission(perm string) bool {
	for _, s := range c.Permissions {
		if s == perm {
			return true
		}
	}

	return false
}

func (c *NtxToken) SetAudience(aud ...string) {
	c.Aud = aud
}

func (t *NtxToken) FirstAudience() string {
	if len(t.Aud) == 0 {
		return ""
	} else {
		return t.Aud[0]
	}
}

func (c *NtxToken) Issuer() string {
	return c.Iss
}

func (t *NtxToken) WithTask(task *NtxToken_Task) *NtxToken {
	t.Task = task
	return t
}

func (t *NtxToken) WithMaxExpire(maxUnix uint64) *NtxToken {
	if uint64(t.Exp) > maxUnix {
		if maxUnix > math.MaxInt64 {
			maxUnix = math.MaxInt64
		}

		t.Exp = int64(maxUnix)
	}

	return t
}

func (t *NtxToken) Expired() bool {
	return t.Exp < int64(timeutil.NowUTCUnixSeconds())
}

func (t *NtxToken) Expire() time.Time {
	return time.Unix(t.Exp, 0)
}

const NtxTokenTimeSkewToleranceSeconds = int64(60)

func (c *NtxToken) Valid() error {
	now := timeutil.NowUTC()
	// add minute tolerance

	nowUnix := now.Unix()

	if len(c.FirstAudience()) == 0 {
		return errors.New("audience: empty")
	}

	if c.Issuer() == "" {
		return errors.New("issuer: blank")
	}

	if c.Task != nil {
		if err := c.Task.Valid(); err != nil {
			return errors.Wrap(err, "task")
		}

		if err := c.CanRun(c.Task); err != nil {
			return errors.Wrap(err, "task")
		}
	}

	if c.Exp != 0 && (c.Exp+NtxTokenTimeSkewToleranceSeconds) < nowUnix {
		return errors.Errorf("token: expired at '%v'", timeutil.FormatTime(time.Unix(c.Exp, 0)))
	}

	if c.Iat != 0 && (c.Iat-NtxTokenTimeSkewToleranceSeconds) > nowUnix {
		return errors.Errorf("token used before issued, issued at '%v', used at '%v'",
			timeutil.FormatTime(time.Unix(c.Iat, 0)),
			timeutil.FormatTime(now))
	}

	if c.Nbf != 0 && (c.Nbf-NtxTokenTimeSkewToleranceSeconds) > nowUnix {
		return errors.Errorf("token is not valid yet, valid after '%v', used at '%v'",
			timeutil.FormatTime(time.Unix(c.Nbf, 0)),
			timeutil.FormatTime(now))
	}

	return nil
}

func (t *NtxToken) CanRun(taskToAuth *NtxToken_Task) error {
	perms := perm.Parse(t.Permissions...)

	if !perms.UseRole(taskToAuth.Role) {
		return errors.Errorf("not allowed to use role '%v'", taskToAuth.Role)
	} else if taskToAuth.Custom && !perms.CreateTask() {
		return errors.New("not allowed to create task")
	} else if !taskToAuth.Custom && !perms.RunTask(taskToAuth.Id) {
		return errors.Errorf("not allowed to run task '%v'", taskToAuth.Id)
	} else {
		return nil
	}
}

func (t *NtxToken_Task) Valid() error {
	if t == nil {
		return errors.New("is nil")
	}

	if t.Id == "" {
		return errors.New("id: is blank")
	}

	if t.Label == "" {
		return errors.New("label: is blank")
	}

	if t.Role == "" {
		return errors.New("role: is blank")
	}

	if err := t.Cfg.Valid(); err != nil {
		return errors.Wrap(err, "cfg")
	}

	return nil
}

func (t *NtxToken_Task) Clone() *NtxToken_Task {
	return &NtxToken_Task{
		Id:     t.Id,
		Cfg:    t.Cfg.Clone(),
		Label:  t.Label,
		Role:   t.Role,
		Custom: t.Custom,
	}
}

func (c *AuthProviderConfig) Valid() error {
	if c == nil {
		return errors.New("is nil")
	}

	if c.Issuer == "" {
		return errors.New("issuer: is empty")
	}

	if c.Hs256 != nil {
		if err := c.Hs256.Valid(); err != nil {
			return errors.Wrap(err, "hs256")
		}
	} else if c.Rs256 != nil {
		if err := c.Rs256.Valid(); err != nil {
			return errors.Wrap(err, "rs256")
		}
	} else {
		return errors.Errorf("either hs256 or rs256 must be set")
	}

	return nil
}

func (c *AuthProviderConfig_RS256) Valid() error {
	if c == nil {
		return errors.New("is nil")
	}

	if c.Input == "" {
		return errors.New("input: is empty")
	}

	return nil
}

func (c *AuthProviderConfig_HS256) Valid() error {
	if c == nil {
		return errors.New("is nil")
	}

	if c.Secret == "" {
		return errors.New("secret: is empty")
	}

	return nil
}

func (m *NtxTokenRequest) Valid() error {
	if m.Id == "" {
		return errors.New("missing 'id'")
	}
	if m.Label == "" {
		return errors.New("missing 'label'")
	}

	return nil
}

func (m *CustomNtxTokenRequest) Valid() error {

	if m.Label == "" {
		return errors.New("taskLabel missing")
	}

	if err := m.Task.Valid(); err != nil {
		return errors.Wrap(err, "task configuration")
	}

	return nil
}

func (c ProviderConfigs) Valid() error {
	for i, a := range c {
		if err := a.Valid(); err != nil {
			return errors.Wrapf(err, "providers[%v]", i)
		}
	}

	return nil
}

func (r *GetAccessToken) Valid() error {
	if r == nil {
		return errors.New("is nil")
	}

	if r.Username == "" {
		return errors.New("username: is blank")
	}

	if r.Password == "" {
		return errors.New("password: is blank")
	}

	return nil
}

func (r *OAuthResourceOwnerCredentials) Valid() error {
	if r == nil {
		return errors.New("is nil")
	}

	if r.Scope == "" {
		return errors.New("scope: is blank")
	}

	if r.Audience == "" {
		return errors.New("audience: is blank")
	}

	if r.GrantType == "" {
		return errors.New("grantType: is blank")
	}

	if r.ClientId == "" {
		return errors.New("clientId: is blank")
	}

	if r.Username == "" {
		return errors.New("username: is blank")
	}

	if r.Password == "" {
		return errors.New("password: is blank")
	}

	return nil
}

func (o *OAuthResourceOwnerCredentials) Clone() *OAuthResourceOwnerCredentials {
	cpy := *o
	return &cpy
}
