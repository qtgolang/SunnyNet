# hpkp
[![Go Report Card](https://goreportcard.com/badge/github.com/tam7t/hpkp?style=flat-square)](https://goreportcard.com/report/github.com/tam7t/hpkp) [![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/tam7t/hpkp) [![Build Status](http://img.shields.io/travis/tam7t/hpkp.svg?style=flat-square)](https://travis-ci.org/tam7t/hpkp)

Library for performing certificate pin validation for golang applications.

## Motivation

I couldn't find any Golang libraries that make key pinning any easier, so I decided to start my own library for writing HPKP aware clients. This library is aimed at providing:

1. HPKP related tools (generate pins, inspect servers)
1. A convenience functions for writing clients that support pin verification


## Examples

To inspect the HPKP headers from the server:

```
$ hpkp-headers https://github.com
{"Created":1465765483,"MaxAge":5184000,"IncludeSubDomains":true,"Permanent":false,"Sha256Pins":["WoiWRyIOVNa9ihaBciRSC7XHjliYS9VwUGOIud4PB18=","RRM1dGqnDFsCJXBTHky16vi1obOlCgFFn/yOhI/y+ho=","k2v657xBsOVe1PQRwOsHsw3bsGT2VzIqz5K+59sNQws=","K87oWBWM9UZfyddvDfoxL+8lpNyoUB2ptGtn0fv6G2Q=","IQBnNBEiFuhj+8x6X8XLgh01V9Ic5/V3IRQLNFFc7v4=","iie1VXtL7HzAMF+/PVPR9xzT80kQxdZeJ+zduCB3uj0=","LvRiGEjRqfzurezaWuj8Wie2gyHMrW5Q06LspMnox7A="]}
```

And generate pins from the certs a server presents:

```
$ hpkp-pins -server=github.com:443
pL1+qb9HTMRZJmuC/bB/ZI9d302BYrrqiVuRyW+DGrU=
RRM1dGqnDFsCJXBTHky16vi1obOlCgFFn/yOhI/y+ho=
```

Or generate a pin from a PEM-encoded certificate file:

```
$ hpkp-pins -file=cert.pem
AD4C8VGyUrvmReK+D/PYtH52cYJrG9o7VR+uOZIh1Q0=
pL1+qb9HTMRZJmuC/bB/ZI9d302BYrrqiVuRyW+DGrU=
```

And finally, how to use the `hpkp` package to verify pins as part of your application:

```
s := hpkp.NewMemStorage()

s.Add("github.com", &hpkp.Header{
    Permanent: true,
    Sha256Pins: []string{
        "WoiWRyIOVNa9ihaBciRSC7XHjliYS9VwUGOIud4PB18=",
        "RRM1dGqnDFsCJXBTHky16vi1obOlCgFFn/yOhI/y+ho=",
        "k2v657xBsOVe1PQRwOsHsw3bsGT2VzIqz5K+59sNQws=",
        "K87oWBWM9UZfyddvDfoxL+8lpNyoUB2ptGtn0fv6G2Q=",
        "IQBnNBEiFuhj+8x6X8XLgh01V9Ic5/V3IRQLNFFc7v4=",
        "iie1VXtL7HzAMF+/PVPR9xzT80kQxdZeJ+zduCB3uj0=",
        "LvRiGEjRqfzurezaWuj8Wie2gyHMrW5Q06LspMnox7A=",
    },
})

client := &http.Client{}
dialConf := &hpkp.DialerConfig{
	Storage:   s,
	PinOnly:   true,
	TLSConfig: nil,
	Reporter: func(p *hpkp.PinFailure, reportUri string) {
		// TODO: report on PIN failure
		fmt.Println(p)
	},
}

client.Transport = &http.Transport{
	DialTLS: dialConf.NewDialer(),
}
resp, err := client.Get("https://github.com")
```

## References

* https://tools.ietf.org/html/rfc7469
* https://developer.mozilla.org/en-US/docs/Web/Security/Public_Key_Pinning
