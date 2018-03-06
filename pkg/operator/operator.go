package operator

import (
	mdbclientset "github.com/goblain/mariadb-operator/pkg/generated/clientset/versioned"
)

func NewKubeClientset() *mdbclientset.Clientset {
	cfg, err := InClusterConfig()
	if err != nil {
		panic(err)
	}
	return mdbclientset.NewForConfigOrDie(cfg)
}
