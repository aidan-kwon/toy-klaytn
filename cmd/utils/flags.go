// Modifications Copyright 2018 The klaytn Authors
// Copyright 2015 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from cmd/utils/flags.go (2018/06/04).
// Modified and improved for the klaytn development.

package utils

import (
	"github.com/aidan-kwon/toy-klaytn/common"
	"github.com/aidan-kwon/toy-klaytn/log"
	"github.com/aidan-kwon/toy-klaytn/networks/p2p"
	"github.com/aidan-kwon/toy-klaytn/networks/rpc"
	"github.com/aidan-kwon/toy-klaytn/node"
	"github.com/aidan-kwon/toy-klaytn/params"
	"gopkg.in/urfave/cli.v1"
	"os"
	"path/filepath"
	"strings"
)

func InitHelper() {
	cli.AppHelpTemplate = AppHelpTemplate
	cli.CommandHelpTemplate = CommandHelpTemplate
}

// NewApp creates an app with sane defaults.
func NewApp(gitCommit, usage string) *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Author = ""
	//app.Authors = nil
	app.Email = ""
	app.Version = params.Version
	if len(gitCommit) >= 8 {
		app.Version += "-" + gitCommit[:8]
	}
	app.Usage = usage
	return app
}

var (
	DataDirFlag = DirectoryFlag{
		Name:  "datadir",
		Usage: "Data directory for the databases and keystore. This value is only used in local DB.",
		Value: DirectoryString{node.DefaultDataDir()},
	}

	NodeTypeFlag = cli.StringFlag{
		Name:  "nodetype",
		Usage: "Klaytn node type (consensus node (cn), proxy node (pn), endpoint node (en))",
		Value: "en",
	}

	SrvTypeFlag = cli.StringFlag{
		Name:  "srvtype",
		Usage: `json rpc server type ("http", "fasthttp")`,
		Value: "fasthttp",
	}

	// RPC settings
	RPCEnabledFlag = cli.BoolFlag{
		Name:  "rpc",
		Usage: "Enable the HTTP-RPC server",
	}
	RPCListenAddrFlag = cli.StringFlag{
		Name:  "rpcaddr",
		Usage: "HTTP-RPC server listening interface",
		Value: node.DefaultHTTPHost,
	}
	RPCPortFlag = cli.IntFlag{
		Name:  "rpcport",
		Usage: "HTTP-RPC server listening port",
		Value: node.DefaultHTTPPort,
	}
	RPCCORSDomainFlag = cli.StringFlag{
		Name:  "rpccorsdomain",
		Usage: "Comma separated list of domains from which to accept cross origin requests (browser enforced)",
		Value: "",
	}
	RPCVirtualHostsFlag = cli.StringFlag{
		Name:  "rpcvhosts",
		Usage: "Comma separated list of virtual hostnames from which to accept requests (server enforced). Accepts '*' wildcard.",
		Value: strings.Join(node.DefaultConfig.HTTPVirtualHosts, ","),
	}
	RPCApiFlag = cli.StringFlag{
		Name:  "rpcapi",
		Usage: "API's offered over the HTTP-RPC interface",
		Value: "",
	}
	RPCGlobalGasCap = cli.Uint64Flag{
		Name:  "rpc.gascap",
		Usage: "Sets a cap on gas that can be used in klay_call/estimateGas",
	}
	RPCConcurrencyLimit = cli.IntFlag{
		Name:  "rpc.concurrencylimit",
		Usage: "Sets a limit of concurrent connection number of HTTP-RPC server",
		Value: rpc.ConcurrencyLimit,
	}
	WSEnabledFlag = cli.BoolFlag{
		Name:  "ws",
		Usage: "Enable the WS-RPC server",
	}
	WSListenAddrFlag = cli.StringFlag{
		Name:  "wsaddr",
		Usage: "WS-RPC server listening interface",
		Value: node.DefaultWSHost,
	}
	WSPortFlag = cli.IntFlag{
		Name:  "wsport",
		Usage: "WS-RPC server listening port",
		Value: node.DefaultWSPort,
	}
	WSApiFlag = cli.StringFlag{
		Name:  "wsapi",
		Usage: "API's offered over the WS-RPC interface",
		Value: "",
	}
	WSAllowedOriginsFlag = cli.StringFlag{
		Name:  "wsorigins",
		Usage: "Origins from which to accept websockets requests",
		Value: "",
	}
	WSMaxSubscriptionPerConn = cli.IntFlag{
		Name:  "wsmaxsubscriptionperconn",
		Usage: "Allowed maximum subscription number per a websocket connection",
		Value: int(rpc.MaxSubscriptionPerWSConn),
	}
	WSReadDeadLine = cli.Int64Flag{
		Name:  "wsreaddeadline",
		Usage: "Set the read deadline on the underlying network connection in seconds. 0 means read will not timeout",
		Value: rpc.WebsocketReadDeadline,
	}
	WSWriteDeadLine = cli.Int64Flag{
		Name:  "wswritedeadline",
		Usage: "Set the Write deadline on the underlying network connection in seconds. 0 means write will not timeout",
		Value: rpc.WebsocketWriteDeadline,
	}
	WSMaxConnections = cli.IntFlag{
		Name:  "wsmaxconnections",
		Usage: "Allowed maximum websocket connection number",
		Value: 3000,
	}
)

