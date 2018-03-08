package main

import (
	"github.com/goblain/mariadb-operator/pkg/operator"
)

func main() {
	op := operator.NewOperator()
	op.Start()
}
