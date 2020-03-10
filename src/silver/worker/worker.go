package worker

import (
	"bufio"
	"io"
	"log"
	"net"
	"silver/cluster"
	"silver/register"
	"silver/storage"
)

type Server struct {
	storage.Storage
	cluster.Node
	*register.DiscoverClient
}

func (s *Server) Listen(storageTyp,addr string) {
	l, e := net.Listen("tcp", addr)
	defer l.Close()
	if e != nil {
		panic(e)
	}
	for {
		c, e := l.Accept()
		if e != nil {
			panic(e)
		}
		go s.process(c,storageTyp)
	}
}

func (s *Server) GetHealthStatus() bool {
	return true
}


func New(c storage.Storage, n cluster.Node,d *register.DiscoverClient) *Server {
	return &Server{c, n,d}
}

func (s *Server) process(conn net.Conn,storageTyp string) {
	switch storageTyp {
	case "tsStorage":
		r := bufio.NewReader(conn)
		getResultCh:=make(chan chan *storage.Value,5000)
		setResultCh:=make(chan chan *storage.WPoint,5000)
		defer close(getResultCh)
		defer close(setResultCh)
		go tsGetReply(conn,getResultCh)
		go tsSetReply(conn,setResultCh)
		go s.getMetaData("/silver/metaData")
		for {
			op, e := r.ReadByte()
			if e != nil {
				if e != io.EOF {
					log.Println("close connection due to error:", e)
				}
				return
			}
			if op == 'S' {
				e = s.tsSet(setResultCh,conn,r)
			} else if op == 'G' {
				e = s.tsGet(getResultCh,conn,r)
			} else if op == 'D' {
			//	e = s.tsDel(resultCh,conn, r)
			} else {
				log.Println("close connection due to invalid operation:", op)
				return
			}
			if e != nil {
				log.Println("close connection due to error:", e)
				return
			}
		}
	default:
		 log.Println("Not Supported Storage Type",storageTyp)
	}
}
