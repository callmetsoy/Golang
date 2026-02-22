package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"todo/internal/models"
	"todo/internal/repository"
)

type TaskHandler struct {
	Repo *repository.TaskRepo
}

func (h *TaskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		h.getTasks(w, r)
	case http.MethodPost:
		h.createTask(w, r)
	case http.MethodPatch:
		h.updateTask(w, r)
	case http.MethodDelete:
		h.deleteTask(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *TaskHandler) getTasks(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	doneFilter := r.URL.Query().Get("done") // Для фильтрации

	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid id"})
			return
		}
		task, ok := h.Repo.GetByID(id)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(models.ErrorResponse{Error: "task not found"})
			return
		}
		json.NewEncoder(w).Encode(task)
		return
	}

	tasks := h.Repo.GetAll()

	if doneFilter != "" {
		isDone := doneFilter == "true"
		filtered := []models.Task{}
		for _, t := range tasks {
			if t.Done == isDone {
				filtered = append(filtered, t)
			}
		}
		json.NewEncoder(w).Encode(filtered)
		return
	}

	json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) deleteTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid id"})
		return
	}

	if ok := h.Repo.Delete(id); !ok {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "task not found"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"deleted": true})
}

func (h *TaskHandler) createTask(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid title"})
		return
	}

	task := h.Repo.Create(input.Title)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) updateTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid id"})
		return
	}

	var input struct {
		Done *bool `json:"done"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Done == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid status"})
		return
	}

	if ok := h.Repo.UpdateStatus(id, *input.Done); !ok {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "task not found"})
		return
	}

	json.NewEncoder(w).Encode(models.SuccessUpdate{Updated: true})
}
