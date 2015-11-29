package main

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMain(t *testing.T) {

	Convey("main should get flags and call runApp", t, func() {
		// set args for examples sake
		os.Args = []string{"-s", "192.168.99.100:32772", "-p", "5", "-b", "3600", "-m", "0", "-V"}
		// main()
		// So(app.Name, ShouldEqual, "Coyote")
	})

}
