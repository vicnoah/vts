package worker

// getTaskList 获取任务执行实例
func getTaskList() *taskList {
	return &taskList{}
}

type taskList struct {
	is   bool
	list []func() error
}

// Add 添加任务
func (t *taskList) Add(f func() error) {
	t.list = append(t.list, f)
}

// Clean 清除任务
func (t *taskList) Clean() {
	t.list = make([]func() error, 0)
}

// Is 需要执行
func (t *taskList) Is() {
	t.is = false
}

// Not 不需要执行
func (t *taskList) Not() {
	t.is = true
}

// Run 运行所有任务
func (t *taskList) Run() (err error) {
	if !t.is {
		t.Reverse()
		for _, f := range t.list {
			er := f()
			if er != nil {
				err = er
			}
		}
	}
	return
}

// Reverse 反转任务
func (t *taskList) Reverse() {
	for i, j := 0, len(t.list)-1; i < j; i, j = i+1, j-1 {
		t.list[i], t.list[j] = t.list[j], t.list[i]
	}
}
