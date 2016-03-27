package routes

import (
	"fmt"
	"strings"

	"github.com/ashwanthkumar/marathon-alerts/checks"
	"github.com/ryanuber/go-glob"
)

var DefaultRoutes = "*/warning/*;*/critical/*;*/resolved/*"

// Routes holds the routing information for every checks, alert level combination which Notifier
// should be used.
// Routes are of the form
//  `check-name/check-result/notifier-name`
// Ex. min-healthy/warning/slack
// The default Route(s) are of the form
// 	`*/warning/*` and `*/critical/*` and `*/resolved/*`
type Route struct {
	Check      string
	CheckLevel checks.CheckStatus
	Notifier   string
}

func (r *Route) Match(check checks.AppCheck) bool {
	nameMatches := glob.Glob(r.Check, check.CheckName)
	checkLevelMatches := r.CheckLevel == check.Result
	return nameMatches && checkLevelMatches
}

func (r *Route) MatchNotifier(notifier string) bool {
	return glob.Glob(r.Notifier, notifier)
}

func (r *Route) MatchCheckResult(level checks.CheckStatus) bool {
	return r.CheckLevel == level
}

func ParseRoutes(routes string) ([]Route, error) {
	var finalRoutes []Route
	routesAsString := strings.Split(routes, ";")
	for _, routeAsString := range routesAsString {
		if routes != "" && routeAsString == "" {
			continue
		}
		segments := strings.Split(routeAsString, "/")
		if len(segments) != 3 {
			return nil, fmt.Errorf("Expected 3 parts in %s, separated by `/` but %d found", routeAsString, len(segments))
		}
		checkLevel, err := parseCheckLevel(segments[1])
		if err != nil {
			return nil, err
		}
		route := Route{
			Check:      segments[0],
			CheckLevel: checkLevel,
			Notifier:   segments[2],
		}
		finalRoutes = append(finalRoutes, route)
	}
	return finalRoutes, nil
}

func parseCheckLevel(checkLevel string) (checks.CheckStatus, error) {
	switch strings.ToLower(checkLevel) {
	case "warning":
		return checks.Warning, nil
	case "critical":
		return checks.Critical, nil
	case "pass":
		return checks.Pass, nil
	case "resolved":
		return checks.Resolved, nil
	default:
		return checks.Critical, fmt.Errorf("Expected one of warning / critical / pass / resolved but %s found", strings.ToLower(checkLevel))
	}
}
