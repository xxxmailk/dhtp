package serve

import (
	"bytes"
	"dhtp/conf"
	dhcp "github.com/insomniac/dhcp4"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

func init() {
	// set rand seed
	rand.Seed(time.Now().Unix())
}

func DHCPStart(wg *sync.WaitGroup) {
	defer func() {
		if err := recover(); err != nil {
			panic(err)
		}
	}()
	defer wg.Done()
	c := conf.GetConf()
	serverIP := net.ParseIP(c.Dhcp.ListenIP)
	handler := &DHCPHandler{
		ip:            serverIP,
		leaseDuration: 4 * time.Hour,
		start:         net.ParseIP(c.Dhcp.StartIP),
		leaseRange:    c.Dhcp.Range,
		leases:        make(map[int]lease, 200),
		options: dhcp.Options{
			//dhcp.OptionSubnetMask:       []byte(net.ParseIP(c.Dhcp.NetMask)),
			dhcp.OptionSubnetMask: net.ParseIP(c.Dhcp.NetMask).To4(),
			//dhcp.OptionRouter:           []byte(serverIP), // Presuming Server is also your router
			//dhcp.OptionRouter: []byte{192, 168, 181, 134}, // Presuming Server is also your router
			//dhcp.OptionDomainNameServer: []byte(serverIP), // Presuming Server is also your DNS server
			//dhcp.OptionTFTPServerName:   []byte(c.Dhcp.TftpServer), // tftp_files server address
			dhcp.OptionBootFileName: []byte(c.Dhcp.PxeFile), // set boot filename option
		},
	}
	log.Printf("starting dhcp server and linstening on %s:%s", c.Dhcp.ListenIP, c.Dhcp.ListenPort)
	log.Fatal(dhcp.ListenAndServe(handler))
}

type lease struct {
	nic    string    // Client's CHAddr
	expiry time.Time // When the lease expires
}

type DHCPHandler struct {
	ip            net.IP        // Server IP to use
	options       dhcp.Options  // Options to send to DHCP Clients
	start         net.IP        // Start of IP range to distribute
	leaseRange    int           // Number of IPs to distribute (starting from start)
	leaseDuration time.Duration // Lease period
	leases        map[int]lease // Map to keep track of leases
}

func (h *DHCPHandler) ServeDHCP(p dhcp.Packet, msgType dhcp.MessageType, options dhcp.Options) (d dhcp.Packet) {
	switch msgType {

	case dhcp.Discover:
		free, nic := -1, p.CHAddr().String()
		for i, v := range h.leases { // Find previous lease
			if v.nic == nic {
				free = i
				goto reply
			}
		}
		if free = h.freeLease(); free == -1 {
			return
		}
	reply:
		pkg := ReplyPacket(p, dhcp.Offer, h.ip, dhcp.IPAdd(h.start, free), h.leaseDuration,
			SelectOrderOrAll(h.options, options[dhcp.OptionParameterRequestList]))
		log.Printf("DHTP: dhcp replied a package for client %s discovery", p.CHAddr())
		return pkg

	case dhcp.Request:
		if server, ok := options[dhcp.OptionServerIdentifier]; ok && !net.IP(server).Equal(h.ip) {
			return nil // Message not for this dhcp server
		}
		reqIP := net.IP(options[dhcp.OptionRequestedIPAddress])
		if reqIP == nil {
			reqIP = net.IP(p.CIAddr())
		}

		if len(reqIP) == 4 && !reqIP.Equal(net.IPv4zero) {
			if leaseNum := dhcp.IPRange(h.start, reqIP) - 1; leaseNum >= 0 && leaseNum < h.leaseRange {
				if l, exists := h.leases[leaseNum]; !exists || l.nic == p.CHAddr().String() {
					h.leases[leaseNum] = lease{nic: p.CHAddr().String(), expiry: time.Now().Add(h.leaseDuration)}
					pkg := ReplyPacket(p, dhcp.ACK, h.ip, reqIP, h.leaseDuration,
						SelectOrderOrAll(h.options, options[dhcp.OptionParameterRequestList]))
					log.Printf("DHTP: dhcp replied and allocation an ip address %s to client %s", reqIP, p.CHAddr())
					return pkg
				}
			}
		}
		pkg := ReplyPacket(p, dhcp.NAK, h.ip, nil, 0, nil)
		log.Printf("DHTP: dhcp replied a package to client, package: %v", h.options)
		return pkg

	case dhcp.Release, dhcp.Decline:
		nic := p.CHAddr().String()
		for i, v := range h.leases {
			if v.nic == nic {
				delete(h.leases, i)
				break
			}
		}
	}
	return nil
}

func (h *DHCPHandler) freeLease() int {
	now := time.Now()
	b := rand.Intn(h.leaseRange) // Try random first
	for _, v := range [][]int{{b, h.leaseRange}, {0, b}} {
		for i := v[0]; i < v[1]; i++ {
			if l, ok := h.leases[i]; !ok || l.expiry.Before(now) {
				return i
			}
		}
	}
	return -1
}

func SelectOrderOrAll(src dhcp.Options, options []byte) []dhcp.Option {
	if options == nil {
		opts := make([]dhcp.Option, 0, len(src))
		for i, v := range src {
			opts = append(opts, dhcp.Option{Code: i, Value: v})
		}
		if !bytes.Contains(options, []byte{66}) {
			opts = append(opts, dhcp.Option{Code: 66, Value: src[66]})
		}
		if !bytes.Contains(options, []byte{67}) {
			opts = append(opts, dhcp.Option{Code: 67, Value: src[67]})
		}
		return opts
	}
	return SelectOrder(src, options)
}

func SelectOrder(src dhcp.Options, options []byte) []dhcp.Option {
	opts := make([]dhcp.Option, 0, len(options))
	for _, v := range options {
		if data, ok := src[dhcp.OptionCode(v)]; ok {
			opts = append(opts, dhcp.Option{Code: dhcp.OptionCode(v), Value: data})
		}
	}
	if !bytes.Contains(options, []byte{66}) {
		opts = append(opts, dhcp.Option{Code: 66, Value: src[66]})
	}
	if !bytes.Contains(options, []byte{67}) {
		opts = append(opts, dhcp.Option{Code: 67, Value: src[67]})
	}
	return opts
}

// ReplyPacket creates a reply packet that a Server would send to a client.
// It uses the req Packet param to copy across common/necessary fields to
// associate the reply the request.
func ReplyPacket(req dhcp.Packet, mt dhcp.MessageType, serverId, yIAddr net.IP, leaseDuration time.Duration, options []dhcp.Option) dhcp.Packet {
	p := dhcp.NewPacket(dhcp.BootReply)
	p.SetXId(req.XId())
	p.SetFlags(req.Flags())
	p.SetYIAddr(yIAddr)
	p.SetSIAddr(serverId)
	p.SetGIAddr(req.GIAddr())
	p.SetCHAddr(req.CHAddr())
	p.AddOption(dhcp.OptionDHCPMessageType, []byte{byte(mt)})
	p.AddOption(dhcp.OptionServerIdentifier, serverId.To4())
	if leaseDuration > 0 {
		p.AddOption(dhcp.OptionIPAddressLeaseTime, dhcp.OptionsLeaseTime(leaseDuration))
	}
	for _, o := range options {
		p.AddOption(o.Code, o.Value)
	}
	p.PadToMinSize()
	return p
}
