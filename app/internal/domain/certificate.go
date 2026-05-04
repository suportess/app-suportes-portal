package domain

import (
	"fmt"
	"time"

	"br.tec.suportes/portal/internal/apierr"
)

// Certificate armazena os dados de um certificado TLS/mTLS.
type Certificate struct {
	ID             int    `storm:"id,increment" json:"id"`
	Name           string `storm:"unique,index" json:"nome"`
	Description    string `json:"descricao,omitempty"`
	CertFile       string `json:"arquivoCert"`
	KeyFile        string `json:"arquivoChave,omitempty"`
	PfxFile        string `json:"arquivoPfx,omitempty"`
	Password       string `json:"senha,omitempty"`
	CACertFile     string `json:"arquivoCACert,omitempty"`
	Active         bool   `json:"ativo"`
	ExpirationDate string `json:"dataExpiracao,omitempty"`
	CreatedAt      string `json:"criadoEm"`
	UpdatedAt      string `json:"atualizadoEm"`
}

// CreateCertificateRequest é o payload de criação de um certificado.
type CreateCertificateRequest struct {
	Name           string `json:"nome"`
	Description    string `json:"descricao,omitempty"`
	CertFile       string `json:"arquivoCert"`
	KeyFile        string `json:"arquivoChave,omitempty"`
	PfxFile        string `json:"arquivoPfx,omitempty"`
	Password       string `json:"senha,omitempty"`
	CACertFile     string `json:"arquivoCACert,omitempty"`
	ExpirationDate string `json:"dataExpiracao,omitempty"`
}

func (r *CreateCertificateRequest) Validate() apierr.Detail {
	if r.Name == "" {
		return apierr.New("campo 'nome' é obrigatório", nil)
	}
	if r.CertFile == "" && r.PfxFile == "" {
		return apierr.New("é necessário fornecer 'arquivoCert' (PEM base64) ou 'arquivoPfx' (PFX base64)", nil)
	}
	if r.ExpirationDate != "" {
		if _, err := time.Parse("2006-01-02", r.ExpirationDate); err != nil {
			return apierr.UnprocessableEntity(
				fmt.Sprintf("'dataExpiracao' deve estar no formato YYYY-MM-DD: %v", err),
				nil,
			)
		}
	}
	return nil
}

func (r *CreateCertificateRequest) ToDomain() *Certificate {
	now := time.Now().Format(time.RFC3339)
	return &Certificate{
		Name:           r.Name,
		Description:    r.Description,
		CertFile:       r.CertFile,
		KeyFile:        r.KeyFile,
		PfxFile:        r.PfxFile,
		Password:       r.Password,
		CACertFile:     r.CACertFile,
		Active:         true,
		ExpirationDate: r.ExpirationDate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// UpdateCertificateRequest é o payload de atualização de um certificado (nome imutável).
type UpdateCertificateRequest struct {
	Description    string `json:"descricao,omitempty"`
	CertFile       string `json:"arquivoCert,omitempty"`
	KeyFile        string `json:"arquivoChave,omitempty"`
	PfxFile        string `json:"arquivoPfx,omitempty"`
	Password       string `json:"senha,omitempty"`
	CACertFile     string `json:"arquivoCACert,omitempty"`
	Active         *bool  `json:"ativo,omitempty"`
	ExpirationDate string `json:"dataExpiracao,omitempty"`
}

func (r *UpdateCertificateRequest) Validate() apierr.Detail {
	if r.ExpirationDate != "" {
		if _, err := time.Parse("2006-01-02", r.ExpirationDate); err != nil {
			return apierr.UnprocessableEntity(
				fmt.Sprintf("'dataExpiracao' deve estar no formato YYYY-MM-DD: %v", err),
				nil,
			)
		}
	}
	return nil
}

func (r *UpdateCertificateRequest) ApplyTo(cert *Certificate) {
	if r.Description != "" {
		cert.Description = r.Description
	}
	if r.CertFile != "" {
		cert.CertFile = r.CertFile
	}
	if r.KeyFile != "" {
		cert.KeyFile = r.KeyFile
	}
	if r.PfxFile != "" {
		cert.PfxFile = r.PfxFile
	}
	if r.Password != "" {
		cert.Password = r.Password
	}
	if r.CACertFile != "" {
		cert.CACertFile = r.CACertFile
	}
	if r.Active != nil {
		cert.Active = *r.Active
	}
	if r.ExpirationDate != "" {
		cert.ExpirationDate = r.ExpirationDate
	}
	cert.UpdatedAt = time.Now().Format(time.RFC3339)
}
