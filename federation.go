package main

import (
	"encoding/json"
	"errors"
	"gopkg.in/macaron.v1"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

var Federation = &EKCPController{}
var ClientTimeoutSeconds = os.Getenv("CLIENT_TIMEOUT_SECONDS")

type EKCPController struct {
	sync.Mutex
	Clusters []EKCPServer
}

type EKCPServer struct {
	Id       int
	Endpoint string `form:"endpoint" binding:"Required"`
	client   *http.Client
}

func (c *EKCPController) HasSlaves() bool {
	c.Lock()
	defer c.Unlock()
	return len(c.Clusters) != 0
}

func (c *EKCPController) Register(e EKCPServer) {
	c.Lock()
	defer c.Unlock()
	exists := false
	for _, ec := range c.Clusters {
		if ec.Endpoint == e.Endpoint {
			exists = true
		}
	}
	if !exists {
		c.Clusters = append(c.Clusters, e)
	}
}

func (c *EKCPController) Unregister(id int) error {
	c.Lock()
	defer c.Unlock()

	if len(c.Clusters) < id+1 {
		return errors.New("Endpoint not found")
	}
	c.Clusters = c.Clusters[:id+copy(c.Clusters[id:], c.Clusters[id+1:])]
	return nil
}

func (c *EKCPController) Show(id int) (EKCPServer, error) {
	c.Lock()
	defer c.Unlock()
	if len(c.Clusters) < id+1 {
		return EKCPServer{}, errors.New("Endpoint not found")
	}
	cl := c.Clusters[id]
	cl.Id = id
	return cl, nil
}

func (c *EKCPController) List() []KubernetesCluster {
	c.Lock()
	defer c.Unlock()
	kubeClusters := []KubernetesCluster{}
	for _, e := range c.Clusters {
		if res, err := e.Status(); err == nil {
			for _, kubeC := range res.Clusters {
				// annotate that the cluster is federated
				kubeC.Federated = true
				kubeC.InstanceEndpoint = e.Endpoint
				kubeClusters = append(kubeClusters, kubeC)
			}
		}
	}
	return kubeClusters
}

func (c *EKCPController) Registered() []EKCPServer {
	c.Lock()
	defer c.Unlock()
	return c.Clusters
}

func (c *EKCPController) Search(clustername string) (KubernetesCluster, error) {
	c.Lock()
	defer c.Unlock()
	for _, e := range c.Clusters {
		if found, _ := e.Exists(clustername); found {
			if kubeCluster, err := e.GetCluster(clustername); err == nil {
				// annotate that the cluster is federated
				kubeCluster.Federated = true
				kubeCluster.InstanceEndpoint = e.Endpoint
				return kubeCluster, nil
			}
		}
	}
	return KubernetesCluster{}, errors.New("Cluster not found")
}

func (c *EKCPController) Delete(clustername string) error {
	c.Lock()
	defer c.Unlock()
	for _, e := range c.Clusters {
		if found, _ := e.Exists(clustername); found {
			if err := e.DeleteCluster(clustername); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *EKCPController) Allocate(kc KubernetesCluster) error {
	c.Lock()
	defer c.Unlock()
	var active []int
	for _, e := range c.Clusters {
		runningC, err := e.ActiveClusters()
		if err != nil {
			active = append(active, 99999)
			continue
		}
		active = append(active, runningC)
	}

	if len(active) == 0 {
		return errors.New("No servers available to allocate the request")
	}
	chosenC := FindMin(active)

	return c.Clusters[chosenC].CreateCluster(kc)
}

func FindMin(capacity []int) (index int) {
	min := capacity[0]
	index = 0
	for i, c := range capacity {
		if c < min {
			min = c
			index = i
		}
	}
	return
}
func (c *EKCPServer) generateClient() *http.Client {
	if c.client == nil {
		var timeout int
		var err error
		if len(ClientTimeoutSeconds) > 0 {
			if timeout, err = strconv.Atoi(ClientTimeoutSeconds); err != nil {
				panic("Invalid input for CLIENT_TIMEOUT_SECONDS")
			}
		} else {
			timeout = 30
		}
		c.client = &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		}
	}

	return c.client
}

func (c *EKCPServer) Status() (APIResult, error) {
	var res APIResult

	response, err := c.generateClient().Get(c.Endpoint)
	if err != nil {
		return res, err
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return res, err
	}
	err = json.Unmarshal(contents, &res)
	if err != nil {
		return res, err
	}
	if len(res.Error) > 0 {
		return res, errors.New(res.Error)
	}
	return res, nil
}

func (c *EKCPServer) CreateCluster(kc KubernetesCluster) error {
	var res APIResult

	response, err := http.PostForm(c.Endpoint+"/api/v1/cluster/new",
		url.Values{
			"name":       {kc.Name},
			"node_image": {kc.NodeImage},
			"kindconfig": {kc.RawEncodedKindConfig},
			"version":    {kc.Version},
		})
	if err != nil {
		return err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return err
	}
	if len(res.Error) != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (c *EKCPServer) DeleteCluster(clustername string) error {
	var res APIResult

	// Create client
	client := c.generateClient()

	// Create request
	req, err := http.NewRequest("DELETE", c.Endpoint+"/api/v1/cluster/"+clustername, nil)
	if err != nil {
		return err
	}

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		return err

	}
	defer resp.Body.Close()

	// Read Response Body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err

	}
	err = json.Unmarshal(respBody, &res)
	if err != nil {
		return err
	}
	if len(res.Error) > 0 {
		return errors.New(res.Error)
	}

	return nil
}

func (c *EKCPServer) Exists(clustername string) (bool, error) {

	res, err := c.Status()
	if err != nil {
		return false, err
	}

	_, ok := res.Clusters[clustername]

	return ok, nil
}

func (c *EKCPServer) ActiveClusters() (int, error) {
	res, err := c.Status()
	if err != nil {
		return 0, err
	}

	return len(res.AvailableClusters), nil
}

func (c *EKCPServer) Clusters() (map[string]KubernetesCluster, error) {
	res, err := c.Status()
	if err != nil {
		return res.Clusters, err
	}

	return res.Clusters, nil
}

func (c *EKCPServer) GetCluster(clustername string) (KubernetesCluster, error) {
	res, err := c.Clusters()
	if err != nil {
		return KubernetesCluster{}, err
	}

	kubeC, ok := res[clustername]
	if !ok {
		return KubernetesCluster{}, errors.New("No cluster found")
	}
	return kubeC, nil
}

/*
* Macaron plugin
 */

func RegisterClusterToFederation(ctx *macaron.Context, ekcp EKCPServer) {
	Federation.Register(ekcp)
	ctx.JSON(200, NewAPIResult(strconv.Itoa(len(Federation.Registered())-1)))
}

func SendRegistrationRequest() error {
	var res APIResult
	var timeout int
	var err error
	if len(ClientTimeoutSeconds) > 0 {
		if timeout, err = strconv.Atoi(ClientTimeoutSeconds); err != nil {
			panic("Invalid input for CLIENT_TIMEOUT_SECONDS")
		}
	} else {
		timeout = 30
	}
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	response, err := client.PostForm(os.Getenv("FEDERATION_MASTER")+"/api/v1/federation/register", url.Values{"endpoint": {"http://" + os.Getenv("KUBEHOST") + ":" + os.Getenv("PORT")}})
	if err != nil {
		return err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return err
	}
	if len(res.Error) != 0 {
		return errors.New(res.Error)
	}
	return nil
}
