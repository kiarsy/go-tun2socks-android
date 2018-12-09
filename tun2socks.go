package tun2socks

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"

	vcore "v2ray.com/core"
	vproxyman "v2ray.com/core/app/proxyman"
	vbytespool "v2ray.com/core/common/bytespool"
	vinternet "v2ray.com/core/transport/internet"

	"github.com/eycorsican/go-tun2socks/core"
	"github.com/eycorsican/go-tun2socks/proxy/v2ray"
)

var err error
var lwipStack core.LWIPStack
var v *vcore.Instance
var isStopped = false

// VpnService should be implemented in Java/Kotlin.
type VpnService interface {
	// Protect is just a proxy to the VpnService.protect() method.
	// See also: https://developer.android.com/reference/android/net/VpnService.html#protect(int)
	Protect(fd int) bool
}

// PacketFlow should be implemented in Java/Kotlin.
type PacketFlow interface {
	// WritePacket should writes packets to the TUN fd.
	WritePacket(packet []byte)
}

// Write IP packets to the lwIP stack. Call this function in the main loop of
// the VpnService in Java/Kotlin, which should reads packets from the TUN fd.
func InputPacket(data []byte) {
	lwipStack.Write(data)
}

// StartV2Ray sets up lwIP stack, starts a V2Ray instance and registers the instance as the
// connection handler for tun2socks. `exceptionDomains` and `exceptionIPs` are 1-1 corresponding
// domain-IP pairs that separated by comma, each domain name only allow 1 IP for now.
// FIXME: Allow multiple IPs for each domain name.
func StartV2Ray(packetFlow PacketFlow, vpnService VpnService, configBytes []byte, assetPath, exceptionDomains, exceptionIPs string) {
	if packetFlow != nil {
		if lwipStack == nil {
			// Setup the lwIP stack.
			lwipStack = core.NewLWIPStack()
		}

		// Assets
		os.Setenv("v2ray.location.asset", assetPath)

		// Protect file descriptors of connections dial from the VPN process to prevent infinite loop.
		vinternet.RegisterDialerController(func(network, address string, fd uintptr) error {
			if vpnService.Protect(int(fd)) {
				return nil
			} else {
				return errors.New("failed to protect fd")
			}
		})

		// Share the buffer pool.
		core.SetBufferPool(vbytespool.GetPool(core.BufSize))

		// Start the V2Ray instance.
		v, err = vcore.StartInstance("json", configBytes)
		if err != nil {
			log.Fatal("start V instance failed: %v", err)
		}

		// Configure sniffing settings for traffic coming from tun2socks.
		sniffingConfig := &vproxyman.SniffingConfig{
			Enabled:             true,
			DestinationOverride: strings.Split("tls,http", ","),
		}
		ctx := vproxyman.ContextWithSniffingConfig(context.Background(), sniffingConfig)

		// Using an exception domain-IP map in the handler to prevent infinite loop while resolving
		// proxy server domain names.
		domains := strings.Split(exceptionDomains, ",")
		ips := strings.Split(exceptionIPs, ",")
		var domainIPMap = make(map[string]string, len(domains))
		for idx, _ := range domains {
			domainIPMap[domains[idx]] = ips[idx]
		}

		// Register tun2socks connection handlers.
		vhandler := v2ray.NewHandlerWithExceptionDomains(ctx, v, domainIPMap)
		core.RegisterTCPConnectionHandler(vhandler)
		core.RegisterUDPConnectionHandler(vhandler)

		// Write IP packets back to TUN.
		core.RegisterOutputFn(func(data []byte) (int, error) {
			if !isStopped {
				packetFlow.WritePacket(data)
			}
			return len(data), nil
		})

		isStopped = false
	}
}

func StopV2Ray() {
	isStopped = true
	lwipStack.Close()
	v.Close()
}
