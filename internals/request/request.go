package request

import (
	"bytes"
	"fmt"
	"io"
)

type ParseState string;
const (
	StateInit ParseState ="init"
	StateDone ParseState ="done"
)
type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	state ParseState
}


var ErrStartLine = fmt.Errorf("invalid start line")
var SEPARATOR=[]byte("\r\n")
var ErrInvalidHTTP = fmt.Errorf("invalid HTTP")

func newRequest() *Request {

	return &Request{
		state: StateInit,
	}
}


func parseRequest(b []byte ) (*RequestLine, int ,error) {

	idx:=bytes.Index(b,SEPARATOR)
	if idx==-1 {
		return nil,len(b),nil
	}
	startLine:=b[:idx]
	restOfMsg:=b[idx+len(SEPARATOR):]
	parts:=bytes.Split(startLine,[]byte(" "))
	
	
	if len(parts)!=3  {
		return nil,len(b)-len(restOfMsg),ErrStartLine
	}
	http:=bytes.Split(parts[2],[]byte("/"))

	if  len(http)!=2 || !bytes.Equal(http[0],[]byte("HTTP")) || !bytes.Equal(http[1],[]byte("1.1")) {
		return nil,len(b)-len(restOfMsg),ErrStartLine
	}

	r:=&RequestLine{
		Method: string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:string(http[1]),
	}


	return r,len(b)-len(restOfMsg),nil
}


func RequestFromReader(reader io.Reader) (*Request, error) {
	request:=newRequest();
	buff:=make([]byte,1024)
	buffLen:=0

	for !request.isDone(){
		n,err:=reader.Read(buff[buffLen:])
		if err!=nil{
			return nil,err
		}
		buffLen+=n
		readN,err:=request.parse(buff[:buffLen]);
		if err!=nil{
			return nil,err
		}
		copy(buff,buff[readN:buffLen])
		buffLen-=readN

	}


	return request,nil
}

func (r* Request) parse (data []byte) (int ,error) {
	read:=0
	outer:
	for {
		switch r.state{
		case StateInit:
			rl,n,err:=parseRequest(data[read:])
			if err!=nil{
				return 0,err
			}
			if(n==0) {
				break outer
			}
			r.RequestLine=*rl
			read+=n;
			r.state=StateDone

		case StateDone:
			break outer
		}
	}
	return read,nil
}

func (r* Request) isDone() bool{
	return r.state==StateDone
}