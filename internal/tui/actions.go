package tui

import (
	"bufio"
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fdanctl/envy/internal/store"
)

type someMsg struct {
	success bool
	data    string
}

type (
	errMsg error
)

func removeFile(path string) tea.Cmd {
	return func() tea.Msg {
		if err := os.Remove(path); err != nil {
			return someMsg{success: false, data: "not removed"}
		}
		return someMsg{success: true, data: "remove"}
	}
}

func copyFile(srcPath, dstPath string) tea.Cmd {
	return func() tea.Msg {
		src, err := os.Open(srcPath)
		if err != nil {
			return someMsg{success: false}
		}
		defer src.Close()

		dst, err := os.Create(dstPath)
		if err != nil {
			return someMsg{success: false}
		}
		defer dst.Close()

		scanner := bufio.NewScanner(src)
		for scanner.Scan() {
			line := scanner.Bytes()
			line = append(line, '\n')
			io.Writer.Write(dst, line)
		}
		if err := scanner.Err(); err != nil {
			return someMsg{success: false}
		}
		return someMsg{success: true, data: "file copied successfully"}
	}
}

func presetExists(path string) tea.Cmd {
	return func() tea.Msg {
		if _, err := os.Stat(path); err == nil {
			return someMsg{success: false, data: "File exists"}
		}
		return someMsg{success: true, data: "add"}
	}
}

func saveFile(filename, content string) tea.Cmd {
	return func() tea.Msg {
		path := store.GetPresetsFolderPath() + filename + ".env"
		f, err := os.Create(path)
		if err != nil {
			return errMsg(err)
		}
		defer f.Close()
		if _, err := f.WriteString(content); err != nil {
			return errMsg(err)
		}
		return someMsg{success: true, data: "file created"}
	}
}

// utils
func getFileContent(filename string) string {
	path := store.GetPresetsFolderPath() + filename + ".env"
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	return string(b)
}
