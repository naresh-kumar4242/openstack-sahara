// Purpose : Hadoop as a service on Openstack - We'll use Sahara REST-API to launch Hadoop cluster & VM on OpenStack 

package main

import (
	"fmt"
    "encoding/json"
    "github.com/parnurzeal/gorequest"
	"github.com/gin-gonic/gin"
	"os"
	"net/http"
	//"time"
	"strings"
	"github.com/gin-contrib/cors"
	"./email"
	"./config"
	"./vm"
)

type AuthStruct struct {
	Auth struct {
		Identity struct {
			Methods []string `json:"methods"`
			Password struct {
				User struct {
					Domain struct {
						Name string `json:"name"`
					} `json:"domain"`
					Name string `json:"name"`
					Password string `json:"password"`
				} `json:"user"`
			} `json:"password"`
		} `json:"identity"`
		Scope struct {
			Project struct {
				Domain struct {
					Name string `json:"name"`
				} `json:"domain"`
				Name string `json:"name"`
			} `json:"project"`
		} `json:"scope"`
	} `json:"auth"`
}

type launch_cluster struct {
	PluginName               string `json:"plugin_name"`
	HadoopVersion            string `json:"hadoop_version"`
	ClusterTemplateID        string `json:"cluster_template_id"`
	DefaultImageID           string `json:"default_image_id"`
	UserKeypairID            string `json:"user_keypair_id"`
	Name                     string `json:"name"`
	NeutronManagementNetwork string `json:"neutron_management_network"`
  IsPublic                 bool   `json:"is_public"`
}

type AutoGenerated struct {
	Cluster struct {
		ID                  string      `json:"id"`
	}
}

type return_info struct {
	Cluster struct {
		Status                   string        `json:"status"`
		StatusDescription 			 string 			 `json:"status_description"`
		Info                     struct {
			HDFS struct {
				NameNode string `json:"NameNode"`
				WebUI    string `json:"Web UI"`
			} `json:"HDFS"`
			JobFlow struct {
				Oozie string `json:"Oozie"`
			} `json:"JobFlow"`
			MapReduceJobHistoryServer struct {
				WebUI string `json:"Web UI"`
			} `json:"MapReduce JobHistory Server"`
			YARN struct {
				WebUI           string `json:"Web UI"`
				ResourceManager string `json:"ResourceManager"`
			} `json:"YARN"`
		} `json:"info"`
	} `json:"cluster"`
}

type create_vm struct {
	Server struct {
		Name             					   string `json:"name"`
		FlavorRef            			 string `json:"flavorRef"`
		ImageRef        				  string `json:"imageRef"`
		AvailabilityZone 			string `json:"availability_zone"`
                Networks []Network `json:"networks"`
	} `json:"server"`
}

type Network struct {
	Uuid string `json:"uuid"`
}

