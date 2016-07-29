package ptc

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestProvider(t *testing.T) {

	Convey("Setup PTC Provider", t, func() {

		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(200)
			rw.Write([]byte("pokemon"))
		}))

		u, err := url.Parse(server.URL)
		if err != nil {
			log.Fatal("failed parsing url", err)
		}

		pvd := NewProvider("admin", "password")
		pvd.http = &http.Client{Transport: RewriteTransport{URL: u}}

		So(pvd, ShouldNotBeNil)
		So(pvd.GetProviderString(), ShouldEqual, "ptc")

		Convey("Check login process", func() {
			ch := make(chan *HTTPResponses, 1)
			pvd.checkLoginProcess(ch)
			resp := <-ch
			close(ch)

			// log.Printf("RawBody looks like: %v \n", string(resp.rawBody))
			So(resp.err, ShouldBeNil)
			So(resp.response.StatusCode, ShouldEqual, http.StatusOK)
			So(string(resp.rawBody), ShouldEqual, "pokemon")
		})

	})
}

type RewriteTransport struct {
	Transport http.RoundTripper
	URL       *url.URL
}

func (t RewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = t.URL.Scheme
	req.URL.Host = t.URL.Host
	req.URL.Path = path.Join(t.URL.Path, req.URL.Path)
	rt := t.Transport
	if rt == nil {
		rt = http.DefaultTransport
	}
	return rt.RoundTrip(req)
}
