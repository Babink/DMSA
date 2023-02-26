package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/multiformats/go-multiaddr"
)

func strmHandler(s network.Stream) {
	log.Println("Got new stream...")
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go writeMsg(rw)
	go readMsg(rw)
}

func writeMsg(rw *bufio.ReadWriter) {

}

func readMsg(rw *bufio.ReadWriter) {}

func main() {
	hostPort := flag.Uint("port", 0, "Host Node Port")
	destIP := flag.String("dest", "", "Destination node addr")
	flag.Parse()
	H, _ := makeHost(*hostPort, rand.Reader)

	if *destIP == "" {
		startPeer(H, strmHandler)
	} else {
		rw, err := connectToNode(H, *destIP)
		if err != nil {
			log.Println(err.Error())
			return
		}

		go writeMsg(rw)
		go readMsg(rw)
	}
	// waiting forever...
	select {}
}

func makeHost(port uint, randomnes io.Reader) (host.Host, error) {
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, randomnes)
	if err != nil {
		log.Println("Error while generating RSA key")
		log.Println(err.Error())
		return nil, err
	}

	addr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)
	newAddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		log.Println("Error while creating new multiaddr")
		log.Println(err.Error())

		return nil, err
	}

	H, err := libp2p.New(
		libp2p.ListenAddrs(newAddr),
		libp2p.Identity(priv),
	)
	if err != nil {
		log.Println("Error while creating a new p2p host")
		log.Println(err.Error())
		return nil, err
	}

	return H, nil
}

func startPeer(h host.Host, streamHdlr network.StreamHandler) {
	log.Println("Destination address not provided, running this node as a host...")
	h.SetStreamHandler("/chat/1.0.0", streamHdlr)
	var port string
	for _, la := range h.Network().ListenAddresses() {
		ptr, err := la.ValueForProtocol(multiaddr.P_TCP)
		if err == nil {
			port = ptr
			break
		}
	}

	log.Println("To connect to this node, run following command", port)
	log.Printf("chat.exe -dest %s/p2p/%s", h.Addrs()[0].String(), h.ID().Pretty())
}

func connectToNode(H host.Host, destinationAddr string) (*bufio.ReadWriter, error) {
	log.Println("Connecting to specific node...", destinationAddr)
	log.Println("Host is running on -->", H.Addrs()[0].String())

	newDestAddr, err := multiaddr.NewMultiaddr(destinationAddr)
	if err != nil {
		log.Println("Error while generating multiaddr for destination addr")
		log.Println(err.Error())
		return nil, err
	}

	peerInfo, err := peer.AddrInfoFromP2pAddr(newDestAddr)
	if err != nil {
		log.Println("Error while getting information from destination addr")
		log.Println(err.Error())
		return nil, err
	}

	// adding newly connected peer ID into host node's peerstore
	H.Peerstore().AddAddrs(peerInfo.ID, peerInfo.Addrs, peerstore.PermanentAddrTTL)
	S, err := H.NewStream(context.Background(), peerInfo.ID, "/chat/1.0.0")
	if err != nil {
		log.Println("Error while creating stream connection with peer")
		log.Println(err.Error())
		return nil, err
	}

	rw := bufio.NewReadWriter(bufio.NewReader(S), bufio.NewWriter(S))
	return rw, nil
}
