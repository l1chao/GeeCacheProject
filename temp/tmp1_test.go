package main_test

import (
	"fmt"
	"net/http"
	"testing"
)

type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// 这个空结构体创建出来实际上是冗余的，在表达上面是不准确的。
// 可以这么想：这里我要实现Handler接口，并没有用到结构体里面的字段，只是通过这个结构体从形式上面完成了传参。
// 每一个类型都有属于自己的方法，但是实际上调用方法的不是类型，而是类型实例。为什么不写成函数而要写成方法？这必须是因为方法需要依赖类型的信息，比如struct的方法需要依赖struct里面的字段；这里的HandlerFunc的方法ServeHTTP需要依赖HandlerFunc里面存储的函数地址。
// 回头看方法一，为了实现接口，创建了HandlerImp，但是ServeHTTP并没有使用到HandlerImp里面的字段，这种方法跟空结构体之间没有关系，所以在表达上面怪怪的。
type HandlerImp struct {
}

func (i *HandlerImp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, bitch!")
}

func TestMain1(t *testing.T) {
	http.ListenAndServe(":8080", &HandlerImp{})
}

// 第二种实现方法
type HandlerFunc func(http.ResponseWriter, *http.Request)

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f(w, r)
}

func greeting(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, bitch!")
}

func TestMain2(t *testing.T) {
	http.ListenAndServe(":8080", HandlerFunc(greeting))
}

type Person struct {
	Name string
	Age  int
}

func (m Person) print() {
	fmt.Printf("Name:%v, Age:%v", m.Name, m.Age)
}

func TestMain3(t *testing.T) {
	var p = Person{}
	p.print()
}
