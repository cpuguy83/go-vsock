package vsock

import (
	"net"
	"testing"
)

func TestVsock(t *testing.T) {
	t.Run("cid and port", testVsock(t, "1:1234"))
	t.Run("cid with zero port", testVsock(t, "1:0"))
	t.Run("cid only", testVsock(t, "1:"))
	t.Run("port only", testVsock(t, ":1235"))
	t.Run("port only with zero port", testVsock(t, ":0"))
	t.Run("0:0", testVsock(t, "0:0"))
	t.Run("only separator", testVsock(t, ":"))
}

func testVsock(t *testing.T, address string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		l, err := Listen("vsock", address)
		if err != nil {
			t.Fatal(err)
		}
		defer l.Close()

		type acceptRet struct {
			conn net.Conn
			err  error
		}
		ch := make(chan acceptRet, 1)
		go func() {
			conn, err := l.Accept()
			ch <- acceptRet{conn, err}
		}()

		addr := l.Addr().(*VsockAddr)

		conn, err := DialVsock(addr.CID, addr.Port)
		if err != nil {
			t.Log(addr)
			t.Fatal(err)
		}
		defer conn.Close()

		ret := <-ch
		if ret.err != nil {
			t.Fatal(ret.err)
		}
		defer ret.conn.Close()

		if _, err := conn.Write([]byte("ping")); err != nil {
			t.Fatal(err)
		}

		p := make([]byte, 4)
		n, err := ret.conn.Read(p)
		if err != nil {
			t.Fatal(err)
		}
		if string(p[:n]) != "ping" {
			t.Fatalf("unexpected message: %s", string(p[:n]))
		}

		if _, err := ret.conn.Write([]byte("pong")); err != nil {
			t.Fatal(err)
		}

		n, err = conn.Read(p)
		if err != nil {
			t.Fatal(err)
		}
		if string(p[:n]) != "pong" {
			t.Fatalf("unexpected message: %s", string(p[:n]))
		}
	}
}
