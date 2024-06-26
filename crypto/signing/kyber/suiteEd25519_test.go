package kyber_test

import (
	"testing"

	"github.com/numbatx/gn-numbat/crypto/signing/kyber"
	"github.com/stretchr/testify/assert"
)

func TestNewBlakeSHA256Ed25519(t *testing.T) {
	suite := kyber.NewBlakeSHA256Ed25519()

	assert.NotNil(t, suite)
}

func TestSuiteEd25519_RandomStream(t *testing.T) {
	suite := kyber.NewBlakeSHA256Ed25519()
	stream := suite.RandomStream()

	assert.NotNil(t, stream)
}

func TestSuiteEd25519_CreatePoint(t *testing.T) {
	suite := kyber.NewBlakeSHA256Ed25519()

	point1 := suite.CreatePoint()
	point2 := suite.CreatePoint()

	assert.NotNil(t, point1)
	assert.NotNil(t, point2)
	assert.False(t, point1 == point2)
}

func TestSuiteEd25519_String(t *testing.T) {
	suite := kyber.NewBlakeSHA256Ed25519()

	str := suite.String()
	assert.Equal(t, "Ed25519", str)
}

func TestSuiteEd25519_ScalarLen(t *testing.T) {
	suite := kyber.NewBlakeSHA256Ed25519()

	length := suite.ScalarLen()
	assert.Equal(t, 32, length)
}

func TestSuiteEd25519_CreateScalar(t *testing.T) {
	suite := kyber.NewBlakeSHA256Ed25519()

	scalar := suite.CreateScalar()
	assert.NotNil(t, scalar)
}

func TestSuiteEd25519_PointLen(t *testing.T) {
	suite := kyber.NewBlakeSHA256Ed25519()

	pointLength := suite.PointLen()

	assert.Equal(t, 32, pointLength)
}

func TestSuiteEd25519_CreateKey(t *testing.T) {
	suite := kyber.NewBlakeSHA256Ed25519()

	stream := suite.RandomStream()
	private, public := suite.CreateKeyPair(stream)

	assert.NotNil(t, private)
	assert.NotNil(t, public)
}

func TestSuiteEd25519_GetUnderlyingSuite(t *testing.T) {
	suite := kyber.NewBlakeSHA256Ed25519()

	obj := suite.GetUnderlyingSuite()

	assert.NotNil(t, obj)
}
