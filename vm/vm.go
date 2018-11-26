package vm

import (
  "fmt"
  "encoding/json"
  "github.com/parnurzeal/gorequest"
  "github.com/gin-gonic/gin"
  "../config"
  "net/http"
)

type UserInputInfo struct {
    Server struct {
        Name                      string `json:"name"`
        FlavorRef               string `json:"flavorRef"`
        ImageRef               string `json:"imageRef"`
        AvailabilityZone  string `json:"availability_zone"`
        Networks []Network `json:"networks"`
    } `json:"server"`
}

type Network struct {
    Uuid string `json:"uuid"`
}

type VMOutput struct {
    Server struct {
        ID string `json:"id"`
        AdminPass string `json:"adminPass"`
    } `json:"server"`
    Forbidden struct {
        Message string `json:"message"`
    } `json:"forbidden"`
    BadRequest struct {
        Message string `json:"message"`
    } `json:"badRequest"`
    Unauthorized struct {
        Message string `json:"message"`
    } `json:"unauthorized"`
    ItemNotFound struct {
        Message string `json:"message"`
    } `json:"itemNotFound"`
    Conflict struct {
        Message string `json:"message"`
    } `json:"conflict"`
}

type VMsList struct {
    Servers []Server `json:"servers"`
}

type Server struct {
    Name string `json:"name"`
}

type VMStatusInfo struct {
    Server struct {
        Addresses struct {
            Provider []ip `json:"provider"`
        } `json:"addresses"`
        Status string `json:"status"`
        Fault struct {
            Message  string `json:"message"`
            Details  string `json:"details"`
            Created  string `json:"created"`
        } `json:"fault"`
        Created string `json:"created"`
        Updated string `json:"updated"`
        TenantID string `json:"tenant_id"`
        Name string `json:"name"`
    } `json:"server"`
}

type ip struct {
    Addr string `json:"addr"`
}

