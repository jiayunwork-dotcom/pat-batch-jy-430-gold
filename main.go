package main

import (
	"flag"
	"fmt"
	"os"

	"pat-batch/internal/parse"
	"pat-batch/internal/report"
	"pat-batch/internal/stats"
)

func main() {
	measPath := flag.String("measurements", "", "path to measurements CSV")
	specsPath := flag.String("specs", "", "path to specs CSV")
	format := flag.String("format", "text", "output format: text or json")
	flag.Parse()

	if *measPath == "" || *specsPath == "" {
		fmt.Fprintln(os.Stderr, "usage: pat-batch -measurements <csv> -specs <csv> [-format text|json]")
		os.Exit(2)
	}

	ms, err := parse.ReadMeasurements(*measPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	specs, err := parse.ReadSpecs(*specsPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	paramStats := stats.ComputeParamStats(ms, specs)

	byBatch := map[string][]stats.Measurement{}
	var order []string
	for _, m := range ms {
		if _, ok := byBatch[m.Batch]; !ok {
			order = append(order, m.Batch)
		}
		byBatch[m.Batch] = append(byBatch[m.Batch], m)
	}
	batches := make([]stats.BatchResult, 0, len(order))
	for _, b := range order {
		batches = append(batches, stats.EvaluateBatch(byBatch[b], specs))
	}

	out, err := report.Format(paramStats, batches, *format)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Print(out)

	if hasDeviations(batches) {
		os.Exit(1)
	}
}

func hasDeviations(bs []stats.BatchResult) bool {
	for _, r := range bs {
		if len(r.OOS) > 0 || len(r.OOT) > 0 {
			return true
		}
	}
	return false
}
