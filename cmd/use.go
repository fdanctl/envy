/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/fdanctl/envy/internal/store"
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

		if _, err := os.Stat(".env"); err == nil && !yes {
			options := "\033[33mr - replace\nc - change output name\nq - quit\n? - print help\033[0m"
			fmt.Print(
				"\033[31mA .env file found in this directory. What do you want to do?\033[0m\n",
				options,
				"\n\n",
			)
		loop:
			for {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("\033[36m\033[1mOption(r,c,q,?):\033[0m ")
				response, _ := reader.ReadByte()

				switch response {
				case 'r', 'R':
					fmt.Println("replace")
					break loop
				case 'c', 'C':
					fmt.Println("change")
					break loop
				case 'q', 'Q':
					fmt.Println("quit")
					break loop
				case '?':
					fmt.Println(options)
				}
			}
		}

		path, ok := store.FindPresetByName(preset)
		if !ok {
			fmt.Printf("preset '%s' was not found\n", preset[:len(preset)-4])
			return
		}
		src, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
		}
		defer src.Close()

		dst, err := os.Create(".env")
		if err != nil {
			log.Fatal(err)
		}
		defer dst.Close()

		scanner := bufio.NewScanner(src)
		for scanner.Scan() {
			line := scanner.Bytes()
			line = append(line, '\n')
			io.Writer.Write(dst, line)
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("copied %s, to .env\n", preset)
	},
}

func init() {
	// TODO (maybe) rename to --force -f
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
