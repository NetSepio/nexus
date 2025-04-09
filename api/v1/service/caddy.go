package caddy

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/NetSepio/nexus/api/v1/middleware"
	"github.com/NetSepio/nexus/api/v1/service/util"
	"github.com/NetSepio/nexus/core"
	"github.com/NetSepio/nexus/model"
	"github.com/gin-gonic/gin"
)

// ApplyRoutes applies router to gin Router
func ApplyRoutes(r *gin.RouterGroup) {

	g := r.Group("/caddy")
	{
		g.POST("", AddServices)
		g.GET("", getServices)
		g.GET(":name", getService)
		g.DELETE(":name", deleteService)
	}
}

var resp map[string]interface{}

// addTunnel adds new tunnel config
func AddServices(c *gin.Context) {
	//post form parameters
	var payload ServicePayload

	// Bind JSON payload to the struct
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// convert port string to int
	portInt, err := strconv.Atoi(payload.Port)
	if err != nil {
		resp = util.Message(400, "Invalid Port")
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	for {

		// check validity of Services name and port
		value, msg, err := middleware.IsValidService(payload.Name, portInt, payload.IPAddress)

		if err != nil {
			resp = util.Message(500, "Server error, Try after some time or Contact Admin..."+err.Error())
			c.JSON(http.StatusOK, resp)
			break
		} else if value == -1 {
			if msg == "Port Already in use" {
				continue
			}

			resp = util.Message(404, msg)
			c.JSON(http.StatusBadRequest, resp)
			break
		} else if value == 1 {
			//create a Services struct object
			var data model.Service
			data.Name = payload.Name
			data.Type = os.Getenv("NODE_TYPE")
			data.Port = payload.Port
			data.Domain = os.Getenv("DOMAIN")
			data.IpAddress = payload.IPAddress
			data.CreatedAt = time.Now().UTC().Format(time.RFC3339)

			//to add Services config
			err := middleware.AddServices(data)
			if err != nil {
				resp = util.Message(500, "Server error, Try after some time or Contact Admin..."+err.Error())
				c.JSON(http.StatusInternalServerError, resp)
				break
			} else {
				resp = util.MessageService(200, data)
				c.JSON(http.StatusOK, resp)
				break
			}
		}
	}
}

// getServices gets all Services config
func getServices(c *gin.Context) {
	services, err := middleware.ReadServices()
	if err != nil {
		resp = util.Message(500, "Server error, Try after some time or Contact Admin...")
		c.JSON(http.StatusInternalServerError, resp)
		return
	}
	c.JSON(http.StatusOK, services)
}

// getServices get specific Services config
func getService(c *gin.Context) {
	//get parameter
	name := c.Param("name")

	//read Services config
	Services, err := middleware.ReadService(name)
	if err != nil {
		resp = util.Message(500, "Server error, Try after some time or Contact Admin...")
		c.JSON(http.StatusInternalServerError, resp)
	}

	//check if Services exists
	if Services.Name == "" {
		resp = util.Message(404, "Service Doesn't Exists")
		c.JSON(http.StatusNotFound, resp)
	} else {
		port, err := strconv.Atoi(Services.Port)
		if err != nil {
			util.LogError("string conv error: ", err)
			resp = util.Message(500, "Server error, Try after some time or Contact Admin...")
			c.JSON(http.StatusInternalServerError, resp)
		} else {
			status, err := core.ScanPort(port)
			if err != nil {
				resp = util.Message(500, "Server error, Try after some time or Contact Admin...")
				c.JSON(http.StatusInternalServerError, resp)
			} else {
				Services.Status = status
				resp = util.MessageService(200, *Services)
				c.JSON(http.StatusOK, resp)
			}
		}
	}
}

func deleteService(c *gin.Context) {
	//get parameter
	name := c.Param("name")

	//read Services config
	Services, err := middleware.ReadService(name)
	if err != nil {
		resp = util.Message(500, "Server error, Try after some time or Contact Admin...")
		c.JSON(http.StatusInternalServerError, resp)
	}

	//check if Services exists
	if Services.Name == "" {
		resp = util.Message(400, "Service Doesn't Exists")
		c.JSON(http.StatusBadRequest, resp)
	} else {
		//delete Services config
		err = middleware.DeleteService(name)
		if err != nil {
			resp = util.Message(500, "Server error, Try after some time or Contact Admin...")
			c.JSON(http.StatusInternalServerError, resp)
		} else {
			resp = util.Message(200, "Deleted Services "+name)
			c.JSON(http.StatusOK, resp)
		}
	}

}

// func MiddlewareForCaddy(c *gin.Context) {

// 	//check if NODE_CONFIG is set to standard or hpc

// 	if strings.ToLower(os.Getenv("NODE_CONFIG")) != "standard" && strings.ToLower(os.Getenv("NODE_CONFIG")) != "hpc" {
// 		util.LogError("NODE_CONFIG not allowed", nil)
// 		c.JSON(http.StatusNotAcceptable, resp)
// 		os.Exit(1)
// 	}
// }

// NodeConfigMiddleware checks if NODE_CONFIG is set to "standard" or "hpc".
func MiddlewareForCaddy() gin.HandlerFunc {
	return func(c *gin.Context) {
		nodeConfig := os.Getenv("NODE_CONFIG")

		if nodeConfig != "nexus" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid NODE_CONFIG value. It must be 'nexus' or 'titan'.",
			})
			c.Abort() // Stop further processing of the request
			return
		}

		// Pass to the next middleware/handler
		c.Next()
	}
}

// AddServicesDirect adds a service using direct arguments.
func AddServicesDirect(domain string, agentName string, port int) error {
	ipAddress := "127.0.0.1" // Replace with actual IP logic if needed

	// Validate the service
	value, msg, err := middleware.IsValidService(agentName, port, ipAddress)
	if err != nil {
		return fmt.Errorf("server error: %v", err)
	}

	if value == -1 {
		return fmt.Errorf("validation failed: %s", msg)
	}

	// Create a Services struct object
	var data model.Service
	data.Name = agentName
	data.Type = os.Getenv("NODE_TYPE")
	data.Port = strconv.Itoa(port)
	data.Domain = agentName + "." + domain
	data.IpAddress = ipAddress
	data.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	// Add the service
	err = middleware.AddServices(data)
	if err != nil {
		return fmt.Errorf("error adding service: %v", err)
	}

	return nil
}
