package repository

import (
	"fmt"
	"strconv"
	"strings"

	"br.tec.suportes/portal/internal/domain"
	"br.tec.suportes/portal/internal/store"
)

type CommandRepo struct {
	db *store.DB
}

func NewCommandRepo(db *store.DB) *CommandRepo {
	return &CommandRepo{db: db}
}

func (r *CommandRepo) Create(c *domain.Command) error {
	return r.db.Storm.Save(c)
}

func (r *CommandRepo) Update(c *domain.Command) error {
	return r.db.Storm.Update(c)
}

func (r *CommandRepo) FindAll() ([]domain.Command, error) {
	var result []domain.Command
	if err := r.db.Storm.All(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *CommandRepo) FindByID(id string) (*domain.Command, error) {
	n, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("id inválido: %s", id)
	}
	var c domain.Command
	if err := r.db.Storm.One("ID", n, &c); err != nil {
		return nil, fmt.Errorf("command com id %d não encontrado", n)
	}
	return &c, nil
}

func (r *CommandRepo) FindByKey(key string) (*domain.Command, error) {
	var c domain.Command
	if err := r.db.Storm.One("Key", key, &c); err != nil {
		return nil, fmt.Errorf("command com key '%s' não encontrado", key)
	}
	return &c, nil
}

func (r *CommandRepo) FindByKeyPattern(pattern string) ([]domain.Command, error) {
	all, err := r.FindAll()
	if err != nil {
		return nil, err
	}
	var result []domain.Command
	p := strings.ToLower(pattern)
	for _, c := range all {
		if strings.Contains(strings.ToLower(c.Key), p) {
			result = append(result, c)
		}
	}
	return result, nil
}

func (r *CommandRepo) Delete(c *domain.Command) error {
	return r.db.Storm.DeleteStruct(c)
}

func (r *CommandRepo) DeleteByKey(key string) (int, error) {
	items, err := r.FindByKeyPattern(key)
	if err != nil {
		return 0, err
	}
	count := 0
	for i := range items {
		if err := r.db.Storm.DeleteStruct(&items[i]); err == nil {
			count++
		}
	}
	return count, nil
}

func (r *CommandRepo) ExistsByKey(key string) bool {
	var c domain.Command
	err := r.db.Storm.One("Key", key, &c)
	return err == nil && c.ID > 0
}
