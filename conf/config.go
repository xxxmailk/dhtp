package conf

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var conf Config

type Config struct {
	Http Http `yaml:"http"`
	Tftp Tftp `yaml:"tftp"`
	Dhcp Dhcp `yaml:"dhcp"`
}

type Http struct {
	HttpIP    string `yaml:"listen_ip,omitempty"`   // which ip address that http server listening
	HttpPort  string `yaml:"listen_port,omitempty"` // listening port of http server
	MountPath string `yaml:"mount_path,omitempty"`  // http file server path
}

type Tftp struct {
	TftpPath string `yaml:"mount_path,omitempty"` // tftp_files server path
	TftpIP   string `yaml:"listen_ip,omitempty"`  // ip address that tftp_files server listening on
}

type Dhcp struct {
	ListenIP   string `yaml:"listen_ip,omitempty"` // which ip address that dhcp server was listening on
	ListenPort string `yaml:"listen_port,omitempty"`
	TftpServer string `yaml:"tftp_server,omitempty"`
	StartIP    string `yaml:"start_ip"`
	Range      int    `yaml:"lease_range"`       // lease ip address count
	NetMask    string `yaml:"netmask,omitempty"` // default /24
	PxeFile    string `yaml:"pxe_file"`          // pxe file name
}

// refresh runtime configurations
func Refresh() {
	c := new(Config)
	// set default options
	c.Http.HttpIP = "0.0.0.0"
	c.Http.HttpPort = "80"
	c.Http.MountPath = "/mnt/dhtp/http"
	c.Tftp.TftpIP = "0.0.0.0"
	c.Tftp.TftpPath = "/mnt/dhtp/tftp"
	c.Dhcp.ListenIP = "0.0.0.0"
	c.Dhcp.ListenPort = "67"
	c.Dhcp.StartIP = "169.169.181.2"
	c.Dhcp.Range = 50
	c.Dhcp.PxeFile = "pxelinux.0"
	c.Dhcp.NetMask = "255.255.255.0"
	f, err := ioutil.ReadFile("/etc/dhtp/dhtp.yml")
	if err != nil {
		panic(fmt.Sprintf("read config file from /etc/dhtp/dhtp.conf failed, %s", err))
	}
	err = yaml.Unmarshal(f, c)
	if err != nil {
		panic(fmt.Sprintf("parse config file failed, %s", err))
	}
	conf = *c
}

// return runtime configurations
func GetConf() Config {
	return conf
}
