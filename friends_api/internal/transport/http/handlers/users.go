package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/callmetsoy/user_api/internal/domain"
	useCases "github.com/callmetsoy/user_api/internal/use_cases"
	"github.com/google/uuid"
)

type userUseCase interface {
	GetUsers(
		page int, pageSize int, filters map[string]interface{}, orderBy string, orderDir string,
	) (domain.PaginatedUsers, error)
	GetCommonFriends(userID1, userID2 uuid.UUID) ([]domain.User, error)
}

type UserHandler struct {
	userUseCase userUseCase
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		userUseCase: useCases.NewUserUseCase(),
	}
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	page := query.Get("page")
	if page == "" {
		page = "1"
	}
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pageSize := query.Get("page_size")
	if pageSize == "" {
		pageSize = "10"
	}
	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	allowedFilters := []string{"id", "name", "email", "gender", "birth_date"}
	filters := make(map[string]interface{})
	for _, field := range allowedFilters {
		if value := query.Get(field); value != "" {
			filters[field] = value
		}
	}

	orderBy := query.Get("order_by")
	if orderBy == "" {
		orderBy = "id"
	}

	orderDir := query.Get("order_dir")
	if orderDir == "" {
		orderDir = "asc"
	}

	users, err := h.userUseCase.GetUsers(pageInt, pageSizeInt, filters, orderBy, orderDir)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) GetCommonFriends(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	user1ID := query.Get("user1")
	user2ID := query.Get("user2")
	if user1ID == "" || user2ID == "" {
		http.Error(w, "user1 and user2 are required", http.StatusBadRequest)
		return
	}

	// Optional: validate UUID format
	user1UUID, err := uuid.Parse(user1ID)
	if err != nil {
		http.Error(w, "invalid user1 UUID", http.StatusBadRequest)
		return
	}
	user2UUID, err := uuid.Parse(user2ID)
	if err != nil {
		http.Error(w, "invalid user2 UUID", http.StatusBadRequest)
		return
	}

	// Call use case
	friends, err := h.userUseCase.GetCommonFriends(user1UUID, user2UUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(friends)
}
