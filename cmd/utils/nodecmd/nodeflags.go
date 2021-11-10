package nodecmd

import (
	"github.com/aidan-kwon/toy-klaytn/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

var CommonNodeFlags = []cli.Flag{
	utils.DataDirFlag,
}