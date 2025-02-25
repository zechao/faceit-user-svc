package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/zechao/faceit-user-svc/user"
)

type Handler struct {
	service user.Service
}

func NewHandler(service user.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/users", h.handlerCreateUser).Methods("POST")
	router.HandleFunc("/users/{id}", h.handlerUpdateUser).Methods("PATH")
	router.HandleFunc("/users/{id}", h.handlerDeleteUser).Methods("DELETE")
	router.HandleFunc("/users", h.handlerListUsers).Methods("GET")
}

func (h *Handler) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "failed to decode request body", http.StatusBadRequest)
		return
	}

	u := user.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		NickName:  req.NickName,
		Email:     req.Email,
		Country:   req.Country,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	res, err := h.service.CreateUser(r.Context(), &u)
	if err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *Handler) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) handlerDeleteUser(w http.ResponseWriter, r *http.Request) {
}

func (h *Handler) handlerListUsers(w http.ResponseWriter, r *http.Request) {

}

// CreateUserRequest represents the request format for creating a new user.
type CreateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	NickName  string `json:"nick_name"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	Country   string `json:"country"`
}

// UserResponse represents the response format for a user. It doesn't include password, since is a sensitive information
type UserResponse struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	NickName  string    `json:"nick_name"`
	Email     string    `json:"email"`
	Country   string    `json:"country"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListUsersResponse represents the response format for a list of users with its query parameters.
type ListUsersResponse struct {
	Page         int                 `json:"page"`
	PageSize     int                 `json:"page_size"`
	TotalRecords int                 `json:"total_records"`
	SortBy       string              `json:"sort_by"`
	SortOrder    string              `json:"sort_order"`
	Filters      map[string][]string `json:"filters"`
	Users        []UserResponse      `json:"users"`
}
