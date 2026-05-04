package repository

import (
	"fmt"
	"strconv"
	"strings"

	"br.tec.suportes/portal/internal/domain"
	"br.tec.suportes/portal/internal/store"
)

type RouteRepo struct {
	db *store.DB
}

func NewRouteRepo(db *store.DB) *RouteRepo {
	return &RouteRepo{db: db}
}

func (r *RouteRepo) Create(route *domain.Route) error {
	return r.db.Storm.Save(route)
}

func (r *RouteRepo) FindAll() ([]domain.Route, error) {
	var result []domain.Route
	if err := r.db.Storm.All(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *RouteRepo) FindByID(id string) (*domain.Route, error) {
	n, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("id inválido: %s", id)
	}
	var route domain.Route
	if err := r.db.Storm.One("ID", n, &route); err != nil {
		return nil, fmt.Errorf("route com id %d não encontrada", n)
	}
	return &route, nil
}

func (r *RouteRepo) FindByKey(key string) (*domain.Route, error) {
	var route domain.Route
	if err := r.db.Storm.One("Key", key, &route); err != nil {
		return nil, fmt.Errorf("route com key '%s' não encontrada", key)
	}
	return &route, nil
}

func (r *RouteRepo) FindByKeyPattern(pattern string) ([]domain.Route, error) {
	all, err := r.FindAll()
	if err != nil {
		return nil, err
	}
	var result []domain.Route
	p := strings.ToLower(pattern)
	for _, route := range all {
		if strings.Contains(strings.ToLower(route.Key), p) {
			result = append(result, route)
		}
	}
	return result, nil
}

func (r *RouteRepo) FindByMethod(method string) ([]domain.Route, error) {
	all, err := r.FindAll()
	if err != nil {
		return nil, err
	}
	var result []domain.Route
	m := strings.ToUpper(method)
	for _, route := range all {
		if strings.ToUpper(string(route.Method)) == m {
			result = append(result, route)
		}
	}
	return result, nil
}

func (r *RouteRepo) FindByPath(path string) ([]domain.Route, error) {
	all, err := r.FindAll()
	if err != nil {
		return nil, err
	}
	var result []domain.Route
	for _, route := range all {
		if strings.Contains(route.Path, path) {
			result = append(result, route)
		}
	}
	return result, nil
}

func (r *RouteRepo) Delete(route *domain.Route) error {
	return r.db.Storm.DeleteStruct(route)
}

func (r *RouteRepo) DeleteByKey(key string) (int, error) {
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

func (r *RouteRepo) ExistsByKey(key string) bool {
	var route domain.Route
	err := r.db.Storm.One("Key", key, &route)
	return err == nil && route.ID > 0
}
