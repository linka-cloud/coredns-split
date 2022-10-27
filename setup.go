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
	var s Split
	for c.Next() {
		r := Rule{
			Zones: plugin.OriginsFromArgsOrServerBlock(c.RemainingArgs(), c.ServerBlockKeys),
		}
		log.Debug("zones", r.Zones)
		for c.NextBlock() {
			prop := c.Val()
			args := c.RemainingArgs()
			switch prop {
			case "net":
				log.Debug("net", args)
				if len(args) == 0 {
					return c.Errf("net: expected at least 1 argument, got 0")
				}
				var nets []*net.IPNet
				var allow bool
				var allowNets []*net.IPNet
				for _, v := range args {
					switch v {
					case "allow":
						allow = true
					default:
						_, ipnet, err := net.ParseCIDR(v)
						if err != nil {
							return err
						}
						if allow {
							allowNets = append(allowNets, ipnet)
						} else {
							nets = append(nets, ipnet)
						}
					}
				}
				if len(allowNets) == 0 {
					allowNets = nets[:]
				}
				for _, v := range nets {
					r.Networks = append(r.Networks, Network{
						RecordNetwork: v,
						Allowed:       allowNets,
					})
				}
			case "fallback":
				log.Debug("fallback", args)
				if r.Fallback != nil {
					return c.Errf("fallback already set")
				}
				if len(args) != 1 {
					return c.Errf("fallback: expected 1 argument, got %d", len(args))
				}
				ip := net.ParseIP(args[0])
				if ip == nil {
					return c.Errf("fallback: invalid IP %s", args[0])
				}
				r.Fallback = ip
			default:
				return c.Errf("unknown property '%s'", prop)
			}
		}
		s.Rules = append(s.Rules, r)
	}

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		s.Next = next
		return s
	})

	// All OK, return a nil error.
	return nil
}
