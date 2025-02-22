/*
Copyright (c) 2020 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// IMPORTANT: This file has been generated automatically, refrain from modifying it manually as all
// your changes will be lost when the file is generated again.

package v1 // github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1

// IngressBuilder contains the data and logic needed to build 'ingress' objects.
//
// Representation of an ingress.
type IngressBuilder struct {
	bitmap_        uint32
	id             string
	href           string
	dnsName        string
	listening      ListeningMethod
	routeSelectors map[string]string
	default_       bool
}

// NewIngress creates a new builder of 'ingress' objects.
func NewIngress() *IngressBuilder {
	return &IngressBuilder{}
}

// Link sets the flag that indicates if this is a link.
func (b *IngressBuilder) Link(value bool) *IngressBuilder {
	b.bitmap_ |= 1
	return b
}

// ID sets the identifier of the object.
func (b *IngressBuilder) ID(value string) *IngressBuilder {
	b.id = value
	b.bitmap_ |= 2
	return b
}

// HREF sets the link to the object.
func (b *IngressBuilder) HREF(value string) *IngressBuilder {
	b.href = value
	b.bitmap_ |= 4
	return b
}

// Empty returns true if the builder is empty, i.e. no attribute has a value.
func (b *IngressBuilder) Empty() bool {
	return b == nil || b.bitmap_&^1 == 0
}

// DNSName sets the value of the 'DNS_name' attribute to the given value.
func (b *IngressBuilder) DNSName(value string) *IngressBuilder {
	b.dnsName = value
	b.bitmap_ |= 8
	return b
}

// Default sets the value of the 'default' attribute to the given value.
func (b *IngressBuilder) Default(value bool) *IngressBuilder {
	b.default_ = value
	b.bitmap_ |= 16
	return b
}

// Listening sets the value of the 'listening' attribute to the given value.
//
// Cluster components listening method.
func (b *IngressBuilder) Listening(value ListeningMethod) *IngressBuilder {
	b.listening = value
	b.bitmap_ |= 32
	return b
}

// RouteSelectors sets the value of the 'route_selectors' attribute to the given value.
func (b *IngressBuilder) RouteSelectors(value map[string]string) *IngressBuilder {
	b.routeSelectors = value
	if value != nil {
		b.bitmap_ |= 64
	} else {
		b.bitmap_ &^= 64
	}
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *IngressBuilder) Copy(object *Ingress) *IngressBuilder {
	if object == nil {
		return b
	}
	b.bitmap_ = object.bitmap_
	b.id = object.id
	b.href = object.href
	b.dnsName = object.dnsName
	b.default_ = object.default_
	b.listening = object.listening
	if len(object.routeSelectors) > 0 {
		b.routeSelectors = map[string]string{}
		for k, v := range object.routeSelectors {
			b.routeSelectors[k] = v
		}
	} else {
		b.routeSelectors = nil
	}
	return b
}

// Build creates a 'ingress' object using the configuration stored in the builder.
func (b *IngressBuilder) Build() (object *Ingress, err error) {
	object = new(Ingress)
	object.id = b.id
	object.href = b.href
	object.bitmap_ = b.bitmap_
	object.dnsName = b.dnsName
	object.default_ = b.default_
	object.listening = b.listening
	if b.routeSelectors != nil {
		object.routeSelectors = make(map[string]string)
		for k, v := range b.routeSelectors {
			object.routeSelectors[k] = v
		}
	}
	return
}
