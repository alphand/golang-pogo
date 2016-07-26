package ptc

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestProvider(t *testing.T) {
	Convey("Setup PTC Provider", t, func() {
		pvd := NewProvider("admin", "password")
		So(pvd, ShouldNotBeNil)
		So(pvd.GetProviderString(), ShouldEqual, "ptc")

		Convey("Check login process", func() {
			ch := make(chan *HTTPResponses, 1)
			pvd.checkLoginProcess(ch)

			resp := <-ch
			close(ch)

			So(resp.url, ShouldEqual, loginURL)
			So(resp.rawBody, ShouldNotBeNil)
		})
	})
}
