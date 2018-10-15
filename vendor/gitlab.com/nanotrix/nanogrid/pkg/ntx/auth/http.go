package auth

import (
	"fmt"

	"github.com/pkg/errors"
)

const NtxTokenHeaderName = "ntx-token"

// returned when no token is found
var ErrMissingNtxToken = errors.New(fmt.Sprintf("missing header: %v", NtxTokenHeaderName))
var ErrMissingNtxTokenTask = errors.New("ntx-token: no task set")
