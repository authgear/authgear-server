package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cmdes "github.com/authgear/authgear-server/cmd/authgear/elasticsearch"
)

var cmdInternal = &cobra.Command{
	Use:   "internal",
	Short: "Internal commands",
}

var cmdInternalElasticsearch = &cobra.Command{
	Use:   "elasticsearch",
	Short: "Elasticsearch commands",
}

var cmdInternalElasticsearchCreateIndex = &cobra.Command{
	Use:   "create-index",
	Short: "Create the search index of all apps",
	RunE: func(cmd *cobra.Command, args []string) error {
		esURL, err := ArgElasticsearchURL.GetRequired(viper.GetViper())
		if err != nil {
			return err
		}

		client, err := cmdes.MakeClient(esURL)
		if err != nil {
			return err
		}

		err = cmdes.CreateIndex(client)
		if err != nil {
			return err
		}

		return nil
	},
}

var cmdInternalElasticsearchDeleteIndex = &cobra.Command{
	Use:   "delete-index",
	Short: "Delete the search index of all apps",
	RunE: func(cmd *cobra.Command, args []string) error {
		esURL, err := ArgElasticsearchURL.GetRequired(viper.GetViper())
		if err != nil {
			return err
		}

		client, err := cmdes.MakeClient(esURL)
		if err != nil {
			return err
		}

		err = cmdes.DeleteIndex(client)
		if err != nil {
			return err
		}

		return nil
	},
}

var cmdInternalElasticsearchReindex = &cobra.Command{
	Use:   "reindex",
	Short: "Reindex all documents of a given app into the search index",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("It works\n")
		return nil
	},
}

func init() {
	cmdInternal.AddCommand(cmdInternalElasticsearch)
	ArgElasticsearchURL.Bind(cmdInternalElasticsearch.PersistentFlags(), viper.GetViper())
	cmdInternalElasticsearch.AddCommand(cmdInternalElasticsearchCreateIndex)
	cmdInternalElasticsearch.AddCommand(cmdInternalElasticsearchDeleteIndex)
	cmdInternalElasticsearch.AddCommand(cmdInternalElasticsearchReindex)
}