// MakeDataDir retrieves the currently requested data directory, terminating
// if none (or the empty string) is specified. If the node is starting a baobab,
// the a subdirectory of the specified datadir will be used.
func MakeDataDir(ctx *cli.Context) string {
	if path := ctx.GlobalString(DataDirFlag.Name); path != "" {
		return path
	}
	log.Fatalf("Cannot determine default data directory, please set manually (--datadir)")
	return ""
}

// splitAndTrim splits input separated by a comma
// and trims excessive white space from the substrings.
func splitAndTrim(input string) []string {
	result := strings.Split(input, ",")
	for i, r := range result {
		result[i] = strings.TrimSpace(r)
	}
	return result
}

// MigrateFlags sets the global flag from a local flag when it's set.
// This is a temporary function used for migrating old command/flags to the
// new format.
//
// e.g. ken account new --keystore /tmp/mykeystore --lightkdf
//
// is equivalent after calling this method with:
//
// ken --keystore /tmp/mykeystore --lightkdf account new
//
// This allows the use of the existing configuration functionality.
// When all flags are migrated this function can be removed and the existing
// configuration functionality must be changed that is uses local flags
func MigrateFlags(action func(ctx *cli.Context) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		for _, name := range ctx.FlagNames() {
			if ctx.IsSet(name) {
				ctx.GlobalSet(name, ctx.String(name))
			}
		}
		return action(ctx)
	}
}

// setHTTP creates the HTTP RPC listener interface string from the set
// command line flags, returning empty if the HTTP endpoint is disabled.
func setHTTP(ctx *cli.Context, cfg *node.Config) {
	if ctx.GlobalBool(RPCEnabledFlag.Name) && cfg.HTTPHost == "" {
		cfg.HTTPHost = "127.0.0.1"
		if ctx.GlobalIsSet(RPCListenAddrFlag.Name) {
			cfg.HTTPHost = ctx.GlobalString(RPCListenAddrFlag.Name)
		}
	}

	if ctx.GlobalIsSet(RPCPortFlag.Name) {
		cfg.HTTPPort = ctx.GlobalInt(RPCPortFlag.Name)
	}
	if ctx.GlobalIsSet(RPCCORSDomainFlag.Name) {
		cfg.HTTPCors = splitAndTrim(ctx.GlobalString(RPCCORSDomainFlag.Name))
	}
	if ctx.GlobalIsSet(RPCApiFlag.Name) {
		cfg.HTTPModules = splitAndTrim(ctx.GlobalString(RPCApiFlag.Name))
	}
	if ctx.GlobalIsSet(RPCVirtualHostsFlag.Name) {
		cfg.HTTPVirtualHosts = splitAndTrim(ctx.GlobalString(RPCVirtualHostsFlag.Name))
	}
	if ctx.GlobalIsSet(RPCConcurrencyLimit.Name) {
		rpc.ConcurrencyLimit = ctx.GlobalInt(RPCConcurrencyLimit.Name)
		logger.Info("Set the concurrency limit of RPC-HTTP server", "limit", rpc.ConcurrencyLimit)
	}
}

