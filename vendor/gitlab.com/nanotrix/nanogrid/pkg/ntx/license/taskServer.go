package license

import (
	"net/http"

	"gitlab.com/nanotrix/nanogrid/pkg/ntx/httputil"

	"strconv"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/auth"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/auth/perm"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/middleware"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/timeutil"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/version"
)

const (
	printWithoutTaskConfiguration = "no-task-config"
)

type (
	TaskServerConfig struct {
		Store      TaskStore
		Encoder    auth.Encoder
		Expiration Expiration

		SupportsCustomTasks bool

		TokenMiddleware middleware.Http
	}

	TaskServer struct {
		log log.Logger
		TaskServerConfig
	}

	TaskServerOpt func(a *TaskServer)

	writeErrFn func(w http.ResponseWriter, err error, contentType string, code int)
	writeTknFn func(w http.ResponseWriter, tknS string, expire int64, contentType string)
)

func (c TaskServerConfig) Valid() error {
	if c.Store == nil {
		return errors.New("store: is nil")
	}

	if c.Encoder == nil {
		return errors.New("encoder: is nil")
	}

	if c.Expiration == nil {
		return errors.New("expiration: is nil")
	}

	if c.TokenMiddleware == nil {
		return errors.New("tokenMiddleware: is nil")
	}

	return nil
}

func WithTaskServerLogger(l log.Logger) TaskServerOpt {
	return func(a *TaskServer) {
		a.log = l
	}
}

func NewTaskServer(cfg TaskServerConfig, opts ...TaskServerOpt) *TaskServer {
	if err := cfg.Valid(); err != nil {
		panic(errors.Wrap(err, "TaskServerConfig"))
	}

	s := &TaskServer{
		log:              log.NewNopLogger(),
		TaskServerConfig: cfg,
	}

	for _, o := range opts {
		o(s)
	}

	return s
}

func (s *TaskServer) Router(r *mux.Router, pathPrefix string) {
	r.Methods("GET").Path(pathPrefix).Handler(s.TokenMiddleware(http.HandlerFunc(s.tasksSnapshot)))

	// V0 API
	r.Methods("POST").Path(pathPrefix).Handler(
		s.TokenMiddleware(http.HandlerFunc(s.taskFromStore(s.writeNtxAuthTokenV0, s.writeErrFnV0))),
	)
	r.Methods("PUT").Path(pathPrefix).Handler(
		s.TokenMiddleware(http.HandlerFunc(s.customTask(s.writeNtxAuthTokenV0, s.writeErrFnV0))),
	)

	// V1 API
	r.Methods("POST").Path(httputil.RemoveDuplicateSlashesFromPath(pathPrefix, "/ntx-token")).Handler(
		s.TokenMiddleware(http.HandlerFunc(s.taskFromStore(s.writeNtxAuthTokenV1, s.writeErrFnV1))),
	)

	r.Methods("PUT").Path(httputil.RemoveDuplicateSlashesFromPath(pathPrefix, "/ntx-token")).Handler(
		s.TokenMiddleware(http.HandlerFunc(s.customTask(s.writeNtxAuthTokenV1, s.writeErrFnV1))),
	)
}

func (s *TaskServer) writeNtxAuthTokenV0(w http.ResponseWriter, tkn string, _ int64, contentType string) {
	t := &auth.NtxTokenResponse{
		Response: &auth.NtxTokenResponse_Token_{
			Token: &auth.NtxTokenResponse_Token{
				Token:     tkn,
				TokenType: auth.NtxTokenHeaderName,
			},
		},
	}

	httputil.WriteProtoMsg(w, t, contentType, http.StatusOK)
}

func (s *TaskServer) writeErrFnV0(w http.ResponseWriter, err error, contentType string, code int) {
	t := &auth.NtxTokenResponse{
		Response: &auth.NtxTokenResponse_Error_{
			Error: &auth.NtxTokenResponse_Error{
				Message: err.Error(),
			},
		},
	}

	httputil.WriteProtoMsg(w, t, contentType, http.StatusOK)
}

func (s *TaskServer) writeNtxAuthTokenV1(w http.ResponseWriter, tkn string, exp int64, contentType string) {
	httputil.WriteProtoMsg(w, &auth.NtxStoreToken{
		NtxToken:  tkn,
		ExpiresAt: uint64(exp),
	}, contentType, http.StatusOK)
}

func (s *TaskServer) writeErrFnV1(w http.ResponseWriter, err error, contentType string, code int) {
	httputil.WriteError(w, err, contentType, code)
}

func (s *TaskServer) writeTkn(w http.ResponseWriter, tkn string, exp int64, contentType string) {
	httputil.WriteProtoMsg(w, &auth.NtxStoreToken{
		NtxToken:  tkn,
		ExpiresAt: uint64(exp),
	}, contentType, http.StatusOK)
}

func (s *TaskServer) version(w http.ResponseWriter, r *http.Request) {
	httputil.WriteJson(w, version.VersionStringMap(), http.StatusOK)
}

func shouldPrintTaskConfiguration(r *http.Request) bool {
	if compactStr := r.URL.Query().Get(printWithoutTaskConfiguration); compactStr == "" {
		return true
	} else if ok, err := strconv.ParseBool(compactStr); err != nil {
		return true
	} else {
		return !ok
	}
}

