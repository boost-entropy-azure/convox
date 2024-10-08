/*

Copyright 2020 Convox, Inc.

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

// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	"context"
	time "time"

	atomv1 "github.com/convox/convox/pkg/atom/pkg/apis/atom/v1"
	versioned "github.com/convox/convox/pkg/atom/pkg/client/clientset/versioned"
	internalinterfaces "github.com/convox/convox/pkg/atom/pkg/client/informers/externalversions/internalinterfaces"
	v1 "github.com/convox/convox/pkg/atom/pkg/client/listers/atom/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// AtomVersionInformer provides access to a shared informer and lister for
// AtomVersions.
type AtomVersionInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.AtomVersionLister
}

type atomVersionInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewAtomVersionInformer constructs a new informer for AtomVersion type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewAtomVersionInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredAtomVersionInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredAtomVersionInformer constructs a new informer for AtomVersion type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredAtomVersionInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AtomV1().AtomVersions(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AtomV1().AtomVersions(namespace).Watch(context.TODO(), options)
			},
		},
		&atomv1.AtomVersion{},
		resyncPeriod,
		indexers,
	)
}

func (f *atomVersionInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredAtomVersionInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *atomVersionInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&atomv1.AtomVersion{}, f.defaultInformer)
}

func (f *atomVersionInformer) Lister() v1.AtomVersionLister {
	return v1.NewAtomVersionLister(f.Informer().GetIndexer())
}
