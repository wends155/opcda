package main

import (
	"fmt"

	"github.com/wends155/opcda"
	"github.com/wends155/opcda/com"
)

func main() {
	err := com.Initialize()
	if err != nil {
		panic(err)
	}
	defer com.Uninitialize()
	host := "localhost"
	serverInfos, err := opcda.GetOPCServers(host)
	if err != nil {
		panic(err)
	}
	for _, info := range serverInfos {
		fmt.Printf("ProgID: %s, ClsStr: %s, VerIndProgID: %s\n", info.ProgID, info.ClsStr, info.VerIndProgID)
	}
}
