package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/types"
	"github.com/sirupsen/logrus"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

var (
	resyncPeriod = 2 * time.Hour
)

type HandlerFunc func(key string) error

type GenericController interface {
	Informer() cache.SharedIndexInformer
	AddHandler(name string, handler HandlerFunc)
	HandlerCount() int
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type handlerDef struct {
	name    string
	handler HandlerFunc
}

type genericController struct {
	sync.Mutex
	informer cache.SharedIndexInformer
	handlers []handlerDef
	queue    workqueue.RateLimitingInterface
	name     string
	running  bool
	synced   bool
}

func NewGenericController(name string, objectClient *clientbase.ObjectClient) GenericController {
	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc:  objectClient.List,
			WatchFunc: objectClient.Watch,
		},
		objectClient.Factory.Object(), resyncPeriod, cache.Indexers{})

	return &genericController{
		informer: informer,
		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(),
			name),
		name: name,
	}
}

func (g *genericController) HandlerCount() int {
	return len(g.handlers)
}

func (g *genericController) Informer() cache.SharedIndexInformer {
	return g.informer
}

func (g *genericController) Enqueue(namespace, name string) {
	if namespace == "" {
		g.queue.Add(name)
	} else {
		g.queue.Add(namespace + "/" + name)
	}
}

func (g *genericController) AddHandler(name string, handler HandlerFunc) {
	g.handlers = append(g.handlers, handlerDef{
		name:    name,
		handler: handler,
	})
}

func (g *genericController) Sync(ctx context.Context) error {
	g.Lock()
	defer g.Unlock()

	return g.sync(ctx)
}

func (g *genericController) sync(ctx context.Context) error {
	if g.synced {
		return nil
	}

	defer utilruntime.HandleCrash()

	g.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: g.queueObject,
		UpdateFunc: func(_, obj interface{}) {
			g.queueObject(obj)
		},
		DeleteFunc: g.queueObject,
	})

	logrus.Infof("Syncing %s Controller", g.name)

	go g.informer.Run(ctx.Done())

	if !cache.WaitForCacheSync(ctx.Done(), g.informer.HasSynced) {
		return fmt.Errorf("failed to sync controller %s", g.name)
	}
	logrus.Infof("Syncing %s Controller Done", g.name)

	g.synced = true
	return nil
}

func (g *genericController) Start(ctx context.Context, threadiness int) error {
	g.Lock()
	defer g.Unlock()

	if !g.synced {
		if err := g.sync(ctx); err != nil {
			return err
		}
	}

	if !g.running {
		go g.run(ctx, threadiness)
	}

	g.running = true
	return nil
}

func (g *genericController) queueObject(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err == nil {
		g.queue.Add(key)
	}
}

func (g *genericController) run(ctx context.Context, threadiness int) {
	defer utilruntime.HandleCrash()
	defer g.queue.ShutDown()

	for i := 0; i < threadiness; i++ {
		go wait.Until(g.runWorker, time.Second, ctx.Done())
	}

	<-ctx.Done()
	logrus.Infof("Shutting down %s controller", g.name)
}

func (g *genericController) runWorker() {
	for g.processNextWorkItem() {
	}
}

func (g *genericController) processNextWorkItem() bool {
	key, quit := g.queue.Get()
	if quit {
		return false
	}
	defer g.queue.Done(key)

	// do your work on the key.  This method will contains your "do stuff" logic
	err := g.syncHandler(key.(string))
	checkErr := err
	if handlerErr, ok := checkErr.(*handlerError); ok {
		checkErr = handlerErr.err
	}
	if _, ok := checkErr.(*ForgetError); err == nil || ok {
		if ok {
			logrus.Infof("%v %v completed with dropped err: %v", g.name, key, err)
		}
		g.queue.Forget(key)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("%v %v %v", g.name, key, err))
	g.queue.AddRateLimited(key)

	return true
}

func (g *genericController) syncHandler(s string) (err error) {
	defer utilruntime.RecoverFromPanic(&err)

	var errs []error
	for _, handler := range g.handlers {
		if err := handler.handler(s); err != nil {
			errs = append(errs, &handlerError{
				name: handler.name,
				err:  err,
			})
		}
	}
	err = types.NewErrors(errs)
	return
}

type handlerError struct {
	name string
	err  error
}

func (h *handlerError) Error() string {
	return fmt.Sprintf("[%s] failed with : %v", h.name, h.err)
}
