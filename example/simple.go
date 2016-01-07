package main

import (
	"github.com/delicb/ligno"
)

func main() {
	ligno.Info("Some message", "key1", "value1", "key2", "value2")
	ligno.WaitAll()
}
