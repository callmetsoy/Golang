package repository

import (
	"sync"
	"todo/internal/models"
)

func (r *TaskRepo) Delete(id int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tasks[id]; !ok {
		return false
	}
	delete(r.tasks, id)
	return true
}

type TaskRepo struct {
	mu     sync.RWMutex
	tasks  map[int]models.Task
	nextID int
}

func NewTaskRepo() *TaskRepo {
	return &TaskRepo{
		tasks:  make(map[int]models.Task),
		nextID: 1,
	}
}

func (r *TaskRepo) Create(title string) models.Task {
	r.mu.Lock()
	defer r.mu.Unlock()

	task := models.Task{
		ID:    r.nextID,
		Title: title,
		Done:  false,
	}
	r.tasks[task.ID] = task
	r.nextID++
	return task
}

func (r *TaskRepo) GetAll() []models.Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make([]models.Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		res = append(res, t)
	}
	return res
}

func (r *TaskRepo) GetByID(id int) (models.Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]
	return task, ok
}

func (r *TaskRepo) UpdateStatus(id int, done bool) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]
	if !ok {
		return false
	}
	task.Done = done
	r.tasks[id] = task
	return true
}
