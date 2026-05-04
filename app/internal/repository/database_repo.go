package repository

import (
	"fmt"
	"strconv"
	"strings"

	"br.tec.suportes/portal/internal/domain"
	"br.tec.suportes/portal/internal/store"
	"github.com/asdine/storm/v3"
)

type DatabaseRepo struct {
	db *store.DB
}

func NewDatabaseRepo(db *store.DB) *DatabaseRepo {
	return &DatabaseRepo{db: db}
}

func (r *DatabaseRepo) Create(d *domain.Database) error {
	return r.db.Storm.Save(d)
}

func (r *DatabaseRepo) FindAll() ([]domain.Database, error) {
	var result []domain.Database
	if err := r.db.Storm.All(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *DatabaseRepo) FindByID(id string) (*domain.Database, error) {
	n, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("id inválido: %s", id)
	}
	var d domain.Database
	if err := r.db.Storm.One("ID", n, &d); err != nil {
		return nil, fmt.Errorf("database com id %d não encontrado", n)
	}
	return &d, nil
}

func (r *DatabaseRepo) FindByKey(key string) (*domain.Database, error) {
	var d domain.Database
	if err := r.db.Storm.One("Key", key, &d); err != nil {
		return nil, fmt.Errorf("database com key '%s' não encontrada", key)
	}
	return &d, nil
}

func (r *DatabaseRepo) FindByKeyPattern(pattern string) ([]domain.Database, error) {
	all, err := r.FindAll()
	if err != nil {
		return nil, err
	}
	var result []domain.Database
	p := strings.ToLower(pattern)
	for _, d := range all {
		if strings.Contains(strings.ToLower(d.Key), p) {
			result = append(result, d)
		}
	}
	return result, nil
}

func (r *DatabaseRepo) Delete(d *domain.Database) error {
	return r.db.Storm.DeleteStruct(d)
}

func (r *DatabaseRepo) ExistsByKey(key string) bool {
	var d domain.Database
	err := r.db.Storm.One("Key", key, &d)
	return err == nil && d.ID > 0
}

// stormNotFound returns true when the error is storm.ErrNotFound.
func stormNotFound(err error) bool {
	return err == storm.ErrNotFound
}
