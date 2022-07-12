package libjuscaba

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

const FichaType = "ficha"
const ActuacionType = "actuacion"
const DocumentType = "document"
const RegularAttachment = 0
const ActuacionesNotificadasAttachment = 1
const AdjuntosAttachment = 1
const CedulaAttachment = 2

type FichaRadicaciones struct {
	SecretariaPrimeraInstancia string `json:"secretariaPrimeraInstancia"`
	OrganismoSegundaInstancia  string `json:"organismoSegundaInstancia"`
	SecretariaSegundaInstancia string `json:"secretariaSegundaInstancia"`
	OrganismoPrimeraInstancia  string `json:"organismoPrimeraInstancia"`
}

type FichaObjetosJuicio struct {
	ObjetoJuicio string `json:"objetoJuicio"`
	Categoria    string `json:"categoria"`
	EsPrincipal  int    `json:"esPrincipal"`
	Materia      string `json:"materia"`
}

type FichaUbicacion struct {
	Organismo   string `json:"organismo"`
	Dependencia string `json:"dependencia"`
}
type Ficha struct {
	ExpId            int
	Radicaciones     FichaRadicaciones    `json:"radicaciones"`
	Numero           int                  `json:"numero"`
	Anio             int                  `json:"anio"`
	Sufijo           int                  `json:"sufijo"`
	ObjetosJuicio    []FichaObjetosJuicio `json:"objetosJuicio"`
	Ubicacion        FichaUbicacion       `json:"ubicacion"`
	FechaInicio      int                  `json:"fechaInicio"`
	UltimoMovimiento int                  `json:"ultimoMovimiento"`
	TieneSentencia   int                  `json:"tieneSentencia"`
	EsPrivado        int                  `json:"esPrivado"`
	TipoExpediente   string               `json:"tipoExpediente"`
	CUIJ             string               `json:"cuij"`
	Caratula         string               `json:"caratula"`
	Monto            float64              `json:"monto"`
	Etiquetas        string               `json:"etiquetas"`
}

func FichaID(expedienteID string) string {
	return fmt.Sprintf("ficha %v", expedienteID)
}

func (ficha *Ficha) NumeroDeExpediente(separator string) string {
	return fmt.Sprintf("%d%s%d", ficha.Numero, separator, ficha.Anio)
}

func (ficha *Ficha) Id() string {
	return FichaID(ficha.NumeroDeExpediente("-"))
}

type ActuacionesPage struct {
	TotalPages       int                     `json:"totalPages"`
	TotalElements    int                     `json:"totalElements"`
	NumberOfElements int                     `json:"numberOfElements"`
	Last             bool                    `json:"last"`
	First            bool                    `json:"first"`
	Size             int                     `json:"size"`
	Number           int                     `json:"number"`
	Pageable         ActuacionesPagePageable `json:"pageable"`
	Content          []*Actuacion            `json:"content"`
}

type Actuacion struct {
	EsCedula               int    `json:"esCedula"`
	Codigo                 string `json:"codigo"`
	ActuacionesNotificadas string `json:"actuacionesNotificadas"`
	Numero                 int    `json:"-"`
	FechaFirma             int    `json:"fechaFirma"`
	Firmantes              string `json:"firmantes"`
	ActId                  int    `json:"actId"`
	Titulo                 string `json:"titulo"`
	FechaNotificacion      int    `json:"fechaNotificacion"`
	PoseeAdjunto           int    `json:"poseeAdjunto"`
	CUIJ                   string `json:"cuij"`
	Anio                   int    `json:"-"`
}

func (actuacion *Actuacion) Id() string {
	return fmt.Sprintf("actuacion %d", actuacion.ActId)
}

type ActuacionWithExpediente struct {
	Actuacion
	NumeroDeExpediente string `json:"numeroDeExpediente"`
}

type ActuacionesPagePageable struct {
	PageNumber int `json:"pageNumber"`
	PageSize   int `json:"pageSize"`
	Offset     int `json:"offset"`
}

type Documento struct {
	URL                string
	MirrorURL          string
	ActuacionID        string `json:"actuacionId"`
	NumeroDeExpediente string `json:"numeroDeExpediente"`
	Type               int    `json:"type"`
	Nombre             string `json:"nombre"`
	Content            string `json:"content"`
}

type SearchFormFilter struct {
	Identificador string `json:"identificador"`
}
type SearchForm struct {
	Filter       string `json:"filter"`
	TipoBusqueda string `json:"tipoBusqueda"`
	Page         int    `json:"page"`
	Size         int    `json:"size"`
}

type SearchResultContent struct {
	ExpId int `json:"expId"`
}

type SearchResult struct {
	Content []SearchResultContent `json:"content"`
}

