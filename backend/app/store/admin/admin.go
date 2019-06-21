// Package admin defines and implements store for admin-level data like secret key, list of admins and so on
package admin

import (
	"errors"

	log "github.com/go-pkgz/lgr"
)

// Store defines interface returning admins info for given site
type Store interface {
	Key() (key string, err error)
	Admins(siteID string) (ids []string, err error)
	Email(siteID string) (email string, err error)
}

// StaticStore implements keys.Store with a single set of admins and email for all sites
type StaticStore struct {
	admins []string
	email  string
	key    string
}

// NewStaticStore makes StaticStore instance with given key
func NewStaticStore(key string, admins []string, email string) *StaticStore {
	log.Printf("[DEBUG] admin users %+v, email %s", admins, email)
	return &StaticStore{key: key, admins: admins, email: email}
}

// NewStaticKeyStore is a shortcut for making StaticStore for key consumers only
func NewStaticKeyStore(key string) *StaticStore {
	return &StaticStore{key: key, admins: []string{}, email: ""}
}

// Key returns static key, same for all sites
func (s *StaticStore) Key() (key string, err error) {
	if s.key == "" {
		return "", errors.New("empty key for static key store")
	}
	return s.key, nil
}

// Admins returns static list of admin's ids, the same for all sites
func (s *StaticStore) Admins(string) (ids []string, err error) {
	return s.admins, nil
}

// Email gets static email address
func (s *StaticStore) Email(string) (email string, err error) {
	return s.email, nil
}
