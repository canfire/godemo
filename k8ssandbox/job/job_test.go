package job

import (
	"fmt"
	"testing"
)

func TestA(t *testing.T) {
	a, err := RunPythonInK8sJob("print('Hello from sandbox')")
	println(a)
	fmt.Println(err)
}
