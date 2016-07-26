package auth

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthProvider(t *testing.T) {
	Convey("Given Setting up Unknown Provider", t, func() {

		unkProvider, err := NewProvider("noprovider", "admin", "password")
		So(err, ShouldNotBeNil)

		Convey("Unknown provider is provided", func() {

			Convey("Access token should be empty and error thrown", func() {
				token := unkProvider.GetAccessToken()
				So(token, ShouldEqual, "")
			})

			Convey("Provider string is unknonwn", func() {
				So(unkProvider.GetProviderString(), ShouldEqual, "unknown")
			})

			Convey("Provider cannot login", func() {
				login, err := unkProvider.Login()
				So(login, ShouldEqual, "")
				So(err, ShouldNotBeNil)
				So(unkProvider.GetProviderString(), ShouldEqual, "unknown")
			})
		})
	})
}
