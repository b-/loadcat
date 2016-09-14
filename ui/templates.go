// Copyright 2015 The Loadcat Authors. All rights reserved.

package ui

import (
	"fmt"
	"html/template"
	"path/filepath"
	"github.com/radkoa/loadcat/cfg"
)

var (
	TplLayout = template.Must(template.New("layout.html").ParseFiles(filepath.Join(cfg.Current.Core.Dir, "ui/templates/layout.html")))

	fmt.Printf("cfgdir: %#v \n", cfg.Current.Core.Dir)

	TplBalancerList     = template.Must(template.Must(TplLayout.Clone()).ParseFiles(filepath.Join(cfg.Current.Core.Dir, "ui/templates/balancerList.html")))
	TplBalancerNewForm  = template.Must(template.Must(TplLayout.Clone()).ParseFiles(filepath.Join(cfg.Current.Core.Dir, "ui/templates/balancerNewForm.html")))
	TplBalancerView     = template.Must(template.Must(TplLayout.Clone()).ParseFiles(filepath.Join(cfg.Current.Core.Dir, "ui/templates/balancerView.html")))
	TplBalancerEditForm = template.Must(template.Must(TplLayout.Clone()).ParseFiles(filepath.Join(cfg.Current.Core.Dir, "ui/templates/balancerEditForm.html")))

	TplServerNewForm  = template.Must(template.Must(TplLayout.Clone()).ParseFiles(filepath.Join(cfg.Current.Core.Dir, "ui/templates/serverNewForm.html")))
	TplServerEditForm = template.Must(template.Must(TplLayout.Clone()).ParseFiles(filepath.Join(cfg.Current.Core.Dir, "ui/templates/serverEditForm.html")))
)
