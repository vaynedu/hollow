package hidgenerator

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUuidGenerateRequestID(t *testing.T) {
	Convey("Uuid 实现 IdGenerator", t, func() {
		var gen IdGenerator = NewUuid()
		id1 := gen.GenerateRequestID()
		id2 := gen.GenerateRequestID()
		So(id1, ShouldNotBeEmpty)
		So(id2, ShouldNotBeEmpty)
		So(id1, ShouldNotEqual, id2)
		So(len(id1), ShouldEqual, 36) // 标准 UUID 长度
	})
}
