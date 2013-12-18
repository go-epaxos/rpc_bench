package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"testing"
	"strconv"
)

const (
	ServerHTTP = 0
	ServerTCP = 1
)

type Args struct {
	C string
}

type Foo int

func (t *Foo) Dummy(args *Args, reply *string) error {
	*reply = "hello " + args.C
	return nil
}

var start = 0

func BenchmarkHttpSync(b *testing.B) {
	done := make(chan bool, 10)
	if start != 1 {
		startRPC(ServerHTTP, &start)	
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < 10; i++ {
			go clientSync(ServerHTTP, start, done)
		}
		for i := 0; i < 10; i++ {
			<-done
		}
	}
}

func BenchmarkHttpAsync(b *testing.B) {
	done := make(chan bool, 10)
	if start != 2{
		startRPC(ServerHTTP, &start)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < 10; i++ {
			go clientAsync(ServerHTTP, start, done)
		}
		for i := 0; i < 10; i++ {
			<-done
		}
	}
}

func BenchmarkTCPSync(b *testing.B) {
	done := make(chan bool, 10)
	if start != 3 {
		startRPC(ServerTCP, &start)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < 10; i++ {
			go clientSync(ServerTCP, start, done)
		}
		for i := 0; i < 10; i++ {
			<-done
		}
	}
}

func BenchmarkTCPAsync(b *testing.B) {
	done := make(chan bool, 10)
	if start != 4 {
		startRPC(ServerTCP, &start)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < 10; i++ {
			go clientAsync(ServerTCP, start, done)
		}
		for i := 0; i < 10; i++ {
			<-done
		}
	}
}

func startRPC(stype int, start *int) {
	*start = *start + 1
	foo := new(Foo)
	port := strconv.Itoa(*start+3000)

	rpc.Register(foo)
	if stype == ServerHTTP {
		http.DefaultServeMux = http.NewServeMux() // avoid panic for dup-registering handler
		rpc.HandleHTTP()
		go http.ListenAndServe("localhost:"+port, nil)
	}
	if stype == ServerTCP {
		ln, err := net.Listen("tcp", "localhost:"+port)
		if err != nil {
			log.Fatal("listen error:", err)
		}

		go func() {
			for {
				conn, err := ln.Accept()
				if err != nil {
					log.Fatal("Accept error:", err)
				}
				go rpc.ServeConn(conn)
			}
		}()
	}
}


func clientDial(stype, pt int) *rpc.Client {
	port := strconv.Itoa(pt+3000)
	if stype == ServerHTTP {
		client, err := rpc.DialHTTP("tcp", "localhost:"+port)
		if err != nil {
			log.Fatal("diaHTTP fail:", err)
		}
		return client
	}

	//TCP
	client, err := rpc.Dial("tcp", "localhost:"+port)
	if err != nil {
		log.Fatal("dialTCP fail:", err)
	}
	return client
}

func clientSync(stype, pt int, done chan bool) {
	client := clientDial(stype, pt)
	defer client.Close()

	args := &Args{"Rock Gopher"}
	var reply string

	err := client.Call("Foo.Dummy", args, &reply)
	if err != nil {
		log.Fatal("Dummy error:", err)
	}
	//log.Println(reply)
	done <- true
}

func clientAsync(stype, pt int, done chan bool) {
	client := clientDial(stype, pt)
	defer client.Close()

	args := &Args{"yifan"}
	var reply string

	call := client.Go("Foo.Dummy", args, &reply, nil)
	<-call.Done
	if call.Error != nil {
		log.Fatal("Dummy error:", call.Error)
	}
	//log.Println(reply)
	done <- true
}
