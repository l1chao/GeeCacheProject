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
	var x = X{
		Interface: Y{},
	}
	x.M1()
	x.M2()

	u := User{Settings: &Settings{Theme: "dark"}}
	fmt.Println(u.Theme) // "dark"
	u.Theme = "light"    // 等同于: u.Settings.Theme = "light"
	fmt.Println(u.Theme) // 通过指针完成的修改，即使有自动解引用的中间环节，最终效果也是指针修改而不是值修改。
}
