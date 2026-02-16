/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:   "use",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("use called")

		var preset string

		if len(args) == 0 {
			cmd.Help()
			return
		} else if args[0] != "-" {
			// In Unix and CLI convention, a single dash "-" often represents stdin instead of a filename.
			preset = args[0]
		} else {
			b, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				log.Fatal("err", err)
			}
			preset = strings.TrimSpace(string(b))
		}
		fmt.Println("'use' is looking for", preset)
		preset = fmt.Sprint(preset, ".env")

		yes, err := cmd.Flags().GetBool("yes")
		if err != nil {
			log.Fatal(err)
		}

		if !yes {
			// works for now, add gum (Charm.sh) later
			// more options like output file name
			reader := bufio.NewReader(os.Stdin)
			fmt.Printf("A .env file found in this directory. Do you want to replace it? [y/N]: ")
			response, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(response)) != "y" {
				os.Exit(0)
			}
		}

		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("err", err)
		}

		path := fmt.Sprintf("%s/.config/envy/", home)
		dir, err := os.ReadDir(path)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				os.MkdirAll(path, 0o755)
			}
		}

		for _, v := range dir {
			if v.Name() == preset {
				f, err := os.Open(fmt.Sprint(path, v.Name()))
				if err != nil {
					log.Fatal(err)
				}
				defer f.Close()

				outFile, err := os.Create(".env")
				if err != nil {
					log.Fatal(err)
				}
				defer outFile.Close()

				_, err = io.Copy(f, outFile)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("copied %s, to .env\n", preset)
				return
			}
		}

		fmt.Println("not found")
	},
}

func init() {
	useCmd.Flags().BoolP("yes", "y", false, "Replace current .env")
	rootCmd.AddCommand(useCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// useCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// useCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
