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

## Ready

This plugin reports readiness to the ready plugin. It will be immediately ready.

## Examples

In this configuration, we forward all queries to 10.10.10.1 and to 9.9.9.9 if 10.10.10.1 did not respond.

**If only used with the forward plugin, the private dns server must be configured as the first forwarded server in the list. The policy must be configured as sequential, so that the first server is always tried first and the second only if the first do not return any answer.**

We filter out A / CNAME / SRV / PTR records pointing to an IP address in the 10.10.10.0/24 network except for queries coming from the 192.168.0.0/24 and 192.168.1.0/24 networks.
If the allowed networks are not defined, the plugin will allow the requests from the same network, e.g. 10.10.10.0/24.

If the record exists both as public and private, the private record will be filtered, resulting with no records at all.
So you can provide a fallback server that will be used to get the public record.

~~~ corefile
. {
  split {
    net 10.10.10.0/24 allow 192.168.0.0/24 192.168.1.0/24
    net 10.1.1.0/24 10.1.2.0/24 # implicitely: allow 10.1.1.0/24 10.1.2.0/24
    fallback 8.8.8.8
  }
  # we could also use any records source
  # e.g.: file example.org
  forward . 10.10.10.1 9.9.9.9 {
    policy sequential
  }
}
~~~

## Also See

See the [manual](https://coredns.io/manual).
