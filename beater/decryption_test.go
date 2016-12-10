package beater

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecryptString(t *testing.T) {
	saltkey := "tututoto"
	encrypted := "-LzSzX_qXMdlIq-DZ9s59mHDJv5fdIm6"
	expectedResult := "totopopo"

	//if everything works well
	realResult, err := decryptString(encrypted, saltkey)
	assert.Nil(t, err)
	assert.Equal(t, expectedResult, realResult, "This seems not be base64 + aes + CFB encrypted.")

	//if encrypted is too short
	realResult2, err := decryptString("tt", saltkey)
	expectedError := errors.New("cipherText too short")
	assert.Equal(t, "", realResult2)
	assert.Equal(t, expectedError, err)

	//if something fails with the Cipher creation
	realResult3, err := decryptString(encrypted, "")
	assert.Equal(t, "", realResult3)
	assert.NotNil(t, err)
}
