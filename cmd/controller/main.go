package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/transport"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog"

	"github.com/damnever/useless/controller"
	clientset "github.com/damnever/useless/pkg/generated/clientset/versioned"
	informers "github.com/damnever/useless/pkg/generated/informers/externalversions"
)

func main() {
	var (
		flagMasterURL  string
		flagKubeConfig string
	)
	flag.StringVar(&flagMasterURL, "master", "",
		"The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	if _, err := os.Stat(kubeconfig); err == nil {
		flag.StringVar(&flagKubeConfig, "kubeconfig", kubeconfig,
			"Absolute path to the kubeconfig file. Only required if out-of-cluster.")
	} else {
		flag.StringVar(&flagKubeConfig, "kubeconfig", "",
			"Absolute path to the kubeconfig file. Only required if out-of-cluster.")
	}
	flag.Parse()
	klog.SetOutput(os.Stdout)

	config, err := clientcmd.BuildConfigFromFlags(flagMasterURL, flagKubeConfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err)
	}
	kubeClient := kubernetes.NewForConfigOrDie(config)
	klog.Infof("KubeClient configured")

	// Ref: https://github.com/kubernetes/client-go/blob/master/examples/leader-election/main.go
	// FIXME(damnever): keeps followers' cache synced so it can respond quickly when the leader step down
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	watchStopSignals(cancel)
	config.Wrap(transport.ContextCanceller(ctx, fmt.Errorf("the leader is shutting down")))

	const leaseLockNamespace = "useless"
	const leaseLockName = "useless-controller"
	ID := os.Getenv("ID")
	klog.Infof("%s: starting leader election", ID)
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock: &resourcelock.LeaseLock{
			LeaseMeta: metav1.ObjectMeta{
				Name:      leaseLockName,
				Namespace: leaseLockNamespace,
			},
			Client: kubeClient.CoordinationV1(),
			LockConfig: resourcelock.ResourceLockConfig{
				Identity: ID,
			},
		},
		ReleaseOnCancel: true,
		LeaseDuration:   60 * time.Second,
		RenewDeadline:   15 * time.Second,
		RetryPeriod:     5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				// we're notified when we start - this is where you would
				// usually put your code
				klog.Infof("%s: leading", ID)
				runController(kubeClient, config, ctx.Done())
			},
			OnStoppedLeading: func() {
				// we can do cleanup here, or after the RunOrDie method
				// returns
				klog.Infof("%s: lost", ID)
			},
			OnNewLeader: func(identity string) {
				// we're notified when new leader elected
				if identity == ID {
					// I just got the lock
					return
				}
				klog.Infof("new leader elected: %v", identity)
			},
		},
	})
	// because the context is closed, the client should report errors
	_, err = kubeClient.CoordinationV1().Leases(leaseLockNamespace).Get(leaseLockName, metav1.GetOptions{})
	if err == nil || !strings.Contains(err.Error(), "the leader is shutting down") {
		klog.Fatalf("%s: expected to get an error when trying to make a client call: %v", ID, err)
	}
}

func runController(kubeClient *kubernetes.Clientset, config *rest.Config, stopCh <-chan struct{}) {
	uselessClient := clientset.NewForConfigOrDie(config)
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	uselessInformerFactory := informers.NewSharedInformerFactory(uselessClient, time.Second*30)
	controller := controller.New(kubeClient, uselessClient,
		kubeInformerFactory.Apps().V1().Deployments(),
		kubeInformerFactory.Core().V1().Services(),
		kubeInformerFactory.Autoscaling().V1().HorizontalPodAutoscalers(),
		uselessInformerFactory.Useless().V1().Functions())

	kubeInformerFactory.Start(stopCh)
	uselessInformerFactory.Start(stopCh)
	if err := controller.Run(5, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}

func watchStopSignals(cancel context.CancelFunc) {
	sigc := make(chan os.Signal, 2)
	signal.Notify(sigc, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigc
		cancel()
		<-sigc
		os.Exit(1)
	}()
}
