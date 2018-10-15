package license

import (
	"net/http"

	"gitlab.com/nanotrix/nanogrid/pkg/ntx/httputil"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/middleware"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
)

type LocalServer struct {
	store           Store
	tokenMiddleware middleware.Http
	log             log.Logger
}

type LocalServerOpt func(m *LocalServer)

func WithLocalServerLogger(log log.Logger) LocalServerOpt {
	return func(m *LocalServer) {
		m.log = log
	}
}

func WithLocalServerMiddleware(middleware middleware.Http) LocalServerOpt {
	return func(m *LocalServer) {
		m.tokenMiddleware = middleware
	}
}

func NewLocalServer(m Store, opts ...LocalServerOpt) *LocalServer {
	s := &LocalServer{
		store:           m,
		tokenMiddleware: middleware.None(),
		log:             log.NewNopLogger(),
	}

	for _, o := range opts {
		o(s)
	}

	return s
}

func (l *LocalServer) Router(r *mux.Router, pathPrefix string) *mux.Router {
	r.Methods("GET").
		Path(httputil.RemoveDuplicateSlashesFromPath(pathPrefix, "/info")).
		Handler(l.tokenMiddleware(middleware.NtxTokenIsSystemAdmin()(http.HandlerFunc(l.info))))

	r.Methods("HEAD").
		Path(pathPrefix).
		Handler(l.tokenMiddleware(http.HandlerFunc(l.licenseStateHeader)))

	r.Methods("GET").
		Path(httputil.RemoveDuplicateSlashesFromPath(pathPrefix, "/export")).
		Handler(l.tokenMiddleware(middleware.NtxTokenIsSystemAdmin()(http.HandlerFunc(l.export))))

	r.Methods("GET").
		Path(pathPrefix).
		Handler(l.tokenMiddleware(middleware.NtxTokenIsSystemAdmin()(http.HandlerFunc(l.newLicense))))

	r.Methods("POST").
		Path(pathPrefix).
		Handler(l.tokenMiddleware(middleware.NtxTokenIsSystemAdmin()(http.HandlerFunc(l.activate))))

	return r
}

func (l *LocalServer) info(w http.ResponseWriter, r *http.Request) {
	if lic, err := l.store.License(); err != nil {
		if err == ErrNoLicenseFound {
			httputil.WriteProtoMsg(w, &LicenseInfo{State: LicenseInfo_WAITING}, r.Header.Get(httputil.AcceptHeader), http.StatusOK)
		} else {
			l.log.Log("ev", "get_info", "err", err)
			httputil.WriteErrorS(w, "unable to obtain license info", r.Header.Get(httputil.AcceptHeader), http.StatusInternalServerError)
		}
	} else {
		httputil.WriteProtoMsg(w, InfoFromLicense(lic), r.Header.Get(httputil.AcceptHeader), http.StatusOK)
	}
}

func (l *LocalServer) licenseStateHeader(w http.ResponseWriter, r *http.Request) {
	if lic, err := l.store.License(); err != nil {
		if err == ErrNoLicenseFound {
			w.WriteHeader(http.StatusPaymentRequired)
		} else {
			l.log.Log("ev", "get_state_header", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else if !lic.Expired() {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusPaymentRequired)
	}
}

func (l *LocalServer) export(w http.ResponseWriter, r *http.Request) {
	if lic, err := l.store.License(); err != nil {
		if err == ErrNoLicenseFound {
			httputil.WriteProtoMsg(w, &LicenseInfo{State: LicenseInfo_WAITING}, r.Header.Get(httputil.AcceptHeader), http.StatusOK)
		} else {
			l.log.Log("ev", "export", "err", err)
			httputil.WriteErrorS(w, "unable to obtain license", r.Header.Get(httputil.AcceptHeader), http.StatusInternalServerError)
		}
	} else {
		httputil.WriteProtoMsg(w, lic, r.Header.Get(httputil.AcceptHeader), http.StatusOK)
	}
}

func (l *LocalServer) newLicense(w http.ResponseWriter, r *http.Request) {
	if challenge, err := l.store.Challenge(); err != nil {
		l.log.Log("ev", "net_license", "err", err)
		httputil.WriteErrorS(w, "unable to generate challenge", r.Header.Get(httputil.AcceptHeader), http.StatusInternalServerError)
	} else {
		httputil.WriteProtoMsg(w, challenge, r.Header.Get(httputil.AcceptHeader), http.StatusOK)
	}
}

func (l *LocalServer) activate(w http.ResponseWriter, r *http.Request) {
	lic := &LicenseResponse{}

	if err := util.ReadCloserToJsonPb(r.Body, lic); err != nil {
		httputil.WriteError(w, err, r.Header.Get(httputil.AcceptHeader), http.StatusBadRequest)
		return
	}

	if err := l.store.Activate(lic); err != nil {
		l.log.Log("ev", "activate", "err", err)
		httputil.WriteErrorS(w, "invalid license", r.Header.Get(httputil.AcceptHeader), http.StatusForbidden)
	}

	// return cluster info
	l.info(w, r)
}
