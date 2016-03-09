package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleParseRoutes(t *testing.T) {
	routeString := "min-healthy/warning/slack"
	routes, err := ParseRoutes(routeString)
	assert.NoError(t, err)
	assert.Len(t, routes, 1)
	route := routes[0]
	expectedRoute := Route{
		Check:      "min-healthy",
		CheckLevel: Warning,
		Notifier:   "slack",
	}

	assert.Equal(t, expectedRoute, route)
}

func TestParseRoutesForEmptyString(t *testing.T) {
	routes, err := ParseRoutes("")
	assert.Error(t, err)
	assert.Nil(t, routes)
}

func TestParseRoutesForMultipleRoutes(t *testing.T) {
	routeString := "min-healthy/warning/slack;min-healthy/critical/slack"
	routes, err := ParseRoutes(routeString)
	assert.NoError(t, err)
	assert.Len(t, routes, 2)
	route := routes[0]
	expectedRoute := Route{
		Check:      "min-healthy",
		CheckLevel: Warning,
		Notifier:   "slack",
	}
	assert.Equal(t, expectedRoute, route)

	route = routes[1]
	expectedRoute = Route{
		Check:      "min-healthy",
		CheckLevel: Critical,
		Notifier:   "slack",
	}
	assert.Equal(t, expectedRoute, route)
}

func TestParseInvalidCheckLevel(t *testing.T) {
	routeString := "min-healthy/blahblah/slack"
	_, err := ParseRoutes(routeString)
	assert.Error(t, err)
}

func TestParseCheckLevel(t *testing.T) {
	expected := make(map[string]CheckStatus)
	expected["Warning"] = Warning
	expected["WARNING"] = Warning
	expected["warning"] = Warning
	expected["WaRnInG"] = Warning
	expected["Critical"] = Critical
	expected["CRITICAL"] = Critical
	expected["critical"] = Critical
	expected["CrItIcAl"] = Critical
	expected["Pass"] = Pass
	expected["pass"] = Pass
	expected["PASS"] = Pass
	expected["PaSs"] = Pass
	expected["Resolved"] = Resolved
	expected["RESOLVED"] = Resolved
	expected["ReSoLvEd"] = Resolved
	for input, expectedOutput := range expected {
		output, err := parseCheckLevel(input)
		assert.NoError(t, err)
		assert.Equal(t, expectedOutput, output)
	}

	_, err := parseCheckLevel("invalid-check-level")
	assert.Error(t, err)
}

func TestDefaultRoutes(t *testing.T) {
	allWarningRoute := DefaultRoutes[0]
	expectedWarningRoute := Route{
		Check:      "*",
		CheckLevel: Warning,
		Notifier:   "*",
	}
	assert.Equal(t, expectedWarningRoute, allWarningRoute)

	allCriticalRoute := DefaultRoutes[1]
	expectedCriticalRoute := Route{
		Check:      "*",
		CheckLevel: Critical,
		Notifier:   "*",
	}
	assert.Equal(t, expectedCriticalRoute, allCriticalRoute)
}
