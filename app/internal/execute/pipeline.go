package execute

import (
	"maps"
	"reflect"
	"sync"

	"br.tec.suportes/portal/internal/apierr"
	"br.tec.suportes/portal/internal/domain"
	"br.tec.suportes/portal/internal/enum"
	"github.com/jmoiron/sqlx"
)

// StepConfig holds the resolved configuration for one pipeline step.
type StepConfig struct {
	Alias        string
	CommandKey   string
	CommandType  enum.CommandType
	DatabaseKey  string
	ReturnResult bool
	AbortNoData  bool
	SingleResult bool
	Parameters   []StepParam
}

// StepParam is a parameter binding that can extract values from prior step results.
type StepParam struct {
	Name    string
	Type    string
	Extract bool
	Value   interface{}
	Field   string
}

// DBProvider resolves a database connection by key.
type DBProvider func(key string) (*sqlx.DB, apierr.Detail)

// CommandProvider resolves a command by key.
type CommandProvider func(key string) (*domain.Command, apierr.Detail)

// Pipeline orchestrates execution of multiple steps.
type Pipeline struct {
	steps        []StepConfig
	executors    map[string]Executor
	singleThread bool
	dbProvider   DBProvider
	cmdProvider  CommandProvider
}

func NewPipeline(
	steps []StepConfig,
	executors map[string]Executor,
	singleThread bool,
	db DBProvider,
	cmd CommandProvider,
) *Pipeline {
	return &Pipeline{
		steps:        steps,
		executors:    executors,
		singleThread: singleThread,
		dbProvider:   db,
		cmdProvider:  cmd,
	}
}

func (p *Pipeline) Execute(params, body map[string]interface{}, headers map[string]string) (interface{}, apierr.Detail) {
	if !p.singleThread {
		return p.executeParallel(params, body, headers)
	}
	return p.executeSequential(params, body, headers)
}

func (p *Pipeline) executeSequential(params, body map[string]interface{}, headers map[string]string) (interface{}, apierr.Detail) {
	var lastMap map[string]interface{}
	var resultSlice []map[string]interface{}
	var firstAlias string
	processed := make(map[string]bool)

	for _, step := range p.steps {
		exec, ok := p.executors[string(step.CommandType)]
		if !ok {
			return nil, apierr.New("executor não encontrado para type: "+string(step.CommandType), nil)
		}

		cmd, aerr := p.cmdProvider(step.CommandKey)
		if aerr != nil {
			return nil, aerr
		}

		if processed[step.CommandKey] {
			continue
		}

		// Inject parameters extracted from prior step results
		merged := make(map[string]interface{})
		maps.Copy(merged, params)
		maps.Copy(merged, extractStepParams(step.Parameters, resultSlice))

		var db *sqlx.DB
		if step.DatabaseKey != "" {
			var dbErr apierr.Detail
			db, dbErr = p.dbProvider(step.DatabaseKey)
			if dbErr != nil {
				return nil, dbErr
			}
		}

		result, aerr := exec.Run(&ExecContext{
			Command:      cmd,
			DB:           db,
			Params:       merged,
			Body:         body,
			Headers:      headers,
			SingleResult: step.SingleResult,
		})

		if aerr != nil {
			return nil, aerr
		}

		if step.ReturnResult && result != nil {
			kind := reflect.TypeOf(result).Kind()

			if kind == reflect.Ptr {
				return result, nil
			} else if kind == reflect.Map {
				rm := result.(map[string]interface{})
				if firstAlias == "" {
					firstAlias = step.Alias
				}
				if lastMap == nil {
					if step.Alias == "" {
						lastMap = rm
					} else {
						lastMap = map[string]interface{}{step.Alias: rm}
					}
				} else {
					if step.Alias == "" {
						for k, v := range rm {
							lastMap[k] = v
						}
					} else {
						lastMap[step.Alias] = rm
					}
				}
				processed[step.CommandKey] = true

			} else if kind == reflect.Slice {
				if !step.SingleResult {
					return result, nil
				}
				sv := reflect.ValueOf(result)
				for i := 0; i < sv.Len(); i++ {
					if m, ok := sv.Index(i).Interface().(map[string]interface{}); ok {
						resultSlice = append(resultSlice, m)
					}
				}
			}
		} else if step.AbortNoData && result == nil {
			break
		}
	}

	if len(lastMap) > 0 {
		return lastMap, nil
	}
	if len(resultSlice) > 0 {
		return resultSlice, nil
	}
	return nil, nil
}

func (p *Pipeline) executeParallel(params, body map[string]interface{}, headers map[string]string) (interface{}, apierr.Detail) {
	var mu sync.Mutex
	var wg sync.WaitGroup
	var results []interface{}

	for _, step := range p.steps {
		exec, ok := p.executors[string(step.CommandType)]
		if !ok {
			continue
		}
		cmd, aerr := p.cmdProvider(step.CommandKey)
		if aerr != nil {
			continue
		}
		var db *sqlx.DB
		if step.DatabaseKey != "" {
			db, _ = p.dbProvider(step.DatabaseKey)
		}

		wg.Add(1)
		go func(e Executor, c *domain.Command, d *sqlx.DB, s StepConfig) {
			defer wg.Done()
			r, err := e.Run(&ExecContext{
				Command:      c,
				DB:           d,
				Params:       params,
				Body:         body,
				Headers:      headers,
				SingleResult: s.SingleResult,
			})
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				results = append(results, map[string]interface{}{"error": err.Error()})
			} else {
				results = append(results, r)
			}
		}(exec, cmd, db, step)
	}

	wg.Wait()
	return results, nil
}

func extractStepParams(params []StepParam, prior []map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for _, p := range params {
		if p.Extract {
			var values []interface{}
			for _, row := range prior {
				if v, ok := row[p.Field]; ok {
					values = append(values, v)
				}
			}
			out[p.Name] = values
		} else if p.Value != nil {
			out[p.Name] = p.Value
		}
	}
	return out
}
