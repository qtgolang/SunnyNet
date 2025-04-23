package hpkp

import (
	"strings"
	"sync"
)

// MemStorage is threadsafe hpkp host storage backed by an in-memory map
type MemStorage struct {
	domains map[string]Header
	mutex   sync.Mutex
}

// NewMemStorage initializes hpkp in-memory datastructure
func NewMemStorage() *MemStorage {
	m := &MemStorage{}
	m.domains = make(map[string]Header)
	return m
}

// Lookup returns the corresponding hpkp header information for a given host
func (s *MemStorage) Lookup(host string) *Header {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	d, ok := s.domains[host]
	if ok {
		return copy(d)
	}

	// is h a subdomain of an hpkp domain, walk the domain to see if it is a sub
	// sub ... sub domain of a domain that has the `includeSubDomains` rule
	l := len(host)
	for l > 0 {
		i := strings.Index(host, ".")
		if i > 0 {
			host = host[i+1:]
			d, ok := s.domains[host]
			if ok {
				if d.IncludeSubDomains {
					return copy(d)
				}
			}
			l = len(host)
		} else {
			break
		}
	}

	return nil
}

func copy(h Header) *Header {
	d := h
	return &d
}

// Add a domain to hpkp storage
func (s *MemStorage) Add(host string, d *Header) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.domains == nil {
		s.domains = make(map[string]Header)
	}

	if d.MaxAge == 0 && !d.Permanent {
		check, ok := s.domains[host]
		if ok {
			if !check.Permanent {
				delete(s.domains, host)
			}
		}
	} else {
		s.domains[host] = *d
	}
}
