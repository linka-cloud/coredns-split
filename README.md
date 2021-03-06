# split

## Name

*split* - Filter DNS Server response Records based on network definitions and request source IP.

## Description

The split plugin allows filtering DNS Server responses Records based on network definitions. That way
you do not need to run multiple DNS servers to handle split DNS.

If there are multiple A Records in the response, only the records matching the defined network will be returned
to a matching querier, and the records not matching the network to the other sources.

⚠️ This plugin is not much about security, it is designed only to give a better answer to the incoming source IP,
if you need to apply security filtering rules, please consider using the [**coredns** *acl*](https://coredns.io/plugins/acl/) plugin. 

## Compilation

This package will always be compiled as part of CoreDNS and not in a standalone way. It will require you to use `go get` or as a dependency on [plugin.cfg](https://github.com/coredns/coredns/blob/master/plugin.cfg).

The [manual](https://coredns.io/manual/toc/#what-is-coredns) will have more information about how to configure and extend the server with external plugins.

A simple way to consume this plugin, is by adding the following on [plugin.cfg](https://github.com/coredns/coredns/blob/master/plugin.cfg), and recompile it as [detailed on coredns.io](https://coredns.io/2017/07/25/compile-time-enabling-or-disabling-plugins/#build-with-compile-time-configuration-file).

~~~
split:go.linka.cloud/coredns-split
~~~

Put this higher in the plugin list, so that *split* is before after any of the other plugins.

After this you can compile coredns by:

``` sh
go generate
go build
```

Or you can instead use make:

``` sh
make
```

## Syntax

~~~ txt
split
# TODO: docs
~~~

## Metrics

If monitoring is enabled (via the *prometheus* directive) the following metric is exported:

* `coredns_example_request_count_total{server}` - query count to the *example* plugin.

The `server` label indicated which server handled the request, see the *metrics* plugin for details.

## Ready

This plugin reports readiness to the ready plugin. It will be immediately ready.

## Examples

In this configuration, we forward all queries to 9.9.9.9 and filter out A records pointing to an IP address
in the 10.10.10.0/24 network except for queries coming from the 192.168.0.0/24 and 192.168.1.0/24 networks.

~~~ corefile
. {
  example {
    10.10.10.0/24 {
        net 192.168.0.0/24 192.168.1.0/24
    }
  }
  forward . 9.9.9.9
}
~~~

## Also See

See the [manual](https://coredns.io/manual).
