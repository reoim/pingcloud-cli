package ping

import (
	"crypto/tls"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"net/http/httptrace"
	"os"
	"runtime"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
)

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

// CmdOption is to save cmd option
type CmdOption struct {
	Option     string
	OptionName string
	ListFlg    bool
	Args       []string
}

// StartCmd is to excute functions and differentiate report depends on cmd option
func (c *CmdOption) StartCmd() {

	var regions = make(map[string]string)
	var endpoints = make(map[string]string)
	var err error
	var csvpath string
	homedir := os.Getenv("PINGCLOUD_DIR")

	fmt.Println("")
	// Read endpoints

	if runtime.GOOS == "windows" {
		csvpath = fmt.Sprintf("\\endpoints\\%s.csv", c.Option)
	} else {
		csvpath = fmt.Sprintf("/endpoints/%s.csv", c.Option)
	}

	regions, endpoints, err = ReadEndpoints(homedir + csvpath)
	if err != nil {
		log.Fatal(err)
	}

	if c.ListFlg {

		// Init tabwriter
		tr := tabwriter.NewWriter(os.Stdout, 40, 8, 2, '\t', 0)
		fmt.Fprintf(tr, "%s Region Code\t%s Region Name", c.OptionName, c.OptionName)
		fmt.Fprintln(tr)
		fmt.Fprintf(tr, "------------------------------\t------------------------------")
		fmt.Fprintln(tr)
		for r, n := range regions {
			fmt.Fprintf(tr, "[%v]\t[%v]", r, n)
			fmt.Fprintln(tr)
		}
		// Flush tabwriter
		tr.Flush()
	} else if len(c.Args) == 0 {

		// Init tabwriter
		tr := tabwriter.NewWriter(os.Stdout, 40, 8, 2, '\t', 0)
		fmt.Fprintf(tr, "%s Region Code\t%s Region Name\tLatency", c.OptionName, c.OptionName)
		fmt.Fprintln(tr)
		fmt.Fprintf(tr, "------------------------------\t------------------------------\t------------------------------")
		fmt.Fprintln(tr)

		// Flush tabwriter
		tr.Flush()

		for r, i := range endpoints {
			p := PingDto{
				Region:  r,
				Name:    regions[r],
				Address: i,
			}
			p.Ping()
		}
		fmt.Println("")
		fmt.Println("You can also add region after command if you want http trace information of the specific region")
		if c.Option == "gcp" {
			fmt.Printf("ex> pingcloud-cli %s us-central1\n", c.Option)
		} else if c.Option == "aws" {
			fmt.Printf("ex> pingcloud-cli %s us-east-1\n", c.Option)
		} else if c.Option == "azure" {
			fmt.Printf("ex> pingcloud-cli %s koreasouth\n", c.Option)
		}

	} else {
		for _, r := range c.Args {
			if i, ok := endpoints[r]; ok {
				p := PingDto{
					Region:  r,
					Name:    endpoints[r],
					Address: i,
				}
				p.VerbosePing()
			} else {
				fmt.Printf("Region code [%v] is wrong.  To check available region codes run the command with -l or --list flag\n", r)
				fmt.Printf("Usage: pingcloud-cli %s -l\n", c.Option)
				fmt.Printf("Usage: pingcloud-cli %s --list\n", c.Option)

			}
		}

	}
	fmt.Println("")
}

// ReadEndpoints parse CSV file into map of region names and endpoints
func ReadEndpoints(filename string) (map[string]string, map[string]string, error) {

	region := make(map[string]string)
	endpoints := make(map[string]string)

	// Open CSV file
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Read File into a Variable
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	for i, line := range lines {
		if i != 0 {
			region[line[0]] = line[1]
			endpoints[line[0]] = line[2]
		}
	}
	return region, endpoints, err
}

// PingDto is for endpoint and reponse time information
type PingDto struct {
	Region  string
	Name    string
	Address string
	Latency time.Duration
}

// TestPrint prints the PingDto struct.
func (p *PingDto) TestPrint() {
	fmt.Println("Region: " + p.Region)
	fmt.Println("Name: " + p.Name)
	fmt.Println("Address: " + p.Address)
}

