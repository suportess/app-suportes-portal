package execute

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"

	"br.tec.suportes/portal/internal/apierr"
	"br.tec.suportes/portal/internal/domain"
	"br.tec.suportes/portal/internal/enum"
	"software.sslmate.com/src/go-pkcs12"
)

// CertificateProvider is a function that fetches a certificate by name.
type CertificateProvider func(name string) (*domain.Certificate, apierr.Detail)

// ---- POST ----

type PostExecute struct {
	certProvider CertificateProvider
}

func NewPostExecute(cp CertificateProvider) *PostExecute { return &PostExecute{certProvider: cp} }
func (e *PostExecute) Type() string                      { return string(enum.CommandTypePost) }

func (e *PostExecute) Run(ctx *ExecContext) (interface{}, apierr.Detail) {
	return runHTTP(ctx, e.certProvider)
}

// ---- PUT ----

type PutExecute struct {
	certProvider CertificateProvider
}

func NewPutExecute(cp CertificateProvider) *PutExecute { return &PutExecute{certProvider: cp} }
func (e *PutExecute) Type() string                     { return string(enum.CommandTypePut) }

func (e *PutExecute) Run(ctx *ExecContext) (interface{}, apierr.Detail) {
	return runHTTP(ctx, e.certProvider)
}

// ---- GET ----

type GetExecute struct {
	certProvider CertificateProvider
}

func NewGetExecute(cp CertificateProvider) *GetExecute { return &GetExecute{certProvider: cp} }
func (e *GetExecute) Type() string                     { return string(enum.CommandTypeGet) }

func (e *GetExecute) Run(ctx *ExecContext) (interface{}, apierr.Detail) {
	return runHTTP(ctx, e.certProvider)
}

// ---- REMOVE ----

type RemoveExecute struct {
	certProvider CertificateProvider
}

func NewRemoveExecute(cp CertificateProvider) *RemoveExecute {
	return &RemoveExecute{certProvider: cp}
}
func (e *RemoveExecute) Type() string { return string(enum.CommandTypeRemove) }

func (e *RemoveExecute) Run(ctx *ExecContext) (interface{}, apierr.Detail) {
	return runHTTP(ctx, e.certProvider)
}

// ---- shared HTTP logic ----

func runHTTP(ctx *ExecContext, cp CertificateProvider) (interface{}, apierr.Detail) {
	cmd := ctx.Command

	if aerr := validateBodyFields(ctx.Body, cmd.Body.Fields); aerr != nil {
		return nil, aerr
	}

	route := replacePlaceholders(cmd.Route, ctx.Params, ctx.Body)

	var req *http.Request
	var err error

	ct := strings.ToLower(cmd.ContentType)
	switch ct {
	case string(enum.ContentTypeMultipartForm):
		req, err = buildMultipartReq(cmd.Type, route, ctx.Body, ctx.Headers)
	case string(enum.ContentTypeFormURLEncoded):
		req, err = buildFormReq(cmd.Type, route, ctx.Body, ctx.Headers)
	default:
		req, err = buildJSONReq(cmd.Type, route, ctx.Body, ctx.Headers)
	}
	if err != nil {
		return nil, apierr.New(err.Error(), nil)
	}

	client, aerr := buildClient(cmd.CertificateName, cp)
	if aerr != nil {
		return nil, aerr
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, apierr.BadGateway("erro ao executar requisição HTTP: "+err.Error(), nil)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, apierr.NewWithStatus(
			fmt.Sprintf("upstream retornou %s", resp.Status),
			string(body),
			resp.StatusCode,
		)
	}

	return readResp(resp)
}

func validateBodyFields(body map[string]interface{}, fields []domain.BodyField) apierr.Detail {
	if len(fields) == 0 {
		return nil
	}
	allowed := make(map[string]domain.BodyField, len(fields))
	for _, f := range fields {
		allowed[f.Name] = f
	}
	for k := range body {
		if _, ok := allowed[k]; !ok {
			return apierr.New("campo não permitido no corpo: "+k, nil)
		}
	}
	for _, f := range fields {
		if f.Required {
			if _, ok := body[f.Name]; !ok {
				return apierr.New("campo obrigatório ausente no corpo: "+f.Name, nil)
			}
		}
		if v, ok := body[f.Name]; ok && v != nil {
			if aerr := validateFieldValue(f, v); aerr != nil {
				return aerr
			}
		}
	}
	return nil
}

