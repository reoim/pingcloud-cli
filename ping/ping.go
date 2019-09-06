package ping

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httptrace"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
)

type PingOption struct {
	icmp bool
	top  bool
}

type PingDto struct {
	Region  string
	Name    string
	Address string
	Latency time.Duration
}

const (
	httpsTemplate = `` +
		`  DNS Lookup   TCP Connection   TLS Handshake   Server Processing   Content Transfer` + "\n" +
		`[%s  |     %s  |    %s  |        %s  |       %s  ]` + "\n" +
		`            |                |               |                   |                  |` + "\n" +
		`   namelookup:%s      |               |                   |                  |` + "\n" +
		`                       connect:%s     |                   |                  |` + "\n" +
		`                                   pretransfer:%s         |                  |` + "\n" +
		`                                                     starttransfer:%s        |` + "\n" +
		`                                                                                total:%s` + "\n"
)

func (p *PingDto) TestPrint() {
	fmt.Println("Region: " + p.Region)
	fmt.Println("Name: " + p.Name)
	fmt.Println("Address: " + p.Address)
}

func (p *PingDto) Ping() {

	// Init tabwriter
	tr := tabwriter.NewWriter(os.Stdout, 40, 8, 2, '\t', 0)

	// Ping start time
	start := time.Now()

	// Create a new HTTP request
	req, err := http.NewRequest("GET", p.Address, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Send request by default HTTP client
	client := http.DefaultClient
	res, err := client.Do(req)
	result := PingDto{
		Region:  p.Region,
		Name:    p.Name,
		Address: p.Address,
		Latency: time.Since(start), // latency = (current time) -(ping start time)
	}
	if err != nil || res.StatusCode != http.StatusOK {
		fmt.Fprintf(tr, "[%v]\t[%v]\tPing failed with status code: %v", result.Region, result.Name, res.StatusCode)
		fmt.Fprintln(tr)
	} else {
		fmt.Fprintf(tr, "[%v]\t[%v]\t%v", result.Region, result.Name, result.Latency)
		fmt.Fprintln(tr)
	}

	// Flush tabwriter
	tr.Flush()

}

func (p *PingDto) VerbosePing() {

	// Create a new HTTP request
	req, err := http.NewRequest("GET", p.Address, nil)
	if err != nil {
		log.Fatal(err)
	}

	var start, connect, dns, tlsHandshake, tlsHandshakeEnd, firstResponse time.Time

	t := TimeTrace{}
	trace := &httptrace.ClientTrace{
		DNSStart: func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
		DNSDone: func(ddi httptrace.DNSDoneInfo) {
			// fmt.Printf("DNS Done: %v\n", time.Since(dns))
			t.DNSLookup = time.Since(dns)
		},

		ConnectStart: func(network, addr string) { connect = time.Now() },
		ConnectDone: func(network, addr string, err error) {
			// fmt.Printf("Connect time: %v\n", time.Since(connect))
			t.TCPConnect = time.Since(connect)
			t.ConnectTotal = time.Since(dns)
		},

		TLSHandshakeStart: func() { tlsHandshake = time.Now() },
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			// fmt.Printf("TLS Handshake: %v\n", time.Since(tlsHandshake))
			tlsHandshakeEnd = time.Now()
			t.TLSHandshake = time.Since(tlsHandshake)
			t.PreTransfer = time.Since(dns)
		},

		GotFirstResponseByte: func() {
			// fmt.Printf("Time from start to first byte: %v\n", time.Since(start))
			firstResponse = time.Now()
			t.ServerProcess = time.Since(tlsHandshakeEnd)
			t.StartTransfer = time.Since(dns)
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	start = time.Now()
	if _, err := http.DefaultTransport.RoundTrip(req); err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("Total time: %v\n", time.Since(start))
	t.ContentTransfer = time.Since(firstResponse)
	t.TotalTime = time.Since(start)
	t.PrintHttpTrace()
}

type TimeTrace struct {
	DNSLookup       time.Duration
	TCPConnect      time.Duration
	TLSHandshake    time.Duration
	ServerProcess   time.Duration
	ContentTransfer time.Duration
	ConnectTotal    time.Duration
	PreTransfer     time.Duration
	StartTransfer   time.Duration
	TotalTime       time.Duration
}

func (t *TimeTrace) PrintHttpTrace() {

	fmta := func(d time.Duration) string {
		return color.CyanString("%7dms", int(d/time.Millisecond))
	}

	fmtb := func(d time.Duration) string {
		return color.CyanString("%-9s", strconv.Itoa(int(d/time.Millisecond))+"ms")
	}

	colorize := func(s string) string {
		v := strings.Split(s, "\n")
		v[0] = grayscale(16)(v[0])
		return strings.Join(v, "\n")
	}

	printf(colorize(httpsTemplate),
		fmta(t.DNSLookup),       // dns lookup
		fmta(t.TCPConnect),      // tcp connection
		fmta(t.TLSHandshake),    // tls handshake
		fmta(t.ServerProcess),   // server processing
		fmta(t.ContentTransfer), // content transfer
		fmtb(t.DNSLookup),       // namelookup
		fmtb(t.ConnectTotal),    // connect
		fmtb(t.PreTransfer),     // pretransfer
		fmtb(t.StartTransfer),   // starttransfer
		fmtb(t.TotalTime),       // total
	)
}

func printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(color.Output, format, a...)
}

func grayscale(code color.Attribute) func(string, ...interface{}) string {
	return color.New(code + 232).SprintfFunc()
}