func getExpedienteCandidates(criteria string) ([]int, error) {
	filter, _ := json.Marshal(SearchFormFilter{
		Identificador: criteria,
	})
	info, _ := json.Marshal(SearchForm{
		Filter:       string(filter),
		TipoBusqueda: "CAU",
		Page:         0,
		Size:         10,
	})

	u := "https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/lista"
	resp, err := http.PostForm(u, url.Values{
		"info": {string(info)},
	})
	if err != nil {
		log.WithFields(log.Fields{
			"expediente": criteria,
			"url":        u,
			"error":      err.Error(),
		}).Warn("Failed to get expediente")
		return nil, err
	}
	defer resp.Body.Close()

	sr := SearchResult{}
	err = json.NewDecoder(resp.Body).Decode(&sr)
	if err != nil {
		log.WithFields(log.Fields{
			"expediente": criteria,
			"url":        u,
			"httpStatus": resp.StatusCode,
		}).Warn("Failed to decode json")
		return nil, err
	}
	res := make([]int, len(sr.Content))
	for i, s := range sr.Content {
		res[i] = s.ExpId
	}
	return res, nil
}

func getFicha(candidate int) (*Ficha, error) {
	u := fmt.Sprintf("https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/ficha?expId=%d", candidate)
	resp, err := http.Get(u)
	if err != nil {
		log.WithFields(log.Fields{
			"expId": candidate,
			"url":   u,
		}).Warn("Failed to get ficha")
		return nil, err
	}
	defer resp.Body.Close()

	ficha := Ficha{ExpId: candidate}
	err = json.NewDecoder(resp.Body).Decode(&ficha)
	if err != nil {
		log.WithFields(log.Fields{
			"expId":      candidate,
			"url":        u,
			"httpStatus": resp.StatusCode,
		}).Warn("Failed to decode json")
		return nil, err
	}
	return &ficha, nil
}

func GetExpediente(criteria string) (*Ficha, error) {
	candidates, err := getExpedienteCandidates(criteria)
	if err != nil {
		return nil, err
	}

	for _, candidate := range candidates {
		ficha, err := getFicha(candidate)
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(criteria, fmt.Sprintf("%d/%d", ficha.Numero, ficha.Anio)) {
			return ficha, nil
		}
	}
	log.WithFields(log.Fields{
		"expediente": criteria,
	}).Info("cannot find expediente")
	return nil, fmt.Errorf("cannot find ficha for criteria: %s", criteria)
}

func (ficha *Ficha) getActuacionesPage(pagenum int) (*ActuacionesPage, error) {
	expId := ficha.ExpId
	log.WithFields(log.Fields{
		"page": pagenum,
	}).Info("getting actuaciones")
	size := 100
	u := fmt.Sprintf("https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/actuaciones?filtro=%%7B%%22cedulas%%22%%3Atrue%%2C%%22escritos%%22%%3Atrue%%2C%%22despachos%%22%%3Atrue%%2C%%22notas%%22%%3Atrue%%2C%%22expId%%22%%3A%d%%2C%%22accesoMinisterios%%22%%3Afalse%%7D&page=%d&size=%d",
		expId,
		pagenum,
		size,
	)
	res, err := http.Get(u)
	if err != nil {
		log.WithFields(log.Fields{
			"expId":   expId,
			"pagenum": pagenum,
			"url":     u,
		}).Warn("Failed to get actuaciones")
		return nil, err
	}
	page := ActuacionesPage{}
	err = json.NewDecoder(res.Body).Decode(&page)
	if err != nil {
		log.WithFields(log.Fields{
			"expId":      expId,
			"pagenum":    pagenum,
			"url":        u,
			"httpStatus": res.StatusCode,
		}).Warn("Failed to decode json")
		return nil, err
	}
	return &page, nil
}

func (ficha *Ficha) GetActuaciones() ([]*Actuacion, error) {
	actuaciones := make([]*Actuacion, 0, 1)
	pagenum := 0
	for {
		page, err := ficha.getActuacionesPage(pagenum)
		if err != nil {
			return nil, err
		}
		if len(page.Content) == 0 {
			break
		}
		actuaciones = append(actuaciones, page.Content...)
		pagenum++
	}
	return actuaciones, nil
}

