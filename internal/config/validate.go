package config

import (
	"fmt"
	"strings"
)

// validate checks for hard errors and collects warnings.
// Returns (warnings, errors).
func validate(cfg *Config) ([]string, []string) {
	var warnings, errs []string

	seen := map[string]int{} // "METHOD /path" -> first index

	for i, ep := range cfg.Endpoints {
		method := strings.ToUpper(ep.Method)
		key := method + " " + ep.Path

		if prev, exists := seen[key]; exists {
			errs = append(errs, fmt.Sprintf(
				"duplicate route %s (endpoint %d conflicts with endpoint %d)",
				key, i+1, prev+1,
			))
		} else {
			seen[key] = i
		}

		// Warn: count on a non-array body
		if ep.Response != nil && ep.Response.Count > 0 {
			if _, isSlice := ep.Response.Body.([]interface{}); !isSlice {
				if ep.Response.Body != nil {
					errs = append(errs, fmt.Sprintf(
						"endpoint %s %s: 'count' requires body to be an array",
						method, ep.Path,
					))
				}
			}
		}

		// Warn: count + paginated together
		if ep.Response != nil && ep.Response.Paginated && ep.Response.Count > 0 {
			warnings = append(warnings, fmt.Sprintf(
				"endpoint %s %s: 'count' is ignored when paginated: true",
				method, ep.Path,
			))
		}

		// Warn: variants-only endpoint with no base response and no default variant
		if ep.Response == nil && len(ep.Variants) > 0 {
			hasDefault := false
			for _, v := range ep.Variants {
				if v.Default {
					hasDefault = true
					break
				}
			}
			if !hasDefault {
				warnings = append(warnings, fmt.Sprintf(
					"endpoint %s %s: has variants but no default or base response — unmatched requests return 200 {}",
					method, ep.Path,
				))
			}
		}

		// Warn: base_path already in path
		if cfg.Info.BasePath != "" && strings.HasPrefix(ep.Path, cfg.Info.BasePath) {
			warnings = append(warnings, fmt.Sprintf(
				"endpoint %s %s: path already contains base_path %q — did you mean %s?",
				method, ep.Path, cfg.Info.BasePath,
				strings.TrimPrefix(ep.Path, cfg.Info.BasePath),
			))
		}
	}

	return warnings, errs
}

// normalize uppercases all methods in place.
func normalize(cfg *Config) {
	for i := range cfg.Endpoints {
		cfg.Endpoints[i].Method = strings.ToUpper(cfg.Endpoints[i].Method)
	}
}
