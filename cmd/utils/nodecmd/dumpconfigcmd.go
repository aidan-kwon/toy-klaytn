package nodecmd

import (
	"github.com/aidan-kwon/toy-klaytn/cmd/utils"
	"github.com/aidan-kwon/toy-klaytn/log"
	"github.com/aidan-kwon/toy-klaytn/node"
	"github.com/aidan-kwon/toy-klaytn/params"
	"gopkg.in/urfave/cli.v1"
)

type klayConfig struct {
	// CN   cn.Config
	Node node.Config
}

func defaultNodeConfig() node.Config {
	cfg := node.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = params.VersionWithCommit(gitCommit)
	cfg.HTTPModules = append(cfg.HTTPModules, "klay", "shh")
	cfg.WSModules = append(cfg.WSModules, "klay", "shh")
	cfg.IPCPath = "klay.ipc"
	return cfg
}


func makeConfigNode(ctx *cli.Context) (*node.Node, klayConfig) {
	// Load defaults.
	cfg := klayConfig{
		// CN:   *cn.GetDefaultConfig(),
		Node: defaultNodeConfig(),
	}

	// TODO: apply Klaytn blockchain specific flags
	utils.SetNodeConfig(ctx, &cfg.Node)
	stack, err := node.New(&cfg.Node)
	if err != nil {
		log.Fatalf("Failed to create the protocol stack: %v", err)
	}

	return stack, cfg
}


func MakeFullNode(ctx *cli.Context) *node.Node {
	stack, _ := makeConfigNode(ctx)

	// TODO: resister CN service later
	// utils.RegisterCNService(stack, &cfg.CN)
	return stack
}

