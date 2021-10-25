package geerpc

import (
	"encoding/json"
	"geerpc/codec"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
)

const MagicNumber =0x3bef5c

type Option struct {
	MagicNumber int
	CodecType codec.Type
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType: codec.GobType,
}

/*
* | option / header1 | body1 | header2 | body2 |
*/

type Server struct {

}

func NewServer () *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func (s *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server 接收请求错误")
			return 
		}
		go s.ServerConn(conn)
	}
}

func (s *Server) ServerConn(conn io.ReadWriteCloser) {
	defer func() { _ = conn.Close() }()
	
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc option 错误")
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Println("rpc 无效的 MagicNumber ")
		return
	}
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server 无效的 code type %v", opt.CodecType)
		return
	}
	s.serverCodec(f(conn))
}

var invaliRequest = struct {}{}

func (s *Server) serverCodec(cc codec.Codec) {
	_ = new(sync.Mutex)
	_ = new(sync.WaitGroup)
	
	for {
		req, err := s.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}
		}
	}
	
}

type request struct {
	h	*codec.Header
	argv, reply reflect.Value
}

func (s *Server) readRequest(cc codec.Codec) (interface{}, interface{}) {
	h, err := s.readRequestHeader(cc)

	if err != nil {
		return nil, err
	}
	req := &request{h:h}
	req.argv = reflect.New(reflect.TypeOf(""))
	if err = cc.ReadBody(req.argv.Interface()); err != nil {
		log.Println("rpc 读取argv 错误")
	}
	return req, nil
}

func (s *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err!= io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc 读取header错误")
		}
		return nil, err
	}
	return &h, nil
}

func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}



