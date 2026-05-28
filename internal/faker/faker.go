package faker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/brianvoe/gofakeit/v6"
)

var tokenRe = regexp.MustCompile(`\{\{([^}]+)\}\}`)

// Resolve walks any value (string, map, slice) and replaces all {{tokens}}.
// r is the incoming request, used for {{param.*}}, {{query.*}}, {{body.*}}.
func Resolve(v interface{}, r *http.Request, params map[string]string) interface{} {
	switch val := v.(type) {
	case string:
		return resolveString(val, r, params)
	case map[string]interface{}:
		out := make(map[string]interface{}, len(val))
		for k, v2 := range val {
			out[k] = Resolve(v2, r, params)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(val))
		for i, item := range val {
			out[i] = Resolve(item, r, params)
		}
		return out
	default:
		return v
	}
}

// Warnings returns unknown tokens found in a value at startup (deduplicated).
func Warnings(v interface{}, params map[string]string) []string {
	seen := map[string]bool{}
	var walk func(interface{})
	walk = func(v interface{}) {
		switch val := v.(type) {
		case string:
			for _, m := range tokenRe.FindAllStringSubmatch(val, -1) {
				tok := strings.TrimSpace(m[1])
				if !isKnown(tok, params) && !seen[tok] {
					seen[tok] = true
				}
			}
		case map[string]interface{}:
			for _, v2 := range val {
				walk(v2)
			}
		case []interface{}:
			for _, item := range val {
				walk(item)
			}
		}
	}
	walk(v)

	var out []string
	for tok := range seen {
		out = append(out, fmt.Sprintf("unknown token {{%s}} — will render as \"\"", tok))
	}
	return out
}

func resolveString(s string, r *http.Request, params map[string]string) interface{} {
	// If the entire string is a single token, return the native type (int, bool, etc.)
	if m := tokenRe.FindStringSubmatch(s); m != nil && m[0] == s {
		return resolveToken(strings.TrimSpace(m[1]), r, params)
	}
	// Otherwise inline-replace all tokens with their string representation
	return tokenRe.ReplaceAllStringFunc(s, func(raw string) string {
		tok := strings.TrimSpace(raw[2 : len(raw)-2])
		v := resolveToken(tok, r, params)
		return fmt.Sprintf("%v", v)
	})
}

func resolveToken(tok string, r *http.Request, params map[string]string) interface{} {
	// Dynamic lookups
	if strings.HasPrefix(tok, "param.") {
		key := strings.TrimPrefix(tok, "param.")
		return params[key] // empty string if not found — decision #5
	}
	if strings.HasPrefix(tok, "query.") && r != nil {
		key := strings.TrimPrefix(tok, "query.")
		return r.URL.Query().Get(key)
	}
	if strings.HasPrefix(tok, "body.") && r != nil {
		key := strings.TrimPrefix(tok, "body.")
		return bodyField(r, key) // empty string on GET — decision #6
	}

	// int(min,max)
	if strings.HasPrefix(tok, "int(") {
		return parseRangeInt(tok)
	}
	// float(min,max)
	if strings.HasPrefix(tok, "float(") {
		return parseRangeFloat(tok)
	}

	switch tok {
	case "uuid":
		return gofakeit.UUID()
	case "fullName":
		return gofakeit.Name()
	case "firstName":
		return gofakeit.FirstName()
	case "lastName":
		return gofakeit.LastName()
	case "email":
		return gofakeit.Email()
	case "phone":
		return gofakeit.Phone()
	case "isoDate":
		return gofakeit.Date().Format("2006-01-02T15:04:05Z")
	case "date":
		return gofakeit.Date().Format("2006-01-02")
	case "bool":
		return gofakeit.Bool()
	case "sentence":
		return gofakeit.Sentence(6)
	case "word":
		return gofakeit.Word()
	case "url":
		return gofakeit.URL()
	default:
		return "" // unknown token — decision #4
	}
}

func isKnown(tok string, params map[string]string) bool {
	if strings.HasPrefix(tok, "param.") ||
		strings.HasPrefix(tok, "query.") ||
		strings.HasPrefix(tok, "body.") ||
		strings.HasPrefix(tok, "int(") ||
		strings.HasPrefix(tok, "float(") {
		return true
	}
	known := map[string]bool{
		"uuid": true, "fullName": true, "firstName": true, "lastName": true,
		"email": true, "phone": true, "isoDate": true, "date": true,
		"bool": true, "sentence": true, "word": true, "url": true,
	}
	return known[tok]
}

func bodyField(r *http.Request, key string) interface{} {
	if r == nil || r.Body == nil {
		return ""
	}
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return ""
	}
	if v, ok := payload[key]; ok {
		return v
	}
	return ""
}

func parseRangeInt(tok string) interface{} {
	inner := tok[4 : len(tok)-1]
	parts := strings.SplitN(inner, ",", 2)
	if len(parts) != 2 {
		return 0
	}
	min, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	max, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil {
		return 0
	}
	return gofakeit.IntRange(min, max)
}

func parseRangeFloat(tok string) interface{} {
	inner := tok[6 : len(tok)-1]
	parts := strings.SplitN(inner, ",", 2)
	if len(parts) != 2 {
		return 0.0
	}
	min, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	max, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err1 != nil || err2 != nil {
		return 0.0
	}
	return gofakeit.Float64Range(min, max)
}
