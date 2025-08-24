/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"deifzar/asmm8/pkg/amqpM8"
	"deifzar/asmm8/pkg/api8"
	"deifzar/asmm8/pkg/log8"
	"deifzar/asmm8/pkg/notification8"
	"deifzar/asmm8/pkg/utils"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

// launchCmd represents the launch command
var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launch the ASSM API service, indicating the IP address and port to bind.",
	// 	Long: `A longer description that spans multiple lines and likely contains examples
	// and usage of using your command. For example:
	// Cobra is a CLI library for Go that empowers applications.
	// This application is a tool to generate the needed files
	// to quickly create a Cobra application.`,
	Args: cobra.MatchAll(cobra.MaximumNArgs(2), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		ipFlag, err := cmd.Flags().GetString("ip")
		portFlag, err2 := cmd.Flags().GetInt("port")
		if err != nil || err2 != nil {
			log8.BaseLogger.Debug().Msg(err.Error())
			log8.BaseLogger.Debug().Msg(err2.Error())
			log8.BaseLogger.Fatal().Msg("Error in `Launch` command line with some of the arguments.")
			return err
		} else {
			if !utils.IsValidIPAddress(ipFlag) {
				log8.BaseLogger.Fatal().Msg("Error in `Launch` command line. Invalid IP address.")
				return errors.New("no valid IP address")
			}
			if portFlag < 8000 || portFlag > 9000 {
				log8.BaseLogger.Fatal().Msg("Error in `Launch` command line. Error port range: 8000 - 8999.")
				return errors.New("port number between 8000 and 8999")
			}
			address := ipFlag + ":" + fmt.Sprint(portFlag)

			// Set up graceful shutdown (connection pool will be initialized by api8.Init())
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sigChan
				log8.BaseLogger.Info().Msg("Shutdown signal received, cleaning up...")
				amqpM8.CleanupConnectionPool()
				os.Exit(0)
			}()

			var a api8.Api8
			err = a.Init()
			if err != nil {
				log8.BaseLogger.Debug().Msg(err.Error())
				log8.BaseLogger.Fatal().Msg("Error in `Launch` command line when initialising the API endpoint.")
				// Use pool helper for notifications now
				notification8.PoolHelper.PublishSysErrorNotification("Error in `ASMM8 Launch` command line when initialising the API endpoint", "urgent", "asmm8")
				amqpM8.CleanupConnectionPool()
				return err
			}
			a.Routes()

			log8.BaseLogger.Info().Msg("API service successfully starting on " + address)
			a.Run(address) // This blocks until server shutdown

			// Cleanup when server stops
			log8.BaseLogger.Info().Msg("API service stopped, cleaning up connection pool...")
			amqpM8.CleanupConnectionPool()
			return nil
		}
	},
}

func init() {
	rootCmd.AddCommand(launchCmd)

	// Here you will define your flags and configuration settings.
	launchCmd.Flags().String("ip", "0.0.0.0", "IP address bind to the service. By default, it will listen to all IP addresses.")
	launchCmd.Flags().Int("port", 8000, "Port bind to the service. By default, it will listen to port 8000.")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// launchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// launchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