// setWS creates the WebSocket RPC listener interface string from the set
// command line flags, returning empty if the HTTP endpoint is disabled.
func setWS(ctx *cli.Context, cfg *node.Config) {
	if ctx.GlobalBool(WSEnabledFlag.Name) && cfg.WSHost == "" {
		cfg.WSHost = "127.0.0.1"
		if ctx.GlobalIsSet(WSListenAddrFlag.Name) {
			cfg.WSHost = ctx.GlobalString(WSListenAddrFlag.Name)
		}
	}

	if ctx.GlobalIsSet(WSPortFlag.Name) {
		cfg.WSPort = ctx.GlobalInt(WSPortFlag.Name)
	}
	if ctx.GlobalIsSet(WSAllowedOriginsFlag.Name) {
		cfg.WSOrigins = splitAndTrim(ctx.GlobalString(WSAllowedOriginsFlag.Name))
	}
	if ctx.GlobalIsSet(WSApiFlag.Name) {
		cfg.WSModules = splitAndTrim(ctx.GlobalString(WSApiFlag.Name))
	}
	rpc.MaxSubscriptionPerWSConn = int32(ctx.GlobalInt(WSMaxSubscriptionPerConn.Name))
	rpc.WebsocketReadDeadline = ctx.GlobalInt64(WSReadDeadLine.Name)
	rpc.WebsocketWriteDeadline = ctx.GlobalInt64(WSWriteDeadLine.Name)
	rpc.MaxWebsocketConnections = int32(ctx.GlobalInt(WSMaxConnections.Name))
}

func SetP2PConfig(ctx *cli.Context, cfg *p2p.Config) {
	var nodeType string
	if ctx.GlobalIsSet(NodeTypeFlag.Name) {
		nodeType = ctx.GlobalString(NodeTypeFlag.Name)
	} else {
		nodeType = NodeTypeFlag.Value
	}

	cfg.ConnectionType = convertNodeType(nodeType)
	if cfg.ConnectionType == common.UNKNOWNNODE {
		logger.Crit("Unknown node type", "nodetype", nodeType)
	}
	logger.Info("Setting connection type", "nodetype", nodeType, "conntype", cfg.ConnectionType)

	// set bootnodes via this function by check specified parameters
}

func convertNodeType(nodetype string) common.ConnType {
	switch strings.ToLower(nodetype) {
	case "cn", "scn":
		return common.CONSENSUSNODE
	case "pn", "spn":
		return common.PROXYNODE
	case "en", "sen":
		return common.ENDPOINTNODE
	default:
		return common.UNKNOWNNODE
	}
}

// SetNodeConfig applies node-related command line flags to the config.
func SetNodeConfig(ctx *cli.Context, cfg *node.Config) {
	SetP2PConfig(ctx, &cfg.P2P)

	// httptype is http or fasthttp
	if ctx.GlobalIsSet(SrvTypeFlag.Name) {
		cfg.HTTPServerType = ctx.GlobalString(SrvTypeFlag.Name)
	}

	setHTTP(ctx, cfg)
	setWS(ctx, cfg)

	cfg.DataDir = ctx.GlobalString(DataDirFlag.Name)
}

// RegisterCNService adds a CN client to the stack.
// func RegisterCNService(stack *node.Node, cfg *cn.Config) {
// 	err := stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
// 		cfg.WsEndpoint = stack.WSEndpoint()
// 		// fullNode, err := cn.New(ctx, cfg)
// 		// return fullNode, err
// 		// TODO: Klaytn full node will be
// 	})
// 	if err != nil {
// 		log.Fatalf("Failed to register the CN service: %v", err)
// 	}
// }