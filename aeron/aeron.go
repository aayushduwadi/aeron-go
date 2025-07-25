// Copyright 2016 Stanislav Liberman
// Copyright 2022 Talos, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aeron

import (
	"time"

	"github.com/lirm/aeron-go/aeron/broadcast"
	"github.com/lirm/aeron-go/aeron/counters"
	"github.com/lirm/aeron-go/aeron/driver"
	"github.com/lirm/aeron-go/aeron/logging"
	rb "github.com/lirm/aeron-go/aeron/ringbuffer"
	"github.com/lirm/aeron-go/aeron/util/memmap"
)

// NullValue is used to represent a null value for when some value is not yet set.
const NullValue = -1

// NewPublicationHandler is the handler type for new publication notification from the media driver
type NewPublicationHandler func(string, int32, int32, int64)

// NewSubscriptionHandler is the handler type for new subscription notification from the media driver
type NewSubscriptionHandler func(string, int32, int64)

// AvailableImageHandler is the handler type for image available notification from the media driver
type AvailableImageHandler func(Image)

// UnavailableImageHandler is the handler type for image unavailable notification from the media driver
type UnavailableImageHandler func(Image)

// Aeron is the primary interface to the media driver for managing subscriptions and publications
type Aeron struct {
	context            *Context
	conductor          ClientConductor
	toDriverRingBuffer rb.ManyToOne
	driverProxy        driver.Proxy

	counters *counters.MetaDataFlyweight
	cncFile  *memmap.File

	toClientsBroadcastReceiver *broadcast.Receiver
	toClientsCopyReceiver      *broadcast.CopyReceiver
}

var logger = logging.MustGetLogger("aeron")

// Connect is the factory method used to create a new instance of Aeron based on Context settings
func Connect(ctx *Context) (*Aeron, error) {
	aeron := new(Aeron)
	aeron.context = ctx
	logger.Debugf("Connecting with context: %v", ctx)

	ctr, cnc, err := counters.MapFile(ctx.CncFileName())
	if err != nil {
		return nil, err
	}
	aeron.counters = ctr
	aeron.cncFile = cnc

	aeron.toDriverRingBuffer.Init(aeron.counters.ToDriverBuf.Get())

	aeron.driverProxy.Init(&aeron.toDriverRingBuffer)

	aeron.toClientsBroadcastReceiver, err = broadcast.NewReceiver(aeron.counters.ToClientsBuf.Get())
	if err != nil {
		return nil, err
	}

	aeron.toClientsCopyReceiver = broadcast.NewCopyReceiver(aeron.toClientsBroadcastReceiver)

	clientLivenessTo := time.Duration(aeron.counters.ClientLivenessTo.Get())

	aeron.conductor.Init(&aeron.driverProxy, aeron.toClientsCopyReceiver, clientLivenessTo, ctx.mediaDriverTo,
		ctx.publicationConnectionTo, ctx.resourceLingerTo, aeron.counters)

	aeron.conductor.onAvailableImageHandler = ctx.availableImageHandler
	aeron.conductor.onUnavailableImageHandler = ctx.unavailableImageHandler
	aeron.conductor.onNewPublicationHandler = ctx.newPublicationHandler
	aeron.conductor.onNewSubscriptionHandler = ctx.newSubscriptionHandler

	aeron.conductor.errorHandler = ctx.errorHandler

	aeron.conductor.Start(ctx.idleStrategy)

	return aeron, nil
}

// Close will terminate client conductor and remove all publications and subscriptions from the media driver
func (aeron *Aeron) Close() error {
	err := aeron.conductor.Close()
	if nil != err {
		aeron.context.errorHandler(err)
	}

	err = aeron.cncFile.Close()
	if nil != err {
		aeron.context.errorHandler(err)
	}

	return err
}

// AddSubscription will add a new subscription to the driver and wait until it is ready.
func (aeron *Aeron) AddSubscription(channel string, streamID int32) (*Subscription, error) {
	registrationID, err := aeron.conductor.AddSubscription(channel, streamID)
	if err != nil {
		return nil, err
	}
	for {
		subscription, err := aeron.conductor.FindSubscription(registrationID)
		if subscription != nil || err != nil {
			return subscription, err
		}
		aeron.context.idleStrategy.Idle(0)
	}
}

// AddSubscriptionDeprecated will add a new subscription to the driver.
// Returns a channel, which can be used for either blocking or non-blocking want for media driver confirmation
func (aeron *Aeron) AddSubscriptionDeprecated(channel string, streamID int32) chan *Subscription {
	ch := make(chan *Subscription, 1)
	registrationID, err := aeron.conductor.AddSubscription(channel, streamID)
	if err != nil {
		// Preserve the legacy functionality.  The original AddSubscription would result in the ClientConductor calling
		// onError on this, as well as subsequently from the FindSubscription call below.
		aeron.conductor.onError(err)
	}
	go func() {
		for {
			subscription, err := aeron.conductor.FindSubscription(registrationID)
			if subscription != nil || err != nil {
				if err != nil {
					aeron.conductor.onError(err)
				}
				ch <- subscription
				close(ch)
				return
			}
			aeron.context.idleStrategy.Idle(0)
		}
	}()
	return ch
}

// AsyncAddSubscription will add a new subscription to the driver and return its registration ID.  That ID can be used
// to get the Subscription with GetSubscription().
func (aeron *Aeron) AsyncAddSubscription(channel string, streamID int32) (int64, error) {
	return aeron.conductor.AddSubscription(channel, streamID)
}

