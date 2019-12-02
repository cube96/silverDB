package cluster

import (
	"github.com/hashicorp/memberlist"
	"io/ioutil"
	"log"
	"stathat.com/c/consistent"
	"strconv"
	"strings"
	"time"
)

type Node interface {
	ShouldProcess(key string) (string,bool)
	Members() []string
    Addr() string
}

type node struct {
	*consistent.Consistent
	addr string
}

func (n *node) Addr() string {
	return n.addr
}

func (n *node) ShouldProcess(key string) (string,bool) {
    addr,_:=n.Get(key)
    return addr,addr==n.addr
}

//var bindPort = flag.Int("port", 8001, "gossip port")


func New(addr string,cluster string) (Node,error) {
	var p1 string
	//a1=strings.Split(addr,":")[0]
	p1=strings.Split(addr,":")[1]
	p, err := strconv.Atoi(p1)
	//rand.Seed(time.Now().Unix())
	//n := rand.Intn(2)
	//conf:=memberlist.DefaultLANConfig()
   // bindPorts:=[]int{7946,7947}
	//bindPort, err := strconv.Atoi(strings.Split(addr,":")[1])
	Localconf:=memberlist.DefaultLocalConfig()
	//hostName,_:=os.Hostname()
	Localconf.Name=addr
	Localconf.BindAddr=addr
	Localconf.BindPort=p
	Localconf.AdvertisePort=p
	Localconf.LogOutput=ioutil.Discard
	l,err:=memberlist.Create(Localconf)
	if err !=nil {
		return nil,err
	}
	if cluster=="" {
		cluster=addr
	}
	clu:=[]string{cluster}
	_,err=l.Join(clu)
	if err !=nil {
		return nil,err
	}
	log.Println(l.Members())
	circle:=consistent.New()
	circle.NumberOfReplicas=256
	go func(){
       for {
       	m:=l.Members()
       	nodes:=make([]string,len(m))
       	for i,node:=range m {
       		nodes[i]=node.Name
		}
       	circle.Set(nodes)
       	time.Sleep(time.Second)
	   }
	}()
	return &node{circle,addr},nil
}



















































