package hpkp_test

import (
	"fmt"
	"log"
	"net/http"

	"github.com/tam7t/hpkp"
)

func Example() {
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
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.StatusCode)
	// Output: 200
}
