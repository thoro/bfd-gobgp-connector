package main

import (
	"context"
	"io"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	/* module name is api, but might be confusing - use bfdapi instead */
	bfdapi "bitbucket.cf-it.at/creamfinance/gobgpd-bfdd-interconnect/bfd-api"
	bgpapi "bitbucket.cf-it.at/creamfinance/gobgpd-bfdd-interconnect/gobgp-api"
	"bitbucket.cf-it.at/creamfinance/gobgpd-bfdd-interconnect/logging"
)

type InterconnectService struct {
	config    *Config
	bgpClient bgpapi.GobgpApiClient
}

func NewInterconnectService(config *Config) *InterconnectService {
	return &InterconnectService{
		config,
		nil,
	}
}

func (s *InterconnectService) Start() {
	bfdConn, cancel, err := s.newGrpcConnection(s.config.Bfd)
	defer cancel()

	if err != nil {
		log.Errorf("failed to dial bfdd: %v", err)
		return
	}
	defer bfdConn.Close()

	bfdClient := bfdapi.NewBfdApiClient(bfdConn)
	peers, err := listPeers(bfdClient)
	if err != nil {
		log.Errorf("failed to list peers: %v", err)
		return
	}

	bgpConn, cancel, err := s.newGrpcConnection(s.config.Gobgp)
	defer cancel()

	if err != nil {
		log.Errorf("failed to dial gobgpd: %v", err)
		cancel()
		return
	}
	defer bgpConn.Close()

	s.bgpClient = bgpapi.NewGobgpApiClient(bgpConn)

	/*
	 * Needed so that this goroutine won't return before
	 * all monitor-requests are finished servicing.
	 * e.g. basically never, and this runs inifinte until the
	 * interconnecter service is killed using ^C
	 */
	var wg sync.WaitGroup
	wg.Add(len(s.config.Peers))
	defer wg.Wait()

	for name := range s.config.Peers {
		uuid := peers[name]

		stream, err := bfdClient.MonitorPeer(
			context.Background(),
			&bfdapi.MonitorPeerRequest{
				Uuid: uuid,
			},
		)
		if err != nil {
			log.Errorf("failed to create monitor peer request: %v", err)
			wg.Done()
			return
		}

		go s.serviceEvents(name, stream, &wg)
	}
}

/* Handles all stream events from a single peer-monitor client
 * from the bfdd-service
 */
func (s *InterconnectService) serviceEvents(
	name string,
	stream bfdapi.BfdApi_MonitorPeerClient,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Errorf("failed to read bfd monitoring stream: %+v", err)
			break
		}

		local := response.Local
		log.Infof("bfd peer %s changed to %s", name, local.State.String())
		s.handleBfdPeerStateChange(name, local)
	}
}

/* Retrieves all peers available from the bfdd-service
 * Returns a map with `peername` -> `peer-uuid`
 */
func listPeers(client bfdapi.BfdApiClient) (map[string][]byte, error) {
	peers := make(map[string][]byte)

	stream, err := client.ListPeer(context.Background(), &bfdapi.ListPeerRequest{})
	if err != nil {
		return nil, err
	}

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		peers[response.Peer.Name] = response.Uuid
	}

	return peers, nil
}

/* Handles the transitioning for a GoBGP-peer according to the state change
 * of its according bfd peer
 */
func (s *InterconnectService) handleBfdPeerStateChange(bfdName string, peerState *bfdapi.PeerState) {
	bgpPeer := s.config.Peers[bfdName]

	switch peerState.State {
	case bfdapi.SessionState_ADMIN_DOWN:
		fallthrough
	case bfdapi.SessionState_DOWN:
		s.bgpClient.DisablePeer(context.Background(), &bgpapi.DisablePeerRequest{
			Address:       bgpPeer,
			Communication: "disabled by bfd", /* doesn't seem to have any significant value to it */
		})

	case bfdapi.SessionState_UP:
		s.bgpClient.EnablePeer(context.Background(), &bgpapi.EnablePeerRequest{
			Address: bgpPeer,
		})

	default:
		/* This only handles the INIT state, which does not really interest us */
		log.Infof("ignoring session state change %s for peer %s", peerState.State.String(), bfdName)
	}
}

func (s *InterconnectService) newGrpcConnection(server ServerConfig) (*grpc.ClientConn, context.CancelFunc, error) {
	options := []grpc.DialOption{grpc.WithBlock()}

	if server.Tls.Enable {
		var creds credentials.TransportCredentials

		if server.Tls.CertFile == "" {
			creds = credentials.NewClientTLSFromCert(nil, "")
		} else {
			var err error

			log.Infof("%s", server.Tls.CertFile)
			creds, err = credentials.NewClientTLSFromFile(server.Tls.CertFile, "")
			if err != nil {
				return nil, nil, err
			}
		}

		options = append(options, grpc.WithTransportCredentials(creds))
	} else {
		options = append(options, grpc.WithInsecure())
	}

	context, cancel := context.WithTimeout(context.Background(), time.Second)
	conn, err := grpc.DialContext(context, server.Host, options...)
	if err != nil {
		return nil, cancel, err
	}

	return conn, cancel, nil
}
