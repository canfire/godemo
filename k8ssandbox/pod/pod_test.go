package pod

import (
	"fmt"
	"testing"
)

func TestPod(t *testing.T) {
	mgr, _ := NewSandboxManager("default")

	err := mgr.InitPods(1)
	if err != nil {
		println("1111111", err)
	}
	fmt.Printf("%+v/n", mgr.pods)
	result, err := mgr.RunCode("import pandas; print(pandas.__version__)")
	if err != nil {
		fmt.Println("22222222", err)
	}
	fmt.Println(result)
}
