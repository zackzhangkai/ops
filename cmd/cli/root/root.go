package root

import (
	"fmt"
	"os"

	"github.com/shaowenchen/ops/cmd/cli/host"
	"github.com/shaowenchen/ops/cmd/cli/kube"
	"github.com/shaowenchen/ops/cmd/cli/pipeline"
	"github.com/shaowenchen/ops/cmd/cli/storage"
	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
)

func Execute() {
	RootCmd.AddCommand(host.HostCmd)
	RootCmd.AddCommand(kube.KubeCmd)
	RootCmd.AddCommand(storage.StorageCmd)
	RootCmd.AddCommand(pipeline.PipelineCmd)
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var RootCmd = &cobra.Command{
	Use:   "opscli",
	Short: "a cli tool",
	Long:  `This is a cli tool for ops.`,
}

func init() {
	utils.CreateDir(constants.GetOpscliLogsDir())
}