// Ping Send HTTP(S) request to endpoint and report its response time
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
	if err != nil || res == nil {
		log.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		fmt.Fprintf(tr, "[%v]\t[%v]\tPing failed with status code: %v", p.Region, p.Name, res.StatusCode)
		fmt.Fprintln(tr)
	} else {
		p.Latency = time.Since(start)
		lowLatency, _ := time.ParseDuration("200ms")
		highLatency, _ := time.ParseDuration("999ms")
		if p.Latency < lowLatency {
			fmt.Fprintf(tr, "[%v]\t[%v]\t%v", p.Region, p.Name, color.GreenString(p.Latency.String()))
		} else if p.Latency < highLatency {
			fmt.Fprintf(tr, "[%v]\t[%v]\t%v", p.Region, p.Name, color.YellowString(p.Latency.String()))
		} else {
			fmt.Fprintf(tr, "[%v]\t[%v]\t%v", p.Region, p.Name, color.RedString(p.Latency.String()))
		}
		fmt.Fprintln(tr)
	}

	// Flush tabwriter
	tr.Flush()

}

// VerbosePing send HTTP(S) request to endpoint and report its respons time in httpstat style
func (p *PingDto) VerbosePing() {

	// Create a new HTTP request
	req, err := http.NewRequest("GET", p.Address, nil)
	if err != nil {
		log.Fatal(err)
	}

	var dnsStart, dnsEnd, tcpConnectStart, tcpConnectEnd, tlsHandshakeStart, tlsHandshakeEnd, serverStart, serverEnd, httpEnd time.Time

	trace := &httptrace.ClientTrace{
		DNSStart: func(dsi httptrace.DNSStartInfo) { dnsStart = time.Now() },
		DNSDone: func(ddi httptrace.DNSDoneInfo) {
			dnsEnd = time.Now()
		},

		ConnectStart: func(network, addr string) {
			tcpConnectStart = time.Now()

			if dnsStart.IsZero() {
				dnsStart = tcpConnectStart
				dnsEnd = tcpConnectStart
			}
		},
		ConnectDone: func(network, addr string, err error) {
			tcpConnectEnd = time.Now()
		},

		TLSHandshakeStart: func() { tlsHandshakeStart = time.Now() },
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			tlsHandshakeEnd = time.Now()
		},

		WroteRequest: func(info httptrace.WroteRequestInfo) {
			serverStart = time.Now()

			if tlsHandshakeStart.IsZero() {
				tlsHandshakeStart = serverStart
				tlsHandshakeEnd = serverStart
			}
		},

		GotFirstResponseByte: func() {
			serverEnd = time.Now()
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	if _, err := http.DefaultTransport.RoundTrip(req); err != nil {
		log.Fatal(err)
	}
	httpEnd = time.Now()
	t := TimeTrace{
		DNSLookup:       dnsEnd.Sub(dnsStart),
		TCPConnect:      tcpConnectEnd.Sub(tcpConnectStart),
		TLSHandshake:    tlsHandshakeEnd.Sub(tlsHandshakeStart),
		ServerProcess:   serverEnd.Sub(serverStart),
		ContentTransfer: httpEnd.Sub(serverEnd),
		NameLookup:      dnsEnd.Sub(dnsStart),
		ConnectTotal:    tcpConnectEnd.Sub(dnsStart),
		PreTransfer:     tlsHandshakeEnd.Sub(dnsStart),
		StartTransfer:   serverEnd.Sub(dnsStart),
		TotalTime:       httpEnd.Sub(dnsStart),
	}

	fmt.Println("")
	fmt.Printf("HTTP(S) trace of region [%v] - %v\n", p.Region, p.Name)
	fmt.Println("-------------------------------------------------------------------------------------------")
	t.PrintHttpTrace()
	fmt.Println("-------------------------------------------------------------------------------------------")
}

// TimeTrace will be used for report httpstat
type TimeTrace struct {
	DNSLookup       time.Duration
	TCPConnect      time.Duration
	TLSHandshake    time.Duration
	ServerProcess   time.Duration
	ContentTransfer time.Duration
	NameLookup      time.Duration
	ConnectTotal    time.Duration
	PreTransfer     time.Duration
	StartTransfer   time.Duration
	TotalTime       time.Duration
}

// PrintHttpTrace prints TimeTrace struct in httpstat style
// This function's codes are from https://github.com/reoim/httpstat-1
// They are slightly different from the original codes. But I used the same template and print format.
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
		fmta(t.DNSLookup),
		fmta(t.TCPConnect),
		fmta(t.TLSHandshake),
		fmta(t.ServerProcess),
		fmta(t.ContentTransfer),
		fmtb(t.DNSLookup),
		fmtb(t.ConnectTotal),
		fmtb(t.PreTransfer),
		fmtb(t.StartTransfer),
		fmtb(t.TotalTime),
	)
}

func printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(color.Output, format, a...)
}

func grayscale(code color.Attribute) func(string, ...interface{}) string {
	return color.New(code + 232).SprintfFunc()
}
