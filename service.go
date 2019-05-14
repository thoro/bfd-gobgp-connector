package main

import (
	"context"
	"io"
	"sync"

	"google.golang.org/grpc"

	/* module name is api, but might be confusing - use bfdapi instead */
	bfdapi "bitbucket.cf-it.at/creamfinance/gobgpd-bfdd-interconnect/bfdapi"
	"bitbucket.cf-it.at/creamfinance/gobgpd-bfdd-interconnect/logging"
)

type InterconnectService struct {
	config *Config
}

func NewInterconnectService(config *Config) *InterconnectService {
	return &InterconnectService{
		config,
	}
}

func (s *InterconnectService) Start() {
	// TODO: should probably use TLS
	bfdConn, err := grpc.DialContext(context.Background(), s.config.BfdHost, grpc.WithInsecure())
	if err != nil {
		log.Errorf("failed to dial bfdd: %v", err)
		return
	}

	bfdClient := bfdapi.NewBfdApiClient(bfdConn)
	peers, err := listPeers(bfdClient)
	if err != nil {
		log.Errorf("failed to list peers: %v", err)
		return
	}

	/*
	 * Needed so that this goroutine won't return before
	 * all monitor-requests are finished servicing.
	 * e.g. basically never, and this runs inifinte until the
	 * interconnecter service is killed using ^C
	 */
	var wg sync.WaitGroup
	wg.Add(len(s.config.Peers))
	defer wg.Wait()

	for _, mapping := range s.config.Peers {
		uuid := peers[mapping.BfdPeer]

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

		go s.serviceEvents(mapping.BfdPeer, stream, &wg)
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
			log.Errorf("%+v", err)
			break
		}

		log.Infof("peer '%s' changed: %+v", name, response)
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
