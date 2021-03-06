package mtest

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"golang.org/x/crypto/ssh"
)

const sshTimeout = 3 * time.Minute

var (
	sshClients = make(map[string]*ssh.Client)
)

func sshTo(address string, sshKey ssh.Signer) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(sshKey),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	return ssh.Dial("tcp", address+":22", config)
}

func parsePrivateKey() (ssh.Signer, error) {
	f, err := os.Open(os.Getenv("SSH_PRIVKEY"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return ssh.ParsePrivateKey(data)
}

func prepareSSHClients(addresses ...string) error {
	sshKey, err := parsePrivateKey()
	if err != nil {
		return err
	}

	ch := time.After(sshTimeout)
	for _, a := range addresses {
	RETRY:
		select {
		case <-ch:
			return errors.New("timed out")
		default:
		}
		client, err := sshTo(a, sshKey)
		if err != nil {
			time.Sleep(time.Second)
			goto RETRY
		}
		sshClients[a] = client
	}

	return nil
}

func stopEtcd(client *ssh.Client) error {
	command := "sudo systemctl stop my-etcd.service; sudo rm -rf /home/ubuntu/default.etcd"
	sess, err := client.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	sess.Run(command)
	return nil
}

func runEtcd(client *ssh.Client) error {
	command := "sudo systemd-run --unit=my-etcd.service /data/etcd --listen-client-urls=http://0.0.0.0:2379 --advertise-client-urls=http://localhost:2379 --data-dir /home/ubuntu/default.etcd"
	sess, err := client.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	return sess.Run(command)
}

func stopSabakan() error {
	for _, c := range sshClients {
		sess, err := c.NewSession()
		if err != nil {
			return err
		}

		sess.Run("sudo systemctl reset-failed sabakan.service; sudo systemctl stop sabakan.service; sudo rm -rf /var/lib/sabakan")
		sess.Close()
	}
	return nil
}

func runSabakan() error {
	for _, c := range sshClients {
		sess, err := c.NewSession()
		if err != nil {
			return err
		}

		err = sess.Run("sudo systemd-run --unit=sabakan.service /data/sabakan -config-file /etc/sabakan.yml")
		sess.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func execAt(host string, args ...string) (stdout, stderr []byte, e error) {
	client := sshClients[host]
	sess, err := client.NewSession()
	if err != nil {
		return nil, nil, err
	}
	defer sess.Close()

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	sess.Stdout = outBuf
	sess.Stderr = errBuf
	err = sess.Run(strings.Join(args, " "))
	return outBuf.Bytes(), errBuf.Bytes(), err
}

func execSafeAt(host string, args ...string) string {
	stdout, _, err := execAt(host, args...)
	ExpectWithOffset(1, err).To(Succeed())
	return string(stdout)
}

func localTempFile(body string) *os.File {
	f, err := ioutil.TempFile("", "sabakan-mtest")
	Expect(err).ShouldNot(HaveOccurred())
	f.WriteString(body)
	f.Close()
	return f
}

func sabactl(args ...string) []byte {
	args = append([]string{"--server", "http://" + host1 + ":10080"}, args...)
	command := exec.Command(sabactlPath, args...)
	stdout := new(bytes.Buffer)
	session, err := gexec.Start(command, stdout, GinkgoWriter)
	Ω(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
	return stdout.Bytes()
}

func etcdctl(args ...string) []byte {
	args = append([]string{"--endpoints", "http://" + host1 + ":2379"}, args...)
	command := exec.Command(etcdctlPath, args...)
	command.Env = append(command.Env, "ETCDCTL_API=3")
	stdout := new(bytes.Buffer)
	session, err := gexec.Start(command, stdout, GinkgoWriter)
	Ω(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
	return stdout.Bytes()
}
