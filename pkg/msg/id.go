package msg

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"sync"
	"template/pkg/logger/gormx"

	uuid "github.com/satori/go.uuid"
	"go.opentelemetry.io/otel/trace"
)

type idGenerator struct {
	sync.Mutex
	randSource *rand.Rand
	getTraceID func(context.Context) string
	form       func(context.Context) gormx.Logger
}

func (u *idGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	u.Lock()
	sid := trace.SpanID{}
	_, _ = u.randSource.Read(sid[:])
	u.Unlock()
	tidString := u.getTraceID(ctx)
	tid, err := uuid.FromString(tidString)
	if err != nil {
		tid = uuid.NewV4()
		u.form(ctx).Errorf("pre_trace_id: %s,trace_id: %s,err: %v", tidString, tid.String(), err)
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

func DefaultIDGenerator(getTraceID func(context.Context) string, form func(context.Context) gormx.Logger) *idGenerator {
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	return &idGenerator{
		randSource: rand.New(rand.NewSource(rngSeed)),
		getTraceID: getTraceID,
		form:       form,
	}
}
