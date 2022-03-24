package split

import (
	"net"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

// init registers this plugin.
func init() { plugin.Register("split", setup) }

// setup is the function that gets called when the config parser see the token "example". Setup is responsible
// for parsing any extra options the example plugin may have. The first token this function sees is "example".
func setup(c *caddy.Controller) error {
	s := Split{}
	for c.Next() {
		r := rule{}
		args := c.RemainingArgs()
		r.zones = plugin.OriginsFromArgsOrServerBlock(args, c.ServerBlockKeys)
		if c.NextBlock() {
			n := network{}
			_, ipnet, err := net.ParseCIDR(c.Val())
			if err != nil {
				return err
			}
			n.record = ipnet
			for c.NextBlock() {
				for c.NextLine() {
					a := c.Val()
					_ = a
				argsLoop:
					for _, v := range c.RemainingArgs() {
						_, ipnet, err := net.ParseCIDR(v)
						if err != nil {
							return err
						}
						for _, vv := range n.allowed {
							if vv.Contains(ipnet.IP) {
								continue argsLoop
							}
						}
						n.allowed = append(n.allowed, ipnet)
					}
				}
				r.networks = append(r.networks, n)
			}
			s.Rule = append(s.Rule, r)
		}
	}

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		s.Next = next
		return s
	})

	// All OK, return a nil error.
	return nil
}
