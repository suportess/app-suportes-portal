package service

import (
	"log"

	"br.tec.suportes/portal/internal/apierr"
	"br.tec.suportes/portal/internal/domain"
	"br.tec.suportes/portal/internal/enum"
	"br.tec.suportes/portal/internal/repository"
	"github.com/xwb1989/sqlparser"
)

type CommandService struct {
	repo *repository.CommandRepo
}

func NewCommandService(repo *repository.CommandRepo) *CommandService {
	return &CommandService{repo: repo}
}

func (s *CommandService) Create(req *domain.CreateCommandRequest) (*domain.Command, apierr.Detail) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if s.repo.ExistsByKey(req.Key) {
		return nil, apierr.Conflict("comando com chave '" + req.Key + "' já existe")
	}

	cmd := req.ToDomain()

	if cmd.Type.IsSQL() && cmd.Type != enum.CommandTypeProcedure && cmd.Type != enum.CommandTypeAnonymous {
		if err := validateSQL(cmd.SQL); err != nil {
			return nil, apierr.UnprocessableEntity("SQL inválido: "+err.Error(), nil)
		}
	}

	cmd.ProcessParameters()

	if err := s.repo.Create(cmd); err != nil {
		return nil, apierr.New("falha ao persistir comando: "+err.Error(), nil)
	}
	log.Printf("[command-service] comando '%s' criado (tipo=%s)", cmd.Key, cmd.Type)
	return cmd, nil
}

func (s *CommandService) Update(id string, req *domain.UpdateCommandRequest) (*domain.Command, apierr.Detail) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	cmd, err := s.repo.FindByID(id)
	if err != nil {
		return nil, apierr.NotFound(err.Error())
	}

	req.ApplyTo(cmd)

	if cmd.Type.IsSQL() && cmd.Type != enum.CommandTypeProcedure && cmd.Type != enum.CommandTypeAnonymous && cmd.SQL != "" {
		if err := validateSQL(cmd.SQL); err != nil {
			return nil, apierr.UnprocessableEntity("SQL inválido: "+err.Error(), nil)
		}
	}

	cmd.ProcessParameters()

	if err := s.repo.Update(cmd); err != nil {
		return nil, apierr.New("falha ao atualizar comando: "+err.Error(), nil)
	}
	log.Printf("[command-service] comando '%s' atualizado", cmd.Key)
	return cmd, nil
}

func (s *CommandService) List() ([]domain.Command, apierr.Detail) {
	all, err := s.repo.FindAll()
	if err != nil {
		return nil, apierr.New(err.Error(), nil)
	}
	return all, nil
}

func (s *CommandService) GetByID(id string) (*domain.Command, apierr.Detail) {
	cmd, err := s.repo.FindByID(id)
	if err != nil {
		return nil, apierr.NotFound(err.Error())
	}
	return cmd, nil
}

func (s *CommandService) GetByKey(key string) (*domain.Command, apierr.Detail) {
	cmd, err := s.repo.FindByKey(key)
	if err != nil {
		return nil, apierr.NotFound(err.Error())
	}
	return cmd, nil
}

func (s *CommandService) FindByKeyPattern(pattern string) ([]domain.Command, apierr.Detail) {
	result, err := s.repo.FindByKeyPattern(pattern)
	if err != nil {
		return nil, apierr.New(err.Error(), nil)
	}
	return result, nil
}

func (s *CommandService) Delete(id string) apierr.Detail {
	cmd, err := s.repo.FindByID(id)
	if err != nil {
		return apierr.NotFound(err.Error())
	}
	if err := s.repo.Delete(cmd); err != nil {
		return apierr.New(err.Error(), nil)
	}
	return nil
}

func (s *CommandService) DeleteByKey(key string) (int, apierr.Detail) {
	count, err := s.repo.DeleteByKey(key)
	if err != nil {
		return 0, apierr.New(err.Error(), nil)
	}
	if count == 0 {
		return 0, apierr.NotFound("nenhum comando encontrado com chave: " + key)
	}
	return count, nil
}

func validateSQL(sql string) error {
	_, err := sqlparser.Parse(sql)
	return err
}
