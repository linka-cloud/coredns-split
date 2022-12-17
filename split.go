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
	"github.com/coredns/coredns/plugin/pkg/upstream"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

// Define log to be a logger with the plugin name in it. This way we can just use log.Info and
// friends to log.
var log = clog.NewWithPlugin("split")

const noFallback = "split-no-fallback"

func isNoFallback(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	if v, ok := ctx.Value(noFallback).(bool); ok {
		return v
	}
	return false
}

// Split is an example plugin to show how to write a plugin.
type Split struct {
	Next plugin.Handler

	Rules []Rule
}

type Rule struct {
	Zones    []string
	Networks []Network
	Fallback net.IP
}

type Network struct {
	RecordNetwork *net.IPNet
	Allowed       []*net.IPNet
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
	pw := s.NewResponsePrinter(ctx, w, r)

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
	ctx   context.Context
	state request.Request
	r     *dns.Msg
	src   net.IP
	rules []Rule
}

// NewResponsePrinter returns ResponseWriter.
func (s Split) NewResponsePrinter(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) *ResponsePrinter {
	state := request.Request{W: w, Req: r}
	ip := net.ParseIP(state.IP())
	return &ResponsePrinter{ctx: ctx, ResponseWriter: w, r: r, src: ip, rules: s.Rules, state: state}
}

// WriteMsg calls the underlying ResponseWriter's WriteMsg method and prints "example" to standard output.
func (r *ResponsePrinter) WriteMsg(res *dns.Msg) error {
	filter := func(rec *dns.A) (rule Rule, allowed, match bool) {
		for _, v := range r.rules {
			zone := plugin.Zones(v.Zones).Matches(r.state.Name())
			if zone == "" {
				continue
			}
			rule = v
			break
		}
		var net *Network
		for _, vv := range rule.Networks {
			if vv.RecordNetwork.Contains(rec.A) {
				net = &vv
				break
			}
		}
		if net == nil {
			return rule, true, false
		}

		for _, vv := range net.Allowed {
			if vv.Contains(r.src) {
				return rule, true, true
			}
		}
		return rule, false, true
	}
	var (
		rule       Rule
		answers    []dns.RR
		netAnswers []dns.RR
	)

	for _, v := range res.Answer {
		switch rec := v.(type) {
		case *dns.A:
			var allowed, match bool
			rule, allowed, match = filter(rec)
			if !match {
				answers = append(answers, v)
				continue
			}
			if allowed {
				answers = append(answers, v)
				netAnswers = append(netAnswers, v)
				continue
			}
			log.Infof("request source %s: %s: filtering %s", r.src.String(), rec.Hdr.Name, rec.A)
		case *dns.CNAME:
			res, err := r.query(rec.Target)
			if err != nil {
				log.Errorf("error querying %s: %s", rec.Target, err)
				continue
			}
			if res == nil || len(res.Answer) == 0 {
				log.Debugf("no answers for %s", rec.Target)
				continue
			}
			answers = append(answers, v)
		case *dns.SRV:
			res, err := r.query(rec.Target)
			if err != nil {
				log.Errorf("error querying %s: %s", rec.Target, err)
				continue
			}
			if res == nil || len(res.Answer) == 0 {
				log.Debugf("no answers for %s", rec.Target)
				continue
			}
			answers = append(answers, v)
		case *dns.PTR:
			a, err := r.query(rec.Ptr)
			if err != nil {
				log.Errorf("error querying %s: %s", rec.Ptr, err)
				continue
			}
			if res == nil || len(a.Answer) == 0 {
				log.Debugf("no answer for %s", rec.Ptr)
				continue
			}
			answers = append(answers, v)
		default:
			return r.ResponseWriter.WriteMsg(res)
		}
	}
	if len(netAnswers) != 0 {
		res.Answer = netAnswers
	} else {
		res.Answer = answers
	}
	if len(res.Answer) != 0 || len(rule.Zones) == 0 {
		return r.ResponseWriter.WriteMsg(res)
	}
	if isNoFallback(r.ctx) {
		log.Debugf("no fallback requested for %s", r.state.Name())
		return r.ResponseWriter.WriteMsg(res)
	}
	if rule.Fallback == nil {
		log.Debugf("no fallback configured for zones %v", rule.Zones)
		return r.ResponseWriter.WriteMsg(res)
	}
	log.Debugf("request source %s: %s: using fallback %s", r.src.String(), r.state.Name(), rule.Fallback)
	c := new(dns.Client)
	req := r.state.Req.Copy()
	req.Id = dns.Id()
	in, _, err := c.Exchange(req, rule.Fallback.String()+":53")
	if err != nil {
		return err
	}
	res.Answer = append(res.Answer, in.Answer...)
	return r.ResponseWriter.WriteMsg(res)
}

func (r *ResponsePrinter) query(name string) (*dns.Msg, error) {
	log.Debugf("internally querying %s", name)
	ctx := context.WithValue(r.ctx, noFallback, true)
	res, err := upstream.New().Lookup(ctx, r.state, name, dns.TypeA)
	if err != nil {
		return nil, err
	}
	return res, nil
}
