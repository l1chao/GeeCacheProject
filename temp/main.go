package main

import "fmt"

type T struct {
	a int
}

func (t T) Get() int {
	return t.a
}

func (t *T) Set(a int) int {
	t.a = a
	return t.a
}

// 这个是编译器将上面的Set自动转换的结果，虽然能手动调用但是不规范。最多能够用“方法表达式”进行调用。
// func Set(t *T, a int) int {
// 	t.a = a
// 	return t.a
// }

type Myinterface interface {
	Get() int
	Set(int) int
}

func ForTest(mi Myinterface) {
	fmt.Printf("mi.Get(): %v\n", mi.Get())
}

func main() {
	ForTest(&T{a: 9527})

	var t = T{}
	var pt = &T{}
	fmt.Printf("T.Get(t): %v\n", T.Get(t)) // 方法表达式
	// T.Get(pt) // 方法表达式里面没有自动解引用。因为自动解引用是语法糖，语法糖用于新手场景，方法表达式并不是新手场景。
	// (*T).Set(t, 1)
	(*T).Set(pt, 1)

	fmt.Println("-------------------------")

	f1 := (*T).Set
	fmt.Printf("%T", f1)

	(&t).Get()
	(&t).Set(1)
	t.Get()
	t.Set(1)
}

type Handler interface {
	Handle(int) error // 签名: (int) error
}
