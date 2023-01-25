package vsock

import (
	"bytes"
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
		l, err := Listen("vsock", address)
		assertNilError(t, err)
		defer l.Close()

		conn1, conn2 := testDial(t, l)

		pingPong(t, conn1, conn2)
		pingPong(t, conn1, conn2)

		conn3, conn4 := testDial(t, l)

		// Check the original connection is still usable.
		pingPong(t, conn1, conn2)
		pingPong(t, conn1, conn2)

		// Check the new connection is usable.
		pingPong(t, conn3, conn4)
		pingPong(t, conn3, conn4)

		// Check the original connection is still usable.
		pingPong(t, conn1, conn2)
		pingPong(t, conn1, conn2)

		// Check the new connection is usable.
		pingPong(t, conn3, conn4)
		pingPong(t, conn3, conn4)
	}
}

func testDial(t *testing.T, l net.Listener) (net.Conn, net.Conn) {
	t.Helper()

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
	assertNilError(t, err)
	t.Cleanup(func() {
		conn.Close()
	})

	ret := <-ch
	if ret.err != nil {
		t.Fatal(ret.err)
	}
	t.Cleanup(func() {
		ret.conn.Close()
	})
	return conn, ret.conn
}

func assertNilError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertLen[T any](t *testing.T, v []T, n int) {
	t.Helper()
	if len(v) != n {
		t.Fatalf("exepcted length %d, got: %d", n, len(v))
	}
}

func assertEq(t *testing.T, a, b []byte) {
	t.Helper()
	if !bytes.Equal(a, b) {
		t.Fatalf("expected %v, got: %v", a, b)
	}
}

func pingPong(t *testing.T, conn1, conn2 net.Conn) {
	ping := []byte("ping")
	pong := []byte("pong")

	n, err := conn1.Write(ping)
	assertNilError(t, err)
	assertLen(t, ping, n)

	buf := make([]byte, 4)
	n, err = conn2.Read(buf)
	assertNilError(t, err)
	assertLen(t, buf, n)
	assertEq(t, ping, buf)

	n, err = conn2.Write(pong)
	assertNilError(t, err)
	assertLen(t, pong, n)

	n, err = conn1.Read(buf)
	assertNilError(t, err)
	assertLen(t, buf, n)
	assertEq(t, pong, buf)
}
