package main

import (
	"dhtp/conf"
	"dhtp/serve"
	"fmt"
	"log"
	"sync"
)

func main() {
	log.Printf("starting dhtp server...")
	// refresh runtime configurations
	conf.Refresh()
	var wg = new(sync.WaitGroup)
	// starting http server
	go serve.HTTPStart(wg)
	wg.Add(1)
	// starting dhcp server
	go serve.DHCPStart(wg)
	wg.Add(1)
	// starting tftp_files server
	go serve.TFTPStart(wg)
	wg.Add(1)
	// make server wait for http dhcp tftp_files server exit
	wg.Wait()
	fmt.Println("???")
}