func validateFieldValue(f domain.BodyField, v interface{}) apierr.Detail {
	switch f.Type {
	case enum.FieldTypeNumber:
		if _, ok := v.(float64); !ok {
			return apierr.New(fmt.Sprintf("campo '%s' deve ser do tipo numérico", f.Name), nil)
		}
	case enum.FieldTypeBoolean:
		if _, ok := v.(bool); !ok {
			return apierr.New(fmt.Sprintf("campo '%s' deve ser do tipo booleano", f.Name), nil)
		}
	case enum.FieldTypeString:
		s, ok := v.(string)
		if !ok {
			return apierr.New(fmt.Sprintf("campo '%s' deve ser do tipo texto", f.Name), nil)
		}
		if f.Maximum > 0 && len(s) > f.Maximum {
			return apierr.New(fmt.Sprintf("campo '%s' excede o máximo de %d caracteres", f.Name, f.Maximum), nil)
		}
		if f.Minimum > 0 && len(s) < f.Minimum {
			return apierr.New(fmt.Sprintf("campo '%s' requer mínimo de %d caracteres", f.Name, f.Minimum), nil)
		}
	case enum.FieldTypeBase64:
		s, ok := v.(string)
		if !ok {
			return apierr.New(fmt.Sprintf("campo '%s' deve ser um texto base64", f.Name), nil)
		}
		if err := decodeBase64Field(f.Name, s); err != nil {
			return err
		}
	}
	return nil
}

func decodeBase64Field(name, s string) apierr.Detail {
	data := s
	if idx := strings.Index(s, ","); idx != -1 {
		data = s[idx+1:]
	}
	if _, err := base64.StdEncoding.DecodeString(data); err != nil {
		return apierr.New(fmt.Sprintf("campo '%s' contém base64 inválido: %v", name, err), nil)
	}
	return nil
}

func replacePlaceholders(route string, params map[string]interface{}, body map[string]interface{}) string {
	result := route
	for k, v := range params {
		result = strings.ReplaceAll(result, "{"+k+"}", fmt.Sprintf("%v", v))
	}
	for k, v := range body {
		result = strings.ReplaceAll(result, "{"+k+"}", fmt.Sprintf("%v", v))
	}
	return result
}

func buildJSONReq(method enum.CommandType, url string, body map[string]interface{}, headers map[string]string) (*http.Request, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(string(method), url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

func buildFormReq(method enum.CommandType, url string, body map[string]interface{}, headers map[string]string) (*http.Request, error) {
	parts := make([]string, 0, len(body))
	for k, v := range body {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	req, err := http.NewRequest(string(method), url, strings.NewReader(strings.Join(parts, "&")))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", string(enum.ContentTypeFormURLEncoded))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

func buildMultipartReq(method enum.CommandType, url string, body map[string]interface{}, headers map[string]string) (*http.Request, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	for k, v := range body {
		switch val := v.(type) {
		case UploadedFile:
			if val.IsMultipart {
				part, err := w.CreateFormFile(k, val.FileName)
				if err != nil {
					return nil, err
				}
				part.Write(val.Content)
			} else {
				w.WriteField(k, string(val.Content))
			}
		case string:
			if k == "body" {
				h := make(textproto.MIMEHeader)
				h.Set("Content-Disposition", `form-data; name="body"`)
				h.Set("Content-Type", "application/json")
				part, err := w.CreatePart(h)
				if err != nil {
					return nil, err
				}
				io.WriteString(part, val)
			} else {
				w.WriteField(k, val)
			}
		default:
			w.WriteField(k, fmt.Sprintf("%v", v))
		}
	}
	w.Close()

	req, err := http.NewRequest(string(method), url, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

func buildClient(certName string, cp CertificateProvider) (*http.Client, apierr.Detail) {
	if certName == "" || cp == nil {
		return &http.Client{}, nil
	}
	cert, aerr := cp(certName)
	if aerr != nil {
		return nil, aerr
	}
	tlsCfg, aerr := buildTLSConfig(cert)
	if aerr != nil {
		return nil, aerr
	}
	return &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsCfg},
	}, nil
}

func buildTLSConfig(cert *domain.Certificate) (*tls.Config, apierr.Detail) {
	pfxData, err := base64.StdEncoding.DecodeString(cert.CertFile)
	if err != nil {
		return nil, apierr.New("erro ao decodificar arquivoCert: "+err.Error(), nil)
	}

	privateKey, leaf, caCerts, err := pkcs12.DecodeChain(pfxData, cert.Password)
	if err != nil {
		return nil, apierr.New("erro ao decodificar PFX: "+err.Error(), nil)
	}

	tlsCert := tls.Certificate{
		Certificate: [][]byte{leaf.Raw},
		PrivateKey:  privateKey,
		Leaf:        leaf,
	}

	pool := x509.NewCertPool()
	for _, ca := range caCerts {
		pool.AddCert(ca)
	}
	if cert.CACertFile != "" {
		raw, err2 := base64.StdEncoding.DecodeString(cert.CACertFile)
		if err2 == nil {
			pool.AppendCertsFromPEM(raw)
		} else {
			pool.AppendCertsFromPEM([]byte(cert.CACertFile))
		}
	}

	return &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		RootCAs:            pool,
		InsecureSkipVerify: true, //nolint:gosec
	}, nil
}

func readResp(resp *http.Response) (interface{}, apierr.Detail) {
	ct := resp.Header.Get("Content-Type")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apierr.New("erro ao ler resposta HTTP: "+err.Error(), nil)
	}
	mediaType, _, _ := mime.ParseMediaType(ct)
	if mediaType == "application/json" {
		var result interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			return string(body), nil
		}
		return result, nil
	}
	return string(body), nil
}