func VMCreate( c *gin.Context ) {
  request := gorequest.New()
  token_resp, token_body, token_err := request.Post(config.TokenApi). //Here we are taking this Identity API endpoint from local environment
      Set("accept", "application/json").
      Set("Content-Type", "application/json").
      Send(config.Token()).
      End()

  if token_err != nil {
      fmt.Println("Unable to retrieve token")
  }

  auth_token := token_resp.Header["X-Subject-Token"][0]
  fmt.Println(token_resp,"\n+++\n",token_body)
  fmt.Println(auth_token)

  launch_vm := UserInputInfo{}
  launch_vm.Server.Name = c.PostForm("name")

  flavor_input := c.PostForm("flavorRef")
  if flavor_input == "Silver" {
    launch_vm.Server.FlavorRef =  config.VMFlavSmall
  } else if flavor_input == "Gold" {
    launch_vm.Server.FlavorRef = config.VMFlavMedium
  } else if flavor_input == "Platinum" {
    launch_vm.Server.FlavorRef = config.VMFlavLarge
  }

  image_input := c.PostForm("imageRef")
  if image_input == "Ubuntu16.04" {
      launch_vm.Server.ImageRef =  config.UbuntuID
  } else if image_input == "CentOS7" {
      launch_vm.Server.ImageRef = config.CentOSID
  } else if image_input == "Cirros" {
      launch_vm.Server.ImageRef = config.CirrosID
  }
  launch_vm.Server.Networks = make([]Network, 1)
  launch_vm.Server.Networks[0].Uuid = config.NetworkID

  launch_vm.Server.AvailabilityZone = c.PostForm("availability_zone")

  vm_info, _ := json.Marshal(&launch_vm)
  fmt.Println(string(vm_info))

  _, list_vms, _ := request.Get(config.ComputeApi).
          Set("X-Auth-Token",auth_token).
          Set("accept", "application/json").
          Set("Content-Type", "application/json").
          End()

  fmt.Println("\n\n==++++++++++++++++++++==\n\n",list_vms,"\n\n=++++++++++++++++++++=\n\n")

  var ListObj VMsList
  _ = json.Unmarshal([]byte(list_vms),&ListObj)

  for _, v := range ListObj.Servers {
     if v.Name == launch_vm.Server.Name {
        c.JSON(http.StatusOK, gin.H{
          "message": fmt.Sprintf("VM instance with name %s already exists on OpenStack -+- Please try a different name",launch_vm.Server.Name),
          "error": true,})
        return
      }
  }

  _, VMOutputInfo, _ := request.Post(config.ComputeApi).
          Set("X-Auth-Token",auth_token).
          Set("accept", "application/json").
          Set("Content-Type", "application/json").
          Send(string(vm_info)).
          End()

  fmt.Println("\n\n==++++++++++++++++++++==\n\n",VMOutputInfo,"\n\n=++++++++++++++++++++=\n\n")

  var msg VMOutput

  _ = json.Unmarshal([]byte(VMOutputInfo), &msg)

  vm_id := msg.Server.ID
  admin_pass := msg.Server.AdminPass

  response := make(map[string]string)
  response = map[string]string{
      "token": auth_token,
      "server_id": vm_id,
      "admin_pass": admin_pass,
  }

  if msg.Forbidden.Message != "" {
      c.JSON(http.StatusOK, gin.H{
           "message": msg.Forbidden.Message,
           "error": true,})
         return
  } else if msg.BadRequest.Message != "" {
      c.JSON(http.StatusOK, gin.H{
           "message": msg.BadRequest.Message,
           "error": true,})
         return
  } else if msg.Unauthorized.Message != "" {
      c.JSON(http.StatusOK, gin.H{
           "message": msg.Unauthorized.Message,
           "error": true,})
         return
  } else if msg.ItemNotFound.Message != "" {
      c.JSON(http.StatusOK, gin.H{
           "message": msg.ItemNotFound.Message,
           "error": true,})
         return
  } else if msg.Conflict.Message != "" {
      c.JSON(http.StatusOK, gin.H{
           "message": msg.Conflict.Message,
           "error": true,})
         return
  } else {
      c.JSON(http.StatusOK, gin.H{
           "message": response,
           "error": false,})
         return
  }
  fmt.Println("\n\n==++++++++++++++++++++==\n\n", vm_id,"\n\n==++++++++++++++++++++==\n\n",admin_pass,"+++",msg.Forbidden.Message)
}

func VMStatus( c *gin.Context ) {
  request := gorequest.New()
  auth_token := c.Request.FormValue("token")
  vm_id := c.Request.FormValue("id")

  status_api := fmt.Sprintf("%s/%s",config.ComputeApi,vm_id)

  _, status_body, _ := request.Get(status_api).
          Set("X-Auth-Token",auth_token).
          Set("accept", "application/json").
          Set("Content-Type", "application/json").
          End()

  var out VMStatusInfo

  _ = json.Unmarshal([]byte(status_body), &out)

  response_error := make(map[string]string)
  response_error = map[string]string{
      "Status": out.Server.Status,
      "Message": out.Server.Fault.Message,
      "created": out.Server.Fault.Created,
      "TenantID": out.Server.TenantID,
      "Name": out.Server.Name,
  }

  if out.Server.Status == "ACTIVE" {
      response_active := make(map[string]string)
      response_active = map[string]string{
        "Status": out.Server.Status,
        "IpAddr": out.Server.Addresses.Provider[0].Addr,
        "Created": out.Server.Created,
        "Updated": out.Server.Updated,
        "TenantID": out.Server.TenantID,
        "Name": out.Server.Name,
      }
      c.JSON(http.StatusOK, gin.H{
          "message": response_active,
          "error": false,})
  } else if out.Server.Status == "ERROR" {
      c.JSON(http.StatusOK, gin.H{
          "message": response_error,
          "error": true,})
  }
}
