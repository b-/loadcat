// Copyright 2015 The Loadcat Authors. All rights reserved.

package data

import (
	"crypto/sha1" // we will need these when we
	"crypto/x509"
	"encoding/pem"

	"github.com/boltdb/bolt"
	"gopkg.in/mgo.v2/bson"
)

// Balancer - a host that we're going to listen for and load-balance
type Balancer struct {
	Id       bson.ObjectId
	Label    string
	Settings BalancerSettings
}

// BalancerSettings - the settings for a Balancer
type BalancerSettings struct {
	Hostname   string
	Port       int
	Protocol   Protocol
	Algorithm  Algorithm
	SSLOptions SSLOptions
}

// SSLOptions - the SSL options for a Balancer.
// If BalancerSettings.Protocol = HTTP this is ignored
type SSLOptions struct {
	CipherSuite CipherSuite
	Certificate []byte
	PrivateKey  []byte
	LetsEncrypt bool // This determines whether we manage the SSL certificate ourselves
	DNSNames    []string
	Fingerprint []byte
}

// ListBalancers - return a list of Balancers
func ListBalancers() ([]Balancer, error) {
	bals := []Balancer{}
	err := DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("balancers"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			bal := Balancer{}
			err := bson.Unmarshal(v, &bal)
			if err != nil {
				return err
			}
			bals = append(bals, bal)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return bals, nil
}

// GetBalancer - get Balancer by id
func GetBalancer(id bson.ObjectId) (*Balancer, error) {
	bal := &Balancer{}
	err := DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("balancers"))
		v := b.Get([]byte(id.Hex()))
		if v == nil {
			bal = nil
			return nil
		}
		err := bson.Unmarshal(v, bal)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return bal, nil
}

// Servers - return a list of servers for a given Balancer
func (l *Balancer) Servers() ([]Server, error) {
	return ListServersByBalancer(l)
}

// Put - save a new balancer
func (l *Balancer) Put() error {
	if !l.Id.Valid() {
		l.Id = bson.NewObjectId()
	}
	if l.Label == "" {
		l.Label = "Unlabelled"
	}
	if l.Settings.Protocol == "https" {
		if !l.Settings.SSLOptions.LetsEncrypt {
			buf := []byte{}
			raw := l.Settings.SSLOptions.Certificate
			for {
				p, rest := pem.Decode(raw)
				raw = rest
				if p == nil {
					break
				}
				buf = append(buf, p.Bytes...)
			}
			certs, err := x509.ParseCertificates(buf)
			if err != nil {
				return err
			}
			l.Settings.SSLOptions.DNSNames = certs[0].DNSNames
			sum := sha1.Sum(certs[0].Raw)
			l.Settings.SSLOptions.Fingerprint = sum[:]
		}
	} else {
		l.Settings.SSLOptions.CipherSuite = ""
		l.Settings.SSLOptions.Certificate = nil
		l.Settings.SSLOptions.PrivateKey = nil
		l.Settings.SSLOptions.DNSNames = nil
		l.Settings.SSLOptions.Fingerprint = nil
	}
	return DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("balancers"))
		p, err := bson.Marshal(l)
		if err != nil {
			return err
		}
		return b.Put([]byte(l.Id.Hex()), p)
	})
}
