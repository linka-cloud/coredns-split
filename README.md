# split

## Name

*split* - Filter DNS Server response Records based on network definitions and request source IP.

## Description

The split plugin allows filtering DNS Server response Records based on network definitions. That way
you do not need to run multiple DNS servers to handle split DNS.

## Compilation

This package will always be compiled as part of CoreDNS and not in a standalone way. It will require you to use `go get` or as a dependency on [plugin.cfg](https://github.com/coredns/coredns/blob/master/plugin.cfg).

The [manual](https://coredns.io/manual/toc/#what-is-coredns) will have more information about how to configure and extend the server with external plugins.

A simple way to consume this plugin, is by adding the following on [plugin.cfg](https://github.com/coredns/coredns/blob/master/plugin.cfg), and recompile it as [detailed on coredns.io](https://coredns.io/2017/07/25/compile-time-enabling-or-disabling-plugins/#build-with-compile-time-configuration-file).

~~~
split:go.linka.cloud/coredns-split
~~~

Put this lower in the plugin list, so that *split* is executed after any of the other plugins.

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

In this configuration, we forward all queries to 9.9.9.9 and print "example" whenever we receive
a query.

~~~ corefile
. {
  forward . 9.9.9.9
  example
}
~~~

Or without any external connectivity:

~~~ corefile
. {
  whoami
  example
}
~~~

## Also See

See the [manual](https://coredns.io/manual).