func main() {
	// Download OpenStack RC file v3 from API access page in Dashboard === Then run "source admin-openrc.sh" on the machine where this Go binary will run
	router := gin.Default()

	//allowing CORS
	cors_config := cors.DefaultConfig()
	cors_config.AllowAllOrigins = true
	//cors_config.AllowOrigins = []string{"http://localhost"}
	router.Use(cors.New(cors_config))


	payload := AuthStruct{}
  	payload.Auth.Identity.Methods = []string{"password"}
  	payload.Auth.Identity.Password.User.Domain.Name = "Default"
	payload.Auth.Identity.Password.User.Name = os.ExpandEnv("$OS_USERNAME")
	payload.Auth.Identity.Password.User.Password = os.ExpandEnv("$OS_PASSWORD")
  	payload.Auth.Scope.Project.Domain.Name = "Default"
  	payload.Auth.Scope.Project.Name = os.ExpandEnv("$OS_PROJECT_NAME")

	body, err := json.Marshal(&payload)
    if err != nil {
        panic (err)
    }


	identity_api := os.ExpandEnv("$OS_AUTH_URL")
	token_api := fmt.Sprintf("%s/auth/tokens",identity_api)
	s := strings.Split(identity_api, ":")
	sahara_api := fmt.Sprintf("%s:%s:8386/v1.1/%s/clusters",s[0],s[1],os.ExpandEnv("$OS_PROJECT_ID"))

	request := gorequest.New()

	router.POST("/sahara/pass_config", func(c *gin.Context) {

		resp, _, errs := request.Post(token_api). //Here we are taking this Identity API endpoint from local environment
			Set("accept", "application/json").
			Set("Content-Type", "application/json").
			Send(string(body)).
			End()

		if errs != nil {
			fmt.Println("Unable to retrieve token")
		}

		auth_token := resp.Header["X-Subject-Token"][0]
		fmt.Println(auth_token)


		launch := launch_cluster{}

    OS_Flavor := c.PostForm("os_flavor")

		launch.PluginName = "vanilla"  // Pass "vanilla"
		launch.HadoopVersion = c.PostForm("hadoop_version")  //pass "2.7.1"
		//launch.ClusterTemplateID = config.ClusterTemplateID  //pass "030725ec-fa10-4a8f-88b2-0e0999d2f417"  //os.ExpandEnv("$ClusterTemplateID")

		if OS_Flavor == "Ubuntu 14.04" {
			launch.DefaultImageID = config.Ubuntu_14_04_ID   //pass c7cbafef-66e4-4a80-b224-64df965abd04 //os.ExpandEnv("$DefaultImageID")
		} else if OS_Flavor == "CentOS 7" {
			launch.DefaultImageID = config.CentOS_7_ID
		}
		launch.ClusterTemplateID = c.PostForm("subscription_id")

		launch.UserKeypairID = config.KeyPair // pass test
		launch.Name = c.PostForm("name")
		launch.NeutronManagementNetwork = config.InternalNetworkID  //pass 1139ad07-ef8d-49b0-b6f4-b591b14446aa //User will input public N/W.How to convert that to this string?
    launch.IsPublic = true


		body1, err1 := json.Marshal(&launch)
		fmt.Println(string(body1))
		if err1 != nil {
			fmt.Print(err1.Error())
		}

		resp2, body2, errs2 := request.Post(sahara_api).
			Set("X-Auth-Token",auth_token).
			Set("accept", "application/json").
			Set("Content-Type", "application/json").
			Send(string(body1)).
			End()

		fmt.Println("\n\n==++++++++++++++++++++==\n\n", resp2,"\n\n==++++++++++++++++++++==\n\n",body2,"\n\n=++++++++++++++++++++=\n\n")

		if errs2 != nil {
			panic (errs2)
		}

		var msg AutoGenerated
		err = json.Unmarshal([]byte(body2), &msg)
		if err != nil {
			panic (err)
		}
		cluster_id := msg.Cluster.ID
		//cluster_api := fmt.Sprintf("%s/%s",sahara_api,cluster_id)

		response := make(map[string]string)
		response = map[string]string{
			"token": auth_token,
			"cluster_id": cluster_id,
		}
		fmt.Println(cluster_id,response)

		var data map[string]interface{}
		_ = json.Unmarshal([]byte(body2), &data)

		if data["error_name"] == "NAME_ALREADY_EXISTS" {
			c.JSON(http.StatusOK, gin.H{
				"message": fmt.Sprintf("Cluster with name %s already exists.Please try different name.",launch.Name),
				"error": true,})
			return
		} else if data["error_name"] == "VALIDATION_ERROR" {
			c.JSON(http.StatusOK, gin.H{
				"message": fmt.Sprintf("Unable to launch cluster %s. Please fill all details.",launch.Name),
				"error": true,})
			return
		}  else if data["error_name"] == "NOT_FOUND" {
			c.JSON(http.StatusOK, gin.H{
				"message": fmt.Sprintf("Error : %s",data["error_message"]),
				"error": true,})
			return
		} else if data["error_name"] != nil {
			c.JSON(http.StatusOK, gin.H{
				"message": fmt.Sprintf("Error : %s",data["error_message"]),
				"error": true,})
			return
        } else {
					c.JSON(http.StatusOK,gin.H{
					"message":response,
					"error":false	,})
			return
					}})


	router.GET("/sahara/status", func(c *gin.Context) {
		auth_token := c.Request.FormValue("token")
		cluster_id := c.Request.FormValue("cluster-id")
		name := c.Request.FormValue("name")
		receiver_mail  := c.Request.FormValue("email")
		horizon_addr := c.Request.FormValue("horizon")

		cluster_api := fmt.Sprintf("%s/%s",sahara_api,cluster_id)

		creds := fmt.Sprintf(`{"project_id":%s,"cluster_id":%s}`,os.ExpandEnv("$OS_PROJECT_ID"),cluster_id)

		fmt.Println(auth_token,name,creds)

		_, body3, _ := request.Get(cluster_api).
				Set("X-Auth-Token",auth_token).
				Set("accept", "application/json").
				Set("Content-Type", "application/json").
				Send(creds).
				End()

		var output return_info
		err = json.Unmarshal([]byte(body3), &output)

		if output.Cluster.Status == "Active" {
			subj := "Hadoop Cluster Details"
			body := fmt.Sprintf("Hello there mate :)\n\nYour Hadoop Cluster Status & Details ::: Your Hadoop cluster having name : [  %s  ] successfully launched on OpenStack.\n\nHere are the details ::: \n\n   Horizon Address :  %s  \n\n   HDFC WebUI Address :  %s  \n\n   Job Flow oozie Address :  %s \n\n   History UI Address :  %s  \n\n   YARN WebUI Address :  %s  \n\n   YARN Resource Manager Address :  %s ",name, horizon_addr, output.Cluster.Info.HDFS.WebUI, output.Cluster.Info.JobFlow.Oozie, output.Cluster.Info.MapReduceJobHistoryServer.WebUI, output.Cluster.Info.YARN.WebUI, output.Cluster.Info.YARN.ResourceManager)

			email.SendEmail(os.ExpandEnv("$FROM_ADDR"), receiver_mail, subj, body)

			response := make(map[string]string)
			response = map[string]string{
				"Message": fmt.Sprintf("Your Hadoop cluster having name : [ %s ] successfully launched on OpenStack :)\n We've emailed you the following details as well ",name),
				"Status": "Active",
				"Horizon": horizon_addr,
				"HDFC_WebUI_Address": output.Cluster.Info.HDFS.WebUI,
				"Job_Flow_oozie_Address": output.Cluster.Info.JobFlow.Oozie,
				"History_UI" : output.Cluster.Info.MapReduceJobHistoryServer.WebUI,
				"YARN_WebUI" : output.Cluster.Info.YARN.WebUI,
				"YARN_Resource_Manager" : output.Cluster.Info.YARN.ResourceManager,
			}
			c.JSON(http.StatusOK, gin.H{
						"message": response,
						"error": false,
		})
					} else if output.Cluster.Status == "Error" {
			c.JSON(http.StatusOK, gin.H{
				"message": output.Cluster.StatusDescription,
				"error": true,
			})
		} else {
						response := make(map[string]string)
						response = map[string]string{
							"Message": "Please wait while we are launching your Hadoop Cluster . . . Please Be patient :)",
							"Status": output.Cluster.Status,
						}
						c.JSON(http.StatusOK, gin.H{
							"message": response,
							"error": false,
			})}
		})

		router.POST("/sahara/vm",vm.VMCreate)
	  router.GET("/sahara/vm", vm.VMStatus)

		router.Run(":"+config.ServerPort)
}
