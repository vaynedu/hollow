package hexcel

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetCSVData(t *testing.T) {
	Convey("GetCSVData", t, func() {
		tempDir := t.TempDir()

		Convey("正常读取 CSV", func() {
			f := filepath.Join(tempDir, "ok.csv")
			So(os.WriteFile(f, []byte("name,age\nAlice,30\nBob,25"), 0644), ShouldBeNil)

			rows, err := GetCSVData(f)
			So(err, ShouldBeNil)
			So(len(rows), ShouldEqual, 3)
			So(rows[0], ShouldResemble, []string{"name", "age"})
			So(rows[1], ShouldResemble, []string{"Alice", "30"})
		})

		Convey("文件不存在返回 open error", func() {
			rows, err := GetCSVData(filepath.Join(tempDir, "nope.csv"))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "open file error")
			So(rows, ShouldBeNil)
		})

		Convey("空文件返回 csv file is empty", func() {
			f := filepath.Join(tempDir, "empty.csv")
			So(os.WriteFile(f, []byte(""), 0644), ShouldBeNil)
			rows, err := GetCSVData(f)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "csv file is empty")
			So(rows, ShouldBeNil)
		})

		Convey("格式错误返回 read csv file error", func() {
			f := filepath.Join(tempDir, "bad.csv")
			So(os.WriteFile(f, []byte("name,age\nAlice,30\nBob,25,"), 0644), ShouldBeNil)
			rows, err := GetCSVData(f)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "read csv file error")
			So(rows, ShouldBeNil)
		})
	})
}
