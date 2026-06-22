package nuclei

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/secureta/s2http-request/internal/config"
	"gopkg.in/yaml.v3"
)

// KeyValue is the ordered key-value representation emitted for converted queries.
type KeyValue = config.KeyValue

// Report describes the conversion result and any Nuclei features that were not converted.
type Report struct {
	Source      string                  `json:"source,omitempty" yaml:"source,omitempty"`
	TemplateID  string                  `json:"template_id,omitempty" yaml:"template_id,omitempty"`
	Info        map[string]interface{}  `json:"info,omitempty" yaml:"info,omitempty"`
	Requests    []*config.RequestConfig `json:"requests" yaml:"requests"`
	Unsupported []UnsupportedFeature    `json:"unsupported,omitempty" yaml:"unsupported,omitempty"`
}

// UnsupportedFeature records a Nuclei feature that needs a later evaluator/runner layer.
type UnsupportedFeature struct {
	Path    string `json:"path" yaml:"path"`
	Feature string `json:"feature" yaml:"feature"`
	Reason  string `json:"reason" yaml:"reason"`
}

// Convert analyzes a Nuclei template and converts the supported HTTP request subset.
func Convert(data []byte, source string) (*Report, error) {
	var template map[string]interface{}
	if err := yaml.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to parse nuclei template: %w", err)
	}

	report := &Report{
		Source:   source,
		Requests: []*config.RequestConfig{},
	}
	if id, ok := template["id"].(string); ok {
		report.TemplateID = id
	}
	if info, ok := template["info"].(map[string]interface{}); ok {
		report.Info = info
	}

	reportUnsupportedProtocolSections(report, template)

	sections := httpSections(template)
	for sectionIndex, section := range sections {
		if _, ok := section["raw"]; ok {
			report.Unsupported = append(report.Unsupported, UnsupportedFeature{
				Path:    fmt.Sprintf("http[%d].raw", sectionIndex),
				Feature: "raw",
				Reason:  "raw HTTP blocks need byte-level request support and are not converted yet",
			})
			continue
		}
		if _, ok := section["matchers"]; ok {
			report.Unsupported = append(report.Unsupported, UnsupportedFeature{
				Path:    fmt.Sprintf("http[%d].matchers", sectionIndex),
				Feature: "matchers",
				Reason:  "matchers require a response evaluator and are preserved as unsupported metadata for now",
			})
		}
		if _, ok := section["extractors"]; ok {
			report.Unsupported = append(report.Unsupported, UnsupportedFeature{
				Path:    fmt.Sprintf("http[%d].extractors", sectionIndex),
				Feature: "extractors",
				Reason:  "extractors require response-to-request state handling and are not converted yet",
			})
		}

		converted, err := convertHTTPSection(section)
		if err != nil {
			report.Unsupported = append(report.Unsupported, UnsupportedFeature{
				Path:    fmt.Sprintf("http[%d]", sectionIndex),
				Feature: "http",
				Reason:  err.Error(),
			})
			continue
		}
		report.Requests = append(report.Requests, converted...)
	}

	return report, nil
}

func reportUnsupportedProtocolSections(report *Report, template map[string]interface{}) {
	for _, feature := range []string{"dns", "tcp", "ssl", "headless", "code", "javascript", "network", "file", "flow"} {
		if _, ok := template[feature]; ok {
			report.Unsupported = append(report.Unsupported, UnsupportedFeature{
				Path:    feature,
				Feature: feature,
				Reason:  "only HTTP request conversion is supported in the initial nuclei converter",
			})
		}
	}
}

func httpSections(template map[string]interface{}) []map[string]interface{} {
	for _, key := range []string{"http", "requests"} {
		value, ok := template[key]
		if !ok {
			continue
		}
		items, ok := value.([]interface{})
		if !ok {
			continue
		}
		sections := make([]map[string]interface{}, 0, len(items))
		for _, item := range items {
			section, ok := item.(map[string]interface{})
			if ok {
				sections = append(sections, section)
			}
		}
		return sections
	}
	return nil
}

func convertHTTPSection(section map[string]interface{}) ([]*config.RequestConfig, error) {
	method := "GET"
	if value, ok := section["method"].(string); ok && value != "" {
		method = strings.ToUpper(value)
	}

	paths := stringList(section["path"])
	if len(paths) == 0 {
		return nil, fmt.Errorf("HTTP section has no supported path entries")
	}

	headers := map[string]interface{}{}
	if headerMap, ok := section["headers"].(map[string]interface{}); ok {
		for key, value := range headerMap {
			headers[key] = value
		}
	}

	requests := make([]*config.RequestConfig, 0, len(paths))
	for _, rawPath := range paths {
		pathValue, query := splitNucleiPath(rawPath)
		request := &config.RequestConfig{
			Method:  method,
			Path:    convertedPath(pathValue),
			Headers: headers,
		}
		if len(query) > 0 {
			request.Query = query
		}
		if body, ok := section["body"]; ok {
			request.Body = body
		}
		requests = append(requests, request)
	}

	return requests, nil
}

func stringList(value interface{}) []string {
	switch typed := value.(type) {
	case string:
		return []string{typed}
	case []interface{}:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			if str, ok := item.(string); ok {
				values = append(values, str)
			}
		}
		return values
	default:
		return nil
	}
}

func splitNucleiPath(value string) (string, []KeyValue) {
	path := stripBaseURLPlaceholder(value)
	queryStart := strings.Index(path, "?")
	if queryStart == -1 {
		return defaultPath(path), nil
	}

	rawQuery := path[queryStart+1:]
	path = path[:queryStart]
	pairs := strings.Split(rawQuery, "&")
	query := make([]KeyValue, 0, len(pairs))
	for _, pair := range pairs {
		if pair == "" {
			continue
		}
		key, val, hasValue := strings.Cut(pair, "=")
		decodedKey, err := url.QueryUnescape(key)
		if err != nil {
			decodedKey = key
		}
		decodedValue := ""
		if hasValue {
			decodedValue, err = url.QueryUnescape(val)
			if err != nil {
				decodedValue = val
			}
		}
		query = append(query, KeyValue{Key: decodedKey, Value: decodedValue})
	}

	return defaultPath(path), query
}

func stripBaseURLPlaceholder(value string) string {
	for _, prefix := range []string{"{{BaseURL}}", "{{baseURL}}", "{{RootURL}}", "{{Hostname}}"} {
		value = strings.TrimPrefix(value, prefix)
	}
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		parsed, err := url.Parse(value)
		if err == nil {
			requestURI := parsed.EscapedPath()
			if requestURI == "" {
				requestURI = "/"
			}
			if parsed.RawQuery != "" {
				requestURI += "?" + parsed.RawQuery
			}
			return requestURI
		}
	}
	return value
}

func defaultPath(path string) string {
	if path == "" {
		return "/"
	}
	return path
}

func convertedPath(path string) interface{} {
	if strings.Contains(path, "%") {
		return map[string]interface{}{
			"value": path,
			"raw":   true,
		}
	}
	return path
}
