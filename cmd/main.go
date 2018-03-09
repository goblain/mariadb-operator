package main

import (
	"github.com/goblain/mariadb-operator/pkg/initializer"
	"github.com/goblain/mariadb-operator/pkg/operator"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "mdbc",
		Short: "MDBC : MariaDBCluster is a Kubernetes oprator for MariaDB Galera Clusters",
		Long: `An operator that will spin up your cluster and keep it alive
					  as much as possible, recovering from common failures`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	var clusterCmd = &cobra.Command{
		Use:   "cluster",
		Short: "Run the cluster operator",
		Run: func(cmd *cobra.Command, args []string) {
			op := operator.NewOperator()
			op.Start()
		},
	}

	i := &initializer.Initializer{}

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Run as initialization process inside an InitContainer of cluster pods",
		Run: func(cmd *cobra.Command, args []string) {
			i.Run()
		},
	}

	rootCmd.AddCommand(clusterCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.Execute()
}