// GetSubscription will attempt to get a Subscription from a registrationID.  See AsyncAddSubscription.  A pending
// Subscription will return nil,nil signifying that there is neither a Subscription nor an error.
func (aeron *Aeron) GetSubscription(registrationID int64) (*Subscription, error) {
	return aeron.conductor.FindSubscription(registrationID)
}

// AddPublicationDeprecated will add a new publication to the driver. If such publication already exists within ClientConductor
// the same instance will be returned.
// Returns a channel, which can be used for either blocking or non-blocking want for media driver confirmation
func (aeron *Aeron) AddPublicationDeprecated(channel string, streamID int32) chan *Publication {
	ch := make(chan *Publication, 1)

	registrationID, err := aeron.conductor.AddPublication(channel, streamID)
	if err != nil {
		// Preserve the legacy functionality.  The original AddPublication would result in the ClientConductor calling
		// onError on this, as well as subsequently from the FindPublication call below.
		aeron.conductor.onError(err)
	}
	go func() {
		for {
			publication, err := aeron.conductor.FindPublication(registrationID)
			if publication != nil || err != nil {
				if err != nil {
					aeron.conductor.onError(err)
				}
				ch <- publication
				close(ch)
				return
			}
			aeron.context.idleStrategy.Idle(0)
		}
	}()

	return ch
}

// AddPublication will add a new publication to the driver. If such publication already exists within ClientConductor
// the same instance will be returned.
func (aeron *Aeron) AddPublication(channel string, streamID int32) (*Publication, error) {
	registrationID, err := aeron.conductor.AddPublication(channel, streamID)
	if err != nil {
		return nil, err
	}
	for {
		publication, err := aeron.conductor.FindPublication(registrationID)
		if publication != nil || err != nil {
			return publication, err
		}
		aeron.context.idleStrategy.Idle(0)
	}
}

// AsyncAddPublication will add a new publication to the driver and return its registration ID.  That ID can be used to
// get the added Publication with GetPublication().
func (aeron *Aeron) AsyncAddPublication(channel string, streamID int32) (int64, error) {
	return aeron.conductor.AddPublication(channel, streamID)
}

// GetPublication will attempt to get a Publication from a registrationID.  See AsyncAddPublication.  A pending
// Publication will return nil,nil signifying that there is neither a Publication nor an error.
func (aeron *Aeron) GetPublication(registrationID int64) (*Publication, error) {
	return aeron.conductor.FindPublication(registrationID)
}

// AddExclusivePublication will add a new exclusive publication to the driver. If such publication already
// exists within ClientConductor the same instance will be returned.
func (aeron *Aeron) AddExclusivePublication(channel string, streamID int32) (*Publication, error) {
	registrationID, err := aeron.conductor.AddExclusivePublication(channel, streamID)
	if err != nil {
		return nil, err
	}
	for {
		publication, err := aeron.conductor.FindPublication(registrationID)
		if publication != nil || err != nil {
			return publication, err
		}
		aeron.context.idleStrategy.Idle(0)
	}
}

// AddExclusivePublicationDeprecated will add a new exclusive publication to the driver. If such publication already
// exists within ClientConductor the same instance will be returned.
// Returns a channel, which can be used for either blocking or non-blocking want for media driver confirmation
func (aeron *Aeron) AddExclusivePublicationDeprecated(channel string, streamID int32) chan *Publication {
	ch := make(chan *Publication, 1)

	registrationID, err := aeron.conductor.AddExclusivePublication(channel, streamID)
	if err != nil {
		// Preserve the legacy functionality.  The original AddExclusivePublication would result in the ClientConductor
		// calling onError on this, as well as subsequently from the FindPublication call below.
		aeron.conductor.onError(err)
	}
	go func() {
		for {
			publication, err := aeron.conductor.FindPublication(registrationID)
			if publication != nil || err != nil {
				if err != nil {
					aeron.conductor.onError(err)
				}
				ch <- publication
				close(ch)
				return
			}
			aeron.context.idleStrategy.Idle(0)
		}
	}()

	return ch
}

// AsyncAddExclusivePublication will add a new exclusive publication to the driver and return its registration ID.  That
// ID can be used to get the added exclusive Publication with GetExclusivePublication().
func (aeron *Aeron) AsyncAddExclusivePublication(channel string, streamID int32) (int64, error) {
	return aeron.conductor.AddExclusivePublication(channel, streamID)
}

// GetExclusivePublication will attempt to get an exclusive Publication from a registrationID.  See
// AsyncAddExclusivePublication.  A pending Publication will return nil,nil signifying that there is neither a
// Publication nor an error.
// Also note that while aeron-go currently handles GetPublication and GetExclusivePublication the same way, they may
// diverge in the future.  Other Aeron languages already handle these calls differently.
func (aeron *Aeron) GetExclusivePublication(registrationID int64) (*Publication, error) {
	return aeron.conductor.FindPublication(registrationID)
}

// NextCorrelationID generates the next correlation id that is unique for the connected Media Driver.
// This is useful generating correlation identifiers for pairing requests with responses in a clients own
// application protocol.
//
// This method is thread safe and will work across processes that all use the same media driver.
func (aeron *Aeron) NextCorrelationID() int64 {
	return aeron.driverProxy.NextCorrelationID()
}

// ClientID returns the client identity that has been allocated for communicating with the media driver.
func (aeron *Aeron) ClientID() int64 {
	return aeron.driverProxy.ClientID()
}

// CounterReader returns Aeron's clientconductor's counterReader
func (aeron *Aeron) CounterReader() *counters.Reader {
	return aeron.conductor.CounterReader()
}

// IsClosed returns true if this connection is closed.
func (aeron *Aeron) IsClosed() bool {
	return !aeron.conductor.running.Get()
}
