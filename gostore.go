package gostore

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

// StoreOptions store init options
type StoreOptions struct {
	RedisHost string
}

// Store save persistent status
type Store struct {
	Namespace string
	Pool      *redis.Pool
}

// Init init store
func (s *Store) Init(options *StoreOptions) {
	s.Pool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			client, err := redis.Dial("tcp", options.RedisHost)
			if err != nil {
				log.Fatalf("redis connecting error: %s", err.Error())
			}
			return client, err
		},
	}
}

// Set set values
func (s *Store) Set(key string, value interface{}) error {
	v := encode(value)
	conn := s.Pool.Get()
	_, err := conn.Do("SET", s.Namespace+":"+key, v)
	return err
}

// Get get values
func (s *Store) Get(key string, result interface{}) (bool, error) {
	conn := s.Pool.Get()
	resp, err := conn.Do("GET", s.Namespace+":"+key)
	if err != nil {
		return false, err
	}
	if resp == nil {
		return false, nil
	}
	decode(resp.([]byte), result)
	return true, nil
}

func encode(value interface{}) string {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(value)
	if err != nil {
		panic(err)
	}

	return string(buf.Bytes())
}

func decode(value []byte, result interface{}) {
	buf := bytes.NewBuffer(value)
	enc := gob.NewDecoder(buf)
	err := enc.Decode(result)
	if err != nil {
		panic(err)
	}
}
