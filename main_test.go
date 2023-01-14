package main

import (
	"fmt"
	"testing"
)

func TestEngine_Generate(t *testing.T) {
	engine := &Engine{
		IsDebug: false,
		SrcRoot: "./example/src",
		DstRoot: "./example/dst",
	}
	err := engine.Generate()
	if err != nil {
		fmt.Println(err)
	}
}
