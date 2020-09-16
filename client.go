package eml

import (
	"time"
)

type bearerToken struct {
	Value   string
	Expires time.Time
}

func LazyStore(rs string, s *Settings) Store {
	return &emlStore{
		_restSecret: rs,
		_env:        s,
	}
}

func (b *bearerToken) ShouldRefresh() bool {
	return b == nil || b.Expires.Before(time.Now().Add(time.Duration(10)*time.Minute))
}

func (b *bearerToken) IsValid() bool {
	return b != nil && b.Expires.After(time.Now().Add(time.Duration(1)*time.Minute))
}
