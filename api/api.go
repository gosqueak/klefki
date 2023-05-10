package api

// Allows for 2 users to form shared secrets.
// the following chart represents the flow for creating a shared secret between
// A and B.
// Alice -------------------------database------------------------- Bob
//    # Alice Initiates and stores an
//       exchange with her her key and id.
// |POST xchange request ----------->x
//
//    # Alice uses message system to notify Bob of the exchange id
//    # Bob stores his id and public key in the exchange
//       and the server responds with Alice's public key.
//                                    x<-------------- UPDATE xchange|
//    # Bob derives shared secret.
//    # Bob uses message system to notify Alice of xchange update.
//    # Alice downloads Bob's public key.
// | GET Bob's public key from xchange
//    # Server deletes the xchange
//    # Alice derives shared secret.
//

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	kit "github.com/gosqueak/apikit"
	"github.com/gosqueak/jwt"
	"github.com/gosqueak/klefki/database"
)

type Server struct {
	db   *sql.DB
	addr string
	aud  jwt.Audience
}

func NewServer(addr string, db *sql.DB, aud jwt.Audience) *Server {
	return &Server{db, addr, aud}
}

func (s *Server) Run() {
	http.HandleFunc(
		"/", kit.LogMiddleware(
			kit.CorsMiddleware(
				kit.CookieTokenMiddleware(
					s.aud.Name, s.aud, s.handleExchange,
				),
			),
		),
	)
	// start serving
	log.Fatal(http.ListenAndServe(s.addr, nil))
}

func (s *Server) handleExchange(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleUserAKey(w, r)
	case http.MethodPatch:
		s.handleUserBKey(w, r)
	case http.MethodGet:
		s.handleFinishExchange(w, r)
	}
}

// This handler is for A to start an exchange with her public key.
func (s *Server) handleUserAKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		B64KeyUserA string `json:"b64KeyUserA"`
	}
	var resp struct {
		ExchangeUuid string `json:"exchangeUuid"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		kit.ErrBadRequest(w)
		return
	}

	exchangeUuid, err := database.MakeNewExchange(s.db, req.B64KeyUserA)

	if err != nil {
		panic(err)
	}

	resp.ExchangeUuid = exchangeUuid

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		kit.ErrInternal(w)
		return
	}
}

// handler for updating an exchange with B's public key.
func (s *Server) handleUserBKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ExchangeUuid string `json:"exchangeUuid"`
		B64KeyUserB  string `json:"b64KeyUserB"`
	}
	var resp struct {
		B64KeyUserA string `json:"b64KeyUserA"`
	}

	keyUserA, err := database.UserBSwapKey(s.db, req.ExchangeUuid, req.B64KeyUserB)

	if err != nil {
		panic(err)
	}

	resp.B64KeyUserA = keyUserA

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		kit.ErrInternal(w)
		return
	}
}

// handler for A to download B's key and finally delete the exchange.
func (s *Server) handleFinishExchange(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ExchangeUuid string `json:"exchangeUuid"`
	}
	var resp struct {
		B64KeyUserB string `json:"b64KeyUserB"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		kit.ErrBadRequest(w)
		return
	}

	keyUserB, err := database.FinishExchange(s.db, req.ExchangeUuid)

	if err != nil {
		panic(err)
	}

	resp.B64KeyUserB = keyUserB

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		kit.ErrInternal(w)
		return
	}
}
