package store

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
)

func GetPresetsFolderPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("err", err)
	}

	return fmt.Sprintf("%s/.config/envy/", home)
}

// Returns (preset_file_path, exitst)
func FindPresetByName(name string) (string, bool) {
	path := GetPresetsFolderPath()
	dir := FindAllPresets()
	for _, v := range dir {
		if v.Name() == name {
			return fmt.Sprint(path, v.Name()), true
		}
	}
	return "", false
}

func FindAllPresets() []os.DirEntry {
	path := GetPresetsFolderPath()
	dir, err := os.ReadDir(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err := os.MkdirAll(path, 0o755)
			if err != nil {
				log.Fatal(err)
			}
			FindAllPresets()
		}
	}

	return dir
}
