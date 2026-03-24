package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func processLastIdx(out io.Writer, dir os.DirEntry, tabs string, printFiles bool) error {
	var sb strings.Builder
	if printFiles {
		if dir.IsDir() {
			fmt.Fprintln(out, tabs+"└───"+dir.Name())
		} else {
			fileInfo, err := dir.Info()
			if err != nil {
				return err
			}
			if fileInfo.Size() == 0 {
				sb.WriteString("(empty)")
			} else {
				sb.WriteString("(" + strconv.Itoa(int(fileInfo.Size())) + "b)")
			}
			fmt.Fprintln(out, tabs+"└───"+dir.Name(), sb.String())
		}
	} else {
		fmt.Fprintln(out, tabs+"└───"+dir.Name())
	}
	return nil

}

func walkDir(out io.Writer, path string, printFiles bool, tabs string) error {
	dirs, err := os.ReadDir(path)
	var strSizeFormat string
	if err != nil {
		return err
	}
	if printFiles {
		for idx, dir := range dirs {
			if idx == len(dirs)-1 {
				err = processLastIdx(out, dir, tabs, printFiles)
				if err != nil {
					return err
				}
			} else {
				if dir.IsDir() {
					fmt.Fprintln(out, tabs+"├───"+dir.Name())
				} else {
					fileInfo, err := dir.Info()
					if err != nil {
						return err
					}
					if fileInfo.Size() == 0 {
						strSizeFormat = "(empty)"
					} else {
						strSizeFormat = "(" + strconv.Itoa(int(fileInfo.Size())) + "b)"
					}
					fmt.Fprintln(out, tabs+"├───"+dir.Name(), strSizeFormat)
				}

			}
			if dir.IsDir() {
				p := filepath.Join(path, dir.Name())
				if idx == len(dirs)-1 {
					walkDir(out, p, printFiles, tabs+"\t")
				} else {
					walkDir(out, p, printFiles, tabs+"│\t")
				}
			}
		}
	} else {
		var onlyDirs []os.DirEntry
		for _, dir := range dirs {
			if dir.IsDir() {
				onlyDirs = append(onlyDirs, dir)
			}
		}
		for idx, dir := range onlyDirs {
			if idx == len(onlyDirs)-1 {
				processLastIdx(out, dir, tabs, printFiles)
			} else {
				fmt.Fprintln(out, tabs+"├───"+dir.Name())
			}
			p := filepath.Join(path, dir.Name())
			if idx == len(onlyDirs)-1 {
				walkDir(out, p, printFiles, tabs+"\t")
			} else {
				walkDir(out, p, printFiles, tabs+"│\t")
			}
		}

	}
	return nil
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	tabs := ""
	err := walkDir(out, path, printFiles, tabs)
	return err
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err)
	}

}
