package ping

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"text/tabwriter"
	"time"
)

type Endpoints struct {
	Region  string
	Name    string
	Address string
	Latency time.Duration
}

func (e *Endpoints) TestPrint() {
	fmt.Println("Region: " + e.Region)
	fmt.Println("Name: " + e.Name)
	fmt.Println("Address: " + e.Address)
}

func (e *Endpoints) Ping() {

	// Init tabwriter
	tr := tabwriter.NewWriter(os.Stdout, 40, 8, 2, '\t', 0)

	// Ping start time
	start := time.Now()

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
		Latency: time.Now().Sub(start), // latency = (current time) -(ping start time)
	}
	if err != nil || res.StatusCode != http.StatusOK {
		fmt.Fprintf(tr, "[%v]\t[%v]\tPing failed with status code: %v", result.Region, result.Name, res.StatusCode)
		fmt.Fprintln(tr)
	} else {
		fmt.Fprintf(tr, "[%v]\t[%v]\tLatency: %v", result.Region, result.Name, result.Latency)
		fmt.Fprintln(tr)
	}

	// Flush tabwriter
	tr.Flush()

}
