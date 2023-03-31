package main

import (
	"fmt"
	"github.com/dop251/goja"
)

func main() {
	const SCRIPT = `
function f(param) {
    return +param + 2;
}
`
	const SCRIPT1 = `
f(1)
`

	vm := goja.New()
	_, err := vm.RunString(SCRIPT)
	if err != nil {
		panic(err)
	}

	v, _ := vm.RunString(SCRIPT1)
	if err != nil {
		panic(err)
	}
	a := v.Export()
	fmt.Println(a)
	//var fn func(string) string
	//err = vm.ExportTo(vm.Get("f"), &fn)
	//if err != nil {
	//	panic(err)
	//}
	//
	//fmt.Println(fn("40"))
}
