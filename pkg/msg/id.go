package msg

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"strings"
	"sync"

	uuid "github.com/satori/go.uuid"
	"go.opentelemetry.io/otel/trace"
)

type idGenerator struct {
	sync.Mutex
	randSource *rand.Rand
	getTraceID func(context.Context) string
}

func (u *idGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	u.Lock()
	sid := trace.SpanID{}
	_, _ = u.randSource.Read(sid[:])
	u.Unlock()
	tid, err := uuid.FromString(strings.TrimPrefix(u.getTraceID(ctx), "req-"))
	if err != nil {
		tid = uuid.NewV4()
	}

	return trace.TraceID(tid), sid
}

func (u *idGenerator) NewSpanID(ctx context.Context, traceID trace.TraceID) trace.SpanID {
	u.Lock()
	sid := trace.SpanID{}
	_, _ = u.randSource.Read(sid[:])
	u.Unlock()
	return sid
}

func DefaultIDGenerator(getTraceID func(context.Context) string) *idGenerator {
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	return &idGenerator{
		randSource: rand.New(rand.NewSource(rngSeed)),
		getTraceID: getTraceID,
	}
}
