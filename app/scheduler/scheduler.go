package scheduler

import (
	"context"
	"fake-deviceplugin/pkg/k8s"
	"fake-deviceplugin/pkg/log"
	"fake-deviceplugin/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	schedulerapi "k8s.io/kube-scheduler/extender/v1"
)

type Schedueler struct {
}

func NewScheduler() *Schedueler {
	return &Schedueler{}
}

func (sched *Schedueler) Start() {
	r := gin.Default()

	r.POST("/scheduler/predicates", sched.Filter)
	r.POST("/scheduler/priorities", sched.Score)
	r.POST("/scheduler/bind", sched.Bind)

	// Test Kubernetes APIServer
	_ = k8s.GetKubeClient(utils.GetInitContext())
	go func() { utils.PanicIfError(r.Run(":8888")) }()
}

func (sched *Schedueler) Filter(c *gin.Context) {
	var request schedulerapi.ExtenderArgs
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	// ctx := utils.NewContext()
	// log.Debug(ctx, "Scheduler Filter called, request: %s", utils.JsonString(request))

	var result schedulerapi.ExtenderFilterResult
	result.Nodes = request.Nodes
	result.NodeNames = request.NodeNames
	c.JSON(http.StatusOK, &result)
}

func (sched *Schedueler) Score(c *gin.Context) {
	var request schedulerapi.ExtenderArgs
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	ctx := utils.NewContext()

	var result schedulerapi.HostPriorityList
	if request.NodeNames != nil {
		for _, nodeName := range *request.NodeNames {
			result = append(result, schedulerapi.HostPriority{Host: nodeName, Score: 10})
		}
	} else if request.Nodes != nil {
		for _, node := range request.Nodes.Items {
			result = append(result, schedulerapi.HostPriority{Host: node.Name, Score: 10})
		}
	}
	log.Debug(ctx, "Scheduler Score called, request: %s, response: %s", utils.JsonString(request), utils.JsonString(result))

	c.JSON(http.StatusOK, &result)
}

func (sched *Schedueler) Bind(c *gin.Context) {
	var request schedulerapi.ExtenderBindingArgs
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := utils.NewContext()
	log.Debugf(ctx, "Scheduler Bind called, request: %s", utils.JsonString(request))

	binding := &v1.Binding{
		ObjectMeta: metav1.ObjectMeta{Name: request.PodName, UID: request.PodUID},
		Target:     v1.ObjectReference{Kind: "Node", Name: request.Node},
	}

	err := k8s.GetKubeClient(ctx).CoreV1().Pods(request.PodNamespace).Bind(context.Background(), binding, metav1.CreateOptions{})
	utils.PanicIfError(err)

	var result schedulerapi.ExtenderBindingResult
	c.JSON(http.StatusOK, &result)
}
