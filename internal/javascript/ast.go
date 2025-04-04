package javascript

import (
	"fmt"
	"github.com/robertkrimen/otto"
)

func main() {
	vm := otto.New()

	// JavaScript code to parse
	jsCode := `
		function add(a, b) {
			return a + b;
		}
	`

	// Evaluate the JavaScript code
	_, err := vm.Run(jsCode)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Example: Accessing a function or variable
	value, err := vm.Get("add")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Function 'add':", value)
}