func (s *TaskServer) tasksSnapshot(w http.ResponseWriter, r *http.Request) {
	if tkn, ok := auth.GetNtxTokenFromContext(r.Context()); !ok {
		httputil.WriteError(w, auth.ErrMissingNtxToken, r.Header.Get(httputil.AcceptHeader), http.StatusUnauthorized)
	} else if snapshot, err := s.Store.Snapshot(); err != nil {
		httputil.WriteError(w, err, r.Header.Get(httputil.AcceptHeader), http.StatusInternalServerError)
	} else {
		snapshot.Tasks = FilterTasks(snapshot.Tasks, perm.Parse(tkn.Permissions...))

		if !shouldPrintTaskConfiguration(r) {
			RemoveTaskConfigurations(snapshot.Tasks)
		}

		httputil.WriteProtoMsg(w, snapshot, r.Header.Get(httputil.AcceptHeader), http.StatusOK)
	}
}

func (s *TaskServer) taskFromStore(tknFn writeTknFn, errFn writeErrFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		query := &auth.NtxTokenRequest{}

		if err := util.ReadCloserToJsonPb(r.Body, query); err != nil {
			errFn(w, err, r.Header.Get(httputil.AcceptHeader), http.StatusBadRequest)
		} else if err := query.Valid(); err != nil {
			errFn(w, errors.Wrap(err, "request"), r.Header.Get(httputil.AcceptHeader), http.StatusBadRequest)
		} else if storeItem, err := s.Store.Get(query.Id); err != nil {
			if util.IsNotFound(err) {
				errFn(w, err, r.Header.Get(httputil.AcceptHeader), http.StatusNotFound)
			} else {
				errFn(w, err, r.Header.Get(httputil.AcceptHeader), http.StatusUnauthorized)
			}
		} else if !storeItem.ContainsLabel(query.Label) {
			errFn(w, errors.Errorf("task '%v' does not contain label '%v'", query.Id, query.Label), r.Header.Get(httputil.AcceptHeader), http.StatusBadRequest)
		} else if tkn, err := s.handleTaskRequest(r, &auth.NtxToken_Task{
			Id:    query.Id,
			Cfg:   storeItem.Task,
			Label: query.Label}); err != nil {
			errFn(w, err, r.Header.Get(httputil.AcceptHeader), http.StatusUnauthorized)
		} else if encodedToken, err := s.Encoder.Encode(tkn); err != nil {
			errFn(w, err, r.Header.Get(httputil.AcceptHeader), http.StatusInternalServerError)
		} else {
			s.log.Log(
				"ev", "existing_task",
				"addr", r.RemoteAddr,
				"tid", query.Id,
				"label", query.Label,
				"debug", true)

			tknFn(w, encodedToken, tkn.Exp, r.Header.Get(httputil.AcceptHeader))
		}
	}
}

func (s *TaskServer) customTask(tknFn writeTknFn, errFn writeErrFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		if !s.SupportsCustomTasks {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		query := &auth.CustomNtxTokenRequest{}

		if err := util.ReadCloserToJsonPb(r.Body, query); err != nil {
			errFn(w, err, r.Header.Get(httputil.AcceptHeader), http.StatusBadRequest)
		} else if err := query.Valid(); err != nil {
			errFn(w, errors.Wrap(err, "request"), r.Header.Get(httputil.AcceptHeader), http.StatusBadRequest)
		} else if tkn, err := s.handleTaskRequest(r, &auth.NtxToken_Task{
			Id:     "custom_task",
			Cfg:    query.Task,
			Label:  query.Label,
			Custom: true}); err != nil {
			errFn(w, err, r.Header.Get(httputil.AcceptHeader), http.StatusUnauthorized)
		} else if encodedToken, err := s.Encoder.Encode(tkn); err != nil {
			errFn(w, err, r.Header.Get(httputil.AcceptHeader), http.StatusInternalServerError)
		} else {
			s.log.Log(
				"ev", "custom_task",
				"addr", r.RemoteAddr,
				"label", query.Label,
				"debug", true)
			tknFn(w, encodedToken, tkn.Exp, r.Header.Get(httputil.AcceptHeader))
		}
	}
}

func (s *TaskServer) handleTaskRequest(r *http.Request, taskToAuth *auth.NtxToken_Task) (*auth.NtxToken, error) {
	ntxToken, ok := auth.GetNtxTokenFromContext(r.Context())
	if !ok {
		return nil, auth.ErrMissingNtxToken
	}

	taskToAuth.Role = perm.Parse(ntxToken.Permissions...).Role()

	if err := ntxToken.CanRun(taskToAuth); err != nil {
		return nil, err
	} else if s.Expiration.Expire() < timeutil.NowUTCUnixSeconds() {
		return nil, errors.Errorf("store license expired")

	} else if ntxToken = ntxToken.WithTask(taskToAuth).WithMaxExpire(s.Expiration.Expire()); ntxToken.Expired() {
		return nil, errors.Errorf("token expired '%v'", taskToAuth.Id)
	} else {
		return ntxToken, nil
	}
}
