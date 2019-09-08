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

type PingDto struct {
	Region  string
	Name    string
	Address string
	Latency time.Duration
}

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

/* Below codes are from https://github.com/reoim/httpstat-1
They are slightly different from the original codes. But I used the same template and print format. */
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
