package main

import (
	"fmt"
	"os"

	jsonschema "github.com/danilboiko1302/json-schema"
)

func main() {
	dir := "./cmd/test"
	items, _ := os.ReadDir(dir)
	for _, item := range items {
		fmt.Println("Testing " + item.Name())
		dataType := dir + "/" + item.Name()
		subitems, _ := os.ReadDir(dataType)
		for _, subitem := range subitems {
			fmt.Println("Testing " + subitem.Name())
			validation := dataType + "/" + subitem.Name()
			if IsFiles(validation) {
				test(validation)
			} else {
				subitems, _ := os.ReadDir(validation)
				for _, subitem := range subitems {
					fmt.Println("Testing " + subitem.Name())
					validation := validation + "/" + subitem.Name()
					test(validation)
				}
			}
		}
	}

	return
}
func IsFiles(dir string) bool {
	subitems, _ := os.ReadDir(dir)

	if len(subitems) == 0 {
		return false
	}

	if subitems[0].IsDir() {
		return false
	}

	return true
}

func test(validation string) {
	subitems, _ := os.ReadDir(validation)
	for i := 0; i < len(subitems); i += 3 {
		err := jsonschema.Validate(validation+"/"+subitems[i].Name(), validation+"/"+subitems[i+1].Name())
		if err != nil {
			fmt.Println("error " + subitems[i].Name() + " " + err.Error())
		} else {
			fmt.Println("successful " + subitems[i].Name())
		}

		err = jsonschema.Validate(validation+"/"+subitems[i+2].Name(), validation+"/"+subitems[i+1].Name())
		if err != nil {
			fmt.Println("error " + subitems[i+2].Name() + " " + err.Error())
		} else {
			fmt.Println("successful " + subitems[i+2].Name())
		}
	}
}
