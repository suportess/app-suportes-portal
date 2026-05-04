package repository

import (
	"fmt"
	"strconv"

	"br.tec.suportes/portal/internal/domain"
	"br.tec.suportes/portal/internal/store"
)

type CertificateRepo struct {
	db *store.DB
}

func NewCertificateRepo(db *store.DB) *CertificateRepo {
	return &CertificateRepo{db: db}
}

func (r *CertificateRepo) Create(c *domain.Certificate) error {
	return r.db.Storm.Save(c)
}

func (r *CertificateRepo) Update(c *domain.Certificate) error {
	return r.db.Storm.Update(c)
}

func (r *CertificateRepo) FindAll() ([]domain.Certificate, error) {
	var result []domain.Certificate
	if err := r.db.Storm.All(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *CertificateRepo) FindByID(id string) (*domain.Certificate, error) {
	n, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("id inválido: %s", id)
	}
	var c domain.Certificate
	if err := r.db.Storm.One("ID", n, &c); err != nil {
		return nil, fmt.Errorf("certificate com id %d não encontrado", n)
	}
	return &c, nil
}

func (r *CertificateRepo) FindByName(name string) (*domain.Certificate, error) {
	var c domain.Certificate
	if err := r.db.Storm.One("Name", name, &c); err != nil {
		return nil, fmt.Errorf("certificate com name '%s' não encontrado", name)
	}
	return &c, nil
}

func (r *CertificateRepo) Delete(c *domain.Certificate) error {
	return r.db.Storm.DeleteStruct(c)
}

func (r *CertificateRepo) ExistsByName(name string) bool {
	var c domain.Certificate
	err := r.db.Storm.One("Name", name, &c)
	return err == nil && c.ID > 0
}
