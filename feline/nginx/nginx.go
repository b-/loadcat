// Copyright 2015 The Loadcat Authors. All rights reserved.

package nginx

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/coreos/go-systemd/dbus"

	"github.com/radkoa/loadcat/cfg"
	"github.com/radkoa/loadcat/data"
	"github.com/radkoa/loadcat/feline"
)

var TplNginxConf = template.Must(template.New("").Parse(`
server {
	listen 80 default_server;
	listen [::]:80 default_server;

	root /var/www/html;
	index index.html index.htm index.nginx-debian.html;

	location / {
		# First attempt to serve request as file, then
		# as directory, then fall back to displaying a 404.
		try_files $uri $uri/ =404;
  }

	server_name _;
#	server_name  {{.Balancer.Settings.Hostname}};

	{{if eq .Balancer.Settings.Protocol "https"}}
		ssl                  on;
		ssl_certificate      {{.Dir}}/server.crt;
		ssl_certificate_key  {{.Dir}}/server.key;
	{{end}}

{{range $srv := .Balancer.Servers}}
	#{{$srv.Label}}
	location /{{$srv.Settings.Path}} {
		proxy_set_header  Host $host;
		proxy_set_header  X-Real-IP $remote_addr;
		proxy_set_header  X-Forwarded-For $proxy_add_x_forwarded_for;
		proxy_set_header  X-Forwarded-Proto $scheme;

		{{if $srv.Settings.Header}}proxy_set_header {{$srv.Settings.Header}}{{end}};
		proxy_pass  {{$srv.Settings.Address}};

		proxy_http_version  1.1;
		proxy_set_header  Upgrade $http_upgrade;
		proxy_set_header  Connection 'upgrade';
	}

{{end}}
}
`))

type Nginx struct {
	sync.Mutex

	Systemd *dbus.Conn
}

func (n Nginx) Generate(dir string, bal *data.Balancer) error {
	n.Lock()
	defer n.Unlock()

	f, err := os.Create(filepath.Join(dir, "nginx.conf"))
	if err != nil {
		return err
	}
	err = TplNginxConf.Execute(f, struct {
		Dir      string
		Balancer *data.Balancer
	}{
		Dir:      dir,
		Balancer: bal,
	})
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}

	if bal.Settings.Protocol == "https" {
		err = ioutil.WriteFile(filepath.Join(dir, "server.crt"), bal.Settings.SSLOptions.Certificate, 0666)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(filepath.Join(dir, "server.key"), bal.Settings.SSLOptions.PrivateKey, 0666)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n Nginx) Reload() error {
	n.Lock()
	defer n.Unlock()

	switch cfg.Current.Nginx.Mode {
	case "systemd":
		if n.Systemd == nil {
			c, err := dbus.NewSystemdConnection()
			if err != nil {
				return err
			}
			n.Systemd = c
		}

		ch := make(chan string)
		_, err := n.Systemd.ReloadUnit(cfg.Current.Nginx.Systemd.Service, "replace", ch)
		if err != nil {
			return err
		}
		<-ch

		return nil

	default:
		return errors.New("unknown Nginx mode")
	}

	panic("unreachable")
}

func init() {
	feline.Register("nginx", Nginx{})
}
