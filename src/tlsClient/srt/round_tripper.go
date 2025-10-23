package srt

import (
	"fmt"
	"github.com/qtgolang/SunnyNet/src/http"
	tlsclient "github.com/qtgolang/SunnyNet/src/tlsClient/tlsClient"
)

type SpoofedRoundTripper struct {
	Client tlsclient.HttpClient
}

func NewSpoofedRoundTripper(httpClientOption ...tlsclient.HttpClientOption) (*SpoofedRoundTripper, error) {
	c, err := tlsclient.NewHttpClient(tlsclient.NewNoopLogger(), httpClientOption...)
	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}
	c.SetProxy("socks5://127.0.0.1:2024")
	return &SpoofedRoundTripper{
		Client: c,
	}, nil
}

func (s SpoofedRoundTripper) RoundTrip(hReq *http.Request) (*http.Response, error) {

	fResp, err := s.Client.Do(hReq)
	if err != nil {
		return nil, fmt.Errorf("error fetching response: %w", err)
	}
	return &http.Response{
		Status:           fResp.Status,
		StatusCode:       fResp.StatusCode,
		Proto:            fResp.Proto,
		ProtoMajor:       fResp.ProtoMajor,
		ProtoMinor:       fResp.ProtoMinor,
		Header:           http.Header(fResp.Header),
		Body:             fResp.Body,
		ContentLength:    fResp.ContentLength,
		TransferEncoding: fResp.TransferEncoding,
		Close:            fResp.Close,
		Uncompressed:     fResp.Uncompressed,
		Trailer:          http.Header(fResp.Trailer),
		Request:          hReq,
		TLS:              nil,
	}, nil
}
