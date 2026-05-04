package domain

import (
	"context"
	"fmt"
	"log"
	"time"

	"br.tec.suportes/portal/internal/apierr"
	"br.tec.suportes/portal/internal/enum"
	"github.com/jmoiron/sqlx"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/godror/godror"
	_ "github.com/lib/pq"
	_ "github.com/microsoft/go-mssqldb"
)

// Database é a entidade persistida no BoltDB representando uma conexão configurada.
type Database struct {
	ID            int           `storm:"id,increment" json:"id"`
	Key           string        `storm:"unique,index" json:"key"`
	Driver        enum.DBDriver `json:"driver"`
	Host          string        `json:"host"`
	Port          int           `json:"port"`
	DBName        string        `json:"dbName"`
	ServiceName   string        `json:"serviceName,omitempty"`
	SID           string        `json:"sid,omitempty"`
	ConnectString string        `json:"connectString,omitempty"`
	User          string        `json:"user"`
	Password      string        `json:"password,omitempty"`
	Pool          DBPool        `json:"pool"`
}

type DBPool struct {
	MaxOpenConns    int `json:"maxOpenConns"`
	MaxIdleConns    int `json:"maxIdleConns"`
	ConnMaxLifetime int `json:"connMaxLifetimeSec"`
	ConnMaxIdleTime int `json:"connMaxIdleTimeSec"`
}

// CreateDatabaseRequest é o payload de criação de uma conexão.
type CreateDatabaseRequest struct {
	Key           string        `json:"key"`
	Driver        enum.DBDriver `json:"driver"`
	Host          string        `json:"host"`
	Port          int           `json:"port"`
	DBName        string        `json:"dbName"`
	ServiceName   string        `json:"serviceName,omitempty"`
	SID           string        `json:"sid,omitempty"`
	ConnectString string        `json:"connectString,omitempty"`
	User          string        `json:"user"`
	Password      string        `json:"password"`
	Pool          DBPool        `json:"pool"`
}

func (r *CreateDatabaseRequest) Validate() apierr.Detail {
	if r.Key == "" {
		return apierr.New("campo 'key' é obrigatório", nil)
	}
	if !r.Driver.IsValid() {
		return apierr.UnprocessableEntity(
			fmt.Sprintf("driver '%s' inválido. Valores aceitos: %v", r.Driver, enum.ValidDBDrivers()),
			nil,
		)
	}
	if r.Host == "" {
		return apierr.New("campo 'host' é obrigatório", nil)
	}
	if r.Port <= 0 || r.Port > 65535 {
		return apierr.New("campo 'port' deve ser um número entre 1 e 65535", nil)
	}
	if r.DBName == "" {
		return apierr.New("campo 'dbName' é obrigatório", nil)
	}
	if r.User == "" {
		return apierr.New("campo 'user' é obrigatório", nil)
	}
	if r.Password == "" {
		return apierr.New("campo 'password' é obrigatório", nil)
	}
	return nil
}

func (r *CreateDatabaseRequest) ToDomain() *Database {
	pool := r.Pool
	if pool.MaxOpenConns <= 0 {
		pool.MaxOpenConns = 10
	}
	if pool.MaxIdleConns <= 0 {
		pool.MaxIdleConns = 5
	}
	if pool.ConnMaxLifetime <= 0 {
		pool.ConnMaxLifetime = 90
	}
	if pool.ConnMaxIdleTime <= 0 {
		pool.ConnMaxIdleTime = 30
	}
	return &Database{
		Key:           r.Key,
		Driver:        r.Driver,
		Host:          r.Host,
		Port:          r.Port,
		DBName:        r.DBName,
		ServiceName:   r.ServiceName,
		SID:           r.SID,
		ConnectString: r.ConnectString,
		User:          r.User,
		Password:      r.Password,
		Pool:          pool,
	}
}

// DatabaseResponse é a resposta pública sem a senha.
type DatabaseResponse struct {
	ID            int           `json:"id"`
	Key           string        `json:"key"`
	Driver        enum.DBDriver `json:"driver"`
	Host          string        `json:"host"`
	Port          int           `json:"port"`
	DBName        string        `json:"dbName"`
	ServiceName   string        `json:"serviceName,omitempty"`
	SID           string        `json:"sid,omitempty"`
	ConnectString string        `json:"connectString,omitempty"`
	User          string        `json:"user"`
	Pool          DBPool        `json:"pool"`
	DSNUsed       string        `json:"dsnUsed,omitempty"`
}

func (d *Database) ToResponse() *DatabaseResponse {
	dsn, _ := d.DSN()
	return &DatabaseResponse{
		ID:            d.ID,
		Key:           d.Key,
		Driver:        d.Driver,
		Host:          d.Host,
		Port:          d.Port,
		DBName:        d.DBName,
		ServiceName:   d.ServiceName,
		SID:           d.SID,
		ConnectString: d.ConnectString,
		User:          d.User,
		Pool:          d.Pool,
		DSNUsed:       dsn,
	}
}

// DSN formula a string de conexão de acordo com o driver.
func (d *Database) DSN() (string, error) {
	switch d.Driver {
	case enum.DBDriverMySQL:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			d.User, d.Password, d.Host, d.Port, d.DBName), nil
	case enum.DBDriverPostgres:
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			d.Host, d.Port, d.User, d.Password, d.DBName), nil
	case enum.DBDriverOracle:
		return d.oracleDSN()
	case enum.DBDriverSQLServer:
		return fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;encrypt=disable",
			d.Host, d.User, d.Password, d.Port, d.DBName), nil
	default:
		return "", fmt.Errorf("driver desconhecido: %s", d.Driver)
	}
}

// oracleDSN monta o DSN para o godror seguindo a prioridade:
// 1. connectString customizado (passado direto, sem modificação)
// 2. SID  → host:port:SID  (formato Easy Connect antigo)
// 3. serviceName → host:port/serviceName
// 4. dbName como service name → host:port/dbName
func (d *Database) oracleDSN() (string, error) {
	var connect string
	switch {
	case d.ConnectString != "":
		connect = d.ConnectString
	case d.SID != "":
		connect = fmt.Sprintf("%s:%d:%s", d.Host, d.Port, d.SID)
	case d.ServiceName != "":
		connect = fmt.Sprintf("%s:%d/%s", d.Host, d.Port, d.ServiceName)
	default:
		connect = fmt.Sprintf("%s:%d/%s", d.Host, d.Port, d.DBName)
	}
	return fmt.Sprintf(`user="%s" password="%s" connectString="%s"`,
		d.User, d.Password, connect), nil
}

// OpenConnection abre e valida a conexão com o banco externo.
func (d *Database) OpenConnection() (*sqlx.DB, error) {
	dsn, err := d.DSN()
	if err != nil {
		return nil, err
	}

	driver := string(d.Driver)
	if d.Driver == enum.DBDriverOracle {
		driver = "godror"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := sqlx.ConnectContext(ctx, driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar em '%s': %w", d.Key, err)
	}

	pool := d.Pool
	db.SetMaxOpenConns(pool.MaxOpenConns)
	db.SetMaxIdleConns(pool.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(pool.ConnMaxLifetime) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(pool.ConnMaxIdleTime) * time.Second)

	log.Printf("[database] conexão '%s' (%s) estabelecida", d.Key, d.Driver)
	return db, nil
}
