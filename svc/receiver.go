package svc

import (
	"fmt"

	"github.com/hyprspace/hyprspace/config"
	"github.com/libp2p/go-libp2p/core/network"
	"go.uber.org/zap"
)

func (sn *ServiceNetwork) streamHandler() func(network.Stream) {
	return func(stream network.Stream) {
		if _, ok := config.FindPeer(sn.config.Peers, stream.Conn().RemotePeer()); !ok {
			logger.Debug("Connection attempt from untrusted peer")
			stream.Reset()
			return
		}
		defer stream.Close()
		buf := make([]byte, 2)
		_, err := stream.Read(buf)
		if err != nil {
			logger.With(err).Error("Failed to read stream")
			return
		}
		svcId := [2]byte(buf)
		if proxy, ok := sn.listeners[svcId]; ok {
			remotePeer := stream.Conn().RemotePeer()
			if _, ok := sn.acl[svcId]; ok {
				if sn.acl[svcId].Blacklist != nil {
					if _, isBlacklisted := sn.acl[svcId].Blacklist[remotePeer]; isBlacklisted {
						logger.With(zap.ByteString("service ID", svcId[:])).Warn("Connection from blacklisted peer")
						stream.Reset()
						return
					}
				}
			}
			// TODO check whitelist
			//
			//
			//
			//
			//
			//
			_, err := stream.Write([]byte{byte(RS_OK)})
			if err != nil {
				logger.With(err).Error("Failed to write stream")
				return
			}
			proxy.Handle(WrapStream(stream))
		} else {
			logger.Info(fmt.Sprintf("%s tried to connect to unknown service %x", stream.Conn().RemotePeer(), svcId))
			_, err := stream.Write([]byte{byte(RS_NOT_SUPPORTED)})
			if err != nil {
				logger.With(err).Error("Failed to send RS_NOT_SUPPORTED")
				return
			}
		}
	}
}
