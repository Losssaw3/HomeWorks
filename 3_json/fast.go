package main

import (
	"fmt"
	"hw3/models"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/mailru/easyjson"
)

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	/*
		!!! !!! !!!
		обратите внимание - в задании обязательно нужен отчет
		делать его лучше в самом начале, когда вы видите уже узкие места, но еще не оптимизировалм их
		так же обратите внимание на команду в параметром -http
		перечитайте еще раз задание
		!!! !!! !!!
	*/
	data, err := os.ReadFile("data/users.txt")
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(data), "\n")
	var user models.User
	uniqueBrowsers := make(map[string]bool, 150)
	var sb strings.Builder
	var android bool
	var msie bool
	fmt.Fprintln(out, "found users:")
	for idx, line := range lines {
		err := easyjson.Unmarshal([]byte(line), &user)

		if err != nil {
			panic(err)
		}

		for _, browser := range user.Browser {
			if okA, okM := strings.Contains(browser, "Android"), strings.Contains(browser, "MSIE"); okA || okM {
				if _, status := uniqueBrowsers[browser]; !status {
					uniqueBrowsers[browser] = true

				}
				if okA {
					android = true

				} else if okM {
					msie = true
				}

			}
		}
		if msie && android {
			email := strings.Replace(user.Email, "@", " [at] ", 1)
			toOut := FormatWithGrow(&sb, idx, user.Name, email)
			fmt.Fprintln(out, toOut)
			sb.Reset()
		}
		android = false
		msie = false

	}

	fmt.Fprintln(out, "\nTotal unique browsers", len(uniqueBrowsers))
}

func FormatWithGrow(sb *strings.Builder, i int, name, email string) string {
	// Вычисляем размер
	size := 3 + len(strconv.Itoa(i)) + len(name) + 2 + len(email) + 2
	sb.Grow(size)

	sb.WriteByte('[')
	sb.WriteString(strconv.Itoa(i))
	sb.WriteString("] ")
	sb.WriteString(name)
	sb.WriteString(" <")
	sb.WriteString(email)
	sb.WriteString(">")
	return sb.String()
}

func main() {

	out := os.Stdout
	FastSearch(out)
}
