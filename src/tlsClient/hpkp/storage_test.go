package hpkp

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

var createdAt = time.Now().Unix()

func TestMemStorage_Lookup(t *testing.T) {
	m := NewMemStorage()
	m.Add("example.org", &Header{
		IncludeSubDomains: false,
		Permanent:         false,
		Created:           createdAt,
		MaxAge:            100,
	})
	m.Add("a.example.org", &Header{
		IncludeSubDomains: true,
		Permanent:         false,
		Created:           createdAt,
		MaxAge:            100,
	})
	m.Add("a.example.com", &Header{
		IncludeSubDomains: true,
		Permanent:         false,
		Created:           createdAt,
		MaxAge:            100,
	})
	m.Add("b.example.com", &Header{
		IncludeSubDomains: false,
		Permanent:         false,
		Created:           createdAt,
		MaxAge:            100,
	})

	done := make(chan bool)

	var orgErr error
	go func() {
		orgErr = orgTest(m, t)
		// try to make a data race
		m.Add("example.org", &Header{
			IncludeSubDomains: false,
			Permanent:         false,
			Created:           createdAt,
			MaxAge:            100,
		})
		done <- true
	}()

	var comErr error
	go func() {
		comErr = comTest(m, t)
		// try to make a data race
		m.Add("a.example.com", &Header{
			IncludeSubDomains: true,
			Permanent:         false,
			Created:           createdAt,
			MaxAge:            100,
		})
		done <- true
	}()

	// wait for tests to finish
	<-done
	<-done

	if orgErr != nil {
		t.Fatal(orgErr)
	}

	if comErr != nil {
		t.Fatal(comErr)
	}
}

func orgTest(m Storage, t *testing.T) error {
	tests := []struct {
		name     string
		host     string
		expected *Header
	}{
		{
			name: "root match org",
			host: "example.org",
			expected: &Header{
				IncludeSubDomains: false,
				Permanent:         false,
				Created:           createdAt,
				MaxAge:            100,
			},
		},
		{
			name: "subdomain match org",
			host: "a.example.org",
			expected: &Header{
				IncludeSubDomains: true,
				Permanent:         false,
				Created:           createdAt,
				MaxAge:            100,
			},
		},
		{
			name:     "subdomain miss-match org",
			host:     "b.example.org",
			expected: nil,
		},
	}

	for _, test := range tests {
		out := m.Lookup(test.host)
		if !reflect.DeepEqual(out, test.expected) {
			t.Logf("host: %s", test.host)
			t.Logf("want:%v", test.expected)
			t.Logf("got:%v", out)
			return fmt.Errorf("test case failed: %s", test.name)
		}
	}
	return nil
}

func comTest(m Storage, t *testing.T) error {
	tests := []struct {
		name     string
		host     string
		expected *Header
	}{
		{
			name: "subdomain enabled",
			host: "z.a.example.com",
			expected: &Header{
				IncludeSubDomains: true,
				Permanent:         false,
				Created:           createdAt,
				MaxAge:            100,
			},
		},
		{
			name: "sub-subdomain",
			host: "z.y.a.example.com",
			expected: &Header{
				IncludeSubDomains: true,
				Permanent:         false,
				Created:           createdAt,
				MaxAge:            100,
			},
		},
		{
			name:     "subdomain disabled",
			host:     "z.b.example.com",
			expected: nil,
		},
		{
			name: "exact match",
			host: "b.example.com",
			expected: &Header{
				IncludeSubDomains: false,
				Permanent:         false,
				Created:           createdAt,
				MaxAge:            100,
			},
		},
		{
			name:     "complete missmatch",
			host:     "z.example.com",
			expected: nil,
		},
	}

	for _, test := range tests {
		out := m.Lookup(test.host)
		if !reflect.DeepEqual(out, test.expected) {
			t.Logf("host: %s", test.host)
			t.Logf("want:%v", test.expected)
			t.Logf("got:%v", out)
			return fmt.Errorf("test case failed: %s", test.name)
		}
	}
	return nil
}

func TestMemStorage_Add(t *testing.T) {
	m := &MemStorage{}

	// permanent
	permanentDomain := Header{
		IncludeSubDomains: false,
		Permanent:         true,
		Created:           time.Now().Unix(),
		MaxAge:            0,
	}

	m.Add("example.org", &permanentDomain)

	expected := map[string]Header{
		"example.org": permanentDomain,
	}

	if !reflect.DeepEqual(m.domains, expected) {
		t.Logf("want:%v", expected)
		t.Logf("got:%v", m.domains)
		t.Fatal("Add failed after permanent")
	}

	// normal
	normalDomain := Header{
		IncludeSubDomains: false,
		Permanent:         false,
		Created:           time.Now().Unix(),
		MaxAge:            100,
	}

	m.Add("a.example.org", &normalDomain)

	expected = map[string]Header{
		"example.org":   permanentDomain,
		"a.example.org": normalDomain,
	}

	if !reflect.DeepEqual(m.domains, expected) {
		t.Logf("want:%v", expected)
		t.Logf("got:%v", m.domains)
		t.Fatal("Add failed after adding normal")
	}

	// remove normal
	removeNormalDomain := Header{
		IncludeSubDomains: false,
		Permanent:         false,
		Created:           time.Now().Unix(),
		MaxAge:            0,
	}

	m.Add("a.example.org", &removeNormalDomain)

	expected = map[string]Header{
		"example.org": permanentDomain,
	}

	if !reflect.DeepEqual(m.domains, expected) {
		t.Logf("want:%v", expected)
		t.Logf("got:%v", m.domains)
		t.Fatal("Add failed after removing normal")
	}

	// attempt to remove the permanent
	removePermanetDomain := Header{
		IncludeSubDomains: false,
		Permanent:         false,
		Created:           time.Now().Unix(),
		MaxAge:            0,
	}

	m.Add("example.org", &removePermanetDomain)

	expected = map[string]Header{
		"example.org": permanentDomain,
	}

	if !reflect.DeepEqual(m.domains, expected) {
		t.Logf("want:%v", expected)
		t.Logf("got:%v", m.domains)
		t.Fatal("Add failed after attempting to remove the permanent domain")
	}
}
