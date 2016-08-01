package ptc

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestProvider(t *testing.T) {

	Convey("Setup PTC Provider", t, func() {

		loginReqTest := &LoginRequest{Lt: "lt1234", Execution: "execme"}

		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
			case "/sso/login":

				if req.Method == http.MethodPost {
					http.Redirect(rw, req, "https://google.com?ticket=abc12", http.StatusMovedPermanently)
					return
				}

				rw.WriteHeader(http.StatusOK)
				json.NewEncoder(rw).Encode(loginReqTest)

			case "/sso/oauth2.0/accessToken":
				rw.WriteHeader(http.StatusMovedPermanently)
				rw.Write([]byte("access_token=abc123"))
			}

		}))

		u, err := url.Parse(server.URL)
		if err != nil {
			log.Fatal("failed parsing url", err)
		}

		pvd := NewProvider("admin", "password")
		pvd.http = &http.Client{Transport: RewriteTransport{URL: u}}

		So(pvd, ShouldNotBeNil)
		So(pvd.GetProviderString(), ShouldEqual, "ptc")

		Convey("Test Check login process", func() {
			resp, err := pvd.checkLoginProcess()

			So(err, ShouldBeNil)
			So(resp.response.StatusCode, ShouldEqual, http.StatusOK)

			loginReqMarshal, _ := json.Marshal(loginReqTest)
			So(strings.TrimSpace(string(resp.rawBody)), ShouldEqual, string(loginReqMarshal))
		})

		Convey("Test process login", func() {
			loginReqMarshal, _ := json.Marshal(loginReqTest)
			respData := &HTTPResponses{
				rawBody: []byte(string(loginReqMarshal)),
			}

			resp, err := pvd.processLogin(respData)

			So(err, ShouldBeNil)
			So(resp.response.StatusCode, ShouldEqual, http.StatusMovedPermanently)
		})

		Convey("Test process ticket and access token", func() {
			pvd.processTicket("abc12")
			So(pvd.GetAccessToken(), ShouldEqual, "abc123")
		})

		Convey("Test full Login", func() {
			accCode, err := pvd.Login()

			So(err, ShouldBeNil)
			So(accCode, ShouldEqual, "abc123")
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
