package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc/grpclog"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

var config string

var calicoEtcdConfig calicoEctd

type calicoEctd struct {
	Endpoints   []string `json:"endpoints"`
	EtcdCert    string   `json:"etcd_cert"`
	EtcdCertKey string   `json:"etcd_cert_key"`
	EtcdCa      string   `json:"etcd_ca"`
}

func init() {
	// get calico etcd server config file path
	flag.StringVar(&config, "c", "./tinyRouter.json", "etcd connector info")
	flag.Parse()
	fmt.Println("etcd config path:", config)
	// get calico etcd server config for connect
	if config != "" {
		filePtr, err := os.Open(config)
		if err != nil {
			fmt.Println(err)
			return
		}
		decoder := json.NewDecoder(filePtr)
		err = decoder.Decode(&calicoEtcdConfig)
		if err != nil {
			fmt.Println("Decoder failed", err.Error())
			return
		}
	}

}

type calicoRouteSpec struct {
	Cidr    string `json:"cidr"`
	Deleted string `json:"deleted"`
	Node    string `json:"node"`
	State   string `json:"state"`
}

type calicoStaticRoutes struct {
	Spec calicoRouteSpec `json:"spec"`
}

func routeHandler(specByte []byte) {
	t_struct := calicoStaticRoutes{}
	err := json.Unmarshal(specByte, &t_struct)
	if err != nil {
		//fmt.Println(t_struct.Spec.Cidr, t_struct.Spec.Node, t_struct.Spec.Deleted, t_struct.Spec.State)
	}
	if t_struct.Spec.Deleted == "true" && t_struct.Spec.State == "pendingDeletion" {
		fmt.Println("delete the static route", t_struct.Spec.Cidr, t_struct.Spec.Node)
		routeCmd("del", t_struct.Spec.Cidr, t_struct.Spec.Node)
	}
	if t_struct.Spec.Deleted == "false" && t_struct.Spec.State == "confirmed" {
		fmt.Println("add the static route", t_struct.Spec.Cidr, t_struct.Spec.Node)
		routeCmd("add", t_struct.Spec.Cidr, t_struct.Spec.Node)
	}
}

func main() {
	calicoBgpRouteWatcher()
}

func routeCmd(action string, cidr string, gateway string) {
	// ip route add 10.244.229.192/26 via 12.1.0.252
	cmdStr := "ip route" + " " + action + " " + cidr + " " + "via" + " " + gateway
	cmd := exec.Command("/bin/bash", "-c", cmdStr)
	if err := cmd.Run(); err != nil {
		fmt.Println("Linux command exec err please check your command", err.Error())
		return
	}
}

func calicoBgpRouteWatcher() {
	cert, err := tls.LoadX509KeyPair(calicoEtcdConfig.EtcdCert, calicoEtcdConfig.EtcdCertKey)
	if err != nil {
		fmt.Printf("cert failed, err:%v", err)
		return
	}
	caData, err := ioutil.ReadFile(calicoEtcdConfig.EtcdCa)
	if err != nil {
		return
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caData)

	_tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
	}

	cfg := clientv3.Config{
		Endpoints:   calicoEtcdConfig.Endpoints,
		DialTimeout: 5 * time.Second,
		TLS:         _tlsConfig,
	}

	clientv3.SetLogger(grpclog.NewLoggerV2(os.Stderr, os.Stderr, os.Stderr))
	cli, err := clientv3.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close() // make sure to close the client
	// watch
	rch := cli.Watch(context.Background(), "/registry/crd.projectcalico.org/blockaffinities/", clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			if ev.Type == 0 { // PUT
				routeHandler(ev.Kv.Value)
			}
		}
	}
}
