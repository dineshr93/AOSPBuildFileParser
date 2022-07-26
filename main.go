package main

import (
	bkparser "AOSPBuildFileParser/blueprint/parser"
	"fmt"
	"os"
)

// mkparser "AOSPBuildFileParser/androidmk/parser"
// set GOOS=linux
// set GOOS=windows
// set GOARCH=amd64 go build
// go build -o aospparse.exe

func main() {

	var keyName, filename string
	if len(os.Args) == 3 {
		filename = os.Args[1]
		keyName = os.Args[2]
	} else {
		fmt.Println("aospparse.exe filename Keyname")
		fmt.Println("./aospparse filename Keyname")
		os.Exit(1)
	}
	a, b := GetValueFromBP(filename, keyName)
	_ = a
	if b != nil {
		for _, v := range b {
			fmt.Println(v)
		}
	} else {
		fmt.Println(a)
		fmt.Println("list is nil")
	}
	// b, err := ioutil.ReadFile("Android.mk")
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }
	// mkp := mkparser.NewParser("Android.mk", bytes.NewBuffer(b))
	// nodes, errs := mkp.Parse()
	// _ = errs

	// fmt.Println("__________Assignment______________")
	// fmt.Println(nodes[15].Dump())
	// fmt.Println("")
	// var a *mkparser.Assignment = nodes[15].(*mkparser.Assignment)
	// fmt.Println("Target:", a.Target)
	// fmt.Println("Name:", *a.Name)
	// fmt.Println("Value:", *a.Value)
	// fmt.Println("Variables:")
	// for _, v := range a.Value.Variables {
	// 	fmt.Println(*v.Name)
	// 	fmt.Println(*v.Name.Variables[0].Name)
	// }
	// fmt.Println("Type:", a.Type)
	// fmt.Println("")
	// fmt.Println("")
	// fmt.Println("__________Rule______________")
	// fmt.Println(nodes[163].Dump())
	// fmt.Println("")
	// var r *mkparser.Rule = nodes[163].(*mkparser.Rule)
	// fmt.Println("Target:", *r.Target)
	// fmt.Println("Prerequisites:", *r.Prerequisites)
	// fmt.Println("RecipePos:", r.RecipePos)
	// fmt.Println("Recipe:")
	// fmt.Println(r.Recipe)
	// fmt.Println("")
	// fmt.Println("")
	// fmt.Println("__________Directive______________")
	// fmt.Println(nodes[37].Dump())
	// fmt.Println("")
	// var d *mkparser.Directive = nodes[37].(*mkparser.Directive)
	// fmt.Println("NamePos:", d.NamePos)
	// fmt.Println("Name:", d.Name)
	// fmt.Println("Args:")
	// for _, v := range d.Args.Variables {
	// 	fmt.Println(*v.Name)
	// 	// fmt.Println(*v.Name.Variables[0].Name)
	// }
	// fmt.Println("EndPos:", d.EndPos)
	// fmt.Println("")
	// fmt.Println("")
	// fmt.Println("__________Comment______________")
	// fmt.Println(nodes[18].Dump())
	// fmt.Println("")
	// var c *mkparser.Comment = nodes[18].(*mkparser.Comment)
	// fmt.Println("Value:", c.CommentPos)
	// fmt.Println("Type:")
	// fmt.Println(c.Comment)
	// fmt.Println("")
	// fmt.Println("")

	// fmt.Println(nodes[17].End())
	// fmt.Println(mkp.Unpack(nodes[17].Pos()))
	// fmt.Println(mkp.Unpack(nodes[17].End()))
	// fmt.Println(mkp.Unpack(nodes[17].End()).Offset)
	// fmt.Println(mkp.Unpack(nodes[18].End()).Offset)
	// for _, n := range nodes {
	// 	fmt.Println(n.Dump())
	// }
}

func GetValueFromBP(filename string, keyName string) (string, []string) {
	f, err := os.Open(filename)
	if err != nil {
		panic("Error in opening the file")
	}
	defer f.Close()

	file, errs := bkparser.Parse(filename, f, bkparser.NewScope(nil))
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Fprintln(os.Stderr, err)
		}
		fmt.Printf("%d parsing errors \n", len(errs))
	}

	result := make([]string, 0)

	// start of the functionality
	for n, def := range file.Defs {
		_ = n

		switch def.(type) {
		case *bkparser.Module:
			m := def.(*bkparser.Module)
			// fmt.Println("Type is ", m.Type)
			// fmt.Println("===============================")
			// fmt.Println(n+1, "Cheking value for key", keyName, " in pkg definition", m.Type)
			// fmt.Println("===============================")

			p, found := m.GetProperty(keyName)
			if found {
				expValue := (*p).Value
				// fmt.Println("expType is ", expValue.Type())

				switch v := expValue.(type) {
				// Value is of string type
				case *bkparser.String:
					// fmt.Println(v.Value)
					result = append(result, v.Value)
					// return "", result
				// Value is of list type
				case *bkparser.List:
					items := expValue.(*bkparser.List)
					// results := make([]string, 0)
					for _, item := range items.Values {
						v := item.(*bkparser.String)

						// fmt.Println(v.Value)
						result = append(result, v.Value)

					}
					// if len(result) > 0 {
					// 	return "", result
					// }
				default:
					fmt.Println("Add case for new value type inside definition of type *bkparser.Module")
				}

			} else {
				// fmt.Println("Key not found")
			}

		case *bkparser.Assignment:
			a := def.(*bkparser.Assignment)
			// fmt.Println(a.Name)
			if a.Name == keyName {
				v, ok := a.Value.(*bkparser.List)
				// fmt.Println(v.Type())
				if ok {
					results := make([]string, 0)
					for _, v := range v.Values {
						v := v.(*bkparser.String)
						// fmt.Println(v.Value)
						results = append(results, v.Value)
					}
					if len(results) > 0 {
						return "", results
					}
				}
			}

		// here v has type S
		default:
			fmt.Println("New Definition interface implementation Type")
		}

	}
	if len(result) > 0 {
		return "", unique(result)
	}
	return "", nil
}

func unique(s []string) []string {
	inResult := make(map[string]bool)
	var result []string
	for _, str := range s {
		if _, ok := inResult[str]; !ok {
			inResult[str] = true
			result = append(result, str)
		}
	}
	return result
}
