# Loadcat

Loadcat is an Nginx configurator that allows you to use Nginx as a load balancer. The project is inspired by the various Nginx load balancing tutorial articles available online and also the existence of Linode's load balancer service [NodeBalancers](https://www.linode.com/nodebalancers). So far the tool covers some of HTTP and HTTPS load balancing features, such as SSL termination, adding servers on the fly, marking them as unavailable or backup as necessary, and setting their weights to distribute load fairly.

## Installation

### Arch Linux

Install Loadcat using a pre-built .pkg file:

~~~
$ wget https://github.com/radkoa/loadcat/releases/download/v0.1-alpha.1/loadcat-0.1_alpha.1-1-x86_64.pkg.tar.xz
# pacman -U loadcat-0.1_alpha.1-1-x86_64.pkg.tar.xz
~~~

Or, from AUR using Yaourt:

~~~
$ yaourt loadcat
~~~

Or, manually:

~~~
$ git clone https://aur.archlinux.org/loadcat.git
$ cd loadcat
$ makepkg
# pacman -U loadcat-0.1_alpha.1-1-x86_64.pkg.tar.xz
~~~

### From Source

Install Loadcat using the go get command:

```
$ go get github.com/radkoa/loadcat/cmd/loadcatd
```

## Usage

If you installed Loadcat using the distribution specific package, you can start it as a service using systemctl:

```
# systemctl start loadcat.service
```

If you installed Loadcat from source, you can launch it with:

```
$ cd $GOPATH/src/github.com/radkoa/loadcat
# $GOPATH/bin/loadcatd
```

Loadcat parses a TOML encoded configuration file. In case one is not found, Loadcat will create one with same sane defaults. The location of the configuration file can be specified with the _-config_ flag.

Loadcat works by generating Nginx configuration files under a particular directory (a directory named "out" under Loadcat's working directory, as set in loadcat.conf). Nginx must be configured to load configuration files from this directory. For example on Arch Linux, when installed from AUR, Loadcat uses /var/lib/loadcat/out as the directory where generated Nginx configuration files are stored. You must include the following line inside the `http {}` block of /etc/nginx/nginx.conf to load configuration files generated by Loadcat:

```
include  /var/lib/loadcat/out/*/nginx.conf
```

Once Loadcat is running, you can navigate to http://{host}:26590 on your web browser, where _{host}_ is the domain name or IP address of the machine where Loadcat is running (for example http://localhost:26590 when running locally). To save a thousand words, here is a (kind of big) picture:

![4 steps primer](http://i.imgur.com/7l6zN5n.png)

Make sure that Nginx is running as a systemd service and is configured to load generated configuration files from the appropriate directory.

## Caution

As this is a very young project, and pretty experimental, you may encounter bugs and issues here and there. It would be really appreciated if you could open an issue outlining details of the bug or issue, or any feature that you would like to request. Any contribution or criticism (constructive or destructive) is really appreciated.

A lot of Nginx load balancing features is still not covered by this tool and at this moment that makes this rather limited in context of practical applications. However, solving this problem is just a matter of time. Although Nginx Plus specific features may have to wait for a while - at least until I get my hands on an instance or someone else with access to Nginx Plus starts contributing.

## Documentation

- [Reference](http://godoc.org/github.com/radkoa/loadcat)

## Resources

- [DigitalOcean's Tutorial on Using Nginx as a Load Balancer](https://www.digitalocean.com/community/tutorials/how-to-set-up-nginx-load-balancing)
- [Nginx's HTTP Load Balancing Documentation](http://nginx.org/en/docs/http/load_balancing.html)

## Contributing

Contributions are welcome.

## License

Loadcat is available under the [BSD (3-Clause) License](http://opensource.org/licenses/BSD-3-Clause).
