package main_test

import (
	"fmt"
	"testing"
)

// 1. 对于结构体内部嵌入的内结构体s1和内结构体s2，s1和s2之间如果方法重叠，直接报错。这和内嵌了两个接口导致了方法重叠不一样。
type WriterA struct{}

func (WriterA) Write() { fmt.Println("A writes") }

type WriterB struct{}

func (WriterB) Write() { fmt.Println("B writes") }

type DualWriter struct {
	WriterA
	WriterB
}

func Test1(t *testing.T) {
	dw := DualWriter{}

	// dw.Write() // 编译错误：ambiguous selector dw.Write

	// 显式调用
	dw.WriterA.Write() // "A writes"
	dw.WriterB.Write() // "B writes"
}

// 2. 所有匿名字段的调用，都是直接用类型名就可以。包括int(x.int)、结构体Internal(x.Intercal)等。
type X struct {
	int
}

func Test2(t *testing.T) {
	x := X{}
	x.int = 998 // 确实直接使用x.int即可。大胆使用类型名来调用匿名字段！
	fmt.Println(x.int)
}

type Myif1 interface {
	M1()
}
type Myif2 interface {
	M1()
}
type If1Impl struct{}

func (i If1Impl) M1() {
	fmt.Println("Myif1 M1")
}

type If2Impl struct{}

func (i If2Impl) M1() {
	fmt.Println("Myif1 M1")
}

type Outer struct {
	Myif1
	Myif2
}

func Test3(t *testing.T) {
	// outer := Outer{
	// 	Myif1: If1Impl{},
	// 	Myif2: If2Impl{},
	// }
	// outer.M1()
}

type Test4Interface interface {
	M1()
	M2()
	M3()
}
type Test4Struct struct {
	Test4Interface
}

func (s Test4Struct) M1() {
	fmt.Println("Test4Struct M1()")
}

func Test4(t *testing.T) {
	s := Test4Struct{}
	s.M1()
	s.M2()
}

func Test5(t *testing.T) {
	var v int
	if v.(type) == v.(type) {

	}
}

func TestX(t *testing.T) {

}
