package datasource

import (
	"fmt"
)

// var a slice for ip addr
var ipArray []string

func getLocalIP() (ips []string, err error) {
	ips = append(ips, "192.168.0.142")
	//ifaces, err := net.Interfaces()
	//if err != nil {
	//	fmt.Println("get ip interfaces error:", err)
	//	return
	//}
	//
	//for _, i := range ifaces {
	//	addrs, errRet := i.Addrs()
	//	if errRet != nil {
	//		continue
	//	}
	//
	//	for _, addr := range addrs {
	//		var ip net.IP
	//		switch v := addr.(type) {
	//		case *net.IPNet:
	//			ip = v.IP
	//			if ip.IsGlobalUnicast() {
	//				ips = append(ips, ip.String())
	//			}
	//		}
	//	}
	//}
	fmt.Println("111111111111111 ips :", ips)
	return
}
