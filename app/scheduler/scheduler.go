package scheduler

import (
	"fake-deviceplugin/pkg/log"
	"fake-deviceplugin/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	//  "k8s.io/kubernetes/pkg/scheduler/apis/extender/v1beta1"
	schedulerapi "k8s.io/kube-scheduler/extender/v1"
)

type Schedueler struct {
}

func NewScheduler() *Schedueler {
	return &Schedueler{}
}

func (sched *Schedueler) Start() error {
	r := gin.Default()

	r.POST("/filter", sched.Filter)
	r.POST("/score", sched.Score)
	r.POST("/bind", sched.Bind)

	return r.Run(":8888")
}

func (sched *Schedueler) Filter(c *gin.Context) {
	var request schedulerapi.ExtenderArgs
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	log.Debug("Scheduler Filter called, request: %s", utils.JsonString(request))

	var result schedulerapi.ExtenderFilterResult
	result.Nodes = request.Nodes
	result.NodeNames = request.NodeNames
	c.JSON(http.StatusOK, &result)
}

func (sched *Schedueler) Score(c *gin.Context) {
	// var args schedulerapi.ExtenderArgs
	// if err := c.ShouldBindJSON(&args); err != nil {
	// 	c.JSON(400, gin.H{"error": err.Error()})
	// 	return
	// }

	// var result schedulerapi.ExtenderScoreResult
	// c.JSON(http.StatusOK, &result)

}

func (sched *Schedueler) Bind(c *gin.Context) {
	var request schedulerapi.ExtenderBindingArgs
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	log.Debug("Scheduler Bind called, request: %s", utils.JsonString(request))

	var result schedulerapi.ExtenderBindingResult
	c.JSON(http.StatusOK, &result)
}
