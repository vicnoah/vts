package worker

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTask(t *testing.T) {
	tl := getTaskList()

	Convey("task test", t, func() {
		tl.Add(func() error {
			return nil
		})
		tl.Clean()
		tl.Add(func() error {
			fmt.Println(1)
			return nil
		})
		tl.Add(func() error {
			fmt.Println(2)
			return nil
		})
		tl.Reverse()
		tl.Run()
		tl.Add(func() error {
			fmt.Println(3)
			return nil
		})
		tl.Not()
		tl.Run()
		tl.Is()
		tl.Run()
		tl.Add(func() error {
			fmt.Println(4)
			return nil
		})
		tl.Clean()
		tl.Run()
	})
}
