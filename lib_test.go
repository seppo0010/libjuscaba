package libjuscaba

import (
	"testing"

	"gotest.tools/assert"
)

func TestGetExpediente(t *testing.T) {
	exp, err := GetExpediente("182908/2020-0")
	assert.NilError(t, err)
	assert.Assert(t, exp != nil)
}
