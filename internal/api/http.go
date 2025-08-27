package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"web_demoservice/internal/cache"
	"web_demoservice/internal/repo"
)

type Server struct {
	cache cache.Cache
	repo  repo.OrderRepository
	mux   *http.ServeMux
}

func NewServer(cache cache.Cache, repo repo.OrderRepository) *Server {
	s := &Server{cache: cache, repo: repo, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/order/", s.handleGetOrderByID)
}

func (s *Server) handleGetOrderByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 2 {
		http.Error(w, "use /api/order/{uid}", http.StatusBadRequest)
		return
	}
	uidStr := parts[1]
	log.Printf("INFO: GET /api/order/%s", uidStr)
	if o, ok := s.cache.Get(uidStr); ok {
		log.Printf("INFO: get from CACHE by uid:%s", uidStr)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(o)
		return
	}
	o, err := s.repo.GetOrderByUID(r.Context(), uidStr)
	log.Printf("INFO: get from DB by uid:%s", uidStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	s.cache.Set(uidStr, o)
	log.Printf("INFO: set to cache UID:%s", uidStr)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(o); err != nil {
		log.Println("ERROR: Json Encode Error:", err)
	}
}
