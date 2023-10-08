package util

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func Unmarshal[T any](t *testing.T, body *bytes.Buffer) *T {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotResult T
	err = json.Unmarshal(data, &gotResult)
	require.NoError(t, err)

	return &gotResult
}
