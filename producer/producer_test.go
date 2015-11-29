package producer

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMakeString(t *testing.T) {

	Convey("Given an integer value", t, func() {

		Convey("it should return a string of length equivalent to the passed in value.", func() {
			result, err := makeString(1000)
			So(len(result), ShouldEqual, 1000)
			So(err, ShouldEqual, nil)
		})

	})

	Convey("Given an integer value > 36000", t, func() {

		Convey("it should return an error.", func() {
			result, err := makeString(37000)
			So(len(result), ShouldEqual, 0)
			So(err.Error(), ShouldEqual, "too many bytes")
		})
	})

	Convey("Given an integer value < 1", t, func() {

		Convey("it should return an error.", func() {
			result, err := makeString(0)
			So(len(result), ShouldEqual, 0)
			So(err.Error(), ShouldEqual, "not enough bytes")
		})
	})

}
