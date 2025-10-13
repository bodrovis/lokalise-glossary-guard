package checks

import (
	"fmt"
	"sort"
	"sync"
)

type Status string

const (
	Pass  Status = "PASS"
	Warn  Status = "WARN"
	Fail  Status = "FAIL"
	Error Status = "ERROR"
)

type Result struct {
	Name    string
	Status  Status
	Message string
}

func Passf(name, format string, args ...any) Result {
	return Result{Name: name, Status: Pass, Message: sprintf(format, args...)}
}

func Warnf(name, format string, args ...any) Result {
	return Result{Name: name, Status: Warn, Message: sprintf(format, args...)}
}

func Failf(name, format string, args ...any) Result {
	return Result{Name: name, Status: Fail, Message: sprintf(format, args...)}
}

func Errorf(name, format string, args ...any) Result {
	return Result{Name: name, Status: Error, Message: sprintf(format, args...)}
}

func sprintf(format string, args ...any) string {
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}

type Check interface {
	Name() string
	Run(data []byte, path string, langs []string) Result
	FailFast() bool
	Priority() int
}

type FuncCheck struct {
	name     string
	failFast bool
	priority int
	run      func([]byte, string, []string) Result
}

func (f FuncCheck) Name() string { return f.name }
func (f FuncCheck) Run(data []byte, path string, langs []string) Result {
	return f.run(data, path, langs)
}
func (f FuncCheck) FailFast() bool { return f.failFast }
func (f FuncCheck) Priority() int  { return f.priority }

func FromFunc(name string, failFast bool, priority int, run func([]byte, string, []string) Result) Check {
	return FuncCheck{name: name, failFast: failFast, priority: priority, run: run}
}

var (
	mu  sync.RWMutex
	All []Check
)

func Register(c Check) {
	mu.Lock()
	defer mu.Unlock()

	name := c.Name()

	for i := range All {
		if All[i].Name() == name {
			All[i] = c
			return
		}
	}

	All = append(All, c)
}

func Reset() {
	mu.Lock()
	defer mu.Unlock()
	All = nil
}

func snapshot() []Check {
	mu.RLock()
	defer mu.RUnlock()
	if len(All) == 0 {
		return nil
	}
	out := make([]Check, len(All))
	copy(out, All)
	return out
}

func Split() (critical, normal []Check) {
	list := snapshot()

	for _, c := range list {
		if c.FailFast() {
			critical = append(critical, c)
		} else {
			normal = append(normal, c)
		}
	}

	sort.SliceStable(critical, func(i, j int) bool {
		pi, pj := critical[i].Priority(), critical[j].Priority()
		if pi != pj {
			return pi < pj
		}
		return critical[i].Name() < critical[j].Name()
	})

	return critical, normal
}
