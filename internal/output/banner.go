package output

import "fmt"

const bannerASCII = `
   ___                              _      ____        _   
  / _ \___ _  _____ _______ ___ ___| | /| / / /  ___  (_)__
 / , _/ -_) |/ / -_) __(_-</ -_)___/ |/ |/ / _ \/ _ \/ (_-<
/_/|_|\__/|___/\__/_/ /___/\__/    |__/|__/_//_/\___/_/___/
                                                           
`

func PrintBanner(pr *Printer, toolName string, toolVersion string) {
	if pr == nil {
		return
	}

	pr.Println(bannerASCII)
	pr.Println(" haltman.io (https://github.com/haltman-io)")
	pr.Println("")
	pr.Println(fmt.Sprintf(" [codename: %s] - [release: %s]", toolName, toolVersion))
	pr.Println("")
}
