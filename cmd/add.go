/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/fdanctl/envy/internal/store"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("add called")

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
		preset = fmt.Sprint(preset, ".env")

		if _, ok := store.FindPresetByName(preset); ok {
			fmt.Println("file already exits")
			return
		}

		presetPath := fmt.Sprint(store.GetPresetsFolderPath(), preset)
		editor := os.Getenv("EDITOR")

		// because can have args
		parts := strings.Fields(editor) // returns empty slice if only white space
		if len(parts) == 0 {
			parts = append(parts, "vi")
		}
		fmt.Printf("parts: %v\n", parts)

		editorCmd := exec.Command(parts[0], append(parts[1:], presetPath)...)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr

		if err := editorCmd.Run(); err != nil {
			log.Fatal(err)
		}

		if _, err := os.Stat(presetPath); err != nil {
			fmt.Println("aborted: file was not created")
			return
		}
		fmt.Println("file created")
	},
}

func init() {
	presetCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
