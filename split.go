// Package split is a CoreDNS plugin that prints "example" to stdout on every packet received.
//
// It serves as an example CoreDNS plugin with numerous code comments.
package split

import (
	"context"
	"net"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

// Define log to be a logger with the plugin name in it. This way we can just use log.Info and
// friends to log.
var log = clog.NewWithPlugin("split")

// Split is an example plugin to show how to write a plugin.
type Split struct {
	Next plugin.Handler

	Rule []rule
}

type rule struct {
	zones    []string
	networks []network
}

type network struct {
	record  *net.IPNet
	allowed []*net.IPNet
}

// ServeDNS implements the plugin.Handler interface. This method gets called when example is used
// in a Server.
func (s Split) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	// This function could be simpler. I.e. just fmt.Println("example") here, but we want to show
	// a slightly more complex example as to make this more interesting.
	// Here we wrap the dns.ResponseWriter in a new ResponseWriter and call the next plugin, when the
	// answer comes back, it will print "example".

	// Debug log that we've seen the query. This will only be shown when the debug plugin is loaded.
	log.Debug("Received response")

	// Wrap.
	pw := s.NewResponsePrinter(w, r)

	// Export metric with the server label set to the current server handling the request.
	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()

	// Call next plugin (if any).
	return plugin.NextOrFailure(s.Name(), s.Next, ctx, pw, r)
}

// Name implements the Handler interface.
func (s Split) Name() string { return "split" }

// ResponsePrinter wrap a dns.ResponseWriter and will write example to standard output when WriteMsg is called.
type ResponsePrinter struct {
	dns.ResponseWriter
	state request.Request
	r     *dns.Msg
	src   net.IP
	rules []rule
}

// NewResponsePrinter returns ResponseWriter.
func (s Split) NewResponsePrinter(w dns.ResponseWriter, r *dns.Msg) *ResponsePrinter {
	state := request.Request{W: w, Req: r}
	ip := net.ParseIP(state.IP())
	return &ResponsePrinter{ResponseWriter: w, r: r, src: ip, rules: s.Rule, state: state}
}

// WriteMsg calls the underlying ResponseWriter's WriteMsg method and prints "example" to standard output.
func (r *ResponsePrinter) WriteMsg(res *dns.Msg) error {
	var rule rule
	for _, v := range r.rules {
		zone := plugin.Zones(v.zones).Matches(r.state.Name())
		if zone == "" {
			continue
		}
		rule = v
		break
	}
	var answers []dns.RR
	var netAnswers []dns.RR
	for _, v := range res.Answer {
		rec, ok := v.(*dns.A)
		if !ok {
			answers = append(answers, v)
			continue
		}
		var net *network
		for _, vv := range rule.networks {
			if vv.record.Contains(rec.A) {
				net = &vv
				break
			}
		}
		if net == nil {
			answers = append(answers, v)
			continue
		}
		allowed := false
		for _, vv := range net.allowed {
			if vv.Contains(r.src) {
				allowed = true
				break
			}
		}
		if allowed {
			answers = append(answers, v)
			netAnswers = append(netAnswers, v)
			continue
		}
		log.Infof("request source %s: %s: filtering %s", r.src.String(), rec.Hdr.Name, rec.A)
	}
	if len(netAnswers) != 0 {
		res.Answer = netAnswers
	} else {
		res.Answer = answers
	}
	return r.ResponseWriter.WriteMsg(res)
}
