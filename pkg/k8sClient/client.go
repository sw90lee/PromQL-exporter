/*
* Samsung-cpc version 1.0
*
*  Copyright ⓒ 2023 kt corp. All rights reserved.
*
*  This is a proprietary software of kt corp, and you may not use this file except in
*  compliance with license agreement with kt corp. Any redistribution or use of this
*  software, with or without modification shall be strictly prohibited without prior written
*  approval of kt corp, and the copyright notice above does not evidence any actual or
*  intended publication of such software.
 */
package k8sClient

import (
	"context"
	"fmt"
	"io"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"kt.com/p5g/cnf-exporter/samsung-cpc/logger"
	"os"
	"path"
	"path/filepath"
)

type Client struct {
	Config    *rest.Config
	Clientset *kubernetes.Clientset
}

func CreateClientSet() (*kubernetes.Clientset, *rest.Config) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		path := filepath.Join(home, ".kube", "config")
		kubeconfig = &path
	} else {
		logger.LogWarn("kubeconfig을 찾을 수 없습니다.")
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return clientset, config
}

func CreateCustomClientSet(configPath string) (*kubernetes.Clientset, *rest.Config) {
	var kubeconfig *string

	kubeconfig = &configPath

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return clientset, config
}

func NewK8sClient(config rest.Config, clientset *kubernetes.Clientset) *Client {
	return &Client{
		Config:    &config,
		Clientset: clientset,
	}
}

func (c *Client) PodList(Namespace string) {
	pods, _ := c.Clientset.CoreV1().Pods(Namespace).List(context.TODO(), v1.ListOptions{})

	for i, pod := range pods.Items {
		fmt.Printf("[%d] %s\n", i, pod.GetName())
	}
}

func CreateToken(client *kubernetes.Clientset, namespace, serviceAccount string) (string, error) {
	treq := &authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			Audiences: []string{"https://kubernetes.default.svc"},
		},
	}
	tokenResp, err := client.CoreV1().ServiceAccounts(namespace).CreateToken(context.TODO(), serviceAccount, treq, v1.CreateOptions{})
	if err != nil {
		logger.LogErr("crate not serviceAccount Token", err)
		return "", err
	}
	if tokenResp.Status.Token == "" {
		logger.LogErr("no service account token returned", fmt.Errorf("no service account token returned"))
		return "", err

	}

	return tokenResp.Status.Token, nil
}

func (c *Client) CopyFromPod(podName, Namespace, containerName string, srcPath string, destPath string) error {
	var stdin io.Reader

	reader, outStream := io.Pipe()
	//todo some containers failed : tar: Refusing to write archive contents to terminal (missing -f option?) when execute `tar cf -` in container
	cmdArr := []string{"tar", "cf", "-", srcPath}

	//pods, _ := c.Clientset.CoreV1().Pods(podName).List(context.TODO(), v1.ListOptions{})

	req := c.Clientset.CoreV1().RESTClient().Post().Resource("pods").Name(podName).Namespace(Namespace).SubResource("exec")

	option := &corev1.PodExecOptions{
		Container: containerName,
		Command:   cmdArr,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}

	if stdin == nil {
		option.Stdin = false
	}

	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)

	exec, err := remotecommand.NewSPDYExecutor(c.Config, "POST", req.URL())
	if err != nil {
		fmt.Println(err)
	}
	go func() {
		defer outStream.Close()
		err = exec.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
			Stdin:  stdin,
			Stdout: outStream,
			Stderr: os.Stderr,
			Tty:    false,
		})
		cmdutil.CheckErr(err)
	}()

	prefix := getPrefix(srcPath)
	prefix = path.Clean(prefix)
	destPath = path.Join(destPath, path.Base(prefix))
	err = untarAll(reader, destPath, prefix)

	return err
}
