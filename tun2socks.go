package tun2socks

import (
	"context"
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

type VpnService interface {
	Protect(fd int)
}

type PacketFlow interface {
	WritePacket(packet []byte)
}

func InputPacket(data []byte) {
	lwipStack.Write(data)
}

func StartV2Ray(packetFlow PacketFlow, vpnService VpnService, configBytes []byte, assetPath string) {
	if packetFlow != nil {
		if lwipStack == nil {
			lwipStack = core.NewLWIPStack()
		}

		// Assets
		os.Setenv("v2ray.location.asset", assetPath)

		// Protect file descriptors of connections dial from the VPN process to prevent infinite loop.
		vinternet.RegisterDialerController(func(network, address string, fd uintptr) error {
			vpnService.Protect(int(fd))
			return nil
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

		// Register tun2socks connection handlers.
		vhandler := v2ray.NewHandler(ctx, v)
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
