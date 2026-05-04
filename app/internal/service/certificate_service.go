package service

import (
	"log"

	"br.tec.suportes/portal/internal/apierr"
	"br.tec.suportes/portal/internal/domain"
	"br.tec.suportes/portal/internal/repository"
)

type CertificateService struct {
	repo *repository.CertificateRepo
}

func NewCertificateService(repo *repository.CertificateRepo) *CertificateService {
	return &CertificateService{repo: repo}
}

func (s *CertificateService) Create(req *domain.CreateCertificateRequest) (*domain.Certificate, apierr.Detail) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if s.repo.ExistsByName(req.Name) {
		return nil, apierr.Conflict("certificado com nome '" + req.Name + "' já existe")
	}
	cert := req.ToDomain()
	if err := s.repo.Create(cert); err != nil {
		return nil, apierr.New("falha ao persistir certificado: "+err.Error(), nil)
	}
	log.Printf("[certificate-service] certificado '%s' criado", cert.Name)
	return cert, nil
}

func (s *CertificateService) Update(id string, req *domain.UpdateCertificateRequest) (*domain.Certificate, apierr.Detail) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	cert, err := s.repo.FindByID(id)
	if err != nil {
		return nil, apierr.NotFound(err.Error())
	}
	req.ApplyTo(cert)
	if err := s.repo.Update(cert); err != nil {
		return nil, apierr.New("falha ao atualizar certificado: "+err.Error(), nil)
	}
	return cert, nil
}

func (s *CertificateService) List() ([]domain.Certificate, apierr.Detail) {
	all, err := s.repo.FindAll()
	if err != nil {
		return nil, apierr.New(err.Error(), nil)
	}
	return all, nil
}

func (s *CertificateService) GetByID(id string) (*domain.Certificate, apierr.Detail) {
	cert, err := s.repo.FindByID(id)
	if err != nil {
		return nil, apierr.NotFound(err.Error())
	}
	return cert, nil
}

func (s *CertificateService) GetByName(name string) (*domain.Certificate, apierr.Detail) {
	cert, err := s.repo.FindByName(name)
	if err != nil {
		return nil, apierr.NotFound(err.Error())
	}
	return cert, nil
}

func (s *CertificateService) Delete(id string) apierr.Detail {
	cert, err := s.repo.FindByID(id)
	if err != nil {
		return apierr.NotFound(err.Error())
	}
	if err := s.repo.Delete(cert); err != nil {
		return apierr.New(err.Error(), nil)
	}
	log.Printf("[certificate-service] certificado '%s' removido", cert.Name)
	return nil
}

func (s *CertificateService) DeleteByName(name string) apierr.Detail {
	cert, err := s.repo.FindByName(name)
	if err != nil {
		return apierr.NotFound(err.Error())
	}
	if err := s.repo.Delete(cert); err != nil {
		return apierr.New(err.Error(), nil)
	}
	return nil
}
