package etcd

import (
	"context"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
)

const (
	etcdClientURL = "http://localhost:12379"
	etcdPeerURL   = "http://localhost:12380"
)

func testMain(m *testing.M) int {
	circleci := os.Getenv("CIRCLECI") == "true"
	if circleci {
		code := m.Run()
		os.Exit(code)
	}

	etcdPath, err := ioutil.TempDir("", "sabakan-test")
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("etcd",
		"--data-dir", etcdPath,
		"--initial-cluster", "default="+etcdPeerURL,
		"--listen-peer-urls", etcdPeerURL,
		"--initial-advertise-peer-urls", etcdPeerURL,
		"--listen-client-urls", etcdClientURL,
		"--advertise-client-urls", etcdClientURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
		os.RemoveAll(etcdPath)
	}()

	return m.Run()
}

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func newEtcdClient(prefix string) (*clientv3.Client, error) {
	var clientURL string
	circleci := os.Getenv("CIRCLECI") == "true"
	if circleci {
		clientURL = "http://localhost:2379"
	} else {
		clientURL = etcdClientURL
	}
	c, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{clientURL},
		DialTimeout: 2 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	c.KV = namespace.NewKV(c.KV, prefix)
	c.Watcher = namespace.NewWatcher(c.Watcher, prefix)
	c.Lease = namespace.NewLease(c.Lease, prefix)
	return c, nil
}

func testNewDriver(t *testing.T) (*driver, <-chan struct{}) {
	client, err := newEtcdClient(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse("http://localhost:10080")
	if err != nil {
		t.Fatal(err)
	}
	d := &driver{
		client:       client,
		mi:           newMachinesIndex(),
		advertiseURL: u,
	}
	ch := make(chan struct{}, 8) // buffers post-modify-done signals, up to 8
	go d.startWatching(context.Background(), ch, nil)
	<-ch
	return d, ch
}
