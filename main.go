package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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

	FilePath := flag.String("dir", "", "Base project folder")
	Verbose := flag.Bool("v", false, "Display each hibernate file location inside the table")
	flag.Parse()

	files, _ := WalkMatch(*FilePath, "*.xml")

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
			if hiber.Class.Mutable == "" || hiber.Class.Mutable != "true" {
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
