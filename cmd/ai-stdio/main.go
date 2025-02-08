package main

import (
	"flag"
	"fmt"
)

func main() {
	isNew := flag.Bool("new", false, "Create a new AI chat")
	isSend := flag.Bool("send", false, "Send the current AI chat to the LLM")
	flag.Parse()

	if *isNew == *isSend {
		fmt.Printf("invalid usage\nusage: ai-stdio -new [model] | -send\n\n")
		flag.PrintDefaults()
	} else if *isNew {
		actionNew(flag.Args())
	} else if *isSend {
		actionSend(flag.Args())
	}
}
