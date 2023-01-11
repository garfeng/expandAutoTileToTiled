package main

import (
	"fmt"
	"testing"
)

func TestEngine_Generate(t *testing.T) {
	engine := &Engine{
		IsDebug: true,
		SrcRoot: "./example/test",
		DstRoot: "./example/dst",
	}
	err := engine.Generate()
	if err != nil {
		fmt.Println(err)
	}
}
