package enum

// DBDriver representa os drivers de banco de dados suportados.
type DBDriver string

const (
	DBDriverMySQL     DBDriver = "mysql"
	DBDriverPostgres  DBDriver = "postgres"
	DBDriverOracle    DBDriver = "oracle"
	DBDriverSQLServer DBDriver = "sqlserver"
)

func (d DBDriver) IsValid() bool {
	switch d {
	case DBDriverMySQL, DBDriverPostgres, DBDriverOracle, DBDriverSQLServer:
		return true
	}
	return false
}

func (d DBDriver) String() string {
	return string(d)
}

// ValidDBDrivers retorna a lista de drivers válidos como string.
func ValidDBDrivers() []string {
	return []string{
		string(DBDriverMySQL),
		string(DBDriverPostgres),
		string(DBDriverOracle),
		string(DBDriverSQLServer),
	}
}

// CommandType representa os tipos de comandos suportados.
type CommandType string

const (
	CommandTypeQuery     CommandType = "QUERY"
	CommandTypeInsert    CommandType = "INSERT"
	CommandTypeUpdate    CommandType = "UPDATE"
	CommandTypeDelete    CommandType = "DELETE"
	CommandTypeProcedure CommandType = "PROCEDURE"
	CommandTypeAnonymous CommandType = "ANONYMOUS"
	CommandTypePost      CommandType = "POST"
	CommandTypeGet       CommandType = "GET"
	CommandTypePut       CommandType = "PUT"
	CommandTypeRemove    CommandType = "REMOVE"
)

func (c CommandType) IsValid() bool {
	switch c {
	case CommandTypeQuery, CommandTypeInsert, CommandTypeUpdate,
		CommandTypeDelete, CommandTypeProcedure, CommandTypeAnonymous,
		CommandTypePost, CommandTypeGet, CommandTypePut, CommandTypeRemove:
		return true
	}
	return false
}

func (c CommandType) IsHTTP() bool {
	switch c {
	case CommandTypePost, CommandTypeGet, CommandTypePut, CommandTypeRemove:
		return true
	}
	return false
}

func (c CommandType) IsSQL() bool {
	return !c.IsHTTP()
}

func (c CommandType) String() string {
	return string(c)
}

// ValidCommandTypes retorna a lista de tipos válidos.
func ValidCommandTypes() []string {
	return []string{
		string(CommandTypeQuery),
		string(CommandTypeInsert),
		string(CommandTypeUpdate),
		string(CommandTypeDelete),
		string(CommandTypeProcedure),
		string(CommandTypeAnonymous),
		string(CommandTypePost),
		string(CommandTypeGet),
		string(CommandTypePut),
		string(CommandTypeRemove),
	}
}

// ContentType representa os content-types HTTP suportados.
type ContentType string

const (
	ContentTypeJSON           ContentType = "application/json"
	ContentTypeFormURLEncoded ContentType = "application/x-www-form-urlencoded"
	ContentTypeMultipartForm  ContentType = "multipart/form-data"
	ContentTypeXML            ContentType = "application/xml"
	ContentTypePlainText      ContentType = "text/plain"
)

func (ct ContentType) IsValid() bool {
	switch ct {
	case ContentTypeJSON, ContentTypeFormURLEncoded,
		ContentTypeMultipartForm, ContentTypeXML, ContentTypePlainText:
		return true
	}
	return false
}

func (ct ContentType) String() string {
	return string(ct)
}

// HTTPMethod representa os métodos HTTP suportados nas rotas dinâmicas.
type HTTPMethod string

const (
	HTTPMethodGET    HTTPMethod = "GET"
	HTTPMethodPOST   HTTPMethod = "POST"
	HTTPMethodPUT    HTTPMethod = "PUT"
	HTTPMethodDELETE HTTPMethod = "DELETE"
	HTTPMethodPATCH  HTTPMethod = "PATCH"
)

func (m HTTPMethod) IsValid() bool {
	switch m {
	case HTTPMethodGET, HTTPMethodPOST, HTTPMethodPUT,
		HTTPMethodDELETE, HTTPMethodPATCH:
		return true
	}
	return false
}

func (m HTTPMethod) String() string {
	return string(m)
}

// ValidHTTPMethods retorna a lista de métodos HTTP válidos.
func ValidHTTPMethods() []string {
	return []string{
		string(HTTPMethodGET),
		string(HTTPMethodPOST),
		string(HTTPMethodPUT),
		string(HTTPMethodDELETE),
		string(HTTPMethodPATCH),
	}
}

// ParamType representa os tipos de parâmetros de um comando SQL.
type ParamType string

const (
	ParamTypeString  ParamType = "string"
	ParamTypeNumber  ParamType = "number"
	ParamTypeBoolean ParamType = "boolean"
	ParamTypeDate    ParamType = "date"
	ParamTypeOut     ParamType = "out"
	ParamTypeBase64  ParamType = "base64"
	ParamTypeArray   ParamType = "array"
)

func (p ParamType) IsValid() bool {
	switch p {
	case ParamTypeString, ParamTypeNumber, ParamTypeBoolean,
		ParamTypeDate, ParamTypeOut, ParamTypeBase64, ParamTypeArray:
		return true
	}
	return false
}

func (p ParamType) String() string {
	return string(p)
}

// FieldType representa os tipos de campos do body de um comando HTTP.
type FieldType string

const (
	FieldTypeString  FieldType = "string"
	FieldTypeNumber  FieldType = "number"
	FieldTypeBoolean FieldType = "boolean"
	FieldTypeBase64  FieldType = "base64"
	FieldTypeArray   FieldType = "array"
	FieldTypeObject  FieldType = "object"
)

func (f FieldType) IsValid() bool {
	switch f {
	case FieldTypeString, FieldTypeNumber, FieldTypeBoolean,
		FieldTypeBase64, FieldTypeArray, FieldTypeObject:
		return true
	}
	return false
}

func (f FieldType) String() string {
	return string(f)
}

// SQLOperator representa os operadores SQL suportados para filtros dinâmicos.
type SQLOperator string

const (
	SQLOperatorEquals       SQLOperator = "="
	SQLOperatorNotEquals    SQLOperator = "<>"
	SQLOperatorGreaterThan  SQLOperator = ">"
	SQLOperatorLessThan     SQLOperator = "<"
	SQLOperatorGreaterEqual SQLOperator = ">="
	SQLOperatorLessEqual    SQLOperator = "<="
	SQLOperatorLike         SQLOperator = "LIKE"
	SQLOperatorNotLike      SQLOperator = "NOT LIKE"
	SQLOperatorIn           SQLOperator = "IN"
	SQLOperatorNotIn        SQLOperator = "NOT IN"
	SQLOperatorIsNull       SQLOperator = "IS NULL"
	SQLOperatorIsNotNull    SQLOperator = "IS NOT NULL"
)

func (o SQLOperator) IsValid() bool {
	switch o {
	case SQLOperatorEquals, SQLOperatorNotEquals, SQLOperatorGreaterThan,
		SQLOperatorLessThan, SQLOperatorGreaterEqual, SQLOperatorLessEqual,
		SQLOperatorLike, SQLOperatorNotLike, SQLOperatorIn, SQLOperatorNotIn,
		SQLOperatorIsNull, SQLOperatorIsNotNull:
		return true
	}
	return false
}

func (o SQLOperator) String() string {
	return string(o)
}
