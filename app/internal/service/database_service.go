package service

import (
	"log"
	"sync"

	"br.tec.suportes/portal/internal/apierr"
	"br.tec.suportes/portal/internal/domain"
	"br.tec.suportes/portal/internal/repository"
	"github.com/jmoiron/sqlx"
)

type DatabaseService struct {
	repo        *repository.DatabaseRepo
	connections map[string]*sqlx.DB
	mu          sync.RWMutex
}

func NewDatabaseService(repo *repository.DatabaseRepo) *DatabaseService {
	svc := &DatabaseService{
		repo:        repo,
		connections: make(map[string]*sqlx.DB),
	}
	svc.initConnections()
	return svc
}

func (s *DatabaseService) initConnections() {
	all, err := s.repo.FindAll()
	if err != nil {
		log.Printf("[database-service] erro ao carregar conexões: %v", err)
		return
	}
	for _, d := range all {
		db, err := d.OpenConnection()
		if err != nil {
			log.Printf("[database-service] erro ao conectar '%s': %v", d.Key, err)
			continue
		}
		s.connections[d.Key] = db
		log.Printf("[database-service] conexão '%s' carregada", d.Key)
	}
}

func (s *DatabaseService) Create(req *domain.CreateDatabaseRequest) (*domain.DatabaseResponse, apierr.Detail) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if s.repo.ExistsByKey(req.Key) {
		return nil, apierr.Conflict("banco de dados com chave '" + req.Key + "' já existe")
	}

	d := req.ToDomain()
	db, openErr := d.OpenConnection()
	if openErr != nil {
		return nil, apierr.New("falha ao testar conexão: "+openErr.Error(), nil)
	}

	if err := s.repo.Create(d); err != nil {
		db.Close()
		return nil, apierr.New("falha ao persistir banco de dados: "+err.Error(), nil)
	}

	s.mu.Lock()
	s.connections[d.Key] = db
	s.mu.Unlock()

	return d.ToResponse(), nil
}

func (s *DatabaseService) List() ([]domain.DatabaseResponse, apierr.Detail) {
	all, err := s.repo.FindAll()
	if err != nil {
		return nil, apierr.New(err.Error(), nil)
	}
	result := make([]domain.DatabaseResponse, 0, len(all))
	for _, d := range all {
		result = append(result, *d.ToResponse())
	}
	return result, nil
}

func (s *DatabaseService) GetByID(id string) (*domain.DatabaseResponse, apierr.Detail) {
	d, err := s.repo.FindByID(id)
	if err != nil {
		return nil, apierr.NotFound(err.Error())
	}
	return d.ToResponse(), nil
}

func (s *DatabaseService) GetByKey(key string) (*domain.DatabaseResponse, apierr.Detail) {
	d, err := s.repo.FindByKey(key)
	if err != nil {
		return nil, apierr.NotFound(err.Error())
	}
	return d.ToResponse(), nil
}

func (s *DatabaseService) GetConnection(key string) (*sqlx.DB, apierr.Detail) {
	s.mu.RLock()
	db, ok := s.connections[key]
	s.mu.RUnlock()
	if !ok {
		return nil, apierr.NotFound("conexão '" + key + "' não encontrada")
	}
	return db, nil
}

func (s *DatabaseService) Delete(id string) apierr.Detail {
	d, err := s.repo.FindByID(id)
	if err != nil {
		return apierr.NotFound(err.Error())
	}
	s.closeConnection(d.Key)
	if err := s.repo.Delete(d); err != nil {
		return apierr.New(err.Error(), nil)
	}
	return nil
}

func (s *DatabaseService) DeleteByKey(key string) apierr.Detail {
	d, err := s.repo.FindByKey(key)
	if err != nil {
		return apierr.NotFound(err.Error())
	}
	s.closeConnection(d.Key)
	if err := s.repo.Delete(d); err != nil {
		return apierr.New(err.Error(), nil)
	}
	return nil
}

func (s *DatabaseService) closeConnection(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if db, ok := s.connections[key]; ok {
		db.Close()
		delete(s.connections, key)
	}
}

// ConnectionStatus retorna o status de cada conexão registrada.
func (s *DatabaseService) ConnectionStatus() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]string, len(s.connections))
	for key, db := range s.connections {
		if err := db.Ping(); err != nil {
			result[key] = "disconnected"
		} else {
			result[key] = "connected"
		}
	}
	return result
}
