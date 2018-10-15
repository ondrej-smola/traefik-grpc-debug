package license

import (
	"net/http"

	"gitlab.com/nanotrix/nanogrid/pkg/ntx/httputil"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/middleware"
)

type (
	CreditServer struct {
		store           CreditStore
		tokenMiddleware middleware.Http
		log             log.Logger
	}
)

func NewCreditServer(store CreditStore, tknMiddleware middleware.Http, log log.Logger) *CreditServer {
	return &CreditServer{store: store, tokenMiddleware: tknMiddleware, log: log}
}

func (s *CreditServer) Router(r *mux.Router, pathPrefix string) *mux.Router {
	r.Methods("GET").Path(httputil.RemoveDuplicateSlashesFromPath(pathPrefix, "/usage")).
		Handler(s.tokenMiddleware(middleware.NtxTokenIsSystemAdmin()(http.HandlerFunc(s.currentCredit))))

	return r
}

func (s *CreditServer) currentCredit(w http.ResponseWriter, r *http.Request) {
	cr, err := s.store.Snapshot()

	if err != nil {
		s.log.Log("ev", "current_credit", "err", err)
		httputil.WriteErrorS(w, "unable to obtain clientsCache credit info", r.Header.Get(httputil.AcceptHeader), http.StatusServiceUnavailable)
	} else {
		httputil.WriteProtoMsg(w, cr, r.Header.Get(httputil.AcceptHeader), http.StatusOK)
	}
}
