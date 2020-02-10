package worker

import "sync"

// getTaskList 获取任务执行实例
func getTaskList() *taskList {
	return &taskList{}
}

type taskList struct {
	mu   sync.Mutex
	is   bool
	list []func() error
}

// Add 添加任务
func (t *taskList) Add(f func() error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.list = append(t.list, f)
}

// Clean 清除任务
func (t *taskList) Clean() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.list = make([]func() error, 0)
}

// Is 需要执行
func (t *taskList) Is() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.is = false
}

// Not 不需要执行
func (t *taskList) Not() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.is = true
}

// Run 运行所有任务
func (t *taskList) Run() (err error) {
	if !t.is {
		t.Reverse()
		for _, f := range t.list {
			t.Del(0)
			er := f()
			if er != nil {
				err = er
			}
		}
	}
	return
}

// Del 删除任务
func (t *taskList) Del(index int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.list = append(t.list[:index], t.list[index+1:]...)
}

// Reverse 反转任务
func (t *taskList) Reverse() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, j := 0, len(t.list)-1; i < j; i, j = i+1, j-1 {
		t.list[i], t.list[j] = t.list[j], t.list[i]
	}
}
