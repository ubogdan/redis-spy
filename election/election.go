package election

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/europelee/redis-spy/utils"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
)

// Election leader election with raft
type Election struct {
	raftBindAddr utils.NetAddr
	raftDataDir  string
	raftPeers    utils.NetAddrList
}

type Config struct {
	Bind    string `json:"bind"`
	DataDir string `json:"data_dir"`
	Peers   string `json:"peers"`
}

type fsm struct {
}

func (f *fsm) Apply(*raft.Log) interface{} {
	return nil
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	return nil, nil
}

func (f *fsm) Restore(io.ReadCloser) error {
	return nil
}

var configFilePath string

func pathExists(p string) bool {
	if _, err := os.Lstat(p); err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

// New return an Election instance
func New(raftBindAddr utils.NetAddr, raftDataDir string, raftPeers utils.NetAddrList) *Election {
	return &Election{
		raftBindAddr: raftBindAddr,
		raftDataDir:  raftDataDir,
		raftPeers:    raftPeers}
}

//Start start and maintain leader election and monitor
func (p *Election) Start() error {
	err := os.MkdirAll(p.raftDataDir, 0755)
	if err != nil {
		log.Fatal(err)
	}
	newNode := !pathExists(path.Join(p.raftDataDir, "raft_db"))

	cfg := raft.DefaultConfig()
	cfg.LogOutput = os.Stdout
	cfg.LocalID = raft.ServerID(p.raftBindAddr.String())
	dbStore, err := raftboltdb.NewBoltStore(path.Join(p.raftDataDir, "raft_db"))
	if err != nil {
		log.Fatal(err)
	}
	fileStore, err := raft.NewFileSnapshotStore(p.raftDataDir, 1, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
	trans, err := raft.NewTCPTransport(p.raftBindAddr.String(), nil, 3, 5*time.Second, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	r, err := raft.NewRaft(cfg, &fsm{}, dbStore, dbStore, fileStore, trans)
	if err != nil {
		log.Fatal(err)
	}
	if newNode {
		bootstrapConfig := raft.Configuration{
			Servers: []raft.Server{
				{
					Suffrage: raft.Voter,
					ID:       raft.ServerID(p.raftBindAddr.String()),
					Address:  raft.ServerAddress(p.raftBindAddr.String()),
				},
			},
		}
		// Add known peers to bootstrap
		for _, node := range p.raftPeers.NetAddrs {

			if node.String() == p.raftBindAddr.String() {
				continue
			}

			bootstrapConfig.Servers = append(bootstrapConfig.Servers, raft.Server{
				Suffrage: raft.Voter,
				ID:       raft.ServerID(node.String()),
				Address:  raft.ServerAddress(node.String()),
			})
		}

		f := r.BootstrapCluster(bootstrapConfig)

		if err := f.Error(); err != nil {
			log.Fatalf("error bootstrapping: %s", err)
		}
	}
	t := time.NewTicker(time.Duration(1) * time.Second)

	for {
		select {
		case <-t.C:
			future := r.VerifyLeader()

			fmt.Printf("Showing peers known by %s:\n", p.raftBindAddr.String())

			if err = future.Error(); err != nil {
				fmt.Println("Node is a follower")
			} else {
				fmt.Println("Node is leader")
			}

			cfuture := r.GetConfiguration()

			if err = cfuture.Error(); err != nil {
				log.Fatalf("error getting config: %s", err)
			}

			configuration := cfuture.Configuration()

			for _, server := range configuration.Servers {
				fmt.Println(server.Address)
			}
		}
	}
}
func main1() {
	flag.StringVar(&configFilePath, "cfg", "./config.json", "help")
	flag.Parse()
	buf, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatal(err)
	}

	var v Config
	err = json.Unmarshal(buf, &v)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println(v)
	dataDir := v.DataDir
	fmt.Println(dataDir)
	err = os.MkdirAll(dataDir, 0755)
	if err != nil {
		log.Fatal(err)
		return
	}
	newNode := !pathExists(path.Join(dataDir, "raft_db"))

	cfg := raft.DefaultConfig()
	cfg.LogOutput = os.Stdout
	cfg.LocalID = raft.ServerID(v.Bind)
	// cfg.EnableSingleNode = true

	dbStore, err := raftboltdb.NewBoltStore(path.Join(dataDir, "raft_db"))
	if err != nil {
		log.Fatal(err)
	}
	fileStore, err := raft.NewFileSnapshotStore(dataDir, 1, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
	trans, err := raft.NewTCPTransport(v.Bind, nil, 3, 5*time.Second, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	r, err := raft.NewRaft(cfg, &fsm{}, dbStore, dbStore, fileStore, trans)
	if err != nil {
		log.Fatal(err)
	}

	if newNode {
		bootstrapConfig := raft.Configuration{
			Servers: []raft.Server{
				{
					Suffrage: raft.Voter,
					ID:       raft.ServerID(v.Bind),
					Address:  raft.ServerAddress(v.Bind),
				},
			},
		}

		var peers []string
		peers = strings.Split(v.Peers, ",")
		// Add known peers to bootstrap
		for _, node := range peers {

			if node == v.Bind {
				continue
			}

			bootstrapConfig.Servers = append(bootstrapConfig.Servers, raft.Server{
				Suffrage: raft.Voter,
				ID:       raft.ServerID(node),
				Address:  raft.ServerAddress(node),
			})
		}

		f := r.BootstrapCluster(bootstrapConfig)

		if err := f.Error(); err != nil {
			log.Fatalf("error bootstrapping: %s", err)
		}
	}
	t := time.NewTicker(time.Duration(1) * time.Second)

	for {
		select {
		case <-t.C:
			//fmt.Println(r.Leader())
			future := r.VerifyLeader()

			fmt.Printf("Showing peers known by %s:\n", v.Bind)

			if err = future.Error(); err != nil {
				fmt.Println("Node is a follower")
			} else {
				fmt.Println("Node is leader")
			}

			cfuture := r.GetConfiguration()

			if err = cfuture.Error(); err != nil {
				log.Fatalf("error getting config: %s", err)
			}

			configuration := cfuture.Configuration()

			for _, server := range configuration.Servers {
				fmt.Println(server.Address)
			}
		}
	}

}
