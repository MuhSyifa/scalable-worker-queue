package http

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter(jobHandler *JobHandler) *gin.Engine {
	r := gin.Default()

	// Middleware
	r.Use(gin.Recovery())

	api := r.Group("/api/v1")
	{
		jobs := api.Group("/jobs")
		{
			jobs.POST("", jobHandler.CreateJob)
			jobs.GET("/:id", jobHandler.GetJobStatus)
			jobs.DELETE("/:id", jobHandler.CancelJob)
		}

		api.GET("/metrics", jobHandler.GetMetrics)
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "UP"})
		})
	}

	return r
}
