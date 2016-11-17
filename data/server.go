// Copyright 2015 The Loadcat Authors. All rights reserved.

package data

import (
	"errors"

	"github.com/boltdb/bolt"
	"github.com/bradfitz/slice"
	"gopkg.in/mgo.v2/bson"
)

type Server struct {
	Id         bson.ObjectId
	BalancerId bson.ObjectId
	Label      string
	Settings   ServerSettings
}

type ServerSettings struct {
	Address      string
	Path         string
	Header       string
	Page         string
	Weight       int
	Availability Availability
}

func ListServersByBalancer(bal *Balancer) ([]Server, error) {
	srvs := []Server{}
	err := DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("servers"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			srv := Server{}
			err := bson.Unmarshal(v, &srv)
			if err != nil {
				return err
			}
			if srv.BalancerId.Hex() != bal.Id.Hex() {
				continue
			}
			srvs = append(srvs, srv)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	slice.Sort(srvs[:], func(i, j int) bool {
		return srvs[i].Label < srvs[j].Label
	})
	return srvs, nil
}

func GetServer(id bson.ObjectId) (*Server, error) {
	srv := &Server{}
	err := DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("servers"))
		v := b.Get([]byte(id.Hex()))
		if v == nil {
			srv = nil
			return nil
		}
		err := bson.Unmarshal(v, srv)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return srv, nil
}

func (s *Server) Balancer() (*Balancer, error) {
	return GetBalancer(s.BalancerId)
}

func (s *Server) Put() error {
	if !s.Id.Valid() {
		s.Id = bson.NewObjectId()
	}
	return DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("servers"))
		p, err := bson.Marshal(s)
		if err != nil {
			return err
		}
		return b.Put([]byte(s.Id.Hex()), p)
	})
}

func (s *Server) Delete() error {
	if s.Id.Valid() {
		return DB.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("servers"))
			return b.Delete([]byte(s.Id.Hex()))
		})
	}
	err := errors.New("wrong server id")
	return err
}
