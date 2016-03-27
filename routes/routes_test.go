package routes

import (
	"testing"

	"github.com/ashwanthkumar/marathon-alerts/checks"
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
		CheckLevel: checks.Warning,
		Notifier:   "slack",
	}

	assert.Equal(t, expectedRoute, route)
}

func TestSimpleParseRoutesEndingWithSemiColon(t *testing.T) {
	routeString := "min-healthy/warning/slack;"
	routes, err := ParseRoutes(routeString)
	assert.NoError(t, err)
	assert.Len(t, routes, 1)
	route := routes[0]
	expectedRoute := Route{
		Check:      "min-healthy",
		CheckLevel: checks.Warning,
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
		CheckLevel: checks.Warning,
		Notifier:   "slack",
	}
	assert.Equal(t, expectedRoute, route)

	route = routes[1]
	expectedRoute = Route{
		Check:      "min-healthy",
		CheckLevel: checks.Critical,
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
	expected := make(map[string]checks.CheckStatus)
	expected["Warning"] = checks.Warning
	expected["WARNING"] = checks.Warning
	expected["warning"] = checks.Warning
	expected["WaRnInG"] = checks.Warning
	expected["Critical"] = checks.Critical
	expected["CRITICAL"] = checks.Critical
	expected["critical"] = checks.Critical
	expected["CrItIcAl"] = checks.Critical
	expected["Pass"] = checks.Pass
	expected["pass"] = checks.Pass
	expected["PASS"] = checks.Pass
	expected["PaSs"] = checks.Pass
	expected["Resolved"] = checks.Resolved
	expected["resolved"] = checks.Resolved
	expected["RESOLVED"] = checks.Resolved
	expected["ReSoLvEd"] = checks.Resolved
	for input, expectedOutput := range expected {
		output, err := parseCheckLevel(input)
		assert.NoError(t, err)
		assert.Equal(t, expectedOutput, output)
	}

	_, err := parseCheckLevel("invalid-check-level")
	assert.Error(t, err)
}

func TestDefaultRoutes(t *testing.T) {
	defaultRoutes, err := ParseRoutes(DefaultRoutes)
	assert.NoError(t, err)
	assert.Len(t, defaultRoutes, 3)
	allWarningRoute := defaultRoutes[0]
	expectedWarningRoute := Route{
		Check:      "*",
		CheckLevel: checks.Warning,
		Notifier:   "*",
	}
	assert.Equal(t, expectedWarningRoute, allWarningRoute)

	allCriticalRoute := defaultRoutes[1]
	expectedCriticalRoute := Route{
		Check:      "*",
		CheckLevel: checks.Critical,
		Notifier:   "*",
	}
	assert.Equal(t, expectedCriticalRoute, allCriticalRoute)

	allResolvedRoute := defaultRoutes[2]
	expectedResolvedRoute := Route{
		Check:      "*",
		CheckLevel: checks.Resolved,
		Notifier:   "*",
	}
	assert.Equal(t, expectedResolvedRoute, allResolvedRoute)
}

func TestRouteMatch(t *testing.T) {
	defaultRoutes, err := ParseRoutes(DefaultRoutes)
	assert.NoError(t, err)
	assert.Len(t, defaultRoutes, 3)

	allWarningRoute := defaultRoutes[0]
	warningCheck := checks.AppCheck{
		CheckName: "check-name",
		Result:    checks.Warning,
	}
	warningCheckMatch := allWarningRoute.Match(warningCheck)
	assert.True(t, warningCheckMatch)

	allCriticalRoute := defaultRoutes[1]
	criticalCheck := checks.AppCheck{
		CheckName: "check-name",
		Result:    checks.Critical,
	}
	criticalCheckMatch := allCriticalRoute.Match(criticalCheck)
	assert.True(t, criticalCheckMatch)
}

func TestRouteMatchDoesNotWork(t *testing.T) {
	defaultRoutes, err := ParseRoutes(DefaultRoutes)
	assert.NoError(t, err)
	assert.Len(t, defaultRoutes, 3)

	allWarningRoute := defaultRoutes[0]
	resolvedCheck := checks.AppCheck{
		CheckName: "check-name",
		Result:    checks.Resolved,
	}
	resolvedCheckMatch := allWarningRoute.Match(resolvedCheck)
	assert.False(t, resolvedCheckMatch)
}

func TestRouteMatchNotifier(t *testing.T) {
	route := Route{
		Notifier: "*",
	}
	assert.True(t, route.MatchNotifier("slack"))
}

func TestRouteMatchCheckResult(t *testing.T) {
	route := Route{
		CheckLevel: checks.Warning,
	}
	assert.True(t, route.MatchCheckResult(checks.Warning))
	assert.False(t, route.MatchCheckResult(checks.Pass))
}

func BenchmarkSimpleParseRoutes(b *testing.B) {
	routeString := "min-healthy/warning/slack"
	for i := 0; i < b.N; i++ {
		ParseRoutes(routeString)
	}
}

func BenchmarkParseRoutesFor2Routes(b *testing.B) {
	routeString := "min-healthy/warning/slack;min-healthy/critical/slack"
	for i := 0; i < b.N; i++ {
		ParseRoutes(routeString)
	}
}

func BenchmarkParseRoutesFor3Routes(b *testing.B) {
	routeString := "min-healthy/warning/slack;min-healthy/critical/slack;min-healthy/resolved/slack"
	for i := 0; i < b.N; i++ {
		ParseRoutes(routeString)
	}
}

func BenchmarkParseRoutesFor4Routes(b *testing.B) {
	routeString := "min-healthy/warning/slack;min-healthy/critical/slack;min-healthy/resolved/slack;*/pass/*"
	for i := 0; i < b.N; i++ {
		ParseRoutes(routeString)
	}
}
