package portunus

//
// import (
// 	"fmt"
// 	"github.com/spf13/cobra"
// 	"github.com/spf13/viper"
// 	"net/http"
// 	"os"
// )
//
// var (
// 	debug   bool
// 	version = "master"
// 	commit  = "unstable"
// )
//
// func initialiseConfig() {
// 	viper.SetConfigName(cfgName)
// 	viper.AddConfigPath("$HOME/.f5")
// 	viper.AddConfigPath(".")
// 	viper.SetDefault("username", "admin")
// 	viper.SetDefault("debug", false)
// 	viper.SetDefault("token", false)
// 	viper.SetDefault("force", false)
// 	viper.SetDefault("statsPathPrefix", "f5")
// 	viper.SetDefault("statsShowZeroValues", false)
// 	viper.SetDefault("dryrun", false)
// 	viper.SetDefault("mergeStrategy", mergo.UniqueFirstSeen.String())
//
// 	viper.SetEnvPrefix("f5")
// 	viper.BindEnv("device")
// 	viper.BindEnv("username")
// 	viper.BindEnv("passwd")
// 	viper.BindEnv("debug")
// 	viper.BindEnv("token")
// 	viper.BindEnv("dryrun")
//
// 	viper.BindPFlag("f5", f5Cmd.PersistentFlags().Lookup("f5"))
// 	viper.BindPFlag("debug", f5Cmd.PersistentFlags().Lookup("debug"))
// 	viper.BindPFlag("input", f5Cmd.PersistentFlags().Lookup("input"))
// 	viper.BindPFlag("dryrun", patchCmd.PersistentFlags().Lookup("dryrun"))
// 	viper.BindPFlag("mergeStrategy", patchCmd.PersistentFlags().Lookup("merge-strategy"))
// 	viper.BindPFlag("pool", onlinePoolMemberCmd.Flags().Lookup("pool"))
// 	viper.BindPFlag("pool", offlinePoolMemberCmd.Flags().Lookup("pool"))
//
// 	// ignore errors - may be using environment vars or cmdline args
// 	viper.ReadInConfig()
//
// }
//
// func checkFlags(cmd *cobra.Command) {
//
// 	debug = viper.GetBool("debug")
// 	token = viper.GetBool("token")
// 	now = viper.GetBool("now")
// 	username = viper.GetString("username")
// 	passwd = viper.GetString("passwd")
// 	f5Host = viper.GetString("device")
// 	if f5Host == "" {
// 		// look for the f5 cmdline option
// 		f5Host = viper.GetString("f5")
// 	}
// 	statsPathPrefix = viper.GetString("stats_path_prefix")
// 	statsShowZeroValues = viper.GetBool("stats_show_zero_values")
//
// 	if username == "" {
// 		fmt.Fprint(os.Stderr, "\nerror: missing username; use config file or F5_USERNAME environment variable\n\n")
// 		os.Exit(1)
// 	}
// 	if passwd == "" {
// 		fmt.Fprint(os.Stderr, "\nerror: missing password; use config file or F5_PASSWD environment variable\n\n")
// 		os.Exit(1)
// 	}
// 	if f5Host == "" {
// 		fmt.Fprint(os.Stderr, "\nerror: missing f5 device hostname; use config file or F5_DEVICE environment variable\n\n")
// 		os.Exit(1)
// 	}
//
// 	// this has to be done here inside cobraCommand.Execute() inc case cmd line args are passed.
// 	// args are only parsed after cobraCommand.Run() - urgh
// 	appliance = f5.New(f5Host, username, passwd, f5.BASIC_AUTH)
// 	appliance.SetDebug(debug)
// 	appliance.SetTokenAuth(token)
// 	appliance.SetStatsPathPrefix(statsPathPrefix)
// 	appliance.SetStatsShowZeroes(statsShowZeroValues)
//
// }
//
// func checkRequiredFlag(flg string) {
// 	if !viper.IsSet(flg) {
// 		fmt.Fprintf(os.Stdout, "\nerror: missing required option --%s\n\n", flg)
// 		os.Exit(1)
// 	}
// }
//
// func init() {
//
// 	f5Cmd.PersistentFlags().StringVarP(&f5Host, "f5", "f", "", "IP or hostname of F5 to poke")
// 	f5Cmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug output")
// 	f5Cmd.PersistentFlags().BoolVarP(&token, "token", "t", false, "use token auth")
// 	f5Cmd.PersistentFlags().StringVarP(&f5Input, "input", "i", "", "input json f5 configuration")
// 	patchCmd.PersistentFlags().BoolVarP(&dryrun, "dryrun", "r", false, "show what would be sent without making changes")
// 	patchCmd.PersistentFlags().StringVarP(&mergeStrategy, "merge-strategy", "m", mergeStrategy, "Stategy for merging patch data; e.g., overwrite, append,\nunique-keep-patch, unique-keep-original")
// 	offlinePoolMemberCmd.Flags().StringVarP(&f5Pool, "pool", "p", "", "F5 pool name")
// 	offlinePoolMemberCmd.Flags().BoolVarP(&now, "now", "n", false, "force member offline immediately")
// 	onlinePoolMemberCmd.Flags().StringVarP(&f5Pool, "pool", "p", "", "F5 pool name")
//
// 	// version
// 	f5Cmd.AddCommand(versionCmd)
// 	f5Cmd.AddCommand(serverCmd)
//
// 	// read config
// 	initialiseConfig()
//
// }
//
// func main() {
// 	//	f5Cmd.DebugFlags()
// 	f5Cmd.Execute()
// }
