package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

func TestNormalizeStateMachineDefinition_Example(t *testing.T) {
	body, err := ioutil.ReadFile("testdata/example.json")
	require.NoError(t, err)

	expected, err := ioutil.ReadFile("testdata/example_expected.json")
	require.NoError(t, err)

	actual := normalizeStateMachineDefinition(string(body), "dynamoid", "funcarn")
	assert.JSONEq(t, string(expected), actual)
}

func TestNormalizeStateMachineDefinition_MapState(t *testing.T) {
	body, err := ioutil.ReadFile("testdata/mapstate.json")
	require.NoError(t, err)

	expected, err := ioutil.ReadFile("testdata/mapstate_expected.json")
	require.NoError(t, err)

	actual := normalizeStateMachineDefinition(string(body), "dynamoid", "funcarn")
	assert.JSONEq(t, string(expected), actual)
}

func TestNormalizeStateMachineDefinition_ParallelState(t *testing.T) {
	body, err := ioutil.ReadFile("testdata/parallel.json")
	require.NoError(t, err)

	expected, err := ioutil.ReadFile("testdata/parallel_expected.json")
	require.NoError(t, err)

	actual := normalizeStateMachineDefinition(string(body), "dynamoid", "funcarn")
	assert.JSONEq(t, string(expected), actual)
}
