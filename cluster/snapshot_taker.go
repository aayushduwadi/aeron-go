package cluster

import (
	"fmt"
	"time"

	"github.com/lirm/aeron-go/aeron"
	"github.com/lirm/aeron-go/aeron/atomic"
	"github.com/lirm/aeron-go/cluster/codecs"
)

const snapshotTypeId = 2

type snapshotTaker struct {
	marshaller  *codecs.SbeGoMarshaller // currently shared as we're not reentrant (but could be here)
	options     *Options
	publication *aeron.Publication
}

func newSnapshotTaker(
	options *Options,
	publication *aeron.Publication,
) *snapshotTaker {
	return &snapshotTaker{
		marshaller:  codecs.NewSbeGoMarshaller(),
		options:     options,
		publication: publication,
	}
}

func (st *snapshotTaker) markBegin(
	logPosition int64,
	leadershipTermId int64,
	timeUnit codecs.ClusterTimeUnitEnum,
	appVersion int32,
) error {
	return st.markSnapshot(logPosition, leadershipTermId, codecs.SnapshotMark.BEGIN, timeUnit, appVersion)
}

func (st *snapshotTaker) markEnd(
	logPosition int64,
	leadershipTermId int64,
	timeUnit codecs.ClusterTimeUnitEnum,
	appVersion int32,
) error {
	return st.markSnapshot(logPosition, leadershipTermId, codecs.SnapshotMark.END, timeUnit, appVersion)
}

func (st *snapshotTaker) markSnapshot(
	logPosition int64,
	leadershipTermId int64,
	mark codecs.SnapshotMarkEnum,
	timeUnit codecs.ClusterTimeUnitEnum,
	appVersion int32,
) error {
	bytes, err := codecs.SnapshotMarkerPacket(
		st.marshaller,
		st.options.RangeChecking,
		snapshotTypeId,
		logPosition,
		leadershipTermId,
		0,
		mark,
		timeUnit,
		appVersion,
	)
	if err != nil {
		return err
	}
	if ret := st.offer(bytes); ret < 0 {
		return fmt.Errorf("snapshotTaker.offer failed: %d", ret)
	}
	return nil
}

func (st *snapshotTaker) snapshotSession(session ClientSession) error {
	bytes, err := codecs.ClientSessionPacket(st.marshaller, st.options.RangeChecking,
		session.Id(), session.ResponseStreamId(), []byte(session.ResponseChannel()), session.EncodedPrincipal())
	if err != nil {
		return err
	}
	if ret := st.offer(bytes); ret < 0 {
		return fmt.Errorf("snapshotTaker.offer failed: %d", ret)
	}
	return nil
}

// Offer to our request publication
func (st *snapshotTaker) offer(bytes []byte) int64 {
	buffer := atomic.NewBufferSlice(bytes)
	length := int32(len(bytes))
	start := time.Now()
	var ret int64
	for time.Since(start) < st.options.Timeout {
		ret = st.publication.Offer(buffer, 0, length, nil)
		switch ret {
		// Retry on these
		case aeron.NotConnected, aeron.BackPressured, aeron.AdminAction:
			st.options.IdleStrategy.Idle(0)
		// Fail or succeed on other values
		default:
			return ret
		}
	}
	// Give up, returning the last failure
	return ret
}
