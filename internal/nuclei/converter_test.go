package nuclei

import "testing"

func TestConvertHTTPTemplateSplitsPathAndQuery(t *testing.T) {
	input := []byte(`id: cve-demo
info:
  name: Demo template
  severity: high
  tags: cve,rce
http:
  - method: GET
    path:
      - "{{BaseURL}}/%%32%65?a=1&a=2&empty="
    headers:
      X-Test: demo
    matchers:
      - type: status
        status: [200]
`)

	report, err := Convert(input, "demo.yaml")
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}

	if report.TemplateID != "cve-demo" {
		t.Fatalf("TemplateID = %q, want cve-demo", report.TemplateID)
	}
	if len(report.Requests) != 1 {
		t.Fatalf("converted requests = %d, want 1", len(report.Requests))
	}

	request := report.Requests[0]
	if request.Method != "GET" {
		t.Errorf("Method = %q, want GET", request.Method)
	}
	path, ok := request.Path.(map[string]interface{})
	if !ok {
		t.Fatalf("Path type = %T, want map[string]interface{}", request.Path)
	}
	if path["value"] != "/%%32%65" {
		t.Errorf("path.value = %v, want /%%%%32%%65", path["value"])
	}
	if path["raw"] != true {
		t.Errorf("path.raw = %v, want true", path["raw"])
	}

	query, ok := request.Query.([]KeyValue)
	if !ok {
		t.Fatalf("Query type = %T, want []KeyValue", request.Query)
	}
	if len(query) != 3 {
		t.Fatalf("query length = %d, want 3", len(query))
	}
	if query[0] != (KeyValue{Key: "a", Value: "1"}) || query[1] != (KeyValue{Key: "a", Value: "2"}) || query[2] != (KeyValue{Key: "empty", Value: ""}) {
		t.Fatalf("query = %#v, want duplicate-preserving ordered key-value pairs", query)
	}

	headers, ok := request.Headers.(map[string]interface{})
	if !ok {
		t.Fatalf("Headers type = %T, want map[string]interface{}", request.Headers)
	}
	if headers["X-Test"] != "demo" {
		t.Errorf("header X-Test = %q, want demo", headers["X-Test"])
	}
	if len(report.Unsupported) != 1 {
		t.Fatalf("unsupported count = %d, want 1", len(report.Unsupported))
	}
	if report.Unsupported[0].Feature != "matchers" {
		t.Errorf("unsupported feature = %q, want matchers", report.Unsupported[0].Feature)
	}
}

func TestAnalyzeReportsUnsupportedProtocolSections(t *testing.T) {
	input := []byte(`id: mixed-template
info:
  name: Mixed template
dns:
  - name: "{{FQDN}}"
flow: http(1) && dns(1)
`)

	report, err := Convert(input, "mixed.yaml")
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}
	if len(report.Requests) != 0 {
		t.Fatalf("converted requests = %d, want 0", len(report.Requests))
	}

	features := map[string]bool{}
	for _, item := range report.Unsupported {
		features[item.Feature] = true
	}
	for _, feature := range []string{"dns", "flow"} {
		if !features[feature] {
			t.Fatalf("unsupported feature %q not reported; got %#v", feature, report.Unsupported)
		}
	}
}
