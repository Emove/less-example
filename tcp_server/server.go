package main

import (
	"context"
	"fmt"

	"net"
	"time"

	"github.com/emove/less"
	"github.com/emove/less/codec/packet"
	"github.com/emove/less/codec/payload"
	"github.com/emove/less/pkg/io/writer"
	"github.com/emove/less/server"
)

func main() {
	// creates a less server
	// adds OnChannel hook and OnChannelClosed hook
	// adds a router
	srv := server.NewServer(":8080",
		server.WithOnChannel(OnChannelHook),
		server.WithOnChannelClosed(OnChannelClosedHook),
		server.WithRouter(router))

	// serving the network
	srv.Run()
	defer func() {
		// shutdown server context
		srv.Shutdown()
	}()

	// mock a client connection
	client, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	// uses less packet codec
	codec := packet.NewVariableLengthCodec()
	// uses less payload codec
	textCodec := payload.NewTextCodec()

	// creates a buffer writer
	writer := writer.NewBufferWriter(client)

	msg := "[msg%d]hello server!"

	for i := 0; i < 5; i++ {
		// send msg per second
		err = codec.Encode(fmt.Sprintf(msg, i), writer, textCodec)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return
		}
		time.Sleep(time.Second)
	}

	_ = client.Close()
	time.Sleep(time.Second)
}

var IDGenerator uint32

type channelCtxKey struct{}

// ChannelContext custom channel context, used to identify channel
type ChannelContext struct {
	ID uint32
	Ch less.Channel
}

// OnChannelHook identifies each channel and print it
func OnChannelHook(ctx context.Context, ch less.Channel) (context.Context, error) {
	IDGenerator++
	fmt.Printf("new channel, id: %d, remote addr: %s\n", IDGenerator, ch.RemoteAddr().String())
	return context.WithValue(ctx, &channelCtxKey{}, &ChannelContext{ID: IDGenerator, Ch: ch}), nil
}

// OnChannelClosedHook prints channel id when channel closed
func OnChannelClosedHook(ctx context.Context, ch less.Channel, err error) {
	cc := ctx.Value(&channelCtxKey{}).(*ChannelContext)
	fmt.Printf("channel closed, id: %d, remote addr: %s ", cc.ID, ch.RemoteAddr().String())
	if err != nil {
		fmt.Printf("due to err: %v", err)
	}
	fmt.Println()
}

// router returns a handler to handle inbound message, it always return echoHandler
func router(ctx context.Context, channel less.Channel, msg interface{}) (less.Handler, error) {
	return echoHandler, nil
}

// echoHandler logic handler
func echoHandler(ctx context.Context, ch less.Channel, msg interface{}) error {
	cc := ctx.Value(&channelCtxKey{}).(*ChannelContext)
	fmt.Printf("receive msg from channel, id: %d, remote: %s, msg: %v\n", cc.ID, ch.RemoteAddr().String(), msg)
	return nil
}
