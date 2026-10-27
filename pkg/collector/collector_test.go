package collector_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/raphaelthomas/meinberg_ltos_exporter/pkg/collector"
	"github.com/raphaelthomas/meinberg_ltos_exporter/pkg/ltosapi"
)

const metricsPrefix = collector.MetricNamespace + "_"

func TestCollector(t *testing.T) {
	files, err := filepath.Glob("../../tests/testdata/*.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no test data files found")
	}

	for _, jsonFile := range files {
		name := strings.TrimSuffix(filepath.Base(jsonFile), ".json")
		metricsFile := strings.TrimSuffix(jsonFile, ".json") + ".metrics"

		t.Run(name, func(t *testing.T) {
			jsonData, err := os.ReadFile(jsonFile)
			if err != nil {
				t.Fatalf("failed to read %s: %v", jsonFile, err)
			}

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if _, err := w.Write(jsonData); err != nil {
					t.Errorf("failed to write mock response: %v", err)
				}
			}))
			defer srv.Close()

			client, _ := ltosapi.NewClient(srv.URL, "", "", false)
			cfg := collector.Config{
				Timeout:      5 * time.Second,
				System:       true,
				Notification: true,
				Network:      true,
				Storage:      true,
				Clock:        true,
				Receiver:     true,
				NTP:          true,
			}
			c := collector.NewCollector(cfg, client, slog.New(slog.DiscardHandler))

			got := gatherMetrics(t, c)
			gotFiltered := filterMetrics(got, srv.URL)

			if os.Getenv("UPDATE_GOLDEN") == "1" {
				if err := os.WriteFile(metricsFile, []byte(gotFiltered), 0o644); err != nil {
					t.Fatalf("failed to write golden file: %v", err)
				}
				t.Logf("updated golden file %s", metricsFile)
				return
			}

			expected, err := os.ReadFile(metricsFile)
			if err != nil {
				t.Fatalf("failed to read %s (run with UPDATE_GOLDEN=1 to create): %v", metricsFile, err)
			}

			if gotFiltered != string(expected) {
				t.Errorf("metrics mismatch for %s\n\nrun with UPDATE_GOLDEN=1 to update golden files\n\ndiff:\n%s",
					name, lineDiff(string(expected), gotFiltered))
			}
		})
	}
}

// gatherMetrics collects all metrics from the given collector and returns
// them in Prometheus text exposition format.
func gatherMetrics(t *testing.T, c *collector.Collector) string {
	t.Helper()

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(c)

	gathered, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	var buf bytes.Buffer
	enc := expfmt.NewEncoder(&buf, expfmt.NewFormat(expfmt.TypeTextPlain))
	for _, mf := range gathered {
		if err := enc.Encode(mf); err != nil {
			t.Fatalf("failed to encode metric family %s: %v", mf.GetName(), err)
		}
	}
	return buf.String()
}

// metricBlock groups the HELP line, TYPE line, and sample lines for a single metric.
type metricBlock struct {
	help    string
	typ     string
	samples []string
}

// filterMetrics keeps only meinberg_ltos_ metrics, normalises the dynamic
// target URL to a fixed placeholder, and replaces the scrape_duration_seconds
// value with 0 since it varies between runs. The output is sorted by metric
// name for deterministic comparison.
func filterMetrics(input string, target string) string {
	input = strings.ReplaceAll(input, target, "http://localhost")

	blocks := make(map[string]*metricBlock)
	var names []string

	for _, line := range strings.Split(input, "\n") {
		if strings.HasPrefix(line, "# ") {
			parts := strings.Fields(line)
			if len(parts) < 3 || !strings.HasPrefix(parts[2], metricsPrefix) {
				continue
			}
			name := parts[2]
			b := getOrCreateBlock(blocks, name, &names)
			switch parts[1] {
			case "HELP":
				b.help = line
			case "TYPE":
				b.typ = line
			}
			continue
		}

		if !strings.HasPrefix(line, metricsPrefix) {
			continue
		}

		// Replace scrape duration value with placeholder
		if strings.HasPrefix(line, metricsPrefix+"scrape_duration_seconds") {
			if idx := strings.LastIndexByte(line, ' '); idx > 0 {
				line = line[:idx] + " 0"
			}
		}

		name := metricName(line)
		b := getOrCreateBlock(blocks, name, &names)
		b.samples = append(b.samples, line)
	}

	sort.Strings(names)

	var out strings.Builder
	for i, name := range names {
		b := blocks[name]
		if i > 0 {
			out.WriteByte('\n')
		}
		if b.help != "" {
			out.WriteString(b.help)
			out.WriteByte('\n')
		}
		if b.typ != "" {
			out.WriteString(b.typ)
			out.WriteByte('\n')
		}
		sort.Strings(b.samples)
		for _, s := range b.samples {
			out.WriteString(s)
			out.WriteByte('\n')
		}
	}

	return out.String()
}

// getOrCreateBlock returns the metricBlock for the given name, creating it if
// needed and appending the name to the order slice.
func getOrCreateBlock(blocks map[string]*metricBlock, name string, order *[]string) *metricBlock {
	if b, ok := blocks[name]; ok {
		return b
	}
	b := &metricBlock{}
	blocks[name] = b
	*order = append(*order, name)
	return b
}

// metricName extracts the metric name from a sample line, i.e. everything
// before the first '{' or ' '.
func metricName(line string) string {
	if name, _, ok := strings.Cut(line, "{"); ok {
		return name
	}
	name, _, _ := strings.Cut(line, " ")
	return name
}

// lineDiff produces a simple line-by-line diff for test failure output.
func lineDiff(expected, got string) string {
	expectedLines := strings.Split(expected, "\n")
	gotLines := strings.Split(got, "\n")

	var out strings.Builder
	for i := range max(len(expectedLines), len(gotLines)) {
		var e, g string
		if i < len(expectedLines) {
			e = expectedLines[i]
		}
		if i < len(gotLines) {
			g = gotLines[i]
		}
		if e != g {
			fmt.Fprintf(&out, "  line %d:\n    want: %s\n    got:  %s\n", i+1, e, g)
		}
	}
	return out.String()
}
