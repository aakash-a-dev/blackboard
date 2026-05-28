package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aakash-a-dev/blackboard/internal/config"
	"github.com/aakash-a-dev/blackboard/internal/faker"
	"github.com/aakash-a-dev/blackboard/internal/logger"
)

// Build returns an http.Handler for a single endpoint.
func Build(ep config.Endpoint, globalLatencyMs int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		params := extractParams(ep.Path, r)

		// 1. Apply latency
		latency := globalLatencyMs
		if ep.Behavior != nil && ep.Behavior.LatencyMs > 0 {
			latency = ep.Behavior.LatencyMs
		}
		if latency > 0 {
			time.Sleep(time.Duration(latency) * time.Millisecond)
		}

		// 2. Roll error_rate — decision #7
		if ep.Behavior != nil && ep.Behavior.ErrorRate > 0 {
			if rand.Float64() < ep.Behavior.ErrorRate {
				status := ep.Behavior.ErrorStatus
				if status == 0 {
					status = http.StatusInternalServerError
				}
				body := ep.Behavior.ErrorBody
				if body == nil {
					body = map[string]string{"error": http.StatusText(status)}
				}
				writeJSON(w, status, body)
				logger.Request(status, r.Method, r.URL.Path, time.Since(start))
				return
			}
		}

		// 3. Evaluate variants — decision #8, #9
		if resp := matchVariant(ep.Variants, r); resp != nil {
			writeResponse(w, r, resp, params)
			logger.Request(resp.Status, r.Method, r.URL.Path, time.Since(start))
			return
		}

		// 4. Base response
		if ep.Response != nil {
			// Validate request body if rules defined.
			// Read body once, then restore so faker can re-read it.
			if ep.Request != nil {
				bodyBytes, _ := io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
				if errMsg := validateRequest(bodyBytes, ep.Request); errMsg != nil {
					writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
						"error":   "Validation failed",
						"details": errMsg,
					})
					logger.Request(http.StatusUnprocessableEntity, r.Method, r.URL.Path, time.Since(start))
					return
				}
			}
			writeResponse(w, r, ep.Response, params)
			logger.Request(ep.Response.Status, r.Method, r.URL.Path, time.Since(start))
			return
		}

		// 5. Nothing matched — decision #8
		writeJSON(w, http.StatusOK, map[string]interface{}{})
		logger.Request(http.StatusOK, r.Method, r.URL.Path, time.Since(start))
	})
}

func writeResponse(w http.ResponseWriter, r *http.Request, resp *config.Response, params map[string]string) {
	// Custom headers
	for k, v := range resp.Headers {
		resolved := faker.Resolve(v, r, params)
		w.Header().Set(k, toString(resolved))
	}

	if resp.Paginated {
		writePaginated(w, r, resp, params)
		return
	}

	body := faker.Resolve(resp.Body, r, params)

	// Generate array via count
	if resp.Count > 0 {
		if template, ok := resp.Body.([]interface{}); ok && len(template) > 0 {
			items := make([]interface{}, resp.Count)
			for i := range items {
				items[i] = faker.Resolve(template[0], r, params)
			}
			body = items
		}
	}

	writeJSON(w, resp.Status, body)
}

func writePaginated(w http.ResponseWriter, r *http.Request, resp *config.Response, params map[string]string) {
	page := queryInt(r, "page", 1)
	limit := queryInt(r, "limit", 10)
	total := resp.Total

	var template interface{}
	if arr, ok := resp.Body.([]interface{}); ok && len(arr) > 0 {
		template = arr[0]
	} else {
		template = resp.Body
	}

	items := make([]interface{}, limit)
	for i := range items {
		items[i] = faker.Resolve(template, r, params)
	}

	totalPages := 1
	if limit > 0 && total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	writeJSON(w, resp.Status, map[string]interface{}{
		"data":        items,
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": totalPages,
	})
}

// matchVariant finds the first matching variant. Decision #8, #9.
func matchVariant(variants []config.Variant, r *http.Request) *config.Response {
	var defaultResp *config.Response
	for _, v := range variants {
		if v.Default {
			defaultResp = v.Response
			continue
		}
		if v.Condition != nil && matchCondition(v.Condition, r) {
			return v.Response
		}
	}
	return defaultResp
}

// matchCondition checks all keys under query/headers (AND logic — decision #9).
func matchCondition(c *config.Condition, r *http.Request) bool {
	for k, v := range c.Query {
		if r.URL.Query().Get(k) != v {
			return false
		}
	}
	for k, v := range c.Headers {
		if !strings.EqualFold(r.Header.Get(k), v) {
			return false
		}
	}
	return true
}

// validateRequest checks required fields and body schema. Decision #11, #12.
func validateRequest(bodyBytes []byte, req *config.Request) []string {
	var errs []string

	var body map[string]interface{}
	if len(bodyBytes) > 0 {
		json.Unmarshal(bodyBytes, &body) //nolint:errcheck
	}
	if body == nil {
		body = map[string]interface{}{}
	}

	for _, field := range req.RequiredFields {
		if _, ok := body[field]; !ok {
			errs = append(errs, field+": required")
		}
	}

	for field, schema := range req.BodySchema {
		val, present := body[field]
		if !present {
			continue
		}
		switch schema.Type {
		case "string":
			s, ok := val.(string)
			if !ok {
				errs = append(errs, field+": must be a string")
				continue
			}
			if schema.MinLength > 0 && len(s) < schema.MinLength {
				errs = append(errs, field+": too short (min "+strconv.Itoa(schema.MinLength)+")")
			}
			if schema.Format == "email" && !strings.Contains(s, "@") {
				errs = append(errs, field+": invalid email format")
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

// extractParams pulls :param values using Go 1.22 r.PathValue().
func extractParams(pattern string, r *http.Request) map[string]string {
	params := map[string]string{}
	for _, seg := range strings.Split(pattern, "/") {
		if strings.HasPrefix(seg, ":") {
			name := seg[1:]
			params[name] = r.PathValue(name)
		}
	}
	return params
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body != nil {
		json.NewEncoder(w).Encode(body) //nolint:errcheck
	}
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func queryInt(r *http.Request, key string, def int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
