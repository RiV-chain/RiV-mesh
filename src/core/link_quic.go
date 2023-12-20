package core

import (
	"context"
	"crypto/tls"
	"net"
	"net/url"
	"time"

	"github.com/Arceliar/phony"
	"github.com/apernet/quic-go"
)

type linkQUIC struct {
	phony.Inbox
	*links
	tlsconfig  *tls.Config
	quicconfig *quic.Config
}

type linkQUICStream struct {
	quic.Connection
	quic.Stream
}

type linkQUICListener struct {
	*quic.Listener
	ch <-chan *linkQUICStream
}

func (l *linkQUICListener) Accept() (net.Conn, error) {
	qs := <-l.ch
	if qs == nil {
		return nil, context.Canceled
	}
	return qs, nil
}

func (l *links) newLinkQUIC() *linkQUIC {
	lt := &linkQUIC{
		links:     l,
		tlsconfig: l.core.config.tls.Clone(),
		quicconfig: &quic.Config{
			MaxIdleTimeout:                 time.Minute,
			KeepAlivePeriod:                time.Second * 20,
			TokenStore:                     quic.NewLRUTokenStore(255, 255),
			InitialStreamReceiveWindow:     1145141,
			MaxStreamReceiveWindow:         1145142,
			InitialConnectionReceiveWindow: 1145143,
			MaxConnectionReceiveWindow:     1145144,
			EnableDatagrams:                true,
		},
	}
	return lt
}

func (l *linkQUIC) dial(ctx context.Context, url *url.URL, info linkInfo, options linkOptions) (net.Conn, error) {
	qc, err := quic.DialAddr(ctx, url.Host, l.tlsconfig, l.quicconfig)
	if err != nil {
		return nil, err
	}
	qs, err := qc.OpenStreamSync(ctx)
	if err != nil {
		return nil, err
	}
	return &linkQUICStream{
		Connection: qc,
		Stream:     qs,
	}, nil
}

func (l *linkQUIC) listen(ctx context.Context, url *url.URL, _ string) (net.Listener, error) {
	ql, err := quic.ListenAddr(url.Host, l.tlsconfig, l.quicconfig)
	if err != nil {
		return nil, err
	}
	ch := make(chan *linkQUICStream)
	lql := &linkQUICListener{
		Listener: ql,
		ch:       ch,
	}
	go func() {
		for {
			qc, err := ql.Accept(ctx)
			if err != nil {
				ql.Close()
				return
			}
			qs, err := qc.AcceptStream(ctx)
			if err != nil {
				ql.Close()
				return
			}
			ch <- &linkQUICStream{
				Connection: qc,
				Stream:     qs,
			}
		}
	}()
	return lql, nil
}