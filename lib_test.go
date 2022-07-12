package libjuscaba

import (
	"testing"

	"gotest.tools/assert"
)

func TestGetExpediente(t *testing.T) {
	exp, err := GetExpediente("182908/2020-0")
	assert.NilError(t, err)
	assert.Assert(t, exp != nil)
	assert.Equal(t, exp.Numero, 182908)
	assert.Equal(t, exp.Anio, 2020)
}

func TestGetActuaciones(t *testing.T) {
	exp, err := GetExpediente("182908/2020-0")
	assert.NilError(t, err)
	acts, err := exp.GetActuaciones()
	assert.NilError(t, err)
	assert.Assert(t, len(acts) > 50)
}

func TestGetDocumento(t *testing.T) {
	exp, err := GetExpediente("182908/2020-0")
	assert.NilError(t, err)
	acts, err := exp.GetActuaciones()
	assert.NilError(t, err)
	documentos, err := FetchDocumentos(exp, acts[0])
	assert.NilError(t, err)
	assert.Assert(t, len(documentos) > 0)
}
