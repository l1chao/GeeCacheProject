package main

import "fmt"

type Interface interface {
	M1()
	M2()
}

type X struct {
	Interface // 对于匿名字段，如果要调用，可以这么办：x.Interface.M1()
}

func (x X) M1() {
	fmt.Println("X M1")
}

type Y struct {
}

func (y Y) M1() {
	fmt.Println("Y M1")
}

func (y Y) M2() {
	fmt.Println("Y M2")
}

type Settings struct {
	Theme string
}

type User struct {
	*Settings // 匿名指针字段
}
