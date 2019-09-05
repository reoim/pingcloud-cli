package gcp

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Endpoints struct {
	Region  string
	Name    string
	Address string
	Latency time.Duration
	Errors  bool
}

func (e *Endpoints) TestPrint() {
	fmt.Println("Region: " + e.Region)
	fmt.Println("Name: " + e.Name)
	fmt.Println("Address: " + e.Address)
}

func (e *Endpoints) Ping() {

	// Ping start time
	startTm := time.Now()

	// Create a new HTTP request
	req, err := http.NewRequest("GET", e.Address, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Send request by default HTTP client
	client := http.DefaultClient
	res, err := client.Do(req)
	result := Endpoints{
		Region:  e.Region,
		Name:    e.Name,
		Address: e.Address,
		Latency: time.Now().Sub(startTm), // latency = (current time) -(ping start time)
		Errors:  false,
	}
	if err != nil {
		result.Errors = true
	}
	if res.StatusCode != http.StatusOK {
		fmt.Errorf("Request to %s(%s) failed with status code: %v", e.Region, e.Address, res.StatusCode)
	}

}

/*
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
*/
