// Code generated by 'yaegi extract net/http/cgi'. DO NOT EDIT.

//go:build go1.20
// +build go1.20

package base

import (
	"github.com/qtgolang/SunnyNet/src/http/cgi"
	"reflect"
)

func init() {
	Symbols["net/http/cgi/cgi"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Request":        reflect.ValueOf(cgi.Request),
		"RequestFromMap": reflect.ValueOf(cgi.RequestFromMap),
		"Serve":          reflect.ValueOf(cgi.Serve),

		// type definitions
		"Handler": reflect.ValueOf((*cgi.Handler)(nil)),
	}
}