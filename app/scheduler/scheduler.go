package scheduler

import (
	"github.com/gin-gonic/gin"
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

	return r.Run()
}

func (sched *Schedueler) Filter(c *gin.Context) {

}

func (sched *Schedueler) Score(c *gin.Context) {

}

func (sched *Schedueler) Bind(c *gin.Context) {

}
