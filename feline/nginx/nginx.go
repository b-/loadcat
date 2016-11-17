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

var (
	TplConf = template.Must(template.New("conf").Parse(`
{{template "server" .}}
`))
	TplNginxConf = template.Must(template.Must(TplConf.Clone()).ParseFiles(filepath.Join(cfg.Current.Core.Dir, "nginx-template.conf")))
)

//template.Must(template.New("").ParseFiles(filepath.Join(cfg.Current.Core.Dir, "nginx-template.conf")))
//template.Must(template.New("").Parse(``))

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
