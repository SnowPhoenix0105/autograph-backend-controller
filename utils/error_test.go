package utils

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetCurrentFuncName(t *testing.T) {
	assert.Equal(t, "utils.TestGetCurrentFuncName", GetCurrentFuncName())
}

func TestWrapError(t *testing.T) {
	originErr := errors.New("TestError")
	msg := "TestMessage"
	wrappedErr := WrapError(originErr, msg)
	assert.True(t, errors.Is(wrappedErr, originErr))
	assert.NotEqual(t, originErr, wrappedErr)
	t.Log(wrappedErr)

	assert.Nil(t, WrapError(nil, "Msg"))
}
