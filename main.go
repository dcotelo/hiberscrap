package main

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/stevedomin/termtable"
)

//HibernateMapping struct for xml
type HibernateMapping struct {
	XMLName xml.Name `xml:"hibernate-mapping"`
	Text    string   `xml:",chardata"`
	Class   struct {
		Text    string `xml:",chardata"`
		Name    string `xml:"name,attr"`
		Table   string `xml:"table,attr"`
		Mutable string `xml:"mutable,attr"`
		ID      struct {
			Text         string `xml:",chardata"`
			Name         string `xml:"name,attr"`
			UnsavedValue string `xml:"unsaved-value,attr"`
			Column       string `xml:"column,attr"`
			Generator    struct {
				Text  string `xml:",chardata"`
				Class string `xml:"class,attr"`
			} `xml:"generator"`
		} `xml:"id"`
		Property []struct {
			Text   string `xml:",chardata"`
			Name   string `xml:"name,attr"`
			Column string `xml:"column,attr"`
			Unique string `xml:"unique,attr"`
		} `xml:"property"`
	} `xml:"class"`
}

func main() {

	FilePath := flag.String("d", "/var/foo", "Base project directory")
	Verbose := flag.Bool("v", false, "Display each hibernate file location inside the table")
	Usage := flag.Bool("u", false, "Display each class usage across the project")
	flag.Parse()

	//check mandatory -d flag
	required := []string{"d"}
	seen := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { seen[f.Name] = true })
	for _, req := range required {
		if !seen[req] {
			// or possibly use `log.Fatalf` instead of:
			fmt.Fprintf(os.Stderr, "missing required -%s argument\n", req)
			os.Exit(2) // the same exit code flag.Parse uses
		}
	}

	files, _ := WalkMatch(*FilePath, "*.xml")

	filesJava, _ := WalkMatch(*FilePath, "*.java")

	t := termtable.NewTable(nil, &termtable.TableOptions{
		Padding:      3,
		UseSeparator: true,
	})
	//if verbose show column
	if *Verbose == true {
		t.SetHeader([]string{"Class", "Table", "Access", "File"})
	} else {
		t.SetHeader([]string{"Class", "Table", "Access"})
	}

	for _, file := range files {
		//display each file info
		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}

		// Convert []byte to string and print to screen
		data := string(content)

		var hiber HibernateMapping

		err2 := xml.Unmarshal([]byte(data), &hiber)
		if err != nil {
			fmt.Printf("error: %v", err2)
			return
		}
		//check if xml has a valid class
		if hiber.Class.Name != "" {

			//parse boolean including default logic
			Mutable := ""
			if hiber.Class.Mutable == "" || hiber.Class.Mutable == "true" {
				Mutable = "Read/Write"
			} else {
				Mutable = "Read"
			}

			if *Verbose == true {
				res := strings.ReplaceAll(file, *FilePath, "")
				t.AddRow([]string{hiber.Class.Name, hiber.Class.Table, Mutable, res})

			} else {
				t.AddRow([]string{hiber.Class.Name, hiber.Class.Table, Mutable})

			}

			if *Usage == true {
				for _, jfile := range filesJava {
					//fmt.Println(file)
					lines, _ := SearchImport(jfile, "import "+hiber.Class.Name)

					for _, line := range lines {
						relPath := strings.ReplaceAll(line[1], *FilePath, "")
						//println(line[0] + line[1])

						if *Verbose == true {

							t.AddRow([]string{" |--" + relPath + ":" + line[0], "", "", ""})

						} else {
							t.AddRow([]string{" |--" + relPath + ":" + line[0], "", ""})

						}
					}
				}
			}

		}
	}

	fmt.Println(t.Render())
}

//WalkMatch get all files on a given path for specific file extention
func WalkMatch(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

//ShowInfo Display information of hibernate xml
func ShowInfo(FilePath string) {

	content, err := ioutil.ReadFile(FilePath)
	if err != nil {
		log.Fatal(err)
	}

	// Convert []byte to string and print to screen
	data := string(content)

	var hiber HibernateMapping

	err2 := xml.Unmarshal([]byte(data), &hiber)
	if err != nil {
		fmt.Printf("error: %v", err2)
		return
	}
	if hiber.Class.Name != "" {
		//fmt.Println("File:", FilePath)
		fmt.Print(hiber.Class.Name + " ")
		fmt.Print(hiber.Class.Table + "\n")

	}

}

//SearchImport check for imported class
func SearchImport(path, searchText string) ([][]string, error) {
	var lines [][]string
	f, err := os.Open(path)
	if err != nil {
		return lines, err
	}
	defer f.Close()

	// Splits on newlines by default.
	scanner := bufio.NewScanner(f)
	var line int
	line = 1
	// https://golang.org/pkg/bufio/#Scanner.Scan
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), searchText) {
			//		fmt.Println(line)
			founds := []string{strconv.Itoa(line), path}
			lines = append(lines, founds)

		}

		line++
	}

	if err := scanner.Err(); err != nil {
		// Handle the error
		return lines, err
	}
	return lines, err
}
