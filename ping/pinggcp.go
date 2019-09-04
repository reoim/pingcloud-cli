package ping

import (
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"
	"time"
)

var client *http.Client

type endpoints struct {
	region  string
	name    string
	address string
}

func (e *endpoints) gcpHTTP() result {
	return e.ping(func() error {
		req, _ := http.NewRequest("GET", "http://"+e.address+"/ping", nil)
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("status code: %v", res.StatusCode)
		}
		return nil
	})
}

func (e *endpoints) ping(fn func() error) result {

	start := time.Now()
	err := fn()
	duration := time.Since(start)

	r := result{
		region:    e.region,
		durations: []time.Duration{duration},
	}
	if err != nil {
		r.errors++
	}

	return r
}

type result struct {
	region    string
	durations []time.Duration
	errors    int
}

type worker struct {
	inputs  chan input
	outputs chan output
}

func (w *worker) start() {
	for worker := 0; worker < concurrency; worker++ {
		go func() {
			for m := range w.inputs {
				o := m.HTTP()
				w.outputs <- o
			}
		}()
	}
}

func (w *worker) reportAll() {
	w.inputs = make(chan input, concurrency)
	w.outputs = make(chan output, w.size(region))
	for i := 0; i < number; i++ {
		for r, e := range endpoints {
			w.inputs <- input{region: r, endpoint: e}
		}
	}
	close(w.inputs)

	sorted := w.sortOutput()
	tr := tabwriter.NewWriter(os.Stdout, 3, 2, 2, ' ', 0)
	for i, a := range sorted {
		fmt.Fprintf(tr, "%2d.\t[%v]\t%v", i+1, a.region, a.median())
		if a.errors > 0 {
			fmt.Fprintf(tr, "\t(%d errors)", a.errors)
		}
		fmt.Fprintln(tr)
	}
	tr.Flush()
}
