package main

import (
	"fmt"
	"strings"
)

var DefaultRoutes = "*/warning/*;*/critical/*"

// Routes holds the routing information for every checks, alert level combination which Notifier
// should be used.
// Routes are of the form
//  `check-name/check-result/notifier-name`
// Ex. min-healthy/warning/slack
// The default Route(s) are of the form
// 	`*/warning/*` and `*/critical/*` and `*/resolved/*`
type Route struct {
	Check      string
	CheckLevel CheckStatus
	Notifier   string
}

func ParseRoutes(routes string) ([]Route, error) {
	var finalRoutes []Route
	routesAsString := strings.Split(routes, ";")
	for _, routeAsString := range routesAsString {
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

func parseCheckLevel(checkLevel string) (CheckStatus, error) {
	switch strings.ToLower(checkLevel) {
	case "warning":
		return Warning, nil
	case "critical":
		return Critical, nil
	case "pass":
		return Pass, nil
	case "resolved":
		return Resolved, nil
	default:
		return Critical, fmt.Errorf("Expected one of warning / critical / pass / resolved but %s found", strings.ToLower(checkLevel))
	}
}