func GetAdjuntosCedula(url string, ficha *Ficha, actuacion *Actuacion) ([]*Documento, error) {
	u := fmt.Sprintf("https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/cedulas/adjuntos?filter=%%7B%%22cedulaCuij%%22:%%22%v%%22,%%22expId%%22:%v,%%22ministerios%%22:false%%7D",
		actuacion.CUIJ,
		ficha.ExpId,
	)
	resp, err := http.Get(u)
	if err != nil {
		log.WithFields(log.Fields{
			"actId": actuacion.ActId,
			"url":   u,
		}).Warn("Failed to get adjuntos")
		return nil, err
	}
	defer resp.Body.Close()

	adjuntos := []map[string]interface{}{}
	err = json.NewDecoder(resp.Body).Decode(&adjuntos)
	if err != nil {
		log.WithFields(log.Fields{
			"actId":      actuacion.ActId,
			"url":        u,
			"httpStatus": resp.StatusCode,
		}).Warn("Failed to decode json")
		return nil, err
	}
	documentos := make([]*Documento, 0, len(adjuntos))
	for _, adjunto := range adjuntos {
		if val, found := adjunto["adjuntoId"]; !found || val == nil {
			continue
		}
		url := fmt.Sprintf("https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/cedulas/adjuntoPdf?filter=%%7B%%22aacId%%22:%v,%%22expId%%22:%v,%%22ministerios%%22:false%%7D",
			int(adjunto["adjuntoId"].(float64)),
			ficha.ExpId,
		)
		documentos = append(documentos, &Documento{
			URL:                url,
			ActuacionID:        actuacion.Id(),
			NumeroDeExpediente: fmt.Sprintf("%d/%d", ficha.Numero, ficha.Anio),
			Type:               CedulaAttachment,
			Nombre:             adjunto["adjuntoNombre"].(string),
		})
	}
	return documentos, nil
}

func GetAdjuntosNoCedula(url string, ficha *Ficha, actuacion *Actuacion) ([]*Documento, error) {
	u := fmt.Sprintf("https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/actuaciones/adjuntos?actId=%d&expId=%v&accesoMinisterios=false",
		actuacion.ActId,
		ficha.ExpId,
	)
	resp, err := http.Get(u)
	if err != nil {
		log.WithFields(log.Fields{
			"actId": actuacion.ActId,
			"url":   u,
		}).Warn("Failed to get adjuntos")
		return nil, err
	}
	defer resp.Body.Close()

	adjuntos := map[string][]map[string]interface{}{}
	err = json.NewDecoder(resp.Body).Decode(&adjuntos)
	if err != nil {
		log.WithFields(log.Fields{
			"httpStatus": resp.StatusCode,
			"actId":      actuacion.ActId,
			"url":        u,
		}).Warn("Failed to decode json")
		return nil, err
	}
	documentos := make([]*Documento, 0, len(adjuntos["adjuntos"]))
	for _, adjunto := range adjuntos["adjuntos"] {
		if val, found := adjunto["adjId"]; !found || val == nil {
			continue
		}
		url := fmt.Sprintf("https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/actuaciones/adjuntoPdf?filter=%%7B%%22aacId%%22:%v,%%22expId%%22:%v,%%22ministerios%%22:false%%7D",
			int(adjunto["adjId"].(float64)),
			ficha.ExpId,
		)
		documentos = append(documentos, &Documento{
			URL:                url,
			ActuacionID:        actuacion.Id(),
			NumeroDeExpediente: fmt.Sprintf("%d/%d", ficha.Numero, ficha.Anio),
			Type:               AdjuntosAttachment,
			Nombre:             adjunto["titulo"].(string),
		})
	}
	return documentos, nil
}

func GetAdjuntos(url string, ficha *Ficha, actuacion *Actuacion) ([]*Documento, error) {
	if actuacion.EsCedula == 1 {
		return GetAdjuntosCedula(url, ficha, actuacion)
	} else {
		return GetAdjuntosNoCedula(url, ficha, actuacion)
	}
}

func FetchDocumentos(ficha *Ficha, actuacion *Actuacion) ([]*Documento, error) {
	documentos := make([]*Documento, 0)
	url := fmt.Sprintf(
		"https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/actuaciones/pdf?datos=%%7B%%22actId%%22:%d,%%22expId%%22:%d,%%22esNota%%22:false,%%22cedulaId%%22:null,%%22ministerios%%22:false%%7D",
		actuacion.ActId,
		ficha.ExpId,
	)
	documentos = append(documentos, &Documento{
		URL:                url,
		ActuacionID:        actuacion.Id(),
		NumeroDeExpediente: fmt.Sprintf("%d/%d", ficha.Numero, ficha.Anio),
		Type:               RegularAttachment,
		Nombre:             "",
	})
	if actuacion.ActuacionesNotificadas != "" {

		url := fmt.Sprintf(
			"https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/actuaciones/pdf?datos=%%7B%%22actId%%22:%%22%v%%22,%%22expId%%22:%v,%%22esNota%%22:false,%%22cedulaId%%22:%v,%%22ministerios%%22:false%%7D",
			actuacion.ActuacionesNotificadas,
			ficha.ExpId,
			actuacion.ActId,
		)
		documentos = append(documentos, &Documento{
			URL:                url,
			ActuacionID:        actuacion.Id(),
			NumeroDeExpediente: fmt.Sprintf("%d/%d", ficha.Numero, ficha.Anio),
			Type:               ActuacionesNotificadasAttachment,
			Nombre:             "",
		})
	}
	if actuacion.PoseeAdjunto > 0 {
		adjuntos, _ := GetAdjuntos(url, ficha, actuacion)
		documentos = append(documentos, adjuntos...)
	}

	return documentos, nil
}
