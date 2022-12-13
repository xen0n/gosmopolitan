package main

import (
	"fmt"
	"time"
)

type escapeHatch = string

func pri18ntln(a ...any) (n int, err error) {
	return fmt.Println(a...)
}

func main() {
	fmt.Println("当前系统时间:", time.Now().In(time.Local))
	fmt.Println(escapeHatch("XXX 不应该报告这个"))
	_, _ = pri18ntln("XXX 也不应该报告这个字符串，但应该报出 time.Local", time.Now().In(time.Local))
}
